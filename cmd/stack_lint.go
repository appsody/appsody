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
		Use:   "lint [path]",
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
			if len(args) > 1 {
				return errors.Errorf("Too many arguments. Use 'appsody [command] --help' for more information about a command")
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

			var lintErrorFiles = []struct {
				prefix string
				name   string
				err    bool
			}{
				{stackPath, "README.md", true},
				{stackPath, "stack.yaml", true},
				{imagePath, "", true},
				{imagePath, "Dockerfile-stack", true},
				{imagePath, "LICENSE", true},
				{templatePath, "", true},
				{configPath, "", false},
				{configPath, "app-deploy.yaml", false},
				{projectPath, "", false},
				{projectPath, "Dockerfile", false},
			}

			for _, file := range lintErrorFiles {
				fileCheck, err := Exists(filepath.Join(file.prefix, file.name))
				if err != nil {
					rootConfig.Error.log("Error attempting to determine file: ", err)
					stackLintErrorCount++
				} else if !fileCheck {
					if file.name == "" {
						rootConfig.Error.log("Missing directory: ", file.prefix)
					} else {
						rootConfig.Error.log("Missing file: ", file.name, " in ", file.prefix)
					}
					if file.err {
						stackLintErrorCount++
					} else {
						stackLintWarningCount++
					}
				}
			}

			if IsEmptyDir(templatePath) {
				rootConfig.Error.log("No templates found in: ", templatePath)
				stackLintErrorCount++
			}

			templates, _ := ioutil.ReadDir(templatePath)
			for _, f := range templates {
				fileCheck, err := Exists(filepath.Join(templatePath, f.Name(), ".appsody-config.yaml"))
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
