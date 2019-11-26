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
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestList(t *testing.T) {
	// tests that would have run before this and crashed could leave the repo
	// in a bad state - mostly leading to: "a repo with this name already exists."
	// so clean it up pro-actively, ignore any errors.
	_, _ = cmdtest.RunAppsodyCmd([]string{"repo", "remove", "LocalTestRepo"}, ".", t)

	// first add the test repo index
	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml", t)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	output, err := cmdtest.RunAppsodyCmd([]string{"list"}, ".", t)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output, "A Java Microprofile Stack") {
		t.Error("list command should not display the stack name")
	}

	if !strings.Contains(output, "java-microprofile") {
		t.Error("list command should contain id 'java-microprofile'")
	}
}

// test the v2 list functionality
func TestListV2(t *testing.T) {
	// first add the test repo index
	var err error
	var output string
	var cleanup func()
	_, _ = cmdtest.RunAppsodyCmd([]string{"repo", "remove", "LocalTestRepo"}, ".", t)
	_, _ = cmdtest.RunAppsodyCmd([]string{"repo", "remove", "incubatortest"}, ".", t)
	_, cleanup, err = cmdtest.AddLocalFileRepo("incubatortest", "../cmd/testdata/kabanero.yaml", t)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	output, _ = cmdtest.RunAppsodyCmd([]string{"list", "incubatortest"}, ".", t)

	if !(strings.Contains(output, "nodejs") && strings.Contains(output, "incubatortest")) {
		t.Error("list command should contain id 'nodejs'")
	}

	// test the current default repo
	output, _ = cmdtest.RunAppsodyCmd([]string{"list", "incubator"}, ".", t)

	if !strings.Contains(output, "java-microprofile") {
		t.Error("list command should contain id 'java-microprofile'")
	}

	output, _ = cmdtest.RunAppsodyCmd([]string{"list"}, ".", t)

	// we expect 2 instances
	if !(strings.Contains(output, "java-microprofile") && (strings.Count(output, "nodejs ") == 2)) {
		t.Error("list command should contain id 'java-microprofile and 2 nodejs '")
	}

	// test the current default repo
	output, _ = cmdtest.RunAppsodyCmd([]string{"list", "nonexisting"}, ".", t)

	if !(strings.Contains(output, "cannot locate repository ")) {
		t.Error("Failed to flag non-existing repo")
	}

}

func TestListJson(t *testing.T) {
	args := []string{"list", "-o", "json"}
	output, err := cmdtest.RunAppsodyCmd(args, ".", t)

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
	args := []string{"list", "-o", "yaml"}
	output, err := cmdtest.RunAppsodyCmd(args, ".", t)

	if err != nil {
		t.Fatal(err)
	}

	list, err := cmdtest.ParseListYAML(cmdtest.ParseYAML(output))

	if err != nil {
		t.Fatal(err)
	}

	testContentsListOutput(t, list, output)
}

func TestListYamlSingleRepository(t *testing.T) {
	args := []string{"list", "incubator", "-o", "yaml"}
	output, err := cmdtest.RunAppsodyCmd(args, ".", t)

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
