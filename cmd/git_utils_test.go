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
	"bytes"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestStringBefore(t *testing.T) {
	var StringBeforeTests = []struct {
		testName       string
		value          string
		searchString   string
		expectedOutput string
	}{
		{"Empty values", "", "", ""},
		{"Non-empty values", "teststring", "string", "test"},
	}
	for _, testData := range StringBeforeTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			output := cmd.StringBefore(tt.value, tt.searchString)
			if output != tt.expectedOutput {
				t.Errorf("Did not get expected string output: %s", tt.expectedOutput)
			}
		})
	}
}

func TestStringAfter(t *testing.T) {
	var StringAfterTests = []struct {
		testName       string
		value          string
		searchString   string
		expectedOutput string
	}{
		{"Empty values", "", "", ""},
		{"Non-empty values", "teststring", "test", "string"},
		{"Search value is not present in string", "teststring", "stringtest", ""},
	}
	for _, testData := range StringAfterTests {
		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			output := cmd.StringAfter(tt.value, tt.searchString)
			if output != tt.expectedOutput {
				t.Errorf("Did not get expected string output: %s", tt.expectedOutput)
			}
		})
	}
}

func TestStringBetween(t *testing.T) {
	var StringBetweenTests = []struct {
		testName       string
		value          string
		preString      string
		postString     string
		expectedOutput string
	}{
		{"Empty values", "", "", "", ""},
		{"Non-empty values", "teststring", "te", "ing", "ststr"},
		{"Search value preString is not present in string", "teststring", "blah", "string", ""},
		{"Search value postString is not present in string", "teststring", "test", "blah", ""},
	}
	for _, testData := range StringBetweenTests {
		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			output := cmd.StringBetween(tt.value, tt.preString, tt.postString)
			if output != tt.expectedOutput {
				t.Errorf("Did not get expected string output: %s", tt.expectedOutput)
			}
		})
	}
}

func TestGetGitInfoWithNotAGitRepo(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)
	config := &cmd.RootCommandConfig{LoggingConfig: loggingConfig}

	// Change the config ProjectDir to be in the sandboxing folder because that's where
	// we want to execute the commands
	config.ProjectDir = sandbox.ProjectDir

	_, err := cmd.GetGitInfo(config)
	expectedError := "not a git repository"
	expectedErrorCapitalLetter := "Not a git repository"
	if err != nil {
		if !strings.Contains(err.Error(), expectedError) || !strings.Contains(err.Error(), expectedErrorCapitalLetter) {
			t.Errorf("String \"not a git repository\" not found in output: %v", err)
		}
	} else {
		t.Errorf("Test unexpectedly passed")
	}
}

func TestGitInfo(t *testing.T) {
	var gitInfoValues = []struct {
		testName    string
		expectedLog string
	}{
		{"TestGetGitInfoWithNoCommits", "does not have any commits yet"},
		{"TestGetGitInfoWithRemote", "Successfully retrieved remote name"},
	}
	for _, testData := range gitInfoValues {
		tt := testData
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			var outBuffer bytes.Buffer
			loggingConfig := &cmd.LoggingConfig{}
			loggingConfig.InitLogging(&outBuffer, &outBuffer)
			config := &cmd.RootCommandConfig{LoggingConfig: loggingConfig}

			// Change the config ProjectDir to be in the sandboxing folder because that's where
			// we want to execute the commands
			config.ProjectDir = sandbox.ProjectDir

			_, gitErr := cmd.RunGit(loggingConfig, sandbox.ProjectDir, []string{"init"}, false)
			if gitErr != nil {
				t.Error(gitErr)
			}

			if tt.testName == "TestGetGitInfoWithRemote" {
				_, gitErr := cmd.RunGit(loggingConfig, sandbox.ProjectDir, []string{"remote", "add", "testRemote", "testURL"}, false)
				if gitErr != nil {
					t.Error(gitErr)
				}
			}

			output, err := cmd.GetGitInfo(config)

			if tt.testName == "TestGetGitInfoWithRemote" {
				if output.Upstream != "testRemote" {
					t.Errorf("Did not return expected upstream: testRemote, instead got: %v", output.Upstream)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedLog) {
					t.Errorf("Should have flagged error: %v. Got: %v", tt.expectedLog, err)
				}
			}

		})
	}
}
