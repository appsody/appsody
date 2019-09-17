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

	"github.com/appsody/appsody/cmd/cmdtest"
)

// Test parsing environment variable with stack info
func TestParser(t *testing.T) {

	fmt.Println("stacksList is: ", stacksList)
	if stacksList == "" {
		log.Println("stacksList is empty, exiting test...")
		return
	}

	// replace incubator with appsodyhub to match current naming convention for repos
	stacksList = strings.Replace(stacksList, "incubator", "appsodyhub", -1)
	fmt.Println("new stacksList is: ", stacksList)

	stackRaw := strings.Split(stacksList, " ")

	// we don't need to split the repo and stack anymore...
	// stackStack := strings.Split(stackRaw, "/")

	for i := range stackRaw {
		fmt.Println("stackRaw is: ", stackRaw[i])

		// code to sepearate the repos and stacks...
		// stageStack := strings.Split(stackRaw[i], "/")
		// stage := stageStack[0]
		// stack := stageStack[1]
		// fmt.Println("stage is: ", stage)
		// fmt.Println("stack is: ", stack)

	}

}

// Simple test for appsody deploy command. A future enhancement would be to configure a valid deployment environment
func TestDeploySimple(t *testing.T) {

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
		projectDir, err := ioutil.TempDir("", "appsody-deploy-simple-test")
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

		// appsody deploy
		runChannel := make(chan error)
		go func() {
			_, err = cmdtest.RunAppsodyCmdExec([]string{"deploy", "-t", "testdeploy/testimage", "--dryrun"}, projectDir)
			log.Println("Running appsody deploy...")
			runChannel <- err
		}()

		// cleanup tasks
		cleanup()
	}
}
