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

var deployConfigFile string

var deleteDeploymentCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete your deployed Appsody project from a Kubernetes cluster",
	Long:  `This command deletes your deployed Appsody project from the configured Kubernetes cluster using your existing deployment manifest.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		exists, err := Exists(deployConfigFile)
		if err != nil {
			return errors.Errorf("Error checking status of %s", deployConfigFile)
		}
		if !dryrun && !exists {
			return errors.Errorf("Cannot delete deployment. Deployment manifest not found: %s", deployConfigFile)
		}

		Info.log("Deleting deployment using deployment manifest ", deployConfigFile)
		err = KubeDelete(deployConfigFile)
		if err != nil {
			return err
		}
		Info.log("Deployment deleted")
		return nil
	},
}

func init() {
	deployCmd.AddCommand(deleteDeploymentCmd)
	deleteDeploymentCmd.PersistentFlags().StringVarP(&deployConfigFile, "file", "f", "app-deploy.yaml", "The file name to use for the deployment configuration.")

}
