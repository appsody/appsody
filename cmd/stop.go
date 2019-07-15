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

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the local Appsody environment for your project",
	Long:  `If no name flag is specified, the value used is "directory-name-dev". By specifying the name flag, the corresponding development container will be stopped.  The value of name should either be "project-dir-dev" or the name specified in "appsody run".  You can find the list of running docker containers with the commmand: "docker container ls".  See the NAMES Column.`,
	Run: func(cmd *cobra.Command, args []string) {

		Info.log("Stopping development environment")
		dockerStop(containerName)
		//dockerRemove(imageName) is not needed due to --rm flag
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	addNameFlags(stopCmd)

}
