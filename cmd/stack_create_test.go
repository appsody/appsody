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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestStackCreateSampleStack(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--config", filepath.Join(cmdtest.TestDirPath, "default_repository_config", "config.yaml")}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if !exists {
		t.Fatal(err)
	}
	os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackCreateWithCopyTag(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--config", filepath.Join(cmdtest.TestDirPath, "default_repository_config", "config.yaml"), "--copy", "incubator/nodejs"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if !exists {
		t.Fatal(err)
	}
	os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase1(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "incubator/nodej"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase2(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "nodejs"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase3(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "experimental/nodejs"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase4(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "exp/java-microprofile"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackName(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing_stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing_stack"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing_stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidLongStackName(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"stack", "create", "testing_stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stack"}
	_, err := cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing_stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackAlreadyExists(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--config", filepath.Join(cmdtest.TestDirPath, "default_repository_config", "config.yaml")}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if !exists {
		t.Fatal(err)
	}

	_, err = cmdtest.RunAppsody(sandbox, args...)

	if !strings.Contains(err.Error(), "A stack named testing-stack already exists in your directory. Specify a unique stack name") {
		t.Error("String \"A stack named testing-stack already exists in your directory. Specify a unique stack name\" not found in output")
	} else {
		if err == nil {
			t.Error("Expected error but did not receive one.")
		}
	}

	err = os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackCreateMissingArgumentsFail(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"stack", "create"}
	_, err := cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		if !strings.Contains(err.Error(), "Required parameter missing. You must specify a stack name") {
			t.Errorf("String \"Required parameter missing. You must specify a stack name\" not found in output: '%v'", err.Error())
		}
	} else {
		t.Error("Expected an error to be returned from command, but error was nil")
	}

}

func TestStackCreateSampleStackDryrun(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"stack", "create", "testing-stack", "--dryrun", "--config", "testdata/default_repository_config/config.yaml"}
	output, err := cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Errorf("Error running dry run mode: %v", err)
	} else {
		if !strings.Contains(output, "Dry run complete") {
			t.Errorf("String \"Dry run complete\" not found in output: '%v'", err.Error())
		}
	}
}
