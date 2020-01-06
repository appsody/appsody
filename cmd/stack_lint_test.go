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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestAppsodyRunMissingInDockerfileStack(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	restoreLine := ""
	file, err := ioutil.ReadFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"))
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(file), "\n")

	for i, line := range lines {
		if strings.Contains(line, "APPSODY_RUN") {
			restoreLine = lines[i]
			lines[i] = "Testing"
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}

	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "Testing") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAppsodyMountsMissingInDockerfileStack(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	restoreLine := ""
	file, err := ioutil.ReadFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"))
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(file), "\n")

	for i, line := range lines {
		if strings.Contains(line, "APPSODY_MOUNTS") {
			restoreLine = lines[i]
			lines[i] = "Testing"
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}

	_, err = cmdtest.RunAppsody(sandbox, args...)
	if !strings.Contains(err.Error(), "LINT TEST FAILED") {
		t.Error("String \"LINT TEST FAILED\" not found in output")
	} else {
		if err == nil {
			t.Error("Expected error but did not receive one.")
		}
	}
	for i, line := range lines {
		if strings.Contains(line, "Testing") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAppsodyWatchDirPresentAndOnChangeMissingInDockerfileStack(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	restoreLine := ""
	file, err := ioutil.ReadFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"))

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(file), "\n")

	if strings.Contains(string(file), "APPSODY_WATCH_DIR") {
		if strings.Contains(string(file), "_ON_CHANGE") {
			for i, line := range lines {
				if strings.Contains(line, "_ON_CHANGE") {
					restoreLine = lines[i]
					lines[i] = "Testing"
				}
			}

			output := strings.Join(lines, "\n")
			err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

			if err != nil {
				t.Fatal(err)
			}

			args := []string{"stack", "lint"}

			_, err = cmdtest.RunAppsody(sandbox, args...)
			if err == nil { //Lint check should fail, if not fail the test
				t.Fatal(err)
			}

			for i, line := range lines {
				if strings.Contains(line, "Testing") {
					lines[i] = restoreLine
				}
			}

			output = strings.Join(lines, "\n")
			err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

			if err != nil {
				t.Fatal(err)
			}

		} else {
			args := []string{"stack", "lint"}
			_, err = cmdtest.RunAppsody(sandbox, args...)

			if err == nil { //Lint check should fail, if not fail the test
				t.Fatal(err)
			}
		}
	}
}

func Test_KillValue(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	restoreLine := ""
	file, err := ioutil.ReadFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"))

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(file), "\n")

	for i, line := range lines {
		if strings.Contains(line, "_KILL") {
			restoreLine = lines[i]
			lines[i] = "ENV APPSODY_DEBUG_KILL=trued"
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}

	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "ENV APPSODY_DEBUG_KILL=trued") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAppsodyRegexValue(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	restoreLine := ""
	file, err := ioutil.ReadFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"))

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(file), "\n")

	for i, line := range lines {
		if strings.Contains(line, "ENV APPSODY_WATCH_REGEX='^.*(.xml|.java|.properties)$'") {
			restoreLine = lines[i]
			lines[i] = "ENV APPSODY_WATCH_REGEX='['"
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}

	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "ENV APPSODY_WATCH_REGEX='['") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile(filepath.Join(cmdtest.TestDirPath, "test-stack", "image", "Dockerfile-stack"), []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}
func TestLintWithValidStack(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"stack", "lint", filepath.Join(cmdtest.TestDirPath, "test-stack")}

	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}
}

func TestLintWithMissingStackYaml(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}
	removeYaml := filepath.Join(testStackPath, "stack.yaml")
	file, err := ioutil.ReadFile(removeYaml)
	if err != nil {
		t.Fatal(err)
	}
	removeArray := []string{filepath.Join(removeYaml)}

	osErr := os.RemoveAll(removeYaml)
	if osErr != nil {
		t.Fatal(osErr)
	}

	_, appsodyErr := cmdtest.RunAppsody(sandbox, args...)

	if appsodyErr == nil { //Lint check should fail, if not fail the test
		t.Fatal("Expected failure - Missing stack.yaml")
	}

	defer RestoreSampleStack(removeArray, testStackPath, file)
}

