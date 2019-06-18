// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test your project in the local appsody environment",
	Long:  `This starts a docker container for your project and runs your test in it.`,
	Run: func(cmd *cobra.Command, args []string) {

		Info.log("Running test environment")
		commonCmd(cmd, args, "test")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	addDevCommonFlags(testCmd)

}
