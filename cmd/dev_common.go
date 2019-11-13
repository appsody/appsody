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
	"os/exec"
	"os/signal"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

type devCommonConfig struct {
	*RootCommandConfig
	disableWatcher  bool
	containerName   string
	depsVolumeName  string
	ports           []string
	publishAllPorts bool
	interactive     bool
	dockerNetwork   string
	dockerOptions   string
}

func checkDockerRunOptions(options []string) error {
	fmt.Println("testing docker options", options)
	//runOptionsTest := "(^((-p)|(--publish)|(--publish-all)|(-P)|(-u)|(--user)|(--name)|(--network)|(-t)|(--tty)|(--rm)|(--entrypoint)|(-v)|(--volume)|(-e)|(--env))((=?$)|(=.*)))"
	runOptionsTest := "(^((--help)|(-p)|(--publish)|(--publish-all)|(-P)|(-u)|(--user)|(--name)|(--network)|(-t)|(--tty)|(--rm)|(--entrypoint)|(-v)|(--volume))((=?$)|(=.*)))"

	blackListedRunOptionsRegexp := regexp.MustCompile(runOptionsTest)
	for _, value := range options {
		isInBlackListed := blackListedRunOptionsRegexp.MatchString(value)
		if isInBlackListed {
			return errors.Errorf("%s is not allowed in --docker-options", value)

		}
	}
	return nil

}

func addNameFlag(cmd *cobra.Command, flagVar *string, config *RootCommandConfig) {
	projectName, perr := getProjectName(config)
	if perr != nil {
		if _, ok := perr.(*NotAnAppsodyProject); ok {
			//Debug.log("Cannot retrieve the project name - continuing: ", perr)
		} else {
			Error.logf("Error occurred retrieving project name... exiting: %s", perr)
			os.Exit(1)
		}
	}

	defaultName := projectName + "-dev"
	cmd.PersistentFlags().StringVar(flagVar, "name", defaultName, "Assign a name to your development container.")
}

func addDevCommonFlags(cmd *cobra.Command, config *devCommonConfig) {

	projectName, perr := getProjectName(config.RootCommandConfig)
	if perr != nil {
		if _, ok := perr.(*NotAnAppsodyProject); ok {
			// Debug.log("Cannot retrieve the project name - continuing: ", perr)
		} else {
			Error.logf("Error occurred retrieving project name... exiting: %s", perr)
			os.Exit(1)
		}
	}
	defaultDepsVolume := projectName + "-deps"

	addNameFlag(cmd, &config.containerName, config.RootCommandConfig)
	cmd.PersistentFlags().StringVar(&config.dockerNetwork, "network", "", "Specify the network for docker to use.")
	cmd.PersistentFlags().StringVar(&config.depsVolumeName, "deps-volume", defaultDepsVolume, "Docker volume to use for dependencies. Mounts to APPSODY_DEPS dir.")
	cmd.PersistentFlags().StringArrayVarP(&config.ports, "publish", "p", nil, "Publish the container's ports to the host. The stack's exposed ports will always be published, but you can publish addition ports or override the host ports with this option.")
	cmd.PersistentFlags().BoolVarP(&config.publishAllPorts, "publish-all", "P", false, "Publish all exposed ports to random ports")
	cmd.PersistentFlags().BoolVar(&config.disableWatcher, "no-watcher", false, "Disable file watching, regardless of container environment variable settings.")
	cmd.PersistentFlags().BoolVarP(&config.interactive, "interactive", "i", false, "Attach STDIN to the container for interactive TTY mode")
	cmd.PersistentFlags().StringVar(&config.dockerOptions, "docker-options", "", "Specify the docker run options to use.  Value must be in \"\".")

}

