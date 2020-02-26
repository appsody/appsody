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
	"runtime"
	"strings"
	"testing"

	"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
	"sigs.k8s.io/yaml"
)

type expectedDeploymentConfig struct {
	deployFile string
	pullURL    string
	imageTag   string
	namespace  string
	knative    bool
}

func TestSimpleBuildCases(t *testing.T) {
	var buildSimpleTests = []struct {
		testName string
		cmdName  string
		args     []string // input
	}{
		{"Test simple build Docker", "docker", []string{"build"}},
	}
	for _, testData := range buildSimpleTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {

			if tt.cmdName == "buildah" {
				if runtime.GOOS != "linux" {
					t.Skip()
				}
			}
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

				// appsody build
				_, err = cmdtest.RunAppsody(sandbox, tt.args...)
				if err != nil {
					t.Fatal("The appsody build command failed: ", err)
				}

				expectedImageTag := "dev.local/" + sandbox.ProjectName
				expectedImageTag = strings.Replace(expectedImageTag, "_", "-", -1)
				listOutput, listErr := cmdtest.RunCmdExec(tt.cmdName, []string{"images", "-q", expectedImageTag}, t)
				if listErr != nil {
					t.Fatal(listErr)
				}
				if listOutput == "" {
					t.Errorf("Expected appsody build to create image '%s' but it was not found.", expectedImageTag)
				}
				//delete the image
				deleteImage(expectedImageTag, tt.cmdName, t)
			}

		})
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

var appsodyCommitKey = "dev.appsody.image.commit."
var appsodyCommitLabels = []string{
	"message",
	"date",
	"committer",
	"author",
}

func TestBuildLabels(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, false)
	defer cleanup()

	// first add the test repo index
	_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "index.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// appsody init
	_, err = cmdtest.RunAppsody(sandbox, "init", "nodejs")
	t.Log("Running appsody init...")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Copying .appsody-config.yaml to project dir...")
	copyCmd := exec.Command("cp", filepath.Join(sandbox.TestDataPath, ".appsody-config.yaml"), sandbox.ProjectDir)
	err = copyCmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	commitMessage := "initial test commit"
	t.Log("Setting up git for ", sandbox.ProjectDir)
	gitCmd := exec.Command("sh", "-c", "git init && git add . && git commit -m '"+commitMessage+"' && git remote add upstream url && git branch upstream && git branch -u upstream")
	gitCmd.Dir = sandbox.ProjectDir
	err = gitCmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// appsody build
	imageName := "testbuildimage"
	_, err = cmdtest.RunAppsody(sandbox, "build", "--tag", imageName)
	if err != nil {
		t.Fatalf("Error on appsody build: %v", err)
	}

	inspectOutput, inspectErr := cmdtest.RunCmdExec("docker", []string{"inspect", imageName}, t)
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

	for _, label := range appsodyCommitLabels {
		if labelsMap[appsodyCommitKey+label] == nil {
			t.Errorf("Could not find %s%s label in Docker image!", appsodyCommitKey, label)
		}
	}

	if labelsMap[appsodyCommitKey+"message"] != commitMessage {
		t.Errorf("Expected commit message \"%s\" but found \"%s\"", commitMessage, labelsMap[appsodyCommitKey+"message"])
	}

	checkDeploymentConfig(t, expectedDeploymentConfig{filepath.Join(sandbox.ProjectDir, deployFile), "", imageName, "", false})

	//delete the image
	deleteImage(imageName, "docker", t)
}

func deleteImage(imageName string, cmdName string, t *testing.T) {
	_, err := cmdtest.RunCmdExec(cmdName, []string{"image", "rm", imageName}, t)
	if err != nil {
		t.Logf("Ignoring error running docker image rm: %s", err)
	}
}

func TestDeploymentConfig(t *testing.T) {

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

		// appsody build
		imageName := sandbox.ProjectName
		pullURL := "my-pull-url"

		_, err = cmdtest.RunAppsody(sandbox, "build", "--tag", imageName, "--pull-url", pullURL, "--knative")
		if err != nil {
			t.Error("appsody build command returned err: ", err)
		}

		checkDeploymentConfig(t, expectedDeploymentConfig{filepath.Join(sandbox.ProjectDir, deployFile), pullURL, imageName, "", true})

		//delete the image
		deleteImage(imageName, "docker", t)
	}
}

// app-deploy

var knativeFlagTests = []struct {
	testName          string
	knativeFlag       string
	appDeployStart    bool
	appDeployExpected bool
}{
	{"KnativeFlagAndAppDeployTrue", "--knative", true, true},
	{"KnativeFlagAndAppDeployFalse", "--knative", false, true},
	{"NoKnativeFlagAndAppDeployTrue", "", true, true},
	{"NoKnativeFlagAndAppDeployFalse", "", false, false},
	{"KnativeFalseAndAppDeployTrue", "--knative=false", true, false},
	{"KnativeFalseAndAppDeployFalse", "--knative=false", false, false},
	{"KnativeTrueAndAppDeployTrue", "--knative=true", true, true},
	{"KnativeTrueAndAppDeployFalse", "--knative=true", false, true},
}

