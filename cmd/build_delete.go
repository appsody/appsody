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
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// deleteCmd provides the ability to delete a GitHook for a Tekton build pipeline
var deleteCmd = &cobra.Command{
	Use: "delete",
	// disable this command until we have a better plan on how to support ci pipelines
	Hidden: true,
	Short:  "Delete a Githook and build pipeline for your Appsody project",
	Long:   `This allows you to delete a Githook for your Appsody project.`,
	Run: func(cmd *cobra.Command, args []string) {
		// projectDir := getProjectDir()
		// projectName := filepath.Base(projectDir)
		projectName, perr := getProjectName()
		if perr != nil {
			Error.log("Not a valid Appsody project: ", perr)
			os.Exit(1)
		}
		tektonServer := cliConfig.GetString("tektonserver")
		if tektonServer == "" {
			Error.log("No target Tekton server specified in the configuration.")
			os.Exit(1)
		}
		url := fmt.Sprintf("%s/v1/namespaces/default/githubsource/%s", tektonServer, projectName)
		if dryrun {
			Info.log("Dry Run appsody build delete")
		} else {
			req, _ := http.NewRequest("DELETE", url, nil)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			Info.log("Making request to ", url)
			resp, err := client.Do(req)
			if err != nil {
				Error.log(err)
				os.Exit(1)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			bodyStr := string(body)

			if resp.StatusCode >= 300 {
				Error.log(resp.Status)
				Error.log(string(bodyStr))
				os.Exit(1)
			} else {
				Info.log(resp.Status)
				Info.log(string(bodyStr))
			}
		}
	},
}

func init() {
	buildCmd.AddCommand(deleteCmd)

}
