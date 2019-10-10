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
	GithubID string `yaml:"github-id"`
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
	// 1. create a local index yaml
	// 2. create a tar for each stack template
	// 3. build a docker image
	// 4. create an appsody repo for the stack

	var stackPackageCmd = &cobra.Command{
		Use:   "package",
		Short: "Package a stack in the local Appsody environment",
		Long: `This command is a tool for stack developers to package a stack from their local Appsody development environment. Once the stack is packaged it can then be tested via Appsody commands. The package command performs the following:
		- Creates an index file named "index-dev-local.yaml" and stores it in .appsody/stacks/dev.local
		- Creates a tar.gz for each stack template and stores it in .appsody/stacks/dev.local
		- Builds a Docker image named "dev.local/[stack name]:SNAPSHOT
		- Creates an Appsody repository named "dev-local"`,
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

			// check for existing index file
			// if it exists then check for existing stack id
			//   if the stack id exists then delete it
			// append new stack to index file

			check, err = Exists(indexFileLocal)
			if err != nil {
				return errors.New("Error checking index file: " + err.Error())
			}
			if check {
				// index file exists already so see if it contains the stack data
				fmt.Println("***index file exists already***")

				// find the index of the stack
				foundStack := -1
				for i, stack := range indexYaml.Stacks {
					if stack.ID == stackName {
						foundStack = i
						break
					}
				}

				// delete index foundStack from indexYaml.Stacks
				if foundStack != -1 {
					indexYaml.Stacks[foundStack] = IndexYaml.Stacks[len(index.Yaml.Stacks) -1]
					indexYaml.Stacks[len(index.Yaml.Stacks) -1] = ""
					indexYaml.Stacks = indexYaml.Stacks[:len(index.Yaml.Stacks) -1]
				}
			}

			// build up stack struct for the new stack

			newStackStruct := IndexYamlStack{}

			

				// append stack to indexYaml.Stacks

				// write it to a file using marshall


						// get the necessary data from the current stack yaml
						var stackYaml StackYaml

						source, err = ioutil.ReadFile(filepath.Join(stackPath, "stack.yaml"))
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

						newStackStruct.ID = stackName
						newStackStruct.Name = stackYaml.Name
						newStackStruct.Version = stackYaml.Version
						newStackStruct.Description = stackYaml.Description
						newStackStruct.License = stackYaml.License
						newStackStruct.Language = stackYaml.License

						// loop through the Maintainers
						for _, maintainer := range stackYaml.Maintainers {
							Info.Logf("Maintainers Name: %#v", maintainer.Name)
							Info.Logf("Maintainers Email: %#v", maintainer.Email)
							Info.Logf("Maintainers GithubID: %#v", maintainer.GithubID)

							newStackStruct.Maintainers = append(newStackStruct.Maintainers, maintainer)

						}

						newStackStruct.DefaultTemplate = stackYaml.DefaultTemplate

						// write the template data
						_, err = h.WriteString(stackYamlTemplateStr)
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

							if runtime.GOOS == "windows" {
								// for windows, add a leading slash and convert to unix style slashes
								versionArchiveTar = "/" + filepath.ToSlash(versionArchiveTar)
							}
							versionArchiveTar = "file://" + versionArchiveTar

							// create the template tar data string
							templateTarStr := "      - id: " + templates[i] + "\n"
							templateTarStr += "        url: " + versionArchiveTar + "\n"

							// write the template url in the index yaml
							_, err = h.WriteString(templateTarStr)
							if err != nil {
								return errors.Errorf("Error trying to write: %v", err)
							}

							// create a config yaml file for the tarball
							configYaml := filepath.Join(templatePath, templates[i], ".appsody-config.yaml")
							Info.Log("configYaml is: ", configYaml)

							k, err := os.Create(configYaml)
							if err != nil {
								return errors.Errorf("Error trying to create file: %v", err)
							}

							_, err = k.WriteString("stack: " + buildImage)
							if err != nil {
								return errors.Errorf("Error trying to write: %v", err)
							}

							k.Close()

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

					} else {
						// the stack does not match the one in the index 

						// create the stack yaml string to write

						tempStackYamlStr := "  - id: " + stack.ID + "\n"
						tempStackYamlStr += "    name: " + stack.Name + "\n"
						tempStackYamlStr += "    version: " + stack.Version + "\n"
						tempStackYamlStr += "    description: " + stack.Description + "\n"
						tempStackYamlStr += "    license: " + stack.License + "\n"
						tempStackYamlStr += "    language: " + stack.Language + "\n"
						tempStackYamlStr += "    maintainers:\n"

						// write the stack yaml data we have so far
						_, err = h.WriteString(tempStackYamlStr)
						if err != nil {
							return errors.Errorf("Error trying to write: %v", err)
						}

						// loop through the Maintainers
						for _, maintainer := range stack.Maintainers {

							// create maintainer data string
							tempStackYamlMaintainersStr := "     - name: " + maintainer.Name + "\n"
							tempStackYamlMaintainersStr += "       email: " + maintainer.Email + "\n"
							tempStackYamlMaintainersStr += "       github-id: " + maintainer.GithubID + "\n"

							// write the maintainer data
							_, err = h.WriteString(tempStackYamlMaintainersStr)
							if err != nil {
								return errors.Errorf("Error trying to write: %v", err)
							}
						}

						// create template data string
						tempStackYamlTemplateStr := "    default-template: " + stack.DefaultTemplate + "\n"
						tempStackYamlTemplateStr += "    templates:\n"

						// write the template data
						_, err = h.WriteString(tempStackYamlTemplateStr)
						if err != nil {
							return errors.Errorf("Error trying to write: %v", err)
						}

						// loop through Templates
						for _, template := range stack.Templates {

							// create templdate data string
							tempTemplateStr := "      - id: " + template.ID + "\n"
							tempTemplateStr += "        url: " + template.URL + "\n"

							// write the template data
							_, err = h.WriteString(tempTemplateStr)
							if err != nil {
								return errors.Errorf("Error trying to write: %v", err)
							}
						}
					}

				}
				// 		Info.Logf("IndexYaml Id: %#v", stack.ID)
				// 		Info.Logf("IndexYaml Name: %#v", stack.Name)
				// 		Info.Logf("IndexYaml Version: %#v", stack.Version)
				// 		Info.Logf("IndexYaml Description: %#v", stack.Description)
				// 		Info.Logf("IndexYaml License: %#v", stack.License)
				// 		Info.Logf("IndexYaml Language: %#v", stack.Language)
				// 		Info.Logf("IndexYaml DefaultTemplate: %#v", stack.DefaultTemplate)

				// 		for _, template := range stack.Templates {
				// 			Info.Logf("IndexYamlStackTemplate Id: %#v", template.ID)
				// 			Info.Logf("IndexYamlStackTemplate Url: %#v", template.URL)
				// 		}
				// 		for _, maintainer := range stack.Maintainers {
				// 			Info.Logf("IndexYamlStackMaintainer Name: %#v", maintainer.Name)
				// 			Info.Logf("IndexYamlStackMaintainer Email: %#v", maintainer.Email)
				// 			Info.Logf("IndexYamlStackMaintainer Github-id: %#v", maintainer.GithubID)
				// 		}
				// 	}
				// }

			} else {
				// index file did not exist so create a new one
				f, err := os.Create(indexFileLocal)
				if err != nil {
					return errors.Errorf("Error creating file: %v", err)
				}
				defer f.Close()

				// create and write the index yaml

				indexFileStr := "apiVersion: v2\n"
				indexFileStr += "stacks:\n"
				indexFileStr += "  - id: " + stackName + "\n"

				f, err = os.Create(indexFileLocal)
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

				for _, maintainer := range stackYaml.Maintainers {

					//for i := range stackYaml.Maintainers {
					Info.Logf("Maintainers Name: %#v", maintainer.Name)
					Info.Logf("Maintainers Email: %#v", maintainer.Email)
					Info.Logf("Maintainers GithubID: %#v", maintainer.GithubID)
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
				for _, maintainer := range stackYaml.Maintainers {

					// create maintainer data string
					stackYamlMaintainersStr := "     - name: " + maintainer.Name + "\n"
					stackYamlMaintainersStr += "       email: " + maintainer.Email + "\n"
					stackYamlMaintainersStr += "       github-id: " + maintainer.GithubID + "\n"

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

					if runtime.GOOS == "windows" {
						// for windows, add a leading slash and convert to unix style slashes
						versionArchiveTar = "/" + filepath.ToSlash(versionArchiveTar)
					}
					versionArchiveTar = "file://" + versionArchiveTar

					// create the template tar data string
					templateTarStr := "      - id: " + templates[i] + "\n"
					templateTarStr += "        url: " + versionArchiveTar + "\n"

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
			}

			// run appsody repo list -o yaml
			// parse output for dev-local
			// if dev-local exists then...
			//   run apposody repo remove dev-local
			//   parse .appsody/stacks/dev.local/index-dev-local.yaml for the stack name id
			//   if the stack name id exist then overwrite it with the current stack
			//   else append current stack info
			// run appsody repo add dev-local

			// create an appsody repo for the stack
			_, err = AddLocalFileRepo("dev-local", indexFileLocal)
			if err != nil {
				fmt.Println("add local file repo err is: ", err)
				return errors.Errorf("Error running appsody command: %v", err)
			}

			return nil

		},
	}
	return stackPackageCmd
}

