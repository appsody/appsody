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
	"github.com/appsody/appsody/cmd/cmdtest"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

func TestTemplatingNoTemplating(t *testing.T) {

	imageNamespace, projectPath, stackPath, stackYaml, labels, err := setup()

	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	err = cmd.CopyDir(projectPath, stackPath)

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(stackPath)

	// create the template metadata
	templateMetadata, err := cmd.CreateTemplateMap(labels, stackYaml, imageNamespace)

	if err != nil {
		t.Fatal(err)
	}

	// apply templating to stack
	err = cmd.ApplyTemplating(projectPath, stackPath, templateMetadata)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(stackPath)

	if !exists {
		t.Fatal(err)
	}

}

func TestTemplatingAllValues(t *testing.T) {

	imageNamespace, projectPath, stackPath, stackYaml, labels, err := setup()
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	restoreLine := ""
	projectFile, err := ioutil.ReadFile(projectPath + "/templates/simple/hello.sh")
	if err != nil {
		t.Fatal(err)
	}

	projectLines := strings.Split(string(projectFile), "\n")

	for i, line := range projectLines {
		if strings.Contains(line, "Hello from Appsody!") {
			restoreLine = projectLines[i]
			projectLines[i] = "id: {{.stack.id}}, name: {{.stack.name}}, version: {{.stack.version}}, description: {{.stack.description}}, created: {{.stack.created}}, tag: {{.stack.tag}}, maintainers: {{.stack.maintainers}}, semver.major: {{.stack.semver.major}}, semver.minor: {{.stack.semver.minor}}, semver.patch: {{.stack.semver.patch}}, semver.majorminor: {{.stack.semver.majorminor}}, image.namespace: {{.stack.image.namespace}}, customvariable1: {{.stack.variable1}}, customvariable2: {{.stack.variable2}}"
		}
	}
	output := strings.Join(projectLines, "\n")
	err = ioutil.WriteFile((projectPath + "/templates/simple/hello.sh"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	err = cmd.CopyDir(projectPath, stackPath)

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(stackPath)

	// create the template metadata
	templateMetadata, err := cmd.CreateTemplateMap(labels, stackYaml, imageNamespace)

	if err != nil {
		t.Fatal(err)
	}

	// apply templating to stack
	err = cmd.ApplyTemplating(projectPath, stackPath, templateMetadata)

	if err != nil {
		t.Fatal(err)
	}

	stackFile, err := ioutil.ReadFile(stackPath + "/templates/simple/hello.sh")
	if err != nil {
		t.Fatal(err)
	}

	stackLines := strings.Split(string(stackFile), "\n")

	for _, line := range stackLines {
		if line != "id: starter, name: Starter Sample, version: 0.1.1, description: Runnable starter stack, copy to create a new stack, created: 2019-11-12T13:46:10Z, tag: dev.local/starter:SNAPSHOT, maintainers: Henry Nash <henry.nash@uk.ibm.com>, semver.major: 0, semver.minor: 1, semver.patch: 1, semver.majorminor: 0.1, image.namespace: dev.local, customvariable1: value1, customvariable2: value2" {
			t.Fatalf("Templating text does not match")
		}
	}

	for i, line := range projectLines {
		if line == "id: {{.stack.id}}, name: {{.stack.name}}, version: {{.stack.version}}, description: {{.stack.description}}, created: {{.stack.created}}, tag: {{.stack.tag}}, maintainers: {{.stack.maintainers}}, semver.major: {{.stack.semver.major}}, semver.minor: {{.stack.semver.minor}}, semver.patch: {{.stack.semver.patch}}, semver.majorminor: {{.stack.semver.majorminor}}, image.namespace: {{.stack.image.namespace}}, customvariable1: {{.stack.variable1}}, customvariable2: {{.stack.variable2}}" {
			projectLines[i] = restoreLine
		}
	}

	output = strings.Join(projectLines, "\n")
	err = ioutil.WriteFile((projectPath + "/templates/simple/hello.sh"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
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
