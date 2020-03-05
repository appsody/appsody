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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getENVDockerfile(log *LoggingConfig, stackPath string) (dockerfileStack map[string]string) {
	arg := filepath.Join(stackPath, "image/Dockerfile-stack")

	file, err := os.Open(arg)

	if err != nil {
		log.Error.log("failed opening file: ", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "ENV") {
			txtlines = append(txtlines, strings.TrimSpace(scanner.Text()))
		}
	}

	file.Close()

	dockerfileMap := make(map[string]string)

	for _, eachline := range txtlines {
		s := strings.Split(eachline, "=")
		key := strings.TrimPrefix(s[0], "ENV")
		key = strings.TrimSpace(key)
		value := strings.TrimPrefix(eachline, s[0]+"=")
		dockerfileMap[key] = value
	}

	return dockerfileMap
}

func lintDockerFileStack(log *LoggingConfig, stackPath string) (int, int) {
	requiredEnvironmentVariables := [...]string{"APPSODY_MOUNTS", "APPSODY_RUN"}
	optionalEnvironmentVariables := [...]string{"APPSODY_DEBUG", "APPSODY_TEST", "APPSODY_DEPS", "APPSODY_PROJECT_DIR"}

	stackLintErrorCount := 0
	stackLintWarningCount := 0
	arg := filepath.Join(stackPath, "image/Dockerfile-stack")

	log.Info.log("Linting Dockerfile-stack: ", arg)

	dockerfileStack := getENVDockerfile(log, stackPath)

	variableFound := false
	variable := ""

	for i := 0; i < len(requiredEnvironmentVariables); i++ {
		variable = requiredEnvironmentVariables[i]
		for k := range dockerfileStack {
			if k == requiredEnvironmentVariables[i] {
				variableFound = true
			}
		}
		if !variableFound {
			log.Error.log("Missing ", variable)
			stackLintErrorCount++
		}
		variableFound = false
	}

	variableFound = false

	for i := 0; i < len(optionalEnvironmentVariables); i++ {
		variable = optionalEnvironmentVariables[i]
		for k := range dockerfileStack {
			if k == optionalEnvironmentVariables[i] {
				variableFound = true
			}
		}
		if !variableFound {
			log.Warning.log("Missing ", variable)
			stackLintWarningCount++
		}
		variableFound = false
	}

	count := 0
	onChangeFound := false

	for k := range dockerfileStack {
		if strings.Contains(k, "APPSODY_WATCH_DIR") {
			for j := range dockerfileStack {
				count++
				if strings.Contains(j, "_ON_CHANGE") {
					onChangeFound = true

				}
			}
			break
		}
	}

	if count == len(dockerfileStack) && !onChangeFound {
		log.Error.log("APPSODY_WATCH_DIR is defined, but no ON_CHANGE variable is defined")
		stackLintErrorCount++
	}

	for k, v := range dockerfileStack {
		if strings.Contains(k, "APPSODY_INSTALL") {
			log.Warning.log("APPSODY_INSTALL should be deprecated and APPSODY_PREP should be used instead")
			stackLintWarningCount++
		}

		if strings.Contains(k, "_KILL") {
			if !(v == "true" || v == "false") {
				log.Error.log(k, " can only have value true/false")
				stackLintErrorCount++
			}
		}

		if strings.Contains(k, "APPSODY_WATCH_REGEX") {
			_, err := regexp.Compile(v)

			if err != nil {
				log.Error.log(err)
				stackLintErrorCount++
			}
		}
	}
	mountVar := dockerfileStack["APPSODY_MOUNTS"]
	lintMountWarnings, lintMountErrors := lintMountVar(mountVar, log, stackPath)
	stackLintWarningCount = stackLintWarningCount + lintMountWarnings
	stackLintErrorCount = stackLintErrorCount + lintMountErrors
	return stackLintErrorCount, stackLintWarningCount
}

func lintMountVar(mountListSource string, log *LoggingConfig, stackPath string) (int, int) {

	if mountListSource == "" {
		log.Error.log("No APPSODY MOUNTS exists, mount paths can not be validated.")
		return 0, 1
	}
	errCount := 0
	warningCount := 0
	mountList := strings.Split(mountListSource, ";")

	templatePath := filepath.Join(stackPath, "templates")
	log.Debug.log("Checking for template path: ", templatePath)
	fileCheck, err := Exists(templatePath)
	log.Debug.log("Template path exists: ", fileCheck)
	if err != nil {
		log.Error.log("Error attempting to determine if template path exists: ", err)
		return 0, 1
	}
	if !fileCheck {
		log.Error.log("Missing template directory in: ", stackPath)
		return 0, 1
	}
	if IsEmptyDir(templatePath) {
		log.Error.log("No templates found in: ", templatePath)
		return 0, 1
	}
	templates, _ := ioutil.ReadDir(templatePath)

	// loop through the template directories
	for _, f := range templates {

		for _, mount := range mountList {
			traceMount := strings.Trim(mount, "\"")
			log.Debug.log("mount pair: ", traceMount)
			localPaths := strings.Split(traceMount, ":")
			if len(localPaths) != 2 {
				log.Error.log("Mount is not properly formatted it is missing the single colon: ", traceMount)
				errCount++
			} else {

				localPath := localPaths[0]
				if localPath == "" {
					log.Error.logf("Path for mount %s is empty: ", traceMount)
					errCount++
				} else {

					log.Debug.log("local path: ", localPath)
					if localPath == "/" || localPath == "." {
						log.Debug.logf("Path %s for mount %s is a directory.", localPath, traceMount)
					} else if localPath[0:1] == "~" {
						log.Debug.logf("Path %s for mount %s can not be evaluated at this time.", localPath, traceMount)
					} else {
						mountFilePath := filepath.Join(stackPath, "templates", f.Name(), localPath)

						log.Debug.log("mountFilePath: ", mountFilePath)
						warnings, valErrors := validateMountPath(mountFilePath, traceMount, log)
						warningCount = warningCount + warnings
						errCount = errCount + valErrors

					}
				}
			}
		}
	}
	return warningCount, errCount
}
func validateMountPath(path string, mount string, log *LoggingConfig) (int, int) {
	log.Debug.log("Attempting to validate mount path: ", path)

	warningCount, errorCount := 0, 0
	file, err := os.Stat(path)
	if err != nil {
		errorCount = 1
		log.Error.logf("Could not stat path: %s for mount %s", path, mount)
	} else {
		if file.Mode().IsDir() {
			log.Debug.logf("Path %s for mount %s is a directory", path, mount)
		} else {
			warningCount = 1
			log.Warning.logf("Path %s for mount %s points to a single file.  Single file Docker mount paths cause unexpected behavior and will be deprecated in the future.", path, mount)
		}

	}
	return warningCount, errorCount
}
