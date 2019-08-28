package cmd

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

//RunKubeGet issues kubectl get <arg>
func RunKubeGet(args []string) (string, error) {
	Info.log("Attempting to get resource from Kubernetes ...")
	kcmd := "kubectl"
	kargs := []string{"get"}
	kargs = append(kargs, args...)

	if dryrun {
		Info.log("Dry run - skipping execution of: ", kcmd, " ", kargs)
		return "", nil
	}
	Info.log("Running command: ", kcmd, kargs)
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := execCmd.Output()
	if kerr != nil {
		return "", errors.Errorf("kubectl get failed: %s", string(kout[:]))
	}
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
	var deploymentsWithOperatorsGetArgs = []string{"deployments", "-o=jsonpath='{.items[?(@.metadata.name==\"appsody-operator\")].metadata.namespace}'", "-A"}
	getOutput, getErr := RunKubeGet(deploymentsWithOperatorsGetArgs)
	if getErr != nil {
		return false, "", getErr
	}
	getOutput = strings.Trim(getOutput, "'")
	if getOutput == "" {
		Info.log("There are no deployments with appsody-operator")
		return false, "", nil
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

			Debug.logf("An operator exists in namespace %s, that is watching namespace: %s", deploymentNamespace, watchNamespace)
			return true, deploymentNamespace, nil
		}

	}
	return false, "", nil
}
func operatorCount() (int, error) {
	var getAllOperatorsArgs = []string{"deployments", "-o=jsonpath='{.items[?(@.metadata.name==\"appsody-operator\")].metadata.name}'", "-A"}
	getOutput, getErr := RunKubeGet(getAllOperatorsArgs)
	if getErr != nil {
		return 0, getErr
	}
	return strings.Count(getOutput, "appsody-operator"), nil
}
