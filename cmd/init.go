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

var (
	overwrite  bool
	noTemplate bool
)
var whiteListDotDirectories = []string{"github", "vscode", "settings", "metadata"}
var whiteListDotFiles = []string{"git", "project", "DS_Store", "classpath", "factorypath", "gitattributes", "gitignore", "cw-settings", "cw-extension"}

// initCmd represents the init command

var initCmd = &cobra.Command{
	Use:   "init [stack]",
	Short: "Initialize an Appsody project with a stack and template app",
	Long: `This creates a new Appsody project in a local directory or sets up the local dev environment of an existing Appsody project.

With the [stack] argument, this command will setup a new Appsody project. It will create an Appsody stack config file, unzip a template app, and
run the stack init script to setup the local dev environment. It is typically run on an empty directory and may fail
if files already exist. See the --overwrite and --no-template options for more details.
Use 'appsody list' to see the available stack options.

Without the [stack] argument, this command must be run on an existing Appsody project and will only run the stack init script to
setup the local dev environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var index RepoIndex

		var proceedWithTemplate bool

		err := CheckPrereqs()
		if err != nil {
			Warning.logf("Failed to check prerequisites: %v\n", err)
		}

		err = index.getIndex()
		if err != nil {
			return errors.Errorf("Could not read index: %v", err)
		}
		if len(args) >= 1 {

			projectType := args[0]

			if len(index.Projects[projectType]) < 1 {
				return errors.Errorf("Could not find a stack with the id \"%s\". Run `appsody list` to see the available stacks or -h for help.", projectType)

			}
			var projectName = index.Projects[projectType][0].URLs[0]

			// 1. Check for empty directory
			dir, err := os.Getwd()
			if err != nil {
				return errors.Errorf("Error getting current directory %v", err)
			}
			appsodyConfigFile := filepath.Join(dir, ".appsody-config.yaml")

			_, err = os.Stat(appsodyConfigFile)
			if err == nil {
				return errors.New("cannot run `appsody init <stack>` on an existing appsody project")

			}

			if noTemplate || overwrite {
				proceedWithTemplate = true
			} else {
				proceedWithTemplate, err = isFileLaydownSafe(dir)
				if err != nil {
					return err
				}
			}

			if !overwrite && !proceedWithTemplate {
				Error.log("Non-empty directory found with files which may conflict with the template project.")
				Info.log("It is recommended that you run `appsody init <stack>` in an empty directory.")
				Info.log("If you wish to proceed and possibly overwrite files in the current directory, try again with the --overwrite option.")
				return errors.New("non-empty directory found with files which may conflict with the template project")

			}

			Info.log("Running appsody init...")
			Info.logf("Downloading %s template project from %s", projectType, projectName)
			filename := projectType + ".tar.gz"

			err = downloadFileToDisk(projectName, filename)
			if err != nil {
				return errors.Errorf("Error downloading tar %v", err)

			}
			Info.log("Download complete. Extracting files from ", filename)
			//if noTemplate
			errUntar := untar(filename, noTemplate)

			if dryrun {
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
		err = install()
		if err != nil {
			return err
		}
		Info.log("Successfully initialized Appsody project")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "Download and extract the template project, overwriting existing files.")
	initCmd.PersistentFlags().BoolVar(&noTemplate, "no-template", false, "Only create the .appsody-config.yaml file. Do not unzip the template project.")
}

//Runs the .appsody-init.sh/bat files if necessary
func install() error {
	Info.log("Setting up the development environment")
	projectDir, perr := getProjectDir()
	if perr != nil {
		return errors.Errorf("%v", perr)

	}
	projectConfig, configErr := getProjectConfig()
	if configErr != nil {
		return configErr
	}
	platformDefinition := projectConfig.Platform

	Debug.logf("Setting up the development environment for projectDir: %s and platform: %s", projectDir, platformDefinition)

	err := extractAndInitialize()
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

func downloadFileToDisk(url string, destFile string) error {
	if dryrun {
		Info.logf("Dry Run -Skipping download of url: %s to destination %s", url, destFile)

	} else {
		outFile, err := os.Create(destFile)
		if err != nil {
			return err
		}
		defer outFile.Close()

		err = downloadFile(url, outFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func untar(file string, noTemplate bool) error {

	if dryrun {
		Info.log("Dry Run - Skipping untar of file:  ", file)
	} else {
		if !overwrite && !noTemplate {
			err := preCheckTar(file)
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

			filename := header.Name
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

func preCheckTar(file string) error {
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
				fileInfo, err := os.Stat(header.Name)
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
func extractAndInitialize() error {

	var err error

	scriptFile := "./.appsody-init.sh"
	if runtime.GOOS == "windows" {
		scriptFile = ".\\.appsody-init.bat"
	}

	scriptFileName := filepath.Base(scriptFile)
	//Determine if we need to run extract
	//We run it only if there is an initialization script to run locally
	//Checking if the script is present on the image
	projectConfig, configErr := getProjectConfig()
	if configErr != nil {
		return configErr
	}
	stackImage := projectConfig.Platform
	bashCmd := "find /project -type f -name " + scriptFileName
	cmdOptions := []string{"--rm"}
	Debug.log("Attempting to run ", bashCmd, " on image ", stackImage, " with options: ", cmdOptions)
	//DockerRunBashCmd has a pullImage call
	scriptFindOut, err := DockerRunBashCmd(cmdOptions, stackImage, bashCmd)
	if err != nil {
		Debug.log("Failed to run the find command for the ", scriptFileName, " on the stack image: ", stackImage)
		return fmt.Errorf("Failed to run the docker find command: %s", err)
	}

	if scriptFindOut == "" {
		Debug.log("There is no initialization script in the image - skipping extract and initialize")
		return nil
	}

	workdir := ".appsody_init"

	// run the extract command here
	if !dryrun {
		workdirExists, err := exists(workdir)
		if workdirExists && err == nil {
			err = os.RemoveAll(workdir)
			if err != nil {
				return fmt.Errorf("Could not remove temp dir %s  %s", workdir, err)
			}
		}
		// set the --target-dir flag for extract
		targetDir = workdir

		extractError := extractCmd.RunE(extractCmd, nil)
		if extractError != nil {
			return extractError

		}

	} else {
		Info.log("Dry Run skipping extract.")
	}

	scriptPath := filepath.Join(workdir, scriptFile)
	scriptExists, err := exists(scriptPath)

	if scriptExists && err == nil { // if it doesn't exist, don't run it
		Debug.log("Running appsody_init script ", scriptFile)
		err = execAndWaitWithWorkDirReturnErr(scriptFile, nil, InitScript, workdir)
		if err != nil {
			return err
		}
	}

	if !dryrun {
		Debug.log("Removing ", workdir)
		err = os.RemoveAll(workdir)
		if err != nil {
			return fmt.Errorf("Could not remove temp dir %s  %s", workdir, err)
		}
	}

	return err
}
