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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestLintWithValidStack(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()
	args := []string{"stack", "lint", filepath.Join(sandbox.TestDataPath, "test-stack")}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}
}
func TestLinterInvalidValues(t *testing.T) {
	var linterInvalidValues = []struct {
		testName     string
		targetPath   string
		containsLine string
		replaceLine  string
		expectedLog  string
	}{
		{"Invalid Run Value", filepath.Join("image", "Dockerfile-stack"), "APPSODY_RUN", "Testing", "Missing APPSODY_RUN"},
		{"Invalid Kill Value", filepath.Join("image", "Dockerfile-stack"), "_KILL", "ENV APPSODY_DEBUG_KILL=trued", "APPSODY_DEBUG_KILL can only have value true/false"},
		{"Invalid Regex Value", filepath.Join("image", "Dockerfile-stack"), "ENV APPSODY_WATCH_REGEX='^.*(.xml|.java|.properties)$'", "ENV APPSODY_WATCH_REGEX='['", "error parsing regexp: missing closing ]"},
		{"Invalid Mount Seperator", filepath.Join("image", "Dockerfile-stack"), "ENV APPSODY_MOUNTS", "ENV APPSODY_MOUNTS=.,/project/user-app", "Mount is not properly formatted"},
		{"Invalid Mounts", filepath.Join("image", "Dockerfile-stack"), "ENV APPSODY_MOUNTS", "ENV APPSODY_MOUNTS=a:abcde", "Could not stat path"},
		{"Invalid Watch Dir", filepath.Join("image", "Dockerfile-stack"), "_ON_CHANGE", "Testing", "APPSODY_WATCH_DIR is defined, but no ON_CHANGE variable is defined"},
		{"Invalid Version", "stack.yaml", "version: ", "version: invalidVersion", "Version must be formatted in accordance to semver"},
		// Fails unmarshalling the file when searching for "name" - searching for "sample stack" instead
		{"Invalid Name Length", "stack.yaml", "sample stack", "name: This name is far too long to pass and therefore should also fail.", "Stack name must be under "},
		{"Invalid Description Length", "stack.yaml", "description: ", "description: This stack description is far too long (greater than 70 characters) and therefore should also fail.", "Description must be under "},
		{"Invalid License Field", "stack.yaml", "license: ", "license: invalidLicense", "The stack.yaml SPDX license ID is invalid"},
		{"Invalid Templating Value", "stack.yaml", "  key1: ", "  key&@_1: value", "is not in an alphanumeric format"},
		{"Invalid Requirements", "stack.yaml", "  appsody-version:", "  appsody-version: invalid-req", "is not in the correct format. See:"},
	}
	for _, testData := range linterInvalidValues {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()
			testStackPath := filepath.Join(sandbox.TestDataPath, "test-stack")
			targetPath := filepath.Join(testStackPath, tt.targetPath)
			file, err := ioutil.ReadFile(targetPath)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(file), "\n")
			for i, line := range lines {
				if strings.Contains(line, tt.containsLine) {
					lines[i] = tt.replaceLine
				}
			}
			output := strings.Join(lines, "\n")
			err = ioutil.WriteFile(targetPath, []byte(output), 0644)
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"stack", "lint", testStackPath}
			output, err = cmdtest.RunAppsody(sandbox, args...)
			if err == nil {
				t.Fatalf("Expected non-zero exit code: %v", tt.expectedLog)
			}
			if !strings.Contains(output, tt.expectedLog) {
				t.Fatalf("Expected failure to include - %s but instead received %s", tt.expectedLog, output)
			}
		})
	}
}
func TestLinterMissingValues(t *testing.T) {
	var linterMissingValues = []struct {
		testName    string
		removeFile  string
		expectedLog string
	}{
		{"Missing Stack Yaml", "stack.yaml", "stack.yaml: no such file or directory"},
		{"Missing Project And Config Dir", "image", "Missing image directory"},
		{"Missing README", "README.md", "Missing README.md"},
		{"Missing Templates Directory", "templates", "Missing template directory"},
	}
	for _, testData := range linterMissingValues {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()
			testStackPath := filepath.Join(sandbox.TestDataPath, "test-stack")
			removalTarget := filepath.Join(testStackPath, tt.removeFile)
			osErr := os.RemoveAll(removalTarget)
			if osErr != nil {
				t.Fatalf("Failed to remove %s. Error: %v", removalTarget, osErr)
			}
			args := []string{"stack", "lint", testStackPath}
			output, err := cmdtest.RunAppsody(sandbox, args...)
			if err == nil {
				t.Fatalf("Expected non-zero exit code: %s", tt.expectedLog)
			}
			if !strings.Contains(output, tt.expectedLog) {
				t.Fatalf("Expected failure to include - %s but instead received %s", tt.expectedLog, output)
			}
		})
	}
}
