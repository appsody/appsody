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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

// requires clean dir
func TestInit(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express"}, projectDir)
	if err != nil {
		t.Fatal(err)
	}

	appsodyResultsCheck(projectDir, t)
}

//This test makes sure that no project creation occurred because app.js existed prior to the call
func TestNoOverwrite(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-nooverwrite-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	appsodyFile := filepath.Join(projectDir, ".appsody-config.yaml")

	appjs := filepath.Join(projectDir, "app.js")
	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	appjsPath := filepath.Join(projectDir, "app.js")
	_, err = os.Create(appjsPath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(appjs)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express"}, projectDir)

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
	projectDir, err := ioutil.TempDir("", "appsody-init-overwrite-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	appsodyFile := filepath.Join(projectDir, ".appsody-config.yaml")

	appjs := filepath.Join(projectDir, "app.js")
	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	appjsPath := filepath.Join(projectDir, "app.js")
	_, err = os.Create(appjsPath)
	if err != nil {
		t.Fatal(err)
	}
	//file should be 0 bytes
	_, err = os.Stat(appjs)
	if err != nil {
		t.Fatal(err)
	}

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express", "--overwrite"}, projectDir)

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
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-notemplate-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	appsodyFile := filepath.Join(projectDir, ".appsody-config.yaml")

	appjs := filepath.Join(projectDir, "app.js")
	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	appjsPath := filepath.Join(projectDir, "app.js")
	// file size should be 0 bytes
	_, err = os.Create(appjsPath)
	if err != nil {
		t.Fatal(err)
	}

	shouldExist(appjs, t)

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express", "--no-template"}, projectDir)

	shouldExist(appsodyFile, t)

	shouldNotExist(packagejson, t)

	shouldNotExist(packagejsonlock, t)

	fileInfoFinal, err = os.Stat(appjs)
	if err != nil {
		err = errors.New(appjs + " should exist without overwrite.")

		t.Fatal(err)
	}
	// if we accidently overwrite the size would be >0
	if fileInfoFinal.Size() != 0 {
		err = errors.New(appjs + " should NOT have data.")

		t.Fatal(err)
	}
}

// the command should work despite existing artifacts
func TestWhiteList(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-notemplate-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	appjs := filepath.Join(projectDir, "app.js")
	vscode := filepath.Join(projectDir, ".vscode")
	project := filepath.Join(projectDir, ".project")
	cwSet := filepath.Join(projectDir, ".cw-settings")
	cwExtension := filepath.Join(projectDir, ".cw-extension")
	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")
	metadata := filepath.Join(projectDir, ".metadata")
	appsodyFile := filepath.Join(projectDir, ".appsody-config.yaml")

	_, err = os.Create(project)
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
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express"}, projectDir)

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
		t.Fatal(err)
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
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-testDefaultRepo")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "appsodyhub/nodejs"}, projectDir)
	if err != nil {
		t.Error(err)
	}

	appsodyResultsCheck(projectDir, t)
}

func TestInitV2WithNonDefaultRepoSpecified(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-testNonDefaultRepo")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "experimental/nodejs-functions"}, projectDir)
	if err != nil {
		t.Error(err)
	}

}

func TestInitV2WithBadStackSpecified(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-testbadstack")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(projectDir)

	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	output, _ := cmdtest.RunAppsodyCmdExec([]string{"init", "badnodejs-express"}, projectDir)
	if !(strings.Contains(output, "Could not find a stack with the id")) {
		t.Error("Should have flagged non existing stack")
	}

}

func TestInitV2WithBadRepoSpecified(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-testBadRepo")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(projectDir)

	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	output, _ := cmdtest.RunAppsodyCmdExec([]string{"init", "badrepo/nodejs-express"}, projectDir)

	if !(strings.Contains(output, "is not in configured list of repositories")) {
		fmt.Println("Bad repo not flagged")
		t.Error("Bad repo not flagged")
	}

}

func TestInitV2WithDefaultRepoSpecifiedTemplateNonDefault(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-testDefaultRepoTemplateNonDefault")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "appsodyhub/nodejs-express", "skaffold"}, projectDir)
	if err != nil {
		t.Error(err)
	}

	appsodyResultsCheck(projectDir, t)
}

func TestInitV2WithDefaultRepoSpecifiedTemplateDefault(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-testDefaultRepoTemplateDefault")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "appsodyhub/nodejs-express", "simple"}, projectDir)
	if err != nil {
		t.Error(err)
	}

	appsodyResultsCheck(projectDir, t)
}

func TestInitV2WithNoRepoSpecifiedTemplateDefault(t *testing.T) {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-testDefaultRepoTemplateDefault")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	// appsody init nodejs-express
	_, err = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express", "simple"}, projectDir)
	if err != nil {
		t.Error(err)
	}

	appsodyResultsCheck(projectDir, t)
}
func TestNone(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-none-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express", "none"}, projectDir)

	shouldNotExist(packagejson, t)
	shouldNotExist(packagejsonlock, t)

}

func TestNoneAndNoTemplate(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-none-and-no-template-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express", "none", "--no-template"}, projectDir)

	shouldNotExist(packagejson, t)
	shouldNotExist(packagejsonlock, t)

}

func TestNoTemplateOnly(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-no-template-only-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)
	log.Println("Created project dir: " + projectDir)

	packagejson := filepath.Join(projectDir, "package.json")
	packagejsonlock := filepath.Join(projectDir, "package-lock.json")

	// appsody init nodejs-express
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express", "--no-template"}, projectDir)

	shouldNotExist(packagejson, t)
	shouldNotExist(packagejsonlock, t)

}

func TestNoTemplateAndSimple(t *testing.T) {

	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-init-no-template-only-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(projectDir)

	// appsody init nodejs-express
	var output string
	output, _ = cmdtest.RunAppsodyCmdExec([]string{"init", "nodejs-express", "simple", "--no-template"}, projectDir)
	if !strings.Contains(output, "with both a template and --no-template") {
		t.Error("Correct error message not given")
	}
}
