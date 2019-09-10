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
	"strings"
	"time"

	"os"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestStopWithoutName(t *testing.T) {
	// first add the test repo index
	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-stop-test")
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
		fmt.Println("calling docker stop")
		stopOutput, errStop := cmdtest.RunAppsodyCmdExec([]string{"stop"}, projectDir)

		//docker stop appsody-stop-test
		if !strings.Contains(stopOutput, "docker stop appsody-stop-test") {
			t.Fatal("docker stop command not present for appsody-stop-test...")
		}
		if errStop != nil {
			log.Printf("Ignoring error running appsody stop: %s", errStop)

		}
		fmt.Println("calling docker ps")
		pathElements := strings.Split(projectDir, "/")
		containerName := pathElements[len(pathElements)-1]
		dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"ps", "-q", "-f", "name=" + containerName + "-dev"})
		fmt.Println("docker output", dockerOutput)
		if dockerErr != nil {
			log.Print("Ignoring error running docker ps -q -f name=appsody-stop-test-dev", dockerErr)

		}
		if dockerOutput != "" {
			t.Fatal("docker container " + containerName + " was found and should have been stopped")

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

func TestStopWithName(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-stopName-test")
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
		_, err = cmdtest.RunAppsodyCmdExec([]string{"run", "--name", "testStopContainer"}, projectDir)
		runChannel <- err
	}()

	// defer the appsody stop to close the docker container

	defer func() {
		fmt.Println("about to run stop for with name")
		stopOutput, errStop := cmdtest.RunAppsodyCmdExec([]string{"stop", "--name", "testStopContainer"}, projectDir)
		if !strings.Contains(stopOutput, "docker stop testStopContainer") {
			t.Fatal("docker stop command not present for container testStopContainer")
		}
		if errStop != nil {
			log.Printf("Ignoring error running appsody stop: %s", errStop)

		}
		fmt.Println("about to do docker ps")
		dockerOutput, dockerErr := cmdtest.RunDockerCmdExec([]string{"ps", "-q", "-f", "name=testStopContainer"})
		if dockerErr != nil {
			log.Print("Ignoring error running docker ps -q -f name=testStopContainer", dockerErr)

		}
		if dockerOutput != "" {
			t.Fatal("docker container testStopContainer was found and should have been stopped")

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
