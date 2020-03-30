// Copyright © 2019 IBM Corporation and others.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"unicode"

	"encoding/json"
	"fmt"

	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/mitchellh/go-spdx"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"
)

type ProjectConfig struct {
	Stack           string
	ProjectName     string `mapstructure:"project-name"`
	ApplicationName string `mapstructure:"application-name"`
	Version         string
	Description     string
	License         string
	Maintainers     []Maintainer
}
type OwnerReference struct {
	APIVersion         string `yaml:"apiVersion"`
	Kind               string `yaml:"kind"`
	BlockOwnerDeletion bool   `yaml:"blockOwnerDeletion"`
	Controller         bool   `yaml:"controller"`
	Name               string `yaml:"name"`
	UID                string `yaml:"uid"`
}
type ProjectFile struct {
	Projects []*ProjectEntry `yaml:"projects"`
}
type ProjectEntry struct {
	ID      string    `yaml:"id"`
	Path    string    `yaml:"path"`
	Volumes []*Volume `yaml:"volumes,omitempty"`
}
type Volume struct {
	Name, Path string
}

type NotAnAppsodyProject string

func (e NotAnAppsodyProject) Error() string { return string(e) }

const ConfigFile = ".appsody-config.yaml"

const LatestVersionURL = "https://github.com/appsody/appsody/releases/latest"
const workDirNotSet = ""

const ociKeyPrefix = "org.opencontainers.image."

const appsodyStackKeyPrefix = "dev.appsody.stack."

const appsodyImageCommitKeyPrefix = "dev.appsody.image.commit."

// Checks whether an inode (it does not bother
// about file or folder) exists or not.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

