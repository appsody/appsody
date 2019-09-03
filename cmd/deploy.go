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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var configFile, namespace, watchspace, tag string
var generate, force, push bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Build and deploy your Appsody project to your Kubernetes cluster",
	Long: `This command extracts the code from your project, builds a local Docker image for deployment,
generates a deployment manifest (yaml) file if one is not present, and uses it to deploy your image to a Kubernetes cluster, either via the Appsody operator or as a Knative service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if generate {
			return generateDeploymentConfig()
		}
		// Check for the Appsody Operator

		operatorExists, existingNamespace, operatorExistsErr := operatorExistsWithWatchspace(namespace)
		if operatorExistsErr != nil {
			return operatorExistsErr
		}

		//kargs := []string{"service/appsody-operator"}
		//_, err := KubeGet(kargs)
		// Performing the kubectl apply
		if !operatorExists {
			Debug.logf("Failed to find Appsody operator that watches namespace %s. Attempting to install...", namespace)
			err := installCmd.RunE(cmd, args)
			if err != nil {
				return errors.Errorf("Failed to install an Appsody operator in namespace %s watching namespace %s. Error was: %v", namespace, namespace, err)
			}
		} else {
			Debug.logf("Operator exists in %s, watching %s ", existingNamespace, namespace)

		}

		exists, err := exists(configFile)
		if err != nil {
			return errors.Errorf("Error checking status of %s", configFile)
		}
		if !exists || (exists && force) {
			err = generateDeploymentConfig()
			if err != nil {
				if err.Error() == "docker cp command failed: exit status 1" {
					Warning.log("No deployment config is present in the stack. Falling back to default deploy config using Knative.")
					return deployWithKnative(cmd, args)
				}
				return err
			}
		}
		Info.log("Found existing deployment manifest ", configFile)

		// Extract code and build the image - and tags it if -t is specified
		buildErr := buildCmd.RunE(cmd, args)
		if buildErr != nil {
			return buildErr
		}

		//Retrieve the project name and lowercase it
		projectName, perr := getProjectName()
		if perr != nil {
			return errors.Errorf("%v", perr)
		}
		deployImage := projectName // if not tagged, this is the deploy image name
		if tag != "" {
			deployImage = tag //Otherwise, it's the tag
		}
		// Edit the deployment manifest to reflect the new tag
		yamlFile, err := os.Open(configFile)
		if !dryrun && err != nil {
			if os.IsNotExist(err) {
				return errors.Errorf("Config file does not exist %s. ", configFile)
			}
			return errors.Errorf("Failed reading file %s", configFile)
		}
		defer yamlFile.Close()
		scanner := bufio.NewScanner(yamlFile)
		scanner.Split(bufio.ScanLines)
		var txtlines []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "applicationImage:") {
				index := strings.Index(line, ": ")
				start := line[0:(index + 2)]
				line = start + deployImage
				Info.log("Using applicationImage of: ", deployImage)
			}
			txtlines = append(txtlines, line)
		}
		tempConfigFile := "temp-" + configFile
		file, err := os.Create(tempConfigFile)
		if err != nil {
			return err
		}

		defer closeAndRemoveFile(file, tempConfigFile)
		w := bufio.NewWriter(file)
		for _, line := range txtlines {
			fmt.Fprintln(w, line)
		}
		w.Flush()

		if push {
			err = DockerPush(deployImage)
			if err != nil {
				return errors.Errorf("Could not push the docker image - exiting. Error: %v", err)
			}
		}
		err = KubeApply(tempConfigFile)
		// Performing the kubectl apply
		if err != nil {
			return errors.Errorf("Failed to deploy to your Kubernetes cluster: %v", err)
		}
		Info.log("Deployment succeeded.")

		// Ensure hostname and IP config is set up for deployment
		time.Sleep(1 * time.Second)

		out, err := KubeGetDeploymentURL(projectName)
		// Performing the kubectl apply
		if err != nil {
			return errors.Errorf("Failed to find deployed service IP and Port: %s", err)
		}
		if !dryrun {
			Info.log("Deployed project running at ", out)
		} else {
			Info.log("Dry run complete")
		}

		return nil
	},
}

func closeAndRemoveFile(file *os.File, fileName string) {
	file.Close()
	Debug.log("Removing file: ", fileName)
	err := os.Remove(fileName)
	if err != nil {
		Warning.logf("Failed to remove %s.", fileName)
	}
}
func deployWithKnative(cmd *cobra.Command, args []string) error {
	buildErr := buildCmd.RunE(cmd, args)
	if buildErr != nil {
		return buildErr
	}
	//Generate the KNative yaml
	//Get the container port first
	port, err := getEnvVarInt("PORT")
	if err != nil {
		//try and get the exposed ports and use the first one
		Warning.log("Could not detect a container port (PORT env var).")
		portsStr, portsErr := getExposedPorts()
		if portsErr != nil {
			return portsErr
		}
		if len(portsStr) == 0 {
			//No ports exposed
			Warning.log("This container exposes no ports. The service will not be accessible.")
			port = 0 //setting this to 0
		} else {
			portStr := portsStr[0]
			Warning.log("Picking the first exposed port as the KNative service port. This may not be the correct port.")
			port, err = strconv.Atoi(portStr)
			if err != nil {
				Warning.log("The exposed port is not a valid integer. The service will not be accessible.")
				port = 0
			}
		}
	}
	//Get the KNative template file
	knativeTempl := getKNativeTemplate()

	//Retrieve the project name and lowercase it
	projectName, perr := getProjectName()
	if perr != nil {
		return errors.Errorf("%v", perr)
	}
	//Get the project name and make it the KNative service name
	serviceName := projectName
	deployImage := projectName // if not tagged, this is the deploy image name
	if tag != "" {
		deployImage = tag //Otherwise, it's the tag
	}
	// We're not pushing to a repository, so we need to use dev.local for Knative to be able to find it
	if !push {
		localtag := "dev.local/" + projectName
		// Tagging the image using the tag as the deployImage for KNative
		err = DockerTag(deployImage, localtag)
		if err != nil {
			return errors.Errorf("Tagging the image failed - exiting. Error: %v", err)
		}
		deployImage = localtag // And forcing deployimage to be localtag
	}
	//Generating the KNative yaml file
	Debug.logf("Calling GenKnativeYaml with parms: %s %d %s %s \n", knativeTempl, port, serviceName, deployImage)
	yamlFileName, err := GenKnativeYaml(knativeTempl, port, serviceName, deployImage, push)
	if err != nil {
		return errors.Errorf("Could not generate the KNative YAML file: %v", err)
	}
	Info.log("Generated KNative serving deploy file: ", yamlFileName)
	// Pushing the docker image if necessary
	if push {
		err = DockerPush(deployImage)
		if err != nil {
			return errors.Errorf("Could not push the docker image - exiting. Error: %v", err)
		}
	}
	err = KubeApply(yamlFileName)
	// Performing the kubectl apply
	if err != nil {
		return errors.Errorf("Failed to deploy to your Kubernetes cluster: %v", err)
	}
	Info.log("Deployment succeeded.")
	url, err := KubeGetKnativeURL(serviceName)
	if err != nil {
		return errors.Errorf("Failed to find deployed service in your Kubernetes cluster: %v", err)
	}
	Info.log("Your deployed service is available at the following URL: ", url)

	return nil
}

func generateDeploymentConfig() error {
	containerConfigDir := "/config/app-deploy.yaml"

	exists, err := exists(configFile)
	if err != nil {
		return errors.Errorf("Error checking status of %s", configFile)
	}
	if exists && !force {
		return errors.Errorf("Error, deploy config file %s already exists. Specify an alternative file using --file or using --force to overwrite.", configFile)
	}

	setupErr := setupConfig()
	if setupErr != nil {
		return setupErr
	}
	projectConfig, configErr := getProjectConfig()
	if configErr != nil {
		return configErr
	}
	err = CheckPrereqs()
	if err != nil {
		Warning.logf("Failed to check prerequisites: %v\n", err)
	}

	stackImage := projectConfig.Platform
	Debug.log("Stack image: ", stackImage)
	Debug.log("Config directory: ", containerConfigDir)

	var cmdName string
	var cmdArgs []string
	dockerPullErr := dockerPullImage(stackImage)
	if dockerPullErr != nil {
		return dockerPullErr
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := dockerStop(containerName)
		if err != nil {
			Error.log(err)
		}
		os.Exit(1)
	}()
	cmdName = "docker"
	var configDir string
	cmdArgs = []string{"--name", extractContainerName}

	cmdArgs = append([]string{"create"}, cmdArgs...)
	cmdArgs = append(cmdArgs, stackImage)
	err = execAndWaitReturnErr(cmdName, cmdArgs, Debug)
	if err != nil {

		Error.log("docker create command failed: ", err)
		removeErr := dockerRemove(extractContainerName)
		Error.log("Error in dockerRemove", removeErr)
		return err

	}
	configDir = extractContainerName + ":" + containerConfigDir

	cmdArgs = []string{"cp", configDir, "./" + configFile}
	err = execAndWaitReturnErr(cmdName, cmdArgs, Debug)
	if err != nil {
		Error.log("docker cp command failed: ", err)

		removeErr := dockerRemove(extractContainerName)
		if removeErr != nil {
			Error.log("dockerRemove error ", removeErr)
		}
		return errors.Errorf("docker cp command failed: %v", err)
	}

	removeErr := dockerRemove(extractContainerName)
	if removeErr != nil {
		Error.log("dockerRemove error ", removeErr)
	}

	yamlReader, err := ioutil.ReadFile(configFile)
	if !dryrun && err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("Config file does not exist %s. ", configFile)

		}
		return errors.Errorf("Failed reading file %s", configFile)

	}

	projectName, perr := getProjectName()
	if perr != nil {
		return errors.Errorf("%v", perr)
	}

	port, err := getEnvVarInt("PORT")
	if err != nil {
		//try and get the exposed ports and use the first one
		Warning.log("Could not detect a container port (PORT env var).")
		portsStr, portsErr := getExposedPorts()
		if portsErr != nil {
			return portsErr
		}
		if len(portsStr) == 0 {
			//No ports exposed
			Warning.log("This container exposes no ports. The service will not be accessible.")
			port = 0 //setting this to 0
		} else {
			portStr := portsStr[0]
			Warning.log("Picking the first exposed port as the KNative service port. This may not be the correct port.")
			port, err = strconv.Atoi(portStr)
			if err != nil {
				Warning.log("The exposed port is not a valid integer. The service will not be accessible.")
				port = 0
			}
		}
	}
	portStr := strconv.Itoa(port)

	split := strings.Split(stackImage, ":")
	stack := split[len(split)-2]
	split = strings.Split(stack, "/")
	stack = split[len(split)-1]
	if !dryrun {
		output := bytes.Replace(yamlReader, []byte("APPSODY_PROJECT_NAME"), []byte(projectName), -1)
		output = bytes.Replace(output, []byte("APPSODY_DOCKER_IMAGE"), []byte(projectName), -1)
		output = bytes.Replace(output, []byte("APPSODY_STACK"), []byte(stack), -1)
		output = bytes.Replace(output, []byte("APPSODY_PORT"), []byte(portStr), -1)

		err = ioutil.WriteFile(configFile, output, 0666)
		if err != nil {
			return errors.Errorf("Failed to write local application configuration file: %s", err)
		}
	} else {
		Info.logf("Dry run skipped construction of file %s", configFile)
	}
	Info.log("Created deployent manifest: ", configFile)
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.PersistentFlags().BoolVar(&generate, "generate-only", false, "Only generate the deployment configuration file. Do not deploy the project.")
	deployCmd.PersistentFlags().StringVarP(&configFile, "file", "f", "app-deploy.yaml", "The file name to use for the deployment configuration.")
	deployCmd.PersistentFlags().BoolVar(&force, "force", false, "Force the reuse of the deployment configuration file if one exists.")
	deployCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Target namespace in your Kubernetes cluster")
	deployCmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "Docker image name and optionally a tag in the 'name:tag' format")
	deployCmd.PersistentFlags().BoolVar(&push, "push", false, "Push this image to an external Docker registry. Assumes that you have previously successfully done docker login")

}
