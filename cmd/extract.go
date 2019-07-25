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
	"os"

	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var targetDir string
var extractContainerName string

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract the stack and your Appsody project to a local directory",
	Long: `This copies the full project, stack plus app, into a local directory
in preparation to build the final docker image.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectName, perr := getProjectName()
		if perr != nil {
			Error.log(perr)
			os.Exit(1)
		}
		projectConfig := getProjectConfig()
		Info.log("Extracting project from development environment")

		if targetDir != "" {
			// the user specified a target dir, quit if it already exists
			targetDir, _ = filepath.Abs(targetDir)
			Debug.log("Checking if target-dir exists: ", targetDir)
			targetExists, err := exists(targetDir)
			if err != nil {
				Error.log("Error checking target directory: ", err)
				os.Exit(1)
			}
			if targetExists {
				Error.log("Cannot extract to an existing target-dir: ", targetDir)
				os.Exit(1)
			}
			targetDirParent := filepath.Dir(targetDir)
			targetDirParentExists, err := exists(targetDirParent)
			if err != nil {
				Error.log("Error checking directory: ", err)
				os.Exit(1)
			}
			if !targetDirParentExists {
				Error.log(targetDirParent, " does not exist")
				os.Exit(1)
			}
		}

		extractDir := filepath.Join(getHome(), "extract")
		extractDirExists, err := exists(extractDir)
		if err != nil {
			Error.log("Error checking directory: ", err)
			os.Exit(1)
		}
		if !extractDirExists {
			if dryrun {
				Info.log("Dry Run - Skip creating extract dir: ", extractDir)
			} else {
				Debug.log("Creating extract dir: ", extractDir)
				err = os.MkdirAll(extractDir, os.ModePerm)
				if err != nil {
					Error.log("Error creating directories ", extractDir, " ", err)
					os.Exit(1)
				}
			}
		}
		extractDir = filepath.Join(extractDir, projectName)
		extractDirExists, err = exists(extractDir)
		if err != nil {
			Error.log("Error checking directory: ", err)
			os.Exit(1)
		}
		if extractDirExists {
			if dryrun {
				Info.log("Dry Run - Skip deleting extract dir: ", extractDir)
			} else {
				Debug.log("Deleting extract dir: ", extractDir)
				os.RemoveAll(extractDir)
			}
		}

		stackImage := projectConfig.Platform

		dockerPullImage(stackImage)

		containerProjectDir := "/project"
		Debug.log("Container project dir: ", containerProjectDir)
		volumeMaps := getVolumeArgs()
		cmdName := "docker"
		var appDir string
		cmdArgs := []string{"--name", extractContainerName}
		if len(volumeMaps) > 0 {
			cmdArgs = append(cmdArgs, volumeMaps...)
		}

		if runtime.GOOS != "windows" {
			// On Linux and OS/X we run docker create
			cmdArgs = append([]string{"create"}, cmdArgs...)
			cmdArgs = append(cmdArgs, stackImage)
			err = execAndWaitReturnErr(cmdName, cmdArgs, Debug)

			if err != nil {
				Error.log("docker create command failed: ", err)
				dockerRemove(extractContainerName)
				os.Exit(1)
			}
			appDir = extractContainerName + ":" + containerProjectDir

		} else {
			// On Windows, we need to run the container to copy of the /project dir in /tmp/project
			// and navigate all the symlinks using cp -rL
			// then extract /tmp/project and remove the container

			bashCmd := "cp -rfL " + filepath.ToSlash(containerProjectDir) + " " + filepath.ToSlash(filepath.Join("/tmp", containerProjectDir))

			Debug.log("Attempting to run ", bashCmd, " on image: ", stackImage, " with args: ", cmdArgs)
			_, err = DockerRunBashCmd(cmdArgs, stackImage, bashCmd)
			if err != nil {
				Debug.log("Error attempting to run copy command ", bashCmd, " on image ", stackImage)
				dockerRemove(extractContainerName)
				os.Exit(1)
			}
			//If everything went fine, we need to set the source project directory to /tmp/...
			appDir = extractContainerName + ":" + filepath.Join("/tmp", containerProjectDir)
		}
		cmdArgs = []string{"cp", appDir, extractDir}
		err = execAndWaitReturnErr(cmdName, cmdArgs, Debug)

		if err != nil {
			Error.log("docker cp command failed: ", err)
			dockerRemove(extractContainerName)
			os.Exit(1)
		}
		dockerRemove(extractContainerName)
		if targetDir == "" {
			if !dryrun {
				Info.log("Project extracted to ", extractDir)
			}
		} else {
			if dryrun {
				Info.log("Dry Run - Skip moving ", extractDir, " to ", targetDir)
			} else {
				err = MoveDir(extractDir, targetDir)
				if err != nil {
					Error.log("Extract failed when moving ", extractDir, " to ", targetDir, " ", err)
					os.Exit(1)
				}
				Info.log("Project extracted to ", targetDir)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.PersistentFlags().StringVar(&targetDir, "target-dir", "", "Directory path to place the extracted files. This dir must not exist, it will be created.")
	// curDir, err := os.Getwd()
	// if err != nil {
	//		Error.log("Error getting current directory ", err)
	//	os.Exit(1)
	//}
	//defaultName := filepath.Base(curDir) + "-extract"
	projectName, perr := getProjectName()

	if perr != nil {
		if pmsg, ok := perr.(*NotAnAppsodyProject); ok {
			Debug.log("Cannot retrieve the project name - continuing: ", perr)
		} else {
			Error.log("Error occurred retrieving project name... exiting: ", pmsg)
			os.Exit(1)
		}
	}
	defaultName := projectName + "-extract"
	extractCmd.PersistentFlags().StringVar(&extractContainerName, "name", defaultName, "Assign a name to your development container.")
}
