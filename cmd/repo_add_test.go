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
package cmd_test

import (
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestRepoAdd(t *testing.T) {
	// see how many repos we currently have
	output, err := cmdtest.RunAppsodyCmdExec([]string{"repo", "list"}, ".")
	if err != nil {
		t.Fatal(err)
	}
	startRepos := cmdtest.ParseRepoList(output)

	addRepoName := "LocalTestRepo"
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"repo", "remove", addRepoName}, ".")
	addRepoURL, cleanup, err := cmdtest.AddLocalFileRepo(addRepoName, "testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	output, err = cmdtest.RunAppsodyCmdExec([]string{"repo", "list"}, ".")
	if err != nil {
		t.Fatal(err)
	}
	endRepos := cmdtest.ParseRepoList(output)

	if len(endRepos) != (len(startRepos) + 1) {
		t.Errorf("Expected %d repos but found %d", (len(startRepos) + 1), len(endRepos))
	} else {
		// check that the correct repo name and url were added
		found := false
		for _, repo := range endRepos {
			if repo.Name == addRepoName && repo.URL == addRepoURL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected repo with name '%s' and url '%s'", addRepoName, addRepoURL)
		}
	}
}

var repoAddErrorTests = []struct {
	testName      string
	args          []string // input
	expectedError string   // expected to be in the error message
}{
	{"No args", nil, "you must specify repository name and URL"},
	{"One arg", []string{"reponame"}, "you must specify repository name and URL"},
	{"No url scheme", []string{"test", "localhost"}, "unsupported protocol scheme"},
	{"Non-existing url", []string{"test", "http://localhost/doesnotexist"}, "refused"},
}

func TestRepoAddErrors(t *testing.T) {
	for _, tt := range repoAddErrorTests {
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			args := append([]string{"repo", "add"}, tt.args...)
			output, err := cmdtest.RunAppsodyCmdExec(args, ".")

			if err == nil {
				t.Error("Expected non-zero exit code")
			}
			if !strings.Contains(output, tt.expectedError) {
				t.Errorf("Did not find expected error '%s' in output", tt.expectedError)
			}
		})

	}
}
