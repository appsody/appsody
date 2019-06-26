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

// repo list represent repo list cmd
var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured Appsody repositories",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var repos RepositoryFile
		repos.getRepos()
		var repoList = repos.listRepos()
		Info.log("\n", repoList)
	},
}

func init() {
	repoCmd.AddCommand(repoListCmd)

}
