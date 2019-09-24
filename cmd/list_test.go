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

func TestList(t *testing.T) {

	// tests that would have run before this and crashed could leave the repo
	// in a bad state - mostly leading to: "a repo with this name already exists."
	// so clean it up pro-actively, ignore any errors.
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"repo", "remove", "LocalTestRepo"}, ".")

	// first add the test repo index
	_, cleanup, err := cmdtest.AddLocalFileRepo("LocalTestRepo", "../cmd/testdata/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	output, err := cmdtest.RunAppsodyCmdExec([]string{"list"}, ".")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output, "A Java Microprofile Stack") {
		t.Error("list command should not display the stack name")
	}

	if !strings.Contains(output, "java-microprofile") {
		t.Error("list command should contain id 'java-microprofile'")
	}
}

// test the v2 list functionality
func TestListV2(t *testing.T) {
	// first add the test repo index
	var err error
	var output string
	var cleanup func()
	_, _ = cmdtest.RunAppsodyCmdExec([]string{"repo", "remove", "incubatortest"}, ".")
	_, cleanup, err = cmdtest.AddLocalFileRepo("incubatortest", "../cmd/testdata/kabanero.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	output, _ = cmdtest.RunAppsodyCmdExec([]string{"list", "incubatortest"}, ".")

	if !(strings.Contains(output, "nodejs") && strings.Contains(output, "incubatortest")) {
		t.Error("list command should contain id 'nodejs'")
	}

	// test the current default hub
	output, _ = cmdtest.RunAppsodyCmdExec([]string{"list", "appsodyhub"}, ".")

	if !strings.Contains(output, "java-microprofile") {
		t.Error("list command should contain id 'java-microprofile'")
	}

	output, _ = cmdtest.RunAppsodyCmdExec([]string{"list", "appsodyhub"}, ".")

	// we expect 2 instances
	if !(strings.Count(output, "java-microprofile") == 1) {
		t.Error("list command should contain id 'java-microprofile'")
	}
	output, _ = cmdtest.RunAppsodyCmdExec([]string{"list"}, ".")

	// we expect 2 instances
	if !(strings.Contains(output, "java-microprofile") && (strings.Count(output, "nodejs ") == 2)) {
		t.Error("list command should contain id 'java-microprofile and 2 nodejs '")
	}

	// test the current default hub
	output, _ = cmdtest.RunAppsodyCmdExec([]string{"list", "nonexisting"}, ".")

	if !(strings.Contains(output, "cannot locate repository ")) {
		t.Error("Failed to flag non-existing repo")
	}

}

func TestRepoJson(t *testing.T) {
	args := []string{"list", "-o", "json"}
	output, err := cmdtest.RunAppsodyCmdExec(args, ".")
	if err != nil {
		t.Fatal(err)
	}

	list, err := cmdtest.ParseListJSON(cmdtest.ParseJSON(output))
	if err != nil {
		t.Fatal(err)
	}
	if list[0].ID != "java-microprofile" {
		t.Error("list command should contain id 'java-microprofile'", list[0].ID, output)
	}
}
