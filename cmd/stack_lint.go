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
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newStackLintCmd(rootConfig *RootCommandConfig) *cobra.Command {
	var lintCmd = &cobra.Command{
		Use:   "lint",
		Short: "Check your stack structure.",
		Long: `Check that the structure of your stack is valid. Error messages indicate critical issues in your stack structure, such as missing files, directories, or stack variables. Warning messages suggest optional stack enhancements.

Run this command from the root directory of your stack, or specify the path to your stack.`,
		Example: `  appsody stack lint
  Checks the structure of the stack in the current directory"
		
  appsody stack lint path/to/my-stack
  Checks the structure of the stack "my-stack" in the path "path/to/my-stack"`,
		RunE: func(cmd *cobra.Command, args []string) error {

			var stackLintErrorCount int
			var stackLintWarningCount int

			stackPath := rootConfig.ProjectDir

			if len(args) > 0 {
				stackPath = args[0]
			}

			imagePath := filepath.Join(stackPath, "image")
			templatePath := filepath.Join(stackPath, "/templates")
			configPath := filepath.Join(imagePath, "/config")
			projectPath := filepath.Join(imagePath, "/project")

			stackID := filepath.Base(stackPath)
			rootConfig.Info.log("LINTING ", stackID)

			validStackID, err := IsValidProjectName(stackID)
			if !validStackID {
				rootConfig.Error.log("Stack directory name is invalid. ", err)
				stackLintErrorCount++
			}

			fileCheck, err := Exists(filepath.Join(stackPath, "/README.md"))
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Error.log("Missing README.md in: ", stackPath)
				stackLintErrorCount++
			}
			fileCheck, err = Exists(filepath.Join(stackPath, "/stack.yaml"))
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Error.log("Missing stack.yaml in: ", stackPath)
				stackLintErrorCount++
			}

			fileCheck, err = Exists(imagePath)
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Error.log("Missing image directory in ", stackPath)
				stackLintErrorCount++
			}

			fileCheck, err = Exists(filepath.Join(imagePath, "/Dockerfile-stack"))
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Error.log("Missing Dockerfile-stack in ", imagePath)
				stackLintErrorCount++
			}

			fileCheck, err = Exists(filepath.Join(imagePath, "/LICENSE"))
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Error.log("Missing LICENSE in ", imagePath)
				stackLintErrorCount++
			}

			fileCheck, err = Exists(configPath)
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Warning.log("Missing config directory in ", imagePath, " (Knative deployment will be used over Kubernetes)")
				stackLintWarningCount++
			}

			fileCheck, err = Exists(filepath.Join(configPath, "/app-deploy.yaml"))
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Warning.log("Missing app-deploy.yaml in ", configPath, " (Knative deployment will be used over Kubernetes)")
				stackLintWarningCount++
			}

			fileCheck, err = Exists(filepath.Join(projectPath, "/Dockerfile"))
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Warning.log("Missing Dockerfile in ", projectPath)
				stackLintWarningCount++
			}

			fileCheck, err = Exists(templatePath)
			if err != nil {
				rootConfig.Error.log("Error attempting to determine file: ", err)
				stackLintErrorCount++
			} else if !fileCheck {
				rootConfig.Error.log("Missing template directory in: ", stackPath)
				stackLintErrorCount++
			}

			if IsEmptyDir(templatePath) {
				rootConfig.Error.log("No templates found in: ", templatePath)
				stackLintErrorCount++
			}

			templates, _ := ioutil.ReadDir(templatePath)
			for _, f := range templates {
				fileCheck, err = Exists(filepath.Join(templatePath, f.Name(), ".appsody-config.yaml"))
				if (err != nil) && f.Name() != ".DS_Store" {
					rootConfig.Error.log("Error attempting to determine file: ", err)
					stackLintErrorCount++
				} else if fileCheck && f.Name() != ".DS_Store" {
					rootConfig.Error.log("Unexpected .appsody-config.yaml in ", filepath.Join(templatePath, f.Name()))
					stackLintErrorCount++
				}
			}

			dockerFileErrorCount, dockerFileWarningCount := lintDockerFileStack(rootConfig.LoggingConfig, stackPath)
			stackLintErrorCount += dockerFileErrorCount
			stackLintWarningCount += dockerFileWarningCount

			var stackDetails StackYaml
			stackYamlErrorCount, stackYamlWarningCount := stackDetails.validateYaml(rootConfig, stackPath)
			stackLintErrorCount += stackYamlErrorCount
			stackLintWarningCount += stackYamlWarningCount

			rootConfig.Info.log("TOTAL ERRORS: ", stackLintErrorCount)
			rootConfig.Info.log("TOTAL WARNINGS: ", stackLintWarningCount)

			if stackLintErrorCount > 0 {
				return errors.Errorf("LINT TEST FAILED")
			}

			rootConfig.Info.log("LINT TEST PASSED")
			return nil
		},
	}
	return lintCmd
}
