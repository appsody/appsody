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
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

// Test appsody run of the nodejs-express stack and check the http://localhost:3000/health endpoint
func TestRun(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, false)
	defer cleanup()

	// first add the test repo index
	_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "dev.local-index.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	stacksList := cmdtest.GetEnvStacksList()

	if stacksList == "dev.local/starter" {
		t.Skip()
	}

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsody(sandbox, "init", "nodejs-express")
	if err != nil {
		t.Fatal(err)
	}

	// appsody run
	runChannel := make(chan error)
	go func() {
		_, err = cmdtest.RunAppsody(sandbox, "run")
		runChannel <- err
		close(runChannel)
	}()

	// defer the appsody stop to close the docker container
	defer func() {
		_, err = cmdtest.RunAppsody(sandbox, "stop")
		if err != nil {
			t.Logf("Ignoring error running appsody stop: %s", err)
		}
		// wait for the appsody command/goroutine to finish
		runErr := <-runChannel
		if runErr != nil {
			t.Logf("Ignoring error from the appsody command: %s", runErr)
		}
	}()

	healthCheckFrequency := 2 // in seconds
	healthCheckTimeout := 60  // in seconds
	healthCheckWait := 0
	healthCheckOK := false
	for !(healthCheckOK || healthCheckWait >= healthCheckTimeout) {
		select {
		case err = <-runChannel:
			// appsody run exited, probably with an error
			t.Fatalf("appsody run quit unexpectedly: %s", err)
		case <-time.After(time.Duration(healthCheckFrequency) * time.Second):
			// check the health endpoint
			healthCheckWait += healthCheckFrequency
			resp, err := http.Get("http://localhost:3000/health")
			if err != nil {
				t.Logf("Health check error. Ignore and retry: %s", err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != 200 {
					t.Logf("Health check response code %d. Ignore and retry.", resp.StatusCode)
				} else {
					t.Logf("Health check OK")
					// may want to check body
					healthCheckOK = true
				}
			}
		}
	}

	if !healthCheckOK {
		t.Errorf("Did not receive an OK health check within %d seconds.", healthCheckTimeout)
	}
}

// Simple test for appsody run command. A future enhancement would be to verify the endpoint or console output if there is no web endpoint
func TestRunSimple(t *testing.T) {

	stacksList := cmdtest.GetEnvStacksList()

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		t.Log("***Testing stack: ", stackRaw[i], "***")

		sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
		defer cleanup()

		// z and p use locally packaged dev.local so we need to add it to the config of the sandbox for it to work
		cmdtest.ZAndPDevLocal(t, sandbox)

		// first add the test repo index
		_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "dev.local-index.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsody(sandbox, "init", stackRaw[i])
		if err != nil {
			t.Fatal(err)
		}

		// appsody run
		runChannel := make(chan error)
		containerName := "testRunSimpleContainer"
		go func() {
			_, err = cmdtest.RunAppsody(sandbox, "run", "--name", containerName)
			runChannel <- err
			close(runChannel)
		}()

		// defer the appsody stop to close the docker container
		defer func() {
			_, err = cmdtest.RunAppsody(sandbox, "stop", "--name", containerName)
			if err != nil {
				t.Logf("Ignoring error running appsody stop: %s", err)
			}
			// wait for the appsody command/goroutine to finish
			runErr := <-runChannel
			if runErr != nil {
				t.Logf("Ignoring error from the appsody command: %s", runErr)
			}
		}()

		// It will take a while for the container to spin up, so let's use docker ps to wait for it
		t.Log("calling docker ps to wait for container")
		containerRunning := false
		count := 100
		for {
			dockerOutput, dockerErr := cmdtest.RunCmdExec("docker", []string{"ps", "-q", "-f", "name=" + containerName}, t)
			if dockerErr != nil {
				t.Log("Ignoring error running docker ps -q -f name="+containerName, dockerErr)
			}
			if dockerOutput != "" {
				t.Log("docker container " + containerName + " was found")
				containerRunning = true
			} else {
				time.Sleep(2 * time.Second)
				count = count - 1
			}
			if count == 0 || containerRunning {
				break
			}
		}

		if !containerRunning {
			t.Fatal("container never appeared to start")
		}

	}
}

func TestRunTooManyArgs(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"run", "too", "many", "args"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err == nil {
		t.Error("Expected non-zero exit code")
	}
	if !strings.Contains(output, "Unexpected argument.") {
		t.Error("Failed to flag too many arguments.")
	}
}

//check appsody run is using the same volumes as in project.yaml
func TestRunUsesCorrectProjectVolumes(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	args = []string{"run", "--dryrun"}

	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	config := new(cmd.RootCommandConfig)
	_, project, _ := getCurrentProjectEntry(t, sandbox, config)

	depsMount := project.Volumes[0].Name + ":" + project.Volumes[0].Path

	if !strings.Contains(output, depsMount) {
		t.Fatalf("Did not find expected docker volume mount in run command output: %s", depsMount)
	}
}