func TestLintWithMissingImageProjectAndConfigDir(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint"}
	removeImage := filepath.Join(testStackPath, "image")
	file, readErr := ioutil.ReadFile(filepath.Join(removeImage, "Dockerfile-stack"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	removeArray := []string{removeImage, filepath.Join(removeImage, "config"), filepath.Join(removeImage, "project"), filepath.Join(removeImage, "config", "app-deploy.yaml"), filepath.Join(removeImage, "project", "Dockerfile"), filepath.Join(removeImage, "LICENSE"), filepath.Join(removeImage, "Dockerfile-stack")}

	osErr := os.RemoveAll(removeImage)
	if osErr != nil {
		t.Fatal(osErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	defer RestoreSampleStack(removeArray, testStackPath, file)

	if !strings.Contains(output, "Missing image directory") {
		t.Fatal(err, ": Expected failure - Missing image directory")
	}
}

func TestLintWithMissingREADME(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}
	removeReadme := filepath.Join(testStackPath, "README.md")
	removeArray := []string{removeReadme}

	osErr := os.RemoveAll(removeReadme)
	if osErr != nil {
		t.Fatal(osErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	var b []byte
	defer RestoreSampleStack(removeArray, testStackPath, b)

	if !strings.Contains(output, "Missing README.md") {
		t.Fatal(err, ": Expected failure - Missing README")
	}
}

func TestLintWithMissingTemplatesDirectory(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}

	removeTemplatesDir := filepath.Join(testStackPath, "templates")
	removeArray := []string{removeTemplatesDir, filepath.Join(removeTemplatesDir, "default"), filepath.Join(removeTemplatesDir, "default", "app.js")}

	osErr := os.RemoveAll(removeTemplatesDir)
	if osErr != nil {
		t.Fatal(osErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	var b []byte
	RestoreSampleStack(removeArray, testStackPath, b)

	if !strings.Contains(output, "Missing template directory") && !strings.Contains(output, "No templates found in") {
		t.Fatal(err, ": Expected failure - Missing templates directory")
	}
}

func TestLintWithInvalidVersion(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}

	stackYaml := filepath.Join(testStackPath, "stack.yaml")

	file, readErr := ioutil.ReadFile(stackYaml)
	if readErr != nil {
		t.Fatal(readErr)
	}

	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "version: ") {
			lines[i] = "version: invalidVersion"
		}
	}

	invalidYaml := strings.Join(lines, "\n")
	writeErr := ioutil.WriteFile(stackYaml, []byte(invalidYaml), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "Version must be formatted in accordance to semver") {
		t.Fatal(err, ": Expected failure - Version in stack.yaml not formatted in accordance to semver")
	}

}

func TestLintWithLongNameAndDescription(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}

	stackYaml := filepath.Join(testStackPath, "stack.yaml")

	file, readErr := ioutil.ReadFile(stackYaml)
	if readErr != nil {
		t.Fatal(readErr)
	}

	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "name: ") {
			lines[i] = "name: This stack name is far too long and therefore should fail"
		}
		if strings.HasPrefix(line, "description: ") {
			lines[i] = "description: This stack description is far too long (greater than 70 characters) and therefore should also fail."
		}
	}

	invalidYaml := strings.Join(lines, "\n")
	writeErr := ioutil.WriteFile(stackYaml, []byte(invalidYaml), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "Description must be under ") && !strings.Contains(output, "Stack name must be under ") {
		t.Fatal(err, ": Expected failure - Stack name and description in stack.yaml is too long")
	}
}

func TestLintWithInvalidLicenseField(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}

	stackYaml := filepath.Join(testStackPath, "stack.yaml")

	file, readErr := ioutil.ReadFile(stackYaml)
	if readErr != nil {
		t.Fatal(readErr)
	}

	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "license: ") {
			lines[i] = "license: invalidLicense"
		}
	}

	invalidYaml := strings.Join(lines, "\n")
	writeErr := ioutil.WriteFile(stackYaml, []byte(invalidYaml), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "The stack.yaml SPDX license ID is invalid") {
		t.Fatal(err, ": Expected failure - License value in stack.yaml is invalid")
	}
}

func TestLintWithInvalidTemplatingValues(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}

	stackYaml := filepath.Join(testStackPath, "stack.yaml")

	file, readErr := ioutil.ReadFile(stackYaml)
	if readErr != nil {
		t.Fatal(readErr)
	}

	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "  key1: ") {
			lines[i] = "  key&@_1: value"
		}
	}

	invalidYaml := strings.Join(lines, "\n")
	writeErr := ioutil.WriteFile(stackYaml, []byte(invalidYaml), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "is not in an alphanumeric format") {
		t.Fatal(err, ": Expected failure - Templating data in stack.yaml is not in alphanumeric format.")
	}
}

func TestLintWithInvalidRequirements(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackPath := filepath.Join(cmdtest.TestDirPath, "test-stack")
	args := []string{"stack", "lint", testStackPath}

	stackYaml := filepath.Join(testStackPath, "stack.yaml")

	file, readErr := ioutil.ReadFile(stackYaml)
	if readErr != nil {
		t.Fatal(readErr)
	}

	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "  appsody-version:") {
			lines[i] = "  appsody-version: invalid-req"
		}
	}

	invalidYaml := strings.Join(lines, "\n")
	writeErr := ioutil.WriteFile(stackYaml, []byte(invalidYaml), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "is not in the correct format. See:") {
		t.Fatal(err, ": Expected failure - Requirement constraint is not in the correct format.")
	}

}

func RestoreSampleStack(fixStack []string, testStackPath string, writeContents []byte) {
	for _, missingContent := range fixStack {
		if missingContent == filepath.Join(testStackPath, "image") || missingContent == filepath.Join(testStackPath, "image", "config") || missingContent == filepath.Join(testStackPath, "image/project") || missingContent == filepath.Join(testStackPath, "templates") || missingContent == filepath.Join(testStackPath, "templates", "default") {
			osErr := os.Mkdir(missingContent, os.ModePerm)
			if osErr != nil {
				fmt.Println(osErr)
			}
		} else {
			_, osErr := os.Create(missingContent)
			if osErr != nil {
				fmt.Println(osErr)
			}
			if missingContent == filepath.Join(testStackPath, "image", "Dockerfile-stack") || missingContent == filepath.Join(testStackPath, "stack.yaml") {
				writeErr := ioutil.WriteFile(missingContent, []byte(writeContents), 0644)
				if writeErr != nil {
					fmt.Println(writeErr)
				}
			}
		}
	}
}
