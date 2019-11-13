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
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

func TestTemplatingAllValues(t *testing.T) {

	imageNamespace, projectPath, stackPath, stackYaml, labels, err := setup()
	if err != nil {
		t.Fatalf("Error during setup: %v", err)
	}

	file, err := os.Create("./testdata/starter/templating.txt")

	if err != nil {
		t.Fatal("Error creating templating file: %v")
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
