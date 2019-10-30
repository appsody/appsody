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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type buildCommandConfig struct {
	*RootCommandConfig
	tag                string
	dockerBuildOptions string
}

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

func newBuildCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &buildCommandConfig{RootCommandConfig: rootConfig}
	// buildCmd provides the ability run local builds, or setup/delete Tekton builds, for an appsody project
	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Locally build a docker image of your appsody project",
		Long:  `This allows you to build a local Docker image from your Appsody project. Extract is run before the docker build.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return build(config)
		},
	}

	buildCmd.PersistentFlags().StringVarP(&config.tag, "tag", "t", "", "Docker image name and optionally a tag in the 'name:tag' format")
	buildCmd.PersistentFlags().StringVar(&config.dockerBuildOptions, "docker-options", "", "Specify the docker build options to use.  Value must be in \"\".")

	buildCmd.AddCommand(newBuildDeleteCmd(config))
	buildCmd.AddCommand(newSetupCmd(config))
	return buildCmd
}

func build(config *buildCommandConfig) error {
	// This needs to do:
	// 1. appsody Extract
	// 2. docker build -t <project name> -f Dockerfile ./extracted

	extractConfig := &extractCommandConfig{RootCommandConfig: config.RootCommandConfig}
	extractErr := extract(extractConfig)
	if extractErr != nil {
		return extractErr
	}

	projectName, perr := getProjectName(config.RootCommandConfig)
	if perr != nil {
		return perr
	}

	extractDir := filepath.Join(getHome(config.RootCommandConfig), "extract", projectName)
	dockerfile := filepath.Join(extractDir, "Dockerfile")
	buildImage := projectName //Lowercased

	// If a tag is specified, change the buildImage
	if config.tag != "" {
		buildImage = config.tag
	}
	cmdArgs := []string{"-t", buildImage}

	if config.dockerBuildOptions != "" {
		dockerBuildOptions := strings.TrimPrefix(config.dockerBuildOptions, " ")
		dockerBuildOptions = strings.TrimSuffix(dockerBuildOptions, " ")
		options := strings.Split(dockerBuildOptions, " ")
		err := checkDockerBuildOptions(options)
		if err != nil {
			return err
		}
		cmdArgs = append(cmdArgs, options...)
	}

	labels, err := getLabels(config)
	if err != nil {
		return err
	}

	// It would be nicer to only call the --label flag once. Could also use the --label-file flag.
	for _, label := range labels {
		cmdArgs = append(cmdArgs, "--label", label)
	}

	cmdArgs = append(cmdArgs, "-f", dockerfile, extractDir)
	Debug.log("final cmd args", cmdArgs)
	execError := DockerBuild(cmdArgs, DockerLog, config.Verbose, config.Dryrun)

	if execError != nil {
		return execError
	}
	if !config.Dryrun {
		Info.log("Built docker image ", buildImage)
	}
	return nil
}

func getLabels(config *buildCommandConfig) ([]string, error) {
	var labels []string

	stackLabels, err := getStackLabels(config.RootCommandConfig)
	if err != nil {
		return labels, err
	}

	projectConfig, projectConfigErr := getProjectConfig(config.RootCommandConfig)
	if projectConfigErr != nil {
		return labels, projectConfigErr
	}

	configLabels, err := getConfigLabels(*projectConfig)
	if err != nil {
		return labels, err
	}

	gitLabels, err := getGitLabels(config.RootCommandConfig)
	if err != nil {
		Info.log(err)
	}

	for key, value := range stackLabels {

		key = strings.Replace(key, "org.opencontainers.image", "dev.appsody.stack", -1)

		// This is temporarily until we update the labels in stack dockerfile
		if key == "appsody.stack" {
			key = "dev.appsody.stack.id"
		}

		delete(configLabels, key)

		labelString := fmt.Sprintf("%s=%s", key, value)
		labels = append(labels, labelString)
	}

	for key, value := range configLabels {
		labelString := fmt.Sprintf("%s=%s", key, value)
		labels = append(labels, labelString)
	}

	for key, value := range gitLabels {
		labelString := fmt.Sprintf("%s=%s", key, value)
		labels = append(labels, labelString)
	}

	return labels, nil
}
