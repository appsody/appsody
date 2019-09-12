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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func lintDockerFileStack() (int, int) {
	mendatoryEnvironmentVariables := [...]string{"APPSODY_MOUNTS", "APPSODY_RUN", "APPSODY_RUN_ON_CHANGE", "APPSODY_RUN_KILL", "APPSODY_DEBUG", "APPSODY_DEBUG_ON_CHANGE", "APPSODY_DEBUG_KILL", "APPSODY_TEST", "APPSODY_TEST_ON_CHANGE", "APPSODY_TEST_KILL"}
	optionalEnvironmentVariables := [...]string{"APPSODY_DEPS", "APPSODY_WATCH_DIR"}
	errorCount := 0
	warningCount := 0

	stackPath, _ := os.Getwd()

	if len(os.Args) > 3 {
		stackPath = os.Args[3]
	}

	arg := filepath.Join(stackPath, "image/Dockerfile-stack")

	dockerfileStack, err := ioutil.ReadFile(arg)

	if err != nil {
		Error.log("Error attempting to read file: ", err)
		errorCount++
	}

	for i := 0; i < len(mendatoryEnvironmentVariables); i++ {
		if strings.Contains(string(dockerfileStack), mendatoryEnvironmentVariables[i]) == false {
			Error.log("Missing ", mendatoryEnvironmentVariables[i], " in: ", arg)
			errorCount++
		}
	}

	for i := 0; i < len(optionalEnvironmentVariables); i++ {
		if strings.Contains(string(dockerfileStack), optionalEnvironmentVariables[i]) == false {
			Warning.log("Missing ", optionalEnvironmentVariables[i], " in: ", arg)
			warningCount++
		}
	}

	return errorCount, warningCount
}
