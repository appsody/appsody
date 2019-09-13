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

func TestRunWithDockerOptionsRegex(t *testing.T) {

	var runOutput string
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-docker-options-test")
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

	//var testOptions = []string{"-p", "--publish", "--publish-all", "-P", "-u", "--user", "--name", "--network", "-t", "--tty", "--rm", "--entrypoint", "-v", "--volume", "-e", "--env"}
	var testOptions = []string{"-p", "--publish",
		"--publish-all",
		"-P",
		"-u", "--user",
		"--name",
		"--network",
		"-t",
		"--tty",
		"--rm",
		"--entrypoint",
		"-v", "--volume"}

	for _, value := range testOptions {
		fmt.Println("Option is", value)
		runOutput, err = cmdtest.RunAppsodyCmdExec([]string{"run", "--docker-options", value, "--dryrun"}, projectDir)
		fmt.Println("err ", err)
		if !strings.Contains(runOutput, value+" is not allowed in --docker-options") {
			t.Fatal("Error message not found:" + value + " is not allowed in --docker-options")

		}
		runOutput, err = cmdtest.RunAppsodyCmdExec([]string{"run", "--docker-options", value + "=", "--dryrun"}, projectDir)
		fmt.Println("err ", err)
		if !strings.Contains(runOutput, value+"="+" is not allowed in --docker-options") {
			t.Fatal("Error message not found:" + value + "=" + " is not allowed in --docker-options")

		}

	}
}