func TestKnativeFlagOnBuild(t *testing.T) {

	stacksList := cmdtest.GetEnvStacksList()

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {
		for _, testData := range knativeFlagTests {
			// need to set testData to a new variable scoped under the for loop
			// otherwise tests run in parallel may get the wrong testData
			// because the for loop reassigns it before the func runs
			tt := testData

			t.Run(tt.testName, func(t *testing.T) {
				t.Log("***Testing stack: ", stackRaw[i], "***")
				sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
				defer cleanup()

				// appsody init
				t.Log("Running appsody init...")
				_, err := cmdtest.RunAppsody(sandbox, "init", stackRaw[i])
				if err != nil {
					t.Fatal(err)
				}

				err = makeAppDeployYaml(sandbox.ProjectDir, tt.appDeployStart)
				if err != nil {
					t.Fatal(err)
				}

				// appsody build
				if tt.knativeFlag == "" {
					_, err = cmdtest.RunAppsody(sandbox, "build")
				} else {
					_, err = cmdtest.RunAppsody(sandbox, "build", tt.knativeFlag)
				}
				if err != nil {
					t.Error("appsody build command returned err: ", err)
				}
				expectedImageName := "dev.local/" + sandbox.ProjectName
				checkDeploymentConfig(t, expectedDeploymentConfig{filepath.Join(sandbox.ProjectDir, deployFile), "", expectedImageName, "", tt.appDeployExpected})

				//delete the image
				deleteImage(expectedImageName, "docker", t)
			})
		}
	}
}

func makeAppDeployYaml(projectDir string, createKnativeService bool) error {
	appsodyApplication := v1beta1.AppsodyApplication{}
	appsodyApplication.Spec.CreateKnativeService = &createKnativeService
	data, err := yaml.Marshal(appsodyApplication)
	if err != nil {
		return fmt.Errorf("error marshalling yaml: %v", err)
	}

	// write to file
	deployFilePath := filepath.Join(projectDir, deployFile)
	err = ioutil.WriteFile(deployFilePath, data, 0666)
	if err != nil {
		return fmt.Errorf("error writing deployment yaml to file %s: %v", deployFilePath, err)
	}
	return nil
}

func checkDeploymentConfig(t *testing.T, expectedDeploymentConfig expectedDeploymentConfig) {
	deployFile := expectedDeploymentConfig.deployFile
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

	expectedApplicationImage := expectedDeploymentConfig.imageTag
	if expectedDeploymentConfig.pullURL != "" {
		expectedApplicationImage = expectedDeploymentConfig.pullURL + "/" + expectedDeploymentConfig.imageTag
	}

	if appsodyApplication.Spec.ApplicationImage != expectedApplicationImage {
		t.Errorf("Incorrect ApplicationImage in app-deploy.yaml. Expected %s but found %s", expectedApplicationImage, appsodyApplication.Spec.ApplicationImage)
	}

	if *appsodyApplication.Spec.CreateKnativeService != expectedDeploymentConfig.knative {
		t.Error("CreateKnativeService not set to true in the app-deploy.yaml when using --knative flag")
	}

	if appsodyApplication.Namespace != expectedDeploymentConfig.namespace {
		t.Errorf("Incorrect Namespace in app-deploy.yaml. Expected %s but found %s", expectedDeploymentConfig.namespace, appsodyApplication.Namespace)
	}

	verifyImageAndConfigLabelsMatch(t, appsodyApplication, expectedDeploymentConfig.imageTag)
}

func verifyImageAndConfigLabelsMatch(t *testing.T, appsodyApplication v1beta1.AppsodyApplication, imageTag string) {
	args := []string{"inspect", "--format='{{json .Config.Labels }}'", imageTag}
	output, err := cmdtest.RunCmdExec("docker", args, t)
	if err != nil {
		t.Errorf("Error inspecting docker image: %s", err)
	}
	output = strings.Trim(output, "\n'")

	var imageLabels map[string]string
	err = json.Unmarshal([]byte(output), &imageLabels)
	if err != nil {
		t.Errorf("Could not unmarshall docker labels: %s", err)
	}

	for key, value := range imageLabels {
		key, err = cmd.ConvertLabelToKubeFormat(key)
		if key == "app.appsody.dev/name" {
			key = "app.kubernetes.io/part-of"
		}
		if err != nil {
			t.Errorf("Could not convert label to Kubernetes format: %s", err)
		}

		label := appsodyApplication.Labels[key]
		annotation := appsodyApplication.Annotations[key]
		if label == "" && annotation == "" {
			t.Errorf("Could not find label %s in deployment config", key)
		}

		if label != "" && label != value {
			t.Errorf("Mismatch of %s label between built image and deployment config. Expected %s but found %s", key, value, label)
		}

		if annotation != "" && annotation != value {
			t.Errorf("Mismatch of %s annotation between built image and deployment config. Expected %s but found %s", key, value, annotation)
		}
	}

}

func TestBuildMissingTagFail(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init
	t.Log("Running appsody init...")
	_, err := cmdtest.RunAppsody(sandbox, "init", "starter")
	if err != nil {
		t.Fatal(err)
	}

	// set push flag to true with no tag
	args := []string{"build", "--push"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {

		// As tag is missing, appsody verifies user input and shows error
		if !strings.Contains(output, "Cannot specify --push or --push-url without a --tag") {
			t.Errorf("String \"Cannot specify --push or --push-url without a --tag\" not found in output: %v", err)
		}

		// If an error is not returned, the test should fail
	} else {
		t.Error("Build with missing tag did not fail as expected")
	}

}
