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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	cmdtest "github.com/appsody/appsody/cmd/cmdtest"
)

// Simple test for appsody extract command.
// A future enhancement is to extract with buildah.
// I could have very well taken one single stack
// and test, but due to the fact that most pain points
// come from diverse mounts that exist for different
// stacks, it is nice to loop through each and make sure
// nothing is broken.

func TestExtract(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	stacksList := cmdtest.GetEnvStacksList()

	// z and p use locally packaged dev.local so we need to add it to the config of the sandbox for it to work
	if stacksList == "dev.local/starter" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		devlocal := filepath.Join(home, ".appsody", "stacks", "dev.local", "dev.local-index.yaml")
		devlocalPath := "file://" + devlocal
		_, err = cmdtest.RunAppsody(sandbox, "repo", "add", "dev.local", devlocalPath)
		if err != nil {
			t.Fatal(err)
		}
	}

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		// create a temporary dir to extract the project, sibling to projectDir
		parentDir := filepath.Dir(sandbox.ProjectDir)

		extractDir := parentDir + "/appsody-extract-test-extract-" + strings.ReplaceAll(stackRaw[i], "/", "_")

		t.Log("Created extraction dir: " + extractDir)

		// appsody init inside projectDir
		t.Log("Now running appsody init...")
		args := []string{"init", stackRaw[i]}
		_, err := cmdtest.RunAppsody(sandbox, args...)
		if err != nil {
			t.Fatal(err)
		}

		// appsody extract: running in projectDir, extracting into extractDir
		t.Log("Now running appsody extract...")
		args = []string{"extract", "--target-dir", extractDir, "-v"}
		_, err = cmdtest.RunAppsody(sandbox, args...)
		if err != nil {
			t.Fatal(err)
		}

		// Main extraction test logic:
		// 1. Get the environment variables from the images that was
		// just extracted.
		// 2. For each mount that is a file, make sure the source and
		// destination file sizes match.
		// 3. For each mount that is a folder, make sure the source and
		// destination folders have same content (file name match)
		// 4. Skip mounts that were not extracted (e.g:- /mvn/...)
		// 5. For all cases, make sure we have a Dockerfile extracted
		// at the vortex of the extraction.

		var outBuffer bytes.Buffer
		loggingConfig := &cmd.LoggingConfig{}
		loggingConfig.InitLogging(&outBuffer, &outBuffer)
		config := &cmd.RootCommandConfig{LoggingConfig: loggingConfig}

		config.ProjectDir = sandbox.ProjectDir

		err = cmd.InitConfig(config)
		if err != nil {
			t.Fatal("Could not init appsody config", err)
		}
		mounts, _ := cmd.GetEnvVar("APPSODY_MOUNTS", config)
		pDir, _ := cmd.GetEnvVar("APPSODY_PROJECT_DIR", config)
		t.Log(outBuffer.String())

		t.Log("Stack mounts:", mounts)
		if pDir == "" {
			pDir = "/project"
		}
		t.Log("Stack's project dir:", pDir)

		mountlist := strings.Split(mounts, ";")
		for _, mount := range mountlist {
			t.Log("mount:", mount)
			src := strings.Split(mount, ":")[0]
			dest := strings.Split(mount, ":")[1]
			if !strings.HasPrefix(dest, pDir) {
				t.Log("Skipping un-extracted content", src, "and", dest)
				continue
			}
			remote := strings.Replace(dest, pDir, extractDir, -1)
			var local string
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				t.Fatal("Unable to find user home location:", homeErr)
			}
			if strings.HasPrefix(src, "~") {
				local = strings.Replace(src, "~", homeDir, 1)
			} else {
				local = sandbox.ProjectDir + "/" + src
			}
			t.Log("local: ", local)
			t.Log("remote: ", remote)

			fileInfoLocal, err := os.Lstat(local)
			if err != nil {
				t.Fatal("Mount inspection error:", err)
			}
			if fileInfoLocal.IsDir() {
				localData, err := ioutil.ReadDir(local)
				if err != nil {
					t.Fatal("Mount inspection error:", err)
				}
				extractData, err := ioutil.ReadDir(remote)
				if err != nil {
					t.Fatal("Mount inspection error:", err)
				}
				localContent := []string{}
				extractContent := []string{}
				for _, file := range localData {
					localContent = append(localContent, file.Name())
				}
				for _, file := range extractData {
					extractContent = append(extractContent, file.Name())
				}
				if !reflect.DeepEqual(extractContent, localContent) {
					t.Fatal("Extraction failure, ", local, " is not extracted into ", remote)
				} else {
					t.Log("Folder contents match.")
				}
			} else {
				fileInfoRemote, err := os.Lstat(remote)
				if err != nil {
					t.Fatal("Mount inspection error:", err)
				}
				dSize := fileInfoRemote.Size()
				lSize := fileInfoLocal.Size()
				if lSize != dSize {
					t.Fatal("Extraction failure, ", local, " is not extracted into ", remote, " properly: source file size: ", lSize, " destination file size: ", dSize)
				} else {
					t.Log("File sizes match.")
				}

			}

		}
		dockerFile := filepath.Join(extractDir, "Dockerfile")
		_, err = cmd.Exists(dockerFile)
		if err != nil {
			t.Fatal("Extraction failure, Dockerfile was not extracted into ", extractDir)
		}
	}
}