//ExtractDockerEnvFile returns a map with the env vars specified in docker env file
func ExtractDockerEnvFile(envFileName string) (map[string]string, error) {
	envVars := make(map[string]string)
	file, err := os.Open(envFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		equal := strings.Index(line, "=")
		hash := strings.Index(line, "#")
		if equal >= 0 && hash != 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				envVars[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {

		return nil, err
	}

	return envVars, nil
}

//ExtractDockerEnvVars returns a map with the env vars specified in docker options
func ExtractDockerEnvVars(dockerOptions string) (map[string]string, error) {
	//Check whether there's --env-file, this needs to be processed first
	var envVars map[string]string
	envFilePos := strings.Index(dockerOptions, "--env-file=")
	lenFlag := len("--env-file=")
	if envFilePos < 0 {
		envFilePos = strings.Index(dockerOptions, "--env-file")
		lenFlag = len("--env-file")
	}
	if envFilePos >= 0 {
		tokens := strings.Fields(dockerOptions[envFilePos+lenFlag:])
		if len(tokens) > 0 {
			var err error
			envVars, err = ExtractDockerEnvFile(tokens[0])
			if err != nil {
				return nil, err
			}
		}
	} else {
		envVars = make(map[string]string)
	}
	tokens := strings.Fields(dockerOptions)
	for idx, token := range tokens {
		nextToken := ""
		if token == "-e" || token == "--env" {
			if len(tokens) > idx+1 {
				nextToken = tokens[idx+1]
			}
		} else if strings.Contains(token, "-e=") || strings.Contains(token, "-env=") {
			posEqual := strings.Index(token, "=")
			nextToken = token[posEqual+1:]
		}
		if nextToken != "" && strings.Contains(nextToken, "=") {
			nextToken = strings.ReplaceAll(nextToken, "\"", "")
			nextToken = strings.ReplaceAll(nextToken, "'", "")
			//Note that Appsody doesn't support quotes in -e, use --env-file
			keyValuePair := strings.Split(nextToken, "=")
			if len(keyValuePair) > 1 {
				envVars[keyValuePair[0]] = keyValuePair[1]
			}
		}
	}
	return envVars, nil
}

//GetEnvVar obtains a Stack environment variable from the Stack image
func GetEnvVar(searchEnvVar string, config *RootCommandConfig) (string, error) {
	if config.cachedEnvVars == nil {
		config.cachedEnvVars = make(map[string]string)
	}

	if value, present := config.cachedEnvVars[searchEnvVar]; present {
		config.Debug.logf("Environment variable found cached: %s Value: %s", searchEnvVar, value)
		return value, nil
	}

	// Docker and Buildah produce slightly different output
	// for `inspect` command. Array of maps vs. maps
	var dataBuildah map[string]interface{}
	var dataDocker []map[string]interface{}
	projectConfig, projectConfigErr := getProjectConfig(config)
	if projectConfigErr != nil {
		return "", projectConfigErr
	}
	imageName := projectConfig.Stack
	pullErrs := pullImage(imageName, config)
	if pullErrs != nil {
		return "", pullErrs
	}

	inspectOut, inspectErr := inspectImage(imageName, config)
	if inspectErr != nil {
		return "", inspectErr
	}
	var err error
	var envVars []interface{}
	if config.Buildah {
		err = json.Unmarshal([]byte(inspectOut), &dataBuildah)
		if err != nil {
			return "", errors.New("error unmarshaling data from inspect command - exiting")
		}
		buildahConfig := dataBuildah["config"].(map[string]interface{})
		envVars = buildahConfig["Env"].([]interface{})

	} else {
		err = json.Unmarshal([]byte(inspectOut), &dataDocker)
		if err != nil {
			return "", errors.New("error unmarshaling data from inspect command - exiting")

		}
		dockerConfig := dataDocker[0]["Config"].(map[string]interface{})

		envVars = dockerConfig["Env"].([]interface{})
	}

	config.Debug.log("Number of environment variables in stack image: ", len(envVars))
	config.Debug.log("All environment variables in stack image: ", envVars)
	var varFound = false
	for _, envVar := range envVars {
		nameValuePair := strings.SplitN(envVar.(string), "=", 2)
		name, value := nameValuePair[0], nameValuePair[1]
		config.cachedEnvVars[name] = value
		if name == searchEnvVar {
			varFound = true
		}
	}
	if varFound {
		config.Debug.logf("Environment variable found: %s Value: %s", searchEnvVar, config.cachedEnvVars[searchEnvVar])
		return config.cachedEnvVars[searchEnvVar], nil
	}
	config.Debug.log("Could not find env var: ", searchEnvVar)
	return "", nil
}

func getEnvVarBool(searchEnvVar string, config *RootCommandConfig) (bool, error) {
	strVal, envErr := GetEnvVar(searchEnvVar, config)
	if envErr != nil {
		return false, envErr
	}
	return strings.Compare(strings.TrimSpace(strings.ToUpper(strVal)), "TRUE") == 0, nil
}

func getEnvVarInt(searchEnvVar string, config *RootCommandConfig) (int, error) {

	strVal, envErr := GetEnvVar(searchEnvVar, config)
	if envErr != nil {
		return 0, envErr
	}
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, err
	}
	return intVal, nil

}

func getExtractDir(config *RootCommandConfig) (string, error) {
	extractDir, envErr := GetEnvVar("APPSODY_PROJECT_DIR", config)
	if envErr != nil {
		return "", envErr
	}
	if extractDir == "" {
		config.Warning.log("The stack image does not contain APPSODY_PROJECT_DIR. Using /project")
		return "/project", nil
	}
	return extractDir, nil
}

func getVolumeArgs(config *RootCommandConfig) ([]string, error) {
	volumeArgs := []string{}
	stackMounts, envErr := GetEnvVar("APPSODY_MOUNTS", config)
	if envErr != nil {
		return nil, envErr
	}
	if stackMounts == "" {
		config.Warning.log("The stack image does not contain APPSODY_MOUNTS")
		return volumeArgs, nil
	}
	stackMountList := strings.Split(stackMounts, ";")
	homeDir := UserHomeDir(config.LoggingConfig)
	homeDirOverride := os.Getenv("APPSODY_MOUNT_HOME")
	homeDirOverridden := false
	if homeDirOverride != "" {
		config.Debug.logf("Overriding home mount dir from '%s' to APPSODY_MOUNT_HOME value '%s' ", homeDir, homeDirOverride)
		homeDir = homeDirOverride
		homeDirOverridden = true
	}
	projectDir, perr := getProjectDir(config)
	if perr != nil {
		return volumeArgs, perr

	}
	projectDirOverride := os.Getenv("APPSODY_MOUNT_PROJECT")
	projectDirOverridden := false
	if projectDirOverride != "" {
		config.Debug.logf("Overriding project mount dir from '%s' to APPSODY_MOUNT_PROJECT value '%s' ", projectDir, projectDirOverride)
		projectDir = projectDirOverride
		projectDirOverridden = true
	}

	namedVolumeCount := 0
	for _, mount := range stackMountList {
		if mount == "" {
			continue
		}
		var mappedMount string
		var overridden bool
		if strings.HasPrefix(mount, "~") {
			if homeDirOverride != "" && !strings.ContainsAny(homeDir, "/\\") {
				// home dir was overridden with a named volume (rather than a path)
				namedVolumeCount++
				start := strings.LastIndex(mount, ":")
				mappedMount = fmt.Sprintf("%s%d%s", homeDir, namedVolumeCount, mount[start:])
			} else {
				mappedMount = strings.Replace(mount, "~", homeDir, 1)
			}
			overridden = homeDirOverridden
		} else {
			if strings.HasPrefix(mount, ".:") {
				mount = strings.TrimPrefix(mount, ".")
			}
			mappedMount = filepath.Join(projectDir, mount)
			overridden = projectDirOverridden
		}
		// mappedMount contains local and container (linux) paths. When on windows, the Join above replaces all '/' with '\' and
		// breaks the linux paths. This method is to always use '/' because windows docker tolerates this.
		mappedMount = filepath.ToSlash(mappedMount)

		if !overridden && !mountExistsLocally(config.LoggingConfig, mappedMount) {
			config.Warning.log("Could not mount ", mappedMount, " because the local file was not found.")
			continue
		}
		volumeArgs = append(volumeArgs, "-v", mappedMount)
	}
	config.Debug.log("Mapped mount args: ", volumeArgs)
	return volumeArgs, nil
}

func mountExistsLocally(log *LoggingConfig, mount string) bool {
	localFile := strings.Split(mount, ":")
	if runtime.GOOS == "windows" {
		//Windows may prepend the drive ID to the path
		//ex. C:\whatever\path\:/linux/dir
		if len(localFile) > 2 {
			//This is the case where we have three strings, the first one being
			//the drive ID
			// C: \whatever\path and /linux/dir
			localFile[0] += ":" + localFile[1]
			//We append the second string to the first
			//thus reconstituting the entire local path
		}
	}
	log.Debug.log("Checking for existence of local file or directory to mount: ", localFile[0])
	fileExists, _ := Exists(localFile[0])
	lintMountPathForSingleFile(localFile[0], log)
	return fileExists
}

func getProjectDir(config *RootCommandConfig) (string, error) {
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	projectDir, err := Exists(appsodyConfig)
	if err != nil {
		config.Error.log(err)
		return config.ProjectDir, err
	}
	if !projectDir {
		var e NotAnAppsodyProject = "The current directory is not a valid appsody project. Run `appsody init <stack>` to create one. Run `appsody list` to see the available stacks."
		return config.ProjectDir, &e
	}
	return config.ProjectDir, nil
}

func IsValidProjectName(name string) (bool, error) {
	return isValidParamName(name, "project-name")
}

func IsValidApplicationName(name string) (bool, error) {
	return isValidParamName(name, "application-name")
}

// IsValidParamName tests the given string against Appsody name rules.
// This common set of name rules for Appsody must comply to Kubernetes
// resource name, Kubernetes label value, and Docker container name rules.
// The current rules are:
// 1. Must start with a lowercase letter
// 2. Must contain only lowercase letters, digits, and dashes
// 3. Must end with a letter or digit
// 4. Must be 68 characters or less
func isValidParamName(name, param string) (bool, error) {
	if name == "" {
		return false, errors.Errorf("Invalid %s. The name cannot be an empty string", param)
	}
	if len(name) > 68 {
		return false, errors.Errorf("Invalid %s \"%s\". The name must be 68 characters or less", param, name)
	}

	match, err := regexp.MatchString("^[a-z]([a-z0-9-]*[a-z0-9])?$", name)
	if err != nil {
		return false, err
	}

	if match {
		return true, nil
	}
	return false, errors.Errorf("Invalid %s \"%s\". The name must start with a lowercase letter, contain only lowercase letters, numbers, or dashes, and cannot end in a dash.", param, name)
}

func IsValidKubernetesLabelValue(value string) (bool, error) {
	if value == "" {
		return true, nil
	}
	if len(value) > 63 {
		return false, errors.New("The label must be 63 characters or less")
	}

	match, err := regexp.MatchString("^[a-z0-9A-Z]([a-z0-9A-Z-_.]*[a-z0-9A-Z])?$", value)
	if err != nil {
		return false, err
	}

	if match {
		return true, nil
	}
	return false, errors.Errorf("Invalid label \"%s\". The label must begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between.", value)
}

// ConvertToValidProjectName takes an existing string or directory path
// and returns a name that conforms to isValidContainerName rules
func ConvertToValidProjectName(projectDir string) (string, error) {
	projectName := strings.ToLower(filepath.Base(projectDir))
	valid, _ := IsValidProjectName(projectName)

	if !valid {
		projectName = strings.ToLower(filepath.Base(projectDir))
		if len(projectName) > 68 {
			projectName = projectName[0:68]
		}

		if projectName[0] < 'a' || projectName[0] > 'z' {
			projectName = "appsody-" + projectName
		}

		reg, err := regexp.Compile("[^a-z0-9]+")
		if err != nil {
			return "", err
		}
		projectName = reg.ReplaceAllString(projectName, "-")

		if projectName[len(projectName)-1] == '-' {
			projectName = projectName + "app"
		}

		valid, err := IsValidProjectName(projectName)
		if !valid {
			return projectName, err
		}
	}

	return projectName, nil
}

func getProjectName(config *RootCommandConfig) (string, error) {
	defaultProjectName := "my-project"
	dir, err := getProjectDir(config)
	if err != nil {
		return defaultProjectName, err
	}
	// check to see if project-name is set in .appsody-config.yaml
	projectConfig, err := getProjectConfig(config)
	if err != nil {
		return defaultProjectName, err
	}
	if projectConfig.ProjectName != "" {
		// project-name is in .appsody-config.yaml
		valid, err := IsValidProjectName(projectConfig.ProjectName)
		if !valid {
			return defaultProjectName, err
		}
		return projectConfig.ProjectName, nil
	}
	// project-name is not in .appsody-config.yaml so use the directory name and save
	projectName, err := ConvertToValidProjectName(dir)
	if err != nil {
		return defaultProjectName, err
	}

	err = saveProjectNameToConfig(projectName, config)
	if err != nil {
		config.Warning.Log("Unable to save project name to ", ConfigFile)
	}

	return projectName, nil
}

func GetDeprecated(config *RootCommandConfig) error {
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	stackInfo := v.Get("stack")
	if stackInfo == nil {
		return errors.New("stack information not found in .appsody-config.yaml file")
	}
	stackLabels, err := getStackLabels(config)
	if err != nil {
		return err
	}
	if stackLabels["dev.appsody.stack.deprecated"] != "" {
		config.Info.logf("*\n*\n*\nStack deprecated: %v \n*\n*\n*", stackLabels["dev.appsody.stack.deprecated"])
	}

	return nil
}

func getStackIndexYaml(repoID string, stackID string, config *RootCommandConfig) (*IndexYamlStack, error) {

	if config.Dryrun {
		return nil, nil
	}

	var stackEntry *IndexYamlStack
	extractDir := filepath.Join(getHome(config), "extract")

	// Get Repository directory and unmarshal
	var repoFile RepositoryFile
	source, err := ioutil.ReadFile(getRepoFileLocation(config))
	if err != nil {
		return stackEntry, errors.Errorf("Error trying to read: %v", err)
	}

	err = yaml.Unmarshal(source, &repoFile)
	if err != nil {
		return stackEntry, errors.Errorf("Error parsing the repository.yaml file: %v", err)
	}

	// get specificed repo and unmarshal
	repoEntry := repoFile.GetRepo(repoID)

	// error if repo not found in repository.yaml
	if repoEntry == nil {
		return stackEntry, errors.Errorf("Repository: '%s' was not found in the repository.yaml file", repoID)
	}
	repoEntryURL := repoEntry.URL

	if repoEntryURL == "" {
		return stackEntry, errors.Errorf("URL for specified repository is empty")
	}

	var repoIndex IndexYaml
	tempRepoIndex := filepath.Join(extractDir, "index.yaml")
	err = downloadFileToDisk(config.LoggingConfig, repoEntryURL, tempRepoIndex, config.Dryrun)
	if err != nil {
		return stackEntry, err
	}
	defer os.Remove(tempRepoIndex)
	tempRepoIndexFile, err := ioutil.ReadFile(tempRepoIndex)
	if err != nil {
		return stackEntry, errors.Errorf("Error trying to read: %v", err)
	}

	err = yaml.Unmarshal(tempRepoIndexFile, &repoIndex)
	if err != nil {
		return stackEntry, errors.Errorf("Error parsing the index.yaml file: %v", err)
	}

	// get specified stack and get URL
	stackEntry = getStack(&repoIndex, stackID)
	if stackEntry == nil {
		return stackEntry, errors.New("Could not find stack specified in repository index")
	}

	return stackEntry, nil

}

func getDefaultStackRegistry(config *RootCommandConfig) string {
	defaultStackRegistry := config.CliConfig.Get("images").(string)
	if defaultStackRegistry == "" {
		defaultStackRegistry = "docker.io"
	}
	return defaultStackRegistry
}

func saveProjectNameToConfig(projectName string, config *RootCommandConfig) error {
	valid, err := IsValidProjectName(projectName)
	if !valid {
		return err
	}

	// update the in-memory project name
	projectConfig, err := getProjectConfig(config)
	if err != nil {
		return err
	}
	projectConfig.ProjectName = projectName

	// save the project name to the .appsody-config.yaml
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err = v.ReadInConfig()
	if err != nil {
		return err
	}
	v.Set("project-name", projectName)
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	config.Info.log("Your Appsody project name has been set to ", projectName)
	return nil
}

func saveApplicationNameToConfig(applicationName string, config *RootCommandConfig) error {
	valid, err := IsValidProjectName(applicationName)
	if !valid {
		return err
	}

	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err = v.ReadInConfig()
	if err != nil {
		return err
	}
	v.Set("application-name", applicationName)
	err = v.WriteConfig()
	if err != nil {
		return err
	}

	config.Info.log("Your Appsody application name has been set to ", applicationName)
	return nil
}

func setStackRegistry(stackRegistry string, config *RootCommandConfig) error {

	// Read in the config
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	stackImageName, err := OverrideStackRegistry(stackRegistry, v.Get("stack").(string))
	if err != nil {
		return err
	}
	stackImageName, err = NormalizeImageName(stackImageName)
	if err != nil {
		return err
	}
	v.Set("stack", stackImageName)
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	config.Info.log("Your Appsody project stack has been set to ", stackImageName)
	return nil
}
func getProjectConfigFileContents(config *RootCommandConfig) (*ProjectConfig, error) {

	dir, perr := getProjectDir(config)
	if perr != nil {
		return nil, perr
	}
	appsodyConfig := filepath.Join(dir, ConfigFile)

	v := viper.New()
	v.SetConfigFile(appsodyConfig)

	err := v.ReadInConfig()

	if err != nil {
		return nil, errors.Errorf("Error reading project config %v", err)
	}

	var projectConfig ProjectConfig
	err = v.Unmarshal(&projectConfig)
	if err != nil {
		return &projectConfig, errors.Errorf("Error reading project config %v", err)
	}
	return &projectConfig, nil
}
func getStackRegistryFromConfigFile(config *RootCommandConfig) (string, error) {
	projectConfig, err := getProjectConfigFileContents(config)
	if err != nil {
		return "", err
	}

	if stack := projectConfig.Stack; stack != "" {
		// stack is in .appsody-config.yaml
		stackElements := strings.Split(stack, "/")
		if len(stackElements) == 3 {
			return stackElements[0], nil
		}
		if len(stackElements) < 3 {
			return "", nil
		}
		return "", errors.Errorf("Invalid stack image name detected in project config file: %s", stack)
	}
	return "", errors.New("No stack image name detected in project config file")

}
func getProjectConfig(config *RootCommandConfig) (*ProjectConfig, error) {
	if config.ProjectConfig != nil {
		return config.ProjectConfig, nil
	}
	projectConfig, err := getProjectConfigFileContents(config)
	if err != nil {
		return nil, err
	}
	config.Debug.logf("Project stack before override: %s", projectConfig.Stack)

	imageComponents := strings.Split(projectConfig.Stack, "/")
	if len(imageComponents) < 3 {
		imageRepo := config.CliConfig.GetString("images")
		config.Debug.log("Image repository in the appsody config file: ", imageRepo)
		projectConfig.Stack = imageRepo + "/" + projectConfig.Stack
	}
	//Override the stack registry URL
	projectConfig.Stack, err = OverrideStackRegistry(config.StackRegistry, projectConfig.Stack)
	if err != nil {
		return projectConfig, err
	}

	//Buildah cannot pull from index.docker.io - only pulls from docker.io
	projectConfig.Stack, err = NormalizeImageName(projectConfig.Stack)
	if err != nil {
		return projectConfig, err
	}

	config.Debug.Logf("Project stack after override: %s is: %s", config.StackRegistry, projectConfig.Stack)
	config.ProjectConfig = projectConfig
	return config.ProjectConfig, nil
}

func getOperatorHome(config *RootCommandConfig) string {
	operatorHome := config.CliConfig.GetString("operator")
	config.Debug.log("Operator home set to: ", operatorHome)
	return operatorHome
}

func execAndWait(log *LoggingConfig, command string, args []string, logger appsodylogger, dryrun bool) error {

	return execAndWaitWithWorkDir(log, command, args, logger, workDirNotSet, dryrun)
}
func execAndWaitWithWorkDir(log *LoggingConfig, command string, args []string, logger appsodylogger, workdir string, dryrun bool) error {

	err := execAndWaitWithWorkDirReturnErr(log, command, args, logger, workdir, dryrun)
	if err != nil {
		return errors.Errorf("Error running %s command: %v", command, err)

	}
	return nil

}

// CopyFile uses OS commands to copy a file from a source to a destination
func CopyFile(log *LoggingConfig, source string, dest string) error {
	_, err := os.Stat(source)
	if err != nil {
		log.Error.logf("Cannot find source file %s to copy", source)
		return err
	}

	var execCmd string
	var execArgs = []string{source, dest}

	if runtime.GOOS == "windows" {
		execCmd = "CMD"
		winArgs := []string{"/C", "COPY"}
		execArgs = append(winArgs[0:], execArgs...)

	} else {
		execCmd = "cp"
	}
	copyCmd := exec.Command(execCmd, execArgs...)
	cmdOutput, cmdErr := SeparateOutput(copyCmd)

	if cmdErr != nil {
		log.Error.logf("Could not copy %s to %s - output of copy command %s %s\n", source, dest, cmdOutput, cmdErr)
		return errors.New("Error in copy: " + cmdOutput)
	}

	log.Debug.logf("Copy of %s to %s was successful \n", source, dest)
	return nil
}

// MoveDir moves a directory to another directory, even if they are on different partitions
func MoveDir(log *LoggingConfig, fromDir string, toDir string) error {
	log.Debug.log("Moving ", fromDir, " to ", toDir)
	// Let's try os.Rename first
	err := os.Rename(fromDir, toDir)
	if err == nil {
		// We did it - returning
		return nil
	}
	// If we are here, we need to use copy
	log.Debug.log("os.Rename did not work to move directories... attempting copy. From dir:", fromDir, " target dir: ", toDir)
	err = CopyDir(log, fromDir, toDir)
	if err != nil {
		log.Error.log("Could not move ", fromDir, " to ", toDir)
		return err
	}
	return nil
}

// CopyDir Copies folder from source destination to target destination
func CopyDir(log *LoggingConfig, fromDir string, toDir string) error {
	// fail if fromDir does not exist on the file system
	fromDirExists, err := Exists(fromDir)
	if err != nil {
		return errors.Errorf("Error checking source %v", err)
	}

	if !fromDirExists {
		log.Error.logf("Source %s does not exist.", fromDir)
		return errors.Errorf("Source %s does not exist.", fromDir)
	}

	// fail if toDir exists on the file system
	// toDir should just be the name of the target directory, not an existing file or directory
	// if toDir is an existing file or directory it causes inconsistent results between windows and non-windows
	toDirExists, err := Exists(toDir)
	if err != nil {
		return errors.Errorf("Error checking target %v", err)
	}

	if toDirExists {
		log.Error.logf("Target %s exists. It must only be a name of the target directory for the copy.", toDir)
		return errors.Errorf("Target %s exists. It should only be the name of the target directory for the copy", toDir)
	}

	var execCmd string
	var execArgs = []string{fromDir, toDir}

	if runtime.GOOS == "windows" {
		execCmd = "CMD"
		winArgs := []string{"/C", "XCOPY", "/I", "/E", "/H", "/K"}
		execArgs = append(winArgs[0:], execArgs...)

	} else {
		execCmd = "cp"
		bashArgs := []string{"-rf"}
		execArgs = append(bashArgs[0:], execArgs...)
	}
	log.Debug.log("About to run: ", execCmd, execArgs)
	copyCmd := exec.Command(execCmd, execArgs...)
	cmdOutput, cmdErr := SeparateOutput(copyCmd)
	if cmdErr != nil {
		return errors.Errorf("Could not copy %s to %s: %v", fromDir, toDir, cmdOutput)
	}
	_, err = os.Stat(toDir)
	if err != nil {
		log.Error.logf("Could not copy %s to %s - output of copy command %s %s\n", fromDir, toDir, cmdOutput, cmdErr)
		return errors.New("Error in copy: " + err.Error())
	}
	log.Debug.logf("Directory copy of %s to %s was successful \n", fromDir, toDir)
	return nil
}

// CheckPrereqs checks the prerequisites to run the CLI
func CheckPrereqs(config *RootCommandConfig) error {
	if config.Buildah {
		buildahCmd := "buildah"
		buildahArgs := []string{"containers"}
		checkBuildahCmd := exec.Command(buildahCmd, buildahArgs...)
		_, buildahCmdErr := checkBuildahCmd.Output()
		if buildahCmdErr != nil {
			return errors.Errorf("buildah does not seem to be installed or capable of running in this environment - failed to execute buildah containers: %v", buildahCmdErr)
		}
		return nil
	}
	dockerCmd := "docker"
	dockerArgs := []string{"ps"}
	checkDockerCmd := exec.Command(dockerCmd, dockerArgs...)
	_, cmdErr := checkDockerCmd.Output()
	if cmdErr != nil {
		return errors.New("docker does not seem to be installed or running - failed to execute docker ps")
	}
	return nil
}

// UserHomeDir returns the current user's home directory or '.'
func UserHomeDir(log *LoggingConfig) string {
	homeDir, homeErr := os.UserHomeDir()

	if homeErr != nil {
		log.Error.log("Unable to find user's home directory", homeErr)
		return "."
	}
	return homeDir
}

func getConfigLabels(projectConfig ProjectConfig, filename string, log *LoggingConfig) (map[string]string, error) {
	var labels = make(map[string]string)

	t := time.Now()

	labels[ociKeyPrefix+"created"] = t.Format(time.RFC3339)

	var maintainersString string
	for index, maintainer := range projectConfig.Maintainers {
		maintainersString += maintainer.Name + " <" + maintainer.Email + ">"
		if index < len(projectConfig.Maintainers)-1 {
			maintainersString += ", "
		}
	}

	if maintainersString != "" {
		labels[ociKeyPrefix+"authors"] = maintainersString
	}

	if projectConfig.Version != "" {
		if valid, err := IsValidKubernetesLabelValue(projectConfig.Version); !valid {
			return labels, errors.Errorf("%s version value is invalid. %v", ConfigFile, err)
		}
		labels[ociKeyPrefix+"version"] = projectConfig.Version
	}

	if projectConfig.License != "" {
		if valid, err := IsValidKubernetesLabelValue(projectConfig.License); !valid {
			return labels, errors.Errorf("%s license value is invalid. %v", ConfigFile, err)
		} else if err := checkValidLicense(log, projectConfig.License); err != nil {
			return labels, errors.Errorf("The %v SPDX license ID is invalid: %v.", filename, err)
		}
		labels[ociKeyPrefix+"licenses"] = projectConfig.License
	}

	if projectConfig.ProjectName != "" {
		labels[ociKeyPrefix+"title"] = projectConfig.ProjectName
	}
	if projectConfig.Description != "" {
		labels[ociKeyPrefix+"description"] = projectConfig.Description
	}

	if projectConfig.Stack != "" {
		labels[appsodyStackKeyPrefix+"configured"] = projectConfig.Stack
	}

	if projectConfig.ApplicationName != "" {
		if valid, err := IsValidKubernetesLabelValue(projectConfig.ApplicationName); !valid {
			return labels, errors.Errorf("%s application-name value is invalid. %v", ConfigFile, err)
		}
		labels["dev.appsody.app.name"] = projectConfig.ApplicationName
	}

	return labels, nil
}

func getGitLabels(config *RootCommandConfig) (map[string]string, error) {
	gitInfo, err := GetGitInfo(config)
	if err != nil {
		return nil, err
	}

	var labels = make(map[string]string)

	if gitInfo.RemoteURL != "" {
		labels[ociKeyPrefix+"url"] = gitInfo.RemoteURL
		labels[ociKeyPrefix+"documentation"] = gitInfo.RemoteURL
		labels[ociKeyPrefix+"source"] = gitInfo.RemoteURL + "/tree/" + gitInfo.Branch
		upstreamSplit := strings.Split(gitInfo.Upstream, "/")
		if len(upstreamSplit) > 1 {
			labels[ociKeyPrefix+"source"] = gitInfo.RemoteURL + "/tree/" + upstreamSplit[1]
		}

	}

	var commitInfo = gitInfo.Commit
	revisionKey := ociKeyPrefix + "revision"
	if commitInfo.SHA != "" {
		labels[revisionKey] = commitInfo.SHA
		if gitInfo.ChangesMade {
			labels[revisionKey] += "-modified"
		}
		if !gitInfo.Commit.Pushed {
			labels[revisionKey] += "-not-pushed"
		}
	}

	if commitInfo.Author != "" {
		labels[appsodyImageCommitKeyPrefix+"author"] = commitInfo.Author
	}

	if commitInfo.AuthorEmail != "" {
		labels[appsodyImageCommitKeyPrefix+"author"] += " <" + commitInfo.AuthorEmail + ">"
	}

	if commitInfo.Committer != "" {
		labels[appsodyImageCommitKeyPrefix+"committer"] = commitInfo.Committer
	}

	if commitInfo.CommitterEmail != "" {
		labels[appsodyImageCommitKeyPrefix+"committer"] += " <" + commitInfo.CommitterEmail + ">"
	}

	if commitInfo.Date != "" {
		labels[appsodyImageCommitKeyPrefix+"date"] = commitInfo.Date
	}

	if commitInfo.Message != "" {
		labels[appsodyImageCommitKeyPrefix+"message"] = commitInfo.Message
	}

	if commitInfo.contextDir != "" {
		labels[appsodyImageCommitKeyPrefix+"contextDir"] = commitInfo.contextDir
	}

	return labels, nil
}

func getStackLabels(config *RootCommandConfig) (map[string]string, error) {
	labels := make(map[string]string)
	var data []map[string]interface{}
	var buildahData map[string]interface{}
	var containerConfig map[string]interface{}
	projectConfig, projectConfigErr := getProjectConfig(config)
	if projectConfigErr != nil {
		return nil, projectConfigErr
	}
	imageName := projectConfig.Stack
	pullErrs := pullImage(imageName, config)
	if pullErrs != nil {
		return nil, pullErrs
	}
	inspectOut, err := inspectImage(imageName, config)
	if err != nil {
		return labels, err
	}
	if config.Buildah {
		err = json.Unmarshal([]byte(inspectOut), &buildahData)
		if err != nil {
			return labels, errors.Errorf("Error unmarshaling data from inspect command - exiting %v", err)
		}
		containerConfig = buildahData["config"].(map[string]interface{})
		config.Debug.Log("Config inspected by buildah: ", config)
	} else {
		err := json.Unmarshal([]byte(inspectOut), &data)
		if err != nil {
			return labels, errors.Errorf("Error unmarshaling data from inspect command - exiting %v", err)
		}
		containerConfig = data[0]["Config"].(map[string]interface{})
	}
	if containerConfig["Labels"] != nil {
		labelsMap := containerConfig["Labels"].(map[string]interface{})

		for key, value := range labelsMap {
			labels[key] = value.(string)
		}
	}

	imageAndDigest := data[0]["RepoDigests"].([]interface{})
	if len(imageAndDigest) > 0 { //Check that the image has a digest
		imageAndDigestStr := fmt.Sprintf("%v", imageAndDigest[0])
		digest := strings.Split(imageAndDigestStr, "@")
		labels[appsodyStackKeyPrefix+"digest"] = digest[1]
	}

	return labels, nil
}

func getExposedPorts(config *RootCommandConfig) ([]string, error) {
	// TODO cache this so the docker inspect command only runs once per cli invocation
	var data []map[string]interface{}
	var buildahData map[string]interface{}
	var portValues []string
	var containerConfig map[string]interface{}
	projectConfig, projectConfigErr := getProjectConfig(config)
	if projectConfigErr != nil {
		return nil, projectConfigErr
	}
	imageName := projectConfig.Stack
	pullErrs := pullImage(imageName, config)
	if pullErrs != nil {
		return nil, pullErrs
	}

	inspectOut, inspectErr := inspectImage(imageName, config)
	if inspectErr != nil {
		return portValues, errors.Errorf("Could not inspect the image: %v", inspectErr)
	}
	if config.Buildah {
		err := json.Unmarshal([]byte(inspectOut), &buildahData)
		if err != nil {
			return portValues, errors.Errorf("Error unmarshaling data from inspect command - exiting %v", err)
		}
		containerConfig = buildahData["config"].(map[string]interface{})
		config.Debug.Log("Config inspected by buildah: ", config)
	} else {
		err := json.Unmarshal([]byte(inspectOut), &data)
		if err != nil {
			return portValues, errors.Errorf("Error unmarshaling data from inspect command - exiting %v", err)
		}
		containerConfig = data[0]["Config"].(map[string]interface{})
	}

	if containerConfig["ExposedPorts"] != nil {
		exposedPorts := containerConfig["ExposedPorts"].(map[string]interface{})

		portValues = make([]string, 0, len(exposedPorts))
		for k := range exposedPorts {
			portValues = append(portValues, strings.Split(k, "/tcp")[0])
		}

	}
	return portValues, nil

}

//GenDeploymentYaml generates a simple yaml for a plaing K8S deployment
func GenDeploymentYaml(log *LoggingConfig, appName string, imageName string, controllerImageName string, ports []string, pdir string, dockerMounts []string, dockerEnvVars map[string]string, depsMount string, dryrun bool) (fileName string, err error) {

	// Codewind workspace root dir constant
	codeWindWorkspace := "/"
	// Codewind project ID if provided
	codeWindProjectID := os.Getenv("CODEWIND_PROJECT_ID")
	// Codewind onwner ref name and uid
	codeWindOwnerRefName := os.Getenv("CODEWIND_OWNER_NAME")
	codeWindOwnerRefUID := os.Getenv("CODEWIND_OWNER_UID")
	// Deployment YAML structs
	type Port struct {
		Name          string `yaml:"name,omitempty"`
		ContainerPort int    `yaml:"containerPort"`
	}
	type EnvVar struct {
		Name  string `yaml:"name,omitempty"`
		Value string `yaml:"value,omitempty"`
	}
	type VolumeMount struct {
		Name      string `yaml:"name"`
		MountPath string `yaml:"mountPath"`
		SubPath   string `yaml:"subPath,omitempty"`
	}
	type Container struct {
		Args            []string  `yaml:"args,omitempty"`
		Command         []string  `yaml:"command,omitempty"`
		Env             []*EnvVar `yaml:"env,omitempty"`
		Image           string    `yaml:"image"`
		ImagePullPolicy string    `yaml:"imagePullPolicy,omitempty"`
		Name            string    `yaml:"name,omitempty"`
		Ports           []*Port   `yaml:"ports,omitempty"`
		SecurityContext struct {
			Privileged bool `yaml:"privileged"`
		} `yaml:"securityContext,omitempty"`
		VolumeMounts []VolumeMount `yaml:"volumeMounts"`
		WorkingDir   string        `yaml:"workingDir,omitempty"`
	}
	type InitContainer struct {
		Args            []string  `yaml:"args,omitempty"`
		Command         []string  `yaml:"command,omitempty"`
		Env             []*EnvVar `yaml:"env,omitempty"`
		Image           string    `yaml:"image"`
		ImagePullPolicy string    `yaml:"imagePullPolicy,omitempty"`
		Name            string    `yaml:"name,omitempty"`
		Ports           []*Port   `yaml:"ports,omitempty"`
		SecurityContext struct {
			Privileged bool `yaml:"privileged"`
		} `yaml:"securityContext,omitempty"`
		VolumeMounts []VolumeMount `yaml:"volumeMounts"`
		WorkingDir   string        `yaml:"workingDir,omitempty"`
	}
	type Volume struct {
		Name                  string `yaml:"name"`
		PersistentVolumeClaim struct {
			ClaimName string `yaml:"claimName"`
		} `yaml:"persistentVolumeClaim,omitempty"`
		EmptyDir struct {
			Medium string `yaml:"medium"`
		} `yaml:"emptyDir,omitempty"`
	}

	type Deployment struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name            string            `yaml:"name"`
			Namespace       string            `yaml:"namespace,omitempty"`
			Labels          map[string]string `yaml:"labels,omitempty"`
			OwnerReferences []OwnerReference  `yaml:"ownerReferences,omitempty"`
		} `yaml:"metadata"`
		Spec struct {
			Selector struct {
				MatchLabels map[string]string `yaml:"matchLabels"`
			} `yaml:"selector"`
			Replicas    int `yaml:"replicas"`
			PodTemplate struct {
				Metadata struct {
					Labels map[string]string `yaml:"labels"`
				} `yaml:"metadata"`
				Spec struct {
					ServiceAccountName string           `yaml:"serviceAccountName,omitempty"`
					InitContainers     []*InitContainer `yaml:"initContainers"`
					Containers         []*Container     `yaml:"containers"`
					Volumes            []*Volume        `yaml:"volumes"`
				} `yaml:"spec"`
			} `yaml:"template"`
		} `yaml:"spec"`
	}

	yamlMap := Deployment{}
	yamlTemplate := getDeploymentTemplate()
	err = yaml.Unmarshal([]byte(yamlTemplate), &yamlMap)
	if err != nil {
		log.Error.log("Could not create the YAML structure from template. Exiting.")
		return "", err
	}
	//Set the name
	yamlMap.Metadata.Name = appName

	//Set the codewind label if present
	if codeWindProjectID != "" {
		yamlMap.Metadata.Labels = make(map[string]string)
		yamlMap.Metadata.Labels["projectID"] = codeWindProjectID
	}
	//Set the owner ref if present
	if codeWindOwnerRefName != "" && codeWindOwnerRefUID != "" {
		yamlMap.Metadata.OwnerReferences = []OwnerReference{
			{
				APIVersion:         "apps/v1",
				BlockOwnerDeletion: true,
				Controller:         true,
				Kind:               "ReplicaSet",
				Name:               codeWindOwnerRefName,
				UID:                codeWindOwnerRefUID},
		}
	}
	//Set the service account if provided by an env var
	serviceAccount := os.Getenv("SERVICE_ACCOUNT_NAME")
	if serviceAccount != "" {
		log.Debug.Log("Detected service account name env var: ", serviceAccount)
		yamlMap.Spec.PodTemplate.Spec.ServiceAccountName = serviceAccount
	} else {
		log.Debug.log("No service account name env var, leaving the appsody-sa default")
	}
	//Set the controller image
	yamlMap.Spec.PodTemplate.Spec.InitContainers[0].Image = controllerImageName
	//Set the image
	yamlMap.Spec.PodTemplate.Spec.Containers[0].Name = appName
	yamlMap.Spec.PodTemplate.Spec.Containers[0].Image = imageName

	//Set the containerPort
	containerPorts := make([]*Port, 0)
	for i, port := range ports {
		//KNative only allows a single port entry
		if i == 0 {
			yamlMap.Spec.PodTemplate.Spec.Containers[0].Ports = containerPorts
		}
		log.Debug.Log("Adding port to yaml: ", port)
		newContainerPort := new(Port)
		newContainerPort.ContainerPort, err = strconv.Atoi(port)
		if err != nil {
			return "", err
		}
		yamlMap.Spec.PodTemplate.Spec.Containers[0].Ports = append(yamlMap.Spec.PodTemplate.Spec.Containers[0].Ports, newContainerPort)
	}
	//Set the env vars from docker run, if any
	if len(dockerEnvVars) > 0 {
		envVars := make([]*EnvVar, len(dockerEnvVars))
		idx := 0
		for key, value := range dockerEnvVars {
			envVars[idx] = &EnvVar{key, value}
			idx++
		}
		yamlMap.Spec.PodTemplate.Spec.Containers[0].Env = envVars
	}
	//Set the Pod release label to the container name
	yamlMap.Spec.PodTemplate.Metadata.Labels["release"] = appName
	//Set the workspace volume PVC
	workspaceVolumeName := "appsody-workspace"
	workspacePvcName := os.Getenv("PVC_NAME")
	if workspacePvcName == "" {
		workspacePvcName = "appsody-workspace"
	}

	workspaceVolume := Volume{Name: workspaceVolumeName}
	workspaceVolume.PersistentVolumeClaim.ClaimName = workspacePvcName
	volumeIdx := len(yamlMap.Spec.PodTemplate.Spec.Volumes)
	if volumeIdx < 1 {
		yamlMap.Spec.PodTemplate.Spec.Volumes = make([]*Volume, 1)
		yamlMap.Spec.PodTemplate.Spec.Volumes[0] = &workspaceVolume
	} else {
		yamlMap.Spec.PodTemplate.Spec.Volumes = append(yamlMap.Spec.PodTemplate.Spec.Volumes, &workspaceVolume)
	}
	//Set the code mounts
	//We need to iterate through the docker mounts
	volumeMounts := &yamlMap.Spec.PodTemplate.Spec.Containers[0].VolumeMounts
	for _, appsodyMount := range dockerMounts {
		if appsodyMount == "-v" {
			continue
		}
		appsodyMountComponents := strings.Split(appsodyMount, ":")
		targetMount := appsodyMountComponents[1]
		sourceMount, err := filepath.Rel(codeWindWorkspace, appsodyMountComponents[0])
		if err != nil {
			log.Debug.Log("Problem with the appsody mount: ", appsodyMountComponents[0])
			return "", err
		}

		sourceSubpath := filepath.Join(".", sourceMount)
		newVolumeMount := VolumeMount{"appsody-workspace", targetMount, sourceSubpath}
		log.Debug.Log("Appending volume mount: ", newVolumeMount)
		*volumeMounts = append(*volumeMounts, newVolumeMount)
	}

	//Set the deployment selector and pod label
	projectLabel := appName
	yamlMap.Spec.Selector.MatchLabels["app"] = projectLabel
	yamlMap.Spec.PodTemplate.Metadata.Labels["app"] = projectLabel

	log.Debug.logf("YAML map: \n%v\n", yamlMap)
	yamlStr, err := yaml.Marshal(&yamlMap)
	if err != nil {
		log.Error.log("Could not create the YAML string from Map. Exiting.")
		return "", err
	}
	log.Debug.logf("Generated YAML: \n%s\n", yamlStr)
	// Generate file based on supplied config, defaulting to app-deploy.yaml
	yamlFile := filepath.Join(pdir, "app-deploy.yaml")
	if dryrun {
		log.Info.log("Skipping creation of yaml file with prefix: ", yamlFile)
		return yamlFile, nil
	}
	err = ioutil.WriteFile(yamlFile, yamlStr, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not create the yaml file for deployment %v", err)
	}
	return yamlFile, nil
}
func getDeploymentTemplate() string {
	yamltempl := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: APPSODY_APP_NAME
spec:
  selector:
    matchLabels:
      app: appsody
  replicas: 1
  template:
    metadata:
      labels:
        app: appsody
    spec:
      serviceAccountName: appsody-sa
      initContainers:
      - name: init-appsody-controller
        image: appsody/appsody-controller
        resources: {}
        volumeMounts:
        - name: appsody-controller
          mountPath: /.appsody
        imagePullPolicy: IfNotPresent 
      containers:
      - name: APPSODY_APP_NAME
        image: APPSODY_STACK
        imagePullPolicy: Always
        command: ["/.appsody/appsody-controller"]
        volumeMounts:
        - name: appsody-controller
          mountPath: /.appsody
      volumes:
      - name: appsody-controller
        emptyDir: {}
`
	return yamltempl
}

//GenServiceYaml returns the file name of a generated K8S Service yaml
func GenServiceYaml(log *LoggingConfig, appName string, ports []string, pdir string, dryrun bool) (fileName string, err error) {

	type Port struct {
		Name       string `yaml:"name,omitempty"`
		Port       int    `yaml:"port"`
		TargetPort int    `yaml:"targetPort"`
	}
	type Service struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name            string            `yaml:"name"`
			Labels          map[string]string `yaml:"labels,omitempty"`
			OwnerReferences []OwnerReference  `yaml:"ownerReferences,omitempty"`
		} `yaml:"metadata"`
		Spec struct {
			Selector    map[string]string `yaml:"selector"`
			ServiceType string            `yaml:"type"`
			Ports       []Port            `yaml:"ports"`
		} `yaml:"spec"`
	}

	// Codewind project ID if provided
	codeWindProjectID := os.Getenv("CODEWIND_PROJECT_ID")
	// Codewind onwner ref name and uid
	codeWindOwnerRefName := os.Getenv("CODEWIND_OWNER_NAME")
	codeWindOwnerRefUID := os.Getenv("CODEWIND_OWNER_UID")

	var service Service

	service.APIVersion = "v1"
	service.Kind = "Service"
	service.Metadata.Name = fmt.Sprintf("%s-%s", appName, "service")

	//Set the release and projectID labels
	service.Metadata.Labels = make(map[string]string)
	service.Metadata.Labels["release"] = appName
	if codeWindProjectID != "" {
		service.Metadata.Labels["projectID"] = codeWindProjectID
	}

	//Set the owner ref if present
	if codeWindOwnerRefName != "" && codeWindOwnerRefUID != "" {
		service.Metadata.OwnerReferences = []OwnerReference{
			{
				APIVersion:         "apps/v1",
				BlockOwnerDeletion: true,
				Controller:         true,
				Kind:               "ReplicaSet",
				Name:               codeWindOwnerRefName,
				UID:                codeWindOwnerRefUID},
		}
	}

	service.Spec.Selector = make(map[string]string, 1)
	service.Spec.Selector["app"] = appName
	service.Spec.ServiceType = "NodePort"
	service.Spec.Ports = make([]Port, len(ports))
	for i, port := range ports {
		service.Spec.Ports[i].Name = fmt.Sprintf("port-%d", i)
		iPort, err := strconv.Atoi(port)
		if err != nil {
			return "", err
		}
		service.Spec.Ports[i].Port = iPort
		service.Spec.Ports[i].TargetPort = iPort
	}

	yamlStr, err := yaml.Marshal(&service)
	if err != nil {
		log.Error.log("Could not create the YAML string from Map. Exiting.")
		return "", err
	}
	log.Debug.logf("Generated YAML: \n%s\n", yamlStr)
	// Generate file based on supplied config, defaulting to app-deploy.yaml
	yamlFile := filepath.Join(pdir, "app-service.yaml")
	if dryrun {
		log.Info.log("Skipping creation of yaml file with prefix: ", yamlFile)
		return yamlFile, nil
	}
	err = ioutil.WriteFile(yamlFile, yamlStr, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not create the yaml file for the service %v", err)
	}
	return yamlFile, nil
}

//GenRouteYaml returns the file name of a generated K8S Service yaml
func GenRouteYaml(log *LoggingConfig, appName string, pdir string, port int, dryrun bool) (fileName string, err error) {
	type IngressPath struct {
		Path    string `yaml:"path"`
		Backend struct {
			ServiceName string `yaml:"serviceName"`
			ServicePort int    `yaml:"servicePort"`
		} `yaml:"backend"`
	}
	type IngressRule struct {
		Host string `yaml:"host"`
		HTTP struct {
			Paths []IngressPath `yaml:"paths"`
		} `yaml:"http"`
	}

	type Ingress struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec struct {
			Rules []IngressRule `yaml:"rules"`
		} `yaml:"spec"`
	}

	var ingress Ingress
	ingress.APIVersion = "extensions/v1beta1"
	ingress.Kind = "Ingress"
	ingress.Metadata.Name = fmt.Sprintf("%s-%s", appName, "ingress")

	ingress.Spec.Rules = make([]IngressRule, 1)
	//cheIngressHost := os.Getenv("CHE_INGRESS_HOST")
	//Ignore the CW variable for now
	ingressHost := ""
	if ingressHost != "" {
		ingress.Spec.Rules[0].Host = ingressHost
	} else {
		// We set it to a host name that's resolvable by nip.io
		ingress.Spec.Rules[0].Host = fmt.Sprintf("%s.%s.%s", appName, getK8sMasterIP(log, dryrun), "nip.io")
	}

	ingress.Spec.Rules[0].HTTP.Paths = make([]IngressPath, 1)
	ingress.Spec.Rules[0].HTTP.Paths[0].Path = "/"
	ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName = fmt.Sprintf("%s-%s", appName, "service")
	ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort = port

	yamlStr, err := yaml.Marshal(&ingress)
	if err != nil {
		log.Error.log("Could not create the YAML string from Map. Exiting.")
		return "", err
	}
	log.Debug.logf("Generated YAML: \n%s\n", yamlStr)
	// Generate file based on supplied config, defaulting to app-deploy.yaml
	yamlFile := filepath.Join(pdir, "app-ingress.yaml")
	if dryrun {
		log.Info.log("Skipping creation of yaml file with prefix: ", yamlFile)
		return yamlFile, nil
	}
	err = ioutil.WriteFile(yamlFile, yamlStr, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not create the yaml file for the route %v", err)
	}
	return yamlFile, nil
}

func getK8sMasterIP(log *LoggingConfig, dryrun bool) string {
	cmdParms := []string{"node", "--selector", "node-role.kubernetes.io/master", "-o", "jsonpath={.items[0].status.addresses[?(.type==\"InternalIP\")].address}"}
	ip, err := KubeGet(log, cmdParms, "", dryrun)
	if err == nil {
		return ip
	}
	log.Debug.log("Could not retrieve the master IP address - returning x.x.x.x: ", err)
	return "x.x.x.x"
}

func getIngressPort(config *RootCommandConfig) int {
	ports, err := getExposedPorts(config)

	knownHTTPPorts := []string{"80", "8080", "8008", "3000", "9080"}
	if err != nil {
		config.Debug.Log("Error trying to obtain the exposed ports: ", err)
		return 0
	}
	if len(ports) < 1 {
		config.Debug.log("Container doesn't expose any port - returning 0")
		return 0
	}
	iPort := 0
	for _, port := range ports {
		for _, knownPort := range knownHTTPPorts {
			if port == knownPort {
				iPort, err := strconv.Atoi(port)
				if err == nil {
					return iPort
				}
			}
		}
	}
	//If we haven't returned yet, there was no match
	//Pick the first port and return it
	config.Debug.Log("No known HTTP port detected, returning the first one on the list.")
	iPort, err = strconv.Atoi(ports[0])
	if err == nil {
		return iPort
	}
	config.Debug.Logf("Error converting port %s - returning 0: %v", ports[0], err)
	return 0
}

//ImagePush pushes a docker image to a docker registry (assumes that the user has done docker login)
func ImagePush(log *LoggingConfig, imageToPush string, buildah bool, dryrun bool) error {
	log.Info.log("Pushing image ", imageToPush)
	cmdName := "docker"
	if buildah {
		cmdName = "buildah"
	}

	cmdArgs := []string{"push", imageToPush}
	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", cmdName, " ", strings.Join(cmdArgs, " "))
		return nil
	}

	pushCmd := exec.Command(cmdName, cmdArgs...)

	pushOut, pushErr := SeparateOutput(pushCmd)
	if pushErr != nil {
		if !(strings.Contains(pushErr.Error(), "[DEPRECATION NOTICE] registry v2") || strings.Contains(string(pushOut[:]), "[DEPRECATION NOTICE] registry v2")) {
			log.Error.log("Could not push the image: ", pushErr, " ", string(pushOut[:]))

			return errors.New("Error in pushing image: " + pushOut)
		}
		return errors.New("Error in pushing image: " + pushOut)

	}
	return pushErr
}

// DockerRunBashCmd issues a shell command in a docker image, overriding its entrypoint
// Assume this is only used for Stack images
func DockerRunBashCmd(options []string, image string, bashCmd string, config *RootCommandConfig) (string, error) {
	cmdName := "docker"
	var cmdArgs []string
	pullErrs := pullImage(image, config)
	if pullErrs != nil {
		return "", pullErrs
	}
	if len(options) >= 0 {
		cmdArgs = append([]string{"run"}, options...)
	} else {
		cmdArgs = []string{"run"}
	}

	cmdArgs = append(cmdArgs, "--entrypoint", "/bin/bash", image, "-c", bashCmd)
	config.Info.log("Running command: ", cmdName, " ", ArgsToString(cmdArgs))
	dockerCmd := exec.Command(cmdName, cmdArgs...)

	kout, kerr := SeparateOutput(dockerCmd)
	if kerr != nil {
		return kout, kerr
	}
	return strings.TrimSpace(string(kout[:])), nil
}

//KubeGet issues kubectl get <arg>
func KubeGet(log *LoggingConfig, args []string, namespace string, dryrun bool) (string, error) {
	log.Info.log("Attempting to get resource from Kubernetes ...")
	kcmd := "kubectl"
	kargs := []string{"get"}
	kargs = append(kargs, args...)
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return "", nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return "", errors.Errorf("kubectl get failed: %s", kout)
	}
	return kout, kerr
}

//KubeApply issues kubectl apply -f <filename>
func KubeApply(log *LoggingConfig, fileToApply string, namespace string, dryrun bool) error {
	log.Info.log("Attempting to apply resource in Kubernetes ...")
	kcmd := "kubectl"
	kargs := []string{"apply", "-f", fileToApply}
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return errors.Errorf("kubectl apply failed: %s", kout)
	}
	log.Debug.log("kubectl apply success: ", string(kout[:]))
	return kerr
}

//KubeDelete issues kubectl delete -f <filename>
func KubeDelete(log *LoggingConfig, fileToApply string, namespace string, dryrun bool) error {
	log.Info.log("Attempting to delete resource from Kubernetes...")
	kcmd := "kubectl"
	kargs := []string{"delete", "-f", fileToApply}
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)

	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return errors.Errorf("kubectl delete failed: %s", kout)
	}
	log.Debug.log("kubectl delete success: ", kout)
	return kerr
}

//KubeGetNodePortURLIBMCloud issues several kubectl commands and prints the concatenated URL
func KubeGetNodePortURLIBMCloud(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kargs := []string{"pod"}
	kargs = append(kargs, "-l", "app.kubernetes.io/name="+service, "-o", "jsonpath={.items[].spec.nodeName}")
	nodeName, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil {
		return "", errors.Errorf("Failed to find nodeName for deployed service: %s", err)
	}

	kargs = append([]string{"node"}, nodeName)
	kargs = append(kargs, "-o", "jsonpath=http://{.status.addresses[?(@.type=='ExternalIP')].address}")
	hostURL, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil {
		return "", errors.Errorf("Failed to find deployed service IP and Port: %s", err)
	}
	kargs = append([]string{"svc"}, service)
	kargs = append(kargs, "-o", "jsonpath="+hostURL+":{.spec.ports[0].nodePort}")
	out, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil {
		return "", errors.Errorf("Failed to find deployed service IP and Port: %s", err)
	}
	return out, nil
}

//KubeGetClusterURL kubectl get svc <service> -o jsonpath=http://{.spec.clusterIP}:{.spec.ports[0].port} and prints the return URL
func KubeGetClusterURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kargs := append([]string{"svc"}, service)
	kargs = append(kargs, "-o", "jsonpath=http://{.spec.clusterIP}:{.spec.ports[0].port}")
	out, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil {
		return "", errors.Errorf("Failed to find deployed service IP and Port: %s", err)
	}
	out = out + "\nHowever, as the ServiceType was specified as ClusterIP this url is only accessible to other applications in the same cluster." +
		"\nTo access it try using 'kubectl port-forward', or exposing it further by 'oc expose'"
	return out, nil
}

//KubeGetNodePortURL kubectl get svc <service> -o jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort} and prints the return URL
func KubeGetNodePortURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kargs := append([]string{"svc"}, service)
	kargs = append(kargs, "-o", "jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort}")
	out, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil || strings.Contains(out, "://:") {
		return "", errors.Errorf("Failed to find deployed service IP and Port: %s", err)
	}
	return out, nil
}

//KubeGetRouteURL issues kubectl get svc <service> -o jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort} and prints the return URL
func KubeGetRouteURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kargs := append([]string{"route"}, service)
	kargs = append(kargs, "-o", "jsonpath={.status.ingress[0].host}")
	out, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil {
		return "", errors.Errorf("Failed to find deployed service IP and Port: %s", err)
	}
	return out, nil
}

//KubeGetKnativeURL issues kubectl get rt <service> -o jsonpath="{.status.url}" and prints the return URL
func KubeGetKnativeURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kcmd := "kubectl"
	kargs := append([]string{"get", "rt"}, service)
	kargs = append(kargs, "-o", "jsonpath=\"{.status.url}\"")
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return "", nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return "", errors.Errorf("kubectl get failed: %s", kout)
	}
	return kout, kerr
}

//KubeGetDeploymentURL searches for an exposed hostname and port for the deployed service
func KubeGetDeploymentURL(log *LoggingConfig, serviceName string, service map[string]interface{}, namespace string, dryrun bool) (url string, err error) {
	serviceType := ""
	if service != nil {
		serviceType = service["type"].(string)
	}
	if serviceType == "ClusterIP" {
		// We have a ClusterIP type
		url, err = KubeGetClusterURL(log, serviceName, namespace, dryrun)
		if err == nil {
			return url, nil
		}
	} else {
		url, err = KubeGetKnativeURL(log, serviceName, namespace, dryrun)
		if err == nil {
			return url, nil
		}
		url, err = KubeGetRouteURL(log, serviceName, namespace, dryrun)
		if err == nil {
			return url, nil
		}
		url, err = KubeGetNodePortURL(log, serviceName, namespace, dryrun)
		if err == nil {
			return url, nil
		}
		url, err = KubeGetNodePortURLIBMCloud(log, serviceName, namespace, dryrun)
		if err == nil {
			return url, nil
		}
	}
	log.Error.log("Failed to get deployment hostname and port: ", err)
	return "", err
}

//pullCmd
// enable extract to use `buildah` sequences for image extraction.
// Pull the given docker image
func pullCmd(log *LoggingConfig, imageToPull string, buildah bool, dryrun bool) error {
	cmdName := "docker"
	if buildah {
		cmdName = "buildah"
	}
	pullArgs := []string{"pull", imageToPull}
	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", cmdName, " ", strings.Join(pullArgs, " "))
		return nil
	}
	log.Info.log("Pulling docker image ", imageToPull)
	err := execAndWaitReturnErr(log, cmdName, pullArgs, log.Info, dryrun)
	if err != nil {
		log.Warning.log("Docker image pull failed: ", err)
		return err
	}
	return nil
}

func checkImageExistsLocally(log *LoggingConfig, imageToPull string, buildah bool) bool {

	var cmdName string
	if buildah {
		cmdName = "buildah"
	} else {
		cmdName = "docker"
	}
	imageNameComponents := strings.Split(imageToPull, "/")
	if len(imageNameComponents) == 3 {
		if imageNameComponents[0] == "index.docker.io" || imageNameComponents[0] == "docker.io" {
			imageToPull = fmt.Sprintf("%s/%s", imageNameComponents[1], imageNameComponents[2])
		}
	}

	cmdArgs := []string{"image", "ls", "-q", imageToPull}
	imagelsCmd := exec.Command(cmdName, cmdArgs...)
	imagelsOut, imagelsErr := SeparateOutput(imagelsCmd)
	log.Debug.log("Docker image ls command output: ", imagelsOut)

	if imagelsErr != nil {
		log.Warning.log("Could not run docker image ls -q for the image: ", imageToPull, " error: ", imagelsErr, " Check to make sure docker is available.")
		return false
	}
	if imagelsOut != "" {
		return true
	}
	return false
}

//pullImage
// pulls buildah / docker image, if APPSODY_PULL_POLICY set to IFNOTPRESENT
//it checks for image in local repo and pulls if not in the repo

func pullImage(imageToPull string, config *RootCommandConfig) error {
	if config.imagePulled == nil {
		config.imagePulled = make(map[string]bool)
	}

	config.Debug.logf("%s image pulled status: %t", imageToPull, config.imagePulled[imageToPull])
	if config.imagePulled[imageToPull] {
		config.Debug.log("Image has been pulled already: ", imageToPull)
		return nil
	}
	config.imagePulled[imageToPull] = true

	localImageFound := false
	pullPolicyAlways := true
	pullPolicy := os.Getenv("APPSODY_PULL_POLICY") // Always or IfNotPresent

	// for local stack development path such as stack validate, stack create, ...
	if strings.Contains(imageToPull, "dev.local/") {
		pullPolicy = "IFNOTPRESENT"
	}

	if pullPolicy == "" || strings.ToUpper(pullPolicy) == "ALWAYS" {
		config.Debug.log("Pull policy Always")
	} else if strings.ToUpper(pullPolicy) == "IFNOTPRESENT" {
		config.Debug.log("Pull policy IfNotPresent, checking for local image")
		pullPolicyAlways = false
	}
	if !pullPolicyAlways {
		localImageFound = checkImageExistsLocally(config.LoggingConfig, imageToPull, config.Buildah)
	}

	if pullPolicyAlways || (!pullPolicyAlways && !localImageFound) {
		err := pullCmd(config.LoggingConfig, imageToPull, config.Buildah, config.Dryrun)
		if err != nil {
			if pullPolicyAlways {
				localImageFound = checkImageExistsLocally(config.LoggingConfig, imageToPull, config.Buildah)
			}
			if !localImageFound {
				return errors.Errorf("Could not find the image either in docker hub or locally: %s", imageToPull)

			}
		}
	}
	if localImageFound {
		config.Info.log("Using local cache for image ", imageToPull)
	}
	return nil
}

func inspectImage(imageToInspect string, config *RootCommandConfig) (string, error) {

	cmdName := "docker"
	cmdArgs := []string{"image", "inspect", imageToInspect}
	if config.Buildah {
		cmdName = "buildah"
		cmdArgs = []string{"inspect", "--format={{.Config}}", imageToInspect}
	}

	inspectCmd := exec.Command(cmdName, cmdArgs...)
	inspectOut, inspectErr := SeparateOutput(inspectCmd)
	if inspectErr != nil {
		return "", errors.Errorf("Could not inspect the image: %s", inspectOut)
	}
	return inspectOut, nil
}

//OverrideStackRegistry allows you to change the image registry URL
func OverrideStackRegistry(override string, imageName string) (string, error) {
	if override == "" {
		return imageName, nil
	}
	match, err := ValidateHostNameAndPort(override)
	if err != nil {
		return "", err
	}
	if !match {
		return "", errors.Errorf("This is an invalid host name: %s", override)
	}
	imageNameComponents := strings.Split(imageName, "/")
	if len(imageNameComponents) == 3 {
		imageNameComponents[0] = override
	}
	if len(imageNameComponents) == 2 || len(imageNameComponents) == 1 {
		newComponent := []string{override}
		imageNameComponents = append(newComponent, imageNameComponents...)
	}
	if len(imageNameComponents) > 3 {
		return "", errors.Errorf("Image name is invalid and needs to be changed in the project config file (.appsody-config.yaml): %s. Too many slashes (/) - the override cannot take place.", imageName)
	}
	return strings.Join(imageNameComponents, "/"), nil
}

//ValidateHostNameAndPort validates that hostNameAndPort conform to the DNS naming conventions
func ValidateHostNameAndPort(hostNameAndPort string) (bool, error) {
	match, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])($|:[0-9]{1,5}$)`, hostNameAndPort)
	return match, err
}

