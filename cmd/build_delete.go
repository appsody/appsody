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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newBuildDeleteCmd(config *buildCommandConfig) *cobra.Command {
	// deleteCmd provides the ability to delete a GitHook for a Tekton build pipeline
	var deleteCmd = &cobra.Command{
		Use: "delete",
		// disable this command until we have a better plan on how to support ci pipelines
		Hidden: true,
		Short:  "Delete a Githook and build pipeline for your Appsody project",
		Long:   `This allows you to delete a Githook for your Appsody project.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			projectName, perr := getProjectName(config.RootCommandConfig)
			if perr != nil {
				return errors.Errorf("%v", perr)

			}
			tektonServer := config.CliConfig.GetString("tektonserver")
			if tektonServer == "" {
				return errors.New("no target Tekton server specified in the configuration")

			}
			url := fmt.Sprintf("%s/v1/namespaces/default/githubsource/%s", tektonServer, projectName)
			if config.Dryrun {
				config.Info.log("Dry Run appsody build delete")
			} else {
				req, _ := http.NewRequest("DELETE", url, nil)
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				config.Info.log("Making request to ", url)
				resp, err := client.Do(req)
				if err != nil {
					return errors.Errorf("%v", err)

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
	return deleteCmd
}
