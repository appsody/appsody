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
	"path/filepath"

	"gopkg.in/yaml.v2"
)

//Structure for an app deploy file
type AppDeploy struct {
	APIVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   MetadataField `yaml:"metadata"`
	Spec       SpecField     `yaml:"spec"`
}

type MetadataField struct {
	Name string `yaml:"name"`
}

type SpecField struct {
	Version          string       `yaml:"version"`
	ApplicationImage string       `yaml:"applicationImage"`
	Stack            string       `yaml:"stack"`
	Expose           string       `yaml:"expose"`
	Service          ServiceField `yaml:"service"`
}

type ServiceField struct {
	Type string `yaml:"type"`
	Port string `yaml:"port"`
}

func (a *AppDeploy) validateAppDeploy(stackPath string, appDeployKeys []string) int {
	stackLintErrorCount := 0
	arg := filepath.Join(stackPath, "image", "config", "app-deploy.yaml")
	var deployFileContents []string

	Info.log("LINTING app-deploy.yaml: ", arg)

	appdeploy, err := ioutil.ReadFile(arg)
	if err != nil {
		Error.log("app deploy err ", err)
		stackLintErrorCount++
	}

	appDeployMap := make(map[string]interface{})
	err = yaml.UnmarshalStrict([]byte(appdeploy), &appDeployMap)
	if err != nil {
		Error.log("Unmarshal: Error unmarshalling app-deploy.yaml")
		stackLintErrorCount++
	}

	mapString := make(map[string]interface{})

	for x, v := range appDeployMap {
		if b, ok := v.(map[interface{}]interface{}); ok {
			for key, value := range b {
				strKey := fmt.Sprintf("%v", key)
				mapString[strKey] = value
			}
			for z := range mapString {
				deployFileContents = append(deployFileContents, z)
			}
		} else {
			deployFileContents = append(deployFileContents, x)
		}
	}

	variableFound := false

	for _, keys := range appDeployKeys {
		for _, key1 := range deployFileContents {
			if keys == key1 {
				variableFound = true
			}
		}
		if !variableFound {
			Error.log("Missing field: ", keys)
			stackLintErrorCount++
		}
		variableFound = false
	}

	return stackLintErrorCount
}
