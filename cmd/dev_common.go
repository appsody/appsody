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
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

type devCommonConfig struct {
	*RootCommandConfig
	disableWatcher  bool
	containerName   string
	ports           []string
	publishAllPorts bool
	interactive     bool
	dockerNetwork   string
	dockerOptions   string
}

func checkDockerRunOptions(options []string, config *RootCommandConfig) error {
	runOptionsTest := "(^((--help)|(-p)|(--publish)|(--publish-all)|(-P)|(-u)|(--user)|(--name)|(--network)|(-t)|(--tty)|(--rm)|(--entrypoint))((=?$)|(=.*)))"
	blackListedRunOptionsRegexp := regexp.MustCompile(runOptionsTest)

	for ind, value := range options {
		isInBlackListed := blackListedRunOptionsRegexp.MatchString(value)
		if isInBlackListed {
			return errors.Errorf("%s is not allowed in --docker-options", value)
		}
		if value == "-v" || value == "--volume" {
			var p ProjectFile
			project, _, err := p.EnsureProjectIDAndEntryExists(config)
			if err != nil {
				return err
			}

			if ind+1 == len(options) {
				return errors.Errorf("-v or --volume flag passed without associated fields. Options passed: %s", options)
			}

			userSpecifiedMount := options[ind+1]
			userSpecifiedMountSplit := strings.Split(userSpecifiedMount, ":")
			if len(userSpecifiedMountSplit) != 2 {
				return errors.Errorf("User specified mount %s is not in the correct format.", userSpecifiedMount)
			}
			userSpecifiedMountPath := userSpecifiedMountSplit[1]

			if strings.HasPrefix(userSpecifiedMountPath, "/.appsody") {
				return errors.Errorf("User specified mount %s cannot override /.appsody folder.", userSpecifiedMount)
			}

			stackMounts, err := getStackMounts(config)
			if err != nil {
				return err
			}

			// Checking against mounts specified in APPSODY_MOUNTS
			for _, mount := range stackMounts {
				mountSplit := strings.Split(mount, ":")
				if len(mountSplit) != 2 {
					return errors.Errorf("Stack specified mount %s is not in the correct format.", mount)
				}
				mountPath := mountSplit[1]
				if userSpecifiedMountPath == mountPath {
					return errors.Errorf("User specified mount path %s is not allowed in --docker-options, as it interferes with the default specified mount path %s", userSpecifiedMountPath, mountPath)
				}
			}

			// Checking against volume specified in APPSODY_DEPS which are store in the project.yaml
			for _, volume := range project.Volumes {
				if userSpecifiedMountPath == volume.Path {
					return errors.Errorf("User specified mount path %s is not allowed in --docker-options, as it interferes with the stack specified mount path %s", userSpecifiedMountPath, volume.Path)
				}
			}
		}
	}
	return nil
}

func addNameFlag(cmd *cobra.Command, flagVar *string, config *RootCommandConfig) {
	projectName, perr := getProjectName(config)
	if perr != nil {
		if _, ok := perr.(*NotAnAppsodyProject); !ok {
			config.Error.logf("Error occurred retrieving project name... exiting: %s", perr)
			os.Exit(1)
		}
	}

	defaultName := projectName
	cmd.PersistentFlags().StringVar(flagVar, "name", defaultName, "Assign a name to your development container.")
}

func addStackRegistryFlag(cmd *cobra.Command, flagVar *string, config *RootCommandConfig) {

	defaultRegistry := getDefaultStackRegistry(config)
	stackRegistryInConfigFile, err := getStackRegistryFromConfigFile(config)
	if err != nil {
		if _, ok := err.(*NotAnAppsodyProject); !ok {
			config.Debug.Logf("Error retrieving the stack registry from config file: %v", err)
		}
		cmd.PersistentFlags().StringVar(flagVar, "stack-registry", defaultRegistry, "Specify the URL of the registry that hosts your stack images. [WARNING] Your current settings are incorrect - change your project config or use this flag to override the image registry.")
	} else if stackRegistryInConfigFile == "" {
		cmd.PersistentFlags().StringVar(flagVar, "stack-registry", defaultRegistry, "Specify the URL of the registry that hosts your stack images.")
	} else {
		cmd.PersistentFlags().StringVar(flagVar, "stack-registry", stackRegistryInConfigFile, "Specify the URL of the registry that hosts your stack images.")
	}
}

func addDevCommonFlags(cmd *cobra.Command, config *devCommonConfig) {

	addNameFlag(cmd, &config.containerName, config.RootCommandConfig)
	addStackRegistryFlag(cmd, &config.StackRegistry, config.RootCommandConfig)
	cmd.PersistentFlags().StringVar(&config.dockerNetwork, "network", "", "Specify the network for docker to use.")
	cmd.PersistentFlags().StringArrayVarP(&config.ports, "publish", "p", nil, "Publish the container's ports to the host. The stack's exposed ports will always be published, but you can publish addition ports or override the host ports with this option.")
	cmd.PersistentFlags().BoolVarP(&config.publishAllPorts, "publish-all", "P", false, "Publish all exposed ports to random ports")
	cmd.PersistentFlags().BoolVar(&config.disableWatcher, "no-watcher", false, "Disable file watching, regardless of container environment variable settings.")
	cmd.PersistentFlags().BoolVarP(&config.interactive, "interactive", "i", false, "Attach STDIN to the container for interactive TTY mode")
	cmd.PersistentFlags().StringVar(&config.dockerOptions, "docker-options", "", "Specify the docker run options to use.  Value must be in \"\". The following Docker options are not supported:  '--help','-p','--publish-all','-P','-u','-—user','-—name','-—network','-t','-—tty,'—rm','—entrypoint', '--mount'.")
}

