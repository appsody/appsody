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
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestParser(t *testing.T) {
	// test environment variable with stack info
	// test environmet variable for stacks
	fmt.Println("stackList is: ", stacksList)

	stackRaw := strings.Split(stacksList, " ")
	// stackStack := strings.Split(stackRaw, "/")

	for i := range stackRaw {
		fmt.Println("stackRaw is: ", stackRaw[i])
		stageStack := strings.Split(stackRaw[i], "/")
		stage := stageStack[0]
		stack := stageStack[1]
		fmt.Println("stage is: ", stage)
		fmt.Println("stack is: ", stack)

	}

}

func TestDeploy(t *testing.T) {
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

		fmt.Println("***Testing stack: ", stack, "***")

		// first add the test repo index
		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
		if err != nil {
			t.Fatal(err)
		}
		// defer cleanup()
		// create a temporary dir to create the project and run the test
		projectDir, err := ioutil.TempDir("", "appsody-deploy-test")
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

		// appsody deploy
		runChannel := make(chan error)
		go func() {
			_, err = cmdtest.RunAppsodyCmdExec([]string{"deploy", "-t", "testdeploy/testimage", "--dryrun"}, projectDir)
			runChannel <- err
		}()

		cleanup()

	}

}