//NormalizeImageName is a temporary fix for buildah workaround #676
func NormalizeImageName(imageName string) (string, error) {
	imageNameComponents := strings.Split(imageName, "/")
	if len(imageNameComponents) == 2 {
		return imageName, nil
	}

	if len(imageNameComponents) == 1 {
		return fmt.Sprintf("docker.io/%s", imageName), nil
	}

	if len(imageNameComponents) == 3 {
		if imageNameComponents[0] == "index.docker.io" {
			imageNameComponents[0] = "docker.io"
			return strings.Join(imageNameComponents, "/"), nil
		}
		return imageName, nil
	}
	return imageName, errors.Errorf("Image name is invalid: %s", imageName)

}

func execAndListenWithWorkDirReturnErr(log *LoggingConfig, command string, args []string, logger appsodylogger, workdir string, dryrun bool) (*exec.Cmd, error) {
	var execCmd *exec.Cmd
	var err error
	if dryrun {
		log.Info.log("Dry Run - Skipping command: ", command, " ", ArgsToString(args))
	} else {
		log.Info.log("Running command: ", command, " ", ArgsToString(args))
		execCmd = exec.Command(command, args...)
		if workdir != "" {
			execCmd.Dir = workdir
		}
		cmdReader, err := execCmd.StdoutPipe()
		if err != nil {
			log.Error.log("Error creating StdoutPipe for Cmd ", err)
			return nil, err
		}

		errReader, err := execCmd.StderrPipe()
		if err != nil {
			log.Error.log("Error creating StderrPipe for Cmd ", err)
			return nil, err
		}

		outScanner := bufio.NewScanner(cmdReader)
		outScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		go func() {
			for outScanner.Scan() {
				logger.log(outScanner.Text())
			}
		}()

		errScanner := bufio.NewScanner(errReader)
		errScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		go func() {
			for errScanner.Scan() {
				logger.log(errScanner.Text())
			}
		}()

		err = execCmd.Start()
		if err != nil {
			log.Debug.log("Error running ", command, " command: ", errScanner.Text(), err)
			return nil, err
		}
	}
	return execCmd, err
}

