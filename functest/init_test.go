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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// requires clean dir
func TestInit(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	appsodyResultsCheck(sandbox.ProjectDir, t)
}

//This test makes sure that no project creation occurred because app.js existed prior to the call
func TestNoOverwrite(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	appsodyFile := filepath.Join(sandbox.ProjectDir, ".appsody-config.yaml")

	appjs := filepath.Join(sandbox.ProjectDir, "app.js")
	packagejson := filepath.Join(sandbox.ProjectDir, "package.json")
	packagejsonlock := filepath.Join(sandbox.ProjectDir, "package-lock.json")

	appjsPath := filepath.Join(sandbox.ProjectDir, "app.js")
	_, err := os.Create(appjsPath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(appjs)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express"}
	_, _ = cmdtest.RunAppsody(sandbox, args...)

	shouldNotExist(appsodyFile, t)

	shouldNotExist(packagejson, t)
	shouldNotExist(packagejsonlock, t)

}

func shouldNotExist(file string, t *testing.T) {
	var err error
	_, err = os.Stat(file)
	if err == nil {
		err = errors.New(file + " should not exist without overwrite.")

		t.Fatal(err)
	}
}

//This test makes sure that no project creation occurred because app.js existed prior to the call
func TestOverwrite(t *testing.T) {

	var fileInfoFinal os.FileInfo
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	appsodyFile := filepath.Join(sandbox.ProjectDir, ".appsody-config.yaml")

	appjs := filepath.Join(sandbox.ProjectDir, "app.js")
	packagejson := filepath.Join(sandbox.ProjectDir, "package.json")
	packagejsonlock := filepath.Join(sandbox.ProjectDir, "package-lock.json")

	appjsPath := filepath.Join(sandbox.ProjectDir, "app.js")
	_, err := os.Create(appjsPath)
	if err != nil {
		t.Fatal(err)
	}
	//file should be 0 bytes
	_, err = os.Stat(appjs)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "--overwrite"}
	_, _ = cmdtest.RunAppsody(sandbox, args...)

	shouldExist(appsodyFile, t)

	shouldExist(packagejson, t)

	shouldExist(packagejsonlock, t)

	fileInfoFinal, err = os.Stat(appjs)
	if err != nil {
		err = errors.New(appjs + " should exist with overwrite.")

		t.Fatal(err)
	}

	if fileInfoFinal.Size() == 0 {
		err = errors.New(appjs + " should have data.")

		t.Fatal(err)
	}
}

//This test makes sure that no files are created except .appsody-config.yaml
func TestNoTemplate(t *testing.T) {
	var fileInfoFinal os.FileInfo
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	appsodyFile := filepath.Join(sandbox.ProjectDir, ".appsody-config.yaml")

	appjs := filepath.Join(sandbox.ProjectDir, "app.js")
	packagejson := filepath.Join(sandbox.ProjectDir, "package.json")
	packagejsonlock := filepath.Join(sandbox.ProjectDir, "package-lock.json")

	appjsPath := filepath.Join(sandbox.ProjectDir, "app.js")
	// file size should be 0 bytes
	_, err := os.Create(appjsPath)
	if err != nil {
		t.Fatal(err)
	}

	shouldExist(appjs, t)

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "--no-template"}
	_, _ = cmdtest.RunAppsody(sandbox, args...)

	shouldExist(appsodyFile, t)

	shouldNotExist(packagejson, t)

	shouldNotExist(packagejsonlock, t)

	fileInfoFinal, err = os.Stat(appjs)
	if err != nil {
		err = errors.New(appjs + " should exist without overwrite.")

		t.Fatal(err)
	}
	// if we accidentally overwrite the size would be >0
	if fileInfoFinal.Size() != 0 {
		err = errors.New(appjs + " should NOT have data.")

		t.Fatal(err)
	}
}

