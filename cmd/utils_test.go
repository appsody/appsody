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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

var validProjectNameTests = []string{
	"my-project",
	"my---project",
	"my-project1",
	"my-project123",
	"my-pr0ject",
	"myproject",
	"m",
	"m1",
	"appsody-project",
	// 68 chars is valid
	"a2345678901234567890123456789012345678901234567890123456789012345678",
}

func TestValidProjectNames(t *testing.T) {

	for _, testData := range validProjectNameTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		test := testData

		t.Run(fmt.Sprintf("Test Valid Project Name \"%s\"", test), func(t *testing.T) {
			isValid, err := cmd.IsValidProjectName(test)
			if err != nil {
				t.Error(err)
			}
			if !isValid {
				t.Error("Not a valid project name: ", test)
			}
			converted, err := cmd.ConvertToValidProjectName(test)
			if err != nil {
				t.Error(err)
			}
			if test != converted {
				t.Error("Valid project name not the same on conversion: ", test)
			}
		})
	}
}

var invalidProjectNameTests = []struct {
	input     string
	converted string
}{
	{"my-project-", "my-project-app"},
	{"-my-project", "appsody-my-project"},
	{"My-project", "my-project"},
	{"my-Project", "my-project"},
	{"1my-project", "appsody-1my-project"},
	{"my-project----", "my-project-app"},
	{"my-proj%ect", "my-proj-ect"},
	{"my-proj#$&%ect", "my-proj-ect"},
	{"M", "m"},
	{"-", "appsody-app"},
	{".", "appsody-app"},
	{"path/to/pr0ject", "pr0ject"},
	{"/path/to/pr0ject", "pr0ject"},
	{"path/to/1my-project", "appsody-1my-project"},
	// 69 chars is invalid
	{"a23456789012345678901234567890123456789012345678901234567890123456789",
		"a2345678901234567890123456789012345678901234567890123456789012345678"},
}

func TestInvalidProjectNames(t *testing.T) {

	for _, testData := range invalidProjectNameTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		test := testData

		t.Run(fmt.Sprintf("Test Invalid Project Name \"%s\"", test.input), func(t *testing.T) {
			isValid, err := cmd.IsValidProjectName(test.input)
			if err == nil {
				t.Error("Expected an error from IsValidProjectName but did not return one.")
			} else if !strings.Contains(err.Error(), "Invalid project-name") {
				t.Error("Expected the error to contain \"Invalid project-name\"", err)
			}
			if isValid {
				t.Error("Valid project name when expected to be invalid: ", test)
			}
			converted, err := cmd.ConvertToValidProjectName(test.input)
			if err != nil {
				t.Error(err)
			}
			if test.converted != converted {
				t.Errorf("Invalid project name \"%s\" converted to \"%s\" but expected \"%s\"", test.input, converted, test.converted)
			}
		})
	}
}

//Passes in impossibly high minimum versions of Docker and Appsody
func TestInvalidVersionAgainstStack(t *testing.T) {
	reqsMap := map[string]string{
		"Docker":  "402.05.6",
		"Appsody": "402.05.6",
	}
	log := &cmd.LoggingConfig{}
	var outBuffer bytes.Buffer
	log.InitLogging(&outBuffer, &outBuffer)

	err := cmd.CheckStackRequirements(log, reqsMap, false)

	if err == nil {
		t.Log(outBuffer.String())
		t.Fatal("Expected Error NOT thrown", reqsMap)
	}
}

var invalidCmdsTest = []struct {
	cmd      string
	args     []string
	expected string
}{
	{"ls", []string{"invalidname"}, "No such file or directory"},
	{"cp", []string{"invalidname", "alsoinavalidname"}, "No such file or directory"},
}

func TestInvalidCmdOutput(t *testing.T) {

	for _, testData := range invalidCmdsTest {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		test := testData

		t.Run(fmt.Sprintf("Test Invalid "+test.cmd+" Command"), func(t *testing.T) {
			invalidCmd := exec.Command(test.cmd, test.args...)
			out, err := cmd.SeparateOutput(invalidCmd)
			if err == nil {
				t.Error("Expected an error from '", test.cmd, strings.Join(test.args, " "), "' but it did not return one.")
			} else if !strings.Contains(out, test.expected) {
				t.Error("Expected the stdout to contain '" + test.expected + "'. It actually contains: " + out)
			}
		})

	}

}

var convertLabelTests = []struct {
	input          string
	expectedOutput string
}{
	{"org.opencontainers.image.created", "image.opencontainers.org/created"},
	{"dev.appsody.stack.id", "stack.appsody.dev/id"},
	{"dev.appsody.app.name", "app.appsody.dev/name"},
	{"dev.appsody.app-name", "appsody.dev/app-name"},
	{"dev.app-sody.app.name", "dev/app-sody.app.name"},
	{"d.name", "d/name"},
	{"app.name", "app/name"},
	{"app-name", "app-name"},
	{"Description", "Description"},
	{"maintainer", "maintainer"},
	{"dev.appsody.app.a23456789012345678901234567890123456789012345678901234567890123",
		"app.appsody.dev/a23456789012345678901234567890123456789012345678901234567890123"}, // exact length limit on name
}

