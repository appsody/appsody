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
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Tests all templating variables with a starter stack in testdata.  Does not test for .stack.created.
func TestTemplatingAllVariables(t *testing.T) {

	// gets all the necessary data from a setup function
	imageNamespace, projectPath, stackPath, stackYaml, labels, err := setup()
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	// creates templating.txt file where templating variables will appear
	file, err := os.Create("./testdata/starter/templating.txt")
	if err != nil {
		t.Fatalf("Error creating templating file: %v", err)
	}

	defer os.RemoveAll("./testdata/starter/templating.txt")
	defer os.RemoveAll(stackPath)

	// write some text to file
	_, err = file.WriteString("id: {{.stack.id}}, name: {{.stack.name}}, version: {{.stack.version}}, description: {{.stack.description}}, tag: {{.stack.tag}}, maintainers: {{.stack.maintainers}}, semver.major: {{.stack.semver.major}}, semver.minor: {{.stack.semver.minor}}, semver.patch: {{.stack.semver.patch}}, semver.majorminor: {{.stack.semver.majorminor}}, image.namespace: {{.stack.image.namespace}}, customvariable1: {{.stack.variable1}}, customvariable2: {{.stack.variable2}}")
	if err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	// save file changes
	err = file.Sync()
	if err != nil {
		t.Fatalf("Error saving file: %v", err)
	}

	err = cmd.CopyDir(projectPath, stackPath)
	if err != nil {
		t.Fatalf("Error copying directory: %v", err)
	}

	// create the template metadata
	templateMetadata, err := cmd.CreateTemplateMap(labels, stackYaml, imageNamespace)
	if err != nil {
		t.Fatalf("Error creating template map: %v", err)
	}

	// apply templating to stack
	err = cmd.ApplyTemplating(projectPath, stackPath, templateMetadata)
	if err != nil {
		t.Fatalf("Error applying template: %v", err)
	}

	// read the whole file at once
	b, err := ioutil.ReadFile(stackPath + "/templating.txt")
	if err != nil {
		panic(err)
	}
	s := string(b)
	if !strings.Contains(s, "id: starter, name: Starter Sample, version: 0.1.1, description: Runnable starter stack, copy to create a new stack, tag: dev.local/starter:SNAPSHOT, maintainers: Henry Nash <henry.nash@uk.ibm.com>, semver.major: 0, semver.minor: 1, semver.patch: 1, semver.majorminor: 0.1, image.namespace: dev.local, customvariable1: value1, customvariable2: value2") {
		t.Fatal("Templating text did not match expected values")
	}

}

func setup() (string, string, string, cmd.StackYaml, map[string]string, error) {

	var rootConfig = &cmd.RootCommandConfig{}
	var labels = map[string]string{}
	var stackPath string
	var stackYaml cmd.StackYaml
	stackID := "starter"
	imageNamespace := "dev.local"
	buildImage := imageNamespace + "/" + stackID + ":SNAPSHOT"
	// sets stack path to be the copied folder
	projectPath, err := filepath.Abs("./testdata/starter")

	if err != nil {
		return imageNamespace, projectPath, stackPath, stackYaml, labels, errors.Errorf("Error getting labels: %v", err)
	}

	rootConfig.ProjectDir = projectPath
	rootConfig.Dryrun = false
	err = cmd.InitConfig(rootConfig)

	if err != nil {
		return imageNamespace, projectPath, stackPath, stackYaml, labels, errors.Errorf("Error getting labels: %v", err)
	}

	stackPath = filepath.Join(rootConfig.CliConfig.GetString("home"), "stacks", "packaging-"+stackID)

	source, err := ioutil.ReadFile(filepath.Join(projectPath, "stack.yaml"))
	if err != nil {
		return imageNamespace, projectPath, stackPath, stackYaml, labels, errors.Errorf("Error getting labels: %v", err)
	}

	err = yaml.Unmarshal(source, &stackYaml)
	if err != nil {
		return imageNamespace, projectPath, stackPath, stackYaml, labels, errors.Errorf("Error getting labels: %v", err)
	}

	labels, err = cmd.GetLabelsForStackImage(stackID, buildImage, stackYaml, rootConfig)
	if err != nil {
		return imageNamespace, projectPath, stackPath, stackYaml, labels, errors.Errorf("Error getting labels: %v", err)
	}

	return imageNamespace, projectPath, stackPath, stackYaml, labels, err

}