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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// setupCmd allows you to setup a GitHook to drive a Tekton build pipeline for the Appsodys project in Git
var setupCmd = &cobra.Command{
	Use: "setup",
	// disable this command until we have a better plan on how to support ci pipelines
	Hidden: true,
	Short:  "Setup a Githook and build pipeline for your Appsody project",
	Long:   `This allows you to register a Githook for your Appsody project.`,
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: should we dynamically pick up the Git URL from the .git in the project?
		// TODO: add validation of the supplied Git URL
		if len(args) < 1 {
			Error.log("Error, you must specify a Git project URL")
			os.Exit(1)
		}
		gitProject := args[0]

		// Use the "tektonserver" field from the config.
		tektonServer := cliConfig.GetString("tektonserver")
		if tektonServer == "" {
			Error.log("No target Tekton server specified in the configuration.")
			os.Exit(1)
		}
		url := fmt.Sprintf("%s/v1/namespaces/default/githubsource/", tektonServer)

		// projectDir := getProjectDir()
		// projectName := filepath.Base(projectDir)
		projectName := getProjectName()

		// Setup JSON payload for use with the Tekton server
		var jsonStr = fmt.Sprintf(`{"name":"%s", "gitrepositoryurl":"%s","accesstoken":"github-secret","pipeline":"appsody-build-pipeline"}`, projectName, gitProject)
		if dryrun {
			Info.logf("Dry Run appsody build setup project URL: %s\n", url)
		} else {
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
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
	buildCmd.AddCommand(setupCmd)
}
