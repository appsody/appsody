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
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var all bool

// installCmd represents the "appsody deploy install" command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Appsody Operator into the configured Kubernetes cluster",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initConfig()
		if err != nil {
			return err
		}

		operatorNamespace := "default"
		watchNamespace := "''"
		if namespace != "" {
			operatorNamespace = namespace
		}
		if watchspace == "" {
			watchNamespace = operatorNamespace
		}
		if watchspace != "" {
			watchNamespace = watchspace
		}
		if all {
			watchNamespace = ""
		}
		Debug.log("watchNamespace is:  ", watchNamespace)
		operatorExists, existsErr := operatorExistsInNamespace(operatorNamespace)
		if existsErr != nil {
			return existsErr
		}
		if operatorExists {
			existingOperatorWatchspace, err := getOperatorWatchspace(operatorNamespace)
			if err != nil {
				Debug.log("Could not retrieve the watchspace of this operator - this should never happen...")
			}
			if existingOperatorWatchspace == "" {
				existingOperatorWatchspace = "all namespaces"
			}
			return errors.Errorf("An operator already exists in namespace %s and it is watching the %s namespace.", operatorNamespace, existingOperatorWatchspace)
		}

		watchExists, existingNamespace, watchExistsErr := operatorExistsWithWatchspace(watchNamespace)
		if watchExistsErr != nil {

			return watchExistsErr
		}
		if watchExists {
			return errors.Errorf("An operator watching namespace %s or all namespaces already exists in namespace %s", watchNamespace, existingNamespace)
		}

		deployConfigDir, err := getDeployConfigDir()
		if err != nil {
			return errors.Errorf("Error getting deploy config dir: %v", err)
		}

		var crdURL = getOperatorHome() + "/" + appsodyCRDName
		appsodyCRD := filepath.Join(deployConfigDir, appsodyCRDName)
		var file string

		file, err = downloadCRDYaml(crdURL, appsodyCRD)
		if err != nil {
			return err
		}

		err = KubeApply(file)
		if err != nil {
			return err
		}
		rbacYaml := filepath.Join(deployConfigDir, operatorRBACName)
		var rbacURL = getOperatorHome() + "/" + operatorRBACName
		if (operatorNamespace != watchNamespace) || all {
			Debug.log("Downloading: ", rbacURL)
			file, err = downloadRBACYaml(rbacURL, operatorNamespace, rbacYaml)
			if err != nil {
				return err
			}

			err = KubeApply(file)
			if err != nil {
				return err
			}
		}

		operatorYaml := filepath.Join(deployConfigDir, operatorYamlName)
		var operatorURL = getOperatorHome() + "/" + operatorYamlName
		file, err = downloadOperatorYaml(operatorURL, operatorNamespace, watchNamespace, operatorYaml)
		if err != nil {
			return err
		}

		err = KubeApply(file)
		if err != nil {
			return err
		}

		Info.log("Appsody operator deployed to Kubernetes")
		return nil
	},
}

func init() {
	operatorCmd.AddCommand(installCmd)
	installCmd.PersistentFlags().StringVarP(&watchspace, "watchspace", "w", "", "The namespace which the operator will watch.")
	installCmd.PersistentFlags().BoolVar(&all, "watch-all", false, "The operator will watch all namespaces.")

}
