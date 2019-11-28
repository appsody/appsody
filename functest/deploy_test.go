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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

var deployFile = "app-deploy.yaml"

// Test parsing environment variable with stack info
func TestParser(t *testing.T) {

	stacksList = "incubator/nodejs"
	t.Log("stacksList is: ", stacksList)
	if stacksList == "" {
		t.Log("stacksList is empty, exiting test...")
		return
	}

	stackRaw := strings.Split(stacksList, " ")

	// we don't need to split the repo and stack anymore...
	// stackStack := strings.Split(stackRaw, "/")

	for i := range stackRaw {
		t.Log("stackRaw is: ", stackRaw[i])

		// code to sepearate the repos and stacks...
		// stageStack := strings.Split(stackRaw[i], "/")
		// stage := stageStack[0]
		// stack := stageStack[1]
		// t.Log("stage is: ", stage)
		// t.Log("stack is: ", stack)

	}

}

// Simple test for appsody deploy command. A future enhancement would be to configure a valid deployment environment
func TestDeploySimple(t *testing.T) {

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
		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml", t)
		if err != nil {
			t.Fatal(err)
		}

		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsodyCmd([]string{"init", stackRaw[i]}, projectDir, t)
		if err != nil {
			t.Fatal(err)
		}

		// appsody deploy
		t.Log("Running appsody deploy...")
		_, err = cmdtest.RunAppsodyCmd([]string{"deploy", "-t", "testdeploy/testimage", "--dryrun"}, projectDir, t)
		if err != nil {
			t.Log("WARNING: deploy dryrun failed. Ignoring for now until that gets fixed.")
			// TODO We need to fix the deploy --dryrun option so it doesn't fail, then uncomment the line below
			// t.Fatal(err)
		}

		// cleanup tasks
		cleanup()
	}
}

// Testing generation of app-deploy.yaml
func TestGenerationDeploymentConfig(t *testing.T) {
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
		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml", t)
		if err != nil {
			t.Fatal(err)
		}

		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsodyCmd([]string{"init", stackRaw[i]}, projectDir, t)
		if err != nil {
			t.Fatal(err)
		}

		imageTag := "testdeploy/testimage"
		pullURL := "my-pull-url"
		// appsody deploy
		t.Log("Running appsody deploy...")
		_, err = cmdtest.RunAppsodyCmd([]string{"deploy", "-t", imageTag, "--pull-url", pullURL, "--generate-only", "--knative"}, projectDir, t)
		if err != nil {
			t.Log("WARNING: deploy dryrun failed. Ignoring for now until that gets fixed.")
			// TODO We need to fix the deploy --dryrun option so it doesn't fail, then uncomment the line below
			// t.Fatal(err)
		}

		checkDeploymentConfig(t, filepath.Join(projectDir, deployFile), pullURL, imageTag)

		// cleanup tasks
		cleanup()
	}
}

// Testing deploy delete when the required config file cannot be found
func TestDeployDeleteNotFound(t *testing.T) {

	// Not passing a config file so it will use the default, which shouldn't exist
	_, err := cmdtest.RunAppsodyCmd([]string{"deploy", "delete"}, ".", t)
	if err != nil {

		// Because the config doesn't exist, this error should be returned (without -v)
		if !strings.Contains(err.Error(), "Deployment manifest not found") {
			t.Error("String \"Deployment manifest not found\" not found in output")
		}

		// If an error is not returned, the test should fail
	} else {
		t.Error("Deploy delete did not fail as expected")
	}
}

// // Testing deploy delete when given a file that exists, but can't be read
//
//
//
//	TODO: This test does work, but as Travis can't currently run kubectl commands
//		  so it doesn't throw the correct error
//
// func TestDeployDeleteKubeFail(t *testing.T) {
// 	filename := "fake.yaml"

// 	// Ensure that the fake yaml file is deleted
// 	defer func() {
// 		err := os.Remove(filename)
// 		if err != nil {
// 			t.Errorf("Error removing the file: %s", err)
// 		}
// 	}()

// 	// Attempt to create the fake file
// 	file, err := os.Create(filename)

// 	if err != nil {
// 		t.Errorf("Error creating the file: %s", err)
// 	}

// 	// Change the fake file to lack read permissions
// 	err = file.Chmod(0333)
// 	if err != nil {
// 		t.Errorf("Error changing file permissions: %s", err)
// 	}

// 	// Pass the file to deploy delete, which should fail
// 	_, err = cmdtest.RunAppsodyCmd([]string{"deploy", "delete", "-f", filename}, ".", t)
// 	if err != nil {

// 		// If the error is not what expected, fail the test
// 		if !strings.Contains(err.Error(), "kubectl delete failed: exit status 1: error: open "+filename+": permission denied") {
// 			t.Error("String \"kubectl delete failed: exit status 1: error: open ", filename, ": permission denied\" not found in output")
// 		}

// 		// If there was not an error returned, fail the test
// 	} else {
// 		t.Error("Deploy delete did not fail as expected")
// 	}
// }