// check project entry path in project.yaml gets updated when project moves
func TestRunUpdatesProjectPath(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	args = []string{"run", "--dryrun"}

	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	config := new(cmd.RootCommandConfig)

	p, _, _ := getCurrentProjectEntry(t, sandbox, config)
	projectsBefore := len(p.Projects)

	tmpDir := filepath.Join(sandbox.TestDataPath, "tmp")
	err = os.Rename(sandbox.ProjectDir, tmpDir)
	if err != nil {
		log.Fatal(err)
	}
	sandbox.ProjectDir = tmpDir

	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	p, project, _ := getCurrentProjectEntry(t, sandbox, config)
	projectsAfter := len(p.Projects)

	if project.Path != tmpDir {
		t.Fatalf("Expected project entry path to be updated to %s but found %s", tmpDir, project.Path)
	}
	if projectsBefore != projectsAfter {
		t.Fatalf("Expected number of project entries to be %v but found %v", projectsBefore, projectsAfter)
	}
}

// check if id exists in .appsody-config.yaml but not in project.yaml, a new project entry in project.yaml gets created with the same id
func TestRunIfProjectIDNotExistInProjectYaml(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)
	config := &cmd.RootCommandConfig{LoggingConfig: loggingConfig}

	p, _, _ := getCurrentProjectEntry(t, sandbox, config)
	projectsBefore := len(p.Projects)

	err = cmd.SaveIDToConfig("newRandomID", config)
	if err != nil {
		t.Fatal(err)
	}
	args = []string{"run", "--dryrun"}
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	p, project, configID := getCurrentProjectEntry(t, sandbox, config)
	projectsAfter := len(p.Projects)

	if projectsBefore+1 != projectsAfter {
		t.Fatalf("Expected number of project entries to be %v but found %v", projectsBefore+1, projectsAfter)
	}
	if project.ID != configID {
		t.Fatalf("Expected project id in .appsody-config.yaml to have a valid project entry in project.yaml.")
	}
}

// check if id does not exists in .appsody-config.yaml, a new project entry in project.yaml gets created with the same id
func TestRunIfProjectIDNotExistInConfigYaml(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	config := new(cmd.RootCommandConfig)

	p, _, configID := getCurrentProjectEntry(t, sandbox, config)
	projectsBefore := len(p.Projects)

	// delete id from .appsody-config.yaml
	appsodyConfig := filepath.Join(sandbox.ProjectDir, cmd.ConfigFile)
	data, err := ioutil.ReadFile(appsodyConfig)
	if err != nil {
		t.Fatal(err)
	}
	removedID := bytes.Replace(data, []byte("id: \""+configID+"\""), []byte(""), 1)
	err = ioutil.WriteFile(appsodyConfig, []byte(removedID), 0666)
	if err != nil {
		t.Fatal(err)
	}

	args = []string{"run", "--dryrun"}
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	p, project, configID := getCurrentProjectEntry(t, sandbox, config)
	projectsAfter := len(p.Projects)

	if projectsBefore+1 != projectsAfter {
		t.Fatalf("Expected number of project entries to be %v but found %v", projectsBefore+1, projectsAfter)
	}
	if project.ID != configID {
		t.Fatalf("Expected project id in .appsody-config.yaml to have a valid project entry in project.yaml.")
	}
}

// check error if user specified mount interferes with stack mounts
func TestRunUserSpecifiedVolumesStack(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	userSpecifiedMount := "volume:/project/user-app/node_modules"
	userSpecifiedMountSplit := strings.Split(userSpecifiedMount, ":")
	args = []string{"run", "--docker-options", "-v " + userSpecifiedMount, "--dryrun"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err == nil {
		t.Fatal("Expected non-zero exit code")
	}

	expectedError := "User specified mount path " + userSpecifiedMountSplit[1] + " is not allowed in --docker-options, as it interferes with the stack specified mount path /project/user-app/node_modules"
	if !strings.Contains(output, expectedError) {
		t.Fatalf("Expected error not found: %s", expectedError)
	}
}

// check error if user specified mount interferes with stack mounts
func TestRunUserSpecifiedVolumesDefault(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	userSpecifiedMount := "volume:/project/user-app"
	userSpecifiedMountSplit := strings.Split(userSpecifiedMount, ":")
	args = []string{"run", "--docker-options", "-v " + userSpecifiedMount, "--dryrun"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err == nil {
		t.Fatal("Expected non-zero exit code")
	}

	expectedError := "User specified mount path " + userSpecifiedMountSplit[1] + " is not allowed in --docker-options, as it interferes with the default specified mount path /project/user-app"
	if !strings.Contains(output, expectedError) {
		t.Fatalf("Expected error not found: %s", expectedError)
	}
}

// check that user specified mount doesnt interferes with stack mounts
func TestRunUserSpecifiedVolumesSimilar(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	userSpecifiedMount := "volume:/project/user-app/node_modules2"
	userSpecifiedMountSplit := strings.Split(userSpecifiedMount, ":")
	args = []string{"run", "--docker-options", "-v " + userSpecifiedMount, "--dryrun"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		unexpectedError := "User specified mount path " + userSpecifiedMountSplit[1] + " is not allowed in --docker-options, as it interferes with the stack specified mount path /project/user-app/node_modules2"
		if strings.Contains(output, unexpectedError) {
			t.Fatalf("Unexpected error found: %s", unexpectedError)
		} else {
			t.Fatal("Expected zero exit code")
		}
	}
}
