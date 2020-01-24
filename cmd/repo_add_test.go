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

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestRepoAdd(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// see how many repos we currently have
	startRepos, _ := getRepoListOutput(t, sandbox)

	addRepoName := "LocalTestRepo"
	addRepoURL, err := cmdtest.AddLocalRepo(sandbox, addRepoName, filepath.Join(sandbox.TestDataPath, "index.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// see how many repos we have after running repo add
	endRepos, _ := getRepoListOutput(t, sandbox)

	if (len(startRepos) + 1) != len(endRepos) {
		t.Errorf("Expected %d repos but found %d", (len(startRepos) + 1), len(endRepos))
	} else {
		// check that the correct repo name and url were added
		found := false
		for _, repo := range endRepos {
			if repo.Name == addRepoName && repo.URL == addRepoURL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected repo with name '%s' and url '%s'", addRepoName, addRepoURL)
		}
	}
}

func TestRepoAddDryRun(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// see how many repos we currently have
	startRepos, _ := getRepoListOutput(t, sandbox)

	args := []string{"repo", "add", "experimental", "https://github.com/appsody/stacks/releases/latest/download/experimental-index.yaml", "--dryrun", "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml")}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(output, "Dry Run - Skip") {
		t.Error("Did not find expected error 'Dry run - Skip' in output")
	}
	// see how many repos we have after running repo add
	endRepos, _ := getRepoListOutput(t, sandbox)

	if len(startRepos) != len(endRepos) {
		t.Errorf("Expected %d repos but found %d", len(startRepos), len(endRepos))
	}
}

func TestRepoAddErrors(t *testing.T) {

	var repoAddErrorTests = []struct {
		testName      string
		args          []string // input
		configDir     string
		expectedError string // expected to be in the error message
	}{
		{"No args", nil, "", "You must specify a repository name and URL"},
		{"One arg", []string{"reponame"}, "", "You must specify a repository name and URL"},
		{"No url scheme", []string{"test", "localhost"}, "", "unsupported protocol scheme"},
		{"Non-existing url", []string{"test", "http://localhost/doesnotexist"}, "", "refused"},
		{"Repo name over 50 characters", []string{"reponametoolongtestreponametoolongtestreponametoolongtest", "http://localhost/doesnotexist"}, "", "must be less than 50 characters"},
		{"Repo name is invalid", []string{"test!", "http://localhost/doesnotexist"}, "", "Invalid repository name"},
		{"Repo name already exists", []string{"incubator", "http://localhost/doesnotexist"}, "", "already exists"},
		{"Url already exists", []string{"test", "https://github.com/appsody/stacks/releases/latest/download/incubator-index.yaml"}, "", "already exists"},
		{"Badly formatted repo config", []string{"test", "http://localhost/doesnotexist"}, "bad_format_repository_config", "Failed to parse repository file yaml"},
		{"Too many arguments", []string{"too", "many", "arguments"}, "", "Two arguments expected."},
	}

	for _, testData := range repoAddErrorTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {

			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			sandbox.SetConfigInTestData(tt.configDir)

			args := append([]string{"repo", "add"}, tt.args...)
			output, err := cmdtest.RunAppsody(sandbox, args...)

			if err == nil {
				t.Error("Expected non-zero exit code.")
			}

			if !strings.Contains(output, tt.expectedError) {
				t.Errorf("Did not find expected error '%s' in output", tt.expectedError)
			}
		})
	}
}

func getRepoListOutput(t *testing.T, sandbox *cmdtest.TestSandbox) ([]cmdtest.Repository, error) {
	output, err := cmdtest.RunAppsody(sandbox, "repo", "list")
	if err != nil {
		return nil, err
	}
	startRepos := cmdtest.ParseRepoList(output)
	return startRepos, nil
}
