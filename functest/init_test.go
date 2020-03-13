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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

var appsodyFile = ".appsody-config.yaml"
var appjs = "app.js"
var packagejson = "package.json"
var packagejsonlock = "package-lock.json"

func TestInitResultsCheck(t *testing.T) {

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

	for _, testData := range initResultsCheckTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			testDir := filepath.Join(sandbox.ProjectDir, tt.testName)
			err := os.Mkdir(testDir, os.FileMode(0755))
			if err != nil {
				t.Errorf("Error creating directory: %v", err)
			}
			sandbox.ProjectDir = filepath.Join(sandbox.ProjectDir, tt.testName)

			args := append([]string{"init"}, tt.args...)

			_, err = cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatal(err)
			}
			expressExist(sandbox.ProjectDir, true, t)
		})

	}
}

func TestInitErrors(t *testing.T) {

	var initErrorsTests = []struct {
		testName     string
		args         []string //input
		configDir    string
		expectedLogs string //expected output
		outputError  string //error output if test fails
	}{
		{"TestInitV2WithBadStackSpecified", []string{"badnodejs-express"}, "", "Could not find a stack with the id", "Should have flagged non existing stack"},
		{"TestInitV2WithBadRepoSpecified", []string{"badrepo/nodejs-express"}, "", "is not in configured list of repositories", "Bad repo not flagged"},
		{"TestInitWithBadTemplateSpecified", []string{"nodejs-express", "badtemplate"}, "", "Could not find a template", "Should have flagged non existing stack template"},
		{"TestInitNoTemplateAndSimple", []string{"nodejs-express", "simple", "--no-template"}, "", "with both a template and --no-template", "Correct error message not given"},
		{"TestInitWithBadProjectName", []string{"nodejs-express", "--project-name", "badprojectname!"}, "", "Invalid project-name", "Correct error message not given"},
		{"TestInitWithBadApplicationName", []string{"nodejs-express", "--application-name", "badapplicationname!"}, "", "Invalid application-name", "Correct error message not given"},
		{"TestInitWithBadlyFormattedConfig", []string{"nodejs-express"}, "bad_format_repository_config", "Failed to parse repository file yaml", "Correct error message not given"},
		{"TestInitWithEmptyConfig", []string{"nodejs-express"}, "empty_repository_config", "Your stack repository is empty", "Correct error message not given"},
		{"TestInitWithBadRepoUrlConfig", []string{"nodejs-express"}, "bad_repo_url_repository_config", "Does the APIVersion of your repository match what the Appsody CLI currently supports?", "Correct error message not given"},
		{"TestInitV2WithStackHasInitScriptDryrun", []string{"java-microprofile", "--dryrun"}, "Dry Run - Skipping", "", "Commands should be skipped on dry run"},
		{"TestInitDryRun", []string{"nodejs-express", "--dryrun"}, "Dry Run - Skipping", "", "Commands should be skipped on dry run"},
		{"TestInitMalformedStackParm", []string{"/nodejs-express"}, "", "malformed project parameter - slash at the beginning or end should be removed", "Malformed stack parameter should be flagged."},
		{"TestInitStackParmTooManySlashes", []string{"incubator/nodejs-express/bad"}, "", "malformed project parameter - too many slashes", "Malformed stack parameter with too many slashes should be flagged."},
		{"TooManyArguments", []string{"too", "many", "arguments"}, "", "Too many arguments.", "Too many arguments given should be flagged."},
	}

	for _, testData := range initErrorsTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()
			args := append([]string{"init"}, tt.args...)
			// appsody init nodejs-express

			sandbox.SetConfigInTestData(tt.configDir)

			output, err := cmdtest.RunAppsody(sandbox, args...)
			if !strings.Contains(output, tt.expectedLogs) {
				t.Error(tt.outputError, " ", err)
			} else if err == nil {
				t.Errorf("Expected an error from test %v but it did not return one.", tt.testName)
			}
			expressExist(sandbox.ProjectDir, false, t)
		})
	}
}

func TestInitWithApplicationName(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	appsodyConfig := filepath.Join(sandbox.ProjectDir, ".appsody-config.yaml")

	args := []string{"init", "nodejs", "--application-name", "my-big-app"}

	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(appsodyConfig)
	if err != nil {
		t.Fatalf("Error reading %s file: %v", appsodyConfig, err)
	}
	s := string(b)
	if !strings.Contains(s, "application-name: my-big-app") {
		t.Fatal("ApplicationName did not match expected value")
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

	for _, testData := range initTemplateShouldNotExistTests {
		// need to set testData to a new variable scoped under the for loop
		// otherwise tests run in parallel may get the wrong testData
		// because the for loop reassigns it before the func runs
		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			testDir := filepath.Join(sandbox.ProjectDir, tt.testName)
			err := os.Mkdir(testDir, os.FileMode(0755))
			if err != nil {
				t.Errorf("Error creating directory: %v", err)
			}
			sandbox.ProjectDir = filepath.Join(sandbox.ProjectDir, tt.testName)

			args := append([]string{"init", "nodejs-express"}, tt.args...)

			_, err = cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatal(err)
			}
			filesExist([]string{packagejson, packagejsonlock}, sandbox.ProjectDir, false, t)
		})
	}
}

