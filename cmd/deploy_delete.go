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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newDeleteDeploymentCmd(deployConfig *deployCommandConfig) *cobra.Command {
	var deployConfigFile string
	var deleteDeploymentCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete your deployed Appsody project from a Kubernetes cluster.",
		Long:  `Delete your deployed Appsody project from the configured Kubernetes cluster using your existing deployment manifest.

By default, the command looks for the deployed project in the "default" namespace, and uses the generated "app-deploy.yaml" deployment manifest, unless you specify otherwise.`,
		Example:`  appsody deploy delete -f my-deploy.yaml
  Deletes the pod using the type and name specified in the "my-deploy.yaml" deployment manifest, in the "default" namespace.
  
  appsody deploy delete --namespace my-namespace
  Deletes the pod using the type and name specified in the "app-deploy.yaml" deployment manifest, in the "my-namespace" namespace.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			exists, err := Exists(deployConfigFile)
			if err != nil {
				return errors.Errorf("Error checking status of %s", deployConfigFile)
			}
			if !deployConfig.Dryrun && !exists {
				return errors.Errorf("Cannot delete deployment. Deployment manifest not found: %s", deployConfigFile)
			}

			Info.log("Deleting deployment using deployment manifest ", deployConfigFile)
			err = KubeDelete(deployConfigFile, deployConfig.namespace, deployConfig.Dryrun)
			if err != nil {
				return err
			}
			Info.log("Deployment deleted")
			return nil
		},
	}

	deleteDeploymentCmd.PersistentFlags().StringVarP(&deployConfigFile, "file", "f", "app-deploy.yaml", "Name of the deployment configuration you want to use.")
	return deleteDeploymentCmd
}
