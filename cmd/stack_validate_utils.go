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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// RunAppsodyCmdExec runs the appsody CLI with the given args in a new process
// The stdout and stderr are captured, printed, and returned
// args will be passed to the appsody command
// workingDir will be the directory the command runs in
func RunAppsodyCmdExec(args []string, workingDir string, rootConfig *RootCommandConfig) (string, error) {

	rootConfig.ProjectDir = workingDir

	// // Buffer cmd output, to be logged if there is a failure
	var outBuffer bytes.Buffer
	// Direct cmd console output to a buffer
	outReader, outWriter := io.Pipe()

	// copy the output to the buffer, and also to the test log
	outScanner := bufio.NewScanner(outReader)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for outScanner.Scan() {
			out := outScanner.Bytes()
			outBuffer.Write(out)
			outBuffer.WriteByte('\n')
		}
		wg.Done()
	}()

	rootConfig.Info.Log("Running appsody in the test sandbox with args: ", args)
	err := ExecuteE("vlatest", "latest", rootConfig.ProjectDir, outWriter, outWriter, args)
	if err != nil {
		rootConfig.Error.Log("Error returned from appsody command: ", err)
	}

	// close the writer first, so it sends an EOF to the scanner above,
	// then wait for the scanner to finish before closing the reader
	outWriter.Close()
	wg.Wait()
	outReader.Close()

	return outBuffer.String(), err
}

// RunDockerCmdExec runs the docker command with the given args in a new process
// The stdout and stderr are captured, printed, and returned
// args will be passed to the docker command
// workingDir will be the directory the command runs in
func RunDockerCmdExec(args []string) (string, error) {

	cmdArgs := []string{"docker"}
	cmdArgs = append(cmdArgs, args...)
	fmt.Println(cmdArgs)

	execCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		return "", err
	}
	defer func() {
		// Make sure to close the writer first or this will hang on Windows
		outWriter.Close()
		outReader.Close()
	}()
	execCmd.Stdout = outWriter
	execCmd.Stderr = outWriter
	outScanner := bufio.NewScanner(outReader)
	var outBuffer bytes.Buffer
	go func() {
		for outScanner.Scan() {
			out := outScanner.Bytes()
			outBuffer.Write(out)
			outBuffer.WriteByte('\n')
			fmt.Println(string(out))
		}
	}()

	err = execCmd.Start()
	if err != nil {
		return "", err
	}

	// replace the original working directory when this function completes

	err = execCmd.Wait()

	return outBuffer.String(), err
}

func AddLocalFileRepo(repoName string, repoFilePath string, config *RootCommandConfig) (string, error) {
	absPath, err := filepath.Abs(repoFilePath)
	if err != nil {
		return "", err
	}
	var repoURL string
	if runtime.GOOS == "windows" {
		// for windows, add a leading slash and convert to unix style slashes
		absPath = "/" + filepath.ToSlash(absPath)
	}
	repoURL = "file://" + absPath
	// add a new repo
	err = repoAdd(repoName, repoURL, config)
	if err != nil {
		return "", err
	}

	return repoURL, err
}
