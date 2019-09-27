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
	"path/filepath"
	"io/ioutil"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

//Structure for an app deploy file
type AppDeploy struct {
	APIVersion string		`yaml:"apiVersion"`
	Kind string				`yaml:"kind"`
	Metadata MetadataField 	`yaml:"metadata"`	
	Spec SpecField 			`yaml:"spec"`
}

type MetadataField struct {
	Name string   `yaml:"name"`
}

type SpecField struct {
	Version string 			`yaml:"version"`
	ApplicationImage string `yaml:"applicationImage"`
	Stack string 			`yaml:"stack"`
	Expose string 			`yaml:"expose"`
	Service ServiceField 	`yaml:"service"`
}

type ServiceField struct {
	Type string `yaml:"type"`
	Port string `yaml:"port"`
}


func (a *AppDeploy) validateAppDeploy(stackPath string) *AppDeploy {
	arg := filepath.Join(stackPath, "image", "config", "app-deploy.yaml")

	Info.log("LINTING app-deploy.yaml: ", arg)

	appdeploy, err := ioutil.ReadFile(arg) //Read app deploy file
	if err != nil {
		Error.log("app deploy err ", err)
		stackLintErrorCount++
	}

	dynamic := make(map[string]interface{})
	err = yaml.UnmarshalStrict([]byte(appdeploy), &dynamic) //Unmarshal contents of file into var of type AppDeploy (struct)
	if err != nil {
		Error.log("Unmarshal: Error unmarshalling app-deploy.yaml")
		stackLintErrorCount++
	}

	for k, v := range dynamic { 
		Info.log(k, "    ", v)
	}

	Info.log(dynamic["metadata"])


	validateFields(a)

	return a
}

func validateFields(a interface{}) {
	v := reflect.ValueOf(a).Elem()
	yamlValues := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ { //Iterate over values, if app-deploy.yaml field is missing - throw error
		yamlValues[i] = v.Field(i).Interface()
		if yamlValues[i] == "" {
			Error.log("Missing value for field: ", strings.ToLower(v.Type().Field(i).Name))
			stackLintErrorCount++
		}
	}

	// if a.Metadata.Name == "" {
	// 	Error.log("Missing value for field: ", a.Metadata.Name)
	// }
	
	// t := reflect.ValueOf(yamlValues[2]).Elem()
	// yamlValues1 := make([]interface{}, t.NumField())

	// for j := 0; j < t.NumField(); j++ {
	// 	yamlValues1 = t.Field(j).Interface()
	// }
	
}