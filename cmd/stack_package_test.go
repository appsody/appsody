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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Tests all templating variables with a starter stack in testdata.  Does not test for .stack.created.
func TestTemplatingAllVariables(t *testing.T) {

	// gets all the necessary data from a setup function
	imageNamespace, imageRegistry, stackYaml, labels, err := setupStackPackageTests()
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	// creates templating.txt file where templating variables will appear
	templatingPath := filepath.Join(".", "testdata", "templating", "templating.txt")
	err = os.MkdirAll(filepath.Dir(templatingPath), 0777)
	if err != nil {
		t.Fatalf("Error creating templating dir: %v", err)
	}
	file, err := os.Create(templatingPath)
	if err != nil {
		t.Fatalf("Error creating templating file: %v", err)
	}

	defer os.RemoveAll(templatingPath)

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
func TestTemplatingWrongVariables(t *testing.T) {

	// gets all the necessary data from a setup function
	imageNamespace, imageRegistry, stackYaml, labels, err := setupStackPackageTests()
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	// creates templating.txt file where templating variables will appear
	templatingPath := filepath.Join(".", "testdata", "templating", "templating.txt")
	err = os.MkdirAll(filepath.Dir(templatingPath), 0777)
	if err != nil {
		t.Fatalf("Error creating templating dir: %v", err)
	}
	file, err := os.Create(templatingPath)
	if err != nil {
		t.Fatalf("Error creating templating file: %v", err)
	}

	defer os.RemoveAll(templatingPath)

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

func TestTemplatingFilePermissions(t *testing.T) {

	// file permissions do not work the same way on windows
	// user has to specify a RUN chmod in their dockerfile for windows
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	// gets all the necessary data from a setup function
	imageNamespace, imageRegistry, stackYaml, labels, err := setupStackPackageTests()
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	// creates templating.txt file where templating variables will appear
	templatingPath := filepath.Join(".", "testdata", "templating", "templating.txt")
	err = os.MkdirAll(filepath.Dir(templatingPath), 0777)
	if err != nil {
		t.Fatalf("Error creating templating dir: %v", err)
	}
	file, err := os.Create(templatingPath)
	if err != nil {
		t.Fatalf("Error creating templating file: %v", err)
	}

	defer os.RemoveAll(templatingPath)

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

func setupStackPackageTests() (string, string, cmd.StackYaml, map[string]string, error) {
	var loggingConfig = &cmd.LoggingConfig{}
	loggingConfig.InitLogging(os.Stdout, os.Stderr)
	var rootConfig = &cmd.RootCommandConfig{LoggingConfig: loggingConfig}
	var labels = map[string]string{}
	var stackYaml cmd.StackYaml
	stackID := "starter"
	imageNamespace := "appsody"
	imageRegistry := "dev.local"
	buildImage := imageNamespace + "/" + stackID + ":SNAPSHOT"
	projectPath := filepath.Join(".", "testdata", "starter")

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
