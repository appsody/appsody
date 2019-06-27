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
			yamlFileName, err := cmd.GenKnativeYaml(testGetter(), testPortNum, testServiceName, testImageName, testPullPolicy)
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
