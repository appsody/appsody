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

var DefaultFound bool

func (s *StackDetails) validateYaml() *StackDetails {
	stackPath, _ := os.Getwd()

	if len(os.Args) > 3 {
		stackPath = os.Args[3]
	}

	arg := filepath.Join(stackPath, "/stack.yaml")

	stackyaml, err := ioutil.ReadFile(arg)
	if err != nil {
		Error.log("stackyaml.Get err ", err)
		stackLintErrorCount++
	}

	err = yaml.Unmarshal([]byte(stackyaml), s)
	if err != nil {
		Error.log("Unmarshal: ", err)
		stackLintErrorCount++
	}

	s.validateFields(arg)
	return s
}

func (s *StackDetails) validateFields(arg string) *StackDetails {
	v := reflect.ValueOf(s).Elem()
	yamlValues := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		yamlValues[i] = v.Field(i).Interface()
		if yamlValues[i] == "" {
			Error.log("Missing value for field: ", strings.ToLower(v.Type().Field(i).Name), " in ", arg)
			stackLintErrorCount++
		}
	}

	s.checkMaintainer(arg, yamlValues)

	return s

}

func (s *StackDetails) checkMaintainer(arg string, yamlValues []interface{}) *StackDetails {
	Map := make(map[string]interface{})
	Map["maintainerEmails"] = yamlValues[5]

	maintainerEmails := Map["maintainerEmails"].([]StackMaintainer)

	if len(maintainerEmails) == 0 {
		Error.log("Email is not provided under field: maintainers in ", arg)
	}

	s.checkVersion(arg)
	return s
}

func (s *StackDetails) checkVersion(arg string) *StackDetails {
	versionNo := strings.Split(s.Version, ".")

	for _, mmp := range versionNo {
		_, err := strconv.Atoi(mmp)
		if err != nil {
			Error.log("Each version field must be an integer in ", arg)
		}
	}

	if len(versionNo) < 3 {
		Error.log("Version must contain 3 or 4 elements in ", arg)
		stackLintErrorCount++
	}

	s.checkDescLenth(arg)
	return s
}

func (s *StackDetails) checkDescLenth(arg string) *StackDetails {

	if len(s.Description) > 70 {
		Error.log("Description should be under 70 characters in ", arg)
		stackLintErrorCount++
	}

	s.checkDefaultTemplate(arg)
	return s
}

func (s *StackDetails) checkDefaultTemplate(arg string) *StackDetails {
	DefaultFound = false
	file, err := os.Open(arg)
	if err != nil {
		Error.log(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		yamlFields := strings.Split(scanner.Text(), ":")
		if yamlFields[0] == "default-template" {
			DefaultFound = true
		}
	}

	if err := scanner.Err(); err != nil {
		Error.log(err)
	}

	if !DefaultFound {
		Error.log("Missing value for field: default-template in ", arg)
		stackLintErrorCount++
	}

	return s
}
