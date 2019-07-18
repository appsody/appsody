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
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var namespace, tag string
var push bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Build and deploy your Appsody project on a local Kubernetes cluster",
	Long: `This command extracts the code from your project, builds a local Docker image for deployment,
generates a KNative serving deployment manifest (yaml) file, and deploys your image as a KNative
service in your local cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Extract code and build the image - and tags it if -t is specified
		buildCmd.Run(cmd, args)
		//Generate the KNative yaml
		//Get the container port first
		port, err := getEnvVarInt("PORT")
		if err != nil {
			//try and get the exposed ports and use the first one
			Warning.log("Could not detect a container port (PORT env var).")
			portsStr := getExposedPorts()
			if len(portsStr) == 0 {
				//No ports exposed
				Warning.log("This container exposes no ports. The service will not be accessible.")
				port = 0 //setting this to 0
			} else {
				portStr := portsStr[0]
				Warning.log("Picking the first exposed port as the KNative service port. This may not be the correct port.")
				port, err = strconv.Atoi(portStr)
				if err != nil {
					Warning.log("The exposed port is not a valid integer. The service will not be accessible.")
					port = 0
				}
			}
		}
		//Get the KNative template file
		knativeTempl := getKNativeTemplate()
		//Get the project name and make it the KNative service name
		serviceName := getProjectName()
		//Retrieve the project name and lowercase it
		projectName := getProjectName()
		deployImage := projectName // if not tagged, this is the deploy image name
		if tag != "" {
			deployImage = tag //Otherwise, it's the tag
		}
		// We're not pushing to a repository, so we need to use dev.local for Knative to be able to find it
		if !push {
			localtag := "dev.local/" + projectName
			// Tagging the image using the tag as the deployImage for KNative
			err = DockerTag(deployImage, localtag)
			if err != nil {
				Error.log("Tagging the image failed - exiting. Error: ", err)
				os.Exit(1)
			}
			deployImage = localtag // And forcing deployimage to be localtag
		}
		//Generating the KNative yaml file
		Debug.logf("Calling GenKnativeYaml with parms: %s %d %s %s \n", knativeTempl, port, serviceName, deployImage)
		yamlFileName, err := GenKnativeYaml(knativeTempl, port, serviceName, deployImage, push)
		if err != nil {
			Error.log("Could not generate the KNative YAML file: ", err)
			os.Exit(1)
		}
		Info.log("Generated KNative serving deploy file: ", yamlFileName)
		// Pushing the docker image if necessary
		if push {
			err = DockerPush(deployImage)
			if err != nil {
				Error.log("Could not push the docker image - exiting. Error: ", err)
			}
		}
		err = KubeApply(yamlFileName)
		// Performing the kubectl apply
		if err != nil {
			Error.log("Failed to deploy to your Kubernetes cluster: ", err)
		} else {
			Info.log("Deployment succeeded.")
			url, err := KubeGetRouteURL(serviceName)
			if err != nil {
				Error.log("Failed to find deployed service in your Kubernetes cluster: ", err)
			} else {
				Info.log("Your deployed service is available at the following URL: ", url)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Target namespace in your Kubernetes cluster")
	deployCmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "Docker image name and optionally a tag in the 'name:tag' format")
	deployCmd.PersistentFlags().BoolVar(&push, "push", false, "Push this image to an external Docker registry. Assumes that you have previously successfully done docker login")

}
