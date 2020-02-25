// Copyright © 2020 IBM Corporation and others.
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
	"os"
	"strings"

	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

func newStackRemoveFromRepoCmd(rootConfig *RootCommandConfig) *cobra.Command {
	var stackName string
	var repoName string
	var useLocalCache bool

	log := rootConfig.LoggingConfig

	var stackRemoveFromRepoCmd = &cobra.Command{
		Use:   "remove-from-repo <repository> <stack>",
		Short: "Remove stack information from an Appsody repository",
		Long: `Removes stack information from an Appsody repository. 
		
Removes stack information, specified by <stack> from an Appsody repository, specified by the <repository> argument.

The updated repository index file is created in  ~/.appsody/stacks/dev.local directory.`,

		Example: `  appsody stack remove-from-repo incubator nodejs
  Updates the repository index file for the incubator repository, removing the definition of the nodejs stack`,
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Info.Log("******************************************")
			log.Info.Log("Running appsody stack remove-from-repo")
			log.Info.Log("******************************************")

			if len(args) < 2 {
				return errors.New("Required parameter missing. You must specify a repository name and a stack name")
			}
			if len(args) > 2 {
				return errors.Errorf("Two arguments expected. Use 'appsody [command] --help' for more information about a command")
			}

			repoName = args[0]
			stackName = args[1]

			log.Debug.Log("repoName is: ", repoName)
			log.Debug.Log("stackName is: ", stackName)

			var repoFile RepositoryFile

			appsodyHome := getHome(rootConfig)
			log.Debug.Log("appsodyHome is:", appsodyHome)

			devLocal := filepath.Join(appsodyHome, "stacks", "dev.local")
			log.Debug.Log("devLocal is: ", devLocal)

			// create the devLocal directory in appsody home
			err := os.MkdirAll(devLocal, os.FileMode(0755))
			if err != nil {
				return errors.Errorf("Error creating directory: %v", err)
			}

			localIndexFile := filepath.Join(devLocal, repoName+"-index.yaml")
			log.Debug.Log("localIndexFile is: ", localIndexFile)

			var indexYaml IndexYaml

			_, repoErr := repoFile.getRepos(rootConfig)
			if repoErr != nil {
				return repoErr
			}

			if repoFile.Has(repoName) {
				// The repoName exists within the repository list
				log.Debug.Log(repoName, " exists within the repository list")
				repo := repoFile.GetRepo(repoName)
				url := repo.URL
				log.Debug.Log(repoName, " URL is: ", url)

				// Check if the repo URL points to a remote repository
				if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
					log.Debug.log("Remote repository index url: ", url)
					// Check to see whether the local index file exists
					exists, err := Exists(localIndexFile)
					if err != nil {
						return errors.Errorf("Error checking status of %s", localIndexFile)
					}
					if exists && useLocalCache {
						log.Debug.Log(localIndexFile, " exists in the appsody directory and use-local-cache is true")
					} else {
						// local file doesnt exist, download the index from the repote location and use it
						log.Debug.Log(localIndexFile, " doesnt exist in the appsody directory")
						log.Info.Log("Downloading the remote index file from: ", url)
						log.Info.log("Creating repository index file: ", localIndexFile)
						err := downloadFileToDisk(log, url, localIndexFile, false)
						if err != nil {
							return err
						}
					}
				} else {
					log.Debug.log("Local repository index url: ", url)
					localIndexFile = strings.TrimPrefix(url, "file://")
					exists, err := Exists(localIndexFile)
					if err != nil {
						return errors.Errorf("Error checking status of %s", localIndexFile)
					}
					if exists {
						log.Debug.Log(localIndexFile, " exists in the appsody directory")
					} else {
						log.Info.Log("Repository index file not found - unable to remove stack")
						return nil
					}
				}
			} else {
				return errors.Errorf("%v does not exist within the repository list", repoName)
			}

			log.Info.log("Updating repository index file: ", localIndexFile)
			source, err := ioutil.ReadFile(localIndexFile)
			if err != nil {
				return errors.Errorf("Error trying to read: %v", err)
			}
			err = yaml.Unmarshal(source, &indexYaml)
			if err != nil {
				return errors.Errorf("Error trying to unmarshall: %v", err)
			}

			// At this point we should have the indexFile loaded that want to use for updating / adding stack info
			// find the index of the stack
			indexYaml, stackExists := findStackAndRemove(log, stackName, indexYaml)
			if !stackExists {
				log.Info.Logf("Stack: %v does not exist in repository index file", stackName)
				return nil
			}

			// Last thing to do is write the data to the file
			data, err := yaml.Marshal(indexYaml)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(localIndexFile, data, 0666)
			if err != nil {
				return errors.Errorf("Error writing localIndexFile: %v", err)
			}
			err = generateCodewindJSON(log, indexYaml, localIndexFile, repoName)
			if err != nil {
				return errors.Errorf("Could not generate json file from yaml index: %v", err)
			}

			log.Info.Log("Repository index file updated successfully")

			return nil
		},
	}

	stackRemoveFromRepoCmd.PersistentFlags().BoolVar(&useLocalCache, "use-local-cache", false, "Whether to use a local file if exists or create a new file")

	return stackRemoveFromRepoCmd
}
