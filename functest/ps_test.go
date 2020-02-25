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
	"strings"
	"time"

	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestPS(t *testing.T) {
	//
	// Test Plan:
	//
	// - Spin up a named conatainer using appsody run (any one will do)
	// - Use docker ps to wait until it is ready
	// - execute 'appsody ps', and check it we get at least a header line and the
	//   right container in the output
	// - use 'appsody stop' to stop the container
	//

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
	containerName := "testPSContainer"
	go func() {
		args = []string{"run", "--name", containerName}
		_, err = cmdtest.RunAppsody(sandbox, args...)
		runChannel <- err
		close(runChannel)
	}()

	defer func() {
		// run appsody stop to close the docker container
		args = []string{"stop", "--name", containerName}
		_, err = cmdtest.RunAppsody(sandbox, args...)
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
	count := 15 // wait 30 seconds
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

	// now run appsody ps and see if we can spot the container
	t.Log("about to run appsody ps")
	stopOutput, errStop := cmdtest.RunAppsody(sandbox, "ps")
	if !strings.Contains(stopOutput, "CONTAINER") {
		t.Fatal("output doesn't contain header line")
	}
	if !strings.Contains(stopOutput, containerName) {
		t.Fatal("output doesn't contain correct container name")
	}
	if errStop != nil {
		t.Logf("Ignoring error running appsody ps: %s", errStop)
	}
}

func TestPsNoContainers(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody ps with no running containers
	args := []string{"ps"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	} else {
		if !strings.Contains(output, "There are no stack-based containers running in your docker environment") {
			t.Fatalf("String \"There are no stack-based containers running in your docker environment\" not found in output: %v", output)
		}
	}
}

func TestPsArgumentFail(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody ps with extra arguments
	args := []string{"ps", "testing"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		if !strings.Contains(output, "Unexpected argument") {
			t.Fatalf("String \"Unexpected argument\" not found in error: %v", err)
		}
	} else {
		t.Fatalf("Appsody ps passed with an argument: %v", output)
	}
}