func execAndWaitReturnErr(log *LoggingConfig, command string, args []string, logger appsodylogger, dryrun bool) error {
	return execAndWaitWithWorkDirReturnErr(log, command, args, logger, "", dryrun)
}

func execAndWaitWithWorkDirReturnErr(log *LoggingConfig, command string, args []string, logger appsodylogger, workdir string, dryrun bool) error {
	var err error
	var execCmd *exec.Cmd
	if dryrun {
		log.Info.log("Dry Run - Skipping command: ", command, " ", strings.Join(args, " "))
	} else {
		execCmd, err = execAndListenWithWorkDirReturnErr(log, command, args, logger, workdir, dryrun)
		if err != nil {
			return err
		}
		err = execCmd.Wait()
		if err != nil {
			return err
		}
	}
	return err
}

func getLatestVersion(log *LoggingConfig) string {
	var version string
	log.Debug.log("Getting latest version ", LatestVersionURL)
	resp, err := http.Get(LatestVersionURL)
	if err != nil {
		log.Warning.log("Unable to check the most recent version of Appsody in GitHub.... continuing.")
	} else {
		url := resp.Request.URL.String()
		r, _ := regexp.Compile(`[\d]+\.[\d]+\.[\d]+$`)

		version = r.FindString(url)
	}
	log.Debug.log("Got version ", version)
	return version
}

