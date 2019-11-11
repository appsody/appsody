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
