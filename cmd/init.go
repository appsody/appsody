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
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type initCommandConfig struct {
	*RootCommandConfig
	overwrite   bool
	noTemplate  bool
	projectName string
}

// these are global constants
var whiteListDotDirectories = []string{"github", "vscode", "settings", "metadata"}
var whiteListDotFiles = []string{"git", "project", "DS_Store", "classpath", "factorypath", "gitattributes", "gitignore", "cw-settings", "cw-extension"}

func newInitCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &initCommandConfig{RootCommandConfig: rootConfig}

	// initCmd represents the init command
	var initCmd = &cobra.Command{
		Use:   "init [stack] or [repository]/[stack] [template]",
		Short: "Initialize an Appsody project.",
		Long: `Set up the local Appsody development environment. You can do this for an existing project or use the template application provided by the stack. 

By default, the command creates an Appsody stack configuration file and provides a simple default application. You can also initialize a project with a different template application, or no template. 

To initialize a project with a template application, in a directory that is not empty, you need to specify the "overwrite" option [--overwrite].
Use 'appsody list' to see the available stacks and templates.`,
		Example: `  appsody init nodejs-express
  Initializes a project with the default template from the "nodejs-express" stack in the default repository.
  
  appsody init experimental/nodejs-functions
  Initializes a project with the default template from the "nodejs-functions" stack in the "experimental" repository.
  
  appsody init nodejs-express scaffold
  Initializes a project with the "scaffold" template from "nodejs-express" stack in the default repository.

  appsody init nodejs none
  Initializes a project without a template for the "nodejs" stack in the default repository.

  appsody init
  Runs the stack init script to set up the local development environment on an existing Appsody project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var stack string
			var template string
			if len(args) >= 1 {
				stack = args[0]
			}
			if len(args) >= 2 {
				template = args[1]
			}
			return initAppsody(stack, template, config)
		},
	}

	initCmd.PersistentFlags().BoolVar(&config.overwrite, "overwrite", false, "Download and extract the template project, overwriting existing files.  This option is not intended to be used in Appsody project directories.")
	initCmd.PersistentFlags().BoolVar(&config.noTemplate, "no-template", false, "Only create the .appsody-config.yaml file. Do not unzip the template project. [Deprecated]")
	defaultName := defaultProjectName(rootConfig)
	initCmd.PersistentFlags().StringVar(&config.projectName, "project-name", defaultName, "Project Name for Kubernetes Service")
	return initCmd
}

func initAppsody(stack string, template string, config *initCommandConfig) error {
	noTemplate := config.noTemplate
	if noTemplate {
		Warning.log("The --no-template flag has been deprecated.  Please specify a template value of \"none\" instead.")
	}
	valid, err := IsValidProjectName(config.projectName)
	if !valid {
		return err
	}
	//var index RepoIndex
	var repos RepositoryFile
	if _, err := repos.getRepos(config.RootCommandConfig); err != nil {
		return err
	}
	var proceedWithTemplate bool

	err = CheckPrereqs()
	if err != nil {
		Warning.logf("Failed to check prerequisites: %v\n", err)
	}

	//err = index.getIndex()

	indices, err := repos.GetIndices()

	if err != nil {
		Error.logf("The following indices could not be read, skipping:\n%v", err)
	}
	if len(indices) == 0 {
		return errors.Errorf("Your stack repository is empty - please use `appsody repo add` to add a repository.")
	}
	var index *RepoIndex

	if stack != "" {
		var projectName string
		projectParm := stack

		repoName, projectType, err := parseProjectParm(projectParm, config.RootCommandConfig)
		if err != nil {
			return err
		}
		if !repos.Has(repoName) {
			return errors.Errorf("Repository %s is not in configured list of repositories", repoName)
		}
		var templateName string
		var inputTemplateName string
		if template != "" {

			inputTemplateName = template
			if inputTemplateName == "none" {
				noTemplate = true
			}

		}

		templateName = inputTemplateName // so we can keep track

		Debug.log("Attempting to locate stack ", projectType, " in repo ", repoName)
		index = indices[repoName]
		projectFound := false
		stackFound := false

		if strings.Compare(index.APIVersion, supportedIndexAPIVersion) == 1 {
			Warning.log("The repository .yaml for " + repoName + " has a more recent APIVersion than the current Appsody CLI supports (" + supportedIndexAPIVersion + "), it is strongly suggested that you update your Appsody CLI to the latest version.")
		}
		if len(index.Projects[projectType]) >= 1 { //V1 repos
			projectFound = true
			//return errors.Errorf("Could not find a stack with the id \"%s\" in repository \"%s\". Run `appsody list` to see the available stacks or -h for help.", projectType, repoName)
			Debug.log("Project ", projectType, " found in repo ", repoName)

			// need to check template name vs default
			if !noTemplate && !(templateName == "" || templateName == index.Projects[projectType][0].DefaultTemplate) {
				return errors.Errorf("template name is not \"none\" and does not match %s.", index.Projects[projectType][0].DefaultTemplate)
			}
			projectName = index.Projects[projectType][0].URLs[0]

		}
		for _, stack := range index.Stacks {
			if stack.ID == projectType {
				stackFound = true
				Debug.log("Stack ", projectType, " found in repo ", repoName)
				URL := ""
				if templateName == "" || templateName == "none" {
					templateName = stack.DefaultTemplate
					if templateName == "" {
						return errors.Errorf("Cannot proceed, no template or \"none\" was specified and there is no default template.")
					}
				}
				URL = findTemplateURL(stack, templateName)

				projectName = URL
			}
		}
		if !projectFound && !stackFound {
			return errors.Errorf("Could not find a stack with the id \"%s\" in repository \"%s\". Run `appsody list` to see the available stacks or -h for help.", projectType, repoName)
		}

		if projectName == "" && inputTemplateName != "none" {
			return errors.Errorf("Could not find a template \"%s\" for stack id \"%s\" in repository \"%s\"", templateName, projectType, repoName)
		}

		// 1. Check for empty directory
		dir := config.ProjectDir
		appsodyConfigFile := filepath.Join(dir, ".appsody-config.yaml")

		_, err = os.Stat(appsodyConfigFile)
		if err == nil {
			return errors.New("cannot run `appsody init <stack>` on an existing appsody project")

		}

		if noTemplate && !(inputTemplateName == "" || inputTemplateName == "none") {

			return errors.Errorf("cannot specify `appsody init <stack> <template>` with both a template and --no-template")

		}

		if noTemplate || config.overwrite {
			proceedWithTemplate = true
		} else {
			proceedWithTemplate, err = isFileLaydownSafe(dir)
			if err != nil {
				return err
			}
		}

		if !config.overwrite && !proceedWithTemplate {
			Error.log("Non-empty directory found with files which may conflict with the template project.")
			Info.log("It is recommended that you run `appsody init <stack>` in an empty directory.")
			Info.log("If you wish to proceed and possibly overwrite files in the current directory, try again with the --overwrite option.")
			return errors.New("non-empty directory found with files which may conflict with the template project")

		}

		Info.log("Running appsody init...")
		Info.logf("Downloading %s template project from %s", projectType, projectName)
		filename := filepath.Join(dir, projectType+".tar.gz")

		err = downloadFileToDisk(projectName, filename, config.Dryrun)
		if err != nil {
			return errors.Errorf("Error downloading tar %v", err)

		}
		if inputTemplateName != "none" {
			Info.log("Download complete. Extracting files from ", filename)
		} else {
			Info.log("Download complete. Do not unzip the template project. Only extracting .appsody-config.yaml file from ", filename)
		}

		//if noTemplate
		errUntar := untar(filename, noTemplate, config.overwrite, config.Dryrun)

		if config.Dryrun {
			Info.logf("Dry Run - Skipping remove of temporary file for project type: %s project name: %s", projectType, projectName)
		} else {
			err = os.Remove(filename)
			if err != nil {
				Warning.log("Unable to remove temporary file ", filename)
			}
		}
		if errUntar != nil {
			Error.log("Error extracting project template: ", errUntar)
			Info.log("It is recommended that you run `appsody init <stack>` in an empty directory.")
			Info.log("If you wish to proceed and overwrite files in the current directory, try again with the --overwrite option.")
			// this leave the tar file in the dir
			return errors.Errorf("Error extracting project template: %v", errUntar)
		}

	}
	err = install(config)
	if err != nil {
		return err
	}
	if template == "" {
		Info.logf("Successfully initialized Appsody project with the %s stack and the default template.", stack)
	} else if template != "none" {
		Info.logf("Successfully initialized Appsody project with the %s stack and the %s template.", stack, template)
	} else {
		Info.logf("Successfully initialized Appsody project with the %s stack and no template.", stack)
	}

	return nil
}

//Runs the .appsody-init.sh/bat files if necessary
func install(config *initCommandConfig) error {
	Info.log("Setting up the development environment")
	projectDir, perr := getProjectDir(config.RootCommandConfig)
	if perr != nil {
		return perr
	}

	projectConfig, configErr := getProjectConfig(config.RootCommandConfig)
	if configErr != nil {
		return configErr
	}

	// save the project name to .appsody-config.yaml only if it doesn't already exist there
	// or if the user specified --project-name on the command line
	if projectConfig.ProjectName == "" || config.projectName != defaultProjectName(config.RootCommandConfig) {
		err := saveProjectNameToConfig(config.projectName, config.RootCommandConfig)
		if err != nil {
			return err
		}
	}
	platformDefinition := projectConfig.Stack

	Debug.logf("Setting up the development environment for projectDir: %s and platform: %s", projectDir, platformDefinition)

	err := extractAndInitialize(config)
	if err != nil {
		// For some reason without this sleep, the [InitScript] output log would get cut off and
		// intermixed with the following Warning logs when verbose logging. Adding this sleep as a workaround.
		time.Sleep(100 * time.Millisecond)
		Warning.log("The stack init script failed: ", err)
		Warning.log("Your local IDE may not build properly, but the Appsody container should still work.")
		Warning.log("To try again, resolve the issue then run `appsody init` with no arguments.")
		os.Exit(0)
	}
	return nil
}

func untar(file string, noTemplate bool, overwrite bool, dryrun bool) error {

	if dryrun {
		Info.log("Dry Run - Skipping untar of file:  ", file)
	} else {
		untarDir := filepath.Dir(file)
		if !overwrite && !noTemplate {
			err := preCheckTar(file, untarDir)
			if err != nil {
				return err
			}
		}
		fileReader, err := os.Open(file)
		if err != nil {
			return err
		}

		defer fileReader.Close()
		gzipReader, err := gzip.NewReader(fileReader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		tarReader := tar.NewReader(gzipReader)
		for {
			header, err := tarReader.Next()

			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if header == nil {
				continue
			}

			filename := filepath.Join(untarDir, header.Name)
			Debug.log("Untar creating ", filename)

			if header.Typeflag == tar.TypeDir && !noTemplate {
				if _, err := os.Stat(filename); err != nil {
					err := os.MkdirAll(filename, 0755)
					if err != nil {
						return err
					}
				}
			} else if header.Typeflag == tar.TypeReg {
				if !noTemplate || (noTemplate && strings.HasSuffix(filename, ".appsody-config.yaml")) {

					f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
					if err != nil {
						return err
					}
					_, err = io.Copy(f, tarReader)
					if err != nil {
						return err
					}
					f.Close()
				}
			}

		}
	}
	return nil
}

func isFileLaydownSafe(directory string) (bool, error) {

	safe := true
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		Error.logf("Can not read directory %s due to error: %v.", directory, err)
		return false, err

	}
	for _, f := range files {

		whiteListed := inWhiteList(f.Name())
		if !whiteListed {
			safe = false
			Debug.logf("%s file exists and is not safe to extract the project template over", f.Name())
		} else {
			Debug.logf("%s file exists and is safe to extract the project template over", f.Name())
		}
	}
	if safe {
		Debug.log("It is safe to extract the project template")
	} else {
		Debug.log("It is not safe to extract the project template")
	}
	return safe, nil

}

func buildOrList(args []string) string {
	base := ""
	for _, fileName := range args {
		base += fileName
		base += "|"
	}
	if base != "" {
		base = base[:len(base)-1]
	}

	return base
}

func inWhiteList(filename string) bool {
	whiteListTest := "(^(.[/\\\\])?.(" +
		buildOrList(whiteListDotFiles) +
		")$)|(^(.[/\\\\])?.(" + buildOrList(whiteListDotDirectories) + ")[/\\\\]?.*)"

	whiteListRegexp := regexp.MustCompile(whiteListTest)
	isWhiteListed := whiteListRegexp.MatchString(filename)

	return isWhiteListed
}

func preCheckTar(file string, untarDir string) error {
	preCheckOK := true
	fileReader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	// precheck the tar for whitelisted files
	for {
		header, err := tarReader.Next()

		if err == io.EOF {

			break
		} else if err != nil {

			return err
		}
		if header == nil {
			continue
		} else {
			if inWhiteList(header.Name) {
				filename := filepath.Join(untarDir, header.Name)
				fileInfo, err := os.Stat(filename)
				if err == nil {
					if !fileInfo.IsDir() {
						preCheckOK = false
						Warning.log("Conflict: " + header.Name + " exists in the file system and the template project.")

					}

				}
			}
		}
	}
	if !preCheckOK {
		err = errors.New("conflicts exist")
	}
	return err
}
func extractAndInitialize(config *initCommandConfig) error {

	var err error

	scriptFile := "./.appsody-init.sh"
	if runtime.GOOS == "windows" {
		scriptFile = ".\\.appsody-init.bat"
	}

	scriptFileName := filepath.Base(scriptFile)
	//Determine if we need to run extract
	//We run it only if there is an initialization script to run locally
	//Checking if the script is present on the image
	projectConfig, configErr := getProjectConfig(config.RootCommandConfig)
	if configErr != nil {
		return configErr
	}
	stackImage := projectConfig.Stack
	containerProjectDir, containerProjectDirErr := getExtractDir(config.RootCommandConfig)
	if containerProjectDirErr != nil {
		return containerProjectDirErr
	}
	if !config.RootCommandConfig.Buildah { //We can skip extract in some cases
		bashCmd := "find " + containerProjectDir + " -type f -name " + scriptFileName
		cmdOptions := []string{"--rm"}
		Debug.log("Attempting to run ", bashCmd, " on image ", stackImage, " with options: ", cmdOptions)
		//DockerRunBashCmd has a pullImage call
		scriptFindOut, err := DockerRunBashCmd(cmdOptions, stackImage, bashCmd, config.RootCommandConfig)
		if err != nil {
			Debug.log("Failed to run the find command for the ", scriptFileName, " on the stack image: ", stackImage)
			return fmt.Errorf("Failed to run the docker find command: %s", scriptFindOut)
		}

		if scriptFindOut == "" {
			Debug.log("There is no initialization script in the image - skipping extract and initialize")
			return nil
		}
	}
	workdir := ".appsody_init"

	// run the extract command here
	if !config.Dryrun {
		workdirExists, err := Exists(workdir)
		if workdirExists && err == nil {
			err = os.RemoveAll(workdir)
			if err != nil {
				return fmt.Errorf("Could not remove temp dir %s  %s", workdir, err)
			}
		}
		extractConfig := &extractCommandConfig{RootCommandConfig: config.RootCommandConfig}
		extractConfig.targetDir = workdir
		extractError := extract(extractConfig)
		if extractError != nil {
			return extractError
		}

	} else {
		Info.log("Dry Run skipping extract.")
	}

	scriptPath := filepath.Join(workdir, scriptFile)
	scriptExists, err := Exists(scriptPath)

	if scriptExists && err == nil { // if it doesn't exist, don't run it
		Debug.log("Running appsody_init script ", scriptFile)
		err = execAndWaitWithWorkDirReturnErr(scriptFile, nil, InitScript, workdir, config.Dryrun)
		if err != nil {
			return err
		}
	}

	if !config.Dryrun {
		Debug.log("Removing ", workdir)
		err = os.RemoveAll(workdir)
		if err != nil {
			return fmt.Errorf("Could not remove temp dir %s  %s", workdir, err)
		}
	}

	return err
}

func parseProjectParm(projectParm string, config *RootCommandConfig) (string, string, error) {
	parms := strings.Split(projectParm, "/")
	if len(parms) == 1 {
		Debug.log("Non-fully qualified stack - retrieving default repo...")
		var r RepositoryFile
		if _, err := r.getRepos(config); err != nil {
			return "", "", err
		}
		defaultRepoName, err := r.GetDefaultRepoName(config)
		if err != nil {
			return "", parms[0], err
		}
		return defaultRepoName, parms[0], nil
	}

	if len(parms) == 2 {
		Debug.log("Fully qualified stack... determining repo...")
		if len(parms[0]) == 0 || len(parms[1]) == 0 {
			return parms[0], parms[1], errors.New("malformed project parameter - slash at the beginning or end should be removed")
		}
		return parms[0], parms[1], nil
	}
	if len(parms) > 2 {
		return parms[0], parms[1], errors.New("malformed project parameter - too many slashes")
	}

	return "", "", errors.New("malformed project parameter - something unusual happened")
}

func defaultProjectName(config *RootCommandConfig) string {
	projectDirPath, perr := getProjectDir(config)
	if perr != nil {
		if _, ok := perr.(*NotAnAppsodyProject); ok {
			//Debug.log("Cannot retrieve the project dir - continuing: ", perr)
		} else {
			Error.logf("Error occurred retrieving project dir... exiting: %s", perr)
			os.Exit(1)
		}
	}

	projectName, err := ConvertToValidProjectName(projectDirPath)
	if err != nil {
		Error.log(err)
		os.Exit(1)
	}
	return projectName
}
