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
	"path/filepath"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

// test the v2 list functionality
func TestListV2(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// first add the test repo index
	_, err := cmdtest.AddLocalRepo(sandbox, "incubatortest", filepath.Join(sandbox.TestDataPath, "kabanero.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	output, _ := cmdtest.RunAppsody(sandbox, "list", "incubatortest")
	if !(strings.Contains(output, "nodejs") && strings.Contains(output, "incubatortest")) {
		t.Error("list command should contain id 'nodejs'")
	}

	// test the current default repo
	output, _ = cmdtest.RunAppsody(sandbox, "list", "incubator")
	if !strings.Contains(output, "java-microprofile") {
		t.Error("list command should contain id 'java-microprofile'")
	}

	output, _ = cmdtest.RunAppsody(sandbox, "list")
	// we expect 2 instances
	if !(strings.Contains(output, "java-microprofile") && (strings.Count(output, "nodejs ") == 2)) {
		t.Error("list command should contain id 'java-microprofile and 2 nodejs '")
	}

	// test the current default repo
	output, _ = cmdtest.RunAppsody(sandbox, "list", "nonexisting")
	if !(strings.Contains(output, "cannot locate repository ")) {
		t.Error("Failed to flag non-existing repo")
	}

}

func TestListJson(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"list", "-o", "json"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	list, err := cmdtest.ParseListJSON(cmdtest.ParseJSON(output))

	if err != nil {
		t.Fatal(err)
	}

	testContentsListOutput(t, list, output)
}

func TestListYaml(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"list", "-o", "yaml"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Parsing yaml output: \n", output)
	list, err := cmdtest.ParseListYAML(cmdtest.ParseYAML(output))

	if err != nil {
		t.Fatal(err)
	}

	testContentsListOutput(t, list, output)
}

func TestListYamlSingleRepository(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"list", "incubator", "-o", "yaml"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	list, err := cmdtest.ParseListYAML(cmdtest.ParseYAML(output))
	if err != nil {
		t.Fatal(err)
	}

	if len(list.Repositories) != 1 && list.Repositories[0].Name == "incubator" {
		t.Error("Could not find repository 'incubator'")
	}
}

func TestListJsonSingleRepository(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"list", "incubator", "-o", "json"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	list, err := cmdtest.ParseListJSON(cmdtest.ParseJSON(output))
	if err != nil {
		t.Fatal(err)
	}

	if len(list.Repositories) != 1 && list.Repositories[0].Name == "incubator" {
		t.Error("Could not find repository 'incubator'")
	}
}

func testContentsListOutput(t *testing.T, list cmd.IndexOutputFormat, output string) {
	if list.APIVersion == "" {
		t.Error("Could not find 'APIVersion'")
	}

	if len(list.Repositories) != 2 {
		t.Errorf("Expected 2 repositories to be defined, but found %d", len(list.Repositories))
	}

	for _, repo := range list.Repositories {
		if len(repo.Stacks) < 1 {
			t.Errorf("Repository '%s' does not contain any stacks", repo.Name)
		}

		for _, stack := range repo.Stacks {
			if stack.ID == "" {
				t.Errorf("A stack in repo '%s' has no 'ID'", repo.Name)
			}
			if stack.Version == "" {
				t.Errorf("Stack '%s' in repo '%s' has no 'Version'", stack.ID, repo.Name)
			}
			if stack.Description == "" {
				t.Errorf("Stack '%s' in repo '%s' has no 'Description'", stack.ID, repo.Name)
			}
			if len(stack.Templates) == 0 {
				t.Errorf("Stack '%s' in repo '%s' has no 'Templates'", stack.ID, repo.Name)
			}
		}
	}
}
