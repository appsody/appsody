// Copyright © 2019 The Appsody Authors
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
	"path/filepath"

	"github.com/spf13/cobra"
)

// buildCmd provides the ability run local builds, or setup/delete Tekton builds, for an appsody project
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Locally build a docker image of your appsody project",
	Long:  `This allows you to build a local Docker image from your Appsody project. Extract is run before the docker build.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This needs to do:
		// 1. appsody Extract
		// 2. docker build -t <project name> -f Dockerfile ./extracted

		extractCmd.Run(cmd, args)

		projectDir := getProjectDir()
		projectName := filepath.Base(projectDir)
		extractDir := filepath.Join(getHome(), "extract", projectName)
		dockerfile := filepath.Join(extractDir, "Dockerfile")

		cmdName := "docker"
		cmdArgs := []string{"build", "-t", projectName, "-f", dockerfile, extractDir}
		execAndWait(cmdName, cmdArgs, Debug)
		if !dryrun {
			Info.log("Built docker image ", projectName)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
