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
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestPackage(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	args := []string{"stack", "package"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPackageImageTag(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	args := []string{"stack", "package", "--image-namespace", "testnamespace", "--image-registry", "testregistry"}
	_, err := cmdtest.RunAppsody(sandbox, args...)

	file, readErr := ioutil.ReadFile(filepath.Join(sandbox.ConfigDir, "stacks", "dev.local", "dev.local-index.yaml"))
	if readErr != nil {
		t.Fatal(readErr)
	}

	if err != nil {
		t.Fatal(err)
	} else {
		if !strings.Contains(string(file), "image: testregistry/testnamespace") {
			t.Errorf("Image name not found in index. Expecting: image: testregistry/testnamespace/starter:<version>")
		}
	}
}

func TestPackageDockerOptions(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	args := []string{"stack", "package", "--docker-options", "-q"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	} else {
		if strings.Contains(output, "[Docker] Sending build context to Docker daemon") {
			t.Errorf("String \"[Docker] Sending build context to Docker daemon\" found in output")
		}
	}
}

func TestPackageBuildah(t *testing.T) {

	if runtime.GOOS != "linux" {
		t.Skip()
	}

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	args := []string{"stack", "package", "--buildah"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	} else {
		if !strings.Contains(output, "[Buildah] Writing manifest to image destination") {
			t.Errorf("String \"[Buildah] Writing manifest to image destination\" not found in output")
		}
	}

}

func TestPackageBuildahWithOptions(t *testing.T) {

	if runtime.GOOS != "linux" {
		t.Skip()
	}

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "starter")

	args := []string{"stack", "package", "--buildah", "--buildah-options", "--format=docker"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	} else {
		if !strings.Contains(output, "--format=docker") {
			t.Error("Buildah options not passed successfuly")
		}
		if !strings.Contains(output, "[Buildah] Writing manifest to image destination") {
			t.Error("String \"[Buildah] Writing manifest to image destination\" not found in output")
		}
	}

}

func TestPackageDeprecatedStack(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	projectDir := sandbox.ProjectDir
	sandbox.ProjectDir = filepath.Join(sandbox.TestDataPath, "deprecated-stack")

	args := []string{"stack", "package"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	sandbox.ProjectDir = projectDir
	args = []string{"init", "dev.local/deprecated-stack"}
	output, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	expectedOutput := "Stack deprecated: 01/01/0001 - this is a test"
	if !strings.Contains(output, expectedOutput) {
		t.Fatalf("Did not get expected error: %s", expectedOutput)
	}
}
