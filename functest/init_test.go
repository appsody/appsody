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

var initResultsCheckTests = []struct {
	testName string
	args     []string //input
}{
	{"TestInit", []string{"nodejs-express"}},
	{"TestInitV2WithDefaultRepoSpecified", []string{"incubator/nodejs-express"}},
	{"TestInitV2WithDefaultRepoSpecifiedTemplateNonDefault", []string{"incubator/nodejs-express", "scaffold"}},
	{"TestInitV2WithDefaultRepoSpecifiedTemplateDefault", []string{"incubator/nodejs-express", "simple"}},
	{"TestInitV2WithNoRepoSpecifiedTemplateDefault", []string{"nodejs-express", "simple"}},
}

func TestInitResultsCheck(t *testing.T) {
	for _, tt := range initResultsCheckTests {
		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		t.Run(tt.testName, func(t *testing.T) {
			args := append([]string{"init"}, tt.args...)
			_, err := cmdtest.RunAppsodyCmd(args, projectDir, t)
			if err != nil {
				t.Error(err)
			}
			appsodyResultsCheck(projectDir, t)
		})
	}
}

var initLogsTests = []struct {
	testName     string
	args         []string //input
	expectedLogs string   //expected output
	outputError  string   //error output if test fails
}{
	{"TestInitV2WithBadStackSpecified", []string{"badnodejs-express"}, "Could not find a stack with the id", "Should have flagged non existing stack"},
	{"TestInitV2WithBadRepoSpecified", []string{"badrepo/nodejs-express"}, "is not in configured list of repositories", "Bad repo not flagged"},
	{"TestInitWithBadTemplateSpecified", []string{"nodejs-express", "badtemplate"}, "Could not find a template", "Should have flagged non existing stack template"},
	{"TestInitNoTemplateAndSimple", []string{"nodejs-express", "simple", "--no-template"}, "with both a template and --no-template", "Correct error message not given"},
	{"TestInitWithBadProjectName", []string{"nodejs-express", "--project-name", "badprojectname!"}, "Invalid project-name", "Correct error message not given"},
	{"TestInitWithBadlyFormattedConfig", []string{"nodejs-express", "--config", "testdata/bad_format_repository_config/config.yaml"}, "Failed to parse repository file yaml", "Correct error message not given"},
	{"TestInitWithEmptyConfig", []string{"nodejs-express", "--config", "testdata/empty_repository_config/config.yaml"}, "Your stack repository is empty", "Correct error message not given"},
	{"TestInitWithBadRepoUrlConfig", []string{"nodejs-express", "--config", "testdata/bad_repo_url_repository_config/config.yaml"}, "The following indices could not be read, skipping", "Correct error message not given"},
	{"TestInitV2WithNonDefaultRepoSpecified", []string{"experimental/nodejs-functions"}, "Successfully initialized Appsody project", "Init should have passed without errors."},
	{"TestInitV2WithStackHasInitScript", []string{"java-microprofile"}, "Successfully initialized Appsody project", "Init should have passed without errors."},
	{"TestInitV2WithStackHasInitScriptDryrun", []string{"java-microprofile", "--dryrun"}, "Dry Run - Skipping", "Commands should be skipped on dry run"},
	{"TestInitDryRun", []string{"nodejs-express", "--dryrun"}, "Dry Run - Skipping", "Commands should be skipped on dry run"},
	{"TestInitMalformedStackParm", []string{"/nodejs-express"}, "malformed project parameter - slash at the beginning or end should be removed", "Malformed stack parameter should be flagged."},
	{"TestInitStackParmTooManySlashes", []string{"incubator/nodejs-express/bad"}, "malformed project parameter - too many slashes", "Malformed stack parameter with too many slashes should be flagged."},
}

func TestLogsErrors(t *testing.T) {
	for _, tt := range initLogsTests {
		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		t.Run(tt.testName, func(t *testing.T) {
			args := append([]string{"init"}, tt.args...)
			// appsody init nodejs-express
			output, _ := cmdtest.RunAppsodyCmd(args, projectDir, t)
			if !(strings.Contains(output, tt.expectedLogs)) {
				t.Error(tt.outputError)
			}
		})
	}

}

