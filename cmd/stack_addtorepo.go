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
	"bytes"
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

func newStackAddToRepoCmd(rootConfig *RootCommandConfig) *cobra.Command {
	var releaseURL string
	var useLocalCache bool
	var repoName string

	log := rootConfig.LoggingConfig

	var stackAddToRepoCmd = &cobra.Command{
		Use:   "add-to-repo <repo-name>",
		Short: "Add stack information into a production Appsody repository",
		Long: `Adds stack information into an Appsody repository. 
		
Adds stack information to a new or existing Appsody repository, specified by the <repo-name> argument. This enables you to share your stack with others.

The updated repository index file is created in  ~/.appsody/stacks/dev.local directory.`,

		Example: `  appsody stack add-to-repo incubator
  Creates a new repository index file for the incubator repository, setting the template URLs to begin with a default URL of https://github.com/appsody/stacks/releases/latest/download/

  appsody stack add-to-repo myrepository --release-url https://github.com/mygitorg/myrepository/releases/latest/download/
  Create a new index file for the myrepository repository, setting the template URLs to begin with https://github.com/mygitorg/myrepository/releases/latest/download/

  appsody stack add-to-repo myrepository --release-url https://github.com/appsody/stacks/releases/latest/download/ --use-local-cache
  Use an existing index for the myrepository repository or create it if it doesnt exist, setting the template URLs to begin with https://github.com/mygitorg/myrepository/releases/latest/download/`,
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Info.Log("******************************************")
			log.Info.Log("Running appsody stack add-to-repo")
			log.Info.Log("******************************************")

			if len(args) < 1 {
				return errors.New("Required parameter missing. You must specify a repository name")
			}

			repoName = args[0]

			log.Debug.Log("repoName is: ", repoName)
			log.Debug.Log("releaseURL is: ", releaseURL)
			log.Debug.Log("useLocalCache is: ", useLocalCache)

			var repoFile RepositoryFile

			stackPath := rootConfig.ProjectDir
			log.Debug.Log("stackPath is: ", stackPath)

			// check for templates dir, error out if its not there
			check, err := Exists("templates")
			if err != nil {
				return errors.New("Error checking stack root directory: " + err.Error())
			}
			if !check {
				// if we can't find the templates directory then we are not starting from a valid root of the stack directory
				return errors.New("Unable to reach templates directory. Current directory must be the root of the stack")
			}

			appsodyHome := getHome(rootConfig.CliConfig)
			log.Debug.Log("appsodyHome is:", appsodyHome)

			devLocal := filepath.Join(appsodyHome, "stacks", "dev.local")
			log.Debug.Log("devLocal is: ", devLocal)

			// create the devLocal directory in appsody home
			err = os.MkdirAll(devLocal, os.FileMode(0755))
			if err != nil {
				return errors.Errorf("Error creating directory: %v", err)
			}

			// get the stack name from the stack path
			stackID := filepath.Base(stackPath)
			log.Debug.Log("stackName is: ", stackID)

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
					log.Debug.log("Remote url: ", url)
					// Check to see whether the local index file exists
					exists, err := Exists(localIndexFile)
					if err != nil {
						return errors.Errorf("Error checking status of %s", localIndexFile)
					}
					if exists && useLocalCache {
						log.Debug.Log(localIndexFile, " exists in the appsody directory and use-local-cache is true")
						source, err := ioutil.ReadFile(localIndexFile)
						if err != nil {
							return errors.Errorf("Error trying to read: %v", err)
						}

						err = yaml.Unmarshal(source, &indexYaml)
						if err != nil {
							return errors.Errorf("Error trying to unmarshall: %v", err)
						}
					} else {
						// local file doesnt exist or use-local-cache is not set to true
						// so download the index from the repote location and use it
						log.Debug.Log(localIndexFile, " doesnt exist in the appsody directory or use-local-cache is false")
						log.Debug.Log("Downloading the remote index file")
						index, err := downloadFileIndex(log, url)
						if err != nil {
							return err
						}
						indexYaml = *index
					}
				} else {
					log.Debug.log("Local url: ", url)
					log.Debug.Log("Modify the local file in the local directory")
					index, err := downloadFileIndex(log, url)
					if err != nil {
						return err
					}
					indexYaml = *index
					localIndexFile = url
				}
			} else {
				log.Debug.Log(repoName, " does not exist within the repository list")
				exists, err := Exists(localIndexFile)
				if err != nil {
					return errors.Errorf("Error checking status of %s", localIndexFile)
				}
				if exists && useLocalCache {
					log.Debug.Log(localIndexFile, " exists in the appsody directory and use-local-cache is true")
					source, err := ioutil.ReadFile(localIndexFile)
					if err != nil {
						return errors.Errorf("Error trying to read: %v", err)
					}

					err = yaml.Unmarshal(source, &indexYaml)
					if err != nil {
						return errors.Errorf("Error trying to unmarshall: %v", err)
					}
				} else {
					// local file doesnt exist or use-local-cache is not set to true
					// so download the index from the repote location and use it
					log.Debug.Log(localIndexFile, " doesnt exist in the appsody directory or use-local-cache is false")
					log.Debug.Log("Creating a new local file in the local directory")

					// create the beginning of the index yaml
					indexYaml = IndexYaml{}
					indexYaml.APIVersion = "v2"
					indexYaml.Stacks = make([]IndexYamlStack, 0, 1)
				}
			}

			// At this point we should have the indexFile loaded that want to use for updating / adding stack info
			// find the index of the stack
			indexYaml = findStackAndRemove(log, stackID, indexYaml)

			// get the necessary data from the current stack.yaml
			stackYaml, err := getStackData(stackPath)
			if err != nil {
				return err
			}

			// build up stack struct for the new stack
			newStackStruct := initialiseStackData(stackID, stackYaml)

			// find and open the template path so we can loop through the templates
			templatePath := filepath.Join(stackPath, "templates")

			t, err := os.Open(templatePath)
			if err != nil {
				return errors.Errorf("Error opening directory: %v", err)
			}

			templates, err := t.Readdirnames(0)
			if err != nil {
				return errors.Errorf("Error reading directories: %v", err)
			}

			// loop through the template directories and create the id and url
			for i := range templates {
				log.Debug.Log("template is: ", templates[i])
				if strings.Contains(templates[i], ".DS_Store") {
					log.Debug.Log("Ignoring .DS_Store")
					continue
				}

				versionArchiveTar := stackID + ".v" + stackYaml.Version + ".templates." + templates[i] + ".tar.gz"
				log.Debug.Log("versionedArdhiveTar is: ", versionArchiveTar)

				templateURL := releaseURL + versionArchiveTar
				log.Debug.Log("full release URL is: ", templateURL)

				// add the template data to the struct
				newTemplateStruct := IndexYamlStackTemplate{}
				newTemplateStruct.ID = templates[i]
				newTemplateStruct.URL = templateURL

				newStackStruct.Templates = append(newStackStruct.Templates, newTemplateStruct)
			}

			t.Close()

			// add the new stack struct to the existing struct
			indexYaml.Stacks = append(indexYaml.Stacks, newStackStruct)

			// Last thing to do is write the data to the file
			data, err := yaml.Marshal(indexYaml)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(localIndexFile, data, 0666)
			if err != nil {
				return errors.Errorf("Error writing localIndexFile: %v", err)
			}

			return nil
		},
	}

	stackAddToRepoCmd.PersistentFlags().StringVar(&releaseURL, "release-url", "https://github.com/appsody/stacks/releases/download/", "URL to use within the repository to access the stack assets")
	stackAddToRepoCmd.PersistentFlags().BoolVar(&useLocalCache, "use-local-cache", false, "Whether to use a local file if exists or create a new file")

	return stackAddToRepoCmd
}

func downloadFileIndex(log *LoggingConfig, url string) (*IndexYaml, error) {
	log.Debug.log("Downloading appsody repository index from ", url)
	indexBuffer := bytes.NewBuffer(nil)
	err := downloadFile(log, url, indexBuffer)
	if err != nil {
		return nil, err
	}

	yamlFile, err := ioutil.ReadAll(indexBuffer)
	if err != nil {
		return nil, fmt.Errorf("Could not read buffer into byte array")
	}
	var index IndexYaml
	err = yaml.Unmarshal(yamlFile, &index)
	if err != nil {
		log.Debug.logf("Contents of downloaded index from %s\n%s", url, yamlFile)
		return nil, fmt.Errorf("Repository index formatting error: %s", err)
	}
	return &index, nil
}
