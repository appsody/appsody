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
	appDeployFile, namespace, tag, pushURL, pullURL string
	knative, generate, force, push                  bool
	dockerBuildOptions                              string
	buildahBuildOptions                             string
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
		
The stack, along with your Appsody project, is extracted to a local directory before the container image of your Appsody project is built for deployment.
The command generates a deployment manifest file, "app-deploy.yaml", if one is not present. The CLI uses this file to deploy your image to your Kubernetes cluster via the Appsody operator, and optionally as a Knative service if you specify the "--knative" flag.`,
		Example: `  appsody deploy --namespace my-namespace
  Builds and deploys your project to the "my-namespace" namespace in your local Kubernetes cluster.
  
  appsody deploy -t my-repo/nodejs-express --push-url external-registry-url --pull-url internal-registry-url
  Builds the container image, sets the namespace to "my-repo", sets the name to "nodejs-express", pushes image to "external-registry-url/my-repo/nodejs-express", and creates a deployment manifest that tells the K8S cluster to pull the image from "internal-registry-url/my-repo/nodejs-express".`,
		RunE: func(cmd *cobra.Command, args []string) error {

			projectDir, err := getProjectDir(config.RootCommandConfig)
			if err != nil {
				return err
			}

			dryrun := config.Dryrun
			namespace := config.namespace
			configFile := filepath.Join(projectDir, config.appDeployFile)

			buildConfig := &buildCommandConfig{RootCommandConfig: config.RootCommandConfig}
			buildConfig.Verbose = config.Verbose
			buildConfig.pushURL = config.pushURL
			buildConfig.push = config.push
			buildConfig.dockerBuildOptions = config.dockerBuildOptions
			buildConfig.buildahBuildOptions = config.buildahBuildOptions

			buildConfig.tag = config.tag
			buildConfig.pullURL = config.pullURL
			buildConfig.knative = config.knative
			buildConfig.appDeployFile = configFile

			if config.generate {
				return generateDeploymentConfig(buildConfig)
			}

			buildErr := build(buildConfig)
			if buildErr != nil {
				return buildErr
			}

			// Check for the Appsody Operator
			operatorExists, existingNamespace, operatorExistsErr := operatorExistsWithWatchspace(namespace, config.Dryrun)
			if operatorExistsErr != nil {
				return operatorExistsErr
			}

			//kargs := []string{"service/appsody-operator"}
			//_, err := KubeGet(kargs)
			// Performing the kubectl apply
			if !operatorExists {
				Debug.logf("Failed to find Appsody operator that watches namespace %s. Attempting to install...", namespace)
				operatorConfig := &operatorCommandConfig{config.RootCommandConfig, namespace}
				operatorInstallConfig := &operatorInstallCommandConfig{operatorCommandConfig: operatorConfig}
				//	operatorInstallConfig.RootCommandConfig = operatorConfig.RootCommandConfig
				err := operatorInstall(operatorInstallConfig)
				if err != nil {
					return errors.Errorf("Failed to install an Appsody operator in namespace %s watching namespace %s. Error was: %v", namespace, namespace, err)
				}
			} else {
				Debug.logf("Operator exists in %s, watching %s ", existingNamespace, namespace)

			}

			// Performing the kubectl apply
			err = KubeApply(configFile, namespace, dryrun)
			if err != nil {
				return errors.Errorf("Failed to deploy to your Kubernetes cluster: %v", err)
			}

			appsodyApplication, err := getAppsodyApplication(configFile)
			if err != nil {
				return err
			}

			// Ensure hostname and IP config is set up for deployment
			time.Sleep(1 * time.Second)
			Info.log("Appsody Deployment name is: ", appsodyApplication.Name)
			out, err := KubeGetDeploymentURL(appsodyApplication.Name, namespace, dryrun)
			// Performing the kubectl apply
			if err != nil {
				return errors.Errorf("Failed to find deployed service IP and Port: %s", err)
			}
			if !dryrun {
				Info.log("Deployed project running at ", out)
			} else {
				Info.log("Dry run complete")
			}

			return nil
		},
	}

	deployCmd.PersistentFlags().BoolVar(&config.generate, "generate-only", false, "DEPRECATED - Only generate the deployment configuration file. Do not deploy the project.")
	deployCmd.PersistentFlags().StringVarP(&config.appDeployFile, "file", "f", "app-deploy.yaml", "The file name to use for the deployment configuration.")
	deployCmd.PersistentFlags().BoolVar(&config.force, "force", false, "DEPRECATED - Force the reuse of the deployment configuration file if one exists.")
	deployCmd.PersistentFlags().StringVarP(&config.namespace, "namespace", "n", "default", "Target namespace in your Kubernetes cluster")
	deployCmd.PersistentFlags().StringVarP(&config.tag, "tag", "t", "", "Docker image name and optionally a tag in the 'name:tag' format")
	deployCmd.PersistentFlags().BoolVar(&rootConfig.Buildah, "buildah", false, "Build project using buildah primitives instead of docker.")
	deployCmd.PersistentFlags().StringVar(&config.dockerBuildOptions, "docker-options", "", "Specify the docker build options to use. Value must be in \"\".")
	deployCmd.PersistentFlags().StringVar(&config.buildahBuildOptions, "buildah-options", "", "Specify the buildah build options to use. Value must be in \"\".")
	deployCmd.PersistentFlags().BoolVar(&config.push, "push", false, "Push this image to an external Docker registry. Assumes that you have previously successfully done docker login")
	deployCmd.PersistentFlags().BoolVar(&config.knative, "knative", false, "Deploy as a Knative Service")
	deployCmd.PersistentFlags().StringVar(&config.pushURL, "push-url", "", "Remote repository to push image to.  This will also trigger a push if the --push flag is not specified.")
	deployCmd.PersistentFlags().StringVar(&config.pullURL, "pull-url", "", "Remote repository to pull image from.")
	deployCmd.AddCommand(newDeleteDeploymentCmd(config))

	return deployCmd
}
