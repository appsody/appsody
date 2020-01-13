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
	"errors"
	"github.com/spf13/cobra"
)

func newDebugCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &devCommonConfig{RootCommandConfig: rootConfig}
	// debug Cmd represents the debug command
	var debugCmd = &cobra.Command{
		Use:   "debug",
		Short: "Debug your Appsody project.",
		Long: `Start a container-based continuous build environment for your Appsody project, with debugging enabled.
		
Run this command from the root directory of your Appsody project.`,
		Example: `  appsody debug --docker-options "--privileged"
  Starts the debugging environment, passing the "--privileged" option to the "docker run" command as a flag.
  
  appsody debug --name my-project-dev2 -p 3001:3000
  Starts the debugging environment, names the development container "my-project-dev2", and binds the container port 3000 to the host port 3001.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) > 0 {
				return errors.New("Expected no additional arguments")
			}
			config.Info.log("Running debug environment")
			return commonCmd(config, "debug")
		},
	}

	addDevCommonFlags(debugCmd, config)
	return debugCmd
}
