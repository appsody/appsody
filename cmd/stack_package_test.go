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

package cmd_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Tests all templating variables with a starter stack in testdata.  Does not test for .stack.created.
func TestTemplatingAllVariables(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// gets all the necessary data from a setup function
	imageNamespace, imageRegistry, stackYaml, labels, err := setupStackPackageTests(sandbox.TestDataPath)
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	// creates templating.txt file where templating variables will appear
	templatingPath := filepath.Join(sandbox.TestDataPath, "templating", "templating.txt")
	err = os.MkdirAll(filepath.Dir(templatingPath), 0777)
	if err != nil {
		t.Fatalf("Error creating templating dir: %v", err)
	}
	file, err := os.Create(templatingPath)
	if err != nil {
		t.Fatalf("Error creating templating file: %v", err)
	}

	// write some text to file
	_, err = file.WriteString("{{test}}, id: {{.stack.id}}, name: {{.stack.name}}, version: {{.stack.version}}, description: {{.stack.description}}, tag: {{.stack.tag}}, maintainers: {{.stack.maintainers}}, semver.major: {{.stack.semver.major}}, semver.minor: {{.stack.semver.minor}}, semver.patch: {{.stack.semver.patch}}, semver.majorminor: {{.stack.semver.majorminor}}, image.namespace: {{.stack.image.namespace}}, image.registry: {{.stack.image.registry}}, customvariable1: {{.stack.variable1}}, customvariable2: {{.stack.variable2}}")
	if err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	// save file changes
	err = file.Sync()
	if err != nil {
		t.Fatalf("Error saving file: %v", err)
	}

	// create the template metadata
	templateMetadata, err := cmd.CreateTemplateMap(labels, stackYaml, imageNamespace, imageRegistry)
	if err != nil {
		t.Fatalf("Error creating template map: %v", err)
	}

	// apply templating to stack
	err = cmd.ApplyTemplating(templatingPath, templateMetadata)
	if err != nil {
		t.Fatalf("Error applying template: %v", err)
	}

	// read the whole file at once
	b, err := ioutil.ReadFile(templatingPath)
	if err != nil {
		t.Fatalf("Error reading templating file: %v", err)
	}
	s := string(b)
	t.Log(s)
	if !strings.Contains(s, "{{test}}, id: starter, name: Starter Sample, version: 0.1.1, description: Runnable starter stack, copy to create a new stack, tag: appsody/starter:SNAPSHOT, maintainers: Henry Nash <henry.nash@uk.ibm.com>, semver.major: 0, semver.minor: 1, semver.patch: 1, semver.majorminor: 0.1, image.namespace: appsody, image.registry: dev.local, customvariable1: value1, customvariable2: value2") {
		t.Fatal("Templating text did not match expected values")
	}

}

// Test templating with an incorrect variable name
func TestTemplatingWrongVariablesFail(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// gets all the necessary data from a setup function
	imageNamespace, imageRegistry, stackYaml, labels, err := setupStackPackageTests(sandbox.TestDataPath)
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	// creates templating.txt file where templating variables will appear
	templatingPath := filepath.Join(sandbox.TestDataPath, "templating", "templating.txt")
	err = os.MkdirAll(filepath.Dir(templatingPath), 0777)
	if err != nil {
		t.Fatalf("Error creating templating dir: %v", err)
	}
	file, err := os.Create(templatingPath)
	if err != nil {
		t.Fatalf("Error creating templating file: %v", err)
	}

	// write some text to file
	_, err = file.WriteString("id: {{.stack.iad}}")
	if err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	// save file changes
	err = file.Sync()
	if err != nil {
		t.Fatalf("Error saving file: %v", err)
	}

	// create the template metadata
	templateMetadata, err := cmd.CreateTemplateMap(labels, stackYaml, imageNamespace, imageRegistry)
	if err != nil {
		t.Fatalf("Error creating template map: %v", err)
	}

	// apply templating to stack
	err = cmd.ApplyTemplating(templatingPath, templateMetadata)
	if err != nil {
		t.Fatalf("Error applying template: %v", err)
	}

	// read the whole file at once
	b, err := ioutil.ReadFile(templatingPath)
	if err != nil {
		t.Fatalf("Error reading templating file: %v", err)
	}
	s := string(b)
	t.Log(s)
	if !strings.Contains(s, "id: <no value>") {
		t.Fatal("Templating text did not match expected values")
	}

}

func TestTemplatingFilePermissionsFail(t *testing.T) {

	// file permissions do not work the same way on windows
	// user has to specify a RUN chmod in their dockerfile for windows
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// gets all the necessary data from a setup function
	imageNamespace, imageRegistry, stackYaml, labels, err := setupStackPackageTests(sandbox.TestDataPath)
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	// creates templating.txt file where templating variables will appear
	templatingPath := filepath.Join(sandbox.TestDataPath, "templating", "templating.txt")
	err = os.MkdirAll(filepath.Dir(templatingPath), 0777)
	if err != nil {
		t.Fatalf("Error creating templating dir: %v", err)
	}
	file, err := os.Create(templatingPath)
	if err != nil {
		t.Fatalf("Error creating templating file: %v", err)
	}

	// write some text to file
	_, err = file.WriteString("id: {{.stack.id}}")
	if err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}
	// make file read only
	err = file.Chmod(0400)
	if err != nil {
		t.Fatalf("Error changing file permissions: %v", err)
	}

	// save file changes
	err = file.Sync()
	if err != nil {
		t.Fatalf("Error saving file: %v", err)
	}

	// create the template metadata
	templateMetadata, err := cmd.CreateTemplateMap(labels, stackYaml, imageNamespace, imageRegistry)
	if err != nil {
		t.Fatalf("Error creating template map: %v", err)
	}

	// apply templating to stack
	err = cmd.ApplyTemplating(templatingPath, templateMetadata)
	if err != nil {
		t.Fatalf("Error applying template: %v", err)
	}

	// read the whole file at once
	b, err := ioutil.ReadFile(templatingPath)
	if err != nil {
		t.Fatalf("Error reading templating file: %v", err)
	}
	s := string(b)
	t.Log(s)
	if !strings.Contains(s, "id: starter") {
		t.Fatal("Templating text did not match expected values")
	}

	writable, err := canWrite(templatingPath)

	if writable && err == nil {
		t.Fatal("Opened read only file")
	}

}

