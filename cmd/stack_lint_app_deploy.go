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
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func validateAppDeploy(appDeployPath string, appDeployKeys []string) int {
	stackLintWarningCount := 0
	var deployFileContents []string

	Info.log("LINTING app-deploy.yaml: ", appDeployPath)

	appdeploy, err := ioutil.ReadFile(appDeployPath)
	if err != nil {
		Warning.log("app deploy err ", err)
		stackLintWarningCount++
	}

	appDeployMap := make(map[string]interface{})
	err = yaml.UnmarshalStrict([]byte(appdeploy), &appDeployMap)
	if err != nil {
		Warning.log("Unmarshal: Error unmarshalling app-deploy.yaml")
		stackLintWarningCount++
	}

	mapString := make(map[string]interface{})

	for key, value := range appDeployMap {
		if b, ok := value.(map[interface{}]interface{}); ok {
			for nestedKey, nestedValue := range b {
				strKey := fmt.Sprintf("%v", nestedKey)
				mapString[strKey] = nestedValue
			}
			for z := range mapString {
				deployFileContents = append(deployFileContents, z)
			}
		} else {
			deployFileContents = append(deployFileContents, key)
		}
	}

	variableFound := false

	for _, keys := range appDeployKeys {
		for _, content := range deployFileContents {
			if keys == content {
				variableFound = true
			}
		}
		if !variableFound {
			Warning.log("Missing field: ", keys)
			stackLintWarningCount++
		}
		variableFound = false
	}

	return stackLintWarningCount
}
