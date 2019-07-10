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
	"log"
	"os"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var addCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add an Appsody repository",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			Error.log("Error, you must specify repository name and URL")
			os.Exit(1)
		}
		var repoName = args[0]
		var repoURL = args[1]
		if dryrun {
			Info.logf("Dry Run - Skipping appsody repo add repository Name: %s, URL: %s", repoName, repoURL)
		} else {
			var _, err = downloadIndex(repoURL)
			if err != nil {
				Error.log(err)
				os.Exit(1)
			}

			var newEntry = RepositoryEntry{
				Name: repoName,
				URL:  repoURL,
			}

			// Need to check to see if it already exists under a different name?
			var repoFile RepositoryFile
			repoFile.getRepos()
			repoFile.Add(&newEntry)
			err = repoFile.WriteFile(getRepoFileLocation())
			if err != nil {
				log.Fatalf("Failed to write file to repository location: %v", err)
			}
		}
	},
}

func init() {
	repoCmd.AddCommand(addCmd)

}