var initTemplateShouldNotExistTests = []struct {
	testName string
	args     []string //input
}{
	{"TestInitNone", []string{"none"}},
	{"TestNoneAndNoTemplate", []string{"none", "--no-template"}},
	{"TestNoTemplateOnly", []string{"--no-template"}},
}

func TestInitTemplateShouldNotExistTests(t *testing.T) {

	for _, tt := range initTemplateShouldNotExistTests {

		// create a temporary dir to create the project and run the test
		projectDir := cmdtest.GetTempProjectDir(t)
		defer os.RemoveAll(projectDir)
		t.Log("Created project dir: " + projectDir)

		packagejson := filepath.Join(projectDir, "package.json")
		packagejsonlock := filepath.Join(projectDir, "package-lock.json")

		t.Run(tt.testName, func(t *testing.T) {
			args := append([]string{"init", "nodejs-express"}, tt.args...)
			_, _ = cmdtest.RunAppsodyCmd(args, projectDir, t)
			shouldNotExist(packagejson, t)
			shouldNotExist(packagejsonlock, t)
		})
	}
}

func TestInitOnExistingAppsodyProject(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir := cmdtest.GetTempProjectDir(t)
	defer os.RemoveAll(projectDir)
	t.Log("Created project dir: " + projectDir)

	args := []string{"init", "nodejs-express"}

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmd(args, projectDir, t)

	output, _ := cmdtest.RunAppsodyCmd(args, projectDir, t)
	if !(strings.Contains(output, "cannot run `appsody init <stack>` on an existing appsody project")) {
		t.Error("Should have flagged that you cannot init an existing Appsody project.")
	}
}

//This test makes sure that no project creation occurred because app.js existed prior to the call
func TestNoOverwrite(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir := cmdtest.GetTempProjectDir(t)
	defer os.RemoveAll(projectDir)
	t.Log("Created project dir: " + projectDir)

	appsodyFile := filepath.Join(projectDir, ".appsody-config.yaml")
	appjs := filepath.Join(projectDir, "app.js")
	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	_, err := os.Create(appjs)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(appjs)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmd([]string{"init", "nodejs-express"}, projectDir, t)

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
	// create a temporary dir to create the project and run the test
	projectDir := cmdtest.GetTempProjectDir(t)
	defer os.RemoveAll(projectDir)
	t.Log("Created project dir: " + projectDir)

	appjs := filepath.Join(projectDir, "app.js")
	_, err := os.Create(appjs)
	if err != nil {
		t.Fatal(err)
	}
	//file should be 0 bytes
	_, err = os.Stat(appjs)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmd([]string{"init", "nodejs-express", "--overwrite"}, projectDir, t)

	appsodyResultsCheck(projectDir, t)

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
	// create a temporary dir to create the project and run the test
	projectDir := cmdtest.GetTempProjectDir(t)
	defer os.RemoveAll(projectDir)
	t.Log("Created project dir: " + projectDir)

	appsodyFile := filepath.Join(projectDir, ".appsody-config.yaml")
	appjs := filepath.Join(projectDir, "app.js")
	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	// file size should be 0 bytes
	_, err := os.Create(appjs)
	if err != nil {
		t.Fatal(err)
	}

	shouldExist(appjs, t)

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmd([]string{"init", "nodejs-express", "--no-template"}, projectDir, t)

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

	// create a temporary dir to create the project and run the test
	projectDir := cmdtest.GetTempProjectDir(t)
	defer os.RemoveAll(projectDir)
	t.Log("Created project dir: " + projectDir)

	vscode := filepath.Join(projectDir, ".vscode")
	project := filepath.Join(projectDir, ".project")
	cwSet := filepath.Join(projectDir, ".cw-settings")
	cwExtension := filepath.Join(projectDir, ".cw-extension")
	metadata := filepath.Join(projectDir, ".metadata")

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
	_, _ = cmdtest.RunAppsodyCmd([]string{"init", "nodejs-express"}, projectDir, t)

	shouldExist(vscode, t)
	appsodyResultsCheck(projectDir, t)
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
