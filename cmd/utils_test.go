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
	"os/exec"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
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

	for _, test := range validProjectNameTests {
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

	for _, test := range invalidProjectNameTests {
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

	for _, test := range invalidCmdsTest {

		invalidCmd := exec.Command(test.cmd, test.args...)

		t.Run(fmt.Sprintf("Test Invalid "+test.cmd+" Command"), func(t *testing.T) {
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
	for _, test := range convertLabelTests {
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
	for _, test := range invalidConvertLabelTests {
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
	for _, test := range getUpdateStringTests {
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
	for idx, imageName := range testImageNames {

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
	for idx, imageName := range testImageNames {

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
		t.Run("No override", func(t *testing.T) {
			output, err := cmd.OverrideStackRegistry("", "test")
			if err != nil || output != "test" {
				t.Errorf("Test with empty image override failed. Error: %v, output: %s", err, output)
			}
		})

	}
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
	for hostName, val := range testHostNames {

		t.Run(hostName, func(t *testing.T) {
			match, err := cmd.ValidateHostNameAndPort(hostName)

			if err != nil || match != val {
				t.Errorf("Unexpected result for %s - valid should be %v, but it was not detected as such", hostName, val)
			}
		})

	}
}

func TestExtractDockerEnvVars(t *testing.T) {
	testDockerOptions1 := []string{
		"-w /path/to/dir -e A=Val1",
		"-w /path/to/dir    -e     A=Val1  ",
		"-e A=Val1 -w /path/to/dir",
		"-e A=Val1",
		"--env A=Val1",
	}
	testDockerOptions2 := []string{
		"--env A=Val1 -e B=Val2",
		"-w /path/to/dir -e A=Val1 -e B=Val2",
		"-w /path/to/dir     -e A=Val1    -e     B=Val2",
		"--workdir /path/to/dir -e A=Val1 -e B=Val2",
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

	result1 := make(map[string]string)
	result1["A"] = "Val1"

	result2 := make(map[string]string)
	result2["A"] = "Val1"
	result2["B"] = "Val2"

	for _, dockerOption := range testDockerOptions1 {

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != len(result1) {
				t.Errorf("Expected %d element(s) and got %d - %v", len(result1), len(envVars), envVars)
			}
			for key, value := range envVars {
				if value != result2[key] {
					t.Errorf("Expected %s for env var %s and got %s - %v", result2[key], key, value, envVars)
				}
			}
		})
	}
	for _, dockerOption := range testDockerOptions2 {

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != len(result2) {
				t.Errorf("Expected %d element(s) and got %d - %v", len(result2), len(envVars), envVars)
			}
			for key, value := range envVars {
				if value != result2[key] {
					t.Errorf("Expected %s for env var %s and got %s - %v", result2[key], key, value, envVars)
				}
			}
		})
	}
	for _, dockerOption := range testDockerOptions3 {

		t.Run(dockerOption, func(t *testing.T) {
			envVars, err := cmd.ExtractDockerEnvVars(dockerOption)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(envVars) != 0 {
				t.Errorf("Expected 0 element(s) and got %d - %v", len(envVars), envVars)
			}

		})
	}
}
