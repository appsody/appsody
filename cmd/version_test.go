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
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestVersion(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"version"}
	output, err := cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, cmd.VERSION) && !strings.Contains(output, cmd.CONTROLLERVERSION) {
		t.Fatal("Output does not contain CLI or controller version")
	}
}

func TestVersionTooManyArgs(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"version", "too", "many", "arguments"}
	output, err := cmdtest.RunAppsody(sandbox, args...)

	if err == nil {
		t.Fatal("Expected non-zero exit code.")
	}
	if !strings.Contains(output, "Unexpected argument.") {
		t.Fatal("Correct error message not given.")
	}
}

func TestOverrideControllerVersion(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	os.Setenv("APPSODY_CONTROLLER_VERSION", "0.3.0")

	args := []string{"version"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, cmd.VERSION) && !strings.Contains(output, "0.3.0") {
		t.Fatal("Output does not contain CLI or altered controller version")
	}
}

func TestOverrideControllerImage(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	os.Setenv("APPSODY_CONTROLLER_IMAGE", "appsody/new-controller:0.1.1")

	args := []string{"version"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, cmd.VERSION) && !strings.Contains(output, "appsody/new-controller:0.1.1") && !strings.Contains(output, "appsody-controller 0.1.1") {
		t.Fatal("Output does not contain CLI or new controller image")
	}
}
