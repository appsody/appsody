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
	"os"
	"testing"

	"github.com/appsody/appsody/cmd/cmdtest"
)

func TestStackCreateSampleStack(t *testing.T) {
	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--config", "testdata/default_repository_config/config.yaml"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if !exists {
		t.Fatal(err)
	}
	os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackCreateWithCopyTag(t *testing.T) {
	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--config", "testdata/default_repository_config/config.yaml", "--copy", "incubator/nodejs"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if !exists {
		t.Fatal(err)
	}
	os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase1(t *testing.T) {
	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "incubator/nodej"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase2(t *testing.T) {
	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "nodejs"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase3(t *testing.T) {
	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "experimental/nodejs"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackCase4(t *testing.T) {
	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--copy", "exp/java-microprofile"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidStackName(t *testing.T) {
	err := os.RemoveAll("testing_stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing_stack"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing_stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackCreateInvalidLongStackName(t *testing.T) {
	args := []string{"stack", "create", "testing_stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stack"}
	_, err := cmdtest.RunAppsodyCmd(args, ".", t)

	if err == nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing_stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stacktesting-stack")

	if exists {
		t.Fatal(err)
	}
}

func TestStackAlreadyExists(t *testing.T) {
	err := os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"stack", "create", "testing-stack", "--config", "testdata/default_repository_config/config.yaml"}
	_, err = cmdtest.RunAppsodyCmd(args, ".", t)

	if err != nil {
		t.Fatal(err)
	}

	exists, err := cmdtest.Exists("testing-stack")

	if !exists {
		t.Fatal(err)
	}

	_, err1 := cmdtest.RunAppsodyCmd(args, ".", t)

	if err1 == nil {
		t.Fatal(err)
	}

	os.RemoveAll("testing-stack")
	if err != nil {
		t.Fatal(err)
	}
}
