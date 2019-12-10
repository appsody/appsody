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
	"path/filepath"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestPackage(t *testing.T) {
	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	// create a temporary dir to create the project and run the test
	_, err := cmdtest.AddLocalRepo(sandbox, "LocalTestRepo", filepath.Join("..", "cmd", "testdata", "starter"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"stack", "package"}
	_, err = cmdtest.RunAppsody(sandbox, args...)

	if err != nil {
		t.Fatal(err)
	}
}
