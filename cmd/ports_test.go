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
package cmd_test

import (
	"io/ioutil"
	"log"
	"strings"

	"os"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

var runOutput string

func TestPortMap(t *testing.T) {

	runOutput = ""
	// first add the test repo index
	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-ports-test")
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

	runOutput, _ = cmdtest.RunAppsodyCmdExec([]string{"run", "--dryrun", "--publish", "3100:3000", "--publish", "4100:4000", "--publish", "9230:9229"}, projectDir)
	if !strings.Contains(runOutput, "docker[run --rm -p 3100:3000 -p 4100:4000 -p 9230:9229") {

		t.Fatal("Ports are not correctly specified as: -p 3100:3000 -p 4100:4000 -p 9230:9229")

	}

}

// This test tests the setting of --publish-all in dry run mode
func TestPublishAll(t *testing.T) {

	runOutput = ""
	// first add the test repo index

	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-publish-all-test")
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
	runOutput, _ = cmdtest.RunAppsodyCmdExec([]string{"run", "--publish-all", "--dryrun"}, projectDir)

	if !strings.Contains(runOutput, "docker[run --rm -P") {
		t.Fatal("publish all is not found in output as: docker[run --rm -P")

	}

}

// This test tests the setting of --network and --publish-all in dry run mode
func TestRunWithNetwork(t *testing.T) {
	// first add the test repo index

	runOutput = ""
	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-network-test")
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

	runOutput, _ = cmdtest.RunAppsodyCmdExec([]string{"run", "--network", "noSuchNetwork", "--publish-all", "--dryrun"}, projectDir)

	if !strings.Contains(runOutput, "--network noSuchNetwork") {
		t.Fatal("--networkis not found in output as: --network noSuchNetwork")

	}
}
