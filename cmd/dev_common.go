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
	"os/signal"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

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
		if pmsg, ok := perr.(*NotAnAppsodyProject); ok {
			//Debug.log("Cannot retrieve the project name - continuing: ", perr)
		} else {
			Error.logf("Error occurred retrieving project name... exiting: %s", pmsg)
			os.Exit(1)
		}
	}

	defaultName := projectName + "-dev"
	cmd.PersistentFlags().StringVar(flagVar, "name", defaultName, "Assign a name to your development container.")
}

func addDevCommonFlags(cmd *cobra.Command, config *devCommonConfig) {

	projectName, perr := getProjectName(config.RootCommandConfig)
	if perr != nil {
		if pmsg, ok := perr.(*NotAnAppsodyProject); ok {
			// Debug.log("Cannot retrieve the project name - continuing: ", perr)
		} else {
			Error.logf("Error occurred retrieving project name... exiting: %s", pmsg)
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

	platformDefinition := projectConfig.Platform
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
	destController := os.Getenv("APPSODY_MOUNT_CONTROLLER")
	if destController != "" {
		Debug.log("Overriding appsody-controller mount with APPSODY_MOUNT_CONTROLLER env variable: ", destController)
	} else {
		// Copy the controller from the installation directory to the home (.appsody)
		destController = filepath.Join(getHome(config.RootCommandConfig), "appsody-controller")
		// Debug.log("Attempting to load the controller from ", destController)
		//if _, err := os.Stat(destController); os.IsNotExist(err) {
		// Always copy it from the executable dir
		//Retrieving the path of the binaries appsody and appsody-controller
		//Debug.log("Didn't find the controller in .appsody - copying from the binary directory...")
		executable, _ := os.Executable()
		binaryLocation, err := filepath.Abs(filepath.Dir(executable))
		Debug.log("Binary location ", binaryLocation)
		if err != nil {
			return errors.New("fatal error - can't retrieve the binary path... exiting")
		}
		controllerExists, existsErr := Exists(destController)
		if existsErr != nil {
			return existsErr
		}
		Debug.log("appsody-controller exists: ", controllerExists)
		checksumMatch := false
		if controllerExists {
			var checksumMatchErr error
			binaryControllerPath := filepath.Join(binaryLocation, "appsody-controller")
			binaryControllerExists, existsErr := Exists(binaryControllerPath)
			if existsErr != nil {
				return existsErr
			}
			if binaryControllerExists {
				checksumMatch, checksumMatchErr = checksum256TestFile(binaryControllerPath, destController)
				Debug.log("checksum returned: ", checksumMatch)
				if checksumMatchErr != nil {
					return checksumMatchErr
				}
			} else {
				//the binary controller did not exist so skip copying it
				Warning.log("The binary controller could not be found.")
				checksumMatch = true
			}
		}
		// if the controller doesn't exist
		if !controllerExists || (controllerExists && !checksumMatch) {
			Debug.log("Replacing Controller")

			//Construct the appsody-controller mount
			sourceController := filepath.Join(binaryLocation, "appsody-controller")
			if config.Dryrun {
				Info.logf("Dry Run - Skipping copy of controller binary from %s to %s", sourceController, destController)
			} else {
				Debug.log("Attempting to copy the source controller from: ", sourceController)
				//Copy the controller from the binary location to $HOME/.appsody
				copyError := CopyFile(sourceController, destController)
				if copyError != nil {
					return errors.Errorf("Cannot retrieve controller - exiting: %v", copyError)
				}
				// Making the controller executable in case CopyFile loses permissions
				chmodErr := os.Chmod(destController, 0755)
				if chmodErr != nil {
					return errors.Errorf("Cannot make the controller  executable - exiting: %v", chmodErr)
				}
			}
		}
		//} Used to close the "if controller does not exist"
	}
	controllerMount := destController + ":/appsody/appsody-controller"
	Debug.log("Adding controller to volume mounts: ", controllerMount)
	volumeMaps = append(volumeMaps, "-v", controllerMount)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := dockerStop(config.containerName, config.Dryrun)
		if err != nil {
			Error.log(err)
		}
		//containerRemove(containerName) is not needed due to --rm flag
	}()

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
	cmdArgs = append(cmdArgs, "-t", "--entrypoint", "/appsody/appsody-controller", platformDefinition, "--mode="+mode)
	if config.Verbose {
		cmdArgs = append(cmdArgs, "-v")
	}
	if config.disableWatcher {
		cmdArgs = append(cmdArgs, "--no-watcher")
	}
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
		//Linux and Windows return a different error on Ctrl-C
		if error == "signal: interrupt" || error == "exit status 2" {
			Info.log("Closing down, development environment was interrupted.")
		} else {
			return errors.Errorf("Error in 'appsody %s': %s", mode, error)

		}

	} else {
		Info.log("Closing down development environment.")
	}
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
