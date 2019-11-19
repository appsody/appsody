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
	"fmt"
	"os"
	"strings"

	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type extractCommandConfig struct {
	*RootCommandConfig
	targetDir            string
	extractContainerName string
}

func newExtractCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &extractCommandConfig{RootCommandConfig: rootConfig}

	var extractCmd = &cobra.Command{
		Use:   "extract",
		Short: "Extract the stack and your Appsody project to a local directory",
		Long: `This copies the full project, stack plus app, into a local directory
in preparation to build the final container image.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return extract(config)
		},
	}

	extractCmd.PersistentFlags().StringVar(&config.targetDir, "target-dir", "", "Directory path to place the extracted files. This dir must not exist, it will be created.")
	extractCmd.PersistentFlags().BoolVar(&rootConfig.Buildah, "buildah", false, "Extract project using buildah primitives instead of docker.")
	defaultName := defaultExtractContainerName(rootConfig)
	extractCmd.PersistentFlags().StringVar(&config.extractContainerName, "name", defaultName, "Assign a name to your development container.")
	return extractCmd
}

func extract(config *extractCommandConfig) error {
	extractContainerName := config.extractContainerName
	if extractContainerName == "" {
		extractContainerName = defaultExtractContainerName(config.RootCommandConfig)
	}

	projectName, perr := getProjectName(config.RootCommandConfig)
	if perr != nil {
		return errors.Errorf("%v", perr)
	}
	projectConfig, projectErr := getProjectConfig(config.RootCommandConfig)
	if projectErr != nil {
		return projectErr
	}
	config.Info.log("Extracting project from development environment")

	targetDir := config.targetDir
	if targetDir != "" {
		// the user specified a target dir, quit if it already exists
		targetDir, _ = filepath.Abs(targetDir)
		config.Debug.log("Checking if target-dir exists: ", targetDir)
		targetExists, err := Exists(targetDir)
		if err != nil {
			return errors.Errorf("Error checking target directory: %v", err)
		}
		if targetExists {
			return errors.Errorf("Cannot extract to an existing target-dir: %s", targetDir)

		}
		targetDirParent := filepath.Dir(targetDir)
		targetDirParentExists, err := Exists(targetDirParent)
		if err != nil {
			return errors.Errorf("Error checking directory: %v", err)
		}
		if !targetDirParentExists {
			return errors.Errorf("%s does not exist", targetDirParent)
		}
	}

	extractDir := filepath.Join(getHome(config.CliConfig), "extract")
	extractDirExists, err := Exists(extractDir)
	if err != nil {
		return errors.Errorf("Error checking directory: %v", err)
	}
	if !extractDirExists {
		if config.Dryrun {
			config.Info.log("Dry Run - Skip creating extract dir: ", extractDir)
		} else {
			config.Debug.log("Creating extract dir: ", extractDir)
			err = os.MkdirAll(extractDir, os.ModePerm)
			if err != nil {
				return errors.Errorf("Error creating directories %s %v", extractDir, err)
			}
		}
	}
	extractDir = filepath.Join(extractDir, projectName)
	extractDirExists, err = Exists(extractDir)
	if err != nil {
		return errors.Errorf("Error checking directory: %v", err)
	}
	if extractDirExists {
		if config.Dryrun {
			config.Info.log("Dry Run - Skip deleting extract dir: ", extractDir)
		} else {
			config.Debug.log("Deleting extract dir: ", extractDir)
			os.RemoveAll(extractDir)
		}
	}

	if config.Buildah {
		// Buildah fails if the destination does not exist.
		config.Debug.log("Creating extract dir: ", extractDir)
		err = os.MkdirAll(extractDir, os.ModePerm)
		if err != nil {
			return errors.Errorf("Error creating directories %s %v", extractDir, err)
		}
	}

	stackImage := projectConfig.Stack

	pullErr := pullImage(stackImage, config.RootCommandConfig)
	if pullErr != nil {
		return pullErr
	}

	containerProjectDir, containerProjectDirErr := getExtractDir(config.RootCommandConfig)
	if containerProjectDirErr != nil {
		return containerProjectDirErr
	}
	config.Debug.log("Container project dir: ", containerProjectDir)

	volumeMaps, volumeErr := getVolumeArgs(config.RootCommandConfig)
	if volumeErr != nil {
		return volumeErr
	}
	cmdName := "docker"
	if config.Buildah {
		cmdName = "buildah"
	}
	var appDir string
	cmdArgs := []string{"--name", extractContainerName}
	if len(volumeMaps) > 0 {
		cmdArgs = append(cmdArgs, volumeMaps...)
	}

	if runtime.GOOS != "windows" {
		// On Linux and OS/X we run docker create or buildah from
		if config.Buildah {
			cmdArgs = append([]string{"from"}, cmdArgs...)
		} else {
			cmdArgs = append([]string{"create"}, cmdArgs...)
		}
		cmdArgs = append(cmdArgs, stackImage)
		err = execAndWaitReturnErr(config.LoggingConfig, cmdName, cmdArgs, config.Debug, config.Dryrun)
		if err != nil {

			if config.Buildah {
				config.Error.log("buildah from command failed: ", err)
			} else {
				config.Error.log("docker create command failed: ", err)
			}
			removeErr := containerRemove(config.LoggingConfig, extractContainerName, config.Buildah, config.Dryrun)
			config.Error.log("Error in containerRemove", removeErr)
			return err

		}
		appDir = extractContainerName + ":" + containerProjectDir

	} else {
		// On Windows, we need to run the container to copy of the /project dir in /tmp/project
		// and navigate all the symlinks using cp -rL
		// then extract /tmp/project and remove the container

		bashCmd := "cp -rfL " + filepath.ToSlash(containerProjectDir) + " " + filepath.ToSlash(filepath.Join("/tmp", containerProjectDir))

		config.Debug.log("Attempting to run ", bashCmd, " on image: ", stackImage, " with args: ", cmdArgs)
		_, err = DockerRunBashCmd(cmdArgs, stackImage, bashCmd, config.RootCommandConfig)
		if err != nil {
			config.Debug.log("Error attempting to run copy command ", bashCmd, " on image ", stackImage, ": ", err)

			removeErr := containerRemove(config.LoggingConfig, extractContainerName, config.Buildah, config.Dryrun)
			if removeErr != nil {
				config.Error.log("containerRemove error ", removeErr)
			}

			return errors.Errorf("Error attempting to run copy command %s on image %s: %v", bashCmd, stackImage, err)

		}
		//If everything went fine, we need to set the source project directory to /tmp/...
		appDir = extractContainerName + ":" + filepath.Join("/tmp", containerProjectDir)
	}
	cmdArgs = []string{"cp", appDir, extractDir}
	if config.Buildah {
		appDir = containerProjectDir
		cmdName = "/bin/sh"
		script := fmt.Sprintf("x=`buildah mount %s`; cp -rf $x/%s/* %s", extractContainerName, appDir, extractDir)
		cmdArgs = []string{"-c", script}
	}
	err = execAndWaitReturnErr(config.LoggingConfig, cmdName, cmdArgs, config.Debug, config.Dryrun)
	if err != nil {
		if config.Buildah {
			config.Error.log("buildah mount / copy command failed: ", err)
		} else {
			config.Error.log("docker cp command failed: ", err)
		}

		removeErr := containerRemove(config.LoggingConfig, extractContainerName, config.Buildah, config.Dryrun)
		if removeErr != nil {
			config.Error.log("containerRemove error ", removeErr)
		}
		if config.Buildah {
			return errors.Errorf("buildah mount / copy command failed: %v", err)
		}
		return errors.Errorf("docker cp command failed: %v", err)
	}

	// A class of systems (e.g:- RHEL 7.6) exhibit situations wherein
	// the bindmount volumes are not propagated to the child containers
	// Accommodate those systems as well, by performing local copies
	// for the locations that are resident in the host.
	// ref: https://github.com/containers/buildah/issues/1821
	if config.Buildah {
		for _, item := range volumeMaps {
			if strings.Contains(item, ":") {
				config.Debug.log("Appsody mount: ", item)
				var src = strings.Split(item, ":")[0]
				var dest = strings.Split(item, ":")[1]
				if strings.EqualFold(src, ".") {
					// This should probably be using config.ProjectDir instead of os.Getwd()
					src, err = os.Getwd()
					if err != nil {
						return errors.Errorf("Error getting cwd: %v", err)
					}
				}
				dest = strings.Replace(dest, appDir, extractDir, -1)
				config.Debug.log("Local-adjusted mount destination: ", dest)
				fileInfo, err := os.Lstat(src)
				if err != nil {
					return errors.Errorf("Error lstat: %v", err)
				}
				var mkdir string
				if fileInfo.IsDir() {
					mkdir = dest
				} else {
					mkdir = filepath.Dir(dest)
				}
				err = os.MkdirAll(mkdir, os.ModePerm)
				if err != nil {
					return errors.Errorf("Error creating directories %s %v", extractDir, err)
				}

				fileInfo, err = os.Lstat(src)
				if err != nil {
					return errors.Errorf("project file check error %v", err)
				}
				config.Debug.log("Copy source: ", src)
				config.Debug.log("Copy destination: ", dest)
				if fileInfo.IsDir() {
					err = CopyDir(config.LoggingConfig, src+"/.", dest)
					if err != nil {
						return errors.Errorf("folder copy error %v", err)
					}
				} else {
					err = CopyFile(config.LoggingConfig, src, dest)
					if err != nil {
						return errors.Errorf("file copy error %v", err)
					}
				}
				config.Debug.log("Copied ", src, " to ", dest)
			}
		}
	}

	removeErr := containerRemove(config.LoggingConfig, extractContainerName, config.Buildah, config.Dryrun)
	if removeErr != nil {
		config.Error.log("containerRemove error ", removeErr)
	}
	if targetDir == "" {
		if !config.Dryrun {
			config.Info.log("Project extracted to ", extractDir)
		}
	} else {
		if config.Dryrun {
			config.Info.log("Dry Run - Skip moving ", extractDir, " to ", targetDir)
		} else {
			err = MoveDir(config.LoggingConfig, extractDir, targetDir)
			if err != nil {
				return errors.Errorf("Extract failed when moving %s to %s %v", extractDir, targetDir, err)

			}
			config.Info.log("Project extracted to ", targetDir)
		}
	}
	return nil
}

func defaultExtractContainerName(config *RootCommandConfig) string {
	projectName, perr := getProjectName(config)

	if perr != nil {
		if _, ok := perr.(*NotAnAppsodyProject); ok {
			//Debug.log("Cannot retrieve the project name - continuing: ", perr)
		} else {
			config.Error.log("Error occurred retrieving project name... exiting: ", perr)
			os.Exit(1)
		}
	}
	return projectName + "-extract"
}
