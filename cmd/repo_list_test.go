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
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

var repoListTests = []struct {
	configFile       string // input
	expectedNumRepos int    // number of expected repositories to list
}{
	{"testdata/empty_repository_config/config.yaml", 0},
	{"testdata/default_repository_config/config.yaml", 1},
	{"testdata/multiple_repository_config/config.yaml", 2},
}

func TestRepoList(t *testing.T) {
	for _, tt := range repoListTests {
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.configFile, func(t *testing.T) {
			args := []string{"repo", "list"}
			if tt.configFile != "" {
				args = append(args, "--config", tt.configFile)
			}
			output, err := cmdtest.RunAppsodyCmd(args, ".", t)
			if err != nil {
				t.Fatal(err)
			}

			repos := cmdtest.ParseRepoList(output)
			if len(repos) != tt.expectedNumRepos {
				t.Errorf("Expected %d repos but found %d. CLI output:\n%s",
					tt.expectedNumRepos, len(repos), output)
			}
		})

	}
}

func TestRepoListJson(t *testing.T) {
	args := []string{"repo", "list", "--config", "testdata/multiple_repository_config/config.yaml", "-o", "json"}
	output, err := cmdtest.RunAppsodyCmd(args, ".", t)
	if err != nil {
		t.Fatal(err)
	}
	list, err := cmdtest.ParseRepoListJSON(cmdtest.ParseJSON(output))
	if err != nil {
		t.Fatal(err)
	}

	testContentsRepoListOutput(t, list, output)
}

func TestRepoListYaml(t *testing.T) {
	args := []string{"repo", "list", "--config", "testdata/multiple_repository_config/config.yaml", "-o", "yaml"}
	output, err := cmdtest.RunAppsodyCmd(args, ".", t)
	if err != nil {
		t.Fatal(err)
	}

	list, err := cmdtest.ParseRepoListYAML(cmdtest.ParseYAML(output))
	if err != nil {
		t.Fatal(err)
	}

	testContentsRepoListOutput(t, list, output)
}

func testContentsRepoListOutput(t *testing.T, list cmd.RepositoryFile, output string) {
	if list.APIVersion == "" {
		t.Errorf("Could not find APIVersion! CLI output:\n%s", output)
	}

	if list.Repositories == nil {
		t.Errorf("Could not find Repositories! CLI output:\n%s", output)
	}

	if len(list.Repositories) != 2 {
		t.Errorf("Expected 2 repos! CLI output:\n%s", output)
	}
}
