// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// initCmd represents the init command
var addCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add an Appsody repository",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			Error.log("Error, you must specify repository name and URL")
			os.Exit(1)
		}
		var repoName = args[0]
		var repoURL = args[1]
		if dryrun {
			Info.logf("Dry Run - Skipping appsody repo add repository Name: %s, URL: %s", repoName, repoURL)
		} else {
			var indexBuffer, err = downloadIndex(repoURL)
			if err != nil {
				log.Fatalf("Failed to verify repository location err   #%v ", err)
			}
			yamlFile, err := ioutil.ReadAll(indexBuffer)
			if err != nil {
				log.Fatalf("Failed to read from repository location err   #%v ", err)
			}

			var index RepoIndex
			err = yaml.Unmarshal(yamlFile, &index)
			if err != nil {
				log.Fatalf("Failed to format index from repository location: %v", err)
			}

			var newEntry = RepositoryEntry{
				Name: repoName,
				URL:  repoURL,
			}

			// Need to check to see if it already exists under a different name?
			var repoFile RepositoryFile
			repoFile.getRepos()
			repoFile.Add(&newEntry)
			err = repoFile.WriteFile(getRepoFileLocation())
			if err != nil {
				log.Fatalf("Failed to write file to repository location: %v", err)
			}
		}
	},
}

func init() {
	repoCmd.AddCommand(addCmd)

}
