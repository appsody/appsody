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
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// "github.com/google/go-github/v28/github"	// with go modules enabled (GO111MODULE=on or outside GOPATH)

	"github.com/google/go-github/github" // with go modules disabled
	"golang.org/x/oauth2"
)

// Model
type Package struct {
	FullName      string
	Description   string
	StarsCount    int
	ForksCount    int
	LastUpdatedBy string
}

func newStackCreateCmd(rootConfig *RootCommandConfig) *cobra.Command {
	var stackCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new stack as a copy of an existing stack",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			context := context.Background()
			tokenService := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: ""},
			)
			tokenClient := oauth2.NewClient(context, tokenService)

			client := github.NewClient(tokenClient)

			repo, _, err := client.Repositories.Get(context, "appsody", "stacks")

			fmt.Println(repo)

			if err != nil {
				fmt.Printf("Problem in getting repository information %v\n", err)
				os.Exit(1)
			}

			pack := &Package{
				FullName:    *repo.FullName,
				Description: *repo.Description,
				ForksCount:  *repo.ForksCount,
				StarsCount:  *repo.StargazersCount,
			}

			fmt.Printf("%+v\n", pack)

			// // list public repositories for org "github"
			// opt := &github.RepositoryListByOrgOptions{Type: "public"}
			// // repos, _, err := client.Repositories.ListByOrg(context.Background(), "appsody", opt)
			// repos := client.Repositories.Get(context.CancelFunc(), "appsody", "stacks")

			// Info.log(repos)
			return nil
		},
	}
	return stackCmd
}
