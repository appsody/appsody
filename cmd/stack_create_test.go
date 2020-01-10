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
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, false)
	defer cleanup()

	testStackName := "testing-create-sample-stack"

	err := os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName, "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml")}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal(err)
	}
	os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackCreateWithCopyTag(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, false)
	defer cleanup()

	testStackName := "testing-create-stack-with-copy"

	err := os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName, "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml"), "--copy", "incubator/nodejs"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal(err)
	}
	os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase1(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackName := "testing-stack-create-invalid-1"
	err := os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName, "--copy", "incubator/nodej"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		// It SHOULDN'T exist, but it might
		err = os.RemoveAll(testStackName)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase2(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackName := "testing-stack-create-invalid-2"
	err := os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName, "--copy", "nodejs"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		// It SHOULDN'T exist, but it might
		err = os.RemoveAll(testStackName)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase3(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackName := "testing-stack-create-invalid-3"

	err := os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName, "--copy", "experimental/nodejs"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		// It SHOULDN'T exist, but it might
		err = os.RemoveAll(testStackName)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase4(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackName := "testing-stack-create-invalid-4"

	err := os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName, "--copy", "exp/java-microprofile"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		// It SHOULDN'T exist, but it might
		err = os.RemoveAll(testStackName)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackName(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackName := "testing_stack_invalid_name"
	err := os.RemoveAll(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		// It SHOULDN'T exist, but it might
		err = os.RemoveAll(testStackName)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(err)
	}
}

func TestStackCreateInvalidLongStackName(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackName := "testing_stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stack"

	args := []string{"stack", "create", testStackName}
	_, err := cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		// It SHOULDN'T exist, but it might
		err = os.RemoveAll(testStackName)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(err)
	}
}

func TestStackAlreadyExists(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	testStackName := "test-stack-already-exists"

	err := os.RemoveAll(testStackName)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", testStackName, "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml")}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists(testStackName)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal(err)
	}

	_, err = cmdtest.RunAppsody(sandbox, args...)

	if !strings.Contains(err.Error(), "A stack named "+testStackName+" already exists in your directory. Specify a unique stack name") {
		t.Error("String \"A stack named " + testStackName + "already exists in your directory. Specify a unique stack name\" not found in output")
	} else {
		if err == nil {
			t.Error("Expected error but did not receive one.")
		}
	}

	err = os.RemoveAll(testStackName)
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

	args := []string{"stack", "create", "testing-stack", "--dryrun", "--config", filepath.Join(sandbox.TestDataPath, "default_repository_config", "config.yaml")}
	output, err := cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Errorf("Error running dry run mode: %v", err)
	} else {
		if !strings.Contains(output, "Dry run complete") {
			t.Error("String \"Dry run complete\" not found in output")
		}
	}
}
