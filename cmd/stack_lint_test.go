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

func TestLintWithMissingConfig(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint", testStackPath}
	_, err := cmdtest.RunAppsody(sandbox, args...)

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}
}

func TestLintWithMissingProject(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint", testStackPath}

	_, err := cmdtest.RunAppsody(sandbox, args...)

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}
}

func TestLintWithMissingFile(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"stack", "lint"}

	_, err := cmdtest.RunAppsody(sandbox, args...)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}
}
