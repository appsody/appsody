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

	"github.com/appsody/appsody/cmd/cmdtest"
)

var repoRemoveLogsTests = []struct {
	testName     string
	args         []string // input
	expectedLogs string   // expected to be in the error message
}{
	{"No args", nil, "you must specify repository name"},
	{"Existing default repo", []string{"incubator"}, "cannot remove the default repository"},
	{"Non-existing repo", []string{"test"}, "not in configured list of repositories"},
	{"Badly formatted repo config", []string{"test", "--config", "testdata/bad_format_repository_config/config.yaml"}, "Failed to parse repository file yaml"},
}

func TestRepoRemoveLogs(t *testing.T) {

	for _, testData := range repoRemoveLogsTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			// see how many repos we currently have
			startRepos := getRepoListOutput(t, sandbox)

			args := append([]string{"repo", "remove"}, tt.args...)
			output, err := cmdtest.RunAppsody(sandbox, args...)
			if err == nil {
				t.Fatalf("Expected non-zero exit code: %v", tt.expectedLogs)
			}

			if !strings.Contains(output, tt.expectedLogs) {
				t.Errorf("Did not find expected error '%s' in output", tt.expectedLogs)
			}
			// see how many repos we have after running repo add
			endRepos := getRepoListOutput(t, sandbox)
			if len(startRepos) != len(endRepos) {
				t.Errorf("Expected %d repos but found %d", len(startRepos), len(endRepos))
			}
		})
	}
}

func TestRepoRemove(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"repo", "remove", "experimental"}

	// see how many repos we currently have
	startRepos := getRepoListOutput(t, sandbox)

	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(output, "repository has been removed") {
		t.Error("Expected output to contain text 'repository has been removed', but it didn't: ", output)
	}

	// see how many repos we have after running repo add
	endRepos := getRepoListOutput(t, sandbox)
	if (len(startRepos) - 1) != len(endRepos) {
		t.Errorf("Expected %d repos but found %d", len(startRepos), len(endRepos))
	}
}

func TestRepoRemoveDryRun(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// see how many repos we currently have
	startRepos := getRepoListOutput(t, sandbox)

	args := []string{"repo", "remove", "localhub", "--config", "testdata/multiple_repository_config/config.yaml", "--dryrun"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(output, "Dry Run - Skip") {
		t.Error("Did not find expected error 'Dry Run - Skip' in output")
	}
	// see how many repos we have after running repo add
	endRepos := getRepoListOutput(t, sandbox)
	if len(startRepos) != len(endRepos) {
		t.Errorf("Expected %d repos but found %d", len(startRepos), len(endRepos))
	}
}
