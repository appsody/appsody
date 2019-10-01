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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// Simple test for appsody debug command. A future enhancement would be to verify the debug output
func TestDebugSimple(t *testing.T) {

	log.Println("stacksList is: ", stacksList)

	// if stacksList is empty there is nothing to test so return
	if stacksList == "" {
		log.Println("stacksList is empty, exiting test...")
		return
	}

	// replace incubator with appsodyhub to match current naming convention for repos
	stacksList = strings.Replace(stacksList, "incubator", "appsodyhub", -1)

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		log.Println("***Testing stack: ", stackRaw[i], "***")

		// first add the test repo index
		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}

		// create a temporary dir to create the project and run the test
		projectDir, err := ioutil.TempDir("", "appsody-debug-simple-test")
		if err != nil {
			t.Fatal(err)
		}

		defer os.RemoveAll(projectDir)
		log.Println("Created project dir: " + projectDir)

		// appsody init
		_, err = cmdtest.RunAppsodyCmdExec([]string{"init", stackRaw[i]}, projectDir)
		log.Println("Running appsody init...")
		if err != nil {
			t.Fatal(err)
		}

		// appsody debug
		runChannel := make(chan error)
		containerName := "testDebugSimpleContainer" + strings.ReplaceAll(stackRaw[i], "/", "_")
		go func() {
			_, err := cmdtest.RunAppsodyCmdExec([]string{"debug", "--name", containerName}, projectDir)
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

		// stop and cleanup
		_, err = cmdtest.RunAppsodyCmdExec([]string{"stop", "--name", containerName}, projectDir)
		if err != nil {
			fmt.Printf("Ignoring error running appsody stop: %s", err)
		}

		cleanup()
	}
}
