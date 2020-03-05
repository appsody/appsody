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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"bytes"
	"strconv"
)

type buildCommandConfig struct {
	*RootCommandConfig
	tag                  string
	dockerBuildOptions   string
	buildahBuildOptions  string
	pushURL              string
	push                 bool
	pullURL              string
	appDeployFile        string
	knative              bool
	knativeFlagPresent   bool
	namespaceFlagPresent bool
	namespace            string
}

type DeploymentManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   map[string]interface{} `json:"spec,omitempty"`
	Status interface{}            `json:"status,omitempty"`
}

//These are the current supported labels for Kubernetes,
//the rest of the labels provided will be annotations.
var supportedKubeLabels = []string{
	"image.opencontainers.org/title",
	"image.opencontainers.org/version",
	"image.opencontainers.org/licenses",
	"stack.appsody.dev/id",
	"stack.appsody.dev/version",
	"app.kubernetes.io/part-of",
}

func checkBuildOptions(options []string) error {
	buildOptionsTest := "(^((-t)|(--tag)|(--help)|(-f)|(--file))((=?$)|(=.*)))"

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
		Short: "Build a local container image of your Appsody project.",
		Long: `Build a local container image of your Appsody project. The stack, along with your Appsody project, is extracted to a local directory before the container build is run.

By default, the built image is tagged with the project name that you specified when you initialised your Appsody project. If you did not specify a name, the image is tagged with the name of the root directory of your Appsody project.

If you want to push the built image to an image repository using the [--push] options, you must specify the relevant image tag.

Run this command from the root directory of your Appsody project.`,
		Example: `  appsody build -t my-repo/nodejs-express --push
  Builds the container image, tags it with my-repo/nodejs-express, and pushes it to the container registry the Docker CLI is currently logged into.

  appsody build -t my-repo/nodejs-express:0.1 --push-url my-registry-url
  Builds the container image, tags it with my-repo/nodejs-express, and pushes it to my-registry-url/my-repo/nodejs-express:0.1.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("Unexpected argument. Use 'appsody [command] --help' for more information about a command")
			}
			config.knativeFlagPresent = cmd.Flag("knative").Changed

			projectDir, err := getProjectDir(config.RootCommandConfig)
			if err != nil {
				return err
			}
			config.appDeployFile = filepath.Join(projectDir, config.appDeployFile)

			return build(config)
		},
	}
	addStackRegistryFlag(buildCmd, &rootConfig.StackRegistry, rootConfig)

	buildCmd.PersistentFlags().StringVarP(&config.tag, "tag", "t", "", "Container image name and optionally, a tag in the 'name:tag' format.")
	buildCmd.PersistentFlags().BoolVar(&rootConfig.Buildah, "buildah", false, "Build project using buildah primitives instead of Docker.")
	buildCmd.PersistentFlags().StringVar(&config.dockerBuildOptions, "docker-options", "", "Specify the Docker build options to use. Value must be in \"\". The following Docker options are not supported: '--help','-t','--tag','-f','--file'.")
	buildCmd.PersistentFlags().StringVar(&config.buildahBuildOptions, "buildah-options", "", "Specify the buildah build options to use. Value must be in \"\".")
	buildCmd.PersistentFlags().BoolVar(&config.push, "push", false, "Push the container image to the image repository.")
	buildCmd.PersistentFlags().StringVar(&config.pushURL, "push-url", "", "The remote registry to push the image to. This will also trigger a push if the --push flag is not specified.")
	buildCmd.PersistentFlags().StringVar(&config.pullURL, "pull-url", "", "Remote repository to pull image from.")
	buildCmd.PersistentFlags().BoolVar(&config.knative, "knative", false, "Deploy as a Knative Service")
	buildCmd.PersistentFlags().StringVarP(&config.appDeployFile, "file", "f", "app-deploy.yaml", "The file name to use for the deployment configuration.")

	buildCmd.AddCommand(newBuildDeleteCmd(config))
	buildCmd.AddCommand(newSetupCmd(config))
	return buildCmd
}

func build(config *buildCommandConfig) error {
	// This needs to do:
	// 1. appsody Extract
	// 2. docker build -t <project name> -f Dockerfile ./extracted
	buildOptions := ""
	if config.dockerBuildOptions != "" {
		if config.Buildah {
			return errors.New("Cannot specify --docker-options flag with --buildah")
		}
		buildOptions = strings.TrimSpace(config.dockerBuildOptions)
	}
	if config.buildahBuildOptions != "" {
		if !config.Buildah {
			return errors.New("Cannot specify --buildah-options flag without --buildah")
		}
		buildOptions = strings.TrimSpace(config.buildahBuildOptions)
	}

	// Issue 529 - if you specify --push or --push-url without --tag, error out
	if config.tag == "" && (config.push || config.pushURL != "") {
		return errors.New("Cannot specify --push or --push-url without a --tag")
	}

	extractConfig := &extractCommandConfig{RootCommandConfig: config.RootCommandConfig}

	projectName, perr := getProjectName(config.RootCommandConfig)
	if perr != nil {
		return perr
	}

	extractDir := filepath.Join(getHome(config.RootCommandConfig), "extract", projectName)
	dockerfile := filepath.Join(extractDir, "Dockerfile")
	buildImage := "dev.local/" + projectName //Lowercased

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

	if buildOptions != "" {
		options := strings.Split(buildOptions, " ")
		err := checkBuildOptions(options)
		if err != nil {
			return err
		}
		cmdArgs = append(cmdArgs, options...)
	}

	labels, err := getLabels(config.RootCommandConfig)
	if err != nil {
		return err
	}

	labelPairs := CreateLabelPairs(labels)

	// It would be nicer to only call the --label flag once. Could also use the --label-file flag.
	for _, label := range labelPairs {
		cmdArgs = append(cmdArgs, "--label", label)
	}

	cmdArgs = append(cmdArgs, "-f", dockerfile, extractDir)
	config.Debug.log("final cmd args", cmdArgs)
	var execError error
	if !config.Buildah {
		execError = DockerBuild(config.RootCommandConfig, cmdArgs, config.DockerLog)
	} else {
		execError = BuildahBuild(config.RootCommandConfig, cmdArgs, config.BuildahLog)
	}

	if execError != nil {
		return execError
	}
	if config.pushURL != "" || config.push {
		err := ImagePush(config.LoggingConfig, buildImage, config.Buildah, config.Dryrun)
		if err != nil {
			return errors.Errorf("Could not push the docker image - exiting. Error: %v", err)
		}
	}
	if !config.Dryrun {
		config.Info.log("Built docker image ", buildImage)
	}

	// Generate app-deploy
	err = generateDeploymentConfig(config, buildImage, labels)
	if err != nil {
		return err
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

	configLabels, err := getConfigLabels(*projectConfig, ".appsody-config.yaml", config.LoggingConfig)
	if err != nil {
		return labels, err
	}

	gitLabels, err := getGitLabels(config)
	if err != nil {
		config.Warning.log("Not all labels will be set. ", err.Error())
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

func convertLabelsToKubeFormat(log *LoggingConfig, labels map[string]string) map[string]string {
	var kubeLabels = make(map[string]string)

	for key, value := range labels {
		newKey, err := ConvertLabelToKubeFormat(key)
		if newKey == "app.appsody.dev/name" {
			newKey = "app.kubernetes.io/part-of"
		}
		if err != nil {
			log.Debug.logf("Skipping image label \"%s\" - %v", key, err)
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

func CreateLabelPairs(labels map[string]string) []string {
	var labelsArr []string

	for key, value := range labels {
		labelString := fmt.Sprintf("%s=%s", key, value)
		labelsArr = append(labelsArr, labelString)
	}

	return labelsArr
}

func generateDeploymentConfig(config *buildCommandConfig, imageName string, labels map[string]string) error {
	containerConfigDir := "/config/app-deploy.yaml"
	configFile := config.appDeployFile

	projectConfig, configErr := getProjectConfig(config.RootCommandConfig)
	if configErr != nil {
		return configErr
	}

	err := CheckPrereqs(config.RootCommandConfig)
	if err != nil {
		config.Warning.logf("Failed to check prerequisites: %v\n", err)
	}

	stackImage := projectConfig.Stack
	config.Debug.log("Stack image: ", stackImage)
	config.Debug.log("Config directory: ", containerConfigDir)

	exists, err := Exists(configFile)
	if err != nil {
		return errors.Errorf("Error checking status of %s", configFile)
	}

	if exists {
		config.Info.log("Found existing deployment manifest ", configFile)
		err := updateDeploymentConfig(config, imageName, labels)
		if err != nil {
			return err
		}
		config.Info.log("Updated existing deployment manifest ", configFile)
		return nil
	}

	var cmdName string
	var cmdArgs []string
	pullErr := pullImage(stackImage, config.RootCommandConfig)
	if pullErr != nil {
		return pullErr
	}
	extractContainerName := defaultExtractContainerName(config.RootCommandConfig)

	cmdName = "docker"
	if config.Buildah {
		cmdName = "buildah"
	}

	var configDir string
	cmdArgs = []string{"--name", extractContainerName}

	if config.Buildah {
		cmdArgs = append([]string{"from"}, cmdArgs...)
	} else {
		cmdArgs = append([]string{"create"}, cmdArgs...)
	}
	cmdArgs = append(cmdArgs, stackImage)
	err = execAndWaitReturnErr(config.LoggingConfig, cmdName, cmdArgs, config.Debug, config.Dryrun)
	if err != nil {
		config.Error.log("Container create command failed: ", err)

		// TODO: We shouldn't remove the container if it already exists
		removeErr := containerRemove(config.LoggingConfig, extractContainerName, config.Buildah, config.Dryrun)
		if removeErr != nil {
			config.Error.log("Error in containerRemove", removeErr)
		}
		return err
	}
	configDir = extractContainerName + ":" + containerConfigDir

	cmdArgs = []string{"cp", configDir, configFile}
	if config.Buildah {
		// buildah does not support copying from the container to the filesystem
		// we'll need to convert this to a mount, like we do in extract
		//cmdArgs = []string{"copy", configDir, configFile}
		configDir = containerConfigDir
		cmdName = "/bin/sh"
		script := fmt.Sprintf("x=`buildah mount %s`; cp -f $x/%s %s", extractContainerName, configDir, configFile)
		cmdArgs = []string{"-c", script}
	}
	err = execAndWaitReturnErr(config.LoggingConfig, cmdName, cmdArgs, config.Debug, config.Dryrun)

	removeErr := containerRemove(config.LoggingConfig, extractContainerName, config.Buildah, config.Dryrun)
	if removeErr != nil {
		config.Error.log("containerRemove error ", removeErr)
	}

	if err != nil {
		return errors.Errorf("Container copy command failed: %v", err)
	}

	yamlReader, err := ioutil.ReadFile(configFile)

	if !config.Dryrun && err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("Config file does not exist %s. ", configFile)

		}
		return errors.Errorf("Failed reading file %s", configFile)

	}

	projectName, perr := getProjectName(config.RootCommandConfig)
	if perr != nil {
		return errors.Errorf("%v", perr)
	}

	port, err := getEnvVarInt("PORT", config.RootCommandConfig)
	if err != nil {
		//try and get the exposed ports and use the first one
		config.Warning.log("Could not detect a container port (PORT env var).")
		portsStr, portsErr := getExposedPorts(config.RootCommandConfig)
		if portsErr != nil {
			return portsErr
		}
		if len(portsStr) == 0 {
			//No ports exposed
			config.Warning.log("This container exposes no ports. The service will not be accessible.")
			port = 0 //setting this to 0
		} else {
			portStr := portsStr[0]
			config.Warning.log("Picking the first exposed port as the KNative service port. This may not be the correct port.")
			port, err = strconv.Atoi(portStr)
			if err != nil {
				config.Warning.log("The exposed port is not a valid integer. The service will not be accessible.")
				port = 0
			}
		}
	}
	portStr := strconv.Itoa(port)

	split := strings.Split(stackImage, ":")
	stack := split[len(split)-2]
	split = strings.Split(stack, "/")
	stack = split[len(split)-1]

	if !config.Dryrun {
		output := bytes.Replace(yamlReader, []byte("APPSODY_PROJECT_NAME"), []byte(projectName), -1)
		output = bytes.Replace(output, []byte("APPSODY_DOCKER_IMAGE"), []byte(imageName), -1)
		output = bytes.Replace(output, []byte("APPSODY_STACK"), []byte(stack), -1)
		output = bytes.Replace(output, []byte("APPSODY_PORT"), []byte(portStr), -1)

		err = ioutil.WriteFile(configFile, output, 0666)
		if err != nil {
			return errors.Errorf("Failed to write local application configuration file: %s", err)
		}

		err = updateDeploymentConfig(config, imageName, labels)

		if err != nil {
			return errors.Errorf("Failed to update deployment config file: %s", err)
		}
	} else {
		config.Info.logf("Dry run skipped construction of file %s", configFile)
	}
	config.Info.log("Created deployment manifest: ", configFile)
	return nil
}

func updateDeploymentConfig(config *buildCommandConfig, imageName string, labels map[string]string) error {
	configFile := config.appDeployFile

	deploymentManifest, err := getDeploymentManifest(configFile)
	if err != nil {
		return err
	}

	labels = convertLabelsToKubeFormat(config.LoggingConfig, labels)

	var selectedLabels = make(map[string]string)
	for _, label := range supportedKubeLabels {
		if labels[label] != "" {
			selectedLabels[label] = labels[label]
			delete(labels, label)
		}
	}

	if deploymentManifest.Labels == nil {
		deploymentManifest.Labels = selectedLabels
	} else {
		for key, value := range selectedLabels {
			deploymentManifest.Labels[key] = value
		}
	}

	if deploymentManifest.Annotations == nil {
		deploymentManifest.Annotations = labels
	} else {
		for key, value := range labels {
			deploymentManifest.Annotations[key] = value
		}
	}

	if deploymentManifest.Spec == nil {
		deploymentManifest.Spec = make(map[string]interface{})
	}

	if deploymentManifest.Spec["createKnativeService"] == nil || config.knativeFlagPresent {
		deploymentManifest.Spec["createKnativeService"] = config.knative
	}

	if config.pullURL != "" {
		imageName = config.pullURL + "/" + findNamespaceRepositoryAndTag(imageName)
	}

	deploymentManifest.Spec["applicationImage"] = imageName

	// This only applies to the deploy command flow:
	// - if the namespace doesn't exist in the manifest, and a namespace flag is not set: we write a "default" namespace
	// - if the namespace does exist in the manifest, and a namespace flag is set: we verify that they are the same. If they are not we throw an error.
	// - if the namespace doesn't exist in the manifest, and a namespace flag is set: we write the value passed an argument with the flag.
	if deploymentManifest.Namespace == "" && config.namespaceFlagPresent {
		deploymentManifest.Namespace = config.namespace
	}

	err = writeDeploymentManifest(deploymentManifest, config)
	if err != nil {
		return err
	}

	return nil
}

func getDeploymentManifest(configFile string) (DeploymentManifest, error) {
	var deploymentManifest DeploymentManifest
	yamlFileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return deploymentManifest, errors.Errorf("Could not read %s file: %s", configFile, err)
	}

	err = yaml.Unmarshal(yamlFileBytes, &deploymentManifest)
	if err != nil {
		return deploymentManifest, errors.Errorf("%s formatting error: %s", configFile, err)
	}

	return deploymentManifest, err
}

func writeDeploymentManifest(deploymentManifest DeploymentManifest, config *buildCommandConfig) error {
	configFile := config.appDeployFile

	output, err := yaml.Marshal(deploymentManifest)
	if err != nil {
		return errors.Errorf("Could not marshall deployment manifest to YAML when updating the %s: %s", configFile, err)
	}

	err = ioutil.WriteFile(configFile, output, 0666)
	if err != nil {
		return errors.Errorf("Failed to write local deployment manifest configuration file: %s", err)
	}

	return nil
}
