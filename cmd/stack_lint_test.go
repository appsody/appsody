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
	"log"
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
		log.Fatalln(err)
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
		log.Fatalln(err)
	}

	args := []string{"stack", "lint"}
	_, err = cmdtest.RunAppsodyCmdExec(args, "../cmd/testData/test-stack")

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
		log.Fatalln(err)
	}
}

func TestAPPSODY_MOUNTSMissingInDockerfileStack(t *testing.T) {
	restoreLine := ""
	file, err := ioutil.ReadFile("../cmd/testdata/test-stack/image/Dockerfile-stack")
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}

	args := []string{"stack", "lint"}
	_, err = cmdtest.RunAppsodyCmdExec(args, "../cmd/testdata/test-stack")

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
		log.Fatalln(err)
	}
}

func TestAPPSODY_WATCH_DIRPRESENTAndONCHANGEMissingInDockerfileStack(t *testing.T) {
	restoreLine := ""
	file, err := ioutil.ReadFile("../cmd/testdata/test-stack/image/Dockerfile-stack")

	if err != nil {
		log.Fatalln(err)
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
				log.Fatalln(err)
			}

			args := []string{"stack", "lint"}
			_, err = cmdtest.RunAppsodyCmdExec(args, "../cmd/testdata/test-stack")

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
			_, err = cmdtest.RunAppsodyCmdExec(args, "../cmd/testdata/test-stack")

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
		log.Fatalln(err)
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
	_, err = cmdtest.RunAppsodyCmdExec(args, "../cmd/testData/test-stack")

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
		log.Fatalln(err)
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
	_, err = cmdtest.RunAppsodyCmdExec(args, "../cmd/testdata/test-stack")

	if err == nil { //Lint check should fail, if not fail the test
		log.Fatalln(err)
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

	_, err := cmdtest.RunAppsodyCmdExec(args, testStackPath)

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}
}

func TestLintWithMissingConfig(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}
	removeConf := filepath.Join(testStackPath, "image", "config")
	removeArray := []string{removeConf, filepath.Join(removeConf, "app-deploy.yaml")}

	os.RemoveAll(removeConf)

	_, err := cmdtest.RunAppsodyCmdExec(args, testStackPath)

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

	_, err := cmdtest.RunAppsodyCmdExec(args, testStackPath)

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
}

func TestLintWithMissingFile(t *testing.T) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	args := []string{"stack", "lint"}
	removeReadme := filepath.Join(testStackPath, "README.md")
	removeArray := []string{removeReadme}

	osErr := os.RemoveAll(removeReadme)
	if osErr != nil {
		t.Fatal(osErr)
	}

	_, err := cmdtest.RunAppsodyCmdExec(args, testStackPath)

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}

	RestoreSampleStack(removeArray)
}

func RestoreSampleStack(fixStack []string) {
	currentDir, _ := os.Getwd()
	testStackPath := filepath.Join(currentDir, "testdata", "test-stack")
	for _, missingContent := range fixStack {
		if missingContent == filepath.Join(testStackPath, "image/config") || missingContent == filepath.Join(testStackPath, "image/project") {
			osErr := os.Mkdir(missingContent, os.ModePerm)
			if osErr != nil {
				log.Println(osErr)
			}
		} else {
			_, osErr := os.Create(missingContent)
			if osErr != nil {
				log.Println(osErr)
			}
		}
	}
}
