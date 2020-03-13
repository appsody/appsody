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

	//"crypto/sha256"
	"encoding/json"
	"fmt"

	//"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
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
				return err
			}

			if baseDir != "" {
				header.Name = "." + strings.TrimPrefix(path, source)
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
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
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, homeErr
	}
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

	for _, mount := range stackMountList {
		if mount == "" {
			continue
		}
		var mappedMount string
		var overridden bool
		if strings.HasPrefix(mount, "~") {
			mappedMount = strings.Replace(mount, "~", homeDir, 1)
			overridden = homeDirOverridden
		} else {
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

func getProjectConfigFileContents(config *RootCommandConfig) (*ProjectConfig, error) {

	dir, perr := getProjectDir(config)
	if perr != nil {
		return nil, perr
	}
	appsodyConfig := filepath.Join(dir, ConfigFile)

	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	config.Debug.log("Project config file set to: ", appsodyConfig)

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
			config.Debug.Log("Stack registry detected in project config file: ", stackElements[0])
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

func execAndWait(log *LoggingConfig, command string, args []string, logger appsodylogger, dryrun bool) error {

	return execAndWaitWithWorkDir(log, command, args, logger, workDirNotSet, dryrun)
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

	cmdName := "docker"
	if buildah {
		cmdName = "buildah"
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

			link := Links{}
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

func checkOptions(options []string, regex string) error {
	blackListedOptions := regexp.MustCompile(regex)
	for _, value := range options {
		isInBlackListed := blackListedOptions.MatchString(value)
		if isInBlackListed {
			return errors.Errorf("%s is not allowed in --docker-options", value)
		}
	}
	return nil
}
