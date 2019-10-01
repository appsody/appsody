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
	"strings"
	"time"
)

// Simple test for appsody build command. A future enhancement would be to verify the image that gets built.
func TestBuild(projectDir string) error {

	// appsody build

	imageName := "testbuildimage"

	Info.Log("******************************************")
	Info.Log("Running appsody build")
	Info.Log("******************************************")
	_, err := RunAppsodyCmdExec([]string{"build", "--tag", imageName}, projectDir)
	if err != nil {
		Error.Log(err)
		return err
	}

	// It will take a while for the image to build, so lets use docker image ls to wait for it
	fmt.Println("calling docker image ls to wait for the image")
	imageBuilt := false
	count := 900
	for {
		dockerOutput, dockerErr := RunDockerCmdExec([]string{"image", "ls", imageName})
		if dockerErr != nil {
			Error.Log("Ignoring error running docker image ls "+imageName, dockerErr)
			return dockerErr

		}
		if strings.Contains(dockerOutput, imageName) {
			Info.Log("docker image " + imageName + " was found")
			imageBuilt = true
		} else {
			time.Sleep(2 * time.Second)
			count = count - 1
		}
		if count == 0 || imageBuilt {
			break
		}
	}

	if !imageBuilt {
		// *** how to fail the test?
		Error.Log("image was never built")
		return err
	}

	//delete the image
	_, err = RunDockerCmdExec([]string{"image", "rm", imageName})
	if err != nil {
		Error.Log(err)
		return err
	}

	return nil
}
