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
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type deployCommandConfig struct {
	*RootCommandConfig
	appDeployFile, namespace, tag, pushURL, pullURL                             string
	knative, generate, force, push, nobuild, noOperatorCheck, noOperatorInstall bool
	knativeFlagPresent, namespaceFlagPresent                                    bool
	dockerBuildOptions                                                          string
	buildahBuildOptions                                                         string
}

func findNamespaceRepositoryAndTag(image string) string {
	nameSpaceRepositoryAndTag := firstAfter(image, "/")
	return nameSpaceRepositoryAndTag
}

func firstAfter(value string, a string) string {
	values := strings.Split(value, a)
	if len(values) == 3 {
		return values[1] + a + values[2]
	}
	return value
}

func newDeployCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &deployCommandConfig{RootCommandConfig: rootConfig}

	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy your Appsody project to Kubernetes.",
		Long: `Build and deploy a local container image of your Appsody project to your Kubernetes cluster. 
		
The command performs the following steps:

1. Runs the appsody build command to build the container image for deployment.
2. Generates a deployment manifest file, "app-deploy.yaml", if one is not present, then applies it to your Kubernetes cluster.
3. Deploys your image to your Kubernetes cluster via the Appsody operator, or as a Knative service if you specify the "--knative" flag. If an Appsody operator cannot be found, one will be installed on your cluster.

Run this command from the root directory of your Appsody project.`,
		Example: `  appsody deploy --namespace my-namespace
  Builds and deploys your project to the "my-namespace" namespace in your local Kubernetes cluster.
  
  appsody deploy -t my-repo/nodejs-express --push-url external-registry-url --pull-url internal-registry-url
  Builds and tags the image as "my-repo/nodejs-express", pushes the image to "external-registry-url/my-repo/nodejs-express", and creates a deployment manifest that tells the Kubernetes cluster to pull the image from "internal-registry-url/my-repo/nodejs-express".`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) > 0 {
				return errors.New("Unexpected argument. Use 'appsody [command] --help' for more information about a command")
			}
			projectDir, err := getProjectDir(config.RootCommandConfig)
			if err != nil {
				return err
			}
			config.knativeFlagPresent = cmd.Flag("knative").Changed
			config.namespaceFlagPresent = cmd.Flag("namespace").Changed

			dryrun := config.Dryrun
			namespace := config.namespace
			configFile := filepath.Join(projectDir, config.appDeployFile)

			exists, err := Exists(configFile)
			if err != nil {
				return err
			}

			if !exists && config.nobuild {
				return errors.Errorf("--no-build flag was used, but deployment manifest %s was not found. Please remove the --no-build, or generate a deployment manifest first by running \"appsody build\"", configFile)
			}

			if exists {
				config.Info.Logf("Found deployment manifest %s", configFile)

				deploymentManifest, err := getDeploymentManifest(configFile)
				if err != nil {
					return err
				}

				manifestNamespace := deploymentManifest.Namespace
				if manifestNamespace != "" {
					if namespace != "" && manifestNamespace != namespace {
						return errors.Errorf("the namespace \"%s\" from the deployment manifest does not match the namespace \"%s\" passed as an argument.", manifestNamespace, namespace)
					}

					namespace = manifestNamespace
				}
			}

			if namespace == "" {
				namespace = "default"
			}

			config.Info.Logf("Using namespace %s for deployment", namespace)

			if !config.nobuild {
				config.Info.Log("Building the production image")

				buildConfig := &buildCommandConfig{RootCommandConfig: config.RootCommandConfig}
				buildConfig.Verbose = config.Verbose
				buildConfig.pushURL = config.pushURL
				buildConfig.push = config.push
				buildConfig.dockerBuildOptions = config.dockerBuildOptions
				buildConfig.buildahBuildOptions = config.buildahBuildOptions

				buildConfig.tag = config.tag
				buildConfig.pullURL = config.pullURL
				buildConfig.knative = config.knative
				buildConfig.knativeFlagPresent = config.knativeFlagPresent
				buildConfig.appDeployFile = configFile
				buildConfig.namespace = namespace
				buildConfig.namespaceFlagPresent = config.namespaceFlagPresent

				buildErr := build(buildConfig)
				if buildErr != nil {
					return buildErr
				}
			}

			if config.generate {
				return nil
			}

			deploymentManifest, err := getDeploymentManifest(configFile)
			if err != nil {
				return err
			}

			if !config.noOperatorInstall && deploymentManifest.Kind == "AppsodyApplication" {
				// Check for the Appsody Operator
				operatorExists, existingNamespace, operatorExistsErr := operatorExistsWithWatchspace(config.LoggingConfig, namespace, config.Dryrun, config.noOperatorCheck)
				if operatorExistsErr != nil {
					return operatorExistsErr
				}

				//kargs := []string{"service/appsody-operator"}
				//_, err := KubeGet(kargs)
				// Performing the kubectl apply
				if !operatorExists {
					config.Debug.logf("Failed to find Appsody operator that watches namespace %s. Attempting to install...", namespace)
					operatorConfig := &operatorCommandConfig{config.RootCommandConfig, namespace}
					operatorInstallConfig := &operatorInstallCommandConfig{operatorCommandConfig: operatorConfig}
					//	operatorInstallConfig.RootCommandConfig = operatorConfig.RootCommandConfig
					err := operatorInstall(operatorInstallConfig)
					if err != nil {
						return errors.Errorf("Failed to install an Appsody operator in namespace %s watching namespace %s. Error was: %v", namespace, namespace, err)
					}
				} else {
					config.Debug.logf("Operator exists in %s, watching %s ", existingNamespace, namespace)
				}
			} else {
				config.Info.logf("The deployment manifest is of kind: %s, you need to install a matching operator.", deploymentManifest.Kind)
			}

			// Performing the kubectl apply
			err = KubeApply(config.LoggingConfig, configFile, namespace, dryrun)
			if err != nil {
				return errors.Errorf("Failed to deploy to your Kubernetes cluster: %v", err)
			}

			// Ensure hostname and IP config is set up for deployment
			time.Sleep(1 * time.Second)
			config.Info.log("Appsody Deployment name is: ", deploymentManifest.Name)
			out, err := KubeGetDeploymentURL(config.LoggingConfig, deploymentManifest.Name, namespace, dryrun)
			// Performing the kubectl apply
			if err != nil {
				return errors.Errorf("Failed to find deployed service IP and Port: %s", err)
			}
			if !dryrun {
				rootConfig.Info.log("Deployed project running at ", out)
			} else {
				rootConfig.Info.log("Dry run complete")
			}

			return nil
		},
	}

	addStackRegistryFlag(deployCmd, &config.RootCommandConfig.StackRegistry, config.RootCommandConfig)
	deployCmd.PersistentFlags().BoolVar(&config.generate, "generate-only", false, "DEPRECATED - Only generate the deployment manifest file. Do not deploy the project.")
	deployCmd.PersistentFlags().BoolVar(&config.nobuild, "no-build", false, "Deploys the application without building a new image or modifying the deployment manifest file.")
	deployCmd.PersistentFlags().StringVarP(&config.appDeployFile, "file", "f", "app-deploy.yaml", "The file name to use for the deployment manifest.")
	deployCmd.PersistentFlags().BoolVar(&config.force, "force", false, "DEPRECATED - Force the reuse of the deployment manifest file if one exists.")
	deployCmd.PersistentFlags().StringVarP(&config.namespace, "namespace", "n", "", "Target namespace in your Kubernetes cluster.")
	deployCmd.PersistentFlags().StringVarP(&config.tag, "tag", "t", "", "Docker image name and optionally a tag in the 'name:tag' format")
	deployCmd.PersistentFlags().BoolVar(&rootConfig.Buildah, "buildah", false, "Build project using buildah primitives instead of docker.")
	deployCmd.PersistentFlags().StringVar(&config.dockerBuildOptions, "docker-options", "", "Specify the docker build options to use. Value must be in \"\". The following Docker options are not supported: '--help','-t','--tag','-f','--file'.")
	deployCmd.PersistentFlags().StringVar(&config.buildahBuildOptions, "buildah-options", "", "Specify the buildah build options to use. Value must be in \"\".")
	deployCmd.PersistentFlags().BoolVar(&config.push, "push", false, "Push this image to an external Docker registry. Assumes that you have previously successfully done docker login")
	deployCmd.PersistentFlags().BoolVar(&config.knative, "knative", false, "Deploy as a Knative Service")
	deployCmd.PersistentFlags().StringVar(&config.pushURL, "push-url", "", "Remote repository to push image to.  This will also trigger a push if the --push flag is not specified.")
	deployCmd.PersistentFlags().StringVar(&config.pullURL, "pull-url", "", "Remote repository to pull image from.")
	deployCmd.PersistentFlags().BoolVar(&config.noOperatorCheck, "no-operator-check", false, "Do not check whether existing operators are already watching the namespace")
	deployCmd.PersistentFlags().BoolVar(&config.noOperatorInstall, "no-operator-install", false, "Deploy your application without installing the Appsody operator")
	deployCmd.AddCommand(newDeleteDeploymentCmd(config))

	return deployCmd
}
