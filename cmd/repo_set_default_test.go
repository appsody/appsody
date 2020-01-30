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

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestRepoSetDefaultLogs(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var repoSetDefaultLogsTests = []struct {
		testName     string
		args         []string // input
		configDir    string
		expectedLogs string // logs that are expected in the output
	}{
		{"No args", nil, "", "You must specify the repository to set as the default"},
		{"Existing default repo", []string{"incubator"}, "", "default repository has already been set to"},
		{"Non-existing repo", []string{"test"}, "", "not in your configured list of repositories"},
		{"Too many arguments", []string{"too", "many", "arguments"}, "", "One argument expected."},
		{"Badly formatted repo config", []string{"test"}, "bad_format_repository_config", "Failed to parse repository file yaml"},
	}

	for _, testData := range repoSetDefaultLogsTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		sandbox.SetConfigInTestData(tt.configDir)

		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {

			args := append([]string{"repo", "set-default"}, tt.args...)
			output, err := cmdtest.RunAppsody(sandbox, args...)
			if err == nil {
				t.Error("Expected non-zero exit code.")
			}
			if !strings.Contains(output, tt.expectedLogs) {
				t.Errorf("Did not find expected error '%s' in output", tt.expectedLogs)
			}

			// check default repo is unchanged and is still incubator
			output, err = cmdtest.RunAppsody(sandbox, "repo", "list")
			if err != nil {
				if !strings.Contains(err.Error(), tt.expectedLogs) {
					t.Fatalf("Unexpected error when running repo list: %s ", err)
				}
			} else {
				checkExpectedDefaultRepo(t, output, "*incubator")
			}
		})
	}
}

func TestRepoSetDefault(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"repo", "set-default", "localhub", "--config", filepath.Join(sandbox.TestDataPath, "multiple_repository_config", "config.yaml")}
	removeRepo := filepath.Join(sandbox.TestDataPath, "multiple_repository_config", "repository", "repository.yaml")
	file, readErr := ioutil.ReadFile(removeRepo)
	if readErr != nil {
		t.Fatal(readErr)
	}

	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	// check default repo has been changed to localhub
	output, err := cmdtest.RunAppsody(sandbox, "repo", "list", "--config", filepath.Join(sandbox.TestDataPath, "multiple_repository_config", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	checkExpectedDefaultRepo(t, output, "*localhub")
	writeErr := ioutil.WriteFile(filepath.Join(removeRepo), []byte(file), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
}

func TestRepoSetDefaultDryRun(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"repo", "set-default", "localhub", "--config", filepath.Join(sandbox.TestDataPath, "multiple_repository_config", "config.yaml"), "--dryrun"}

	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	// check default repo is unchanged and is still incubator
	output, err := cmdtest.RunAppsody(sandbox, "repo", "list")
	if err != nil {
		t.Fatal(err)
	}
	checkExpectedDefaultRepo(t, output, "*incubator")
}

func checkExpectedDefaultRepo(t *testing.T, output string, checkRepo string) {
	if !strings.Contains(output, checkRepo) {
		t.Errorf("Expected default repo to be %v", checkRepo)
	}
}
