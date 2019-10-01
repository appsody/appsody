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
	"github.com/spf13/cobra"
)

func newRepoListCmd(config *RootCommandConfig) *cobra.Command {
	// repo list represent repo list cmd
	var repoListCmd = &cobra.Command{
		Use:   "list",
		Short: "List configured Appsody repositories",
		Long:  `List configured Appsody repositories. An asterisk denotes the default repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var repos RepositoryFile

			_, repoErr := repos.getRepos(config)
			if repoErr != nil {
				return repoErr
			}
			repoList, err := repos.listRepos(config)
			if err != nil {
				return err
			}
			Info.log("\n", repoList)
			return nil
		},
	}
	return repoListCmd
}
