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
	"strings"

	"github.com/spf13/cobra"
)

// get the list of stacks
var stacksList = os.Getenv("STACKSLIST")

// stackValidateCmd represents the validate command
var stackValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run validation tests of a stack in the local Appsody environment",
	Long:  `This runs a set of validation tests for a stack.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		Info.log("Running test environment")
		Info.log("stacksList is: ", stacksList)

		// if stacksList is empty there is nothing to test so return
		if stacksList == "" {
			Error.log("STACKSLIST is empty")
		}

		// replace incubator with appsodyhub to match current naming convention for repos
		stacksList = strings.Replace(stacksList, "incubator", "appsodyhub", -1)

		// split the appsodyStack env variable
		stackRaw := strings.Split(stacksList, " ")

		// loop through the stacks, execute all the tests on each stack before moving on to the next one
		for i := range stackRaw {
			Info.log("#################################################")
			Info.log("Testing stack: ", stackRaw[i])
			Info.log("#################################################")

			// create a temporary dir to create the project and run the test
			projectDir, err := ioutil.TempDir("", "appsody-build-simple-test")
			if err != nil {
				return err
			}

			Info.log("Created project dir: " + projectDir)

			// call tests...

			// init
			err = TestInit(stackRaw[i], projectDir)
			if err != nil {
				// quit everything if init fails as the other tests rely on init to succeed
				return err
			}

			// run
			err = TestRun(projectDir)
			if err != nil {
				//logs error but keeps going
				Error.log(err)
			}

			// test
			err = TestTest(projectDir)
			if err != nil {
				//logs error but keeps going
				Error.log(err)
			}

			// build
			err = TestBuild(projectDir)
			if err != nil {
				//logs error but keeps going
				Error.log(err)

				//to quit
				//return err
			}

			//cleanup
			Info.log("Removing project dir: " + projectDir)
			os.RemoveAll(projectDir)

		}

		return nil
	},
}

func init() {
	// will use stackCmd eventually
	rootCmd.AddCommand(stackValidateCmd)

}
