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
	"testing"
	"time"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// Test appsody run of the nodejs-express stack and check the http://localhost:3000/health endpoint
func TestRun(t *testing.T) {
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

		err = cmdtest.RunDockerPs(t, 100, containerName)
		if err != nil {
			t.Fatal(err)
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
