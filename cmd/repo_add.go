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
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

func newRepoAddCmd(config *RootCommandConfig) *cobra.Command {
	// initCmd represents the init command
	var addCmd = &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add an Appsody repository.",
		Long:  `Add an Appsody repository to your list of configured Appsody repositories.`,
		Example: `  appsody repo add my-local-repo file://path/to/my-local-repo.yaml
  Adds the "my-local-repo" repository, specified by the "file://path/to/my-local-repo.yaml" file to your list of repositories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {

				return errors.New("Error, you must specify repository name and URL")
			}

			var repoName = args[0]
			var repoURL = args[1]

			if len(repoName) > 50 {
				return errors.Errorf("Invalid repository name. The <name> must be less than 50 characters")

			}
			match, _ := regexp.MatchString("^[a-zA-Z0-9\\-_\\.]{1,50}$", repoName)
			if !match {
				return errors.Errorf("Invalid repository name. The name can contain only letters (lowercase or uppercase), numbers, dashes '-', underscores '_', and periods '.'")

			}

			var repoFile RepositoryFile

			_, repoErr := repoFile.getRepos(config)
			if repoErr != nil {
				return repoErr
			}
			if repoFile.Has(repoName) {
				return errors.Errorf("A repository with the name '%s' already exists.", repoName)

			}
			if repoFile.HasURL(repoURL) {
				return errors.Errorf("A repository with the URL '%s' already exists.", repoURL)

			}
			index, err := downloadIndex(config.LoggingConfig, repoURL)
			if err != nil {

				return err
			}
			if strings.Compare(index.APIVersion, supportedIndexAPIVersion) == 1 {
				config.Warning.log("The repository " + repoName + " contains an APIVersion in its .yaml file more recent than the current Appsody CLI supports(" + supportedIndexAPIVersion + "), it is strongly suggested that you update your Appsody CLI to the latest version.")
			}

			if config.Dryrun {
				config.Info.logf("Dry Run - Skipping appsody repo add repository Name: %s, URL: %s", repoName, repoURL)
			} else {
				var newEntry = RepositoryEntry{
					Name: repoName,
					URL:  repoURL,
				}

				repoFile.Add(&newEntry)
				err = repoFile.WriteFile(getRepoFileLocation(config.CliConfig))
				if err != nil {
					return errors.Errorf("Failed to write file to repository location: %v", err)
				}
			}
			return nil
		},
	}
	return addCmd
}