func doVersionCheck(config *RootCommandConfig) {
	var latest = getLatestVersion(config.LoggingConfig)
	var currentTime = time.Now().Format("2006-01-02 15:04:05 -0700 MST")
	if latest != "" && VERSION != "vlatest" && VERSION != latest {
		updateString := GetUpdateString(runtime.GOOS, VERSION, latest)
		config.Warning.logf(updateString)
	}

	config.CliConfig.Set("lastversioncheck", currentTime)
	if err := config.CliConfig.WriteConfig(); err != nil {
		config.Error.logf("Writing default config file %s", err)

	}
}

// GetUpdateString Returns a format string to advise the user how to upgrade
func GetUpdateString(osName string, version string, latest string) string {
	var updateString string
	switch osName {
	case "darwin":
		updateString = "Please run `brew upgrade appsody` to upgrade"
	default:
		updateString = "Please go to https://appsody.dev/docs/getting-started/installation#upgrading-appsody and upgrade"
	}
	return fmt.Sprintf("\n*\n*\n*\n\nA new CLI update is available.\n%s from %s --> %s.\n\n*\n*\n*\n", updateString, version, latest)
}

func getLastCheckTime(config *RootCommandConfig) string {
	return config.CliConfig.GetString("lastversioncheck")
}

func checkTime(config *RootCommandConfig) {
	var lastCheckTime = getLastCheckTime(config)

	lastTime, err := time.Parse("2006-01-02 15:04:05 -0700 MST", lastCheckTime)
	if err != nil {
		config.Debug.logf("Could not parse the config file's lastversioncheck: %v. Continuing with a new version check...", err)
		doVersionCheck(config)
	} else if time.Since(lastTime).Hours() > 24 {
		doVersionCheck(config)
	}

}

