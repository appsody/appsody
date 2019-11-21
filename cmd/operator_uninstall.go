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

type operatorUninstallCommandConfig struct {
	*operatorCommandConfig
	force bool
}

func newOperatorUninstallCmd(operatorConfig *operatorCommandConfig) *cobra.Command {
	config := &operatorUninstallCommandConfig{operatorCommandConfig: operatorConfig}
	// uninstallCmd represents the "appsody deploy uninstall" command
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the Appsody Operator from the configured Kubernetes cluster",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {

			operatorNamespace := "default"
			if config.namespace != "" {
				operatorNamespace = config.namespace
			}

			removeErr := removeOperator(operatorNamespace, config)
			if removeErr != nil {
				return removeErr
			}

			operCount, operCountErr := operatorCount(config.LoggingConfig, config.Dryrun)
			config.Debug.log("Appsody operator count is: ", operCount)
			if operCountErr != nil {
				return operCountErr
			}
			//If no more operators, remove the CRDs
			//
			if operCount == 0 {
				if err := removeOperatorCRDs(config); err != nil {
					return err
				}
			}

			return nil
		},
	}

	uninstallCmd.PersistentFlags().BoolVar(&config.force, "force", false, "Force removal of appsody apps if present")
	return uninstallCmd
}

func removeOperatorCRDs(config *operatorUninstallCommandConfig) error {
	deployConfigDir, err := getDeployConfigDir(config.RootCommandConfig)
	if err != nil {
		return errors.Errorf("Error getting deploy config dir: %v", err)
	}
	appsodyCRD := filepath.Join(deployConfigDir, appsodyCRDName)
	//Download the CRD yaml
	var crdURL = getOperatorHome(config.RootCommandConfig) + "/" + appsodyCRDName
	_, err = downloadCRDYaml(config.LoggingConfig, crdURL, appsodyCRD)
	if err != nil {
		return err

	}
	err = KubeDelete(config.LoggingConfig, appsodyCRD, config.namespace, config.Dryrun)
	if err != nil {
		return err
	}
	if !config.Dryrun {
		err = os.Remove(appsodyCRD)
		if err != nil {
			return err
		}
	}
	return nil
}
func removeOperatorRBAC(operatorNamespace string, config *operatorUninstallCommandConfig) error {
	deployConfigDir, err := getDeployConfigDir(config.RootCommandConfig)
	if err != nil {
		return errors.Errorf("Error getting deploy config dir: %v", err)
	}
	appsodyRBAC := filepath.Join(deployConfigDir, operatorRBACName)
	// Download the RBAC file
	var rbacURL = getOperatorHome(config.RootCommandConfig) + "/" + operatorRBACName
	_, err = downloadRBACYaml(config.LoggingConfig, rbacURL, operatorNamespace, appsodyRBAC, config.Dryrun)
	if err != nil {
		return err

	}
	err = KubeDelete(config.LoggingConfig, appsodyRBAC, config.namespace, config.Dryrun)
	if err != nil {
		config.Debug.log("Error in KubeDelete: ", err)
		return err
	}
	if !config.Dryrun {
		err = os.Remove(appsodyRBAC)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeOperator(operatorNamespace string, config *operatorUninstallCommandConfig) error {
	var watchNamespace string
	deployConfigDir, err := getDeployConfigDir(config.RootCommandConfig)
	if err != nil {
		return errors.Errorf("Error getting deploy config dir: %v", err)
	}
	operatorYaml := filepath.Join(deployConfigDir, operatorYamlName)
	if !config.Dryrun {
		watchNamespace, err = getOperatorWatchspace(config.LoggingConfig, operatorNamespace, config.Dryrun)
		config.Debug.logf("Operator is watching the '%s' namespace", watchNamespace)
		if err != nil {
			return err
		}
	} else {
		config.Info.log("Dry run - skipping execution of: getOperatorWatchspace(" + operatorNamespace + ")")
	}

	watchSpaces := getWatchSpaces(watchNamespace, config.Dryrun)
	if watchSpaces == nil {
		watchSpaces = append(watchSpaces, "")
	}
	for _, currentWatchSpace := range watchSpaces {
		// If there are running apps...
		appsCount, err := appsodyApplicationCount(config.LoggingConfig, currentWatchSpace, config.Dryrun)
		if err != nil {
			return errors.Errorf("Could not determine if there are AppsodyApplication instances: %v", err)
		}
		if appsCount > 0 {
			if config.force {
				deleteOut, err := deleteAppsodyApps(config.LoggingConfig, currentWatchSpace, config.Dryrun)
				if err != nil {
					return errors.Errorf("Could not remove appsody apps: %v %s", err, deleteOut)
				}
			} else {
				config.Debug.log("There are outstanding appsody applications for this operator - resubmit the command with --force if you want to remove them.")
				return errors.Errorf("There are outstanding appsody applications for this operator - resubmit the command with --force if you want to remove them.")
			}
		}

	}
	// If the operator is watching a different namespace, remove RBACs
	for _, currentWatchSpace := range watchSpaces {
		if currentWatchSpace != operatorNamespace {
			if err := removeOperatorRBAC(operatorNamespace, config); err != nil {
				config.Debug.logf("Error from removeOperatorRBAC: %s", fmt.Sprintf("%v", err))
				if !strings.Contains(fmt.Sprintf("%v", err), "(NotFound)") {
					return err
				}
			}
			break
		}
	}

	var operatorURL = getOperatorHome(config.RootCommandConfig) + "/" + operatorYamlName
	_, err = downloadOperatorYaml(config.LoggingConfig, operatorURL, operatorNamespace, watchNamespace, operatorYaml)
	if err != nil {
		return err
	}

	err = KubeDelete(config.LoggingConfig, operatorYaml, config.namespace, config.Dryrun)
	if err != nil {
		return err
	}
	if !config.Dryrun {
		err = os.Remove(operatorYaml)
		if err != nil {
			return err
		}
	}

	config.Info.log("Appsody operator removed from Kubernetes")
	return nil
}
