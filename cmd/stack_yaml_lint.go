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
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

func (stackDetails *StackYaml) validateYaml(stackPath string) (int, int) {
	stackLintErrorCount := 0
	arg := filepath.Join(stackPath, "/stack.yaml")

	Info.log("LINTING stack.yaml: ", arg)

	stackyaml, err := ioutil.ReadFile(arg)
	if err != nil {
		Error.log("stackyaml.Get err ", err)
		stackLintErrorCount++
	}

	err = yaml.Unmarshal([]byte(stackyaml), stackDetails)
	if err != nil {
		Error.log("Unmarshal: Error unmarshalling stack.yaml")
		stackLintErrorCount++
	}

	stackLintErrorCount += stackDetails.validateFields()
	validSemver := CheckValidSemver(string(stackDetails.Version))
	if validSemver != nil {
		Error.log(validSemver)
		stackLintErrorCount++
	}
	stackLintErrorCount += stackDetails.checkDescLength()
	templateErrorCount, templateWarningCount := stackDetails.checkTemplatingData()
	stackLintErrorCount += templateErrorCount
	return stackLintErrorCount, templateWarningCount
}

func (stackDetails *StackYaml) validateFields() int {
	stackLintErrorCount := 0
	v := reflect.ValueOf(stackDetails).Elem()
	yamlValues := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		yamlValues[i] = v.Field(i).Interface()
		if yamlValues[i] == "" {
			Error.log("Missing value for field: ", strings.ToLower(v.Type().Field(i).Name))
			stackLintErrorCount++
		}
	}

	stackLintErrorCount += stackDetails.checkMaintainer(yamlValues)
	return stackLintErrorCount

}

func (stackDetails *StackYaml) checkMaintainer(yamlValues []interface{}) int {
	stackLintErrorCount := 0
	Map := make(map[string]interface{})
	Map["maintainerEmails"] = yamlValues[5]

	maintainerEmails := Map["maintainerEmails"].([]Maintainer)

	if len(maintainerEmails) == 0 {
		Error.log("Please list a stack maintainer with the following details: Name, Email, and Github ID")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}

func (stackDetails *StackYaml) checkDescLength() int {
	stackLintErrorCount := 0

	if len(stackDetails.Description) > 70 {
		Error.log("Description must be under 70 characters")
		stackLintErrorCount++
	}

	if len(stackDetails.Name) > 30 {
		Error.log("Stack name must be under 30 characters")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}

func (stackDetails *StackYaml) checkTemplatingData() (int, int) {
	stackLintErrorCount := 0
	stackLintWarningCount := 0
	keyRegex := regexp.MustCompile("^[a-zA-Z0-9]*$")

	if len(stackDetails.TemplatingData) == 0 {
		Warning.log("No custom templating variables defined - You will not be able to reuse variables across the stack")
		stackLintWarningCount++
		return stackLintErrorCount, stackLintWarningCount
	}

	for key, value := range stackDetails.TemplatingData {
		checkKey := keyRegex.FindString(string(key))
		checkValue := keyRegex.FindString(string(value))

		if checkKey == "" {
			Error.log("Key variable: ", key, " is not in an alphanumeric format")
			stackLintErrorCount++
		}

		if checkValue == "" {
			Error.log("Value associated with key: ", key, " is not in an alphanumeric format")
			stackLintErrorCount++
		}
	}

	return stackLintErrorCount, stackLintWarningCount
}
