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
	"strconv"

	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint your stack to verify that it conforms to the standard of an Appsody stack",
	Long: `This command will validate that your stack has the structure of an Appsody stack. It will inform you of files/directories
missing and warn you if your stack could be enhanced.

This command can be run from the base directory of your stack or you can supply a path to the stack as an argument.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		stackPath := os.Getenv("PWD")
		errorCount := 0
		warningCount := 0

		if len(args) > 0 {
			stackPath = args[0]
		}

		imagePath := stackPath + "/image"
		templatePath := stackPath + "/templates"
		configPath := imagePath + "/config"
		projectPath := imagePath + "/project"

		Info.log("LINTING " + path.Base(stackPath) + "\n")

		if fileDoesNotExist(stackPath + "/README.md") {
			Info.log("ERROR: Missing README.md in: " + stackPath)
			errorCount++
		}

		if fileDoesNotExist(stackPath + "/stack.yaml") {
			Info.log("ERROR: Missing stack.yaml in: " + stackPath)
			errorCount++
		}

		if fileDoesNotExist(imagePath) {
			Info.log("ERROR: Missing image directory in " + stackPath)
			errorCount++
		}

		if fileDoesNotExist(imagePath + "/Dockerfile-stack") {
			Info.log("ERROR: Missing Dockerfile-stack in " + imagePath)
			errorCount++
		}

		if fileDoesNotExist(imagePath + "/LICENSE") {
			Info.log("ERROR: Missing LICENSE in " + imagePath)
			errorCount++
		}

		if fileDoesNotExist(configPath) {
			Info.log("WARNING: Missing config directory in " + imagePath + " (Knative deployment will be used over Kubernetes)")
			warningCount++

		}

		if fileDoesNotExist(configPath + "/app-deploy.yaml") {
			Info.log("WARNING: Missing app-deploy.yaml in " + configPath + " (Knative deployment will be used over Kubernetes)")
			warningCount++
		}

		if fileDoesNotExist(projectPath + "/Dockerfile") {
			Info.log("WARNING: Missing Dockerfile in " + projectPath)
			warningCount++
		}

		if fileDoesNotExist(templatePath) {
			Info.log("ERROR: Missing template directory in: " + stackPath)
			errorCount++
		}

		if IsEmptyDir(templatePath) {
			Info.log("ERROR: No templates found in: " + templatePath)
			errorCount++
		}

		templates, _ := ioutil.ReadDir(templatePath)
		for _, f := range templates {
			if !fileDoesNotExist(templatePath + "/" + f.Name() + "/" + ".appsody-config.yaml") {
				Info.log("ERROR: Unexpected .appsody-config.yaml in " + templatePath + "/" + f.Name())
				errorCount++
			}
		}

		if errorCount > 0 {
			Info.log("\nLINT TEST FAILED")
			Info.log("\nTOTAL ERRORS: " + strconv.Itoa(errorCount))
			Info.log("TOTAL WARNINGS: " + strconv.Itoa(warningCount))

		} else {
			Info.log("\nLINT TEST PASSED")
			Info.log("TOTAL WARNINGS: " + strconv.Itoa(warningCount))
		}

		return nil
	},
}

func init() {
	stackCmd.AddCommand(lintCmd)

}
