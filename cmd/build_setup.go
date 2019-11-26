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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newSetupCmd(config *buildCommandConfig) *cobra.Command {
	// setupCmd allows you to setup a GitHook to drive a Tekton build pipeline for the Appsodys project in Git
	var setupCmd = &cobra.Command{
		Use: "setup",
		// disable this command until we have a better plan on how to support ci pipelines
		Hidden: true,
		Short:  "Setup a Githook and build pipeline for your Appsody project",
		Long:   `This allows you to register a Githook for your Appsody project.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			// TODO: should we dynamically pick up the Git URL from the .git in the project?
			// TODO: add validation of the supplied Git URL
			if len(args) < 1 {
				return errors.New("error, you must specify a Git project URL")

			}
			gitProject := args[0]

			// Use the "tektonserver" field from the config.
			tektonServer := config.CliConfig.GetString("tektonserver")
			if tektonServer == "" {
				return errors.New("no target Tekton server specified in the configuration")
			}
			url := fmt.Sprintf("%s/v1/namespaces/default/githubsource/", tektonServer)

			// projectDir := getProjectDir()
			// projectName := filepath.Base(projectDir)
			projectName, perr := getProjectName(config.RootCommandConfig)
			if perr != nil {
				return errors.Errorf("%v", perr)
			}
			// Setup JSON payload for use with the Tekton server
			var jsonStr = fmt.Sprintf(`{"name":"%s", "gitrepositoryurl":"%s","accesstoken":"github-secret","pipeline":"appsody-build-pipeline"}`, projectName, gitProject)
			if config.Dryrun {
				config.Info.logf("Dry Run appsody build setup project URL: %s\n", url)
			} else {
				req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				config.Info.log("Making request to ", url)
				resp, err := client.Do(req)
				if err != nil {
					return errors.Errorf("%v", perr)
				}
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				bodyStr := string(body)

				if resp.StatusCode >= 300 {
					return errors.Errorf("Bad Status Code: %s with message: %s", resp.Status, string(bodyStr))
				}
				config.Info.log(resp.Status)
				config.Info.log(string(bodyStr))

			}
			return nil
		},
	}

	return setupCmd
}
