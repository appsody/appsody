// Copyright © 2020 IBM Corporation and others.
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
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
	cmdtest "github.com/appsody/appsody/cmd/cmdtest"
	"github.com/pkg/errors"
)

func TestOperatorInstallCases(t *testing.T) {
	var operatorInstallTests = []struct {
		testName     string
		args         []string
		expectedLogs string
	}{
		{"Run in dryrun mode", []string{"--dryrun"}, "Appsody operator deployed to Kubernetes"},
		{"Install with non existing namespace", []string{"--namespace", "nonexistingnamespace"}, "namespaces \"nonexistingnamespace\" not found"},
	}

	for _, testData := range operatorInstallTests {
		if !cmdtest.TravisTesting {
			t.Skip()
		}

		tt := testData
		t.Run(tt.testName, func(t *testing.T) {
			sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
			defer cleanup()

			t.Log("Now running appsody init")
			args := []string{"init", "starter"}
			_, err := cmdtest.RunAppsody(sandbox, args...)
			if err != nil {
				t.Fatal(err)
			}

			t.Log("Now running appsody operator install")
			operatorInstallArgs := append([]string{"operator", "install"}, tt.args...)
			output, operatorErr := cmdtest.RunAppsody(sandbox, operatorInstallArgs...)

			if !strings.Contains(output, tt.expectedLogs) {
				t.Fatalf("Expected failure to include: %s but instead receieved: %s. Full error: %s", tt.expectedLogs, output, operatorErr)
			}
		})
	}
}

func TestInstallOperatorWithNamespaceAndWatchspace(t *testing.T) {
	if !cmdtest.TravisTesting {
		t.Skip()
	}

	sandbox, cleanup := cmdtest.TestSetupWithSandbox(t, true)
	defer cleanup()

	var outBuffer bytes.Buffer
	log := &cmd.LoggingConfig{}
	log.InitLogging(&outBuffer, &outBuffer)

	defer removeNamespace(log, "namespace-for-test-operator-install")
	expectedLogs := "Appsody operator deployed to Kubernetes"

	namespaceKargs := []string{"create", "namespace", "namespace-for-test-operator-install"}
	_, namespaceErr := cmd.RunKube(log, namespaceKargs, false)
	if namespaceErr != nil {
		t.Fatal(namespaceErr)
	}

	t.Log("Now running appsody init")
	args := []string{"init", "starter"}
	_, err := cmdtest.RunAppsody(sandbox, args...)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Now running appsody operator install")
	operatorInstallArgs := []string{"operator", "install", "--namespace", "namespace-for-test-operator-install", "--watchspace", "testWatchspace"}
	output, operatorErr := cmdtest.RunAppsody(sandbox, operatorInstallArgs...)

	if !strings.Contains(output, expectedLogs) {
		t.Fatalf("Expected failure to include: %s but instead receieved: %s. Full error: %s", expectedLogs, output, operatorErr)
	}
}

func removeNamespace(log *cmd.LoggingConfig, namespace string) error {
	kargs := []string{"delete", "namespace", namespace}
	_, namespaceErr := cmd.RunKube(log, kargs, false)

	if namespaceErr != nil {
		return errors.Errorf("Error removing namespace created for test: %s", namespaceErr)
	}
	return nil
}
