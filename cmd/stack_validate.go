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
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// stack validate is a suite of validation tests for a local stack
// stack validate does the following...
// 1. stack lint test, can be turned off with --no-lint
// 2. stack package, can be turned off with --no-package
// 3. appsody init
// 4. appsody run
// 5. appsody test
// 6. appsody build

func newStackValidateCmd(rootConfig *RootCommandConfig) *cobra.Command {

	// vars for --no-package and --no-lint parms
	var noPackage bool
	var noLint bool

	var stackValidateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Run validation tests of a stack in the local Appsody environment",
		Long:  `This runs a set of validation tests for a stack.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			// vars to store test results
			var testResults []string
			failCount := 0
			passCount := 0

			stackPath := rootConfig.ProjectDir
			Info.Log("stackPath is: ", stackPath)

			// check for temeplates dir, error out if its not there
			err := os.Chdir("templates")
			if err != nil {
				// if we can't find the templates directory then we are not starting from a valid root of the stack directory
				Error.Log("Unable to reach templates directory. Current directory must be the root of the stack.")
				return err
			}

			// get the stack name and repo name from the stack path
			stackPathSplit := strings.Split(stackPath, string(filepath.Separator))
			stackName := stackPathSplit[len(stackPathSplit)-1]
			Info.Log("stackName is: ", stackName)

			repoName := stackPathSplit[len(stackPathSplit)-2]
			Info.Log("repoName is: ", repoName)

			Info.Log("#################################################")
			Info.Log("Validating stack: ", stackName)
			Info.Log("#################################################")

			// create a temporary dir to create the project and run the test
			projectDir, err := ioutil.TempDir("", "appsody-build-simple-test")
			if err != nil {
				return err
			}

			Info.Log("Created project dir: " + projectDir)

			// call tests...

			// lint
			if !noLint {
				_, err = RunAppsodyCmdExec([]string{"stack", "lint"}, stackPath)
				if err != nil {
					//logs error but keeps going
					Error.Log(err)
					testResults = append(testResults, ("FAILED: Lint for stack: " + stackName))
					failCount++
				} else {
					testResults = append(testResults, ("PASSED: Lint for stack: " + stackName))
					passCount++
				}
			}

			// package
			if !noPackage {
				_, err = RunAppsodyCmdExec([]string{"stack", "package"}, stackPath)
				if err != nil {
					//logs error but keeps going
					Error.Log(err)
					testResults = append(testResults, ("FAILED: Package for stack: " + stackName))
					failCount++
				} else {
					testResults = append(testResults, ("PASSED: Package for stack: " + stackName))
					passCount++
				}
			}

			// init
			err = TestInit("dev-local/"+stackName, projectDir)
			if err != nil {
				// quit everything if init fails as the other tests rely on init to succeed
				return err
			}

			testResults = append(testResults, ("PASSED: Init for stack: " + stackName))
			passCount++

			// run
			err = TestRun(projectDir)
			if err != nil {
				//logs error but keeps going
				Error.Log(err)
				testResults = append(testResults, ("FAILED: Run for stack: " + stackName))
				failCount++
			} else {
				testResults = append(testResults, ("PASSED: Run for stack: " + stackName))
				passCount++
			}

			// test
			err = TestTest(projectDir)
			if err != nil {
				//logs error but keeps going
				Error.Log(err)
				testResults = append(testResults, ("FAILED: Test for stack: " + stackName))
				failCount++
			} else {
				testResults = append(testResults, ("PASSED: Test for stack: " + stackName))
				passCount++
			}

			// build
			err = TestBuild(projectDir)
			if err != nil {
				//logs error but keeps going
				Error.Log(err)
				testResults = append(testResults, ("FAILED: Build for stack: " + stackName))
				failCount++
			} else {
				testResults = append(testResults, ("PASSED: Build for stack: " + stackName))
				passCount++
			}

			//cleanup
			Info.Log("Removing project dir: " + projectDir)
			os.RemoveAll(projectDir)

			//}

			Info.Log("@@@@@@@@@ Validate Summary Start @@@@@@@@@@")
			for i := range testResults {
				Info.Log(testResults[i])
			}
			Info.Log("Total PASSED: ", passCount)
			Info.Log("Total FAILED: ", failCount)
			Info.Log("@@@@@@@@@ Validate Summary End @@@@@@@@@@")

			return nil
		},
	}

	stackValidateCmd.PersistentFlags().BoolVar(&noPackage, "no-package", false, "Skips running appsody stack package")
	stackValidateCmd.PersistentFlags().BoolVar(&noLint, "no-lint", false, "Skips running appsody stack lint")

	return stackValidateCmd
}
