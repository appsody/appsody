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

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the local Appsody docker container for your project",
	Long: `Stop the local Appsody docker container for your project.

Stops the docker container specified by the --name flag. 
If --name is not specified, the container name is determined from the current working directory (see default below).
To see a list of all your running docker containers, run the command "docker ps". The name is in the last column.`,

	RunE: func(cmd *cobra.Command, args []string) error {

		Info.log("Stopping development environment")
		err := dockerStop(containerName)
		if err != nil {
			return err
		}
		//dockerRemove(imageName) is not needed due to --rm flag
		//os.Exit(1)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	addNameFlags(stopCmd)

}
