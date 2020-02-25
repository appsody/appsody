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
	"path/filepath"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
	"sigs.k8s.io/yaml"
)

func TestRemoveFromIncubatorRepo(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// run stack remove-from-repo
	args := []string{"stack", "remove-from-repo", "incubator", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	devLocal := filepath.Join(sandbox.ConfigDir, "stacks", "dev.local")
	indexFileLocal := filepath.Join(devLocal, "incubator-index.yaml")
	var indexYaml cmd.IndexYaml

	source, err := ioutil.ReadFile(indexFileLocal)
	if err != nil {
		t.Fatalf("Error trying to read: %v", err)
	}

	err = yaml.Unmarshal(source, &indexYaml)
	if err != nil {
		t.Fatalf("Error trying to unmarshall: %v", err)
	}

	foundStack := -1
	for i, stack := range indexYaml.Stacks {
		if stack.ID == "nodejs" {
			foundStack = i
			break
		}
	}
	if foundStack != -1 {
		t.Fatal("Stack found unexpectedly")
	}
}

func TestRemoveFromRepoLocalCache(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// run stack remove-from-repo
	args := []string{"stack", "remove-from-repo", "incubator", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	// run stack remove-from-repo use local cache
	argsRemoveLC := []string{"stack", "remove-from-repo", "incubator", "kitura", "--use-local-cache"}
	_, err = cmdtest.RunAppsody(sandbox, argsRemoveLC...)
	if err != nil {
		t.Fatal(err)
	}

	devLocal := filepath.Join(sandbox.ConfigDir, "stacks", "dev.local")
	indexFileLocal := filepath.Join(devLocal, "incubator-index.yaml")
	var indexYaml cmd.IndexYaml

	source, err := ioutil.ReadFile(indexFileLocal)
	if err != nil {
		t.Fatalf("Error trying to read: %v", err)
	}

	err = yaml.Unmarshal(source, &indexYaml)
	if err != nil {
		t.Fatalf("Error trying to unmarshall: %v", err)
	}

	foundStack := -1
	for i, stack := range indexYaml.Stacks {
		if stack.ID == "nodejs" || stack.ID == "kitura" {
			foundStack = i
			break
		}
	}
	if foundStack != -1 {
		t.Fatal("Stack found unexpectedly")
	}
}

func TestRemoveFromRepoInvalidRepoName(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// run stack remove-from-repo
	args := []string{"stack", "remove-from-repo", "invalid", "nodejs"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		if !strings.Contains(output, "invalid does not exist within the repository list") {
			t.Errorf("String \"invalid does not exist within the repository list\" not found in output: '%v'", err.Error())
		}

	} else {
		t.Fatal("stack remove-from-repo command unexpectedly passed with an invalid repo name")
	}
}

func TestRemoveFromRepoInvalidStackName(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// run stack remove-from-repo
	args := []string{"stack", "remove-from-repo", "incubator", "invalid"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	} else {
		if !strings.Contains(output, "Stack: invalid does not exist in repository index file") {
			t.Errorf("String \"Stack: invalid does not exist in repository index file\" not found in output: '%v'", output)
		}
	}
}
