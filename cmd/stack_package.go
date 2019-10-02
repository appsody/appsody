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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// structs for parsing the stack.yaml
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
	GithubID string `yaml:"github-id"`
}

func newStackPackageCmd(rootConfig *RootCommandConfig) *cobra.Command {

	// stack package is a tool for local stack developers to package their stack
	// the stack package command does the following...
	// 1. create a local index yaml
	// 2. create a tar for each stack template
	// 3. build a docker image
	// 4. create an appsody repo for the stack

	var stackPackageCmd = &cobra.Command{
		Use:   "package",
		Short: "Package a stack in the local Appsody environment",
		Long:  `This builds a stack and creates an index and adds it to the repository`,
		RunE: func(cmd *cobra.Command, args []string) error {

			Info.Log("******************************************")
			Info.Log("Running appsody stack package")
			Info.Log("******************************************")

			stackPath := rootConfig.ProjectDir
			Info.Log("stackPath is: ", stackPath)

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
			Info.Log("appsodyHome is:", appsodyHome)

			devLocal := filepath.Join(appsodyHome, "stacks", "dev.local")
			Info.Log("devLocal is: ", devLocal)

			// create the devLocal directory in appsody home
			err = os.MkdirAll(devLocal, os.FileMode(0755))
			if err != nil {
				return errors.Errorf("Error creating directory: %v", err)
			}

			// get the stack name from the stack path
			stackName := filepath.Base(stackPath)
			Info.Log("stackName is: ", stackName)

			indexFileLocal := filepath.Join(devLocal, "index-dev-local.yaml")
			Info.Log("indexFileLocal is: ", indexFileLocal)

			// create and write the index yaml

			indexFileStr := "apiVersion: v2\n"
			indexFileStr += "stacks:\n"
			indexFileStr += "  - id: " + stackName + "\n"

			f, err := os.Create(indexFileLocal)
			if err != nil {
				return errors.Errorf("Error creating file: %v", err)
			}
			defer f.Close()

			_, err = f.WriteString(indexFileStr)
			if err != nil {
				return errors.Errorf("Error trying to write: %v", err)
			}

			// get the necessary data from the current stack yaml
			var stackYaml StackYaml

			source, err := ioutil.ReadFile(filepath.Join(stackPath, "stack.yaml"))
			if err != nil {
				return errors.Errorf("Error trying to read: %v", err)
			}

			err = yaml.Unmarshal(source, &stackYaml)
			if err != nil {
				return errors.Errorf("Error trying to unmarshall: %v", err)
			}

			Info.Logf("StackYaml Name: %#v", stackYaml.Name)
			Info.Logf("StackYaml Version: %#v", stackYaml.Version)
			Info.Logf("StackYaml Description: %#v", stackYaml.Description)
			Info.Logf("StackYaml License: %#v", stackYaml.License)
			Info.Logf("StackYaml Language: %#v", stackYaml.Language)
			Info.Logf("StackYaml DefaultTemplate: %#v", stackYaml.DefaultTemplate)

			for i := range stackYaml.Maintainers {
				Info.Logf("Maintainers Name: %#v", stackYaml.Maintainers[i].Name)
				Info.Logf("Maintainers Email: %#v", stackYaml.Maintainers[i].Email)
				Info.Logf("Maintainers GithubID: %#v", stackYaml.Maintainers[i].GithubID)
			}

			// create the stack yaml string to write
			stackYamlStr := "    name: " + stackYaml.Name + "\n"
			stackYamlStr += "    version: " + stackYaml.Version + "\n"
			stackYamlStr += "    description: " + stackYaml.Description + "\n"
			stackYamlStr += "    license: " + stackYaml.License + "\n"
			stackYamlStr += "    language: " + stackYaml.Language + "\n"
			stackYamlStr += "    maintainers:\n"

			// write the stack yaml data we have so far
			_, err = f.WriteString(stackYamlStr)
			if err != nil {
				return errors.Errorf("Error trying to write: %v", err)
			}

			// loop through the Maintainers
			for i := range stackYaml.Maintainers {
				// create maintainer data string

				stackYamlMaintainersStr := "     - name: " + stackYaml.Maintainers[i].Name + "\n"
				stackYamlMaintainersStr += "       email: " + stackYaml.Maintainers[i].Email + "\n"
				stackYamlMaintainersStr += "       github-id: " + stackYaml.Maintainers[i].GithubID + "\n"

				// write the maintainer data
				_, err = f.WriteString(stackYamlMaintainersStr)
				if err != nil {
					return errors.Errorf("Error trying to write: %v", err)
				}
			}

			// create template data string
			stackYamlTemplateStr := "    default-template: " + stackYaml.DefaultTemplate + "\n"
			stackYamlTemplateStr += "    templates:\n"

			// write the template data
			_, err = f.WriteString(stackYamlTemplateStr)
			if err != nil {
				return errors.Errorf("Error trying to write: %v", err)
			}

			// we still need the url for the index but we will write it while taring the templates

			// docker build

			// create the image name to be used for the docker image
			buildImage := "dev.local/" + stackName + ":SNAPSHOT"

			imageDir := filepath.Join(stackPath, "image")
			Info.Log("imageDir is: ", imageDir)

			dockerFile := filepath.Join(imageDir, "Dockerfile-stack")
			Info.Log("dockerFile is: ", dockerFile)

			cmdArgs := []string{"-t", buildImage}

			cmdArgs = append(cmdArgs, "-f", dockerFile, imageDir)
			Info.Log("cmdArgs is: ", cmdArgs)

			err = DockerBuild(cmdArgs, DockerLog, rootConfig.Verbose, rootConfig.Dryrun)
			if err != nil {
				return errors.Errorf("Error during docker build: %v", err)
			}

			// tar the templates

			templatePath := filepath.Join(stackPath, "templates")

			t, err := os.Open(templatePath)
			if err != nil {
				return errors.Errorf("Error opening directory: %v", err)
			}

			templates, err := t.Readdirnames(0)
			if err != nil {
				return errors.Errorf("Error reading directories: %v", err)
			}

			// loop through the template directories
			// write the template url in the index yaml
			// create a tar.gz for each template
			for i := range templates {
				Info.Log("template is: ", templates[i])

				sourceDir := filepath.Join(stackPath, "templates", templates[i])
				Info.Log("sourceDir is: ", sourceDir)

				// create name for the tar files
				versionedArchive := filepath.Join(devLocal, stackName+".v"+stackYaml.Version+".templates.")
				Info.Log("versionedArchive is: ", versionedArchive)

				versionArchiveTar := versionedArchive + templates[i] + ".tar.gz"
				Info.Log("versionedArdhiveTar is: ", versionArchiveTar)

				// create the template tar data string
				templateTarStr := "      - id: " + templates[i] + "\n"
				templateTarStr += "        url: file://" + versionArchiveTar + "\n"

				// write the template url in the index yaml
				_, err = f.WriteString(templateTarStr)
				if err != nil {
					return errors.Errorf("Error trying to write: %v", err)
				}

				// create a config yaml file for the tarball
				configYaml := filepath.Join(templatePath, templates[i], ".appsody-config.yaml")
				Info.Log("configYaml is: ", configYaml)

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

			// create an appsody repo for the stack
			yamlPath := "file://" + indexFileLocal
			_, err = RunAppsodyCmdExec([]string{"repo", "add", "dev-local", yamlPath}, stackPath)
			if err != nil {
				return errors.Errorf("Error running appsody command: %v", err)
			}

			return nil

		},
	}
	return stackPackageCmd
}
