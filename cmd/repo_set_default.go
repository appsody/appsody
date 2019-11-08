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

func newRepoDefaultCmd(config *RootCommandConfig) *cobra.Command {
	// initCmd represents the init command
	var setDefaultCmd = &cobra.Command{
		Use:   "set-default <name>",
		Short: "Sets a default repository.",
		Long: `Sets your specified repository to be the default repository. The "incubator" repository is the initial default repository you will get upon installing Appsody. The default repository cannot be removed from you list of configured Appsody repositories. To remove a default repository, you would first have to set a different default repository.

The default repository is denoted by an asterisk when you run "appsody repo list" or "appsody list". As a result of setting this default repository, you will not have to specify the repository, when initialising a new project using a stack in that repository.`,
		Example: `  appsody repo set-default my-local-repo
  Sets your default repository to "my-local-repo".`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Error, you must specify desired default repository")
			}

			var repoName = args[0]

			var repoFile RepositoryFile
			_, repoErr := repoFile.getRepos(config)
			if repoErr != nil {
				return repoErr
			}
			if config.Dryrun {
				Info.log("Dry Run - Skipping appsody repo set-default ", repoName)
			} else {
				if repoFile.Has(repoName) {
					defaultRepoName, err := repoFile.GetDefaultRepoName(config)
					if err != nil {
						return err
					}
					if repoName != defaultRepoName {
						_, repoFileErr := repoFile.SetDefaultRepoName(repoName, defaultRepoName, config)
						if repoFileErr != nil {
							return repoFileErr
						}
					} else {
						Info.log("Your default repository has already been set to " + repoName)
					}
				} else {
					Error.log("Repository is not in configured list of repositories")
				}
				err := repoFile.WriteFile(getRepoFileLocation(config))
				if err != nil {
					log.Fatalf("Failed to write file repository location: %v", err)
				}
			}
			return nil
		},
	}

	return setDefaultCmd
}
