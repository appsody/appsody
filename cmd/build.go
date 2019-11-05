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
	"os"
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

	projectName, perr := getProjectName(config.RootCommandConfig)
	if perr != nil {
		return perr
	}

	extractDir := filepath.Join(getHome(config.RootCommandConfig), "extract", projectName)
	dockerfile := filepath.Join(extractDir, "Dockerfile")
	buildImage := projectName //Lowercased

	// Regardless of pass or fail, remove the local extracted folder
	defer os.RemoveAll(extractDir)

	extractErr := extract(extractConfig)
	if extractErr != nil {
		return extractErr
	}

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

	gitLabels, err := getGitLabels(config)
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
		newKey, err := ConvertLabelToKubeFormat(key)
		if err != nil {
			Debug.logf("Skipping image label \"%s\" - %v", key, err)
		} else {
			kubeLabels[newKey] = value
		}
	}

	return kubeLabels
}

func ConvertLabelToKubeFormat(key string) (string, error) {
	// regular expression to strip off the domain prefix
	// this matches anything starting with an alphanumeric, followed by
	// alphanumerics or dots, and ending with a dot
	regex, err := regexp.Compile(`^[a-z0-9A-Z][a-z0-9A-Z.]*\.`)
	if err != nil {
		return "", err
	}
	loc := regex.FindStringIndex(key)
	var prefix string
	var name string
	if loc == nil {
		// did not start with a domain so there will be no prefix
		prefix = ""
		name = key
	} else {
		prefix = key[0:loc[1]]
		name = key[loc[1]:]
		// reverse the prefix domain
		domainSections := strings.Split(prefix, ".")
		newPrefix := ""
		for i := len(domainSections) - 1; i >= 0; i-- {
			if domainSections[i] != "" {
				newPrefix += domainSections[i]
				if i > 0 {
					newPrefix += "."
				}
			}
		}
		prefix = newPrefix + "/"
	}
	if name == "" {
		return "", errors.New("Invalid kubernetes metadata name. Must not be empty")
	}
	if len(prefix) > 253 {
		return "", errors.New("Invalid kubernetes metadata prefix. Must be less than 253 characters")
	}
	match, err := IsValidKubernetesLabelValue(name)
	if !match {
		return "", errors.Errorf("Invalid kubernetes metadata name. %v", err)
	}
	return prefix + name, nil
}

func createLabelPairs(labels map[string]string) []string {
	var labelsArr []string

	for key, value := range labels {
		labelString := fmt.Sprintf("%s=%s", key, value)
		labelsArr = append(labelsArr, labelString)
	}

	return labelsArr
}
