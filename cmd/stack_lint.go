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
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var stackLintErrorCount int
var stackLintWarningCount int

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint your stack to verify that it conforms to the standard of an Appsody stack",
	Long: `This command will validate that your stack has the structure of an Appsody stack. It will inform you of files/directories
missing and warn you if your stack could be enhanced.

This command can be run from the base directory of your stack or you can supply a path to the stack as an argument.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		stackPath, _ := os.Getwd()

		if len(args) > 0 {
			stackPath = args[0]
		}

		imagePath := filepath.Join(stackPath, "image")
		templatePath := filepath.Join(stackPath, "/templates")
		configPath := filepath.Join(imagePath, "/config")
		projectPath := filepath.Join(imagePath, "/project")

		Info.log("LINTING ", path.Base(stackPath))

		fileCheck, err := Exists(filepath.Join(stackPath, "/README.md"))
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Error.log("Missing README.md in: ", stackPath)
			stackLintErrorCount++
		}
		fileCheck, err = Exists(filepath.Join(stackPath, "/stack.yaml"))
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Error.log("Missing stack.yaml in: ", stackPath)
			stackLintErrorCount++
		}

		fileCheck, err = Exists(imagePath)
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Error.log("Missing image directory in ", stackPath)
			stackLintErrorCount++
		}

		fileCheck, err = Exists(filepath.Join(imagePath, "/Dockerfile-stack"))
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Error.log("Missing Dockerfile-stack in ", imagePath)
			stackLintErrorCount++
		}

		fileCheck, err = Exists(filepath.Join(imagePath, "/LICENSE"))
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Error.log("Missing LICENSE in ", imagePath)
			stackLintErrorCount++
		}

		fileCheck, err = Exists(configPath)
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Warning.log("Missing config directory in ", imagePath, " (Knative deployment will be used over Kubernetes)")
			stackLintWarningCount++
		}

		fileCheck, err = Exists(filepath.Join(configPath, "/app-deploy.yaml"))
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Warning.log("Missing app-deploy.yaml in ", configPath, " (Knative deployment will be used over Kubernetes)")
			stackLintWarningCount++
		}

		fileCheck, err = Exists(filepath.Join(projectPath, "/Dockerfile"))
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Warning.log("Missing Dockerfile in ", projectPath)
			stackLintWarningCount++
		}

		fileCheck, err = Exists(templatePath)
		if err != nil {
			Error.log("Error attempting to determine file: ", err)
			stackLintErrorCount++
		} else if !fileCheck {
			Error.log("Missing template directory in: ", stackPath)
			stackLintErrorCount++
		}

		if IsEmptyDir(templatePath) {
			Error.log("No templates found in: ", templatePath)
			stackLintErrorCount++
		}

		templates, _ := ioutil.ReadDir(templatePath)
		for _, f := range templates {
			fileCheck, err = Exists(filepath.Join(templatePath, f.Name(), ".appsody-config.yaml"))
			if (err != nil) && f.Name() != ".DS_Store" {
				Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if fileCheck && f.Name() != ".DS_Store" {
				Error.log("Unexpected .appsody-config.yaml in ", filepath.Join(templatePath, f.Name()))
				stackLintErrorCount++
			}
		}

		lintDockerFileStack(stackPath)

		var s StackDetails
		s.validateYaml(stackPath)

		Info.log("TOTAL ERRORS: ", stackLintErrorCount)
		Info.log("TOTAL WARNINGS: ", stackLintWarningCount)

		if stackLintErrorCount > 0 {
			return errors.Errorf("LINT TEST FAILED")
		}

		Info.log("LINT TEST PASSED")
		return nil
	},
}

func init() {
	stackCmd.AddCommand(lintCmd)
}
