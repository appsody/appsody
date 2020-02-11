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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestStackCreateDevLocal(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	packageArgs := []string{"stack", "package"}
	_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
	if err != nil {
		t.Fatal(err)
	}

	createArgs := []string{"stack", "create", "testing-stack", "--copy", "dev.local/starter"}
	_, err = cmdtest.RunAppsody(sandbox, createArgs...)
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

func TestStackCreateCustomRepo(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	packageArgs := []string{"stack", "package"}
	_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
	if err != nil {
		t.Fatal(err)
	}

	devlocalFolder := filepath.Join(sandbox.ConfigDir, "stacks", "dev.local")

	addToRepoArgs := []string{"stack", "add-to-repo", "test-repo", "--release-url", "file://" + devlocalFolder + "/"}

	_, err = cmdtest.RunAppsody(sandbox, addToRepoArgs...)
	if err != nil {
		t.Fatal(err)
	}

	testRepoIndex := filepath.Join(devlocalFolder, "test-repo-index.yaml")

	addRepoArgs := []string{"repo", "add", "test-repo", "file://" + testRepoIndex}
	_, err = cmdtest.RunAppsody(sandbox, addRepoArgs...)
	if err != nil {
		t.Fatal(err)
	}

	createArgs := []string{"stack", "create", "testing-stack", "--copy", "test-repo/starter"}
	_, err = cmdtest.RunAppsody(sandbox, createArgs...)
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

func TestStackCreateInvalidRepoFail(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	createArgs := []string{"stack", "create", "testing-stack", "--copy", "invalid/starter"}
	_, err := cmdtest.RunAppsody(sandbox, createArgs...)
	if err != nil {
		if !strings.Contains(err.Error(), "Repository: invalid not found in repository.yaml file") {
			t.Errorf("String \"Repository: invalid not found in repository.yaml file\" not found in output")
		}
	} else {
		t.Error("Stack create command unexpectededly passed with an invalid repository name")
	}

}

func TestStackCreateInvalidStackFail(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	packageArgs := []string{"stack", "package"}
	_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
	if err != nil {
		t.Fatal(err)
	}

	createArgs := []string{"stack", "create", "testing-stack", "--copy", "dev.local/invalid"}
	_, err = cmdtest.RunAppsody(sandbox, createArgs...)
	if err != nil {
		if !strings.Contains(err.Error(), "Stack not found in index") {
			t.Errorf("String \"Stack not found in index\" not found in output")
		}
	} else {
		t.Error("Stack create command unexpectededly passed with an invalid repository name")
	}

}
func TestStackCreateInvalidURLFail(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	packageArgs := []string{"stack", "package"}
	_, err := cmdtest.RunAppsody(sandbox, packageArgs...)
	if err != nil {
		t.Fatal(err)
	}

	devlocalFolder := filepath.Join(sandbox.ConfigDir, "stacks", "dev.local")

	addToRepoArgs := []string{"stack", "add-to-repo", "test-repo", "--release-url", "file://invalidurl/"}

	_, err = cmdtest.RunAppsody(sandbox, addToRepoArgs...)
	if err != nil {
		t.Fatal(err)
	}

	testRepoIndex := filepath.Join(devlocalFolder, "test-repo-index.yaml")

	addRepoArgs := []string{"repo", "add", "test-repo", "file://" + testRepoIndex}
	_, err = cmdtest.RunAppsody(sandbox, addRepoArgs...)
	if err != nil {
		t.Fatal(err)
	}

	createArgs := []string{"stack", "create", "testing-stack", "--copy", "test-repo/starter"}
	_, err = cmdtest.RunAppsody(sandbox, createArgs...)
	if err != nil {
		if !strings.Contains(err.Error(), "Could not download file://invalidurl") {
			t.Errorf("String \"Could not download file://invalidurl\" not found in output")
		}
	} else {
		t.Error("Stack create command unexpectededly passed with an invalid repository name")
	}

}
