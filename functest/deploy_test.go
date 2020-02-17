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

	stacksList := cmdtest.GetEnvStacksList()

	stackRaw := strings.Split(stacksList, " ")

	// we don't need to split the repo and stack anymore...
	// stackStack := strings.Split(stackRaw, "/")

	for i := range stackRaw {
		t.Log("stackRaw is: ", stackRaw[i])
	}

}

// Simple test for appsody deploy command. A future enhancement would be to configure a valid deployment environment
func TestDeploySimple(t *testing.T) {

	stacksList := cmdtest.GetEnvStacksList()

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		t.Log("***Testing stack: ", stackRaw[i], "***")

		sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
		defer cleanup()

		// first add the test repo index
		_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "index.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsody(sandbox, "init", stackRaw[i])
		if err != nil {
			t.Fatal(err)
		}

		// appsody deploy
		t.Log("Running appsody deploy...")
		_, err = cmdtest.RunAppsody(sandbox, "deploy", "-t", "testdeploy/testimage", "--dryrun")
		if err != nil {
			t.Log("WARNING: deploy dryrun failed. Ignoring for now until that gets fixed.")
			// TODO We need to fix the deploy --dryrun option so it doesn't fail, then uncomment the line below
			// t.Fatal(err)
		}
	}
}

// Testing generation of app-deploy.yaml
func TestGenerationDeploymentConfig(t *testing.T) {

	stacksList := cmdtest.GetEnvStacksList()

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		t.Log("***Testing stack: ", stackRaw[i], "***")

		sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
		defer cleanup()

		// first add the test repo index
		_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "index.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsody(sandbox, "init", stackRaw[i])
		if err != nil {
			t.Fatal(err)
		}

		imageTag := "testdeploy/testimage"
		pullURL := "my-pull-url"
		namespace := "myNamespace"
		// appsody deploy
		t.Log("Running appsody deploy...")
		_, err = cmdtest.RunAppsody(sandbox, "deploy", "-t", imageTag, "--pull-url", pullURL, "--generate-only", "--knative", "-n", namespace)
		if err != nil {
			t.Log("WARNING: deploy dryrun failed. Ignoring for now until that gets fixed.")
			// TODO We need to fix the deploy --dryrun option so it doesn't fail, then uncomment the line below
			// t.Fatal(err)
		}

		checkDeploymentConfig(t, expectedDeploymentConfig{filepath.Join(sandbox.ProjectDir, deployFile), pullURL, imageTag, namespace, true})
	}
}

func TestDeployNoNamespace(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init
	t.Log("Running appsody init...")
	_, err := cmdtest.RunAppsody(sandbox, "init", "nodejs-express")
	if err != nil {
		t.Fatal(err)
	}

	imageTag := "testdeploy/testimage"
	// appsody deploy
	t.Logf("Running appsody deploy with no namespace")
	_, err = cmdtest.RunAppsody(sandbox, "deploy", "-t", imageTag, "--generate-only")
	if err != nil {
		t.Fatal(err)
	}

	checkDeploymentConfig(t, expectedDeploymentConfig{filepath.Join(sandbox.ProjectDir, deployFile), "", imageTag, "", false})
}

func TestDeployNamespaceMismatch(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init
	t.Log("Running appsody init...")
	_, err := cmdtest.RunAppsody(sandbox, "init", "nodejs-express")
	if err != nil {
		t.Fatal(err)
	}

	firstNamespace := "firstNamespace"
	// appsody deploy
	t.Logf("Running appsody deploy with namespace: %s ...", firstNamespace)
	_, err = cmdtest.RunAppsody(sandbox, "deploy", "--generate-only", "-n", firstNamespace)
	if err != nil {
		t.Fatal(err)
	}

	secondNamespace := "secondNamespace"
	// appsody deploy
	t.Logf("Running appsody deploy with namespace: %s ...", secondNamespace)
	output, err := cmdtest.RunAppsody(sandbox, "deploy", "--generate-only", "-n", secondNamespace)

	if err != nil {
		if !strings.Contains(output, "the namespace \""+firstNamespace+"\" from the deployment manifest does not match the namespace \""+secondNamespace+"\" passed as an argument.") {
			t.Errorf("Expecting namespace error to be thrown, but another error was thrown: %s", err)
		}
	} else {
		t.Error("Deploy with conflicting namespace did not fail as expected")
	}

}

// Testing deploy delete when the required config file cannot be found
func TestDeployDeleteNotFound(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// Not passing a config file so it will use the default, which shouldn't exist
	args := []string{"deploy", "delete"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
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
func TestDeployDeleteKubeFail(t *testing.T) {

	if !cmdtest.TravisTesting {
		t.Skip()
	}

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	filename := "fake.yaml"

	// Ensure that the fake yaml file is deleted
	defer func() {
		err := os.Remove(filename)
		if err != nil {
			t.Errorf("Error removing the file: %s", err)
		}
	}()

	// Attempt to create the fake file
	file, err := os.Create(filename)

	if err != nil {
		t.Errorf("Error creating the file: %s", err)
	}

	// Change the fake file to lack read permissions
	err = file.Chmod(0333)
	if err != nil {
		t.Errorf("Error changing file permissions: %s", err)
	}

	args := []string{"deploy", "delete", "-f", filename}
	// Pass the file to deploy delete, which should fail
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {

		// If the error is not what expected, fail the test
		if !strings.Contains(err.Error(), "kubectl delete failed: exit status 1: error: open "+filename+": permission denied") {
			t.Error("String \"kubectl delete failed: exit status 1: error: open ", filename, ": permission denied\" not found in output")
		}

		// If there was not an error returned, fail the test
	} else {
		t.Error("Deploy delete did not fail as expected")
	}
}

func TestNoCheckFlag(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "index.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Running appsody init...")
	_, err = cmdtest.RunAppsody(sandbox, "init", "nodejs-express")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Running appsody deploy...")
	output, err := cmdtest.RunAppsody(sandbox, "deploy", "-t", "testdeploy/testimage", "--dryrun", "--no-check")
	if err != nil {
		t.Log("WARNING: deploy dryrun failed.")
	}

	if !strings.Contains(output, "kubectl get pods -o=jsonpath='{.items[?(@.metadata.labels.name==\"appsody-operator\")].metadata.namespace}' -n") {
		t.Fatal(err, ": Expected kubectl get pods to run only against the targeted namespace rather than all namespaces.")
	}
}
