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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [repo]",
	Short: "List the Appsody stacks available to init",
	Long: `This command lists all the stacks available in your repositories, if you omit the 
	optional [repo] parameter. If you specify the repository name [repo], only the stacks in that
	repositories will be listed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var repos RepositoryFile
		setupErr := setupConfig()
		if setupErr != nil {
			return setupErr
		}		
		//var index RepoIndex
		if len(args) < 1 {
			projects, err := repos.listProjects()
			if err != nil {
				return errors.Errorf("%v", err)
			}
			Info.log("\n", projects)
			return nil
		} else {
			repoName := args[0]
			repos.getRepos()
			if repoProjects, err := repos.listRepoProjects(repoName); err != nil {
				return err
			} else {
				Info.log("\n", repoProjects)
			}


		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

}
