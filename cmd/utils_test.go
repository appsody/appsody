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
}{
	{getKNativeTemplate1, "TESTIMAGE", 9091, "TESTSERVICE"},
	{getKNativeTemplateNoports, "TESTIMAGE", 9091, "TESTSERVICE"},
}

// requires clean dir
func TestGenYAML(t *testing.T) {

	for numTest, test := range yamlGenTests {
		t.Run(fmt.Sprintf("Test YAML template %d", numTest), func(t *testing.T) {
			testServiceName := test.testServiceName
			testImageName := test.testImageName
			testPortNum := test.testPortNum
			testGetter := test.yamlTemplateGetter
			yamlFileName, err := cmd.GenKnativeYaml(testGetter(), testPortNum, testServiceName, testImageName)
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
