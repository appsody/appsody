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

//var stacksList = os.Getenv("STACKSLIST")

//var stacksList = ""

var stacksList = "incubator/java-microprofile experimental/quarkus incubator/nodejs experimental/nodejs-functions"

//var stack = flag.String("stack", "", "Stack to run tests on")

/* func TestTnixa(t *testing.T) {
	fmt.Println(os.Args)
	fmt.Println("Stack is: ", stacks)

} */
func TestRun(t *testing.T) {

	// replace incubator with appsodyhub to match current naming convention for repos
	stacksList = strings.Replace(stacksList, "incubator", "appsodyhub", -1)

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {
		// fmt.Println("stackRaw is: ", stackRaw[i])

		// split out the stage and stack
		// stageStack := strings.Split(stackRaw[i], "/")
		// stage := stageStack[0]
		// stack := stageStack[1]
		// fmt.Println("stage is: ", stage)
		// fmt.Println("stack is: ", stack)
		// first add the test repo index

		fmt.Println("***Testing stack: ", stackRaw[i], "***")

		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}
		//defer cleanup()

		// create a temporary dir to create the project and run the test
		projectDir, err := ioutil.TempDir("", "appsody-run-test")
		if err != nil {
			t.Fatal(err)
		}
		//defer os.RemoveAll(projectDir)
		fmt.Println("Created project dir: " + projectDir)

		// appsody init nodejs-express
		_, err = cmdtest.RunAppsodyCmdExec([]string{"init", stackRaw[i]}, projectDir)
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
		// defer func() {
		// 	_, err = cmdtest.RunAppsodyCmdExec([]string{"stop"}, projectDir)
		// 	if err != nil {
		// 		fmt.Printf("Ignoring error running appsody stop: %s", err)
		// 	}
		// }()

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
					fmt.Printf("Health check error. Ignore and retry: %s", err)
				} else {
					resp.Body.Close()
					if resp.StatusCode != 200 {
						fmt.Printf("Health check response code %d. Ignore and retry.", resp.StatusCode)
					} else {
						fmt.Printf("Health check OK")
						// may want to check body
						healthCheckOK = true
					}
				}
			}
		}

		if !healthCheckOK {
			t.Errorf("Did not receive an OK health check within %d seconds.", healthCheckTimeout)
		}

		cleanup()
		os.RemoveAll(projectDir)
		func() {
			_, err = cmdtest.RunAppsodyCmdExec([]string{"stop"}, projectDir)
			if err != nil {
				fmt.Printf("Ignoring error running appsody stop: %s", err)
			}
		}()

	}
}

func TestRunSimple(t *testing.T) {

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		// fmt.Println("stackRaw is: ", stackRaw[i])

		// split out the stage and stack
		stageStack := strings.Split(stackRaw[i], "/")
		// stage := stageStack[0]
		stack := stageStack[1]
		// fmt.Println("stage is: ", stage)
		// fmt.Println("stack is: ", stack)
		// first add the test repo index

		//stackRaw = strings.Replace(stackRaw, "incubator", "appsodyhub", -1)
		fmt.Println("stackRaw is: ", stackRaw)

		fmt.Println("***Testing stack: ", stackRaw, "***")

		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}
		//defer cleanup()

		// create a temporary dir to create the project and run the test
		projectDir, err := ioutil.TempDir("", "appsody-run-test")
		if err != nil {
			t.Fatal(err)
		}
		//defer os.RemoveAll(projectDir)
		log.Println("Created project dir: " + projectDir)

		// appsody init nodejs-express
		_, err = cmdtest.RunAppsodyCmdExec([]string{"init", stack}, projectDir)
		if err != nil {
			t.Fatal(err)
		}

		// appsody run
		runChannel := make(chan error)
		go func() {
			_, err = cmdtest.RunAppsodyCmdExec([]string{"run"}, projectDir)
			runChannel <- err
		}()

		cleanup()
		os.RemoveAll(projectDir)
		func() {
			_, err = cmdtest.RunAppsodyCmdExec([]string{"stop"}, projectDir)
			if err != nil {
				fmt.Printf("Ignoring error running appsody stop: %s", err)
			}
		}()

	}
}

func TestParseRunOutput(t *testing.T) {

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {
		// fmt.Println("stackRaw is: ", stackRaw[i])

		// split out the stage and stack
		stageStack := strings.Split(stackRaw[i], "/")
		// stage := stageStack[0]
		stack := stageStack[1]
		// fmt.Println("stage is: ", stage)
		// fmt.Println("stack is: ", stack)
		// first add the test repo index

		fmt.Println("***Testing stack: ", stack, "***")

		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}
		//defer cleanup()

		// create a temporary dir to create the project and run the test
		projectDir, err := ioutil.TempDir("", "appsody-run-test")
		if err != nil {
			t.Fatal(err)
		}
		//defer os.RemoveAll(projectDir)
		fmt.Println("Created project dir: " + projectDir)

		// appsody init nodejs-express
		_, err = cmdtest.RunAppsodyCmdExec([]string{"init", stack}, projectDir)
		if err != nil {
			t.Fatal(err)
		}

		// appsody run
		var runOutput string

		runOutput, err = cmdtest.RunAppsodyCmdExec([]string{"run"}, projectDir)
		//fmt.Println("### runOutput:", runOutput)

		// defer the appsody stop to close the docker container
		// defer func() {
		// 	_, err = cmdtest.RunAppsodyCmdExec([]string{"stop"}, projectDir)
		// 	if err != nil {
		// 		fmt.Printf("Ignoring error running appsody stop: %s", err)
		// 	}
		// }()
		fmt.Println("### runOutput before for loop: ", runOutput)
		healthCheckFrequency := 2 // in seconds
		fmt.Println("### health")
		healthCheckTimeout := 60 // in seconds
		healthCheckWait := 0
		healthCheckOK := false
		for !(healthCheckOK || healthCheckWait >= healthCheckTimeout) {
			fmt.Println("### For loop")
			fmt.Println("### runOutput: ", runOutput)
			select {
			//case err = <-runChannel:
			// appsody run exited, probably with an error
			//t.Fatalf("appsody run quit unexpectedly: %s", err)
			case <-time.After(time.Duration(healthCheckFrequency) * time.Second):
				// check the health endpoint
				healthCheckWait += healthCheckFrequency

				if !strings.Contains(runOutput, "3000") {
					fmt.Println("### Health check not found")
					//t.Fatalf("Hello not found in the output")
				} else {
					fmt.Println("### Health check OK")
					// may want to check body
					healthCheckOK = true
				}

			}
		}

		if !healthCheckOK {
			t.Errorf("Did not receive Hello within %d seconds.", healthCheckTimeout)
		}

		cleanup()
		os.RemoveAll(projectDir)
		func() {
			_, err = cmdtest.RunAppsodyCmdExec([]string{"stop"}, projectDir)
			if err != nil {
				fmt.Printf("Ignoring error running appsody stop: %s", err)
			}
		}()

	}
}
