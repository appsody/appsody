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
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

//RunKubeGet issues kubectl get <arg>
func RunKubeGet(args []string) (string, error) {
	Info.log("Attempting to get resource from Kubernetes ...")
	kargs := []string{"get"}
	kargs = append(kargs, args...)
	return RunKube(kargs)

}

//RunKubeDelete issues kubectl delete <args>
func RunKubeDelete(args []string) (string, error) {
	Info.log("Attempting to delete resource from Kubernetes ...")
	kargs := []string{"delete"}
	kargs = append(kargs, args...)
	return RunKube(kargs)
}

//RunKube runs a generic kubectl command
func RunKube(kargs []string) (string, error) {
	kcmd := "kubectl"
	if dryrun {
		Info.log("Dry run - skipping execution of: ", kcmd, " ", strings.Join(kargs, " "))
		return "", nil
	}
	Info.log("Running command: ", kcmd, " ", strings.Join(kargs, " "))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := execCmd.Output()

	if kerr != nil {
		return "", errors.Errorf("kubectl command failed: %s", string(kout[:]))
	}
	Debug.log("Command successful...")
	return string(kout[:]), nil
}

/*
func downloadOperatorYaml(url string, operatorNamespace string, watchNamespace string, target string) (string, error) {
	if dryrun {
		Info.log("Skipping download of operator yaml: ", url)
		return "", nil

	}
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

	//output := bytes.Replace(yamlReader, []byte("APPSODY_OPERATOR_NAMESPACE"), []byte(operatorNamespace), -1)
	output := bytes.Replace(yamlReader, []byte("APPSODY_WATCH_NAMESPACE"), []byte(watchNamespace), -1)

	err = ioutil.WriteFile(target, output, 0666)
	if err != nil {
		return "", errors.Errorf("Failed to write local operator definition file: %s", err)
	}
	return target, nil
}

func downloadRBACYaml(url string, operatorNamespace string, target string) (string, error) {
	if dryrun {
		Info.log("Skipping download of RBAC yaml: ", url)
		return "", nil

	}
	file, err := downloadYaml(url, target)
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
*/
func operatorExistsInNamespace(operatorNamespace string) (bool, error) {

	// check to see if this namespace already has an appsody-operator
	//var args = []string{"deployment", "appsody-operator", "-n", operatorNamespace}
	var args = []string{"deployments", "-o=jsonpath='{.items[?(@.metadata.name==\"appsody-operator\")].metadata.namespace}'", "-n", operatorNamespace}

	getOutput, getErr := RunKubeGet(args)
	if getErr != nil {
		Debug.log("Received an err: ", getErr)
		return false, getErr
	}
	getOutput = strings.Trim(getOutput, "'")
	if getOutput == "" {
		Info.log("There are no deployments with appsody-operator")
		return false, nil
	}
	return true, nil

}

// Check to see if any other operator is watching the watchNameSpace
func operatorExistsWithWatchspace(watchNamespace string) (bool, string, error) {
	Debug.log("Looking for an operator matching watchspace: ", watchNamespace)
	var deploymentsWithOperatorsGetArgs = []string{"deployments", "-o=jsonpath='{.items[?(@.metadata.name==\"appsody-operator\")].metadata.namespace}'", "--all-namespaces"}
	getOutput, getErr := RunKubeGet(deploymentsWithOperatorsGetArgs)
	if getErr != nil {
		return false, "", getErr
	}
	getOutput = strings.Trim(getOutput, "'")
	if getOutput == "" {
		Info.log("There are no deployments with appsody-operator")
		return false, "", nil
	}
	if watchNamespace == "" && getOutput != "" {
		watchAllErr := errors.Errorf("You specified --watch-all, but there are already instances of the appsody operator on the cluster")
		return true, "", watchAllErr
	}
	deployments := strings.Split(getOutput, " ")
	Debug.log("deployments with operators: ", deployments)
	for _, deploymentNamespace := range deployments {
		var getDeploymentWatchNamespaceArgs = []string{"deployment", "-o=jsonpath='{.items[?(@.metadata.name==\"appsody-operator\")].spec.template.spec.containers[0].env[?(@.name==\"WATCH_NAMESPACE\")].value}'", "-n", deploymentNamespace}
		getOutput, getErr = RunKubeGet(getDeploymentWatchNamespaceArgs)
		Debug.logf("Deployment: %s is watching namespace %s", deploymentNamespace, getOutput)
		if getErr != nil {
			return false, "", getErr
		}
		if strings.Trim(getOutput, "'") == watchNamespace {
			Debug.logf("An operator that is watching namespace %s already exists in namespace %s", watchNamespace, deploymentNamespace)
			return true, deploymentNamespace, nil
		}
		// the operator is watching all namespaces
		if strings.Trim(getOutput, "'") == "" {

			Info.logf("An operator exists in namespace %s, that is watching all namespaces", deploymentNamespace)
			return true, deploymentNamespace, nil
		}
	}
	return false, "", nil
}
func operatorCount() (int, error) {
	var getAllOperatorsArgs = []string{"deployments", "-o=jsonpath='{.items[?(@.metadata.name==\"appsody-operator\")].metadata.name}'", "--all-namespaces"}
	getOutput, getErr := RunKubeGet(getAllOperatorsArgs)
	if getErr != nil {
		return 0, getErr
	}
	return strings.Count(getOutput, "appsody-operator"), nil
}

func appsodyApplicationCount(namespace string) (int, error) {
	var getAppsodyAppsArgs = []string{"AppsodyApplication", "-o=jsonpath='{.items[*].kind}'"}
	if namespace == "" {
		getAppsodyAppsArgs = append(getAppsodyAppsArgs, "--all-namespaces")
	} else {
		getAppsodyAppsArgs = append(getAppsodyAppsArgs, "-n", namespace)
	}
	getOutput, getErr := RunKubeGet(getAppsodyAppsArgs)
	if getErr != nil {
		return 0, getErr
	}
	return strings.Count(getOutput, "AppsodyApplication"), nil
}

func deleteAppsodyApps(namespace string) (string, error) {
	var deleteAppsodyAppsArgs = []string{"AppsodyApplication", "--all"}
	if namespace != "" {
		deleteAppsodyAppsArgs = append(deleteAppsodyAppsArgs, "-n", namespace)
	}
	return RunKubeDelete(deleteAppsodyAppsArgs)

}

func getOperatorWatchspace(namespace string) (string, error) {
	operatorExists, existsErr := operatorExistsInNamespace(namespace)
	if existsErr != nil {
		return "", existsErr
	}
	if !operatorExists {
		return "", errors.Errorf("An appsody operator could not be found in namespace: %s", namespace)
	}
	var args = []string{"deployments", "-o=jsonpath='{.items[?(@.metadata.name==\"appsody-operator\")].spec.template.spec.containers[0].env[?(@.name==\"WATCH_NAMESPACE\")].value}'", "-n", namespace}

	getOutput, getErr := RunKubeGet(args)
	if getErr != nil {
		Debug.log("Received an err: ", getErr)
		return "", getErr
	}
	watchspace = strings.Trim(getOutput, "'")
	if watchspace == "" {
		Debug.log("This operator watches the entire cluster ")
	}
	return watchspace, nil
}