// TEMPORARY CODE: sets the old repo name "appsodyhub" to the new name "incubator"
// this code should be removed when we think everyone is using the new name.
func setNewRepoName(config *RootCommandConfig) {
	var repoFile RepositoryFile
	_, repoErr := repoFile.getRepos(config)
	if repoErr != nil {
		config.Warning.log("Unable to read repository file")
	}
	appsodyhubRepo := repoFile.GetRepo("appsodyhub")
	if appsodyhubRepo != nil && appsodyhubRepo.URL == incubatorRepositoryURL {
		config.Info.log("Migrating your repo name from 'appsodyhub' to 'incubator'")
		appsodyhubRepo.Name = "incubator"
		err := repoFile.WriteFile(getRepoFileLocation(config))
		if err != nil {
			config.Warning.logf("Failed to write file to repository location: %v", err)
		}
	}
}

func IsEmptyDir(name string) bool {
	f, err := os.Open(name)

	if err != nil {
		return true
	}
	defer f.Close()

	_, err = f.Readdirnames(1)

	return err == io.EOF
}

func downloadFile(log *LoggingConfig, href string, writer io.Writer) error {

	// allow file:// scheme
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	log.Debug.log("Proxy function for HTTP transport set to: ", &t.Proxy)
	if runtime.GOOS == "windows" {
		// For Windows, remove the root url. It seems to work fine with an empty string.
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("")))
	} else {
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	}

	httpClient := &http.Client{Transport: t}

	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return err
	}

	if strings.Contains(href, "http") {
		token := os.Getenv("GH_READ_TOKEN")
		if token != "" {
			token = "token " + token
			req.Header.Add("Authorization", token)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debug.log("Could not read contents of response body: ", err)
		} else {
			log.Debug.logf("Contents http response:\n%s", buf)
		}
		return fmt.Errorf("Could not download %s: %s", href, resp.Status)
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("Could not copy http response body to writer: %s", err)
	}
	return nil
}