func TestExtractCases(t *testing.T) {
	var extractTests = []struct {
		testName     string
		args         []string
		expectedLogs string
	}{
		{"Non existing target directory", []string{"--target-dir", "/non/existing/dir"}, "/non/existing does not exist"},
		{"Target dir with contents", []string{"--target-dir", "."}, "Cannot extract to an existing target-dir"},
	}

	for _, testData := range extractTests {
		tt := testData

		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			t.Log("Now running appsody init...")
			args := []string{"init", "starter"}
			_, err := cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatal(err)
			}

			t.Log("Now running appsody extract...")
			extractArgs := append([]string{"extract"}, tt.args...)
			output, extractErr := cmdtest.RunAppsody(sandbox, extractArgs...)

			if !strings.Contains(output, tt.expectedLogs) {
				t.Fatalf("Expected failure to include: %s but instead receieved: %s. Full error: %s", tt.expectedLogs, output, extractErr)
			}
		})
	}
}

// check if id exists in .appsody-config.yaml but not in project.yaml, a new project entry in project.yaml gets created with the same id
func TestExtractIfProjectIDNotExistInProjectYaml(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)
	config := &cmd.RootCommandConfig{LoggingConfig: loggingConfig}

	p, _, _ := getCurrentProjectEntry(t, sandbox, config)
	projectsBefore := len(p.Projects)

	err = cmd.SaveIDToConfig("newRandomID", config)
	if err != nil {
		t.Fatal(err)
	}
	args = []string{"extract", "--dryrun"}
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	p, project, configID := getCurrentProjectEntry(t, sandbox, config)
	projectsAfter := len(p.Projects)

	if projectsBefore+1 != projectsAfter {
		t.Fatalf("Expected number of project entries to be %v but found %v", projectsBefore+1, projectsAfter)
	}
	if project.ID != configID {
		t.Fatalf("Expected project id in .appsody-config.yaml to have a valid project entry in project.yaml.")
	}
}

// check if id does not exists in .appsody-config.yaml, a new project entry in project.yaml gets created with the same id
func TestExtractIfProjectIDNotExistInConfigYaml(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	args := []string{"init", "nodejs"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}
	config := new(cmd.RootCommandConfig)

	p, _, configID := getCurrentProjectEntry(t, sandbox, config)
	projectsBefore := len(p.Projects)

	// delete id from .appsody-config.yaml
	appsodyConfig := filepath.Join(sandbox.ProjectDir, cmd.ConfigFile)
	data, err := ioutil.ReadFile(appsodyConfig)
	if err != nil {
		t.Fatal(err)
	}
	removedID := bytes.Replace(data, []byte("id: \""+configID+"\""), []byte(""), 1)
	err = ioutil.WriteFile(appsodyConfig, []byte(removedID), 0666)
	if err != nil {
		t.Fatal(err)
	}

	args = []string{"extract", "--dryrun"}
	_, err = cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	p, project, configID := getCurrentProjectEntry(t, sandbox, config)
	projectsAfter := len(p.Projects)

	if projectsBefore+1 != projectsAfter {
		t.Fatalf("Expected number of project entries to be %v but found %v", projectsBefore+1, projectsAfter)
	}
	if project.ID != configID {
		t.Fatalf("Expected project id in .appsody-config.yaml to have a valid project entry in project.yaml.")
	}
}
