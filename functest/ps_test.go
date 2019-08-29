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
	"fmt"
	"io/ioutil"
	"log"
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
	projectDir, err := ioutil.TempDir("", "appsody-ps-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express"}, projectDir)
	if err != nil {
		t.Fatal(err)
	}

	// appsody run
	runChannel := make(chan error)
	containerName := "testPSContainer"
	go func() {
		_, err = cmdtest.RunAppsodyCmdExec([]string{"run", "--name", containerName}, projectDir)
		runChannel <- err
	}()

	// It will take a while for the container to spin up, so let's use docker ps to wait for it
	fmt.Println("calling docker ps to wait for container")
	containerRunning := false
	count := 100
	for {
		dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"ps", "-q", "-f", "name=" + containerName})
		if dockerErr != nil {
			log.Print("Ignoring error running docker ps -q -f name="+containerName, dockerErr)
		}
		if dockerOutput != "" {
			fmt.Println("docker container " + containerName + " was found")
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
	fmt.Println("about to run appsody ps")
	stopOutput, errStop := cmdtest.RunAppsodyCmd([]string{"ps"}, projectDir)
	if !strings.Contains(stopOutput, "CONTAINER") {
		t.Fatal("output doesn't contain header line")
	}
	if !strings.Contains(stopOutput, containerName) {
		t.Fatal("output doesn't contain correct container name")
	}
	if errStop != nil {
		log.Printf("Ignoring error running appsody ps: %s", errStop)
	}

	// defer the appsody stop to close the docker container
	defer func() {
		_, err = cmdtest.RunAppsodyCmdExec([]string{"stop", "--name", "testPSContainer"}, projectDir)
		if err != nil {
			fmt.Printf("Ignoring error running appsody stop: %s", err)
		}
	}()
}
