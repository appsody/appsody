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
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

		operatorNamespace := "default"
		if namespace != "" {
			operatorNamespace = namespace
		}

		removeErr := removeOperator(operatorNamespace)
		if removeErr != nil {
			return removeErr
		}

		operCount, operCountErr := operatorCount()
		Debug.log("Appsody operator count is: ", operCount)
		if operCountErr != nil {
			return operCountErr
		}
		//If no more operators, remove the CRDs
		//
		if operCount == 0 {
			if err := removeOperatorCRDs(); err != nil {
				return err
			}
		}

		return nil
	},
}

func removeOperatorCRDs() error {
	deployConfigDir, err := getDeployConfigDir()
	if err != nil {
		return errors.Errorf("Error getting deploy config dir: %v", err)
	}
	appsodyCRD := filepath.Join(deployConfigDir, appsodyCRDName)
	//Download the CRD yaml
	var crdURL = getOperatorHome() + "/" + appsodyCRDName
	_, err = downloadCRDYaml(crdURL, appsodyCRD)
	if err != nil {
		return err

	}
	err = KubeDelete(appsodyCRD)
	if err != nil {
		return err
	}
	if !dryrun {
		err = os.Remove(appsodyCRD)
		if err != nil {
			return err
		}
	}
	return nil
}
func removeOperatorRBAC(operatorNamespace string) error {
	deployConfigDir, err := getDeployConfigDir()
	if err != nil {
		return errors.Errorf("Error getting deploy config dir: %v", err)
	}
	appsodyRBAC := filepath.Join(deployConfigDir, operatorRBACName)
	// Download the RBAC file
	var rbacURL = getOperatorHome() + "/" + operatorRBACName
	_, err = downloadRBACYaml(rbacURL, operatorNamespace, appsodyRBAC)
	if err != nil {
		return err

	}
	err = KubeDelete(appsodyRBAC)
	if err != nil {
		Debug.log("Error in KubeDelete: ", err)
		return err
	}
	if !dryrun {
		err = os.Remove(appsodyRBAC)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeOperator(operatorNamespace string) error {
	var watchNamespace string
	deployConfigDir, err := getDeployConfigDir()
	if err != nil {
		return errors.Errorf("Error getting deploy config dir: %v", err)
	}
	operatorYaml := filepath.Join(deployConfigDir, operatorYamlName)
	if !dryrun {
		watchNamespace, err = getOperatorWatchspace(operatorNamespace)
		Debug.logf("Operator is watching the '%s' namespace", watchNamespace)
		if err != nil {
			return err
		}
	} else {
		Info.log("Dry run - skipping execution of: getOperatorWatchspace(" + operatorNamespace + ")")
	}
	// If there are running apps...
	appsCount, err := appsodyApplicationCount(watchNamespace)
	if err != nil {
		return errors.Errorf("Could not determine if there are AppsodyApplication instances: %v", err)
	}
	if appsCount > 0 {
		if force {
			deleteOut, err := deleteAppsodyApps(watchNamespace)
			if err != nil {
				return errors.Errorf("Could not remove appsody apps: %v %s", err, deleteOut)
			}
		} else {
			Debug.log("There are outstanding appsody applications for this operator - resubmit the command with --force if you want to remove them.")
			return errors.Errorf("There are outstanding appsody applications for this operator - resubmit the command with --force if you want to remove them.")
		}
	}

	// If the operator is watching a different namespace, remove RBACs
	if watchspace != operatorNamespace {
		if err := removeOperatorRBAC(operatorNamespace); err != nil {
			Debug.logf("Error from removeOperatorRBAC: %s", fmt.Sprintf("%v", err))
			if !strings.Contains(fmt.Sprintf("%v", err), "(NotFound)") {
				return err
			}
		}
	}

	var operatorURL = getOperatorHome() + "/" + operatorYamlName
	_, err = downloadOperatorYaml(operatorURL, operatorNamespace, watchNamespace, operatorYaml)
	if err != nil {
		return err
	}

	err = KubeDelete(operatorYaml)
	if err != nil {
		return err
	}
	if !dryrun {
		err = os.Remove(operatorYaml)
		if err != nil {
			return err
		}
	}

	Info.log("Appsody operator removed from Kubernetes")
	return nil
}

func init() {

	operatorCmd.AddCommand(uninstallCmd)
	uninstallCmd.PersistentFlags().BoolVar(&force, "force", false, "Force removal of appsody apps if present")
}
