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
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type StackDetails struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	License     string            `yaml:"license"`
	Language    string            `yaml:"language"`
	Maintainers []StackMaintainer `yaml:"maintainers"`
}

type StackMaintainer struct {
	Email string `yaml:"email"`
}

func (s *StackDetails) validateYaml(rootConfig *RootCommandConfig, stackPath string) int {
	stackLintErrorCount := 0
	arg := filepath.Join(stackPath, "/stack.yaml")

	rootConfig.Info.log("LINTING stack.yaml: ", arg)

	stackyaml, err := ioutil.ReadFile(arg)
	if err != nil {
		rootConfig.Error.log("stackyaml.Get err ", err)
		stackLintErrorCount++
	}

	err = yaml.Unmarshal([]byte(stackyaml), s)
	if err != nil {
		rootConfig.Error.log("Unmarshal: Error unmarshalling stack.yaml")
		stackLintErrorCount++
	}

	stackLintErrorCount += s.checkDefaultTemplate(rootConfig.LoggingConfig, arg)
	stackLintErrorCount += s.validateFields(rootConfig)
	stackLintErrorCount += s.checkVersion(rootConfig.LoggingConfig)
	stackLintErrorCount += s.checkDescLength(rootConfig.LoggingConfig)
	return stackLintErrorCount
}

func (s *StackDetails) checkDefaultTemplate(log *LoggingConfig, arg string) int {
	stackLintErrorCount := 0
	defaultTemplateFound := false
	file, err := os.Open(arg)
	if err != nil {
		log.Error.log(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		yamlFields := strings.Split(scanner.Text(), ":")
		if yamlFields[0] == "default-template" && yamlFields[1] != "" {
			defaultTemplateFound = true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error.log(err)
	}

	if !defaultTemplateFound {
		log.Error.log("Missing value for field: default-template")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}

func (s *StackDetails) validateFields(rootConfig *RootCommandConfig) int {
	stackLintErrorCount := 0
	v := reflect.ValueOf(s).Elem()
	yamlValues := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		yamlValues[i] = v.Field(i).Interface()
		if yamlValues[i] == "" {
			rootConfig.Error.log("Missing value for field: ", strings.ToLower(v.Type().Field(i).Name))
			stackLintErrorCount++
		}
	}

	stackLintErrorCount += s.checkMaintainer(rootConfig.LoggingConfig, yamlValues)
	return stackLintErrorCount

}

func (s *StackDetails) checkMaintainer(log *LoggingConfig, yamlValues []interface{}) int {
	stackLintErrorCount := 0
	Map := make(map[string]interface{})
	Map["maintainerEmails"] = yamlValues[5]

	maintainerEmails := Map["maintainerEmails"].([]StackMaintainer)

	if len(maintainerEmails) == 0 {
		log.Error.log("Email is not provided under field: maintainers")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}

func (s *StackDetails) checkVersion(log *LoggingConfig) int {
	stackLintErrorCount := 0
	versionNo := strings.Split(s.Version, ".")

	for _, mmp := range versionNo {
		_, err := strconv.Atoi(mmp)
		if err != nil {
			log.Error.log("Each version field must be an integer")
		}
	}

	if len(versionNo) < 3 {
		log.Error.log("Version must contain 3 or 4 elements")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}

func (s *StackDetails) checkDescLength(log *LoggingConfig) int {
	stackLintErrorCount := 0

	if len(s.Description) > 70 {
		log.Error.log("Description must be under 70 characters")
		stackLintErrorCount++
	}

	if len(s.Name) > 30 {
		log.Error.log("Stack name must be under 30 characters")
		stackLintErrorCount++
	}

	return stackLintErrorCount
}
