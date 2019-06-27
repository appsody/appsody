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

func dockerStop(imageName string) {
	cmdName := "docker"
	cmdArgs := []string{"stop", imageName}
	execAndWait(cmdName, cmdArgs, Debug)
}

func dockerRemove(imageName string) {
	cmdName := "docker"
	//Added "-f" to force removal if container is still running or image has containers
	cmdArgs := []string{"rm", imageName, "-f"}
	execAndWait(cmdName, cmdArgs, Debug)
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the local Appsody environment for your project",
	Long:  `This starts a docker based continuous build environment for your project.`,
	Run: func(cmd *cobra.Command, args []string) {

		Info.log("Running development environment...")
		commonCmd(cmd, args, "run")

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	addDevCommonFlags(runCmd)
}
