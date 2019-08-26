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
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// repo list represent repo list cmd
var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured Appsody repositories",
	Long:  `List configured Appsody repositories. An asterisk denotes the default repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var repos RepositoryFile
		setupErr := setupConfig()
		if setupErr != nil {
			return setupErr
		}
		list, repoErr := repos.getRepos()
		if repoErr != nil {
			return repoErr
		}
		repoList, err := repos.listRepos()
		if err != nil {
			return err
		}

		if output == "" {
			Info.log("\n", repoList)
		} else if output == "yaml" {
			result := executeMarshal(&list, yaml.Marshal)
			Info.log("\n", result)
		} else if output == "json" {
			result := executeMarshal(&list, json.Marshal)
			Info.log("\n", result)
		}
		return nil
	},
}

func executeMarshal(r interface{}, marshalImpl func(v interface{}) ([]byte, error)) string {
	bytes, err := marshalImpl(r)
	if err != nil {
		Error.log("Could not marshal repository", err)
	}
	return string(bytes)
}

func init() {
	repoCmd.AddCommand(repoListCmd)
}
