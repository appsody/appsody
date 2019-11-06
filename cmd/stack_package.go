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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// structs for parsing the yaml files
type StackYaml struct {
	Name            string `yaml:"name"`
	Version         string `yaml:"version"`
	Description     string `yaml:"description"`
	License         string `yaml:"license"`
	Language        string `yaml:"language"`
	Maintainers     []Maintainer
	DefaultTemplate string `yaml:"default-template"`
}
type Maintainer struct {
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	GithubID string `yaml:"github-id" mapstructure:"github-id"`
}

type IndexYaml struct {
	APIVersion string `yaml:"apiVersion"`
	Stacks     []IndexYamlStack
}
type IndexYamlStack struct {
	ID              string `yaml:"id"`
	Name            string `yaml:"name"`
	Version         string `yaml:"version"`
	Description     string `yaml:"description"`
	License         string `yaml:"license"`
	Language        string `yaml:"language"`
	Maintainers     []Maintainer
	DefaultTemplate string `yaml:"default-template"`
	Templates       []IndexYamlStackTemplate
}
type IndexYamlStackTemplate struct {
	ID  string `yaml:"id"`
	URL string `yaml:"url"`
}

func newStackPackageCmd(rootConfig *RootCommandConfig) *cobra.Command {

	// stack package is a tool for local stack developers to package their stack
	// the stack package command does the following...
	// 1. create/update a local index yaml
	// 2. create a tar for each stack template
	// 3. build a docker image
	// 4. create/update an appsody repo for the stack

	var imageNamespace string

	var stackPackageCmd = &cobra.Command{
		Use:   "package",
		Short: "Package a stack in the local Appsody environment",
		Long: `This command is a tool for stack developers to package a stack from their local Appsody development environment. Once the stack is packaged it can then be tested via Appsody commands. The package command performs the following:
- Creates/updates an index file named "dev.local-index.yaml" and stores it in .appsody/stacks/dev.local
- Creates a tar.gz for each stack template and stores it in .appsody/stacks/dev.local
- Builds a Docker image named "dev.local/[stack name]:SNAPSHOT"
- Creates an Appsody repository named "dev.local"
- Adds/updates the "dev.local" repository of your Appsody configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {

			Info.Log("******************************************")
			Info.Log("Running appsody stack package")
			Info.Log("******************************************")

			stackPath := rootConfig.ProjectDir
			Debug.Log("stackPath is: ", stackPath)

			// check for templates dir, error out if its not there
			check, err := Exists("templates")
			if err != nil {
				return errors.New("Error checking stack root directory: " + err.Error())
			}
			if !check {
				// if we can't find the templates directory then we are not starting from a valid root of the stack directory
				return errors.New("Unable to reach templates directory. Current directory must be the root of the stack")
			}

			appsodyHome := getHome(rootConfig)
			Debug.Log("appsodyHome is:", appsodyHome)

			devLocal := filepath.Join(appsodyHome, "stacks", "dev.local")
			Debug.Log("devLocal is: ", devLocal)

			// create the devLocal directory in appsody home
			err = os.MkdirAll(devLocal, os.FileMode(0755))
			if err != nil {
				return errors.Errorf("Error creating directory: %v", err)
			}

			// get the stack name from the stack path
			stackID := filepath.Base(stackPath)
			Debug.Log("stackName is: ", stackID)

			indexFileLocal := filepath.Join(devLocal, "dev.local-index.yaml")
			Debug.Log("indexFileLocal is: ", indexFileLocal)

			// create IndexYaml struct and populate the APIVersion and Stacks header
			var indexYaml IndexYaml

			// check for existing index yaml file
			check, err = Exists(indexFileLocal)
			if err != nil {
				return errors.New("Error checking index file: " + err.Error())
			}
			if check {
				// index file exists already so see if it contains the stack data and remove it if found
				Debug.Log("Index file exists already")

				source, err := ioutil.ReadFile(indexFileLocal)
				if err != nil {
					return errors.Errorf("Error trying to read: %v", err)
				}

				err = yaml.Unmarshal(source, &indexYaml)
				if err != nil {
					return errors.Errorf("Error trying to unmarshall: %v", err)
				}

				// find the index of the stack
				foundStack := -1
				for i, stack := range indexYaml.Stacks {
					if stack.ID == stackID {
						Debug.Log("Existing stack: " + stackID + "found")
						foundStack = i
						break
					}
				}

				// delete index foundStack from indexYaml.Stacks as we will append the new stack later
				if foundStack != -1 {
					indexYaml.Stacks = indexYaml.Stacks[:foundStack+copy(indexYaml.Stacks[foundStack:], indexYaml.Stacks[foundStack+1:])]
				}
			} else {
				// create the beginning of the index yaml
				indexYaml = IndexYaml{}
				indexYaml.APIVersion = "v2"
				indexYaml.Stacks = make([]IndexYamlStack, 0, 1)
			}

			// get the necessary data from the current stack.yaml
			var stackYaml StackYaml

			source, err := ioutil.ReadFile(filepath.Join(stackPath, "stack.yaml"))
			if err != nil {
				return errors.Errorf("Error trying to read: %v", err)
			}

			err = yaml.Unmarshal(source, &stackYaml)
			if err != nil {
				return errors.Errorf("Error trying to unmarshall: %v", err)
			}

			// docker build
			// create the image name to be used for the docker image
			buildImage := imageNamespace + "/" + stackID + ":SNAPSHOT"

			imageDir := filepath.Join(stackPath, "image")
			Debug.Log("imageDir is: ", imageDir)

			dockerFile := filepath.Join(imageDir, "Dockerfile-stack")
			Debug.Log("dockerFile is: ", dockerFile)

			cmdArgs := []string{"-t", buildImage}

			labels, err := getLabelsForStackImage(stackID, buildImage, stackYaml, rootConfig)
			if err != nil {
				return err
			}
			// It would be nicer to only call the --label flag once. Could also use the --label-file flag.
			for _, label := range labels {
				cmdArgs = append(cmdArgs, "--label", label)
			}

			cmdArgs = append(cmdArgs, "-f", dockerFile, imageDir)
			Debug.Log("cmdArgs is: ", cmdArgs)

			Info.Log("Running docker build")

			err = DockerBuild(cmdArgs, DockerLog, rootConfig.Verbose, rootConfig.Dryrun)
			if err != nil {
				return errors.Errorf("Error during docker build: %v", err)
			}

			// build up stack struct for the new stack
			newStackStruct := IndexYamlStack{}
			// set the data in the new stack struct
			newStackStruct.ID = stackID
			newStackStruct.Name = stackYaml.Name
			newStackStruct.Version = stackYaml.Version
			newStackStruct.Description = stackYaml.Description
			newStackStruct.License = stackYaml.License
			newStackStruct.Language = stackYaml.License
			newStackStruct.Maintainers = append(newStackStruct.Maintainers, stackYaml.Maintainers...)
			newStackStruct.DefaultTemplate = stackYaml.DefaultTemplate

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
				Debug.Log("template is: ", templates[i])
				if strings.Contains(templates[i], ".DS_Store") {
					Debug.Log("Ignoring .DS_Store")
					continue
				}

				sourceDir := filepath.Join(stackPath, "templates", templates[i])
				Debug.Log("sourceDir is: ", sourceDir)

				// create name for the tar files
				versionedArchive := filepath.Join(devLocal, stackID+".v"+stackYaml.Version+".templates.")
				Debug.Log("versionedArchive is: ", versionedArchive)

				versionArchiveTar := versionedArchive + templates[i] + ".tar.gz"
				Debug.Log("versionedArdhiveTar is: ", versionArchiveTar)

				if runtime.GOOS == "windows" {
					// for windows, add a leading slash and convert to unix style slashes
					versionArchiveTar = "/" + filepath.ToSlash(versionArchiveTar)
				}
				versionArchiveTar = "file://" + versionArchiveTar

				// add the template data to the struct
				newTemplateStruct := IndexYamlStackTemplate{}
				newTemplateStruct.ID = templates[i]
				newTemplateStruct.URL = versionArchiveTar

				newStackStruct.Templates = append(newStackStruct.Templates, newTemplateStruct)

				// create a config yaml file for the tarball
				configYaml := filepath.Join(templatePath, templates[i], ".appsody-config.yaml")
				Debug.Log("configYaml is: ", configYaml)

				g, err := os.Create(configYaml)
				if err != nil {
					return errors.Errorf("Error trying to create file: %v", err)
				}

				_, err = g.WriteString("stack: " + buildImage)
				if err != nil {
					return errors.Errorf("Error trying to write: %v", err)
				}

				g.Close()

				// tar the files
				Info.Log("Creating tar for: " + templates[i])
				err = Targz(sourceDir, versionedArchive)
				if err != nil {
					return errors.Errorf("Error trying to tar: %v", err)
				}

				// remove the config yaml file
				err = os.Remove(configYaml)
				if err != nil {
					return errors.Errorf("Error trying to remove file: %v", err)
				}
			}

			t.Close()

			// add the new stack struct to the existing struct
			indexYaml.Stacks = append(indexYaml.Stacks, newStackStruct)

			// write yaml data to the index yaml
			source, err = yaml.Marshal(&indexYaml)
			if err != nil {
				return errors.Errorf("Error trying to marshall: %v", err)
			}

			Info.Log("Writing: " + indexFileLocal)
			err = ioutil.WriteFile(indexFileLocal, source, 0644)
			if err != nil {
				return errors.Errorf("Error trying to read: %v", err)
			}

			// list repos
			var repoFile RepositoryFile
			repos, repoErr := repoFile.getRepos(rootConfig)
			if repoErr != nil {
				return repoErr
			}
			// See if a configured repo already points to dev.local, if so remove it
			repoName := "dev.local"

			repo := repos.GetRepo(repoName)
			if repo == nil || !strings.Contains(repo.URL, indexFileLocal) {
				// the repo is setup wrong, delete and recreate it
				if repo != nil {
					Info.logf("Appsody repo %s is configured with the wrong URL. Deleting and recreating it.", repoName)
					repos.Remove(repoName)
				}
				// check for a different repo with the same file url
				var repoNameToDelete string
				for _, repo := range repos.Repositories {
					if strings.Contains(repo.URL, indexFileLocal) {
						repoNameToDelete = repo.Name
						break
					}
				}
				if repoNameToDelete != "" {
					Info.logf("Appsody repo %s is configured with %s's URL. Deleting it to setup %s.", repoNameToDelete, repoName, repoName)
					repos.Remove(repoNameToDelete)
				}
				err = repos.WriteFile(getRepoFileLocation(rootConfig))
				if err != nil {
					return errors.Errorf("Error writing to repo file %s. %v", getRepoFileLocation(rootConfig), err)
				}
				Info.Logf("Creating %s repository", repoName)
				_, err = AddLocalFileRepo(repoName, indexFileLocal)
				if err != nil {
					return errors.Errorf("Error adding local repository. Your stack may not be available to appsody commands. %v", err)
				}
			}

			Info.log("Your local stack is available as part of repo ", repoName)

			return nil
		},
	}

	stackPackageCmd.PersistentFlags().StringVar(&imageNamespace, "image-namespace", "dev.local", "Namespace that the images will be created using (default is dev.local)")

	return stackPackageCmd
}

func getLabelsForStackImage(stackID string, buildImage string, stackYaml StackYaml, config *RootCommandConfig) ([]string, error) {
	var labels []string

	gitLabels, err := getGitLabels(config)
	if err != nil {
		Info.log(err)
	} else {
		if branchURL, ok := gitLabels[ociKeyPrefix+"source"]; ok {
			if contextDir, ok := gitLabels[appsodyImageCommitKeyPrefix+"contextDir"]; ok {
				branchURL += contextDir
				gitLabels[ociKeyPrefix+"url"] = branchURL
			}
			// These are enforced by the stack lint so they should exist
			gitLabels[ociKeyPrefix+"documentation"] = branchURL + "/README.md"
			gitLabels[ociKeyPrefix+"source"] = branchURL + "/image"
		}

		for key, value := range gitLabels {
			labelString := fmt.Sprintf("%s=%s", key, value)
			labels = append(labels, labelString)
		}
	}

	// build a ProjectConfig struct from the stackyaml so we can reuse getConfigLabels() func
	projectConfig := ProjectConfig{
		ProjectName: stackYaml.Name,
		Version:     stackYaml.Version,
		Description: stackYaml.Description,
		License:     stackYaml.License,
		Maintainers: stackYaml.Maintainers,
	}
	configLabels, err := getConfigLabels(projectConfig)
	if err != nil {
		return labels, err
	}
	configLabels[appsodyStackKeyPrefix+"id"] = stackID
	configLabels[appsodyStackKeyPrefix+"tag"] = buildImage

	for key, value := range configLabels {
		labelString := fmt.Sprintf("%s=%s", key, value)
		labels = append(labels, labelString)
	}

	return labels, nil
}
