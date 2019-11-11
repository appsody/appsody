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
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type listCommandConfig struct {
	*RootCommandConfig
	output string
}

func newListCmd(rootConfig *RootCommandConfig) *cobra.Command {
	listConfig := &listCommandConfig{RootCommandConfig: rootConfig}
	// listCmd represents the list command
	var listCmd = &cobra.Command{
		Use:   "list [repository]",
		Short: "List the available Appsody stacks.",
		Long: `List all the Appsody stacks available in your repositories. 

An asterisk in the repository column denotes the default repository. An asterisk in the template column denotes the default template used, when you initialise an Appsody project.`,
		Example: `  appsody list
  Lists all available stacks for each of your repositories.
  
  appsody list my-repo
  Lists available stacks only in your "my-repo" repository.`,
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

				if listConfig.output == "" {
					Info.log("\n", projects)
					return nil
				}

				list, err := repos.getRepositories()
				if err != nil {
					return err
				}

				if listConfig.output == "yaml" {
					bytes, err := yaml.Marshal(&list)
					if err != nil {
						return err
					}
					result := string(bytes)
					Info.log("\n", result)
				} else if listConfig.output == "json" {
					bytes, err := json.Marshal(&list)
					if err != nil {
						return err
					}
					result := string(bytes)
					Info.log("\n", result)
				}

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

				if listConfig.output == "" {
					Info.log("\n", repoProjects)
					return nil
				}

				repoList, err := repos.getRepository(repoName)
				if err != nil {
					return err
				}

				if listConfig.output == "yaml" {
					bytes, err := yaml.Marshal(&repoList)
					if err != nil {
						return err
					}
					result := string(bytes)
					Info.log("\n", result)
				} else if listConfig.output == "json" {
					bytes, err := json.Marshal(&repoList)
					if err != nil {
						return err
					}
					result := string(bytes)
					Info.log("\n", result)
				}
			}

			return nil
		},
	}

	listCmd.PersistentFlags().StringVarP(&listConfig.output, "output", "o", "", "Output list in yaml or json format")
	return listCmd
}
