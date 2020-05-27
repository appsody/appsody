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

	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v2"
)

func (stackDetails *StackYaml) validateYaml(rootConfig *RootCommandConfig, stackPath string) (int, int) {
	stackLintErrorCount := 0
	stackLintWarningCount := 0
	arg := filepath.Join(stackPath, "/stack.yaml")

	rootConfig.Info.log("LINTING stack.yaml: ", arg)

	stackyaml, err := ioutil.ReadFile(arg)
	if err != nil {
		rootConfig.Error.log("stackyaml.Get err ", err)
		stackLintErrorCount++
	}

	err = yaml.Unmarshal([]byte(stackyaml), stackDetails)
	if err != nil {
		rootConfig.Error.log("Unmarshal: Error unmarshalling stack.yaml")
		stackLintErrorCount++
	}

	stackLintErrorCount += stackDetails.validateFields(rootConfig)
	validSemver := CheckValidSemver(string(stackDetails.Version))
	if validSemver != nil {
		rootConfig.Error.log(validSemver)
		stackLintErrorCount++
	}

	stackLintErrorCount += stackDetails.checkDescLength(rootConfig.LoggingConfig)
	stackLintWarningCount += stackDetails.checkLicense(rootConfig.LoggingConfig)
	stackLintErrorCount += stackDetails.checkRequirements(rootConfig.LoggingConfig)
	templateErrorCount, templateWarningCount := stackDetails.checkTemplatingData(rootConfig.LoggingConfig)
	stackLintErrorCount += templateErrorCount
	stackLintWarningCount += templateWarningCount
	return stackLintErrorCount, stackLintWarningCount
}

func (stackDetails *StackYaml) validateFields(rootConfig *RootCommandConfig) int {
	stackLintErrorCount := 0
	v := reflect.ValueOf(stackDetails).Elem()
	yamlValues := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		yamlValues[i] = v.Field(i).Interface()
		if yamlValues[i] == "" && v.Type().Field(i).Name != "Deprecated" {
			rootConfig.Error.log("Missing value for field: ", strings.ToLower(v.Type().Field(i).Name))
			stackLintErrorCount++
		}
	}

	stackLintErrorCount += stackDetails.checkMaintainer(rootConfig.LoggingConfig, yamlValues)
	return stackLintErrorCount

}

func (stackDetails *StackYaml) checkMaintainer(log *LoggingConfig, yamlValues []interface{}) int {
	stackLintErrorCount := 0
	Map := make(map[string]interface{})
	Map["maintainerEmails"] = yamlValues[5]

	maintainerEmails := Map["maintainerEmails"].([]Maintainer)

	if len(maintainerEmails) == 0 {
		log.Error.log("Please list a stack maintainer with the following details: Name, Email, and Github ID")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}

func (stackDetails *StackYaml) checkDescLength(log *LoggingConfig) int {
	stackLintErrorCount := 0

	if len(stackDetails.Description) > 70 {
		log.Error.log("Description must be under 70 characters")
		stackLintErrorCount++
	}

	if len(stackDetails.Name) > 30 {
		log.Error.log("Stack name must be under 30 characters")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}

func (stackDetails *StackYaml) checkTemplatingData(log *LoggingConfig) (int, int) {
	stackLintErrorCount := 0
	stackLintWarningCount := 0
	keyRegex := regexp.MustCompile("^[a-zA-Z0-9]*$")

	if len(stackDetails.TemplatingData) == 0 {
		log.Info.log("No custom stack variables used (see https://appsody.dev/docs/stacks/develop/#custom-stack-variables)")
		return stackLintErrorCount, stackLintWarningCount
	}

	for key := range stackDetails.TemplatingData {
		checkKey := keyRegex.FindString(string(key))

		if checkKey == "" {
			log.Error.log("stack.yaml templating-data key is not alphanumeric: ", key)
			stackLintErrorCount++
		}

	}
	return stackLintErrorCount, stackLintWarningCount
}

func (stackDetails *StackYaml) checkLicense(log *LoggingConfig) int {
	stackLintWarningCount := 0

	if err := checkValidLicense(log, stackDetails.License); err != nil {
		stackLintWarningCount++
		log.Warning.logf("The stack.yaml SPDX license ID is invalid: %v.", err)
	}
	if valid, err := IsValidKubernetesLabelValue(stackDetails.License); !valid {
		stackLintWarningCount++
		log.Warning.logf("The stack.yaml SPDX license ID is invalid: %v.", err)
	}
	return stackLintWarningCount
}

func (stackDetails *StackYaml) checkRequirements(log *LoggingConfig) int {
	stackLintErrorCount := 0

	reqsMap := map[string]string{
		"Docker":  stackDetails.Requirements.Docker,
		"Appsody": stackDetails.Requirements.Appsody,
		"Buildah": stackDetails.Requirements.Buildah,
	}

	for _, req := range reqsMap {
		if req == "" {
			continue
		}
		_, err := semver.NewConstraint(req)
		if err != nil {
			log.Error.log("Requirement: ", req, " is not in the correct format. See: https://github.com/Masterminds/semver for a list of valid requirement constraints.")
			stackLintErrorCount++
		}
	}

	return stackLintErrorCount
}
