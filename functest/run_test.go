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
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// get the STACKSLIST environment variable
var stacksList = os.Getenv("STACKSLIST")

// Test appsody run of the nodejs-express stack and check the http://localhost:3000/health endpoint
func TestRun(t *testing.T) {
	// first add the test repo index
	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-run-test")
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
	go func() {
		_, err = cmdtest.RunAppsodyCmdExec([]string{"run"}, projectDir)
		runChannel <- err
	}()

	// defer the appsody stop to close the docker container
	defer func() {
		_, err = cmdtest.RunAppsodyCmdExec([]string{"stop"}, projectDir)
		if err != nil {
			log.Printf("Ignoring error running appsody stop: %s", err)
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
				log.Printf("Health check error. Ignore and retry: %s", err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != 200 {
					log.Printf("Health check response code %d. Ignore and retry.", resp.StatusCode)
				} else {
					log.Printf("Health check OK")
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
		projectDir, err := ioutil.TempDir("", "appsody-run-simple-test")
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

		// appsody run
		runChannel := make(chan error)
		containerName := "testRunSimpleContainer"
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

		// stop and clean up after the run
		_, err = cmdtest.RunAppsodyCmdExec([]string{"stop", "--name", containerName}, projectDir)
		if err != nil {
			fmt.Printf("Ignoring error running appsody stop: %s", err)
		}

		cleanup()
	}
}