func TestInitV2WithNonDefaultRepoSpecified(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	functionjs := "function.js"

	args := []string{"init", "experimental/nodejs-functions"}

	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	filesExist([]string{appsodyFile, functionjs, packagejson}, sandbox.ProjectDir, true, t)
}

func TestInitV2WithStackHasInitScript(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	pomxml := "pom.xml"

	args := []string{"init", "java-microprofile"}

	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	filesExist([]string{appsodyFile, pomxml}, sandbox.ProjectDir, true, t)
}

func TestInitOnExistingAppsodyProject(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs-express"}

	// appsody init nodejs-express
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	output, err := cmdtest.RunAppsody(sandbox, args...)
	if !(strings.Contains(output, "cannot run `appsody init <stack>` on an existing appsody project")) {
		t.Error("Should have flagged that you cannot run `init` on an existing Appsody project.")
	} else if err == nil {
		t.Errorf("Expected an error from test TestInitOnExistingAppsodyProject but it did not return one.")
	}
}

//This test makes sure that no project creation occurred because app.js existed prior to the call
func TestNoOverwrite(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	createAndStat(filepath.Join(sandbox.ProjectDir, appjs), t)

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if !strings.Contains(err.Error(), "non-empty directory found with files which may conflict with the template project") {
		t.Errorf("Correct error message not given: %v", err)
	}

	filesExist([]string{appsodyFile, packagejson, packagejsonlock}, sandbox.ProjectDir, false, t)

}

//This test makes sure that no project creation occurred because app.js existed prior to the call
func TestOverwrite(t *testing.T) {

	var fileInfoFinal os.FileInfo
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	createAndStat(filepath.Join(sandbox.ProjectDir, appjs), t)

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "--overwrite"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	expressExist(sandbox.ProjectDir, true, t)

	fileInfoFinal, err = os.Stat(filepath.Join(sandbox.ProjectDir, appjs))
	if err != nil {
		t.Fatal(appjs + " should exist with overwrite.")
	}

	if fileInfoFinal.Size() == 0 {
		t.Fatal(appjs + " should have data.")
	}
}

//This test makes sure that no files are created except .appsody-config.yaml
func TestNoTemplate(t *testing.T) {
	var fileInfoFinal os.FileInfo
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	createAndStat(filepath.Join(sandbox.ProjectDir, appjs), t)

	filesExist([]string{appjs}, sandbox.ProjectDir, true, t)

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express", "--no-template"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	filesExist([]string{appsodyFile}, sandbox.ProjectDir, true, t)
	filesExist([]string{packagejson, packagejsonlock}, sandbox.ProjectDir, false, t)

	fileInfoFinal, err = os.Stat(filepath.Join(sandbox.ProjectDir, appjs))
	if err != nil {
		t.Fatal(appjs + " should exist without overwrite.")
	}
	// if we accidentally overwrite the size would be >0
	if fileInfoFinal.Size() != 0 {
		t.Fatal(appjs + " should NOT have data.")
	}
}

// the command should work despite existing artifacts
func TestWhiteList(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	files := []string{".project", ".cw-settings", ".cw-extension"}
	vscode := ".vscode"
	metadata := ".metadata"

	for _, file := range files {
		createAndStat(filepath.Join(sandbox.ProjectDir, file), t)
	}
	err := os.MkdirAll(vscode, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(metadata, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	args := []string{"init", "nodejs-express"}
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	filesExist([]string{vscode}, sandbox.ProjectDir, true, t)
	expressExist(sandbox.ProjectDir, true, t)
}

func filesExist(files []string, projectDir string, expected bool, t *testing.T) {
	for _, file := range files {
		exist, err := cmdtest.Exists(filepath.Join(projectDir, file))
		if err != nil {
			t.Fatal(err)
		}
		if exist != expected {
			if expected {
				t.Fatal(file, " should exist but doesn't.")
			} else {
				t.Fatal(file, " should not exist without overwrite.")
			}
		}
	}
}

func expressExist(projectDir string, expected bool, t *testing.T) {
	filesExist([]string{appsodyFile, appjs, packagejson, packagejsonlock}, projectDir, expected, t)
}

func createAndStat(file string, t *testing.T) {
	_, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
}
