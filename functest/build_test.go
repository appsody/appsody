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
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// Simple test for appsody build command. A future enhancement would be to verify the image that gets built.
func TestBuildSimple(t *testing.T) {

	t.Log("stacksList is: ", stacksList)

	// if stacksList is empty there is nothing to test so return
	if stacksList == "" {
		t.Log("stacksList is empty, exiting test...")
		return
	}

	// replace incubator with appsodyhub to match current naming convention for repos
	stacksList = strings.Replace(stacksList, "incubator", "appsodyhub", -1)

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
		projectDir, err := ioutil.TempDir("", "appsody-build-simple-test")
		if err != nil {
			t.Fatal(err)
		}

		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		// appsody init
		_, err = cmdtest.RunAppsodyCmdExec([]string{"init", stackRaw[i]}, projectDir)
		t.Log("Running appsody init...")
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

		//delete the image
		deleteImage(imageName)

		// clean up
		cleanup()
	}
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
	log.Println("Created project dir: " + projectDir)

	// appsody init
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express"}, projectDir)
	log.Println("Running appsody init...")
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
	fmt.Println("calling docker image ls to wait for the image")
	imageBuilt := false
	count := 900
	for {
		dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"image", "ls", imageName})
		if dockerErr != nil {
			log.Print("Ignoring error running docker image ls "+imageName, dockerErr)
		}
		if strings.Contains(dockerOutput, imageName) {
			fmt.Println("docker image " + imageName + " was found")
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

	if labelsMap["appsody.stack"] == nil {
		t.Error("Could not find stack label in Docker image!")
	}

	if labelsMap["appsody.requested-stack"] == nil {
		t.Error("Could not find requested stack label in Docker image!")
	}

	if labelsMap["org.opencontainers.image.created"] == nil || labelsMap["org.opencontainers.image.version"] == nil || labelsMap["org.opencontainers.image.revision"] == nil {
		t.Error("Could not find opencontainers image labels in Docker image!")
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
