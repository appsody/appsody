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

func TestAPPSODY_RUNMissingInDockerfileStack(t *testing.T) {
	restoreLine := ""
	file, err := ioutil.ReadFile("../cmd/testdata/test-stack/image/Dockerfile-stack")
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
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}
	_, err = cmdtest.RunAppsodyCmd(args, "../cmd/testData/test-stack", t)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "Testing") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAPPSODY_MOUNTSMissingInDockerfileStack(t *testing.T) {
	restoreLine := ""
	file, err := ioutil.ReadFile("../cmd/testdata/test-stack/image/Dockerfile-stack")
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
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}
	_, err = cmdtest.RunAppsodyCmd(args, "../cmd/testdata/test-stack", t)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "Testing") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAPPSODY_WATCH_DIRPRESENTAndONCHANGEMissingInDockerfileStack(t *testing.T) {
	restoreLine := ""
	file, err := ioutil.ReadFile("../cmd/testdata/test-stack/image/Dockerfile-stack")

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
			err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

			if err != nil {
				t.Fatal(err)
			}

			args := []string{"stack", "lint"}
			_, err = cmdtest.RunAppsodyCmd(args, "../cmd/testdata/test-stack", t)

			if err == nil { //Lint check should fail, if not fail the test
				t.Fatal(err)
			}

			for i, line := range lines {
				if strings.Contains(line, "Testing") {
					lines[i] = restoreLine
				}
			}

			output = strings.Join(lines, "\n")
			err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

			if err != nil {
				t.Fatal(err)
			}

		} else {
			args := []string{"stack", "lint"}
			_, err = cmdtest.RunAppsodyCmd(args, "../cmd/testdata/test-stack", t)

			if err == nil { //Lint check should fail, if not fail the test
				t.Fatal(err)
			}
		}
	}
}

func Test_KILLValue(t *testing.T) {
	restoreLine := ""
	file, err := ioutil.ReadFile("../cmd/testdata/test-stack/image/Dockerfile-stack")

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
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}
	_, err = cmdtest.RunAppsodyCmd(args, "../cmd/testData/test-stack", t)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "ENV APPSODY_DEBUG_KILL=trued") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_APPSODY_REGEXValue(t *testing.T) {
	restoreLine := ""
	file, err := ioutil.ReadFile("../cmd/testdata/test-stack/image/Dockerfile-stack")

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
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "lint"}
	_, err = cmdtest.RunAppsodyCmd(args, "../cmd/testdata/test-stack", t)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	for i, line := range lines {
		if strings.Contains(line, "ENV APPSODY_WATCH_REGEX='['") {
			lines[i] = restoreLine
		}
	}

	output = strings.Join(lines, "\n")
	err = ioutil.WriteFile("../cmd/testdata/test-stack/image/Dockerfile-stack", []byte(output), 0644)

	if err != nil {
		t.Fatal(err)
	}
}
func TestLintWithValidStack(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}
}

func TestLintWithInvalidStackName(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	newStackPath := filepath.Join(currentDir, "testdata", "test_stack")
	args := []string{"stack", "lint"}

	renErr := os.Rename(testStackPath, newStackPath)
	if renErr != nil {
		t.Fatal(renErr)
	}
	output, err := cmdtest.RunAppsodyCmd(args, newStackPath, t)

	renErr1 := os.Rename(newStackPath, testStackPath)

	if renErr1 != nil {
		t.Fatal(renErr1)
	}
	if !strings.Contains(output, "Stack directory name is invalid.") {
		t.Fatal(err)
	}

}

func TestLintWithMissingStackYaml(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}
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

	_, appsodyErr := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	if appsodyErr == nil { //Lint check should fail, if not fail the test
		t.Fatal(appsodyErr)
	}

	RestoreSampleStack(removeArray)
	writeErr := ioutil.WriteFile(removeYaml, []byte(file), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
}

func TestLintWithMissingImageProjectAndConfigDir(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
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

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	RestoreSampleStack(removeArray)
	writeErr := ioutil.WriteFile(filepath.Join(removeImage, "Dockerfile-stack"), []byte(file), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	if !strings.Contains(output, "Missing image directory") {
		t.Fatal(err)
	}

}

func TestLintWithMissingREADME(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}
	removeReadme := filepath.Join(testStackPath, "README.md")
	removeArray := []string{removeReadme}

	osErr := os.RemoveAll(removeReadme)
	if osErr != nil {
		t.Fatal(osErr)
	}

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	RestoreSampleStack(removeArray)

	if !strings.Contains(output, "Missing README.md") {
		t.Fatal(err)
	}

}

func TestLintWithMissingDockerfileStackAndLicense(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

	removeDockerfileStack := filepath.Join(testStackPath, "image", "Dockerfile-stack")
	removeLicense := filepath.Join(testStackPath, "image", "LICENSE")
	removeArray := []string{removeDockerfileStack, removeLicense}

	file, readErr := ioutil.ReadFile(removeDockerfileStack)
	if readErr != nil {
		t.Fatal(readErr)
	}
	for _, deleteFile := range removeArray {
		osErr := os.RemoveAll(deleteFile)
		if osErr != nil {
			t.Fatal(osErr)
		}
	}

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	RestoreSampleStack(removeArray)
	writeErr := ioutil.WriteFile(filepath.Join(removeDockerfileStack), []byte(file), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	if !strings.Contains(output, "Missing Dockerfile-stack") && !strings.Contains(output, "Missing LICENSE") {
		t.Fatal(err)
	}

}

func TestLintWithMissingTemplatesDirectory(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

	removeTemplatesDir := filepath.Join(testStackPath, "templates")
	removeArray := []string{removeTemplatesDir, filepath.Join(removeTemplatesDir, "default"), filepath.Join(removeTemplatesDir, "default", "app.js")}

	osErr := os.RemoveAll(removeTemplatesDir)
	if osErr != nil {
		t.Fatal(osErr)
	}

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	RestoreSampleStack(removeArray)

	if !strings.Contains(output, "Missing template directory") && !strings.Contains(output, "No templates found in") {
		t.Fatal(err)
	}

}

func TestLintWithInvalidVersion(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

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

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "Version must be formatted in accordance to semver") {
		t.Fatal(err)
	}

}

func TestLintWithLongNameAndDescription(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

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

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "Description must be under ") && !strings.Contains(output, "Stack name must be under ") {
		t.Fatal(err)
	}
}

func TestLintWithInvalidLicenseField(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

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

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "The stack.yaml SPDX license ID is invalid") {
		t.Fatal(err)
	}
}

func TestLintWithInvalidTemplatingValues(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

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

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "is not in an alphanumeric format") {
		t.Fatal(err)
	}
}

func TestLintWithInvalidRequirements(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

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

	output, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	restoreYaml := ioutil.WriteFile(stackYaml, []byte(file), 0644)
	if restoreYaml != nil {
		t.Fatal(restoreYaml)
	}

	if !strings.Contains(output, "is not in the correct format. See:") {
		t.Fatal(err)
	}

}

func RestoreSampleStack(fixStack []string) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
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
		}
	}
}