// **********************************************************************
// check ability to read an existing index...
// var indexYaml IndexYaml

// source, err := ioutil.ReadFile(indexFileLocal)

// if err != nil {
// 	return errors.Errorf("Error trying to read: %v", err)
// }

// err = yaml.Unmarshal(source, &indexYaml)
// if err != nil {
// 	return errors.Errorf("Error trying to unmarshall: %v", err)
// }

// Info.Logf("IndexYaml ApiVersion: %#v", indexYaml.APIVersion)

// for _, stack := range indexYaml.Stacks {
// 	Info.Logf("IndexYaml Id: %#v", stack.ID)
// 	Info.Logf("IndexYaml Name: %#v", stack.Name)
// 	Info.Logf("IndexYaml Version: %#v", stack.Version)
// 	Info.Logf("IndexYaml Description: %#v", stack.Description)
// 	Info.Logf("IndexYaml License: %#v", stack.License)
// 	Info.Logf("IndexYaml Language: %#v", stack.Language)
// 	Info.Logf("IndexYaml DefaultTemplate: %#v", stack.DefaultTemplate)

// 	for _, template := range stack.Templates {
// 		Info.Logf("IndexYamlStackTemplate Id: %#v", template.ID)
// 		Info.Logf("IndexYamlStackTemplate Url: %#v", template.URL)
// 	}
// 	for _, maintainer := range stack.Maintainers {
// 		Info.Logf("IndexYamlStackMaintainer Name: %#v", maintainer.Name)
// 		Info.Logf("IndexYamlStackMaintainer Email: %#v", maintainer.Email)
// 		Info.Logf("IndexYamlStackMaintainer Github-id: %#v", maintainer.GithubID)
// 	}
// }
// *********************************************************************************
