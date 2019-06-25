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

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a configured Appsody repository",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			Error.log("Error, you must specify repository name")
			os.Exit(1)
		}
		var repoName = args[0]

		var repoFile RepositoryFile
		repoFile.getRepos()
		if dryrun {
			Info.log("Dry Run - Skipping appsody repo remove ", repoName)
		} else {
			if repoFile.Has(repoName) {
				repoFile.Remove(repoName)
			} else {
				Error.log("Repository is not in configured list of repositories")
			}
			err := repoFile.WriteFile(getRepoFileLocation())
			if err != nil {
				log.Fatalf("Failed to write file repository location: %v", err)
			}
		}
	},
}

func init() {
	repoCmd.AddCommand(removeCmd)

}
