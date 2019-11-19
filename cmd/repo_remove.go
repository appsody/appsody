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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newRepoRemoveCmd(config *RootCommandConfig) *cobra.Command {
	// initCmd represents the init command
	var removeCmd = &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an Appsody repository.",
		Long: `Remove an Appsody repository from your list of configured Appsody repositories.
		
You cannot remove the default repository, but you can make a different repository the default (see appsody repo set-default).`,
		Example: `  appsody repo remove my-local-repo
  Removes the "my-local-repo" repository from your list of configured repositories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Error, you must specify repository name")
			}

			var repoName = args[0]

			var repoFile RepositoryFile
			_, repoErr := repoFile.getRepos(config)
			if repoErr != nil {
				return repoErr
			}
			if config.Dryrun {
				config.Info.log("Dry Run - Skipping appsody repo remove ", repoName)
			} else {
				if repoFile.Has(repoName) {
					defaultRepoName, err := repoFile.GetDefaultRepoName(config)
					if err != nil {
						return err
					}
					if repoName != defaultRepoName {
						repoFile.Remove(repoName)
					} else {
						config.Error.log("You cannot remove the default repository " + repoName)
					}
				} else {
					config.Error.log("Repository is not in configured list of repositories")
				}
				err := repoFile.WriteFile(getRepoFileLocation(config.CliConfig))
				if err != nil {
					log.Fatalf("Failed to write file repository location: %v", err)
				}
			}
			return nil
		},
	}
	return removeCmd
}