func TestConvertLabelToKubeFormat(t *testing.T) {

	for _, testData := range convertLabelTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		test := testData

		t.Run(test.input, func(t *testing.T) {
			output, err := cmd.ConvertLabelToKubeFormat(test.input)
			if err != nil {
				t.Error(err)
			} else if output != test.expectedOutput {
				t.Errorf("Expected %s to convert to %s but got %s", test.input, test.expectedOutput, output)
			}
		})

	}
}

var invalidConvertLabelTests = []string{
	"inva$lid",
	".name",
	"dev.appsody.",
	"dev.appsody.app.a234567890123456789012345678901234567890123456789012345678901234", // one over length limit
}

func TestInvalidConvertLabelToKubeFormat(t *testing.T) {

	for _, testData := range invalidConvertLabelTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		test := testData

		t.Run(test, func(t *testing.T) {
			_, err := cmd.ConvertLabelToKubeFormat(test)
			if err == nil {
				t.Errorf("Expected error but got none converting %s", test)
			}
		})
	}
}

var getUpdateStringTests = []struct {
	input        string
	version      string
	latest       string
	updateString string
}{
	{"darwin", "1", "2", "Please run `brew upgrade appsody` to upgrade"},
	{"anythingelse", "1", "2", "Please go to https://appsody.dev/docs/getting-started/installation#upgrading-appsody and upgrade"},
	{"", "1", "2", "Please go to https://appsody.dev/docs/getting-started/installation#upgrading-appsody and upgrade"},
}

func TestGetUpdateString(t *testing.T) {

	for _, testData := range getUpdateStringTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		test := testData

		t.Run(test.input, func(t *testing.T) {
			output := cmd.GetUpdateString(test.input, test.version, test.latest)
			expectedOutput := fmt.Sprintf("\n*\n*\n*\n\nA new CLI update is available.\n%s from %s --> %s.\n\n*\n*\n*\n", test.updateString, test.version, test.latest)
			if output != expectedOutput {
				t.Errorf("Expected %s to convert to %s but got %s", test.input, expectedOutput, output)
			}
		})

	}
}

func TestNormalizeImageName(t *testing.T) {
	testImageNames := []string{"ubuntu", "ubuntu:latest", "ubuntu:17.1", "appsody/nodejs-express:0.2", "docker.io/appsody/nodejs-express:0.2", "index.docker.io/appsody/nodejs-express:0.2", "myregistry.com:8080/appsody/nodejs-express:0.2", "yada/yada/yada/yada"}
	normalizedTestImageNames := []string{"docker.io/ubuntu", "docker.io/ubuntu:latest", "docker.io/ubuntu:17.1", "appsody/nodejs-express:0.2", "docker.io/appsody/nodejs-express:0.2", "docker.io/appsody/nodejs-express:0.2", "myregistry.com:8080/appsody/nodejs-express:0.2"}

	for index, testData := range testImageNames {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		imageName := testData
		idx := index

		t.Run(imageName, func(t *testing.T) {
			output, err := cmd.NormalizeImageName(imageName)

			if err != nil {
				if idx < len(testImageNames)-1 {
					t.Errorf("Unexpected error: %v", err)
				}
			} else {
				expectedOutput := normalizedTestImageNames[idx]
				if output != expectedOutput {
					t.Errorf("Expected %s to convert to %s but got %s", imageName, expectedOutput, output)
				}
			}
		})

	}
}
func TestOverrideStackRegistry(t *testing.T) {
	testImageNames := []string{"ubuntu", "ubuntu:latest", "ubuntu:17.1", "appsody/nodejs-express:0.2", "docker.io/appsody/nodejs-express:0.2", "index.docker.io/appsody/nodejs-express:0.2", "another-registry.com:8080/appsody/nodejs-express:0.2", "yada/yada/yada/yada"}
	override := "my-registry.com:8080"
	normalizedTestImageNames := []string{"my-registry.com:8080/ubuntu", "my-registry.com:8080/ubuntu:latest", "my-registry.com:8080/ubuntu:17.1", "my-registry.com:8080/appsody/nodejs-express:0.2", "my-registry.com:8080/appsody/nodejs-express:0.2", "my-registry.com:8080/appsody/nodejs-express:0.2", "my-registry.com:8080/appsody/nodejs-express:0.2"}

	for index, testData := range testImageNames {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		imageName := testData
		idx := index

		t.Run(imageName, func(t *testing.T) {
			output, err := cmd.OverrideStackRegistry(override, imageName)

			if err != nil {
				if idx < len(testImageNames)-1 {
					t.Errorf("Unexpected error: %v", err)
				}
			} else {
				expectedOutput := normalizedTestImageNames[idx]
				if output != expectedOutput {
					t.Errorf("Expected %s to convert to %s but got %s", imageName, expectedOutput, output)
				}
			}
		})
	}

	t.Run("No override", func(t *testing.T) {
		output, err := cmd.OverrideStackRegistry("", "test")
		if err != nil || output != "test" {
			t.Errorf("Test with empty image override failed. Error: %v, output: %s", err, output)
		}
	})
}
func TestValidateHostName(t *testing.T) {
	testHostNames := make(map[string]bool)
	testHostNames["hostname"] = true
	testHostNames["hostname:80"] = true
	testHostNames["hostname.com"] = true
	testHostNames["hostname.company.com"] = true
	testHostNames["hostname:8080"] = true
	testHostNames["hostname:30080"] = true
	testHostNames["hostname.company.com:30080"] = true
	testHostNames["hostname.company.com:443"] = true
	testHostNames["host-name"] = true
	testHostNames["host/name"] = false
	testHostNames["host-name-"] = false
	testHostNames["host-name.my-company-"] = false
	testHostNames["host-name.-my-company"] = false
	testHostNames["-host-name.-my-company"] = false

	for key, value := range testHostNames {
		// need to set key and value to new variables scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		hostName := key
		val := value

		t.Run(hostName, func(t *testing.T) {
			match, err := cmd.ValidateHostNameAndPort(hostName)

			if err != nil || match != val {
				t.Errorf("Unexpected result for %s - valid should be %v, but it was not detected as such", hostName, val)
			}
		})

	}
}

