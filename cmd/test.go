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

import "github.com/spf13/cobra"

func newTestCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &devCommonConfig{RootCommandConfig: rootConfig}
	// testCmd represents the test command
	var testCmd = &cobra.Command{
		Use:   "test",
		Short: "Test your project in the local Appsody environment",
		Long: `Test your project in the local Appsody environment, starting a container for your project, running the stack provided and user specified tests.
		
You must be in the base directory of your Appsody project when running this command.`,
		Example: `  appsody test --no-watcher
  Runs the tests for your Appsody project with file watching disabled. The process runs the tests, terminates and returns with the result of the executed command.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			Info.log("Running test environment")
			return commonCmd(config, "test")
		},
	}

	addDevCommonFlags(testCmd, config)
	return testCmd
}