func downloadFileToDisk(log *LoggingConfig, url string, destFile string, dryrun bool) error {
	if dryrun {
		log.Info.logf("Dry Run -Skipping download of url: %s to destination %s", url, destFile)

	} else {
		outFile, err := os.Create(destFile)
		if err != nil {
			return err
		}
		defer outFile.Close()

		err = downloadFile(log, url, outFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// tar and zip a directory into .tar.gz
func Targz(log *LoggingConfig, source, target, filename string) error {
	log.Debug.log("source is: ", source)
	log.Debug.log("filename is: ", filename)
	log.Debug.log("target is: ", target)
	target = target + filename + ".tar.gz"
	//target = filepath.Join(target, fmt.Sprintf("%s.tar.gz", filename))
	log.Info.log(filename, " archive file created at : ", target)
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	var fileWriter io.WriteCloser = tarfile
	fileWriter = gzip.NewWriter(tarfile)
	defer fileWriter.Close()

	tarball := tar.NewWriter(fileWriter)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return errors.Wrap(err, "tar.FileInfoHeader")
			}

			log.Debug.logf("FileInfoHeader %s: %+v", info.Name(), header)

			if baseDir != "" {
				header.Name = "." + strings.TrimPrefix(path, source)
			}

			if err := tarball.WriteHeader(header); err != nil {
				return errors.Wrap(err, "tarball.WriteHeader")
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return errors.Wrap(err, "io.Copy")
		})
}

//Compares the minimum requirements of a stack against the user to determine whether they can use the stack or not.
func CheckStackRequirements(log *LoggingConfig, requirementArray map[string]string, buildah bool) error {
	versionRegex := regexp.MustCompile(`(\d)+\.(\d)+\.(\d)+`)
	upgradesRequired := 0

	log.Info.log("Checking stack requirements...")

	for technology, minVersion := range requirementArray {
		if minVersion == "" {
			continue
		}
		if technology == "Docker" && buildah {
			log.Debug.log("Skipping Docker requirement - Buildah is being used.")
			continue
		}
		if technology == "Buildah" && !buildah {
			log.Debug.log("Skipping Buildah requirement - Docker is being used.")
			continue
		}
		log.Debug.logf("Checking version requirement: %s %s", technology, minVersion)

		setConstraint, err := semver.NewConstraint(minVersion)
		if err != nil {
			log.Warning.logf("Skipping %s version requirement because the minimum version is invalid: %s", technology, err)
			continue
		}

		var runVersionCmd string
		if strings.ToLower(technology) == "appsody" {
			if VERSION == "0.0.0" || VERSION == "vlatest" {
				log.Warning.log("Skipping appsody version requirement because this is a local build of appsody ", VERSION)
				continue
			}
			runVersionCmd = VERSION
		} else {
			cmd := exec.Command(strings.ToLower(technology), "version")
			runVersionCmd, err = SeparateOutput(cmd)
			if err != nil {
				log.Error.log(err, " - Are you sure ", technology, " is installed?")
				upgradesRequired++
				continue
			}
			log.Debug.logf("Output of running %s: %s", strings.ToLower(technology)+" version", runVersionCmd)
		}

		cutCmdOutput := versionRegex.FindString(runVersionCmd)
		parseUserVersion, parseErr := semver.NewVersion(cutCmdOutput)
		if parseErr != nil {
			log.Warning.logf("Unable to parse %s version - This stack may not work in your current development environment. %s", technology, parseErr)
			// Continue when the version can not be determined
			continue
		}
		log.Debug.logf("Found version of %s to be %s", technology, parseUserVersion)
		compareVersion := setConstraint.Check(parseUserVersion)

		if compareVersion {
			log.Info.log(technology + " requirements met")
		} else {
			log.Error.log("The required version of " + technology + " to use this stack is " + minVersion + " - Please upgrade.")
			upgradesRequired++
		}
	}
	if upgradesRequired > 0 {
		return errors.Errorf("One or more technologies need upgrading to use this stack. Upgrades required: %v", upgradesRequired)
	}
	return nil
}

func SeparateOutput(cmd *exec.Cmd) (string, error) {
	var stdErr, stdOut bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut
	err := cmd.Run()

	// If there was an error, return the stdErr & err
	if err != nil {
		return err.Error() + ": " + strings.TrimSpace(stdErr.String()), err
	}

	// If there wasn't an error return the stdOut & (lack of) err
	return strings.TrimSpace(stdOut.String()), err
}

func CheckValidSemver(version string) error {
	versionRegex := regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	checkVersionNo := versionRegex.FindString(version)

	if checkVersionNo == "" {
		return errors.Errorf("Version must be formatted in accordance to semver - Please see: https://semver.org/ for valid versions.")
	}

	return nil
}

func checkValidLicense(log *LoggingConfig, license string) error {
	// Get the list of all known licenses
	list, _ := spdx.List()
	if list != nil {
		for _, spdx := range list.Licenses {
			if spdx.ID == license {
				return nil
			}
		}
	} else {
		log.Warning.log("Unable to check if license ID is valid.... continuing.")
		return nil
	}
	return errors.New("file must have a valid license ID, see https://spdx.org/licenses/ for the list of valid licenses")
}
func lintMountPathForSingleFile(path string, log *LoggingConfig) {

	file, err := os.Stat(path)
	if err != nil {
		log.Warning.logf("Could not stat mount path: %s", path)

	} else {
		if file.Mode().IsDir() {
			log.Debug.logf("Path %s for mount is a directory", path)
		} else {

			log.Warning.logf("Path %s for mount points to a single file.  Single file Docker mount paths cause unexpected behavior and will be deprecated in the future.", path)
		}

	}
}

// taken from https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func untar(log *LoggingConfig, dst string, r io.Reader, dryrun bool) error {
	if !dryrun {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)

		for {
			header, err := tr.Next()

			switch {

			// if no more files are found return
			case err == io.EOF:
				return nil

			// return any other error
			case err != nil:
				return err

			// if the header is nil, just skip it (not sure how this happens)
			case header == nil:
				continue
			}

			// the target location where the dir/file should be created
			target := filepath.Join(dst, header.Name)

			// the following switch could also be done using fi.Mode(), not sure if there
			// a benefit of using one vs. the other.
			// fi := header.FileInfo()

			// check the file type
			switch header.Typeflag {

			// if its a dir and it doesn't exist create it
			case tar.TypeDir:
				if _, err := os.Stat(target); err != nil {
					if err := os.MkdirAll(target, 0755); err != nil {
						return err
					}
				}

			// if it's a file create it
			case tar.TypeReg:
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return err
				}

				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					return err
				}

				// manually close here after each file operation; defering would cause each file close
				// to wait until all operations have completed.
				f.Close()
			}
		}
	} else {
		log.Info.logf("Dry Run skipping -Untar of file: %s to destination %s", r, dst)
		return nil
	}
}

