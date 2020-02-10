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
	"github.com/appsody/appsody/cmd/cmdtest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var stackAddToRepoTests = []struct {
	testName       string
	args           []string // input
	repo           string   //name of repository to run appsody list on
	repoFile       string
	repoConfigured bool //if true, repo exists within the repository list
}{
	{"Simple test", []string{"dev.local"}, "dev.local", "", false},
	{"Test with repo URL pointing to remote repository", []string{"incubator"}, "incubator", "", false},
	{"Repository in configured list", []string{"myrepository"}, "myrepository", "dev.local-index.yaml", true},
	{"Repository not in configured list", []string{"myrepository", "--use-local-cache"}, "myrepository", "stacks/dev.local/myrepository-index.yaml", false},
}

func TestStackAddToRepo(t *testing.T) {
	for _, testData := range stackAddToRepoTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

			packageArgs := []string{"stack", "package"}
			_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
			if err != nil {
				t.Fatal(err)
			}

			if tt.repoConfigured == true {
				args := append([]string{"repo", "add", tt.repo}, "file://"+filepath.Join(sandbox.TestDataPath, tt.repoFile))
				_, err := cmdtest.RunAppsody(sandbox, args...)
				if err != nil {
					t.Fatalf("Error adding repo to configured list of repositories: %v", err)
				}
			}

			args := append([]string{"stack", "add-to-repo"}, tt.args...)
			output, err := cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatalf("Error adding stack to repository: %v", err)
			}

			if tt.repoFile != "" && tt.repoConfigured == false {
				args = append([]string{"repo", "add", tt.repo}, "file://"+filepath.Join(sandbox.ConfigDir, tt.repoFile))
				_, err := cmdtest.RunAppsody(sandbox, args...)
				if err != nil {
					t.Fatalf("Error adding repo to configured list of repositories: %v", err)
				}
			}
			output, err = cmdtest.RunAppsody(sandbox, "list", tt.repo)
			if !strings.Contains(output, "starter") {
				t.Errorf("Expected starter stack to be added to the %s repository.", tt.repo)
			}
			if err != nil {
				t.Fatalf("Expected test to pass without errors: %v", err)
			}
		})
	}
}

var stackAddToRepoErrorsTests = []struct {
	testName     string
	args         []string // input
	expectedLogs string   // logs that are expected in the output
	dir          string   //directory to remove
}{
	{"No args", nil, "You must specify a repository", ""},
	{"Too many args", []string{"too", "many", "arguments"}, "One argument expected.", ""},
	{"No templates directory", []string{"dev.local"}, "Current directory must be the root of the stack", "templates"},
	{"Without packaging stack", []string{"dev.local"}, "Run appsody stack package on your stack before running this command.", ""},
}

func TestStackAddToRepoErrors(t *testing.T) {

	for _, testData := range stackAddToRepoErrorsTests {
		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

			if tt.dir != "" {
				err := os.RemoveAll(filepath.Join(sandbox.ProjectDir, tt.dir))
				if err != nil {
					t.Fatalf("Error removing directory: %v", err)
				}
			}

			args := append([]string{"stack", "add-to-repo"}, tt.args...)
			output, err := cmdtest.RunAppsody(sandbox, args...)
			if err == nil {
				t.Error("Expected non-zero exit code.")
			}
			if !strings.Contains(output, tt.expectedLogs) {
				t.Errorf("Did not find expected error in output: %s", tt.expectedLogs)
			}
		})
	}
}

var stackAddToRepoStackExists = []struct {
	testName string
	args     []string // input
	repoName string   // logs that are expected in the output
}{
	{"Test without use local cache", []string{"dev.local"}, "dev.local"},
	{"Test with use local cache", []string{"dev.local", "--use-local-cache"}, "dev.local"},
	{"Test with repo URL pointing to remote repository and use local cache", []string{"incubator", "--use-local-cache"}, "incubator"},
}

func TestStackAddToRepoUseLocalCache(t *testing.T) {
	for _, testData := range stackAddToRepoStackExists {
		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

			packageArgs := []string{"stack", "package"}
			_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
			if err != nil {
				t.Fatal(err)
			}

			args := append([]string{"stack", "add-to-repo"}, tt.args...)
			_, err = cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatalf("Error adding stack to repository: %v", err)
			}
			output, err := cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatalf("Error adding stack to repository: %v", err)
			}

			output, err = cmdtest.RunAppsody(sandbox, "list", tt.repoName)
			if !strings.Contains(output, "starter") {
				t.Errorf("Expected starter stack to be added to the %s repository.", tt.repoName)
			}
			if err != nil {
				t.Fatalf("Expected test to pass without errors: %v", err)
			}
		})
	}
}
