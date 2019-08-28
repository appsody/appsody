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

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the "appsody deploy uninstall" command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Appsody Operator from the configured Kubernetes cluster",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initConfig()
		if err != nil {
			return err
		}

		deployConfigDir, err := getDeployConfigDir()
		if err != nil {
			return errors.Errorf("Error getting deploy config dir: %v", err)
		}

		appsodyCRD := filepath.Join(deployConfigDir, appsodyCRDName)

		// If appsody-app-crd.yaml exists, uninstall using it. Else download a new one
		// and uninstall using that.
		crdFileExists, err := exists(appsodyCRD)
		if err != nil {
			return errors.Errorf("Error checking file: %v", err)
		}
		if !crdFileExists {
			var crdURL = getOperatorHome() + "/" + appsodyCRDName
			_, err := downloadCRDYaml(crdURL, appsodyCRD)
			if err != nil {
				return err
			}
		}
		err = KubeDelete(appsodyCRD)
		if err != nil {
			return err
		}
		err = os.Remove(appsodyCRD)
		if err != nil {
			return err
		}

		operatorYaml := filepath.Join(deployConfigDir, operatorYamlName)

		yamlFileExists, err := exists(operatorYaml)
		if err != nil {
			return errors.Errorf("Error checking file: %v", err)
		}
		if !yamlFileExists {
			operatorNamespace := "default"
			watchNamespace := "''"
			if namespace != "" {
				operatorNamespace = namespace
			}
			if watchspace != "" {
				watchNamespace = watchspace
			}
			var operatorURL = getOperatorHome() + "/" + operatorYamlName
			_, err := downloadOperatorYaml(operatorURL, operatorNamespace, watchNamespace, operatorYaml)
			if err != nil {
				return err
			}
		}
		err = KubeDelete(operatorYaml)
		if err != nil {
			return err
		}
		err = os.Remove(operatorYaml)
		if err != nil {
			return err
		}

		Info.log("Appsody operator removed from Kubernetes")
		return nil
	},
}

func init() {

	operatorCmd.AddCommand(uninstallCmd)
}
