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
	"github.com/appsody/appsody/cmd/cmdtest"
	"testing"
)

func TestLintWithValidStack(t *testing.T) {
	args := []string{"stack", "lint"}

	_, err := cmdtest.RunAppsodyCmdExec(args, "../cmd/testData/sample-stack")

	if err != nil { //Lint check should pass, if not fail the test
		t.Fatal(err)
	}
}

func TestLintWithBadStack(t *testing.T) {
	args := []string{"stack", "lint"}

	_, err := cmdtest.RunAppsodyCmdExec(args, "../cmd/testData/bad-sample-stack")

	if err == nil { //Lint check should fail, if not fail the test
		t.Fatal(err)
	}
}
