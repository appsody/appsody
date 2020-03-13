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
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"

	//"crypto/sha256"

	//"hash"

	"os"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

//ExtractDockerEnvFile returns a map with the env vars specified in docker env file
func ExtractDockerEnvFile(envFileName string) (map[string]string, error) {
	envVars := make(map[string]string)
	file, err := os.Open(envFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		equal := strings.Index(line, "=")
		hash := strings.Index(line, "#")
		if equal >= 0 && hash != 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				envVars[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {

		return nil, err
	}

	return envVars, nil
}

//ExtractDockerEnvVars returns a map with the env vars specified in docker options
func ExtractDockerEnvVars(dockerOptions string) (map[string]string, error) {
	//Check whether there's --env-file, this needs to be processed first
	var envVars map[string]string
	envFilePos := strings.Index(dockerOptions, "--env-file=")
	lenFlag := len("--env-file=")
	if envFilePos < 0 {
		envFilePos = strings.Index(dockerOptions, "--env-file")
		lenFlag = len("--env-file")
	}
	if envFilePos >= 0 {
		tokens := strings.Fields(dockerOptions[envFilePos+lenFlag:])
		if len(tokens) > 0 {
			var err error
			envVars, err = ExtractDockerEnvFile(tokens[0])
			if err != nil {
				return nil, err
			}
		}
	} else {
		envVars = make(map[string]string)
	}
	tokens := strings.Fields(dockerOptions)
	for idx, token := range tokens {
		nextToken := ""
		if token == "-e" || token == "--env" {
			if len(tokens) > idx+1 {
				nextToken = tokens[idx+1]
			}
		} else if strings.Contains(token, "-e=") || strings.Contains(token, "-env=") {
			posEqual := strings.Index(token, "=")
			nextToken = token[posEqual+1:]
		}
		if nextToken != "" && strings.Contains(nextToken, "=") {
			nextToken = strings.ReplaceAll(nextToken, "\"", "")
			nextToken = strings.ReplaceAll(nextToken, "'", "")
			//Note that Appsody doesn't support quotes in -e, use --env-file
			keyValuePair := strings.Split(nextToken, "=")
			if len(keyValuePair) > 1 {
				envVars[keyValuePair[0]] = keyValuePair[1]
			}
		}
	}
	return envVars, nil
}

//GetEnvVar obtains a Stack environment variable from the Stack image
func GetEnvVar(searchEnvVar string, config *RootCommandConfig) (string, error) {
	if config.cachedEnvVars == nil {
		config.cachedEnvVars = make(map[string]string)
	}

	if value, present := config.cachedEnvVars[searchEnvVar]; present {
		config.Debug.logf("Environment variable found cached: %s Value: %s", searchEnvVar, value)
		return value, nil
	}

	// Docker and Buildah produce slightly different output
	// for `inspect` command. Array of maps vs. maps
	var dataBuildah map[string]interface{}
	var dataDocker []map[string]interface{}
	projectConfig, projectConfigErr := getProjectConfig(config)
	if projectConfigErr != nil {
		return "", projectConfigErr
	}
	imageName := projectConfig.Stack
	pullErrs := pullImage(imageName, config)
	if pullErrs != nil {
		return "", pullErrs
	}

	inspectOut, inspectErr := inspectImage(imageName, config)
	if inspectErr != nil {
		return "", inspectErr
	}
	var err error
	var envVars []interface{}
	if config.Buildah {
		err = json.Unmarshal([]byte(inspectOut), &dataBuildah)
		if err != nil {
			return "", errors.New("error unmarshaling data from inspect command - exiting")
		}
		buildahConfig := dataBuildah["config"].(map[string]interface{})
		envVars = buildahConfig["Env"].([]interface{})

	} else {
		err = json.Unmarshal([]byte(inspectOut), &dataDocker)
		if err != nil {
			return "", errors.New("error unmarshaling data from inspect command - exiting")

		}
		dockerConfig := dataDocker[0]["Config"].(map[string]interface{})

		envVars = dockerConfig["Env"].([]interface{})
	}

	config.Debug.log("Number of environment variables in stack image: ", len(envVars))
	config.Debug.log("All environment variables in stack image: ", envVars)
	var varFound = false
	for _, envVar := range envVars {
		nameValuePair := strings.SplitN(envVar.(string), "=", 2)
		name, value := nameValuePair[0], nameValuePair[1]
		config.cachedEnvVars[name] = value
		if name == searchEnvVar {
			varFound = true
		}
	}
	if varFound {
		config.Debug.logf("Environment variable found: %s Value: %s", searchEnvVar, config.cachedEnvVars[searchEnvVar])
		return config.cachedEnvVars[searchEnvVar], nil
	}
	config.Debug.log("Could not find env var: ", searchEnvVar)
	return "", nil
}

func getEnvVarBool(searchEnvVar string, config *RootCommandConfig) (bool, error) {
	strVal, envErr := GetEnvVar(searchEnvVar, config)
	if envErr != nil {
		return false, envErr
	}
	return strings.Compare(strings.TrimSpace(strings.ToUpper(strVal)), "TRUE") == 0, nil
}

func getEnvVarInt(searchEnvVar string, config *RootCommandConfig) (int, error) {

	strVal, envErr := GetEnvVar(searchEnvVar, config)
	if envErr != nil {
		return 0, envErr
	}
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, err
	}
	return intVal, nil

}

func getDefaultStackRegistry(config *RootCommandConfig) string {
	defaultStackRegistry := config.CliConfig.Get("images").(string)
	if defaultStackRegistry == "" {
		config.Debug.Log("Appsody config file does not contain a default stack registry images property - setting it to docker.io")
		defaultStackRegistry = "docker.io"
	}
	config.Debug.Log("Default stack registry set to: ", defaultStackRegistry)
	return defaultStackRegistry
}

//GenDeploymentYaml generates a simple yaml for a plaing K8S deployment
func GenDeploymentYaml(log *LoggingConfig, appName string, imageName string, controllerImageName string, ports []string, pdir string, dockerMounts []string, dockerEnvVars map[string]string, depsMount string, dryrun bool) (fileName string, err error) {
	// Codewind workspace root dir constant
	codeWindWorkspace := "/"
	// Codewind project ID if provided
	codeWindProjectID := os.Getenv("CODEWIND_PROJECT_ID")
	// Codewind onwner ref name and uid
	codeWindOwnerRefName := os.Getenv("CODEWIND_OWNER_NAME")
	codeWindOwnerRefUID := os.Getenv("CODEWIND_OWNER_UID")
	// Deployment YAML structs
	type Port struct {
		Name          string `yaml:"name,omitempty"`
		ContainerPort int    `yaml:"containerPort"`
	}
	type EnvVar struct {
		Name  string `yaml:"name,omitempty"`
		Value string `yaml:"value,omitempty"`
	}
	type VolumeMount struct {
		Name      string `yaml:"name"`
		MountPath string `yaml:"mountPath"`
		SubPath   string `yaml:"subPath,omitempty"`
	}
	type Container struct {
		Args            []string  `yaml:"args,omitempty"`
		Command         []string  `yaml:"command,omitempty"`
		Env             []*EnvVar `yaml:"env,omitempty"`
		Image           string    `yaml:"image"`
		ImagePullPolicy string    `yaml:"imagePullPolicy,omitempty"`
		Name            string    `yaml:"name,omitempty"`
		Ports           []*Port   `yaml:"ports,omitempty"`
		SecurityContext struct {
			Privileged bool `yaml:"privileged"`
		} `yaml:"securityContext,omitempty"`
		VolumeMounts []VolumeMount `yaml:"volumeMounts"`
		WorkingDir   string        `yaml:"workingDir,omitempty"`
	}
	type InitContainer struct {
		Args            []string  `yaml:"args,omitempty"`
		Command         []string  `yaml:"command,omitempty"`
		Env             []*EnvVar `yaml:"env,omitempty"`
		Image           string    `yaml:"image"`
		ImagePullPolicy string    `yaml:"imagePullPolicy,omitempty"`
		Name            string    `yaml:"name,omitempty"`
		Ports           []*Port   `yaml:"ports,omitempty"`
		SecurityContext struct {
			Privileged bool `yaml:"privileged"`
		} `yaml:"securityContext,omitempty"`
		VolumeMounts []VolumeMount `yaml:"volumeMounts"`
		WorkingDir   string        `yaml:"workingDir,omitempty"`
	}
	type Volume struct {
		Name                  string `yaml:"name"`
		PersistentVolumeClaim struct {
			ClaimName string `yaml:"claimName"`
		} `yaml:"persistentVolumeClaim,omitempty"`
		EmptyDir struct {
			Medium string `yaml:"medium"`
		} `yaml:"emptyDir,omitempty"`
	}

	type Deployment struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name            string            `yaml:"name"`
			Namespace       string            `yaml:"namespace,omitempty"`
			Labels          map[string]string `yaml:"labels,omitempty"`
			OwnerReferences []OwnerReference  `yaml:"ownerReferences,omitempty"`
		} `yaml:"metadata"`
		Spec struct {
			Selector struct {
				MatchLabels map[string]string `yaml:"matchLabels"`
			} `yaml:"selector"`
			Replicas    int `yaml:"replicas"`
			PodTemplate struct {
				Metadata struct {
					Labels map[string]string `yaml:"labels"`
				} `yaml:"metadata"`
				Spec struct {
					ServiceAccountName string           `yaml:"serviceAccountName,omitempty"`
					InitContainers     []*InitContainer `yaml:"initContainers"`
					Containers         []*Container     `yaml:"containers"`
					Volumes            []*Volume        `yaml:"volumes"`
				} `yaml:"spec"`
			} `yaml:"template"`
		} `yaml:"spec"`
	}

	yamlMap := Deployment{}
	yamlTemplate := getDeploymentTemplate()
	err = yaml.Unmarshal([]byte(yamlTemplate), &yamlMap)
	if err != nil {
		log.Error.log("Could not create the YAML structure from template. Exiting.")
		return "", err
	}
	//Set the name
	yamlMap.Metadata.Name = appName

	//Set the codewind label if present
	if codeWindProjectID != "" {
		yamlMap.Metadata.Labels = make(map[string]string)
		yamlMap.Metadata.Labels["projectID"] = codeWindProjectID
	}
	//Set the owner ref if present
	if codeWindOwnerRefName != "" && codeWindOwnerRefUID != "" {
		yamlMap.Metadata.OwnerReferences = []OwnerReference{
			{
				APIVersion:         "apps/v1",
				BlockOwnerDeletion: true,
				Controller:         true,
				Kind:               "ReplicaSet",
				Name:               codeWindOwnerRefName,
				UID:                codeWindOwnerRefUID},
		}
	}
	//Set the service account if provided by an env var
	serviceAccount := os.Getenv("SERVICE_ACCOUNT_NAME")
	if serviceAccount != "" {
		log.Debug.Log("Detected service account name env var: ", serviceAccount)
		yamlMap.Spec.PodTemplate.Spec.ServiceAccountName = serviceAccount
	} else {
		log.Debug.log("No service account name env var, leaving the appsody-sa default")
	}
	//Set the controller image
	yamlMap.Spec.PodTemplate.Spec.InitContainers[0].Image = controllerImageName
	//Set the image
	yamlMap.Spec.PodTemplate.Spec.Containers[0].Name = appName
	yamlMap.Spec.PodTemplate.Spec.Containers[0].Image = imageName

	//Set the containerPort
	containerPorts := make([]*Port, 0)
	for i, port := range ports {
		//KNative only allows a single port entry
		if i == 0 {
			yamlMap.Spec.PodTemplate.Spec.Containers[0].Ports = containerPorts
		}
		log.Debug.Log("Adding port to yaml: ", port)
		newContainerPort := new(Port)
		newContainerPort.ContainerPort, err = strconv.Atoi(port)
		if err != nil {
			return "", err
		}
		yamlMap.Spec.PodTemplate.Spec.Containers[0].Ports = append(yamlMap.Spec.PodTemplate.Spec.Containers[0].Ports, newContainerPort)
	}
	//Set the env vars from docker run, if any
	if len(dockerEnvVars) > 0 {
		envVars := make([]*EnvVar, len(dockerEnvVars))
		idx := 0
		for key, value := range dockerEnvVars {
			envVars[idx] = &EnvVar{key, value}
			idx++
		}
		yamlMap.Spec.PodTemplate.Spec.Containers[0].Env = envVars
	}
	//Set the Pod release label to the container name
	yamlMap.Spec.PodTemplate.Metadata.Labels["release"] = appName
	//Set the workspace volume PVC
	workspaceVolumeName := "appsody-workspace"
	workspacePvcName := os.Getenv("PVC_NAME")
	if workspacePvcName == "" {
		workspacePvcName = "appsody-workspace"
	}

	workspaceVolume := Volume{Name: workspaceVolumeName}
	workspaceVolume.PersistentVolumeClaim.ClaimName = workspacePvcName
	volumeIdx := len(yamlMap.Spec.PodTemplate.Spec.Volumes)
	if volumeIdx < 1 {
		yamlMap.Spec.PodTemplate.Spec.Volumes = make([]*Volume, 1)
		yamlMap.Spec.PodTemplate.Spec.Volumes[0] = &workspaceVolume
	} else {
		yamlMap.Spec.PodTemplate.Spec.Volumes = append(yamlMap.Spec.PodTemplate.Spec.Volumes, &workspaceVolume)
	}
	//Set the code mounts
	//We need to iterate through the docker mounts
	volumeMounts := &yamlMap.Spec.PodTemplate.Spec.Containers[0].VolumeMounts
	for _, appsodyMount := range dockerMounts {
		if appsodyMount == "-v" {
			continue
		}
		appsodyMountComponents := strings.Split(appsodyMount, ":")
		targetMount := appsodyMountComponents[1]
		sourceMount, err := filepath.Rel(codeWindWorkspace, appsodyMountComponents[0])
		if err != nil {
			log.Debug.Log("Problem with the appsody mount: ", appsodyMountComponents[0])
			return "", err
		}

		sourceSubpath := filepath.Join(".", sourceMount)
		newVolumeMount := VolumeMount{"appsody-workspace", targetMount, sourceSubpath}
		log.Debug.Log("Appending volume mount: ", newVolumeMount)
		*volumeMounts = append(*volumeMounts, newVolumeMount)
	}

	//Set the deployment selector and pod label
	projectLabel := appName
	yamlMap.Spec.Selector.MatchLabels["app"] = projectLabel
	yamlMap.Spec.PodTemplate.Metadata.Labels["app"] = projectLabel

	log.Debug.logf("YAML map: \n%v\n", yamlMap)
	yamlStr, err := yaml.Marshal(&yamlMap)
	if err != nil {
		log.Error.log("Could not create the YAML string from Map. Exiting.")
		return "", err
	}
	log.Debug.logf("Generated YAML: \n%s\n", yamlStr)
	// Generate file based on supplied config, defaulting to app-deploy.yaml
	yamlFile := filepath.Join(pdir, "app-deploy.yaml")
	if dryrun {
		log.Info.log("Skipping creation of yaml file with prefix: ", yamlFile)
		return yamlFile, nil
	}
	err = ioutil.WriteFile(yamlFile, yamlStr, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not create the yaml file for deployment %v", err)
	}
	return yamlFile, nil
}
func getDeploymentTemplate() string {
	yamltempl := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: APPSODY_APP_NAME
spec:
  selector:
    matchLabels:
      app: appsody
  replicas: 1
  template:
    metadata:
      labels:
        app: appsody
    spec:
      serviceAccountName: appsody-sa
      initContainers:
      - name: init-appsody-controller
        image: appsody/appsody-controller
        resources: {}
        volumeMounts:
        - name: appsody-controller
          mountPath: /.appsody
        imagePullPolicy: IfNotPresent 
      containers:
      - name: APPSODY_APP_NAME
        image: APPSODY_STACK
        imagePullPolicy: Always
        command: ["/.appsody/appsody-controller"]
        volumeMounts:
        - name: appsody-controller
          mountPath: /.appsody
      volumes:
      - name: appsody-controller
        emptyDir: {}
`
	return yamltempl
}

//GenServiceYaml returns the file name of a generated K8S Service yaml
func GenServiceYaml(log *LoggingConfig, appName string, ports []string, pdir string, dryrun bool) (fileName string, err error) {

	type Port struct {
		Name       string `yaml:"name,omitempty"`
		Port       int    `yaml:"port"`
		TargetPort int    `yaml:"targetPort"`
	}
	type Service struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name            string            `yaml:"name"`
			Labels          map[string]string `yaml:"labels,omitempty"`
			OwnerReferences []OwnerReference  `yaml:"ownerReferences,omitempty"`
		} `yaml:"metadata"`
		Spec struct {
			Selector    map[string]string `yaml:"selector"`
			ServiceType string            `yaml:"type"`
			Ports       []Port            `yaml:"ports"`
		} `yaml:"spec"`
	}

	// Codewind project ID if provided
	codeWindProjectID := os.Getenv("CODEWIND_PROJECT_ID")
	// Codewind onwner ref name and uid
	codeWindOwnerRefName := os.Getenv("CODEWIND_OWNER_NAME")
	codeWindOwnerRefUID := os.Getenv("CODEWIND_OWNER_UID")

	var service Service

	service.APIVersion = "v1"
	service.Kind = "Service"
	service.Metadata.Name = fmt.Sprintf("%s-%s", appName, "service")

	//Set the release and projectID labels
	service.Metadata.Labels = make(map[string]string)
	service.Metadata.Labels["release"] = appName
	if codeWindProjectID != "" {
		service.Metadata.Labels["projectID"] = codeWindProjectID
	}

	//Set the owner ref if present
	if codeWindOwnerRefName != "" && codeWindOwnerRefUID != "" {
		service.Metadata.OwnerReferences = []OwnerReference{
			{
				APIVersion:         "apps/v1",
				BlockOwnerDeletion: true,
				Controller:         true,
				Kind:               "ReplicaSet",
				Name:               codeWindOwnerRefName,
				UID:                codeWindOwnerRefUID},
		}
	}

	service.Spec.Selector = make(map[string]string, 1)
	service.Spec.Selector["app"] = appName
	service.Spec.ServiceType = "NodePort"
	service.Spec.Ports = make([]Port, len(ports))
	for i, port := range ports {
		service.Spec.Ports[i].Name = fmt.Sprintf("port-%d", i)
		iPort, err := strconv.Atoi(port)
		if err != nil {
			return "", err
		}
		service.Spec.Ports[i].Port = iPort
		service.Spec.Ports[i].TargetPort = iPort
	}

	yamlStr, err := yaml.Marshal(&service)
	if err != nil {
		log.Error.log("Could not create the YAML string from Map. Exiting.")
		return "", err
	}
	log.Debug.logf("Generated YAML: \n%s\n", yamlStr)
	// Generate file based on supplied config, defaulting to app-deploy.yaml
	yamlFile := filepath.Join(pdir, "app-service.yaml")
	if dryrun {
		log.Info.log("Skipping creation of yaml file with prefix: ", yamlFile)
		return yamlFile, nil
	}
	err = ioutil.WriteFile(yamlFile, yamlStr, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not create the yaml file for the service %v", err)
	}
	return yamlFile, nil
}

//GenRouteYaml returns the file name of a generated K8S Service yaml
func GenRouteYaml(log *LoggingConfig, appName string, pdir string, port int, dryrun bool) (fileName string, err error) {
	type IngressPath struct {
		Path    string `yaml:"path"`
		Backend struct {
			ServiceName string `yaml:"serviceName"`
			ServicePort int    `yaml:"servicePort"`
		} `yaml:"backend"`
	}
	type IngressRule struct {
		Host string `yaml:"host"`
		HTTP struct {
			Paths []IngressPath `yaml:"paths"`
		} `yaml:"http"`
	}

	type Ingress struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec struct {
			Rules []IngressRule `yaml:"rules"`
		} `yaml:"spec"`
	}

	var ingress Ingress
	ingress.APIVersion = "extensions/v1beta1"
	ingress.Kind = "Ingress"
	ingress.Metadata.Name = fmt.Sprintf("%s-%s", appName, "ingress")

	ingress.Spec.Rules = make([]IngressRule, 1)
	//cheIngressHost := os.Getenv("CHE_INGRESS_HOST")
	//Ignore the CW variable for now
	ingressHost := ""
	if ingressHost != "" {
		ingress.Spec.Rules[0].Host = ingressHost
	} else {
		// We set it to a host name that's resolvable by nip.io
		ingress.Spec.Rules[0].Host = fmt.Sprintf("%s.%s.%s", appName, getK8sMasterIP(log, dryrun), "nip.io")
	}

	ingress.Spec.Rules[0].HTTP.Paths = make([]IngressPath, 1)
	ingress.Spec.Rules[0].HTTP.Paths[0].Path = "/"
	ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName = fmt.Sprintf("%s-%s", appName, "service")
	ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort = port

	yamlStr, err := yaml.Marshal(&ingress)
	if err != nil {
		log.Error.log("Could not create the YAML string from Map. Exiting.")
		return "", err
	}
	log.Debug.logf("Generated YAML: \n%s\n", yamlStr)
	// Generate file based on supplied config, defaulting to app-deploy.yaml
	yamlFile := filepath.Join(pdir, "app-ingress.yaml")
	if dryrun {
		log.Info.log("Skipping creation of yaml file with prefix: ", yamlFile)
		return yamlFile, nil
	}
	err = ioutil.WriteFile(yamlFile, yamlStr, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not create the yaml file for the route %v", err)
	}
	return yamlFile, nil
}

func getK8sMasterIP(log *LoggingConfig, dryrun bool) string {
	cmdParms := []string{"node", "--selector", "node-role.kubernetes.io/master", "-o", "jsonpath={.items[0].status.addresses[?(.type==\"InternalIP\")].address}"}
	ip, err := KubeGet(log, cmdParms, "", dryrun)
	if err == nil {
		return ip
	}
	log.Debug.log("Could not retrieve the master IP address - returning x.x.x.x: ", err)
	return "x.x.x.x"
}

func getIngressPort(config *RootCommandConfig) int {
	ports, err := getExposedPorts(config)

	knownHTTPPorts := []string{"80", "8080", "8008", "3000", "9080"}
	if err != nil {
		config.Debug.Log("Error trying to obtain the exposed ports: ", err)
		return 0
	}
	if len(ports) < 1 {
		config.Debug.log("Container doesn't expose any port - returning 0")
		return 0
	}
	iPort := 0
	for _, port := range ports {
		for _, knownPort := range knownHTTPPorts {
			if port == knownPort {
				iPort, err := strconv.Atoi(port)
				if err == nil {
					return iPort
				}
			}
		}
	}
	//If we haven't returned yet, there was no match
	//Pick the first port and return it
	config.Debug.Log("No known HTTP port detected, returning the first one on the list.")
	iPort, err = strconv.Atoi(ports[0])
	if err == nil {
		return iPort
	}
	config.Debug.Logf("Error converting port %s - returning 0: %v", ports[0], err)
	return 0
}

//KubeGet issues kubectl get <arg>
func KubeGet(log *LoggingConfig, args []string, namespace string, dryrun bool) (string, error) {
	log.Info.log("Attempting to get resource from Kubernetes ...")
	kcmd := "kubectl"
	kargs := []string{"get"}
	kargs = append(kargs, args...)
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return "", nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return "", errors.Errorf("kubectl get failed: %s", kout)
	}
	return kout, kerr
}
