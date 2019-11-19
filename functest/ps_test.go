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

	"os"
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
	projectDir := cmdtest.GetTempProjectDir(t)
	defer os.RemoveAll(projectDir)
	t.Log("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err := cmdtest.RunAppsodyCmd([]string{"init", "nodejs-express"}, projectDir, t)
	if err != nil {
		t.Fatal(err)
	}

	// appsody run
	runChannel := make(chan error)
	containerName := "testPSContainer"
	go func() {
		_, err = cmdtest.RunAppsodyCmd([]string{"run", "--name", containerName}, projectDir, t)
		runChannel <- err
	}()

	// It will take a while for the container to spin up, so let's use docker ps to wait for it
	t.Log("calling docker ps to wait for container")
	containerRunning := false
	count := 15 // wait 30 seconds
	for {
		dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"ps", "-q", "-f", "name=" + containerName}, t)
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
	stopOutput, errStop := cmdtest.RunAppsodyCmd([]string{"ps"}, projectDir, t)
	if !strings.Contains(stopOutput, "CONTAINER") {
		t.Fatal("output doesn't contain header line")
	}
	if !strings.Contains(stopOutput, containerName) {
		t.Fatal("output doesn't contain correct container name")
	}
	if errStop != nil {
		t.Logf("Ignoring error running appsody ps: %s", errStop)
	}

	// defer the appsody stop to close the docker container
	defer func() {
		_, err = cmdtest.RunAppsodyCmd([]string{"stop", "--name", "testPSContainer"}, projectDir, t)
		if err != nil {
			t.Logf("Ignoring error running appsody stop: %s", err)
		}
	}()
}
