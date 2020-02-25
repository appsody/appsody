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
		Use:   "add-to-repo <repository>",
		Short: "Add stack information into a production Appsody repository",
		Long: `Adds stack information into an Appsody repository. 
		
Adds stack information to a new or existing Appsody repository, specified by the <repository> argument. This enables you to share your stack with others. This command must be run after appsody stack package on the chosen stack.

The updated repository index file is created in  ~/.appsody/stacks/dev.local directory.

Run this command from the root directory of your Appsody project.`,
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
			if len(args) > 1 {
				return errors.Errorf("One argument expected. Use 'appsody [command] --help' for more information about a command")
			}

			repoName = args[0]

			log.Debug.Log("repoName is: ", repoName)
			log.Debug.Log("releaseURL is: ", releaseURL)
			log.Debug.Log("useLocalCache is: ", useLocalCache)

			var repoFile RepositoryFile
			createNewLocalIndex := false

			stackPath := rootConfig.ProjectDir
			log.Debug.Log("stackPath is: ", stackPath)

			// check for templates dir, error out if its not there
			check, err := Exists(filepath.Join(stackPath, "templates"))
			if err != nil {
				return errors.New("Error checking stack root directory: " + err.Error())
			}
			if !check {
				// if we can't find the templates directory then we are not starting from a valid root of the stack directory
				return errors.New("Unable to reach templates directory. Current directory must be the root of the stack")
			}

			appsodyHome := getHome(rootConfig)
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

			log.Debug.Log("Checking if appsody stack package has been run on this stack")
			checkDevLocalIndex, existsErr := Exists(filepath.Join(devLocal, "dev.local-index.yaml"))
			if existsErr != nil {
				return errors.Errorf("Error checking if dev.local-index.yaml exists in: %s. Error was: %s", devLocal, existsErr)
			}
			if !checkDevLocalIndex {
				return errors.Errorf("Unable to find dev.local-index.yaml in: %s. Run appsody stack package on your stack before running this command.", devLocal)
			}

			devLocalIndexYaml := IndexYaml{}
			devLocalIndex, readErr := ioutil.ReadFile(filepath.Join(devLocal, "dev.local-index.yaml"))
			if readErr != nil {
				return errors.Errorf("Error reading index file: %s", readErr)
			}

			unmarshalErr := yaml.Unmarshal(devLocalIndex, &devLocalIndexYaml)
			if unmarshalErr != nil {
				return errors.Errorf("Unmarshal: Error unmarshalling index.yaml")
			}

			stackFound := false
			var stackToAddImage string
			for i, stack := range devLocalIndexYaml.Stacks {
				if stackID == stack.ID {
					log.Debug.Log("Found stack attempting to add to repo in dev.local-index.yaml")
					stackFound = true
					stackToAddImage = devLocalIndexYaml.Stacks[i].Image
					break
				}
			}
			if !stackFound {
				return errors.Errorf("Couldn't find stack in dev.local-index.yaml. Have you packaged this stack?")
			}

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
						// local file doesnt exist or use-local-cache is not set to true
						// so download the index from the repote location and use it
						log.Debug.Log(localIndexFile, " doesnt exist in the appsody directory or use-local-cache is false")
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
					if exists && useLocalCache {
						log.Debug.Log(localIndexFile, " exists in the appsody directory and use-local-cache is true")
					} else {
						log.Debug.Log(localIndexFile, " doesnt exist in the appsody directory or use-local-cache is false")
						createNewLocalIndex = true
					}
				}
			} else {
				log.Debug.Log(repoName, " does not exist within the repository list")
				exists, err := Exists(localIndexFile)
				if err != nil {
					return errors.Errorf("Error checking status of %s", localIndexFile)
				}
				if exists && useLocalCache {
					log.Debug.Log(localIndexFile, " exists in the appsody directory and use-local-cache is true")
				} else {
					// local file doesnt exist or use-local-cache is not set to true
					// so download the index from the repote location and use it
					log.Debug.Log(localIndexFile, " doesnt exist in the appsody directory or use-local-cache is false")
					createNewLocalIndex = true
				}
			}

			if createNewLocalIndex {
				// create the beginning of the index yaml
				log.Info.log("Creating repository index file: ", localIndexFile)
				indexYaml = IndexYaml{}
				indexYaml.APIVersion = "v2"
				indexYaml.Stacks = make([]IndexYamlStack, 0, 1)
			} else {
				log.Info.log("Updating repository index file: ", localIndexFile)
				source, err := ioutil.ReadFile(localIndexFile)
				if err != nil {
					return errors.Errorf("Error trying to read: %v", err)
				}
				err = yaml.Unmarshal(source, &indexYaml)
				if err != nil {
					return errors.Errorf("Error trying to unmarshall: %v", err)
				}
			}

			// At this point we should have the indexFile loaded that want to use for updating / adding stack info
			// find the index of the stack
			indexYaml, stackExists := findStackAndRemove(log, stackID, indexYaml)

			if stackExists {
				log.Debug.Logf("Stack: %v already exists in repo", stackID)
			}

			// get the necessary data from the current stack.yaml
			stackYaml, err := getStackData(stackPath)
			if err != nil {
				return err
			}

			// build up stack struct for the new stack
			newStackStruct := initialiseStackData(stackID, stackToAddImage, stackYaml)

			versionArchiveTar := stackID + ".v" + stackYaml.Version + ".source.tar.gz"
			log.Debug.Log("versionedArchiveTar is: ", versionArchiveTar)

			sourceURL := releaseURL + versionArchiveTar
			log.Debug.Log("full release URL is: ", sourceURL)

			newStackStruct.SourceURL = sourceURL

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
				log.Debug.Log("versionedArchiveTar is: ", versionArchiveTar)

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

			err = generateCodewindJSON(log, indexYaml, localIndexFile, repoName)
			if err != nil {
				return errors.Errorf("Could not generate json file from yaml index: %v", err)
			}

			log.Info.Log("Repository index file updated successfully")

			return nil
		},
	}

	stackAddToRepoCmd.PersistentFlags().StringVar(&releaseURL, "release-url", "https://github.com/appsody/stacks/releases/download/", "URL to use within the repository to access the stack assets")
	stackAddToRepoCmd.PersistentFlags().BoolVar(&useLocalCache, "use-local-cache", false, "Whether to use a local file if exists or create a new file")

	return stackAddToRepoCmd
}

func getStackData(stackPath string) (StackYaml, error) {
	// get the necessary data from the current stack.yaml
	var stackYaml StackYaml

	source, err := ioutil.ReadFile(filepath.Join(stackPath, "stack.yaml"))
	if err != nil {
		return stackYaml, errors.Errorf("Error trying to read: %v", err)
	}

	err = yaml.Unmarshal(source, &stackYaml)
	if err != nil {
		return stackYaml, errors.Errorf("Error trying to unmarshall: %v", err)
	}

	return stackYaml, nil
}