func TestCopy(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	existingFile := "idoexist.bbb"
	nonExistentFile := "idontexistyet.aaa"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(existingFile)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
		err = os.Remove(nonExistentFile)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
	}()

	_, err := os.Create(existingFile)
	if err != nil {
		t.Errorf("Error creating the file: %v", err)
	}

	err = cmd.CopyFile(log, existingFile, nonExistentFile)

	if err != nil {
		t.Errorf(": '%v'", err.Error())
	}
}

func TestCopyFailFNF(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	existingFile := "idoexist.bbb"
	nonExistentFile := "idontexist.aaa"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(existingFile)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
	}()

	// Attempt to create the fake file
	_, err := os.Create(existingFile)
	if err != nil {
		t.Errorf("Error creating the file: %v", err)
	}

	err = cmd.CopyFile(log, nonExistentFile, existingFile)

	if err != nil {
		if !strings.Contains(err.Error(), "stat "+nonExistentFile+": no such file or directory") {
			t.Errorf("String \"stat "+nonExistentFile+": no such file or directory\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Errorf("Error: %v", err)
	}
}

func TestCopyFailPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	existingFile := "idoexist.bbb"
	nonExistentFile := "idontexist.aaa"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(existingFile)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
	}()

	// Attempt to create the fake file
	file, err := os.Create(existingFile)
	if err != nil {
		t.Errorf("Error creating the file: %s", err)
	}

	err = file.Chmod(0333)
	if err != nil {
		t.Errorf("Error changing file permissions: %s", err)
	}

	err = cmd.CopyFile(log, existingFile, nonExistentFile)

	if err != nil {
		if !strings.Contains(err.Error(), "Permission denied") {
			t.Errorf("String \"Permission denied\" not found in output: '%v'", err.Error())
		}

	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestMoveFailFNF(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	nonExistentFile := "idontexist.aaa"

	err := cmd.MoveDir(log, nonExistentFile, "../")

	if err != nil {
		if !strings.Contains(err.Error(), "Source "+nonExistentFile+" does not exist.") {
			t.Errorf("String \"Source "+nonExistentFile+" does not exist.\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestMove(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip()
	}

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	existingFile := "iamafile"
	newFileName := "iamachangedfile"

	// Ensure that the fake yaml file is deleted
	defer func() {

		if _, err := os.Stat(existingFile); err == nil {
			err := os.Remove(existingFile)
			if err != nil {
				t.Errorf("Error removing the file: %s", err)
			}
		}
		if _, err := os.Stat(newFileName); err == nil {
			err := os.Remove(newFileName)
			if err != nil {
				t.Errorf("Error removing the file: %s", err)
			}
		}

	}()

	// Attempt to create the fake file
	_, err := os.Create(existingFile)

	if err != nil {
		t.Errorf("Error creating the file: %v", err)
	}

	err = cmd.MoveDir(log, existingFile, newFileName)

	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestCopyFileFailPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	existingFile := "iamafile.test"
	existingDir := "iamadir/"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(existingFile)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}

		err = os.RemoveAll(existingDir)
		if err != nil {
			t.Errorf("Error removing the directory: %s", err)
		}
	}()

	// Attempt to create the fake file
	_, err := os.Create(existingFile)
	if err != nil {
		t.Errorf("Error creating the file: %v", err)
	}

	err = os.Mkdir(existingDir, 4440)

	if err != nil {
		t.Errorf("Error creating the directory: %v", err)
	}

	err = cmd.CopyFile(log, existingFile, existingDir)

	if err != nil {
		if !strings.Contains(err.Error(), "Permission denied") {
			t.Errorf("String \"Could not copy "+existingFile+" to"+existingDir+"\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestCopyDirFailSourceNotDir(t *testing.T) {
	// set up logging
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// create the directories in the sandbox

	// create a base directory from sandbox
	baseDir := sandbox.ProjectDir
	t.Log("baseDir is: ", baseDir)

	// create a source directory but don't do the os.Mkdir
	sourceDir := filepath.Join(baseDir, "sourceDir")
	t.Log("sourceDir is: ", sourceDir)

	// create a target directory
	targetDir := filepath.Join(baseDir, "targetDir")
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatal("Error creating targetDir: ", err)
	}
	t.Log("Created targetDir: ", targetDir)

	// copy the source directory into the target
	// this should fail as the target can not be an existing directory
	err = cmd.CopyDir(log, sourceDir, targetDir)
	if err != nil {
		if !strings.Contains(err.Error(), "Source "+sourceDir+" does not exist.") {
			t.Errorf("String \"Source "+sourceDir+" does not exist.\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}

}

func TestCopyDirFailTargetDirExistsDir(t *testing.T) {
	// set up logging
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// create the directories in the sandbox

	// create a base directory from sandbox
	baseDir := sandbox.ProjectDir
	t.Log("baseDir is: ", baseDir)

	// create a source directory
	sourceDir := filepath.Join(baseDir, "sourceDir")
	err := os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatal("Error creating sourceDir: ", err)
	}
	t.Log("Created sourceDir: ", sourceDir)

	// Create test file1 <name.extension>
	file1 := filepath.Join(sourceDir, "file1.test")
	data := []byte("home: " + sourceDir + "\n" + "generated-by-tests: Yes" + "\n")
	err = ioutil.WriteFile(file1, data, 0644)
	if err != nil {
		t.Fatal("Error writing file1: ", err)
	}
	t.Log("Created file1: ", file1)

	// create a target directory
	targetDir := filepath.Join(baseDir, "targetDir")
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatal("Error creating sourceDir: ", err)
	}
	t.Log("Created sourceDir: ", sourceDir)

	// copy the source directory into the target
	// this should fail as the target can not be an existing directory
	err = cmd.CopyDir(log, sourceDir, targetDir)
	if err != nil {
		if !strings.Contains(err.Error(), "Target "+targetDir+" exists. It should only be the name of the target directory for the copy") {
			t.Errorf("String \"Target "+targetDir+" exists. It should only be the name of the target directory for the copy\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}

}

func TestCopyDirFailTargetDirExistsFile(t *testing.T) {
	// set up logging
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// create the directories in the sandbox

	// create a base directory from sandbox
	baseDir := sandbox.ProjectDir
	t.Log("baseDir is: ", baseDir)

	// create a source directory
	sourceDir := filepath.Join(baseDir, "sourceDir")
	err := os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatal("Error creating sourceDir: ", err)
	}
	t.Log("Created sourceDir: ", sourceDir)

	// Create target file <name.extension>
	targetFile := filepath.Join(baseDir, "targetFile.test")
	data := []byte("home: " + baseDir + "\n" + "generated-by-tests: Yes" + "\n")
	err = ioutil.WriteFile(targetFile, data, 0644)
	if err != nil {
		t.Fatal("Error writing targetFile: ", err)
	}
	t.Log("Created targetFile: ", targetFile)

	// copy the source directory into the target
	// this should fail as the target can not be an existing directory
	err = cmd.CopyDir(log, sourceDir, targetFile)
	if err != nil {
		if !strings.Contains(err.Error(), "Target "+targetFile+" exists. It should only be the name of the target directory for the copy") {
			t.Errorf("String \"Target "+targetFile+" exists. It should only be the name of the target directory for the copy\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}

}

func TestCopyDirPass(t *testing.T) {
	// set up logging
	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	// use sandbox directories and cleanup
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// create the directories in the sandbox

	// create a base directory
	baseDir := sandbox.ProjectDir
	t.Log("baseDir is: ", baseDir)

	// create a source directory
	sourceDir := filepath.Join(baseDir, "sourceDir")
	err := os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatal("Error creating sourceDir: ", err)
	}
	t.Log("Created sourceDir: ", sourceDir)

	// Create test file1 <name.extension>
	file1 := filepath.Join(sourceDir, "file1.test")
	data := []byte("home: " + sourceDir + "\n" + "generated-by-tests: Yes" + "\n")
	err = ioutil.WriteFile(file1, data, 0644)
	if err != nil {
		t.Fatal("Error writing file1: ", err)
	}
	t.Log("Created file1: ", file1)

	// Create test file2 <.extension>
	file2 := filepath.Join(sourceDir, ".file2")
	data = []byte("home: " + sourceDir + "\n" + "generated-by-tests: Yes" + "\n")
	err = ioutil.WriteFile(file2, data, 0644)
	if err != nil {
		t.Fatal("Error writing file2: ", err)
	}
	t.Log("Created file2: ", file2)

	// create a source sub directory
	sourceSubDir := filepath.Join(sourceDir, "sourceSubDir")
	err = os.MkdirAll(sourceSubDir, 0755)
	if err != nil {
		t.Fatal("Error creating source sub dir: ", err)
	}
	t.Log("Created sourceSubDir: ", sourceSubDir)

	// Create test file3 <name.extension>
	file3 := filepath.Join(sourceSubDir, "file3.test")
	data = []byte("home: " + sourceSubDir + "\n" + "generated-by-tests: Yes" + "\n")
	err = ioutil.WriteFile(file3, data, 0644)
	if err != nil {
		t.Fatal("Error writing file3: ", err)
	}
	t.Log("Created file3: ", file3)

	// Create test file4 <.extension>
	file4 := filepath.Join(sourceSubDir, ".file4")
	data = []byte("home: " + sourceSubDir + "\n" + "generated-by-tests: Yes" + "\n")
	err = ioutil.WriteFile(file4, data, 0644)
	if err != nil {
		t.Fatal("Error writing file4: ", err)
	}
	t.Log("Created file4: ", file4)

	// Create test file5 symlink
	file5 := filepath.Join(sourceSubDir, "file5")
	err = os.Symlink(file1, file5)
	if err != nil {
		t.Fatal("Error writing file5: ", err)
	}
	t.Log("Created file5: ", file5)

	// create a target directory string but don't create the directory
	targetDir := filepath.Join(baseDir, "targetDir")
	t.Log("targetDir is: ", targetDir)

	// create the expected files
	var copyExpectedFiles []string

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		copyExpectedFiles = append(copyExpectedFiles, path)
		return nil
	})
	if err != nil {
		t.Fatal("Error walking expectedDir", err)
	}
	// println("****source files:****")
	// for _, copyExpectedFiles := range copyExpectedFiles {
	// 	fmt.Println(copyExpectedFiles)
	// }
	// replace sourceDir with targetDir as we expect the copy to just put the files from the sourceDir into the targetDir
	for i := range copyExpectedFiles {
		copyExpectedFiles[i] = strings.Replace(copyExpectedFiles[i], "sourceDir", "targetDir", -1)
	}
	// println("****copyExpectedFiles files:****")
	// for _, copyExpectedFiles := range copyExpectedFiles {
	// 	fmt.Println(copyExpectedFiles)
	// }

	// copy the source directory into the target
	err = cmd.CopyDir(log, sourceDir, targetDir)
	if err != nil {
		t.Fatal("Error copying sourceDir to targetDir ", err)
	}

	// create the result files
	var copyResultFiles []string

	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		copyResultFiles = append(copyResultFiles, path)
		return nil
	})
	if err != nil {
		t.Fatal("Error walking targetDir", err)
	}
	// println("****copyResultFiles files:****")
	// for _, copyResultFiles := range copyResultFiles {
	// 	fmt.Println(copyResultFiles)
	// }

	// compare the expected vs the results
	if len(copyExpectedFiles) != len(copyResultFiles) {
		t.Errorf("length of expected file %s array does not match length of result file array %s", copyExpectedFiles, copyResultFiles)
	}
	for i := range copyExpectedFiles {
		if copyExpectedFiles[i] != copyResultFiles[i] {
			t.Errorf("expected file %s does not match result file %s", copyExpectedFiles[i], copyResultFiles[i])
		}
	}

}

// func TestGetGitLables(t *testing.T)

// 	var outBuffer bytes.Buffer
// 	loggingConfig := &cmd.LoggingConfig{}
// 	loggingConfig.InitLogging(&outBuffer, &outBuffer)
// 	config := &cmd.RootCommandConfig{LoggingConfig: loggingConfig}

// 	gitLabels, err := cmd.GetGitLabels(config)

// 	if err != nil {
// 		t.Error("Error: ", err)
// 	}

// 	for key, value := range gitLabels {
// 		switch key {
// 		case "dev.appsody.image.commit.author", "dev.appsody.image.commit.committer":
// 			matched, err := regexp.MatchString(`^[a-zA-Z0-9-_\s]*\s<([a-zA-Z0-9_\-\.]+)@([a-zA-Z0-9_\-\.]+)\.([a-zA-Z]{2,5})>$`, value)
// 			if err != nil {
// 				t.Errorf("Error performing regular expression: %v", err)
// 			}
// 			if !matched {
// 				t.Errorf("The value '%s' in the label '%s' was not in the expected format", value, key)
// 			}
// 		case "dev.appsody.image.commit.date":
// 			matched, err := regexp.MatchString(`^[a-zA-Z]{3}\s[a-zA-Z]{3}\s[0-9]{1,2}\s[0-9]{2}:[0-9]{2}:[0-9]{2}\s[0-9]{4}\s\+[0-9]{4}$`, value)
// 			if err != nil {
// 				t.Errorf("Error performing regular expression: %v", err)
// 			}
// 			if !matched {
// 				t.Errorf("The value '%s' in the label '%s' was not in the expected format", value, key)
// 			}
// 		case "dev.appsody.image.commit.message":
// 			matched, err := regexp.MatchString(`^[A-Za-z0-9\W\s]*$`, value)
// 			if err != nil {
// 				t.Errorf("Error performing regular expression: %v", err)
// 			}
// 			if !matched {
// 				t.Errorf("The value '%s' in the label '%s' was not in the expected format", value, key)
// 			}
// 		case "org.opencontainers.image.documentation", "org.opencontainers.image.source", "org.opencontainers.image.url":
// 			matched, err := regexp.MatchString(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`, value)
// 			if err != nil {
// 				t.Errorf("Error performing regular expression: %v", err)
// 			}
// 			if !matched {
// 				t.Errorf("The value '%s' in the label '%s' was not in the expected format", value, key)
// 			}
// 		case "org.opencontainers.image.revision":
// 			matched, err := regexp.MatchString(`^[0-9a-z]*(-modified)?$`, value)
// 			if err != nil {
// 				t.Errorf("Error performing regular expression: %v", err)
// 			}
// 			if !matched {
// 				t.Errorf("The value '%s' in the label '%s' was not in the expected format", value, key)
// 			}
// 		default:
// 			t.Errorf("Unexpected value returned from GetGitLabels(): {%s:%s}", key, value)
// 		}
// 	}
// }

func TestImagePushDryrun(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	imageName := "irrelevant"

	err := cmd.ImagePush(loggingConfig, imageName, false, true)

	if err != nil {
		t.Errorf("Unexpected error when pretending to push the image: %s", err)
	}
}

func TestImagePushNoReg(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	imageName := "notvalid"

	err := cmd.ImagePush(loggingConfig, imageName, false, false)

	if err != nil {
		if !strings.Contains(err.Error(), "An image does not exist locally with the tag: "+imageName) {
			t.Errorf("String \"An image does not exist locally with the tag: %s, \" not found in output: '%v'", imageName, err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetFailNoRes(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	cmdParms := []string{}
	_, err := cmd.KubeGet(loggingConfig, cmdParms, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "exit status 1: You must specify the type of resource to get.") {
			t.Errorf("String \"exit status 1: You must specify the type of resource to get.\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetFailIncorrectRes(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	cmdParms := []string{"invalid"}
	_, err := cmd.KubeGet(loggingConfig, cmdParms, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "kubectl get failed: exit status 1: error: the server doesn't have a resource type \"invalid\"") {
			t.Errorf("String \"kubectl get failed: exit status 1: error: the server doesn't have a resource type \"invalid\"\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeApplyDryrun(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	fileName := "file"

	err := cmd.KubeApply(loggingConfig, fileName, "namespace", true)

	if err != nil {
		t.Errorf("Unexpected error from kube apply: %v", err)
	}
}

func TestKubeApplyFailFNF(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	fileName := "file"

	err := cmd.KubeApply(loggingConfig, fileName, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "kubectl apply failed: exit status 1: error: the path \""+fileName+"\" does not exist") {
			t.Errorf("String \"kubectl apply failed: exit status 1: error: the path \"%s\" does not exist\" not found in output: '%v'", fileName, err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeApplyFailFileInvalid(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	fileName := "file"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(fileName)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
	}()

	// Attempt to create the fake file
	_, err := os.Create(fileName)
	if err != nil {
		t.Errorf("Error creating the file: %v", err)
	}

	err = cmd.KubeApply(loggingConfig, fileName, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "kubectl apply failed: exit status 1: error: no objects passed to apply") {
			t.Errorf("String \"kubectl apply failed: exit status 1: error: no objects passed to apply\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeApplyFailPermission(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	fileName := "file"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(fileName)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
	}()

	// Attempt to create the fake file
	file, err := os.Create(fileName)
	if err != nil {
		t.Errorf("Error creating the file: %v", err)
	}

	err = file.Chmod(0333)
	if err != nil {
		t.Errorf("Error changing file permissions: %s", err)
	}

	err = cmd.KubeApply(loggingConfig, fileName, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "kubectl apply failed: exit status 1: error: open file: permission denied") {
			t.Errorf("String \"kubectl apply failed: exit status 1: error: open file: permission denied\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeDeleteDryrun(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	fileName := "file"

	err := cmd.KubeDelete(loggingConfig, fileName, "namespace", true)

	if err != nil {
		t.Errorf("Unexpected error from kube apply: %v", err)
	}
}

func TestKubeDeleteFailFNF(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}

	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	fileName := "file"

	err := cmd.KubeDelete(loggingConfig, fileName, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "kubectl delete failed: exit status 1: error: the path \""+fileName+"\" does not exist") {
			t.Errorf("String \"kubectl delete failed: exit status 1: error: the path \"%s\" does not exist\" not found in output: '%v'", fileName, err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeDeleteFailPermission(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	fileName := "file"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(fileName)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
	}()

	// Attempt to create the fake file
	file, err := os.Create(fileName)
	if err != nil {
		t.Errorf("Error creating the file: %v", err)
	}

	err = file.Chmod(0333)
	if err != nil {
		t.Errorf("Error changing file permissions: %s", err)
	}

	err = cmd.KubeDelete(loggingConfig, fileName, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "kubectl delete failed: exit status 1: error: open file: permission denied") {
			t.Errorf("String \"kubectl delete failed: exit status 1: error: open file: permission denied\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetNodePortURLFailNoService(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	_, err := cmd.KubeGetNodePortURL(loggingConfig, "", "namespace", false)

	if err != nil {
		if !strings.Contains(err.Error(), "Failed to find deployed service IP and Port: kubectl get failed: exit status 1: error: resource name may not be empty") {
			t.Errorf("String \"Failed to find deployed service IP and Port: kubectl get failed: exit status 1: error: resource name may not be empty\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetNodePortURLFailInvalidService(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	service := "definitelynotaservice"

	_, err := cmd.KubeGetNodePortURL(loggingConfig, service, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "Failed to find deployed service IP and Port: kubectl get failed: exit status 1: Error from server (NotFound): services \""+service+"\" not found") {
			t.Errorf("String \"Failed to find deployed service IP and Port: kubectl get failed: exit status 1: Error from server (NotFound): services \"%s\" not found\" not found in output: '%v'", service, err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetNodePortURLDryrun(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	service := "svc"

	_, err := cmd.KubeGetNodePortURL(loggingConfig, service, "", true)

	if err != nil {
		t.Errorf("Unexpected error from kube get: %v", err)
	}
}

func TestKubeGetDeploymentURLFailNoService(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	_, err := cmd.KubeGetDeploymentURL(loggingConfig, "", "namespace", false)

	if err != nil {
		if !strings.Contains(err.Error(), "Failed to find deployed service IP and Port: kubectl get failed: exit status 1: error: resource name may not be empty") {
			t.Errorf("String \"Failed to find deployed service IP and Port: kubectl get failed: exit status 1: error: resource name may not be empty\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetDeploymentURLFailInvalidService(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	service := "definitelynotaservice"

	_, err := cmd.KubeGetDeploymentURL(loggingConfig, service, "", false)

	if err != nil {
		if !strings.Contains(err.Error(), "Failed to find deployed service IP and Port: kubectl get failed: exit status 1: Error from server (NotFound): services \""+service+"\" not found") {
			t.Errorf("String \"Failed to find deployed service IP and Port: kubectl get failed: exit status 1: Error from server (NotFound): services \"%s\" not found\" not found in output: '%v'", service, err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetDeploymentURLDryrun(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	service := "svc"

	_, err := cmd.KubeGetDeploymentURL(loggingConfig, service, "", true)

	if err != nil {
		t.Errorf("Unexpected error from kube get: %v", err)
	}
}

func TestKubeGetRouteURLFailInvalidService(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	_, err := cmd.KubeGetRouteURL(loggingConfig, "", "namespace", false)

	if err != nil {
		if !strings.Contains(err.Error(), "Failed to find deployed service IP and Port: kubectl get failed: exit status 1: error: the server doesn't have a resource type") {
			t.Errorf("String \"Failed to find deployed service IP and Port: kubectl get failed: exit status 1: error: the server doesn't have a resource type\" not found\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetRouteURLDryrun(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	service := "svc"

	_, err := cmd.KubeGetRouteURL(loggingConfig, service, "", true)

	if err != nil {
		t.Errorf("Unexpected error from kube get: %v", err)
	}
}

func TestKubeGetKnativeURLFailInvalidService(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	_, err := cmd.KubeGetKnativeURL(loggingConfig, "", "namespace", false)

	if err != nil {
		if !strings.Contains(err.Error(), "kubectl get failed: exit status 1: error: the server doesn't have a resource type") {
			t.Errorf("String \"kubectl get failed: exit status 1: error: the server doesn't have a resource type\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}
}

func TestKubeGetKnativeURLDryrun(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	service := "svc"

	_, err := cmd.KubeGetKnativeURL(loggingConfig, service, "", true)

	if err != nil {
		t.Errorf("Unexpected error from kube get: %v", err)
	}
}

func TestExtractDockerEnvVars(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testDockerOptions1 := []string{
		"-w /path/to/dir -e A=Val1",
		"-w /path/to/dir    -e     A=Val1  ",
		"-e A=Val1 -w /path/to/dir",
		"-e A=Val1",
		"--env A=Val1",
	}
	testDockerOptions2 := []string{
		"--env A=Val1 -e B=Val2",
		"--env=A=Val1 -e=B=Val2",
		"--env A=Val1 -e=B=Val2",
		"--env=A=Val1 -e B=Val2",
		"-w /path/to/dir -e A=Val1 -e B=Val2",
		"-w /path/to/dir     -e A=Val1    -e     B=Val2",
		"--workdir /path/to/dir -e A=Val1 -e B=Val2",
		"--workdir /path/to/dir     -e A=Val1   -e B=Val2",
		"--workdir /path/to/dir -e A=Val1 -e B=Val2 -m 1024",
		"--workdir /path/to/dir --env A=Val1 --env B=Val2",
		"--workdir /path/to/dir -e A=Val1 --env B=Val2",
		"-e A=Val1 --workdir /path/to/dir --env B=Val2",
	}
	testDockerOptions3 := []string{
		"--workdir /path/to/dir -m 1024",
		"--workdir /path/to/dir -e A Val1",
		"-env A=1",
		"--env A",
		"whatever --env -e",
	}

	testDockerOptions4 := []string{
		"--env-file " + filepath.Join(sandbox.TestDataPath, "test_docker_options", "test_docker_options.env"),
		"--env-file=" + filepath.Join(sandbox.TestDataPath, "test_docker_options", "test_docker_options.env"),
		"--env-file " + filepath.Join(sandbox.TestDataPath, "test_docker_options", "test_docker_options.env") + " -w /whatever/it/is",
		" --env-file " + filepath.Join(sandbox.TestDataPath, "test_docker_options", "test_docker_options.env") + "   -w     /whatever/it/is",
	}

	testDockerOptions5 := []string{
		"--env-file " + filepath.Join(sandbox.TestDataPath, "test_docker_options", "test_docker_options.env") + " -e VAR1=Override -e VAR4=Override -e VAR6=VAL6",
		"-e VAR1=Override -e VAR4=Override --env-file " + filepath.Join(sandbox.TestDataPath, "test_docker_options", "test_docker_options.env") + " -e VAR1=Override -e VAR4=Override -e VAR6=VAL6",
	}

	result1 := make(map[string]string)
	result1["A"] = "Val1"

	result2 := make(map[string]string)
	result2["A"] = "Val1"
	result2["B"] = "Val2"

	result3 := make(map[string]string)
	result3["VAR1"] = "VAL1"
	result3["VAR2"] = "VAL2"
	result3["VAR3"] = ""
	result3["VAR4"] = "VAL4"
	result3["VAR7"] = "VAL\"7"
	result3["VAR'8"] = "VAL'8"

	result4 := make(map[string]string)
	result4["VAR1"] = "Override"
	result4["VAR2"] = "VAL2"
	result4["VAR3"] = ""
	result4["VAR4"] = "Override"
	result4["VAR6"] = "VAL6"
	result4["VAR7"] = "VAL\"7"
	result4["VAR'8"] = "VAL'8"

	for _, testData := range testDockerOptions1 {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		dockerOption := testData

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != len(result1) {
				t.Errorf("TEST 1 - Expected %d element(s) and got %d - %v", len(result1), len(envVars), envVars)
			}
			for key, value := range envVars {
				if value != result2[key] {
					t.Errorf("TEST 1 - Expected %s for env var %s and got %s - %v", result2[key], key, value, envVars)
				}
			}
		})
	}
	for _, testData := range testDockerOptions2 {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		dockerOption := testData

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != len(result2) {
				t.Errorf("TEST 2 - Expected %d element(s) and got %d - %v", len(result2), len(envVars), envVars)
			}
			for key, value := range envVars {
				if value != result2[key] {
					t.Errorf("TEST 2 - Expected %s for env var %s and got %s - %v", result2[key], key, value, envVars)
				}
			}
		})
	}
	for _, testData := range testDockerOptions3 {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		dockerOption := testData

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != 0 {
				t.Errorf("TEST 3 - Expected 0 element(s) and got %d - %v", len(envVars), envVars)
			}

		})
	}
	for _, testData := range testDockerOptions4 {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		dockerOption := testData

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != len(result3) {
				t.Errorf("TEST 4 - Expected %d element(s) and got %d - %v", len(result3), len(envVars), envVars)
			}
			for key, value := range envVars {
				if value != result3[key] {
					t.Errorf("TEST 4 - Expected %s for env var %s and got %s - %v", result3[key], key, value, envVars)
				}
			}

		})
	}
	for _, testData := range testDockerOptions5 {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		dockerOption := testData

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != len(result4) {
				t.Errorf("TEST 5 - Expected %d element(s) and got %d - %v", len(result4), len(envVars), envVars)
			}
			for key, value := range envVars {
				if value != result4[key] {
					t.Errorf("TEST 5 - Expected %s for env var %s and got %s - %v", result4[key], key, value, envVars)
				}
			}

		})
	}
}

var argsToStringTests = []struct {
	input          []string
	expectedOutput string
}{
	{[]string{""}, ""},
	{[]string{}, ""},
	{[]string{"appsody", "list"}, `appsody list`},
	{[]string{"appsody", "list\"quote"}, `appsody "list\"quote"`},
	{[]string{"appsody", "--name", "contains spaces"}, `appsody --name "contains spaces"`},
	{[]string{"kubectl", "get", "svc", "hello", "-o", "jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort}"},
		`kubectl get svc hello -o "jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort}"`},
}

func TestArgsToString(t *testing.T) {

	for _, testData := range argsToStringTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		test := testData

		t.Run(fmt.Sprint(test.input), func(t *testing.T) {
			output := cmd.ArgsToString(test.input)
			if output != test.expectedOutput {
				t.Errorf("Expected %s to convert to %s but got %s", test.input, test.expectedOutput, output)
			}
		})

	}
}
func TestRemoveIfExists(t *testing.T) {

	err := os.MkdirAll("./test/test-child/", 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = cmd.RemoveIfExists("./test/")
	if err != nil {
		t.Fatal(err)
	}

	parentDirExists, err := cmd.Exists("./test/")
	if err != nil {
		t.Fatal(err)
	}
	childDirExists, err := cmd.Exists("./test/test-child/")
	if err != nil {
		t.Fatal(err)
	}
	if parentDirExists || childDirExists {
		t.Fatal("Folders were not removed")
	}

}
