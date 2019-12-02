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
package functest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
	"sigs.k8s.io/yaml"
)

// Simple test for appsody build command. A future enhancement would be to verify the image that gets built.
func TestBuildSimple(t *testing.T) {

	t.Log("stacksList is: ", stacksList)

	// if stacksList is empty there is nothing to test so return
	if stacksList == "" {
		t.Log("stacksList is empty, exiting test...")
		return
	}

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		t.Log("***Testing stack: ", stackRaw[i], "***")

		sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
		defer cleanup()

		// first add the test repo index
		_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsody(sandbox, "init", stackRaw[i])
		if err != nil {
			t.Fatal(err)
		}

		// appsody build
		imageName := "testbuildimage"
		_, err = cmdtest.RunAppsody(sandbox, "build", "--tag", imageName)
		if err != nil {
			t.Fatal("The appsody build command failed: ", err)
		}

		//delete the image
		deleteImage(imageName, t)
	}
}

var ociPrefixKey = "org.opencontainers.image."
var openContainerLabels = []string{
	"created",
	"authors",
	"version",
	"licenses",
	"title",
	"description",
}

var appsodyPrefixKey = "dev.appsody.stack."
var appsodyStackLabels = []string{
	// These will need updating when the stacks CI is updated
	//"id",
	"tag",
	"version",
	"configured",
}

func TestBuildLabels(t *testing.T) {
	stacksList = "incubator/nodejs"
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// first add the test repo index
	_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", "../cmd/testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// appsody init
	_, err = cmdtest.RunAppsody(sandbox, "init", "nodejs-express")
	t.Log("Running appsody init...")
	if err != nil {
		t.Fatal(err)
	}

	copyCmd := exec.Command("cp", "../cmd/testdata/.appsody-config.yaml", sandbox.ProjectDir)
	err = copyCmd.Run()
	t.Log("Copying .appsody-config.yaml to project dir...")
	if err != nil {
		t.Fatal(err)
	}

	// appsody build
	imageName := "testbuildimage"
	_, err = cmdtest.RunAppsody(sandbox, "build", "--tag", imageName)
	if err != nil {
		t.Fatalf("Error on appsody build: %v", err)
	}

	inspectOutput, inspectErr := cmdtest.RunDockerCmdExec([]string{"inspect", imageName}, t)
	if inspectErr != nil {
		t.Fatal(inspectErr)
	}

	var inspect []map[string]interface{}

	err = json.Unmarshal([]byte(inspectOutput), &inspect)
	if err != nil {
		t.Fatal(err)
	}

	config := inspect[0]["Config"].(map[string]interface{})
	labelsMap := config["Labels"].(map[string]interface{})

	for _, label := range appsodyStackLabels {
		if labelsMap[appsodyPrefixKey+label] == nil {
			t.Errorf("Could not find %s%s label in Docker image!", appsodyPrefixKey, label)
		}
	}

	if labelsMap["dev.appsody.app.name"] == nil {
		t.Error("Could not find requested stack label in Docker image!")
	}

	for _, label := range openContainerLabels {
		if labelsMap[ociPrefixKey+label] == nil {
			t.Errorf("Could not find %s%s label in Docker image!", ociPrefixKey, label)
		}
	}

	//delete the image
	deleteImage(imageName, t)
}

func deleteImage(imageName string, t *testing.T) {
	_, err := cmdtest.RunDockerCmdExec([]string{"image", "rm", imageName}, t)
	if err != nil {
		fmt.Printf("Ignoring error running docker image rm: %s", err)
	}
}

func TestDeploymentConfig(t *testing.T) {
	t.Log("stacksList is: ", stacksList)

	// if stacksList is empty there is nothing to test so return
	if stacksList == "" {
		t.Log("stacksList is empty, exiting test...")
		return
	}

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		t.Log("***Testing stack: ", stackRaw[i], "***")

		sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
		defer cleanup()

		// first add the test repo index
		_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsody(sandbox, "init", stackRaw[i])
		if err != nil {
			t.Fatal(err)
		}

		// appsody build
		imageName := filepath.Base(sandbox.ProjectDir)
		pullURL := "my-pull-url"

		_, err = cmdtest.RunAppsody(sandbox, "build", "--tag", imageName, "--pull-url", pullURL, "--knative")
		if err != nil {
			t.Error("appsody build command returned err: ", err)
		}
		checkDeploymentConfig(t, filepath.Join(sandbox.ProjectDir, deployFile), pullURL, imageName)

		//delete the image
		deleteImage(imageName, t)
	}
}

func checkDeploymentConfig(t *testing.T, deployFile string, pullURL string, imageTag string) {
	_, err := os.Stat(deployFile)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Could not find %s", deployFile)
		return
	}
	yamlFileBytes, err := ioutil.ReadFile(deployFile)
	if err != nil {
		t.Errorf("Could not read %s: %s", deployFile, err)
	}

	var appsodyApplication v1beta1.AppsodyApplication

	err = yaml.Unmarshal(yamlFileBytes, &appsodyApplication)
	if err != nil {
		t.Logf("app-deploy.yaml formatting error: %s", err)
	}

	expectedApplicationImage := imageTag
	if pullURL != "" {
		expectedApplicationImage = pullURL + "/" + imageTag
	}

	if appsodyApplication.Spec.ApplicationImage != expectedApplicationImage {
		t.Errorf("Incorrect ApplicationImage in app-deploy.yaml. Expected %s but found %s", expectedApplicationImage, appsodyApplication.Spec.ApplicationImage)
	}

	if *appsodyApplication.Spec.CreateKnativeService != true {
		t.Error("CreateKnativeService not set to true in the app-deploy.yaml when using --knative flag")
	}

	verifyImageAndConfigLabelsMatch(t, appsodyApplication, imageTag)
}

func verifyImageAndConfigLabelsMatch(t *testing.T, appsodyApplication v1beta1.AppsodyApplication, imageTag string) {
	args := []string{"inspect", "--format='{{json .Config.Labels }}'", imageTag}
	output, err := cmdtest.RunDockerCmdExec(args, t)
	if err != nil {
		t.Errorf("Error inspecting docker image: %s", err)
	}

	output = strings.ReplaceAll(output, "\n", "")
	output = strings.ReplaceAll(output, "'", "")

	var imageLabels map[string]string
	err = json.Unmarshal([]byte(output), &imageLabels)
	if err != nil {
		t.Errorf("Could not unmarshall docker labels: %s", err)
	}

	for key, value := range imageLabels {
		key, err = cmd.ConvertLabelToKubeFormat(key)
		if err != nil {
			t.Errorf("Could not convert label to Kubernetes format: %s", err)
		}

		label := appsodyApplication.Labels[key]
		annotation := appsodyApplication.Annotations[key]
		if label == "" && annotation == "" {
			t.Errorf("Could not find label %s in deployment config", key)
		}

		if label != "" && label != value {
			t.Errorf("Mismatch of %s label between built image and deployment config", key)
		}

		if annotation != "" && annotation != value {
			t.Errorf("Mismatch of %s label between built image and deployment config", key)
		}
	}

}
