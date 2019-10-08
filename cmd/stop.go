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

func newStopCmd(rootConfig *RootCommandConfig) *cobra.Command {
	var containerName string
	// stopCmd represents the stop command
	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops the local Appsody docker container for your project",
		Long: `Stop the local Appsody docker container for your project.

Stops the docker container specified by the --name flag. 
If --name is not specified, the container name is determined from the current working directory (see default below).
To see a list of all your running docker containers, run the command "docker ps". The name is in the last column.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if !rootConfig.Buildah {
				Info.log("Stopping development environment")
				err := dockerStop(containerName, rootConfig.Dryrun)
				if err != nil {
					return err
				}
				//dockerRemove(imageName) is not needed due to --rm flag
				//os.Exit(1)
			} else {
				// this is the k8s path, runs kubectl delete for the ingress, service and deployment
				// Note for k8s the containerName does not need -dev

				serviceArgs := []string{containerName + "-service"}
				deploymentArgs := []string{containerName + "-deployment"}
				ingressArgs := []string{containerName + "-ingres"}
				_, err := RunKubeDelete(ingressArgs, rootConfig.Dryrun)
				if err != nil {
					return err
				}

				_, err = RunKubeDelete(serviceArgs, rootConfig.Dryrun)
				if err != nil {
					return err
				}
				_, err = RunKubeDelete(deploymentArgs, rootConfig.Dryrun)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	addNameFlag(stopCmd, &containerName, rootConfig)
	return stopCmd
}
