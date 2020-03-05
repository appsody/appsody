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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func IsValidKubernetesLabelValue(value string) (bool, error) {
	if value == "" {
		return true, nil
	}
	if len(value) > 63 {
		return false, errors.New("The label must be 63 characters or less")
	}

	match, err := regexp.MatchString("^[a-z0-9A-Z]([a-z0-9A-Z-_.]*[a-z0-9A-Z])?$", value)
	if err != nil {
		return false, err
	}

	if match {
		return true, nil
	}
	return false, errors.Errorf("Invalid label \"%s\". The label must begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between.", value)
}

// ConvertToValidProjectName takes an existing string or directory path
// and returns a name that conforms to isValidContainerName rules
func ConvertToValidProjectName(projectDir string) (string, error) {
	projectName := strings.ToLower(filepath.Base(projectDir))
	valid, _ := IsValidProjectName(projectName)

	if !valid {
		projectName = strings.ToLower(filepath.Base(projectDir))
		if len(projectName) > 68 {
			projectName = projectName[0:68]
		}

		if projectName[0] < 'a' || projectName[0] > 'z' {
			projectName = "appsody-" + projectName
		}

		reg, err := regexp.Compile("[^a-z0-9]+")
		if err != nil {
			return "", err
		}
		projectName = reg.ReplaceAllString(projectName, "-")

		if projectName[len(projectName)-1] == '-' {
			projectName = projectName + "app"
		}

		valid, err := IsValidProjectName(projectName)
		if !valid {
			return projectName, err
		}
	}

	return projectName, nil
}

func saveProjectNameToConfig(projectName string, config *RootCommandConfig) error {
	valid, err := IsValidProjectName(projectName)
	if !valid {
		return err
	}

	// update the in-memory project name
	projectConfig, err := getProjectConfig(config)
	if err != nil {
		return err
	}
	projectConfig.ProjectName = projectName

	// save the project name to the .appsody-config.yaml
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err = v.ReadInConfig()
	if err != nil {
		return err
	}
	v.Set("project-name", projectName)
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	config.Info.log("Your Appsody project name has been set to ", projectName)
	return nil
}

func saveApplicationNameToConfig(applicationName string, config *RootCommandConfig) error {
	valid, err := IsValidProjectName(applicationName)
	if !valid {
		return err
	}

	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err = v.ReadInConfig()
	if err != nil {
		return err
	}
	v.Set("application-name", applicationName)
	err = v.WriteConfig()
	if err != nil {
		return err
	}

	config.Info.log("Your Appsody application name has been set to ", applicationName)
	return nil
}

func setStackRegistry(stackRegistry string, config *RootCommandConfig) error {

	// Read in the config
	appsodyConfig := filepath.Join(config.ProjectDir, ConfigFile)
	v := viper.New()
	v.SetConfigFile(appsodyConfig)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	stackImageName, err := OverrideStackRegistry(stackRegistry, v.Get("stack").(string))
	if err != nil {
		return err
	}
	stackImageName, err = NormalizeImageName(stackImageName)
	if err != nil {
		return err
	}
	v.Set("stack", stackImageName)
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	config.Info.log("Your Appsody project stack has been set to ", stackImageName)
	return nil
}

func execAndWaitWithWorkDir(log *LoggingConfig, command string, args []string, logger appsodylogger, workdir string, dryrun bool) error {

	err := execAndWaitWithWorkDirReturnErr(log, command, args, logger, workdir, dryrun)
	if err != nil {
		return errors.Errorf("Error running %s command: %v", command, err)

	}
	return nil

}

//OverrideStackRegistry allows you to change the image registry URL
func OverrideStackRegistry(override string, imageName string) (string, error) {
	if override == "" {
		return imageName, nil
	}
	match, err := ValidateHostNameAndPort(override)
	if err != nil {
		return "", err
	}
	if !match {
		return "", errors.Errorf("This is an invalid host name: %s", override)
	}
	imageNameComponents := strings.Split(imageName, "/")
	if len(imageNameComponents) == 3 {
		imageNameComponents[0] = override
	}
	if len(imageNameComponents) == 2 || len(imageNameComponents) == 1 {
		newComponent := []string{override}
		imageNameComponents = append(newComponent, imageNameComponents...)
	}
	if len(imageNameComponents) > 3 {
		return "", errors.Errorf("Image name is invalid and needs to be changed in the project config file (.appsody-config.yaml): %s. Too many slashes (/) - the override cannot take place.", imageName)
	}
	return strings.Join(imageNameComponents, "/"), nil
}

