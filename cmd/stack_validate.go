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
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// stack validate is a suite of validation tests for a local stack
// stack validate does the following...
// 1. stack lint test, can be turned off with --no-lint
// 2. stack package, can be turned off with --no-package
// 3. appsody init
// 4. appsody run
// 5. appsody test
// 6. appsody build

func newStackValidateCmd(rootConfig *RootCommandConfig) *cobra.Command {

	// vars for --no-package and --no-lint parms
	var noPackage bool
	var noLint bool
	var imageNamespace string

	var stackValidateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Run validation tests against a stack in the local Appsody environment",
		Long: `This command is a tool for stack developers to validate a stack from their local Appsody development environment. It performs the following against the stack:
- Runs the stack lint test. This can be turned off with the --no-lint flag
- Runs the stack package command. This can be turned off with the --no-package flag
- Runs the appsody init command
- Runs the appsody run command
- Runs the appsody test command
- Runs the appsody build command
- Provides a Passed/Failed status and summary of the above operations`,
		RunE: func(cmd *cobra.Command, args []string) error {

			// vars to store test results
			var testResults []string
			var initFail bool // if init fails we can skip the rest of the tests
			failCount := 0
			passCount := 0

			stackPath := rootConfig.ProjectDir
			Info.Log("stackPath is: ", stackPath)

			// check for templates dir, error out if its not there
			check, err := Exists("templates")
			if err != nil {
				return errors.New("Error checking stack root directory: " + err.Error())
			}
			if !check {
				// if we can't find the templates directory then we are not starting from a valid root of the stack directory
				return errors.New("Unable to reach templates directory. Current directory must be the root of the stack")
			}

			// get the stack name from the stack path
			stackName := filepath.Base(stackPath)
			Info.Log("stackName is: ", stackName)

			Info.Log("#################################################")
			Info.Log("Validating stack: ", stackName)
			Info.Log("#################################################")

			// create a temporary dir to create the project and run the test
			projectDir, err := ioutil.TempDir("", "appsody-build-simple-test")
			if err != nil {
				return err
			}

			Info.Log("Created project dir: " + projectDir)

			Debug.Log("Setting environment variable APPSODY_PULL_POLICY=IFNOTPRESENT")
			err = os.Setenv("APPSODY_PULL_POLICY", "IFNOTPRESENT")
			if err != nil {
				return errors.Errorf("Could not set environment variable APPSODY_PULL_POLICY. %v", err)
			}

			// call tests...

			// lint
			if !noLint {
				_, err = RunAppsodyCmdExec([]string{"stack", "lint"}, stackPath)
				if err != nil {
					//logs error but keeps going
					Error.Log(err)
					testResults = append(testResults, ("FAILED: Lint for stack: " + stackName))
					failCount++
				} else {
					testResults = append(testResults, ("PASSED: Lint for stack: " + stackName))
					passCount++
				}
			}

			// package
			if !noPackage {
				_, err = RunAppsodyCmdExec([]string{"stack", "package", "--image-namespace", imageNamespace}, stackPath)
				if err != nil {
					//logs error but keeps going
					Error.Log(err)
					testResults = append(testResults, ("FAILED: Package for stack: " + stackName))
					failCount++
				} else {
					testResults = append(testResults, ("PASSED: Package for stack: " + stackName))
					passCount++
				}
			}

			// init
			err = TestInit("dev.local/"+stackName, projectDir)
			if err != nil {
				Error.Log(err)
				testResults = append(testResults, ("FAILED: Init for stack: " + stackName))
				failCount++
				initFail = true
			} else {
				testResults = append(testResults, ("PASSED: Init for stack: " + stackName))
				passCount++
				initFail = false
			}

			// run
			if !initFail {
				err = TestRun(projectDir)
				if err != nil {
					//logs error but keeps going
					Error.Log(err)
					testResults = append(testResults, ("FAILED: Run for stack: " + stackName))
					failCount++
				} else {
					testResults = append(testResults, ("PASSED: Run for stack: " + stackName))
					passCount++
				}
			}

			// test
			if !initFail {
				err = TestTest(projectDir)
				if err != nil {
					//logs error but keeps going
					Error.Log(err)
					testResults = append(testResults, ("FAILED: Test for stack: " + stackName))
					failCount++
				} else {
					testResults = append(testResults, ("PASSED: Test for stack: " + stackName))
					passCount++
				}
			}

			// build
			if !initFail {
				err = TestBuild(stackName, projectDir)
				if err != nil {
					//logs error but keeps going
					Error.Log(err)
					testResults = append(testResults, ("FAILED: Build for stack: " + stackName))
					failCount++
				} else {
					testResults = append(testResults, ("PASSED: Build for stack: " + stackName))
					passCount++
				}
			}

			//cleanup
			Info.Log("Removing project dir: " + projectDir)
			os.RemoveAll(projectDir)

			//}

			Info.Log("@@@@@@@@@ Validate Summary Start @@@@@@@@@@")
			for i := range testResults {
				Info.Log(testResults[i])
			}
			Info.Log("Total PASSED: ", passCount)
			Info.Log("Total FAILED: ", failCount)
			Info.Log("@@@@@@@@@ Validate Summary End @@@@@@@@@@")

			return nil
		},
	}

	stackValidateCmd.PersistentFlags().BoolVar(&noPackage, "no-package", false, "Skips running appsody stack package")
	stackValidateCmd.PersistentFlags().BoolVar(&noLint, "no-lint", false, "Skips running appsody stack lint")
	stackValidateCmd.PersistentFlags().StringVar(&imageNamespace, "image-namespace", "dev.local", "Namespace that the images will be created using (default is dev.local)")

	return stackValidateCmd
}

