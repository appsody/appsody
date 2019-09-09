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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint your stack to verify that it conforms to the standard of an Appsody stack",
	Long: `This command will validate that your stack has the structure of an Appsody stack. It will inform you of files/directories
missing and warn you if your stack could be enhanced.

This command can be run from the base directory of your stack or you can supply a path to the stack as an argument.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		stackPath, _ := os.Getwd()
		errorCount := 0
		warningCount := 0

		if len(args) > 0 {
			stackPath = args[0]
		}

		imagePath := filepath.Join(stackPath, "image")
		templatePath := filepath.Join(stackPath, "/templates")
		configPath := filepath.Join(imagePath, "/config")
		projectPath := filepath.Join(imagePath, "/project")

		Info.log("LINTING ", path.Base(stackPath))

		if fileDoesNotExist(filepath.Join(stackPath, "/README.md")) != nil {
			Error.log("Missing README.md in: ", stackPath)
			errorCount++
		}

		if fileDoesNotExist(filepath.Join(stackPath, "/stack.yaml")) != nil {
			Error.log("Missing stack.yaml in: ", stackPath)
			errorCount++
		}

		if fileDoesNotExist(imagePath) != nil {
			Error.log("Missing image directory in ", stackPath)
			errorCount++
		}

		if fileDoesNotExist(filepath.Join(imagePath, "/Dockerfile-stack")) != nil {
			Error.log("Missing Dockerfile-stack in ", imagePath)
			errorCount++
		}

		if fileDoesNotExist(filepath.Join(imagePath, "/LICENSE")) != nil {
			Error.log("Missing LICENSE in ", imagePath)
			errorCount++
		}

		if fileDoesNotExist(configPath) != nil {
			Warning.log("Missing config directory in ", imagePath, " (Knative deployment will be used over Kubernetes)")
			warningCount++

		}

		if fileDoesNotExist(filepath.Join(configPath, "/app-deploy.yaml")) != nil {
			Warning.log("Missing app-deploy.yaml in ", configPath, " (Knative deployment will be used over Kubernetes)")
			warningCount++
		}

		if fileDoesNotExist(filepath.Join(projectPath, "/Dockerfile")) != nil {
			Warning.log("Missing Dockerfile in ", projectPath)
			warningCount++
		}

		if fileDoesNotExist(templatePath) != nil {
			Error.log("Missing template directory in: ", stackPath)
			errorCount++
		}

		if IsEmptyDir(templatePath) != nil {
			Error.log("No templates found in: ", templatePath)
			errorCount++
		}

		templates, _ := ioutil.ReadDir(templatePath)
		for _, f := range templates {
			if fileDoesNotExist(filepath.Join(templatePath, f.Name(), ".appsody-config.yaml")) == nil {
				Info.log("ERROR: Unexpected .appsody-config.yaml in ", filepath.Join(templatePath, f.Name()))
				errorCount++
			}
		}

		if errorCount > 0 {
			Info.log("TOTAL ERRORS: ", errorCount)
			Info.log("TOTAL WARNINGS: ", warningCount)
			return errors.Errorf("LINT TEST FAILED")

		}

		Info.log("TOTAL WARNINGS: ", warningCount)
		Info.log("LINT TEST PASSED")
		return nil
	},
}

func IsEmptyDir(name string) error {
	_, err := ioutil.ReadDir(name)
	return err
}

func fileDoesNotExist(filename string) error {
	_, err := os.Stat(filename)
	return err
}

func init() {
	stackCmd.AddCommand(lintCmd)

}
