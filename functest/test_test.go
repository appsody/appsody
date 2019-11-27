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
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// Simple test for appsody test command. A future enhancement would be to verify the test output
func TestTestSimple(t *testing.T) {

	t.Log("stacksList is: ", stacksList)

	// if stacksList is empty there is nothing to test so return
	if stacksList == "" {
		t.Log("stacksList is empty, exiting test...")
		return
	}

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		t.Log("***Testing stack: ", stackRaw[i], "***")

		// first add the test repo index
		_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml", t)
		if err != nil {
			t.Fatal(err)
		}

		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsodyCmd([]string{"init", stackRaw[i]}, projectDir, t)

		if err != nil {
			t.Fatal(err)
		}

		// appsody test
		runChannel := make(chan error)
		go func() {
			log.Println("Running appsody test...")
			_, err = cmdtest.RunAppsodyCmd([]string{"test"}, projectDir, t)
			runChannel <- err
		}()

		waitForError := 20 // in seconds
		stillWaiting := true
		log.Println("Waiting to see if test will fail...")
		for stillWaiting {
			select {
			case err = <-runChannel:
				if err != nil {
					// appsody run exited, probably with an error
					t.Fatalf("appsody test quit unexpectedly: %s", err)
				} else {
					t.Log("appsody test exited successfully")
					stillWaiting = false
				}
			case <-time.After(time.Duration(waitForError) * time.Second):
				fmt.Printf("appsody test kept running for %d seconds with no error so consider this passed\n", waitForError)
				stillWaiting = false
				// stop the container if it is still up
				_, err = cmdtest.RunAppsodyCmd([]string{"stop"}, projectDir, t)
				if err != nil {
					t.Logf("Ignoring error running appsody stop: %s", err)
				}
			}
		}

		// stop and cleanup

		cleanup()
		os.RemoveAll(projectDir)
	}
}
