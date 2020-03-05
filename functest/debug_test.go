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
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// Simple test for appsody debug command. A future enhancement would be to verify the debug output
func TestDebugSimple(t *testing.T) {
	stacksList := cmdtest.GetEnvStacksList()

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		t.Log("***Testing stack: ", stackRaw[i], "***")

		sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
		defer cleanup()

		// first add the test repo index
		_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join(sandbox.TestDataPath, "index.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		// appsody init
		t.Log("Running appsody init...")
		_, err = cmdtest.RunAppsody(sandbox, "init", stackRaw[i])
		if err != nil {
			t.Fatal(err)
		}

		// appsody debug
		runChannel := make(chan error)
		containerName := "testDebugSimpleContainer" + strings.ReplaceAll(stackRaw[i], "/", "_")
		go func() {
			_, err := cmdtest.RunAppsody(sandbox, "debug", "--name", containerName)
			runChannel <- err
			close(runChannel)
		}()

		// defer the appsody stop to close the docker container
		defer func() {
			_, err = cmdtest.RunAppsody(sandbox, "stop", "--name", containerName)
			if err != nil {
				t.Logf("Ignoring error running appsody stop: %s", err)
			}
			// wait for the appsody command/goroutine to finish
			runErr := <-runChannel
			if runErr != nil {
				t.Logf("Ignoring error from the appsody command: %s", runErr)
			}
		}()

		err = cmdtest.RunDockerPs(t, 50, containerName)
		if err != nil {
			t.Fatal(err)
		}
	}
}
