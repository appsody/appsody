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
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestStackValidateNoLintFlag(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	args := []string{"stack", "validate", "--no-lint"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatalf("Stack validate failed unexpectedly: %v", err)
	}
	if !strings.Contains(output, "Total PASSED: 5") && !strings.Contains(output, "Total FAILED: 0") {
		t.Error("Stack validate did not have expected PASS/FAIL amounts")
	}
	if strings.Contains(output, "PASSED: Lint for stack") {
		t.Error("Stack validate --no-package still ran packaging process")
	}
}

func TestStackValidateNoPackageFlag(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, false)
	defer cleanup()
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	_, err := cmdtest.RunAppsody(sandbox, "stack", "package", "--image-namespace", "appsody", "--image-registry", "dev.local")

	if err != nil {
		t.Fatal(err)
	}

	_, err = cmdtest.RunAppsody(sandbox, "repo", "list")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "validate", "--no-package"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatalf("Stack validate failed unexpectedly: %v", err)
	}
	if !strings.Contains(output, "Total PASSED: 5") && !strings.Contains(output, "Total FAILED: 0") {
		t.Error("Stack validate did not have expected PASS/FAIL amounts")
	}
	if strings.Contains(output, "PASSED: Package for stack") {
		t.Error("Stack validate --no-package still ran packaging process")
	}
}