// Converts an array of command arguments to a string of arguments
// properly escaped and quoted for copying and running in sh, bash, or zsh
func ArgsToString(args []string) string {
	charsToQuote := ` =*(|<[{?^@#$"'`
	returnStr := ""
	for i := 0; i < len(args); i++ {
		if strings.ContainsAny(args[i], charsToQuote) {
			returnStr += `"` + strings.Replace(args[i], `"`, `\"`, -1) + `"`
		} else {
			returnStr += args[i]
		}
		if i+1 != len(args) {
			returnStr += " "
		}
	}
	return returnStr
}

func dockerStop(rootConfig *RootCommandConfig, imageName string, dryrun bool) error {
	cmdName := "docker"
	cmdArgs := []string{"stop", imageName}
	err := execAndWait(rootConfig.LoggingConfig, cmdName, cmdArgs, rootConfig.Debug, dryrun)
	if err != nil {
		return err
	}
	return nil
}

func containerRemove(log *LoggingConfig, imageName string, buildah bool, dryrun bool) error {
	cmdName := "docker"
	//Added "-f" to force removal if container is still running or image has containers
	cmdArgs := []string{"rm", imageName, "-f"}
	if buildah {
		cmdName = "buildah"
		cmdArgs = []string{"rm", imageName}
	}
	err := execAndWait(log, cmdName, cmdArgs, log.Debug, dryrun)
	if err != nil {
		return err
	}
	return nil
}

//RemoveIfExists - Checks if inode exists and removes it if it does
func RemoveIfExists(path string) error {
	pathExists, err := Exists(path)
	if err != nil {
		return errors.Errorf("Error checking that: %v exists: %v", path, err)
	}
	if pathExists {
		err = os.RemoveAll(path)
		if err != nil {
			return errors.Errorf("Error removing: %v and children: %v", path, err)
		}
	}
	return nil
}

func generateCodewindJSON(log *LoggingConfig, indexYaml IndexYaml, indexFilePath string, repoName string) error {
	indexJSONStack := make([]IndexJSONStack, 0)
	prefixName := strings.Title(repoName)
	for _, stack := range indexYaml.Stacks {
		for _, template := range stack.Templates {
			stackJSON := IndexJSONStack{}
			stackJSON.DisplayName = prefixName + " " + stack.Name + " " + template.ID + " template"
			stackJSON.Description = stack.Description
			stackJSON.Language = stack.Language
			stackJSON.ProjectType = "appsodyExtension"
			stackJSON.ProjectStyle = "Appsody"
			stackJSON.Location = template.URL

			link := Link{}
			link.Self = "/devfiles/" + stack.ID + "/devfile.yaml"
			stackJSON.Links = link

			indexJSONStack = append(indexJSONStack, stackJSON)
		}
	}

	// Last thing to do is write the data to the file
	data, err := json.MarshalIndent(&indexJSONStack, "", "	")
	if err != nil {
		return err
	}
	indexFilePath = strings.Replace(indexFilePath, ".yaml", ".json", 1)

	err = ioutil.WriteFile(indexFilePath, data, 0666)
	if err != nil {
		return errors.Errorf("Error writing to json file: %v", err)
	}

	log.Info.logf("Succesfully generated file: %s", indexFilePath)
	return nil
}

func getProjectYamlPath(rootConfig *RootCommandConfig) string {
	return filepath.Join(getHome(rootConfig), "project.yaml")
}

// add a new project entry to the project.yaml file
func (p *ProjectFile) add(projectEntry ...*ProjectEntry) {
	p.Projects = append(p.Projects, projectEntry...)
}

// write to the project.yaml file
func (p *ProjectFile) writeFile(path string) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// check if project.yaml file had a project with given id
func (p *ProjectFile) hasID(id string) bool {
	for _, pf := range p.Projects {
		if id == pf.ID {
			return true
		}
	}
	return false
}

// get project from project.yaml with given id
func (p *ProjectFile) GetProject(id string) *ProjectEntry {
	for _, pf := range p.Projects {
		if id == pf.ID {
			return pf
		}
	}
	return nil
}

// get all project entries from project.yaml
func (p *ProjectFile) GetProjects(fileLocation string) (*ProjectFile, error) {
	projectReader, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(projectReader, p)
	if err != nil {
		return nil, errors.Errorf("Failed to parse project file %v", err)
	}
	return p, nil
}

// create unique project id for .appsody-config.yaml
func generateID(log *LoggingConfig) string {
	var id = time.Now().Format("20060102150405.00000000")

	log.Debug.Logf("Successfully generated ID: %s", id)
	return id
}

// add new project entry to ~/.appsody/project.yaml
func (p *ProjectFile) addNewProject(ID string, config *RootCommandConfig) error {
	projectDir, err := getProjectDir(config)
	if err != nil {
		return err
	}
	fileLocation := getProjectYamlPath(config)

	_, err = p.GetProjects(fileLocation)
	if err != nil {
		return err
	}

	var newEntry = ProjectEntry{
		ID:   ID,
		Path: projectDir,
	}
	p.add(&newEntry)
	err = p.writeFile(fileLocation)
	if err != nil {
		return errors.Errorf("Failed to write file to repository location: %v", err)
	}
	config.Info.Logf("Successfully added your project to %s", getProjectYamlPath(config))
	return nil
}

// get APPSODY_DEPS environment variable and split it into and array
func getDepVolumeArgs(config *RootCommandConfig) ([]string, error) {
	stackDeps, envErr := GetEnvVar("APPSODY_DEPS", config)
	if envErr != nil {
		return nil, envErr
	}
	if stackDeps == "" {
		config.Warning.log("The stack image does not contain APPSODY_DEPS")
		return nil, nil
	}
	volumeArgs := strings.Split(stackDeps, ";")
	return volumeArgs, nil
}

// create unique name for APPSODY_DEPS volumes
func generateVolumeName(config *RootCommandConfig) string {
	projectName, perr := getProjectName(config)
	if perr != nil {
		if _, ok := perr.(*NotAnAppsodyProject); !ok {
			config.Error.logf("Error occurred retrieving project name... exiting: %s", perr)
			os.Exit(1)
		}
	}
	ID := generateID(config.LoggingConfig)
	volumeName := "appsody-" + projectName + "-" + ID

	config.Debug.Logf("Using docker volume name: %s", volumeName)
	return volumeName
}

// save project id to .appsody-config.yaml
func SaveIDToConfig(ID string, config *RootCommandConfig) error {
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	v.Set("id", ID)
	err = v.WriteConfig()
	if err != nil {
		return err
	}

	config.Info.log("Your Appsody project ID has been set to ", ID)
	return nil
}

// get project id from .appsody-config.yaml and create project entry if id does not exist
func GetIDFromConfig(config *RootCommandConfig) (string, error) {
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err := v.ReadInConfig()
	if err != nil {
		return "", err
	}
	id := v.GetString("id")
	if id == "" {
		id, err := generateNewProjectAndID(config)
		return id, err
	}
	return id, nil
}

// create new project entry in ~/.appsody/project.yaml and add id to .appsody-config.yaml
func generateNewProjectAndID(config *RootCommandConfig) (string, error) {
	var projectFile ProjectFile
	ID := generateID(config.LoggingConfig)
	err := projectFile.addNewProject(ID, config)
	if err != nil {
		return "", err
	}

	err = SaveIDToConfig(ID, config)
	if err != nil {
		return "", err
	}
	return ID, nil
}

func (p *ProjectFile) ensureProjectIDAndEntryExists(rootConfig *RootCommandConfig) (*ProjectEntry, string, error) {
	id, err := GetIDFromConfig(rootConfig)
	if err != nil {
		return nil, "", err
	}
	var fileLocation = getProjectYamlPath(rootConfig)
	_, err = p.GetProjects(fileLocation)
	if err != nil {
		return nil, "", err
	}
	// if id exists in .appsody-config.yaml but not in project.yaml, add a new project entry in project.yaml with that id, and get projects again
	if !p.hasID(id) {
		err = p.addNewProject(id, rootConfig)
		if err != nil {
			return nil, "", err
		}
	}
	project := p.GetProject(id)

	projectDir, err := getProjectDir(rootConfig)
	if err != nil {
		return nil, "", err
	}
	// if the user moves their Appsody project, (same project id different path), update the project path in project.yaml
	if project.Path != projectDir {
		project.Path = projectDir
		if err := p.writeFile(fileLocation); err != nil {
			return nil, "", err
		}
	}
	return project, id, nil
}

// create docker volume names for every path in APPSODY_DEPS and put it in project.yaml
func (p *ProjectFile) addDepsVolumesToProjectEntry(depsEnvVars []string, volumeMaps []string, rootConfig *RootCommandConfig) ([]string, error) {
	project, _, err := p.ensureProjectIDAndEntryExists(rootConfig)
	if err != nil {
		return nil, err
	}

	// if the project entry does not have existing dependency volumes, for every path in APPSODY_DEPS, generate a new volume name, assign it to that path, and write it to the current project entry in project.yaml
	if project.Volumes == nil {
		for _, volumePath := range depsEnvVars {
			volumeName := generateVolumeName(rootConfig)
			depsMount := volumeName + ":" + volumePath
			rootConfig.Debug.log("Adding dependency cache to volume mounts: ", depsMount)
			// add the volume mounts to volumeMaps
			volumeMaps = append(volumeMaps, "-v", depsMount)

			v := new(Volume)
			v.Name = volumeName
			v.Path = volumePath
			project.Volumes = append(project.Volumes, v)
		}

		var fileLocation = getProjectYamlPath(rootConfig)
		if err := p.writeFile(fileLocation); err != nil {
			return volumeMaps, err
		}
	} else { // else if project entry has existing dependency volumes, loop through volumes in the current project entry, and add each volume mount to volumeMaps
		for _, v := range project.Volumes {
			depsMount := v.Name + ":" + v.Path
			rootConfig.Debug.log("Adding dependency cache to volume mounts: ", depsMount)
			volumeMaps = append(volumeMaps, "-v", depsMount)
		}
	}
	return volumeMaps, nil
}

/**

What it does:
	This function splits the build options by spaces, but only if the space is outside of any quotation mark block
	e.g "option1 option2" ---> ["option1", "option2"]
	e.g "option1='my option1' option2='my option2'" ---> ["option1='my option1'", "option2='my option2"]

How it works:
	It works by iterating over each element in the string.
	When the element is a Quotation Mark, it stores it in the variable `lastQuote`.
	It continues iterating until we find the next matching quote, if the next quote is escaped (\') when don't match.
	While this quote block hasn't been closed,  we don't split if we find a space.
	Once the next matching quote is found, we clear `lastQuote` and if any subsequent space is found we split.

Inspired from: https://play.golang.org/p/gJrqdeCr7k
**/

func SplitBuildOptions(options string) []string {
	slash := rune(92) // \ symbol

	lastQuote := rune(0)
	previousChar := rune(0)
	f := func(c rune) bool {
		result := false
		switch {
		case c == lastQuote:
			if previousChar != slash {
				lastQuote = rune(0)
			}
		case lastQuote != rune(0):
			break
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
		default:
			result = unicode.IsSpace(c)
		}

		previousChar = c
		return result
	}

	return strings.FieldsFunc(options, f)
}
