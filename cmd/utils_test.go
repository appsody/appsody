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
	"os"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
)

var yamlGenTests = []struct {
	yamlTemplateGetter func() string // input
	testImageName      string
	testPortNum        int
	testServiceName    string
	testPullPolicy     bool
}{
	{getKNativeTemplate1, "TESTIMAGE", 9091, "TESTSERVICE", false},
	{getKNativeTemplate2, "TESTIMAGE", 9091, "TESTSERVICE", true},
	{getKNativeTemplateNoports, "TESTIMAGE", 9091, "TESTSERVICE", true},
}

// requires clean dir
func TestGenYAML(t *testing.T) {

	for numTest, test := range yamlGenTests {
		t.Run(fmt.Sprintf("Test YAML template %d", numTest), func(t *testing.T) {
			testServiceName := test.testServiceName
			testImageName := test.testImageName
			testPortNum := test.testPortNum
			testGetter := test.yamlTemplateGetter
			testPullPolicy := test.testPullPolicy
			yamlFileName, err := cmd.GenKnativeYaml(testGetter(), testPortNum, testServiceName, testImageName, testPullPolicy, "app-deploy.yaml", false)
			if err != nil {
				t.Fatal("Can't generate the YAML for KNative serving deploy. Error: ", err)
			}
			if _, err := os.Stat(yamlFileName); os.IsNotExist(err) {
				t.Fatal("Didn't find the file ", yamlFileName)
			} else { //clean up
				os.RemoveAll(yamlFileName)
			}
		})
	}
}
func getKNativeTemplate1() string {
	yamltempl := `
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: test
spec:
  runLatest:
    configuration:
      revisionTemplate:
        spec:
          container:
            image: myimage
            imagePullPolicy: Always
            ports:
            - containerPort: 8080
`
	return yamltempl
}

func getKNativeTemplate2() string {
	yamltempl := `
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: test
spec:
  runLatest:
    configuration:
      revisionTemplate:
        spec:
          container:
            image: myimage
            imagePullPolicy: Never
            ports:
            - containerPort: 8080
`
	return yamltempl
}

func getKNativeTemplateNoports() string {
	yamltempl := `
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: test
spec:
  runLatest:
    configuration:
      revisionTemplate:
        spec:
          container:
            image: myimage
            imagePullPolicy: Always
`
	return yamltempl
}

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
	// 127 chars is valid
	"a234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567",
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
	// 128 chars is invalid
	{"a2345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678",
		"a234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567"},
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

func TestInvalidVersionAgainstStack(t *testing.T) {
	reqArray := []cmd.StackRequirement{{Docker: "102.0.5", Appsody: "21.4.0"}}
	err := cmd.CheckStackRequirements(reqArray)

	if err == nil {
		t.Fatal(err)
	}
}