// the command should work despite existing artifacts
func TestWhiteList(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	appjs := filepath.Join(sandbox.ProjectDir, "app.js")
	vscode := filepath.Join(sandbox.ProjectDir, ".vscode")
	project := filepath.Join(sandbox.ProjectDir, ".project")
	cwSet := filepath.Join(sandbox.ProjectDir, ".cw-settings")
	cwExtension := filepath.Join(sandbox.ProjectDir, ".cw-extension")
	packagejson := filepath.Join(sandbox.ProjectDir, "package.json")
	packagejsonlock := filepath.Join(sandbox.ProjectDir, "package-lock.json")
	metadata := filepath.Join(sandbox.ProjectDir, ".metadata")
	appsodyFile := filepath.Join(sandbox.ProjectDir, ".appsody-config.yaml")

	_, err := os.Create(project)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(cwSet)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(cwExtension)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(vscode, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(metadata, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express"}
	_, _ = cmdtest.RunAppsody(sandbox, args...)

	shouldExist(appsodyFile, t)

	shouldExist(vscode, t)

	shouldExist(appjs, t)

	shouldExist(packagejson, t)

	shouldExist(packagejsonlock, t)

}
func shouldExist(file string, t *testing.T) {
	var err error
	_, err = os.Stat(file)
	if err != nil {
		t.Fatal(file, "should exist but didn't", err)
	}

}
func appsodyResultsCheck(projectDir string, t *testing.T) {

	appsodyFile := filepath.Join(projectDir, ".appsody-config.yaml")

	appjs := filepath.Join(projectDir, "app.js")
	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	shouldExist(appsodyFile, t)

	shouldExist(appjs, t)

	shouldExist(packagejson, t)

	shouldExist(packagejsonlock, t)

}

func TestInitV2WithDefaultRepoSpecified(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "incubator/nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}

	appsodyResultsCheck(sandbox.ProjectDir, t)
}

func TestInitV2WithNonDefaultRepoSpecified(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "experimental/nodejs-functions"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}

}

func TestInitV2WithBadStackSpecified(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "badnodejs-express"}
	output, _ := cmdtest.RunAppsody(sandbox, args...)
	if !(strings.Contains(output, "Could not find a stack with the id")) {
		t.Error("Should have flagged non existing stack")
	}

}

func TestInitV2WithBadRepoSpecified(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "badrepo/nodejs-express"}
	output, _ := cmdtest.RunAppsody(sandbox, args...)

	if !(strings.Contains(output, "is not in configured list of repositories")) {
		t.Log("Bad repo not flagged")
		t.Error("Bad repo not flagged")
	}

}

func TestInitV2WithDefaultRepoSpecifiedTemplateNonDefault(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "incubator/nodejs-express", "scaffold"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}

	appsodyResultsCheck(sandbox.ProjectDir, t)
}

func TestInitV2WithDefaultRepoSpecifiedTemplateDefault(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "incubator/nodejs-express", "simple"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}
	appsodyResultsCheck(sandbox.ProjectDir, t)
}

func TestInitV2WithNoRepoSpecifiedTemplateDefault(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "simple"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Error(err)
	}

	appsodyResultsCheck(sandbox.ProjectDir, t)
}
func TestNone(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	packagejson := filepath.Join(sandbox.ProjectDir, "package.json")
	packagejsonlock := filepath.Join(sandbox.ProjectDir, "package-lock.json")

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "none"}
	_, _ = cmdtest.RunAppsody(sandbox, args...)

	shouldNotExist(packagejson, t)
	shouldNotExist(packagejsonlock, t)

}

func TestNoneAndNoTemplate(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	packagejson := filepath.Join(sandbox.ProjectDir, "package.json")
	packagejsonlock := filepath.Join(sandbox.ProjectDir, "package-lock.json")

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "none", "--no-template"}
	_, _ = cmdtest.RunAppsody(sandbox, args...)

	shouldNotExist(packagejson, t)
	shouldNotExist(packagejsonlock, t)

}

func TestNoTemplateOnly(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	packagejson := filepath.Join(sandbox.ProjectDir, "package.json")
	packagejsonlock := filepath.Join(sandbox.ProjectDir, "package-lock.json")

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "--no-template"}
	_, _ = cmdtest.RunAppsody(sandbox, args...)

	shouldNotExist(packagejson, t)
	shouldNotExist(packagejsonlock, t)

}

func TestNoTemplateAndSimple(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "simple", "--no-template"}
	output, _ := cmdtest.RunAppsody(sandbox, args...)
	if !strings.Contains(output, "with both a template and --no-template") {
		t.Error("Correct error message not given")
	}
}
