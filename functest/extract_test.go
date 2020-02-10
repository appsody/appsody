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
	"runtime"
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

	t.Log("stacksList is: ", stacksList)

	// if stacksList is empty there is nothing to test so return
	if stacksList == "" {
		t.Log("stacksList is empty, exiting test...")
		return
	}

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		// create a temporary dir to extract the project, sibling to projectDir
		parentDir := filepath.Dir(sandbox.ProjectDir)

		extractDir := parentDir + "/appsody-extract-test-extract-" + strings.ReplaceAll(stackRaw[i], "/", "_")

		defer os.RemoveAll(extractDir)
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
		// just extracted - GetEnvVar. Switch the current directory
		// to meet the API's needs. (refactor it so that this isn't needed)
		// 2. For each mount that is a file, make sure the source and
		// destination file sizes match.
		// 3. For each mount that is a folder, make sure the source and
		// destination folders have same content (file name match)
		// 4. Skip mounts that were not extracted (e.g:- /mvn/...)
		// 5. For all cases, make sure we have a Dockerfile extracted
		// at the vortex of the extraction.

		oldDir, err := os.Getwd()
		if err != nil {
			t.Fatal("Error getting current directory: ", err)
		}

		err = os.Chdir(sandbox.ProjectDir)
		if err != nil {
			t.Fatal("Error changing directory: ", err)
		}
		var outBuffer bytes.Buffer
		loggingConfig := &cmd.LoggingConfig{}
		loggingConfig.InitLogging(&outBuffer, &outBuffer)
		config := &cmd.RootCommandConfig{LoggingConfig: loggingConfig}
		err = cmd.InitConfig(config)
		if err != nil {
			t.Fatal("Could not init appsody config", err)
		}
		mounts, _ := cmd.GetEnvVar("APPSODY_MOUNTS", config)
		pDir, _ := cmd.GetEnvVar("APPSODY_PROJECT_DIR", config)
		t.Log(outBuffer.String())

		err = os.Chdir(oldDir)
		if err != nil {
			t.Fatal("Error changing directory: ", err)
		}

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
		_, err = exists(dockerFile)
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
		{"Extract with Buildah", []string{"--buildah"}, "Project extracted"},
	}

	for _, testData := range extractTests {
		tt := testData
		if runtime.GOOS != "linux" && tt.testName == "Extract with Buildah" {
			t.Skip()
		}

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

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
