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
	"os/exec"
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
		Short: "Extract your Appsody project to a local directory.",
		Long: `Extract the full application (the stack and your Appsody project) into a local directory.
		
Your project is extracted into your local '$HOME/.appsody/extract' directory, unless you use the --target-dir flag to specify a different location.

Run this command from the root directory of your Appsody project.`,
		Example: `  appsody extract --target-dir $HOME/my-extract/directory
  Extracts your project from the container to the local '$HOME/my-extract/directory' on your system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("Unexpected argument. Use 'appsody [command] --help' for more information about a command")
			}
			var project ProjectFile
			_, _, err := project.EnsureProjectIDAndEntryExists(config.RootCommandConfig)
			if err != nil {
				return err
			}
			return extract(config)
		},
	}

	extractCmd.PersistentFlags().StringVar(&config.targetDir, "target-dir", "", "The absolute directory path to extract the files into. This directory must not exist, as it will be created.")
	extractCmd.PersistentFlags().BoolVar(&rootConfig.Buildah, "buildah", false, "Extract project using buildah primitives instead of Docker.")
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

	// Even if targetDir is specified, we still extract to ~/.appsody/extract first because if the user's
	// targetDir is in the project it would turn into a recursive copy
	extractDir := filepath.Join(getHome(config.RootCommandConfig), "extract")
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

	stackImage := projectConfig.Stack

	pullErr := pullImage(stackImage, config.RootCommandConfig)
	if pullErr != nil {
		return pullErr
	}

	// Gets the project path within the container to extract
	containerProjectDir, containerProjectDirErr := getExtractDir(config.RootCommandConfig)
	if containerProjectDirErr != nil {
		return containerProjectDirErr
	}
	config.Debug.log("Container project dir: ", containerProjectDir)

	// Now we need to create a container with the same volume mappings as we would with the run/debug/test commands
	// using the `docker create` or `buildah from` command
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

	// When done, or if something goes wrong, remove the temporary container
	defer func() {
		if config.Dryrun {
			config.Info.log("Dry Run - Skip container remove: ", extractContainerName)
		} else {
			removeErr := containerRemove(config.LoggingConfig, extractContainerName, config.Buildah, config.Dryrun)
			if removeErr != nil {
				config.Warning.log("Ignoring container remove error ", removeErr)
			}
		}
	}()

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
			return err

		}
		appDir = extractContainerName + ":" + containerProjectDir

	} else {
		// On Windows, we need to run the container to copy of the /project dir in /tmp/project
		// and navigate all the symlinks using cp -rL
		// then extract /tmp/project and remove the container

		// We can't use CopyDir() func here because it needs to run in the container
		bashCmd := "cp -rfL " + filepath.ToSlash(containerProjectDir) + " " + filepath.ToSlash(filepath.Join("/tmp", containerProjectDir))
		if config.Dryrun {
			config.Info.log("Dry Run - Skip running ", bashCmd, " on image: ", stackImage, " with args: ", cmdArgs)
		} else {
			config.Debug.log("Attempting to run ", bashCmd, " on image: ", stackImage, " with args: ", cmdArgs)
			_, err = DockerRunBashCmd(cmdArgs, stackImage, bashCmd, config.RootCommandConfig)
			if err != nil {
				return errors.Errorf("Error attempting to run copy command %s on image %s: %v", bashCmd, stackImage, err)
			}
		}
		//If everything went fine, we need to set the source project directory to /tmp/...
		appDir = extractContainerName + ":" + filepath.Join("/tmp", containerProjectDir)
	}

	// Now we need to copy files out of the container using the `docker cp` or `buildah mount` commands
	if config.Buildah {
		// In buildah, we need to mount the container filesystem then manually copy the files out
		cmdArgs := []string{"mount", extractContainerName}
		if config.Dryrun {
			config.Info.log("Dry Run - Skip running buildah mount and copying files")
		} else {
			config.Debug.Logf("About to run %s with args %s ", cmdName, cmdArgs)
			buildahMountCmd := exec.Command(cmdName, cmdArgs...)
			buildahMountOutput, err := SeparateOutput(buildahMountCmd)
			config.Debug.Log("Output of buildah mount command: ", buildahMountOutput)
			if err != nil {
				return errors.Errorf("buildah mount command failed: %v", err)
			}

			appDir = filepath.Join(buildahMountOutput, containerProjectDir)
			err = CopyDir(config.LoggingConfig, appDir, extractDir)
			if err != nil {
				return errors.Errorf("Problem copying directory %s to %s: %v", appDir, extractDir, err)
			}

			// A class of systems (e.g:- RHEL 7.6) exhibit situations wherein
			// the bindmount volumes are not propagated to the child containers
			// Accommodate those systems as well, by performing local copies
			// for the locations that are resident in the host.
			// ref: https://github.com/containers/buildah/issues/1821
			for _, item := range volumeMaps {
				if strings.Contains(item, ":") {
					config.Debug.log("Appsody mount: ", item)
					var src = strings.Split(item, ":")[0]
					var dest = strings.Split(item, ":")[1]
					if strings.EqualFold(src, ".") {
						src = config.ProjectDir
					}
					dest = strings.Replace(dest, containerProjectDir, extractDir, -1)
					config.Debug.log("Local-adjusted mount destination: ", dest)

					destExists, err := Exists(dest)
					if err != nil {
						return errors.Errorf("Error checking file exists: %v", err)
					}
					if destExists {
						config.Debug.log("Deleting dest: ", dest)
						os.RemoveAll(dest)
					}

					mkdir := filepath.Dir(dest)
					config.Debug.Log("Running mkdir ", mkdir)
					err = os.MkdirAll(mkdir, os.ModePerm)
					if err != nil {
						return errors.Errorf("Error creating directories %s: %v", extractDir, err)
					}

					fileInfo, err := os.Lstat(src)
					if err != nil {
						return errors.Errorf("project file check error %v", err)
					}
					config.Debug.log("Copy source: ", src)
					config.Debug.log("Copy destination: ", dest)
					if fileInfo.IsDir() {
						err = CopyDir(config.LoggingConfig, src, dest)
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
	} else { // not buildah
		// Extract the files with `docker cp`
		cmdArgs = []string{"cp", appDir, extractDir}
		err = execAndWaitReturnErr(config.LoggingConfig, cmdName, cmdArgs, config.Debug, config.Dryrun)
		if err != nil {
			return errors.Errorf("docker cp command failed: %v", err)
		}
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

	depErr := GetDeprecated(config.RootCommandConfig)
	if depErr != nil {
		return depErr
	}
	return nil
}

func defaultExtractContainerName(config *RootCommandConfig) string {
	projectName, perr := getProjectName(config)

	if perr != nil {
		if _, ok := perr.(*NotAnAppsodyProject); !ok {
			config.Error.log("Error occurred retrieving project name... exiting: ", perr)
			os.Exit(1)
		}
	}
	return projectName + "-extract"
}
