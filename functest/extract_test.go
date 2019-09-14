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
	"log"
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

	log.Println("stacksList is: ", stacksList)

	// if stacksList is empty there is nothing to test so return
	if stacksList == "" {
		log.Println("stacksList is empty, exiting test...")
		return
	}

	// replace incubator with appsodyhub to match current naming convention for repos
	stacksList = strings.Replace(stacksList, "incubator", "appsodyhub", -1)

	// split the appsodyStack env variable
	stackRaw := strings.Split(stacksList, " ")

	// loop through the stacks
	for i := range stackRaw {

		// create a temporary dir to create the project
		projectDir, err := ioutil.TempDir("", "appsody-extract-test-"+strings.ReplaceAll(stackRaw[i], "/", "_"))
		if err != nil {
			t.Fatal(err)
		}

		defer os.RemoveAll(projectDir)
		log.Println("Created project dir: " + projectDir)

		// create a temporary dir to extract the project, sibling to projectDir
		parentDir := filepath.Dir(projectDir)
		if err != nil {
			t.Fatal(err)
		}
		extractDir := parentDir + "/appsody-extract-test-extract-" + strings.ReplaceAll(stackRaw[i], "/", "_")

		defer os.RemoveAll(extractDir)
		log.Println("Created extraction dir: " + extractDir)

		// appsody init inside projectDir
		log.Println("Now running appsody init...")
		_, err = cmdtest.RunAppsodyCmd([]string{"init", stackRaw[i]}, projectDir)
		if err != nil {
			t.Fatal(err)
		}

		// appsody extract: running in projectDir, extracting into extractDir
		log.Println("Now running appsody extract...")
		_, err = cmdtest.RunAppsodyCmd([]string{"extract", "--target-dir", extractDir, "-v"}, projectDir)
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

		err = os.Chdir(projectDir)
		if err != nil {
			t.Fatal("Error changing directory: ", err)
		}
		mounts, _ := cmd.GetEnvVar("APPSODY_MOUNTS")
		pDir, _ := cmd.GetEnvVar("APPSODY_PROJECT_DIR")

		err = os.Chdir(oldDir)
		if err != nil {
			t.Fatal("Error changing directory: ", err)
		}

		log.Println("Stack mounts:", mounts)
		if pDir == "" {
			pDir = "/project"
		}
		log.Println("Stack's project dir:", pDir)

		mountlist := strings.Split(mounts, ";")
		for _, mount := range mountlist {
			log.Println("mount:", mount)
			src := strings.Split(mount, ":")[0]
			dest := strings.Split(mount, ":")[1]
			if !strings.HasPrefix(dest, pDir) {
				log.Println("Skipping un-extracted content", src, "and", dest)
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
				local = projectDir + "/" + src
			}
			log.Println("local: ", local)
			log.Println("remote: ", remote)

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
					log.Println("Folder contents match.")
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
					log.Println("File sizes match.")
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

// private exists fn.
// TODO: Can this be re-used from cmd/utils.go?
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