func commonCmd(config *devCommonConfig, mode string) error {
	// Checking whether the controller is being overridden
	overrideControllerImage := os.Getenv("APPSODY_CONTROLLER_IMAGE")
	if overrideControllerImage == "" {
		overrideVersion := os.Getenv("APPSODY_CONTROLLER_VERSION")
		if overrideVersion != "" {
			CONTROLLERVERSION = overrideVersion
			Warning.Log("You have overridden the Appsody controller version and set it to: ", CONTROLLERVERSION)
		}
	} else {
		Warning.Log("The Appsody CLI detected the APPSODY_CONTROLLER_IMAGE env var. The controller image that will be used is: ", overrideControllerImage)
		imageSplit := strings.Split(overrideControllerImage, ":")
		if len(imageSplit) == 1 {
			// this is an implicit reference to latest
			CONTROLLERVERSION = "latest"
		} else {
			CONTROLLERVERSION = imageSplit[1]
			//This also could be latest
		}
	}
	if CONTROLLERVERSION == "latest" {
		Warning.Log("The Appsody CLI will use the latest version of the controller image. This may result in a mismatch or malfunction.")
	}
	projectDir, perr := getProjectDir(config.RootCommandConfig)
	if perr != nil {
		return perr

	}
	projectConfig, configErr := getProjectConfig(config.RootCommandConfig)
	if configErr != nil {
		return configErr
	}
	err := CheckPrereqs()
	if err != nil {
		Warning.logf("Failed to check prerequisites: %v\n", err)
	}

	platformDefinition := projectConfig.Stack
	Debug.log("Stack image: ", platformDefinition)
	Debug.log("Project directory: ", projectDir)

	var cmdArgs []string
	pullErr := pullImage(platformDefinition, config.RootCommandConfig)
	if pullErr != nil {
		return pullErr
	}

	volumeMaps, volumeErr := getVolumeArgs(config.RootCommandConfig)
	if volumeErr != nil {
		return volumeErr
	}
	// Mount the APPSODY_DEPS cache volume if it exists
	depsEnvVar, envErr := GetEnvVar("APPSODY_DEPS", config.RootCommandConfig)
	if envErr != nil {
		return envErr
	}
	if depsEnvVar != "" {
		depsMount := config.depsVolumeName + ":" + depsEnvVar
		Debug.log("Adding dependency cache to volume mounts: ", depsMount)
		volumeMaps = append(volumeMaps, "-v", depsMount)
	}

	// Mount the controller

	controllerImageName := overrideControllerImage
	if controllerImageName == "" {
		controllerImageName = fmt.Sprintf("%s:%s", "appsody/init-controller", CONTROLLERVERSION)
	}
	controllerVolumeName := fmt.Sprintf("%s-%s", "appsody-controller", CONTROLLERVERSION)
	controllerVolumeMount := fmt.Sprintf("%s:%s", controllerVolumeName, "/.appsody")

	if !config.Buildah {
		//In local mode, run the init-controller image if necessary
		updateController := false
		if CONTROLLERVERSION == "latest" {
			updateController = true
		} else {
			volNames, err := RunDockerVolumeList(controllerVolumeName)
			if err != nil {
				Debug.Log("Error attempting to query volumes for ", controllerVolumeName, " :", err)
				return err
			}
			Debug.Log("Retrieved volume name(s): [", volNames, "]")
			volNamesSlice := strings.Split(volNames, "\n")
			foundVolName := ""
			for _, volName := range volNamesSlice {
				if volName == controllerVolumeName {
					foundVolName = volName
				}
			}

			if foundVolName == "" {
				updateController = true
			}
		}
		// We run the image if there no volume that matches the controller version or if the controller version is "latest"
		if updateController {
			Debug.Logf("Controller volume not found or version is latest - launching the %s image to populate it", controllerImageName)
			downloaderArgs := []string{"--rm", "-v", controllerVolumeMount, controllerImageName}
			controllerDownloader, err := DockerRunAndListen(downloaderArgs, Info, false, config.RootCommandConfig.Verbose, config.RootCommandConfig.Dryrun)
			if config.Dryrun {
				Info.log("Dry Run - Skipping execCmd.Wait")
			} else {
				if err == nil {
					err = controllerDownloader.Wait()
				}
			}
			if err != nil {
				Debug.Log("Error populating the controller volume: ", err)
				return err
			}
		}
	}

	//controllerMount := controllerVolumeName + ":/appsody"
	Debug.log("Adding controller to volume mounts: ", controllerVolumeMount)
	volumeMaps = append(volumeMaps, "-v", controllerVolumeMount)
	if !config.Buildah {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			<-c
			// note we still need this signaling block otherwise the signal is not caught and termination doesn't occur propertly
			//containerRemove(containerName) is not needed due to --rm flag
		}()
	}
	cmdArgs = []string{"--rm"}
	validPorts, portError := checkPortInput(config.ports)
	if !validPorts {
		return errors.Errorf("Ports provided as input to the command are not valid: %v\n", portError)
	}
	var portsErr error
	cmdArgs, portsErr = processPorts(cmdArgs, config)
	if portsErr != nil {
		return portsErr
	}
	cmdArgs = append(cmdArgs, "--name", config.containerName)
	if config.dockerNetwork != "" {
		cmdArgs = append(cmdArgs, "--network", config.dockerNetwork)
	}
	runAsLocal, boolErr := getEnvVarBool("APPSODY_USER_RUN_AS_LOCAL", config.RootCommandConfig)
	if boolErr != nil {
		return boolErr
	}
	if runAsLocal && runtime.GOOS != "windows" {
		current, _ := user.Current()
		cmdArgs = append(cmdArgs, "-u", fmt.Sprintf("%s:%s", current.Uid, current.Gid))
		cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("APPSODY_USER=%s", current.Uid), "-e", fmt.Sprintf("APPSODY_GROUP=%s", current.Gid))
	}

	if len(volumeMaps) > 0 {
		cmdArgs = append(cmdArgs, volumeMaps...)
	}
	if config.dockerOptions != "" {
		Debug.logf("User provided Docker options: \"%s\"", config.dockerOptions)
		dockerOptions := config.dockerOptions
		dockerOptions = strings.TrimPrefix(dockerOptions, " ")
		dockerOptions = strings.TrimSuffix(dockerOptions, " ")
		dockerOptionsCmd := strings.Split(dockerOptions, " ")
		err := checkDockerRunOptions(dockerOptionsCmd)
		if err != nil {
			return err
		}
		cmdArgs = append(cmdArgs, dockerOptionsCmd...)
	}
	if config.interactive {
		cmdArgs = append(cmdArgs, "-i")
	}
	cmdArgs = append(cmdArgs, "-t", "--entrypoint", "/.appsody/appsody-controller", platformDefinition, "--mode="+mode)
	if config.Verbose {
		cmdArgs = append(cmdArgs, "-v")
	}
	if config.disableWatcher {
		cmdArgs = append(cmdArgs, "--no-watcher")
	}
	if !config.Buildah {
		Debug.logf("Attempting to start image %s with container name %s", platformDefinition, config.containerName)
		execCmd, err := DockerRunAndListen(cmdArgs, Container, config.interactive, config.Verbose, config.Dryrun)
		if config.Dryrun {
			Info.log("Dry Run - Skipping execCmd.Wait")
		} else {
			if err == nil {
				err = execCmd.Wait()
			}
		}
		if err != nil {
			// 'signal: interrupt'
			// TODO presumably you can query the error itself
			error := fmt.Sprintf("%s", err)
			Debug.log("CLI exit error is:  ", error)
			//Linux and Windows return a different error on Ctrl-C
			if error == "signal: interrupt" || error == "signal: terminated" || error == "exit status 2" {
				Info.log("Closing down, development environment was interupted will now sleep 10 seconds.")
				err := dockerStop(config.containerName, config.Dryrun)
				if err != nil {
					Error.log(err)
				}
			} else {
				return errors.Errorf("Error in 'appsody %s': %s", mode, error)

			}

		} else {
			Info.log("Closing down development environment.")
		}

	} else {
		//This is buildah path - so do a kube apply instead
		dryrun := config.Dryrun

		portList, portsErr := getExposedPorts(config.RootCommandConfig)
		if portsErr != nil {
			return portsErr
		}

		projectDir, err := getProjectDir(config.RootCommandConfig)
		if err != nil {
			return err
		}
		dockerMounts, err := getVolumeArgs(config.RootCommandConfig)
		if err != nil {
			return err
		}
		depsMount, err := GetEnvVar("APPSODY_DEPS", config.RootCommandConfig)
		if err != nil {
			return err
		}
		deploymentYaml, err := GenDeploymentYaml(config.containerName, platformDefinition, controllerImageName, portList, projectDir, dockerMounts, depsMount, dryrun)
		if err != nil {
			return err
		}
		//hack
		namespace := ""
		//endhack
		err = KubeApply(deploymentYaml, namespace, dryrun)
		if err != nil {
			return err
		}
		serviceYaml, err := GenServiceYaml(config.containerName, portList, projectDir, dryrun)
		if err != nil {
			return err
		}

		err = KubeApply(serviceYaml, namespace, dryrun)
		if err != nil {
			return err
		}
		codeWindProjectID := os.Getenv("CODEWIND_PROJECT_ID")
		if codeWindProjectID == "" {
			port := getIngressPort(config.RootCommandConfig)
			// Generate the Ingress only if it makes sense - i.e. there's a port to expose
			if port > 0 {
				routeYaml, err := GenRouteYaml(config.containerName, projectDir, port, dryrun)
				if err != nil {
					return err
				}

				err = KubeApply(routeYaml, namespace, dryrun)
				if err != nil {
					return err
				}
			}
		}

		deploymentName := "deployment/" + config.containerName
		var timeout = "2m"
		kubeArgs := []string{"logs", deploymentName, "-f", "--pod-running-timeout=" + timeout}

		for {
			var waitErr, kubeErr error
			var execCmd *exec.Cmd
			if config.Dryrun {
				Info.log("Dry Run - Skipping kubectl logs")
				break
			} else {
				Info.Log("Getting the logs ...")
				execCmd, kubeErr = RunKubeCommandAndListen(kubeArgs, Container, config.interactive, config.Verbose, config.Dryrun)
				if kubeErr != nil {
					Debug.Log("kubectl log error: ", kubeErr.Error())
					time.Sleep(5 * time.Second)

				} else {
					waitErr = execCmd.Wait()
					if waitErr != nil {
						Debug.Log("kubectl log wait error: ", waitErr.Error())
						time.Sleep(5 * time.Second)

					}
				}
				if waitErr == nil && kubeErr == nil {
					break

				} // errors are nil

			} // end if not dryrun

		} //end of for loop

	} // end of buildah path
	return nil
}