func TestPackageMissingFilesFail(t *testing.T) {

	var targetFiles = []struct {
		testName           string
		target             string
		expectedLogWindows string
		expectedLogDefault string
	}{
		{"No stack.yaml", "stack.yaml", "stack.yaml: The system cannot find the file specified", "stack.yaml: no such file or directory"},
		{"No Templates Folder", "templates", "Unable to reach templates directory. Current directory must be the root of the stack", "Unable to reach templates directory. Current directory must be the root of the stack"},
	}

	for _, testData := range targetFiles {

		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

			target := filepath.Join(sandbox.ProjectDir, tt.target)

			err := os.RemoveAll(target)
			if err != nil {
				t.Fatal(err)
			}

			// run stack package
			args := []string{"stack", "package"}
			_, err = cmdtest.RunAppsody(sandbox, args...)

			if err != nil {
				if runtime.GOOS == "windows" {
					if !strings.Contains(err.Error(), tt.expectedLogWindows) {
						t.Errorf("String \""+tt.expectedLogWindows+"\" not found in output: '%v'", err.Error())
					}
				} else {
					if !strings.Contains(err.Error(), tt.expectedLogDefault) {
						t.Errorf("String \""+tt.expectedLogDefault+"\" not found in output: '%v'", err.Error())
					}
				}
			} else {
				t.Fatal("Stack package command unexpectedly passed with no stack.yaml present")
			}
		})
	}
}

func TestInvalidStackYaml(t *testing.T) {
	var targetLines = []struct {
		testName        string
		targetLine      string
		replacementLine string
		expectedLog     string
	}{
		{"Invalid default-template", "default-template", "Enrique Was Here", "Error parsing the stack.yaml file"},
		{"Invalid custom variable", "variable1", "  ^variable1: value1", "Error creating templating mal: Variable name didn't start with alphanumeric character"},
		{"Invalid custom varable map", "variable1", "  variable1: \n    value1: s", "cannot unmarshal !!map into string"},
		{"Invalid version", "version:", "version: 0.1", "Error creating templating mal: Version format incorrect"},
	}

	for _, testData := range targetLines {

		tt := testData

		t.Run(tt.testName, func(t *testing.T) {

			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

			yamlPath := filepath.Join(sandbox.ProjectDir, "stack.yaml")

			file, err := ioutil.ReadFile(yamlPath)
			if err != nil {
				t.Fatal(err)
			}

			lines := strings.Split(string(file), "\n")

			for i, line := range lines {
				if strings.Contains(line, tt.targetLine) {
					lines[i] = tt.replacementLine
				}
			}
			output := strings.Join(lines, "\n")
			err = ioutil.WriteFile(yamlPath, []byte(output), 0644)

			if err != nil {
				t.Fatal(err)
			}

			args := []string{"stack", "package"}
			_, err = cmdtest.RunAppsody(sandbox, args...)

			if err != nil {
				if !strings.Contains(err.Error(), tt.expectedLog) {
					t.Errorf("String \""+tt.expectedLog+"\" not found in output: '%v'", err.Error())
				}

			} else {
				t.Fatal("Stack package command unexpectedly passed with invalid stack.yaml")
			}
		})
	}
}

// function that returns a boolean if the file is writable or not
func canWrite(filepath string) (bool, error) {
	file, err := os.OpenFile(filepath, os.O_WRONLY, 0666)
	if err != nil {
		if os.IsPermission(err) {
			return false, err
		}
	}
	file.Close()
	return true, nil

}

func setupStackPackageTests(testDataPath string) (string, string, cmd.StackYaml, map[string]string, error) {
	var loggingConfig = &cmd.LoggingConfig{}
	loggingConfig.InitLogging(os.Stdout, os.Stderr)
	var rootConfig = &cmd.RootCommandConfig{LoggingConfig: loggingConfig}
	var labels = map[string]string{}
	var stackYaml cmd.StackYaml
	stackID := "starter"
	imageNamespace := "appsody"
	imageRegistry := "dev.local"
	buildImage := imageNamespace + "/" + stackID + ":SNAPSHOT"
	projectPath := filepath.Join(testDataPath, "starter")

	rootConfig.ProjectDir = projectPath
	rootConfig.Dryrun = false

	err := cmd.InitConfig(rootConfig)
	if err != nil {
		return imageNamespace, imageRegistry, stackYaml, labels, errors.Errorf("Error getting config: %v", err)
	}

	source, err := ioutil.ReadFile(filepath.Join(projectPath, "stack.yaml"))
	if err != nil {
		return imageNamespace, imageRegistry, stackYaml, labels, errors.Errorf("Error reading stackyaml: %v", err)
	}

	err = yaml.Unmarshal(source, &stackYaml)
	if err != nil {
		return imageNamespace, imageRegistry, stackYaml, labels, errors.Errorf("Error parsing stackyaml: %v", err)
	}

	labels, err = cmd.GetLabelsForStackImage(stackID, buildImage, stackYaml, rootConfig)
	if err != nil {
		return imageNamespace, imageRegistry, stackYaml, labels, errors.Errorf("Error getting labels: %v", err)
	}

	return imageNamespace, imageRegistry, stackYaml, labels, err

}
