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

// debug Cmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run the local Appsody environment in debug mode",
	Long:  `This starts a docker based continuous build environment for your project with debugging enabled.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		Info.log("Running debug environment")
		return commonCmd(cmd, args, "debug")
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	addDevCommonFlags(debugCmd)

}
