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

	var stackPackageCmd = &cobra.Command{
		Use:   "package",
		Short: "Package a stack in the local Appsody environment",
		Long: `This command is a tool for stack developers to package a stack from their local Appsody development environment. Once the stack is packaged it can then be tested via Appsody commands. The package command performs the following:
- Creates/updates an index file named "index-dev-local.yaml" and stores it in .appsody/stacks/dev.local
- Creates a tar.gz for each stack template and stores it in .appsody/stacks/dev.local
- Builds a Docker image named "dev.local/[stack name]:SNAPSHOT"
- Creates an Appsody repository named "dev-local"
- Adds/updates the "dev-local" repository of your Appsody configuration`,
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
			stackName := filepath.Base(stackPath)
			Debug.Log("stackName is: ", stackName)

			indexFileLocal := filepath.Join(devLocal, "index-dev-local.yaml")
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
					if stack.ID == stackName {
						Debug.Log("Existing stack: " + stackName + "found")
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

			// build up stack struct for the new stack
			newStackStruct := IndexYamlStack{}

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

			// set the data in the new stack struct
			newStackStruct.ID = stackName
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
				versionedArchive := filepath.Join(devLocal, stackName+".v"+stackYaml.Version+".templates.")
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

				// docker build

				// create the image name to be used for the docker image
				buildImage := "dev.local/" + stackName + ":SNAPSHOT"

				imageDir := filepath.Join(stackPath, "image")
				Debug.Log("imageDir is: ", imageDir)

				dockerFile := filepath.Join(imageDir, "Dockerfile-stack")
				Debug.Log("dockerFile is: ", dockerFile)

				cmdArgs := []string{"-t", buildImage}
				cmdArgs = append(cmdArgs, "-f", dockerFile, imageDir)
				Debug.Log("cmdArgs is: ", cmdArgs)

				Info.Log("Running docker build")

				err = DockerBuild(cmdArgs, DockerLog, rootConfig.Verbose, rootConfig.Dryrun)
				if err != nil {
					return errors.Errorf("Error during docker build: %v", err)
				}

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
			repos, err := RunAppsodyCmdExec([]string{"repo", "list", "-o", "yaml"}, ".")
			if err != nil {
				return err
			}

			// if dev.local exists then remove it
			if strings.Contains(repos, indexFileLocal) {
				Info.Log("Existing repo found for local index ", indexFileLocal)
			} else {
				// create an appsody repo for the stack
				Info.Log("Creating dev.local repository")
				_, err = AddLocalFileRepo("dev.local", indexFileLocal)
				if err != nil {
					return errors.Errorf("Error running appsody command: %v", err)
				}
			}

			return nil
		},
	}
	return stackPackageCmd
}
