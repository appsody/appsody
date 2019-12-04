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

var repoSetDefaultLogsTests = []struct {
	testName     string
	args         []string // input
	expectedLogs string   // expected to be in the error message
}{
	{"Valid args", []string{"localhub", "--config", "testdata/multiple_repository_config/config.yaml"}, "default repository is now set to localhub"},
	{"No args", nil, "you must specify desired default repository"},
	{"Existing default repo", []string{"incubator"}, "default repository has already been set to"},
	{"Non-existing repo", []string{"test"}, "not in configured list of repositories"},
	{"Badly formatted repo config", []string{"test", "--config", "testdata/bad_format_repository_config/config.yaml"}, "Failed to parse repository file yaml"},
}

func TestRepoSetDefaultLogs(t *testing.T) {
	for _, tt := range repoSetDefaultLogsTests {
		// call t.Run so that we can name and report on individual tests
		t.Run(tt.testName, func(t *testing.T) {
			args := append([]string{"repo", "set-default"}, tt.args...)
			output, _ := cmdtest.RunAppsodyCmd(args, ".", t)

			if !strings.Contains(output, tt.expectedLogs) {
				t.Errorf("Did not find expected error '%s' in output", tt.expectedLogs)
			}
		})
	}
}
