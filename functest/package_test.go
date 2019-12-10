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

package functest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestPackageStarterStack(t *testing.T) {

	args := []string{"stack", "package"}
	_, err := cmdtest.RunAppsodyCmd(args, filepath.Join("..", "cmd", "testdata", "starter"), t)

	if err != nil {
		t.Fatal(err)
	}
}

func TestPackageNoStackYaml(t *testing.T) {

	// rename stack.yaml to test
	stackPath := filepath.Join("..", "cmd", "testdata", "starter")
	stackYaml := filepath.Join(stackPath, "stack.yaml")
	newStackYaml := filepath.Join(stackPath, "test")

	os.Rename(stackYaml, newStackYaml)
	defer os.Rename(newStackYaml, stackYaml)

	// run stack package
	args := []string{"stack", "package"}
	_, err := cmdtest.RunAppsodyCmd(args, stackPath, t)

	if err == nil { // stack package will fail as stack.yaml file does not exist
		t.Fatal(err)
	}

}

func TestPackageInvalidStackYaml(t *testing.T) {

	// add invalid line to stack.yaml
	stackPath := filepath.Join("..", "cmd", "testdata", "starter")
	stackYaml := filepath.Join(stackPath, "stack.yaml")

	restoreLine := ""
	file, err := ioutil.ReadFile(stackYaml)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(file), "\n")

	for i, line := range lines {
		if strings.Contains(line, "default-template") {
			restoreLine = lines[i]
			lines[i] = "Testing"
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(stackYaml, []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	// run stack package
	args := []string{"stack", "package"}
	_, err = cmdtest.RunAppsodyCmd(args, stackPath, t)

	if err == nil { // stack package will fail as stack.yaml has invalid foramtting
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "Testing") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile(stackYaml, []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

}

func TestPackageNoTemplates(t *testing.T) {

	// rename templates directory to test
	stackPath := filepath.Join("..", "cmd", "testdata", "starter")
	templates := filepath.Join(stackPath, "templates")
	newTemplates := filepath.Join(stackPath, "test")

	os.Rename(templates, newTemplates)
	defer os.Rename(newTemplates, templates)

	// run stack package
	args := []string{"stack", "package"}
	_, err := cmdtest.RunAppsodyCmd(args, stackPath, t)

	if err == nil { // stack package will fail as stack.yaml file does not exist
		t.Fatal(err)
	}

}
