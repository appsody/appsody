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
	"regexp"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var addCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add an Appsody repository",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {

			return errors.New("Error, you must specify repository name and URL")
		}
		var repoName = args[0]
		var repoURL = args[1]

		if len(repoName) > 50 {
			return errors.Errorf("Invalid repository name. The <name> must be less than 50 characters")

		}
		match, _ := regexp.MatchString("^[a-zA-Z0-9\\-_]{1,50}$", repoName)
		if !match {
			return errors.Errorf("Invalid repository name. The <name> may only contain digits, numbers, dashes '-', and underscores '_'.")

		}

		var repoFile RepositoryFile
		repoFile.getRepos()
		if repoFile.Has(repoName) {
			return errors.Errorf("A repository with the name '%s' already exists.", repoName)

		}
		if repoFile.HasURL(repoURL) {
			return errors.Errorf("A repository with the URL '%s' already exists.", repoURL)

		}
		_, err := downloadIndex(repoURL)
		if err != nil {

			return err
		}

		if dryrun {
			Info.logf("Dry Run - Skipping appsody repo add repository Name: %s, URL: %s", repoName, repoURL)
		} else {
			var newEntry = RepositoryEntry{
				Name: repoName,
				URL:  repoURL,
			}

			repoFile.Add(&newEntry)
			err = repoFile.WriteFile(getRepoFileLocation())
			if err != nil {
				return errors.Errorf("Failed to write file to repository location: %v", err)
			}
		}
		return nil
	},
}

func init() {
	repoCmd.AddCommand(addCmd)

}
