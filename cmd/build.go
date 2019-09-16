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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func checkDockerBuildOptions(options []string) error {
	buildOptionsTest := "(^((-t)|(--tag)|(-f)|(--file))((=?$)|(=.*)))"

	blackListedBuildOptionsRegexp := regexp.MustCompile(buildOptionsTest)
	for _, value := range options {
		isInBlackListed := blackListedBuildOptionsRegexp.MatchString(value)
		if isInBlackListed {
			return errors.Errorf("%s is not allowed in --docker-options", value)

		}
	}
	return nil

}

var dockerBuildOptions string

// buildCmd provides the ability run local builds, or setup/delete Tekton builds, for an appsody project
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Locally build a docker image of your appsody project",
	Long:  `This allows you to build a local Docker image from your Appsody project. Extract is run before the docker build.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// This needs to do:
		// 1. appsody Extract
		// 2. docker build -t <project name> -f Dockerfile ./extracted

		extractErr := extractCmd.RunE(cmd, args)
		if extractErr != nil {
			return extractErr
		}

		projectName, perr := getProjectName()
		if perr != nil {
			return errors.Errorf("%v", perr)
		}
		extractDir := filepath.Join(getHome(), "extract", projectName)
		dockerfile := filepath.Join(extractDir, "Dockerfile")
		buildImage := projectName //Lowercased
		// If a tag is specified, change the buildImage
		if tag != "" {
			buildImage = tag
		}
		//cmdName := "docker"
		cmdArgs := []string{"-t", buildImage}

		if dockerBuildOptions != "" {
			dockerBuildOptions = strings.TrimPrefix(dockerBuildOptions, " ")
			dockerBuildOptions = strings.TrimSuffix(dockerBuildOptions, " ")
			options := strings.Split(dockerBuildOptions, " ")
			err := checkDockerBuildOptions(options)
			if err != nil {
				return err
			}
			cmdArgs = append(cmdArgs, options...)

		}
		cmdArgs = append(cmdArgs, "-f", dockerfile, extractDir)
		Debug.log("final cmd args", cmdArgs)
		execError := DockerBuild(cmdArgs, DockerLog)

		if execError != nil {
			return execError
		}
		if !dryrun {
			Info.log("Built docker image ", buildImage)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "Docker image name and optionally a tag in the 'name:tag' format")
	buildCmd.PersistentFlags().StringVar(&dockerBuildOptions, "docker-options", "", "Specify the docker build options to use.  Value must be in \"\".")

}
