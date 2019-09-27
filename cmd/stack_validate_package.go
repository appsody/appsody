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
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func check(e error) {
	if e != nil {
		Error.log("Something went wrong")
		Error.log(e)
		return
	}
}

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

// stackValidateCmd represents the validate command
var stackPackageCmd = &cobra.Command{
	Use:   "package",
	Short: "Package a stack in the local Appsody environment",
	Long:  `This builds a stack and creates an index and adds it to the repository`,
	RunE: func(cmd *cobra.Command, args []string) error {

		Info.log("******************************************")
		Info.log("Running appsody stack package")
		Info.log("******************************************")

		stackPath, _ := os.Getwd()
		Info.log("stackPath is: ", stackPath)

		appsodyHome := getHome()
		Info.log("appsodyHome is:", appsodyHome)

		err := os.Chdir(filepath.Join("..", "..", "ci", "assets")))

		err := os.Chdir(filepath.Join("..", "..", "ci", "assets"))
		if err != nil {
			// if we can't find the assets directory then we are not starting from a valid root of the stack directory
			Error.log("Unable to reach assets directory. Current directory must be the root of the stack.")
			return err
		}

		assetsDir, _ := os.Getwd()
		Info.log("assetsDir is: ", assetsDir)

		stackPathSplit := strings.Split(stackPath, string(filepath.Separator))
		stackName := stackPathSplit[len(stackPathSplit)-1]
		Info.log("stackName is: ", stackName)

		repoName := stackPathSplit[len(stackPathSplit)-2]
		Info.log("repoName is: ", repoName)

		indexFileLocal := filepath.Join(assetsDir, repoName) + "-index-local.yaml"
		Info.log("indexFileLocal is: ", indexFileLocal)

		err = os.Chdir(stackPath)
		check(err)

		// create incubator-index.yaml and put it in ci/assets

		f, err := os.Create(indexFileLocal)
		check(err)
		defer f.Close()
		n, err := f.WriteString("apiVersion: v2\n")
		check(err)
		Info.log("wrote bytes: ", n)
		n, err = f.WriteString("stacks\n")
		check(err)
		Info.log("wrote bytes: ", n)
		n, err = f.WriteString("  - id: " + stackName + "\n")
		check(err)
		Info.log("wrote bytes: ", n)

		var stackYaml StackYaml
		source, err := ioutil.ReadFile("stack.yaml")
		check(err)

		err = yaml.Unmarshal(source, &stackYaml)
		check(err)

		fmt.Printf("StackYaml Name: %#v\n", stackYaml.Name)
		fmt.Printf("StackYaml Version: %#v\n", stackYaml.Version)
		fmt.Printf("StackYaml Description: %#v\n", stackYaml.Description)
		fmt.Printf("StackYaml License: %#v\n", stackYaml.License)
		fmt.Printf("StackYaml Language: %#v\n", stackYaml.Language)
		fmt.Printf("StackYaml DefaultTemplate: %#v\n", stackYaml.DefaultTemplate)

		for i := range stackYaml.Maintainers {
			fmt.Printf("Maintainers Name: %#v\n", stackYaml.Maintainers[i].Name)
			fmt.Printf("Maintainers Email: %#v\n", stackYaml.Maintainers[i].Email)
			fmt.Printf("Maintainers GithubID: %#v\n", stackYaml.Maintainers[i].GithubID)
		}

		n, err = f.WriteString("    name: " + stackYaml.Name + "\n")
		check(err)
		Info.log("wrote bytes: ", n)

		n, err = f.WriteString("    version: " + stackYaml.Version + "\n")
		check(err)
		Info.log("wrote bytes: ", n)

		n, err = f.WriteString("    description: " + stackYaml.Description + "\n")
		check(err)
		Info.log("wrote bytes: ", n)

		n, err = f.WriteString("    license: " + stackYaml.License + "\n")
		check(err)
		Info.log("wrote bytes: ", n)

		n, err = f.WriteString("    language: " + stackYaml.Language + "\n")
		check(err)
		Info.log("wrote bytes: ", n)

		n, err = f.WriteString("    maintainers:\n")
		check(err)
		Info.log("wrote bytes: ", n)

		for i := range stackYaml.Maintainers {
			n, err := f.WriteString("     - name: " + stackYaml.Maintainers[i].Name + "\n")
			check(err)
			Info.log("wrote bytes: ", n)

			n, err = f.WriteString("       email: " + stackYaml.Maintainers[i].Email + "\n")
			check(err)
			Info.log("wrote bytes: ", n)

			n, err = f.WriteString("       github-id: " + stackYaml.Maintainers[i].GithubID + "\n")
			check(err)
			Info.log("wrote bytes: ", n)
		}

		n, err = f.WriteString("    default-template: " + stackYaml.DefaultTemplate + "\n")
		check(err)
		Info.log("wrote bytes: ", n)

		n, err = f.WriteString("    templates:\n")
		check(err)
		Info.log("wrote bytes: ", n)

		// we still need the url for the index but we will write it while taring the templates

		// docker build

		buildImage := "appsody/appsody-index:SNAPSHOT"

		err = os.Chdir(filepath.Join(stackPath, "image"))
		check(err)

		imageDir, _ := os.Getwd()
		dockerFile := imageDir + string(filepath.Separator) + "Dockerfile-stack"
		Info.log("dockerFile is: ", dockerFile)

		imageDir = imageDir + string(filepath.Separator)

		cmdArgs := []string{"-t", buildImage}

		cmdArgs = append(cmdArgs, "-f", dockerFile, imageDir)
		Info.log("cmdArgs is: ", cmdArgs)

		err = DockerBuild(cmdArgs, DockerLog)
		check(err)

		// tar the templates

		templatePath := filepath.Join(stackPath, "templates")

		t, err := os.Open(templatePath)
		check(err)

		templates, err := t.Readdirnames(0)
		check(err)

		// loop through the template directories
		// write the template url in the index yaml
		// create a tar.gz for each template
		for i := range templates {
			Info.log("template is: ", templates[i])
			sourceDir := stackPath + string(filepath.Separator) + "templates" + string(filepath.Separator) + templates[i]
			Info.log("sourceDir is: ", sourceDir)

			versionedArchive := assetsDir + string(filepath.Separator) + repoName + "." + stackName + ".v" + stackYaml.Version + ".templates."
			Info.log("versionedArdhive is: ", versionedArchive)

			versionArchiveTar := versionedArchive + templates[i] + ".tar.gz"

			// write the template url in the index yaml
			n, err = f.WriteString("      - id: " + templates[i] + "\n")
			check(err)
			Info.log("wrote bytes: ", n)

			n, err = f.WriteString("        url: file://" + versionArchiveTar + "\n")
			check(err)
			Info.log("wrote bytes: ", n)

			err := Targz(sourceDir, versionedArchive)
			check(err)
		}

		t.Close()

		return nil

	},
}

func init() {
	// will use stackCmd eventually
	stackCmd.AddCommand(stackPackageCmd)

}
