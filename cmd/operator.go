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

var operatorYamlName = "appsody-app-operator.yaml"
var appsodyCRDName = "appsody-app-crd.yaml"
var crdURL = "https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/0.1.0/appsody-app-crd.yaml"
var operatorURL = "https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/0.1.0/appsody-app-operator.yaml"

func downloadYaml(url string, target string) (string, error) {
	Debug.log("Downloading file: ", url)
	fileBuffer := bytes.NewBuffer(nil)
	err := downloadFile(url, fileBuffer)
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

func downloadOperatorYaml(url string, operatorNamespace string, watchNamespace string, target string) (string, error) {
	file, err := downloadYaml(url, target)
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

	output := bytes.Replace(yamlReader, []byte("APPSODY_OPERATOR_NAMESPACE"), []byte(operatorNamespace), -1)
	output = bytes.Replace(output, []byte("APPSODY_WATCH_NAMESPACE"), []byte(watchNamespace), -1)

	err = ioutil.WriteFile(target, output, 0666)
	if err != nil {
		return "", errors.Errorf("Failed to write local operator definition file: %s", err)
	}
	return target, nil
}

func downloadCRDYaml(url string, target string) (string, error) {
	file, err := downloadYaml(url, target)
	if err != nil {
		return "", fmt.Errorf("Could not download AppsodyApplication CRD file %s", url)
	}
	return file, nil
}

func getDeployConfigDir() (string, error) {
	deployConfigDir := filepath.Join(getHome(), "deploy")
	deployConfigDirExists, err := exists(deployConfigDir)
	if err != nil {
		return "", errors.Errorf("Error checking directory: %v", err)
	}
	if !deployConfigDirExists {
		if dryrun {
			Info.log("Dry Run - Skip creating deploy config dir: ", deployConfigDir)
		} else {
			Debug.log("Creating deploy config dir: ", deployConfigDir)
			err = os.MkdirAll(deployConfigDir, os.ModePerm)
			if err != nil {
				return "", errors.Errorf("Error creating directories %s %v", deployConfigDir, err)
			}
		}
	}
	return deployConfigDir, nil
}

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
		deployConfigDir, err := getDeployConfigDir()
		if err != nil {
			return errors.Errorf("Error getting deploy config dir: %v", err)
		}

		appsodyCRD := filepath.Join(deployConfigDir, appsodyCRDName)
		file, err := downloadCRDYaml(crdURL, appsodyCRD)
		if err != nil {
			return err
		}
		err = KubeApply(file)
		if err != nil {
			return err
		}

		operatorNamespace := "default"
		watchNamespace := "''"
		if operatorspace != "" {
			operatorNamespace = operatorspace
		}
		if watchspace != "" {
			watchNamespace = watchspace
		}

		operatorYaml := filepath.Join(deployConfigDir, operatorYamlName)
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

var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Install or uninstall the Appsody operator from your Kubernetes cluster.",
	Long:  `This command allows you to "install" or "uninstall" the Appsody operator from the configured Kubernetes cluster. An installed Appsody operator is required to deploy your Appsody projects.`,
}

func init() {
	rootCmd.AddCommand(operatorCmd)
	operatorCmd.AddCommand(installCmd)
	operatorCmd.AddCommand(uninstallCmd)
	operatorCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "The namespace in which the operator will run.")
	operatorCmd.PersistentFlags().StringVarP(&watchspace, "watchspace", "w", "''", "The namespace which the operator will watch. Use '' for all namespaces.")
}
