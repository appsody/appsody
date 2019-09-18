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
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getENVDockerfile(stackPath string) (dockerfileStack map[string]string) {
	arg := filepath.Join(stackPath, "image/Dockerfile-stack")

	file, err := os.Open(arg)

	if err != nil {
		Error.log("failed opening file: ", err)
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

func lintDockerFileStack(stackPath string) {
	mendatoryEnvironmentVariables := [...]string{"APPSODY_MOUNTS", "APPSODY_RUN"}
	optionalEnvironmentVariables := [...]string{"APPSODY_DEBUG", "APPSODY_TEST", "APPSODY_DEPS", "APPSODY_PROJECT_DIR"}

	arg := filepath.Join(stackPath, "image/Dockerfile-stack")

	Info.log("Linting Dockerfile-stack: ", arg)

	dockerfileStack := getENVDockerfile(stackPath)

	variableFound := false
	variable := ""

	for i := 0; i < len(mendatoryEnvironmentVariables); i++ {
		variable = mendatoryEnvironmentVariables[i]
		for k := range dockerfileStack {
			if k == mendatoryEnvironmentVariables[i] {
				variableFound = true
			}
		}
		if !variableFound {
			Error.log("Missing ", variable)
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
			Warning.log("Missing ", variable)
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
		Error.log("APPSODY_WATCH_DIR is defined, but no ON_CHANGE variable is defined")
		stackLintErrorCount++
	}

	for k, v := range dockerfileStack {
		if strings.Contains(k, "APPSODY_INSTALL") {
			Warning.log("APPSODY_INSTALL should be deprecated and APPSODY_PREP should be used instead")
			stackLintWarningCount++
		}

		if strings.Contains(k, "_KILL") {
			if !(v == "true" || v == "false") {
				Error.log(k, " can only have value true/false")
				stackLintErrorCount++
			}
		}

		if strings.Contains(k, "APPSODY_WATCH_REGEX") {
			_, err := regexp.Compile(v)

			if err != nil {
				Error.log(err)
				stackLintErrorCount++
			}
		}
	}
}
