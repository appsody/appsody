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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

const operatorYamlName = "appsody-app-operator.yaml"
const appsodyCRDName = "appsody-app-crd.yaml"
const operatorRBACName = "appsody-app-cluster-rbac.yaml"

type operatorCommandConfig struct {
	*RootCommandConfig
	namespace string
}

func newOperatorCmd(rootConfig *RootCommandConfig) *cobra.Command {
	operatorConfig := &operatorCommandConfig{RootCommandConfig: rootConfig}
	var operatorCmd = &cobra.Command{
		Use:   "operator",
		Short: "Install or uninstall the Appsody operator from your Kubernetes cluster.",
		Long:  `This command allows you to "install" or "uninstall" the Appsody operator from the configured Kubernetes cluster. An installed Appsody operator is required to deploy your Appsody projects.`,
	}

	// rootCmd.AddCommand(operatorCmd)
	//operatorCmd.AddCommand(installCmd)
	//operatorCmd.AddCommand(uninstallCmd)
	operatorCmd.PersistentFlags().StringVarP(&operatorConfig.namespace, "namespace", "n", "default", "The namespace in which the operator will run.")
	//operatorCmd.PersistentFlags().StringVarP(&watchspace, "watchspace", "w", "''", "The namespace which the operator will watch. Use '' for all namespaces.")
	operatorCmd.AddCommand(newOperatorInstallCmd(operatorConfig))
	operatorCmd.AddCommand(newOperatorUninstallCmd(operatorConfig))
	return operatorCmd
}

func downloadOperatorYaml(log *LoggingConfig, url string, operatorNamespace string, watchNamespace string, target string) (string, error) {

	file, err := downloadYaml(log, url, target)
	if err != nil {
		return "", fmt.Errorf("Could not download Operator YAML file %s", url)
	}

	yamlReader, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.Errorf("Downloaded file does not exist %s. ", target)

		}
		return "", errors.Errorf("Failed reading file %s", target)

	}

	//output := bytes.Replace(yamlReader, []byte("APPSODY_OPERATOR_NAMESPACE"), []byte(operatorNamespace), -1)
	output := bytes.Replace(yamlReader, []byte("APPSODY_WATCH_NAMESPACE"), []byte(watchNamespace), -1)

	err = ioutil.WriteFile(target, output, 0666)
	if err != nil {
		return "", errors.Errorf("Failed to write local operator definition file: %s", err)
	}
	return target, nil
}

func downloadRBACYaml(log *LoggingConfig, url string, operatorNamespace string, target string, dryrun bool) (string, error) {
	if dryrun {
		log.Info.log("Skipping download of RBAC yaml: ", url)
		return "", nil

	}
	file, err := downloadYaml(log, url, target)
	if err != nil {
		return "", fmt.Errorf("Could not download RBAC YAML file %s", url)
	}

	yamlReader, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.Errorf("Downloaded file does not exist %s. ", target)

		}
		return "", errors.Errorf("Failed reading file %s", target)

	}

	output := bytes.Replace(yamlReader, []byte("APPSODY_OPERATOR_NAMESPACE"), []byte(operatorNamespace), -1)
	//output = bytes.Replace(output, []byte("APPSODY_WATCH_NAMESPACE"), []byte(watchNamespace), -1)

	err = ioutil.WriteFile(target, output, 0666)
	if err != nil {
		return "", errors.Errorf("Failed to write local operator definition file: %s", err)
	}
	return target, nil
}

func downloadYaml(log *LoggingConfig, url string, target string) (string, error) {
	log.Debug.log("Downloading file: ", url)

	fileBuffer := bytes.NewBuffer(nil)
	err := downloadFile(log, url, fileBuffer)
	if err != nil {
		return "", errors.Errorf("Failed to get file: %s", err)
	}

	yamlFile, err := ioutil.ReadAll(fileBuffer)
	if err != nil {
		return "", fmt.Errorf("Could not read buffer into byte array")
	}

	err = ioutil.WriteFile(target, yamlFile, 0666)
	if err != nil {
		return "", errors.Errorf("Failed to write local operator definition file: %s", err)
	}
	return target, nil
}

func downloadCRDYaml(log *LoggingConfig, url string, target string) (string, error) {
	file, err := downloadYaml(log, url, target)
	if err != nil {
		return "", fmt.Errorf("Could not download AppsodyApplication CRD file %s", url)
	}
	return file, nil
}

func getDeployConfigDir(rootConfig *RootCommandConfig) (string, error) {
	deployConfigDir := filepath.Join(getHome(rootConfig), "deploy")
	deployConfigDirExists, err := Exists(deployConfigDir)
	if err != nil {
		return "", errors.Errorf("Error checking directory: %v", err)
	}
	if !deployConfigDirExists {

		rootConfig.Debug.log("Creating deploy config dir: ", deployConfigDir)
		err = os.MkdirAll(deployConfigDir, os.ModePerm)
		if err != nil {
			return "", errors.Errorf("Error creating directories %s %v", deployConfigDir, err)
		}

	}
	return deployConfigDir, nil
}
