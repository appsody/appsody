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
	"errors"
	"github.com/spf13/cobra"
)

func newVersionCmd(log *LoggingConfig, rootCmd *cobra.Command) *cobra.Command {
	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version of the Appsody CLI.",
		Long:  `Show the version of the Appsody CLI that is currently in use.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("Unexpected argument. Use 'appsody [command] --help' for more information about a command")
			}
			log.Info.log(rootCmd.Use, " ", VERSION)
			return nil
		},
	}

	return versionCmd
}
