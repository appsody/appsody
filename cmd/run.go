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
	"github.com/spf13/cobra"
)

func dockerStop(imageName string, dryrun bool) error {
	cmdName := "docker"
	signalInterval := "10" // numnrt of seconds to wait prior to sending SIGKILL
	cmdArgs := []string{"stop", imageName, "-t", signalInterval}
	err := execAndWait(cmdName, cmdArgs, Debug, dryrun)
	if err != nil {
		return err
	}
	return nil
}

func containerRemove(imageName string, buildah bool, dryrun bool) error {
	cmdName := "docker"
	//Added "-f" to force removal if container is still running or image has containers
	cmdArgs := []string{"rm", imageName, "-f"}
	if buildah {
		cmdName = "buildah"
		cmdArgs = []string{"rm", imageName}
	}
	err := execAndWait(cmdName, cmdArgs, Debug, dryrun)
	if err != nil {
		return err
	}
	return nil

}

func newRunCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &devCommonConfig{RootCommandConfig: rootConfig}
	// runCmd represents the run command
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run your project in the local Appsody environment.",
		Long: `Run the local Appsody environment, starting a container based continuous build environment for your project.
		
You must be in the base directory of your Appsody project when running this command.`,
		Example: `  appsody run --docker-volume my-volume
  Runs the local Appsody environment using the "my-volume" docker volume to cache your project dependencies.
  
  appsody run --interactive
  Runs the local Appsody environment with STDIN attached to the container, making the container start look like a terminal connection session.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			Info.log("Running development environment...")
			return commonCmd(config, "run")

		},
	}

	addDevCommonFlags(runCmd, config)
	return runCmd
}
