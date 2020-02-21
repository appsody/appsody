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
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestStopWithoutName(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, false)
	defer cleanup()

	// first add the test repo index
	_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "index.yaml"))
	if err != nil {
		t.Fatal(err)
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
		t.Log("calling docker stop")
		stopOutput, errStop := cmdtest.RunAppsody(sandbox, "stop")

		//docker stop appsody-stop-test
		if !strings.Contains(stopOutput, "docker stop "+sandbox.ProjectName) {
			t.Fatal("docker stop command not present for appsody-test...")
		}
		if errStop != nil {
			t.Logf("Ignoring error running appsody stop: %s", errStop)

		}
		t.Log("calling docker ps")
		pathElements := strings.Split(sandbox.ProjectDir, "/")
		containerName := pathElements[len(pathElements)-1]
		dockerOutput, dockerErr := cmdtest.RunCmdExec("docker", []string{"ps", "-q", "-f", "name=" + containerName + "-dev"}, t)
		t.Log("docker output", dockerOutput)
		if dockerErr != nil {
			t.Log("Ignoring error running docker ps -q -f name=appsody-stop-test-dev", dockerErr)

		}
		if dockerOutput != "" {
			t.Fatal("docker container " + containerName + " was found and should have been stopped")

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

func TestStopWithName(t *testing.T) {

	// create a temporary dir to create the project and run the test
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, false)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	// appsody run
	runChannel := make(chan error)
	go func() {
		args = []string{"run", "--name", "testStopContainer"}
		_, err = cmdtest.RunAppsody(sandbox, args...)
		runChannel <- err
		close(runChannel)
	}()

	// defer the appsody stop to close the docker container

	defer func() {
		t.Log("about to run stop for with name")
		args = []string{"stop", "--name", "testStopContainer"}
		stopOutput, errStop := cmdtest.RunAppsody(sandbox, args...)
		if !strings.Contains(stopOutput, "docker stop testStopContainer") {
			t.Fatal("docker stop command not present for container testStopContainer")
		}
		if errStop != nil {
			t.Logf("Ignoring error running appsody stop: %s", errStop)

		}
		t.Log("about to do docker ps")
		dockerOutput, dockerErr := cmdtest.RunCmdExec("docker", []string{"ps", "-q", "-f", "name=testStopContainer"}, t)
		if dockerErr != nil {
			t.Log("Ignoring error running docker ps -q -f name=testStopContainer", dockerErr)

		}
		if dockerOutput != "" {
			t.Fatal("docker container testStopContainer was found and should have been stopped")

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
