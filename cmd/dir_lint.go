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
	"github.com/spf13/cobra"
	"os"
	"io/ioutil"
)

// listCmd represents the list command
var lintCmd = &cobra.Command{
	Use:   "lint ",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		stackPath := os.Getenv("PWD")
		templatePath := os.Getenv("PWD") + "/templates"

		if fileExists(templatePath) {
			Info.log("ERROR: Missing template directory in: " + stackPath )
		} 
		
		if IsEmptyDir(templatePath) {
			Info.log("ERROR: No templates found in: " + templatePath )
		}

		if fileExists(stackPath + "/README.md") {
			Info.log("ERROR: Missing README.md in: " + stackPath)
		}

		if fileExists(stackPath + "/stack.yaml") {
			Info.log("ERROR: Missing stack.yaml in: " + stackPath)
		}


		return nil
	},
}

func fileExists(filename string) (bool) {
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return true;
    } else {
		return false
	}
    
}

func IsEmptyDir(name string) (bool) {
	_, err := ioutil.ReadDir(name)
	if err != nil {
		return true
	} else {
		return false
	}
}

func init() {
	stackCmd.AddCommand(lintCmd)

}
