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

func TestStackCreateValidCases(t *testing.T) {
	var stackCreateValidTests = []struct {
		testName  string
		args      []string
		stackName string
	}{
		//{"No args", []string{}, "test-stack-no-args"},
		{"Existing default repo", []string{"--copy", "incubator/nodejs"}, "test-stack-existing-repo"},
	}
	for _, testData := range stackCreateValidTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			testStackName := tt.stackName
			args := []string{"stack", "create", testStackName, "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml")}
			args = append(args, tt.args...)
			_, err := cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatal("Unexpected error when running appsody stack create: ", err)
			}
			exists, err := cmdtest.Exists(filepath.Join(sandbox.ProjectDir, testStackName))
			if err != nil {
				t.Fatal("Failed to check if the stack exists: ", err)
			}
			if !exists {
				t.Fatal("Stack doesn't exist despite appsody stack create executing correctly")
			}
		})
	}
}
func TestStackCreateInvalidCases(t *testing.T) {
	var stackCreateInvalidTests = []struct {
		testName     string
		stackName    string
		args         []string // input
		expectedLogs string   // logs that are expected in the output
	}{
		{"Invalid args", "testing-stack-create-invalid-args", []string{"--copy", "incubator/nodej"}, "Could not find stack specified in repository index"},
		{"Non-existing repo", "testing-stack-create-non-existing-repo", []string{"--copy", "experimental/nodejs"}, "Could not find stack specified in repository index"},
		{"Invalid repo", "testing-stack-create-invalid-repo", []string{"--copy", "invalid/java-microprofile"}, "Repository: 'invalid' was not found in the repository.yaml file"},
		{"Invalid stack name underscores", "testing_stack_invalid_name", nil, "The name must start with a lowercase letter, contain only lowercase letters, numbers, or dashes, and cannot end in a dash."},
		{"Invalid stack name length", "testing_stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stack", nil, "The name must be 68 characters or less"},
		{"Invalid stack name missing", "", nil, "Invalid project-name. The name cannot be an empty string"},
	}
	for _, testData := range stackCreateInvalidTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()
			args := append([]string{"stack", "create"}, tt.stackName)
			if tt.args != nil {
				args = append(args, tt.args...)
			}
			output, err := cmdtest.RunAppsody(sandbox, args...)
			if err == nil {
				t.Fatalf("Expected non-zero exit code: %v", tt.expectedLogs)
			}
			if !strings.Contains(output, tt.expectedLogs) {
				t.Errorf("Did not find expected error '%s' in output", tt.expectedLogs)
			}
			_, err = cmdtest.Exists(tt.stackName)
			if err != nil {
				t.Fatal("Error attempting to check stack exists: ", err)
			}
		})
	}
}

//func TestStackAlreadyExists(t *testing.T) {
//	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
//	defer cleanup()
//	testStackName := "test-stack-already-exists"
//	expectedLog := "A stack named " + testStackName + " already exists in your directory. Specify a unique stack name"
//	args := []string{"stack", "create", testStackName, "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml")}
//	_, err := cmdtest.RunAppsody(sandbox, args...)
//	if err != nil {
//		t.Fatal(err)
//	}
//	exists, err := cmdtest.Exists(filepath.Join(sandbox.ProjectDir, testStackName))
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !exists {
//		t.Fatal(err)
//	}
//	_, err = cmdtest.RunAppsody(sandbox, args...)
//	if !strings.Contains(err.Error(), expectedLog) {
//		t.Error("String \"" + expectedLog + "\" not found in output")
//	} else {
//		if err == nil {
//			t.Fatalf("Expected non-zero exit code: %v", expectedLog)
//		}
//	}
//}
func TestStackCreateSampleStackDryrun(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()
	args := []string{"stack", "create", "testing-stack", "--dryrun", "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml")}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Errorf("Error running dry run mode: %v", err)
	} else {
		if !strings.Contains(output, "Dry run complete") {
			t.Error("String \"Dry run complete\" not found in output")
		}
	}
}