func commonCmd(config *devCommonConfig, mode string) error {
	depErr := GetDeprecated(config.RootCommandConfig)
	if depErr != nil {
		return depErr
	}
	config.Debug.Log("Default stack registry set to: ", &config.StackRegistry)
	// Checking whether the controller is being overridden
	overrideControllerImage := os.Getenv("APPSODY_CONTROLLER_IMAGE")
	if overrideControllerImage == "" {
		overrideVersion := os.Getenv("APPSODY_CONTROLLER_VERSION")
		if overrideVersion != "" {
			CONTROLLERVERSION = overrideVersion
			config.Warning.Log("You have overridden the Appsody controller version and set it to: ", CONTROLLERVERSION)
		}
	} else {
		config.Warning.Log("The Appsody CLI detected the APPSODY_CONTROLLER_IMAGE env var. The controller image that will be used is: ", overrideControllerImage)
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
		config.Warning.Log("The Appsody CLI will use the latest version of the controller image. This may result in a mismatch or malfunction.")
	}
	projectDir, perr := getProjectDir(config.RootCommandConfig)
	if perr != nil {
		return perr
	}
	config.Debug.log("Project config file set to: ", filepath.Join(projectDir, ConfigFile))

	projectConfig, configErr := getProjectConfig(config.RootCommandConfig)
	if configErr != nil {
		return configErr
	}

	err := CheckPrereqs(config.RootCommandConfig)
	if err != nil {
		config.Warning.logf("Failed to check prerequisites: %v\n", err)
	}

	platformDefinition := projectConfig.Stack
	config.Debug.log("Stack image: ", platformDefinition)
	config.Debug.log("Project directory: ", projectDir)

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
	depsEnvVars, envErr := GetDepVolumeArgs(config.RootCommandConfig)
	if envErr != nil {
		return envErr
	}

	if depsEnvVars != nil {
		var project ProjectFile

		//add volumes to project entry of current Appsody project
		volumeMaps, err = project.AddDepsVolumesToProjectEntry(depsEnvVars, volumeMaps, config.RootCommandConfig)
		if err != nil {
			return err
		}
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
			volNames, err := RunDockerVolumeList(config.LoggingConfig, controllerVolumeName)
			if err != nil {
				config.Debug.Log("Error attempting to query volumes for ", controllerVolumeName, " :", err)
				return err
			}
			config.Debug.Log("Retrieved volume name(s): [", volNames, "]")
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
			config.Debug.Logf("Controller volume not found or version is latest - launching the %s image to populate it", controllerImageName)
			downloaderArgs := []string{"--rm", "-v", controllerVolumeMount, controllerImageName}
			controllerDownloader, err := DockerRunAndListen(config.RootCommandConfig, downloaderArgs, config.Info, false)
			if config.Dryrun {
				config.Info.log("Dry Run - Skipping execCmd.Wait")
			} else {
				if err == nil {
					err = controllerDownloader.Wait()
				}
			}
			if err != nil {
				config.Debug.Log("Error populating the controller volume: ", err)
				return err
			}
		}
	}

	//controllerMount := controllerVolumeName + ":/appsody"
	config.Debug.log("Adding controller to volume mounts: ", controllerVolumeMount)
	volumeMaps = append(volumeMaps, "-v", controllerVolumeMount)

	var wg sync.WaitGroup
	if !config.Buildah {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			<-c
			wg.Add(1)
			defer wg.Done()
			config.Debug.Log("Inside signal handler for appsody command")
			err := dockerStop(config.RootCommandConfig, config.containerName, config.Dryrun)
			if err != nil {
				config.Error.log(err)
			}
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
		config.Debug.logf("User provided Docker options: \"%s\"", config.dockerOptions)
		dockerOptions := config.dockerOptions
		dockerOptions = strings.TrimPrefix(dockerOptions, " ")
		dockerOptions = strings.TrimSuffix(dockerOptions, " ")
		dockerOptionsCmd := strings.Split(dockerOptions, " ")
		err := checkDockerRunOptions(dockerOptionsCmd, config.RootCommandConfig)
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
	if config.interactive {
		cmdArgs = append(cmdArgs, "--interactive")
	}
	if !config.Buildah {
		config.Debug.logf("Attempting to start image %s with container name %s", platformDefinition, config.containerName)
		execCmd, err := DockerRunAndListen(config.RootCommandConfig, cmdArgs, config.Container, config.interactive)
		if config.Dryrun {
			config.Info.log("Dry Run - Skipping execCmd.Wait")
		} else {
			if err == nil {
				err = execCmd.Wait()
			}
		}
		if err != nil {
			// 'signal: interrupt'
			// TODO presumably you can query the error itself
			error := fmt.Sprintf("%s", err)
			config.Debug.log("CLI exit error is:  ", error)
			//Linux and Windows return a different error on Ctrl-C
			if error == "signal: interrupt" || error == "signal: terminated" || error == "exit status 2" {
				config.Info.log("Closing down, development environment was interrupted.")

			} else {
				return errors.Errorf("Error in 'appsody %s': %s", mode, error)
			}

		} else {
			config.Info.log("Closing down development environment.")
		}

	} else {
		//This is buildah path - so do a kube apply instead
		dryrun := config.Dryrun

		portList, portsErr := getExposedPorts(config.RootCommandConfig)
		if portsErr != nil {
			return portsErr
		}

		var debugPortArray []string
		debugPorts, debugPortErr := GetEnvVar("APPSODY_DEBUG_PORT", config.RootCommandConfig)
		if debugPortErr != nil || debugPorts == "" {
			config.Debug.log("No debug port found. Continuing...")
		} else {
			debugPortArray = strings.Split(debugPorts, " ") //Split the string if multiple debug ports exist
			for _, debugPort := range debugPortArray {
				debugPortExists := FindElement(portList, debugPort) //Determine whether all ports specified in env var have actually been exposed
				if !debugPortExists {
					return errors.Errorf("Port: %s specified in APPSODY_DEBUG_PORT could not be found in ports list", debugPort)
				}
			}
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

		dockerEnvVars, err := ExtractDockerEnvVars(config.dockerOptions)
		if err != nil {
			return err
		}
		config.Debug.Logf("Docker env vars extracted from docker options: %v", dockerEnvVars)
		deploymentYaml, err := GenDeploymentYaml(config.LoggingConfig, config.containerName, platformDefinition, controllerImageName, portList, debugPortArray, projectDir, dockerMounts, dockerEnvVars, depsMount, dryrun)
		if err != nil {
			return err
		}
		//hack
		namespace := ""
		//endhack
		err = KubeApply(config.LoggingConfig, deploymentYaml, namespace, dryrun)
		if err != nil {
			return err
		}
		serviceYaml, err := GenServiceYaml(config.LoggingConfig, config.containerName, portList, debugPortArray, projectDir, dryrun)
		if err != nil {
			return err
		}

		err = KubeApply(config.LoggingConfig, serviceYaml, namespace, dryrun)
		if err != nil {
			return err
		}
		codeWindProjectID := os.Getenv("CODEWIND_PROJECT_ID")
		if codeWindProjectID == "" {
			port := getIngressPort(config.RootCommandConfig)
			// Generate the Ingress only if it makes sense - i.e. there's a port to expose
			if port > 0 {
				routeYaml, err := GenRouteYaml(config.LoggingConfig, config.containerName, projectDir, port, dryrun)
				if err != nil {
					return err
				}

				err = KubeApply(config.LoggingConfig, routeYaml, namespace, dryrun)
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
				config.Info.log("Dry Run - Skipping kubectl logs")
				break
			} else {
				config.Info.Log("Getting the logs ...")
				execCmd, kubeErr = RunKubeCommandAndListen(config.RootCommandConfig, kubeArgs, config.Container, config.interactive)
				if kubeErr != nil {
					config.Debug.Log("kubectl log error: ", kubeErr.Error())
					time.Sleep(5 * time.Second)

				} else {
					waitErr = execCmd.Wait()
					if waitErr != nil {
						config.Debug.Log("kubectl log wait error: ", waitErr.Error())
						time.Sleep(5 * time.Second)

					}
				}
				if waitErr == nil && kubeErr == nil {
					break

				} // errors are nil

			} // end if not dryrun

		} //end of for loop

	} // end of buildah path
	wg.Wait()
	return nil
}

func processPorts(cmdArgs []string, config *devCommonConfig) ([]string, error) {
	var exposedPortsMapping []string

	dockerExposedPorts, portsErr := getExposedPorts(config.RootCommandConfig)
	if portsErr != nil {
		return cmdArgs, portsErr
	}

	config.Debug.log("Exposed ports provided by the docker file", dockerExposedPorts)
	// if the container port is not in the lised of exposed ports add it to the list

	containerPort, envErr := GetEnvVar("PORT", config.RootCommandConfig)
	if envErr != nil {
		return cmdArgs, envErr
	}
	containerPortIsExposed := false

	config.Debug.log("Container port set to: ", containerPort)
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

	config.Debug.log("Published ports provided as inputs: ", config.ports)
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
			if !validPortNumber.MatchString(portValues[0]) || !validPortNumber.MatchString(portValues[1]) {
				portError = errors.New("The numeric port input: " + publishedPorts[i] + " is not valid.")
				validPorts = false
				break

			}

		}
	}
	return validPorts, portError
}
