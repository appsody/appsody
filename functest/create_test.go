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
package functest

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestFuncCreateValidCases(t *testing.T) {
	var stackCreateValidTests = []struct {
		testName   string
		addToRepo  bool
		createArgs []string
		stackName  string
	}{
		{"Create with dev.local stack", false, []string{"stack", "create", "testing-stack", "--copy", "dev.local/starter"}, "testing-stack"},
		{"Create with custom repo stack", true, []string{"stack", "create", "testing-stack", "--copy", "test-repo/starter"}, "testing-stack"},
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

			sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

			packageArgs := []string{"stack", "package"}
			_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
			if err != nil {
				t.Fatal(err)
			}
			testStackName := tt.stackName

			if tt.addToRepo {

				devlocalFolder := filepath.Join(sandbox.ConfigDir, "stacks", "dev.local")

				addToRepoArgs := []string{"stack", "add-to-repo", "test-repo", "--release-url", "file://" + devlocalFolder + "/"}

				_, err = cmdtest.RunAppsody(sandbox, addToRepoArgs...)
				if err != nil {
					t.Fatal(err)
				}

				testRepoIndex := filepath.Join(devlocalFolder, "test-repo-index.yaml")

				addRepoArgs := []string{"repo", "add", "test-repo", "file://" + testRepoIndex}
				_, err = cmdtest.RunAppsody(sandbox, addRepoArgs...)
				if err != nil {
					t.Fatal(err)
				}

			}
			_, err = cmdtest.RunAppsody(sandbox, tt.createArgs...)
			if err != nil {
				t.Fatal(err)
			}

			exists, err := cmdtest.Exists(filepath.Join(sandbox.ProjectDir, testStackName))
			if !exists {
				t.Fatal(err)
			}
		})
	}
}

func TestFuncCreateInvalidCases(t *testing.T) {
	var stackCreateInvalidTests = []struct {
		testName       string
		addToRepo      bool
		removeSrc      bool
		createArgs     []string
		expectedOutput string
	}{
		{"Create with invalid stack", false, false, []string{"stack", "create", "testing-stack", "--copy", "dev.local/invalid"}, "Could not find stack specified in repository index"},
		{"Create with no src", false, true, []string{"stack", "create", "testing-stack", "--copy", "dev.local/starter"}, "No source URL specified"},
		{"Create with invalid url", true, false, []string{"stack", "create", "testing-stack", "--copy", "test-repo/starter"}, "Could not download file://invalidurl"},
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

			sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

			packageArgs := []string{"stack", "package"}
			_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
			if err != nil {
				t.Fatal(err)
			}

			if tt.addToRepo {

				devlocalFolder := filepath.Join(sandbox.ConfigDir, "stacks", "dev.local")

				addToRepoArgs := []string{"stack", "add-to-repo", "test-repo", "--release-url", "file://invalidurl/"}

				_, err = cmdtest.RunAppsody(sandbox, addToRepoArgs...)
				if err != nil {
					t.Fatal(err)
				}

				testRepoIndex := filepath.Join(devlocalFolder, "test-repo-index.yaml")

				addRepoArgs := []string{"repo", "add", "test-repo", "file://" + testRepoIndex}
				_, err = cmdtest.RunAppsody(sandbox, addRepoArgs...)
				if err != nil {
					t.Fatal(err)
				}

			}

			if tt.removeSrc {
				devlocalFile := filepath.Join(sandbox.ConfigDir, "stacks", "dev.local", "dev.local-index.yaml")

				file, err := ioutil.ReadFile(devlocalFile)
				if err != nil {
					t.Fatal(err)
				}

				lines := strings.Split(string(file), "\n")

				for i, line := range lines {
					if strings.Contains(line, "src:") {
						lines[i] = ""
					}
				}
				output := strings.Join(lines, "\n")
				err = ioutil.WriteFile(devlocalFile, []byte(output), 0644)

				if err != nil {
					t.Fatal(err)
				}
			}
			output, err := cmdtest.RunAppsody(sandbox, tt.createArgs...)
			if err != nil {
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("String" + tt.expectedOutput + " not found in output")
				}
			} else {
				t.Error("Stack create command unexpectedly passed")
			}
		})
	}
}