func processPorts(cmdArgs []string, config *devCommonConfig) ([]string, error) {

	var exposedPortsMapping []string

	dockerExposedPorts, portsErr := getExposedPorts(config.RootCommandConfig)
	if portsErr != nil {
		return cmdArgs, portsErr
	}
	Debug.log("Exposed ports provided by the docker file", dockerExposedPorts)
	// if the container port is not in the lised of exposed ports add it to the list

	containerPort, envErr := GetEnvVar("PORT", config.RootCommandConfig)
	if envErr != nil {
		return cmdArgs, envErr
	}
	containerPortIsExposed := false

	Debug.log("Container port set to: ", containerPort)
	if containerPort != "" {
		for i := 0; i < len(dockerExposedPorts); i++ {

			if containerPort == dockerExposedPorts[i] {
				containerPortIsExposed = true
			}
		}
		if !containerPortIsExposed {
			dockerExposedPorts = append(dockerExposedPorts, containerPort)
		}
	}

	if config.publishAllPorts {
		cmdArgs = append(cmdArgs, "-P")
		// user specified to publish all EXPOSE ports to random ports with -P, so clear this list so we don't add them with -p
		dockerExposedPorts = []string{}
		if containerPort != "" && !containerPortIsExposed {
			// A PORT var was defined in the stack but not EXPOSE. It won't get published with -P, so add it as -p
			dockerExposedPorts = append(dockerExposedPorts, containerPort)
		}
	}

	Debug.log("Published ports provided as inputs: ", config.ports)
	for i := 0; i < len(config.ports); i++ { // this is the list of input -p's

		exposedPortsMapping = append(exposedPortsMapping, config.ports[i])

	}
	// see if there are any exposed ports (including container port) for which there are no overrides and add those to the list
	for i := 0; i < len(dockerExposedPorts); i++ {
		overrideFound := false
		for j := 0; j < len(config.ports); j++ {
			portMapping := strings.Split(config.ports[j], ":")
			if dockerExposedPorts[i] == portMapping[1] {
				overrideFound = true
			}
		}
		if !overrideFound {
			exposedPortsMapping = append(exposedPortsMapping, dockerExposedPorts[i]+":"+dockerExposedPorts[i])
		}
	}

	for k := 0; k < len(exposedPortsMapping); k++ {
		cmdArgs = append(cmdArgs, "-p", exposedPortsMapping[k])
	}
	return cmdArgs, nil
}
func checkPortInput(publishedPorts []string) (bool, error) {
	validPorts := true
	var portError error
	validPortNumber := regexp.MustCompile("^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$")
	for i := 0; i < len(publishedPorts); i++ {
		if !strings.Contains(publishedPorts[i], ":") {
			validPorts = false
			portError = errors.New("The port input: " + publishedPorts[i] + " is not valid as the : separator is missing.")
			break
		} else {
			// check the numbers
			portValues := strings.Split(publishedPorts[i], ":")
			fmt.Println(portValues)
			if !validPortNumber.MatchString(portValues[0]) || !validPortNumber.MatchString(portValues[1]) {
				portError = errors.New("The numeric port input: " + publishedPorts[i] + " is not valid.")
				validPorts = false
				break

			}

		}
	}
	return validPorts, portError

}