//ValidateHostNameAndPort validates that hostNameAndPort conform to the DNS naming conventions
func ValidateHostNameAndPort(hostNameAndPort string) (bool, error) {
	match, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])($|:[0-9]{1,5}$)`, hostNameAndPort)
	return match, err
}

//NormalizeImageName is a temporary fix for buildah workaround #676
func NormalizeImageName(imageName string) (string, error) {
	imageNameComponents := strings.Split(imageName, "/")
	if len(imageNameComponents) == 2 {
		return imageName, nil
	}

	if len(imageNameComponents) == 1 {
		return fmt.Sprintf("docker.io/%s", imageName), nil
	}

	if len(imageNameComponents) == 3 {
		if imageNameComponents[0] == "index.docker.io" {
			imageNameComponents[0] = "docker.io"
			return strings.Join(imageNameComponents, "/"), nil
		}
		return imageName, nil
	}
	return imageName, errors.Errorf("Image name is invalid: %s", imageName)

}

func execAndWaitWithWorkDirReturnErr(log *LoggingConfig, command string, args []string, logger appsodylogger, workdir string, dryrun bool) error {
	var err error
	var execCmd *exec.Cmd
	if dryrun {
		log.Info.log("Dry Run - Skipping command: ", command, " ", strings.Join(args, " "))
	} else {
		execCmd, err = execAndListenWithWorkDirReturnErr(log, command, args, logger, workdir, dryrun)
		if err != nil {
			return err
		}
		err = execCmd.Wait()
		if err != nil {
			return err
		}
	}
	return err
}

//Compares the minimum requirements of a stack against the user to determine whether they can use the stack or not.
func CheckStackRequirements(log *LoggingConfig, requirementArray map[string]string, buildah bool) error {
	versionRegex := regexp.MustCompile(`(\d)+\.(\d)+\.(\d)+`)
	upgradesRequired := 0

	log.Info.log("Checking stack requirements...")

	for technology, minVersion := range requirementArray {
		if minVersion == "" {
			continue
		}
		if technology == "Docker" && buildah {
			log.Debug.log("Skipping Docker requirement - Buildah is being used.")
			continue
		}
		if technology == "Buildah" && !buildah {
			log.Debug.log("Skipping Buildah requirement - Docker is being used.")
			continue
		}
		log.Debug.logf("Checking version requirement: %s %s", technology, minVersion)

		setConstraint, err := semver.NewConstraint(minVersion)
		if err != nil {
			log.Warning.logf("Skipping %s version requirement because the minimum version is invalid: %s", technology, err)
			continue
		}

		var runVersionCmd string
		if strings.ToLower(technology) == "appsody" {
			if VERSION == "0.0.0" || VERSION == "vlatest" {
				log.Warning.log("Skipping appsody version requirement because this is a local build of appsody ", VERSION)
				continue
			}
			runVersionCmd = VERSION
		} else {
			cmd := exec.Command(strings.ToLower(technology), "version")
			runVersionCmd, err = SeparateOutput(cmd)
			if err != nil {
				log.Error.log(err, " - Are you sure ", technology, " is installed?")
				upgradesRequired++
				continue
			}
			log.Debug.logf("Output of running %s: %s", strings.ToLower(technology)+" version", runVersionCmd)
		}

		cutCmdOutput := versionRegex.FindString(runVersionCmd)
		parseUserVersion, parseErr := semver.NewVersion(cutCmdOutput)
		if parseErr != nil {
			log.Warning.logf("Unable to parse %s version - This stack may not work in your current development environment. %s", technology, parseErr)
			// Continue when the version can not be determined
			continue
		}
		log.Debug.logf("Found version of %s to be %s", technology, parseUserVersion)
		compareVersion := setConstraint.Check(parseUserVersion)

		if compareVersion {
			log.Info.log(technology + " requirements met")
		} else {
			log.Error.log("The required version of " + technology + " to use this stack is " + minVersion + " - Please upgrade.")
			upgradesRequired++
		}
	}
	if upgradesRequired > 0 {
		return errors.Errorf("One or more technologies need upgrading to use this stack. Upgrades required: %v", upgradesRequired)
	}
	return nil
}
