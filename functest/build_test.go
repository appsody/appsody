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
	"strings"
	"testing"
	"time"

	"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
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

		// first add the test repo index
		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}

		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsodyCmd([]string{"init", stackRaw[i]}, projectDir)
		if err != nil {
			t.Fatal(err)
		}

		// appsody build
		runChannel := make(chan error)
		imageName := "testbuildimage"
		go func() {
			_, err = cmdtest.RunAppsodyCmd([]string{"build", "--tag", imageName}, projectDir)
			runChannel <- err
		}()

		// It will take a while for the image to build, so lets use docker image ls to wait for it
		t.Log("calling docker image ls to wait for the image")
		imageBuilt := false
		count := 900
		for {
			dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"image", "ls", imageName})
			if dockerErr != nil {
				t.Log("Ignoring error running docker image ls "+imageName, dockerErr)
			}
			if strings.Contains(dockerOutput, imageName) {
				t.Log("docker image " + imageName + " was found")
				imageBuilt = true
			} else {
				time.Sleep(2 * time.Second)
				count = count - 1
			}
			if count == 0 || imageBuilt {
				break
			}
		}

		if !imageBuilt {
			t.Fatal("image was never built")
		}

		//delete the image
		deleteImage(imageName)

		// clean up
		cleanup()
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
	// first add the test repo index
	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-build-labels-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	t.Log("Created project dir: " + projectDir)

	// appsody init
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express"}, projectDir)
	t.Log("Running appsody init...")
	if err != nil {
		t.Fatal(err)
	}

	copyCmd := exec.Command("cp", "../cmd/testdata/.appsody-config.yaml", projectDir)
	err = copyCmd.Run()
	t.Log("Copying .appsody-config.yaml to project dir...")
	if err != nil {
		t.Fatal(err)
	}

	// appsody build
	runChannel := make(chan error)
	imageName := "testbuildimage"
	go func() {
		_, err = cmdtest.RunAppsodyCmdExec([]string{"build", "--tag", imageName}, projectDir)
		runChannel <- err
	}()

	// It will take a while for the image to build, so lets use docker image ls to wait for it
	t.Log("calling docker image ls to wait for the image")
	imageBuilt := false
	count := 900
	for {
		dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"image", "ls", imageName})
		if dockerErr != nil {
			t.Log("Ignoring error running docker image ls "+imageName, dockerErr)
		}
		if strings.Contains(dockerOutput, imageName) {
			t.Log("docker image " + imageName + " was found")
			imageBuilt = true
		} else {
			time.Sleep(2 * time.Second)
			count = count - 1
		}
		if count == 0 || imageBuilt {
			break
		}
	}

	if !imageBuilt {
		t.Fatal("image was never built")
	}

	inspectOutput, inspectErr := cmdtest.RunDockerCmdExec([]string{"inspect", imageName})
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
	deleteImage(imageName)

	// clean up
	cleanup()
}

func deleteImage(imageName string) {
	_, err := cmdtest.RunDockerCmdExec([]string{"image", "rm", imageName})
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

		// first add the test repo index
		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}

		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsodyCmd([]string{"init", stackRaw[i]}, projectDir)
		if err != nil {
			t.Fatal(err)
		}

		// appsody build
		runChannel := make(chan error)
		imageName := "testbuildimage"
		pullURL := "my-pull-url"

		go func() {
			_, err = cmdtest.RunAppsodyCmd([]string{"build", "--tag", imageName, "--pull-url", pullURL, "--knative"}, projectDir)
			runChannel <- err
		}()

		// It will take a while for the image to build, so lets use docker image ls to wait for it
		t.Log("calling docker image ls to wait for the image")
		imageBuilt := false
		count := 900
		for {
			dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"image", "ls", imageName})
			if dockerErr != nil {
				t.Log("Ignoring error running docker image ls "+imageName, dockerErr)
			}
			if strings.Contains(dockerOutput, imageName) {
				t.Log("docker image " + imageName + " was found")
				imageBuilt = true
			} else {
				time.Sleep(2 * time.Second)
				count = count - 1
			}
			if count == 0 || imageBuilt {
				break
			}
		}

		if !imageBuilt {
			t.Fatal("image was never built")
		}

		checkDeploymentConfig(t, pullURL, imageName)

		//delete the image
		deleteImage(imageName)

		// clean up
		cleanup()
	}
}

func checkDeploymentConfig(t *testing.T, pullURL string, imageTag string) {
	_, err := os.Stat(deployFile)
	if err != nil && os.IsNotExist(err) {
		t.Fatalf("Could not find %s", deployFile)
	}
	yamlFileBytes, err := ioutil.ReadFile(deployFile)

	var appsodyApplication v1beta1.AppsodyApplication

	err = yaml.Unmarshal(yamlFileBytes, &appsodyApplication)
	if err != nil {
		t.Logf("app-deploy.yaml formatting error: %s", err)
	}

	expectedApplicationImage := pullURL + "/" + imageTag
	if appsodyApplication.Spec.ApplicationImage != expectedApplicationImage {
		t.Fatal("Incorrect ApplicationImage in app-deploy.yaml")
	}

	if !*appsodyApplication.Spec.CreateKnativeService {
		t.Fatal("CreateKnativeService not set to true in the app-deploy.yaml when using --knative flag")
	}
}
