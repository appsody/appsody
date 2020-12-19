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
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newVersionCmd(log *LoggingConfig, rootCmd *cobra.Command) *cobra.Command {
	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version of the Appsody CLI and Controller.",
		Long:  `Show the version of the Appsody CLI and Controller that is currently in use.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("Unexpected argument. Use 'appsody [command] --help' for more information about a command")
			}
			// default value in main.go is vlatest
			log.Info.log(rootCmd.Use, " ", VERSION)
			overrideControllerImage := os.Getenv("APPSODY_CONTROLLER_IMAGE")
			if overrideControllerImage == "" {
				log.Debug.Log("Using default controller image...")
				overrideVersion := os.Getenv("APPSODY_CONTROLLER_VERSION")
				if overrideVersion != "" {
					CONTROLLERVERSION = overrideVersion
				}
			} else {
				log.Info.Log("Overriding default controller image with: " + overrideControllerImage)
				imageSplit := strings.Split(overrideControllerImage, ":")
				if len(imageSplit) == 1 {
					// this is an implicit reference to latest
					CONTROLLERVERSION = "latest"
				} else {
					CONTROLLERVERSION = imageSplit[1]
					//This also could be latest
				}
			}
			// default value in main.go is latest
			log.Info.log("appsody-controller", " ", CONTROLLERVERSION)
			return nil
		},
	}

	return versionCmd
}
