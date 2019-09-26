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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newListCmd(rootConfig *RootCommandConfig) *cobra.Command {
	// listCmd represents the list command
	var listCmd = &cobra.Command{
		Use:   "list [repository]",
		Short: "List the Appsody stacks available to init",
		Long:  `This command lists all the stacks available in your repositories. If you omit the  optional [repository] parameter, the stacks for all the repositories are listed. If you specify the repository name [repository], only the stacks in that repository will be listed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var repos RepositoryFile

			if _, err := repos.getRepos(rootConfig); err != nil {
				return err
			}
			//var index RepoIndex
			if len(args) < 1 {
				projects, err := repos.listProjects(rootConfig)
				if err != nil {
					return errors.Errorf("%v", err)
				}
				if len(rootConfig.UnsupportedRepos) > 0 {
					Warning.log("The following repositories .yaml have an  APIVersion greater than "+supportedIndexAPIVersion+" which your installed Appsody CLI supports, it is strongly suggested that you update your Appsody CLI to the latest version: ", rootConfig.UnsupportedRepos)
				}
				Info.log("\n", projects)
			} else {
				repoName := args[0]
				_, err := repos.getRepos(rootConfig)
				if err != nil {
					return err
				}
				repoProjects, err := repos.listRepoProjects(repoName, rootConfig)
				if err != nil {
					return err
				}
				if len(rootConfig.UnsupportedRepos) > 0 {
					Warning.log("The following repositories are of APIVersion greater than "+supportedIndexAPIVersion+" which your installed Appsody CLI supports, it is strongly suggested that you update your Appsody CLI to the latest version: ", rootConfig.UnsupportedRepos)
				}

				Info.log("\n", repoProjects)
			}

			return nil
		},
	}
	return listCmd
}
