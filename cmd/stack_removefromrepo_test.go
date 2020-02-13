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

func TestRemoveFromRepoDefault(t *testing.T) {

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// run stack package
	args := []string{"stack", "remove-from-repo", "incubator", "nodejs"}
	output, err := cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		if !strings.Contains(err.Error(), "Error creating templating mal: Variable name didn't start with alphanumeric character") {
			t.Errorf("String \"Error creating templating mal: Variable name didn't start with alphanumeric character\" not found in output: '%v'", err.Error())
		}

	} else {
		t.Fatal(output)
	}

}
