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

var (
	output string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [repository]",
	Short: "List the Appsody stacks available to init",
	Long:  `This command lists all the stacks available in your repositories. If you omit the  optional [repository] parameter, the stacks for all the repositories are listed. If you specify the repository name [repository], only the stacks in that repository will be listed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var repos RepositoryFile

		setupErr := setupConfig()
		if setupErr != nil {
			return setupErr
		}

		if _, err := repos.getRepos(); err != nil {
			return err
		}
		//var index RepoIndex
		if len(args) < 1 {
			projects, err := repos.listProjects()
			if err != nil {
				return errors.Errorf("%v", err)
			}
			if len(unsupportedRepos) > 0 {
				Warning.log("The following repositories .yaml have an  APIVersion greater than "+supportedIndexAPIVersion+" which your installed Appsody CLI supports, it is strongly suggested that you update your Appsody CLI to the latest version: ", unsupportedRepos)
			}
			if output == "" {
				Info.log("\n", projects)
			} else {
				list, err := values(repos.GetIndices)
				if err != nil {
					return err
				}
				if output == "yaml" {
					result := executeMarshal(&list, yaml.Marshal)
					Info.log("\n", result)
				} else if output == "json" {
					result := executeMarshal(&list, json.Marshal)
					Info.log("\n", result)
				}
			}

		} else {
			repoName := args[0]
			_, err := repos.getRepos()
			if err != nil {
				return err
			}
			repoProjects, err := repos.listRepoProjects(repoName)
			if err != nil {
				return err
			}
			if len(unsupportedRepos) > 0 {
				Warning.log("The following repositories are of APIVersion greater than "+supportedIndexAPIVersion+" which your installed Appsody CLI supports, it is strongly suggested that you update your Appsody CLI to the latest version: ", unsupportedRepos)
			}

			Info.log("\n", repoProjects)
		}

		return nil
	},
}

func values(indicesFunc func() (RepoIndices, error)) ([]Stack, error) {
	var stacks []Stack
	indices, err := indicesFunc()
	if err != nil {
		return nil, errors.Errorf("Could not read indices: %v", err)
	}

	if len(indices) != 0 {
		for repoName, index := range indices {
			var errStack error
			stacks, errStack = index.buildStacksFromIndex(repoName, stacks)
			if errStack != nil {
				return nil, errStack
			}
		}
	}
	return stacks, nil
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output in another type yaml or json")
}
