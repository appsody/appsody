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
	pushURL            string
	push               bool
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
	buildCmd.PersistentFlags().BoolVar(&config.push, "push", false, "Push the Docker image to the image repository.")
	buildCmd.PersistentFlags().StringVar(&config.pushURL, "push-url", "", "The remote registry to push the image to.")
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
	if config.pushURL != "" {
		buildImage = config.pushURL + "/" + buildImage
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

	labels, err := getLabels(config.RootCommandConfig)
	if err != nil {
		return err
	}

	labelPairs := createLabelPairs(labels)

	// It would be nicer to only call the --label flag once. Could also use the --label-file flag.
	for _, label := range labelPairs {
		cmdArgs = append(cmdArgs, "--label", label)
	}

	cmdArgs = append(cmdArgs, "-f", dockerfile, extractDir)
	Debug.log("final cmd args", cmdArgs)
	execError := DockerBuild(cmdArgs, DockerLog, config.Verbose, config.Dryrun)

	if execError != nil {
		return execError
	}
	if config.push {

		err := DockerPush(buildImage, config.Dryrun)
		if err != nil {
			return errors.Errorf("Could not push the docker image - exiting. Error: %v", err)
		}
	}
	if !config.Dryrun {
		Info.log("Built docker image ", buildImage)
	}

	Info.log(convertLabelsToKubeFormat(labels))
	return nil
}

func getLabels(config *RootCommandConfig) (map[string]string, error) {
	var labels = make(map[string]string)

	stackLabels, err := getStackLabels(config)
	if err != nil {
		return labels, err
	}

	projectConfig, projectConfigErr := getProjectConfig(config)
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
		key = strings.Replace(key, ociKeyPrefix, appsodyStackKeyPrefix, 1)
		key = strings.Replace(key, appsodyImageCommitKeyPrefix, appsodyStackKeyPrefix+"commit.", 1)

		// This is temporarily until we update the labels in stack dockerfile
		if key == "appsody.stack" {
			key = "dev.appsody.stack.tag"
		}

		delete(configLabels, key)

		labels[key] = value
	}

	for key, value := range configLabels {
		labels[key] = value
	}

	for key, value := range gitLabels {
		labels[key] = value
	}

	return labels, nil
}

func convertLabelsToKubeFormat(labels map[string]string) map[string]string {
	var kubeLabels = make(map[string]string)

	for key, value := range labels {
		prefixes := strings.Split(key, ".")
		nPrefixes := len(prefixes)
		newKey := ""
		for i := nPrefixes - 2; i >= 0; i-- {
			newKey += prefixes[i]
			if i > 0 {
				newKey += "."
			}
		}

		newKey += "/" + prefixes[nPrefixes-1]
		kubeLabels[newKey] = value
	}

	return kubeLabels
}

func createLabelPairs(labels map[string]string) []string {
	var labelsArr []string

	for key, value := range labels {
		labelString := fmt.Sprintf("%s=%s", key, value)
		labelsArr = append(labelsArr, labelString)
	}

	return labelsArr
}
