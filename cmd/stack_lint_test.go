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
	_, err := cmdtest.RunAppsodyCmd(args, newStackPath, t)

	if err == nil {
		t.Fatal(err)
	}

	renErr1 := os.Rename(newStackPath, testStackPath)

	if renErr1 != nil {
		t.Fatal(renErr1)
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

	os.RemoveAll(removeYaml)

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

func TestLintWithMissingImage(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}
	removeImage := filepath.Join(testStackPath, "image")
	file, readErr := ioutil.ReadFile(filepath.Join(removeImage, "Dockerfile-stack"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	removeArray := []string{removeImage, filepath.Join(removeImage, "config"), filepath.Join(removeImage, "project"), filepath.Join(removeImage, "config", "app-deploy.yaml"), filepath.Join(removeImage, "project", "Dockerfile"), filepath.Join(removeImage, "LICENSE"), filepath.Join(removeImage, "Dockerfile-stack")}

	os.RemoveAll(removeImage)

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
	writeErr := ioutil.WriteFile(filepath.Join(removeImage, "Dockerfile-stack"), []byte(file), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
}

func TestLintWithMissingConfig(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}
	removeConf := filepath.Join(testStackPath, "image", "config")
	removeArray := []string{removeConf, filepath.Join(removeConf, "app-deploy.yaml")}

	os.RemoveAll(removeConf)

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
}

func TestLintWithMissingProject(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}
	removeProj := filepath.Join(testStackPath, "image", "project")
	removeArray := []string{removeProj, filepath.Join(removeProj, "Dockerfile")}

	osErr := os.RemoveAll(removeProj)

	if osErr != nil {
		t.Fatal(osErr)
	}

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
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

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
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
	for _, file := range removeArray {
		osErr := os.RemoveAll(file)
		if osErr != nil {
			t.Fatal(osErr)
		}
	}

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)

	if err == nil {
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
	writeErr := ioutil.WriteFile(filepath.Join(removeDockerfileStack), []byte(file), 0644)
	if writeErr != nil {
		t.Fatal(writeErr)
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

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)
	if err == nil {
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
}

func TestLintWithMissingTemplateInTemplatesDirectory(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

	removeTemplate := filepath.Join(testStackPath, "templates", "default")
	removeArray := []string{removeTemplate, filepath.Join(removeTemplate, "app.js")}

	osErr := os.RemoveAll(removeTemplate)
	if osErr != nil {
		t.Fatal(osErr)
	}

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)
	if err == nil {
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
}

func TestLintWithConfigYamlInTemplate(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}

	addConfigYaml := filepath.Join(testStackPath, "templates", "default", ".appsody-config.yaml")

	_, osErr := os.Create(addConfigYaml)
	if osErr != nil {
		t.Fatal(osErr)
	}

	_, err := cmdtest.RunAppsodyCmd(args, testStackPath, t)
	if err == nil {
		t.Fatal(err)
	}

	removeErr := os.RemoveAll(addConfigYaml)
	if removeErr != nil {
		t.Fatal(removeErr)
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
