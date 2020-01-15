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
	"path/filepath"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestPackage(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	stackDir := filepath.Join(cmdtest.TestDirPath, "starter")
	err := cmd.CopyDir(log, stackDir, sandbox.ProjectDir)
	if err != nil {
		t.Errorf("Problem copying %s to %s: %v", stackDir, sandbox.ProjectDir, err)
	} else {
		t.Logf("Copied %s to %s", stackDir, sandbox.ProjectDir)
	}

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.ProjectDir, "starter")

	args := []string{"stack", "package"}
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPackageDockerOptions(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	stackDir := filepath.Join(cmdtest.TestDirPath, "starter")
	err := cmd.CopyDir(log, stackDir, sandbox.ProjectDir)
	if err != nil {
		t.Errorf("Problem copying %s to %s: %v", stackDir, sandbox.ProjectDir, err)
	} else {
		t.Logf("Copied %s to %s", stackDir, sandbox.ProjectDir)
	}

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.ProjectDir, "starter")

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
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	stackDir := filepath.Join(cmdtest.TestDirPath, "starter")
	err := cmd.CopyDir(log, stackDir, sandbox.ProjectDir)
	if err != nil {
		t.Errorf("Problem copying %s to %s: %v", stackDir, sandbox.ProjectDir, err)
	} else {
		t.Logf("Copied %s to %s", stackDir, sandbox.ProjectDir)
	}

	// Because the 'starter' folder has been copied, the stack.yaml file will be in the 'starter'
	// folder within the temp directory that has been generated for sandboxing purposes, rather than
	// the usual core temp directory
	sandbox.ProjectDir = filepath.Join(sandbox.ProjectDir, "starter")

	args := []string{"stack", "package", "--buildah", "--buildah-options", "--format=docker"}
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
}
