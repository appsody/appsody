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

package cmd

import (
	"fmt"
	"strings"
	"time"
)

// Simple test for appsody build command. A future enhancement would be to verify the image that gets built.
func TestRun(projectDir string) error {

	// appsody run
	runChannel := make(chan error)
	containerName := "testRunContainer"
	go func() {
		Info.log("******************************************")
		Info.log("Running appsody run")
		Info.log("******************************************")
		_, err := RunAppsodyCmdExec([]string{"run", "--name", containerName}, projectDir)
		runChannel <- err
	}()

	// It will take a while for the container to spin up, so let's use docker ps to wait for it
	Info.log("calling docker ps to wait for container")
	containerRunning := false
	count := 100
	for {
		dockerOutput, dockerErr := RunDockerCmdExec([]string{"ps", "-q", "-f", "name=" + containerName})
		if dockerErr != nil {
			Info.log("Ignoring error running docker ps -q -f name="+containerName, dockerErr)
		}
		if dockerOutput != "" {
			Info.log("docker container " + containerName + " was found")
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
		Error.log("container never appeared to start")
	}

	// now run appsody ps and see if we can spot the container
	fmt.Println("about to run appsody ps")
	stopOutput, errStop := RunAppsodyCmdExec([]string{"ps"}, projectDir)
	if !strings.Contains(stopOutput, "CONTAINER") {
		Error.log("output doesn't contain header line")
	}
	if !strings.Contains(stopOutput, containerName) {
		Error.log("output doesn't contain correct container name")
	}
	if errStop != nil {
		Error.log(errStop)
	}

	// stop and clean up after the run
	func() {
		_, err := RunAppsodyCmdExec([]string{"stop", "--name", "testRunContainer"}, projectDir)
		if err != nil {
			Error.log(err)
		}
	}()

	return nil
}