// Simple test for appsody init command
func TestInit(stack string, projectDir string) error {

	Info.Log("******************************************")
	Info.Log("Running appsody init")
	Info.Log("******************************************")
	_, err := RunAppsodyCmdExec([]string{"init", stack}, projectDir)
	return err
}

// Simple test for appsody run command. A future enhancement would be to verify the image that gets built.
func TestRun(projectDir string) error {

	runChannel := make(chan error)
	containerName := "testRunContainer"
	go func() {
		Info.Log("******************************************")
		Info.Log("Running appsody run")
		Info.Log("******************************************")
		_, err := RunAppsodyCmdExec([]string{"run", "--name", containerName}, projectDir)
		runChannel <- err
	}()

	// check to see if we get an error from appsody run
	// log appsody ps output
	// if appsody run doesn't fail after the loop time then assume it passed
	// appsody ps will show a running container even if the app does not run successfully so it is not reliable
	// endpoint checking would be a better way to verify appsody run
	healthCheckFrequency := 2 // in seconds
	healthCheckTimeout := 60  // in seconds
	healthCheckWait := 0
	isHealthy := false
	for !(healthCheckWait >= healthCheckTimeout) {
		select {
		case err := <-runChannel:
			// appsody run exited, probably with an error
			Error.Log("Appsody run failed")
			return err
		case <-time.After(time.Duration(healthCheckFrequency) * time.Second):
			// see if appsody ps has a container
			healthCheckWait += healthCheckFrequency

			Info.Log("about to run appsody ps")
			stopOutput, errStop := RunAppsodyCmdExec([]string{"ps"}, projectDir)
			if !strings.Contains(stopOutput, "CONTAINER") {
				Info.Log("appsody ps output doesn't contain header line")
			}
			if !strings.Contains(stopOutput, containerName) {
				Info.Log("appsody ps output doesn't contain correct container name")
			} else {
				Info.Log("appsody ps contains correct container name")
				isHealthy = true
			}
			if errStop != nil {
				Error.Log(errStop)
				return errStop
			}
		}
	}

	if !isHealthy {
		Error.Log("appsody ps never found the correct container")
		return errors.New("appsody ps never found the correct container")
	}

	Info.Log("Appsody run did not fail")

	// stop and clean up after the run
	_, err := RunAppsodyCmdExec([]string{"stop", "--name", "testRunContainer"}, projectDir)
	if err != nil {
		Error.Log("appsody stop failed")
	}

	return nil
}

// Simple test for appsody build command. A future enhancement would be to verify the image that gets built.
func TestTest(projectDir string) error {

	Info.Log("******************************************")
	Info.Log("Running appsody test")
	Info.Log("******************************************")
	_, err := RunAppsodyCmdExec([]string{"test", "--no-watcher"}, projectDir)
	return err
}

// Simple test for appsody build command. A future enhancement would be to verify the image that gets built.
func TestBuild(stack string, projectDir string) error {

	imageName := "dev.local/" + filepath.Base(projectDir)

	Info.Log("******************************************")
	Info.Log("Running appsody build")
	Info.Log("******************************************")
	_, err := RunAppsodyCmdExec([]string{"build", "--tag", imageName}, projectDir)
	if err != nil {
		Error.Log(err)
		return err
	}

	// use docker image ls to check for the image
	fmt.Println("calling docker image ls to check for the image")
	imageBuilt := false
	dockerOutput, dockerErr := RunDockerCmdExec([]string{"image", "ls", imageName})
	if dockerErr != nil {
		Error.Log("Error running docker image ls "+imageName, dockerErr)
		return dockerErr

	}
	if strings.Contains(dockerOutput, imageName) {
		Info.Log("docker image " + imageName + " was found")
		imageBuilt = true
	}

	if !imageBuilt {
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
