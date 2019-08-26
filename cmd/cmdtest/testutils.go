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
package cmdtest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/appsody/appsody/cmd"
)

// Repository struct represents an appsody repository
type Repository struct {
	Name string
	URL  string
}

type RepositoryFile struct {
	APIVersion   string
	Generated    string
	Repositories []Repository
}

type Stack struct {
	ID          string
	Version     string
	Description string
}

// RunAppsodyCmdExec runs the appsody CLI with the given args in a new process
// The stdout and stderr are captured, printed, and returned
// args will be passed to the appsody command
// workingDir will be the directory the command runs in
func RunAppsodyCmdExec(args []string, workingDir string) (string, error) {
	execDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer func() {
		// replace the original working directory when this function completes
		err := os.Chdir(execDir)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// set the working directory
	if err := os.Chdir(workingDir); err != nil {
		return "", err
	}

	cmdArgs := []string{"go", "run", execDir + "/..", "-v"}
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
	err = os.Chdir(execDir)
	if err != nil {
		log.Fatal(err)
	}
	err = execCmd.Wait()

	return outBuffer.String(), err
}

// RunAppsodyCmd runs the appsody CLI with the given args
// The stdout and stderr are captured and returned
// args will be passed to the appsody command
// workingDir will be the directory the command runs in
func RunAppsodyCmd(args []string, workingDir string) (string, error) {
	// save off the original args and stdout/stderr streams
	osArgs := os.Args
	osStdout := os.Stdout
	osStderr := os.Stderr
	osDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer func() {
		// replace the original args and stdout/stderr streams when
		// RunAppsodyCmd returns
		os.Args = osArgs
		os.Stdout = osStdout
		os.Stderr = osStderr
		err := os.Chdir(osDir)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// set the working directory
	if err := os.Chdir(workingDir); err != nil {
		return "", err
	}

	// need to add "appsody" as the first arg and set os.Args
	os.Args = make([]string, len(args)+1)
	os.Args[0] = "appsody"
	copy(os.Args[1:], args)

	// setup pipes to capture stdout and stderr of the command
	stdoutReader, stdoutWriter, _ := os.Pipe()
	os.Stdout = stdoutWriter
	stderrReader, stderrWriter, _ := os.Pipe()
	os.Stderr = stderrWriter
	go func() {
		// run appsody cli in a goroutine so we don't
		// get infinite blocking pipes
		cmd.Execute(cmd.VERSION)
		defer func() {
			stdoutWriter.Close()
			stderrWriter.Close()
		}()
	}()
	// convert pipes to strings, this blocks until the writers are closed
	stdoutResult, stdoutErr := ioutil.ReadAll(stdoutReader)
	stderrResult, stderrErr := ioutil.ReadAll(stderrReader)
	output := string(stdoutResult) + "\n" + string(stderrResult)

	if stdoutErr != nil {
		err = stdoutErr
	} else if stderrErr != nil {
		err = stderrErr
	}

	return output, err
}

// ParseRepoList takes in the string from 'appsody repo list' command
// and returns an array of Repository structs from the string.
func ParseRepoList(repoListString string) []Repository {
	repoStrs := strings.Split(repoListString, "\n")
	var repos []Repository
	for _, repoStr := range repoStrs {
		fields := strings.Fields(repoStr)
		if len(fields) == 2 {
			if fields[0] != "NAME" && fields[0] != "Using" {
				repos = append(repos, Repository{fields[0], fields[1]})
			}
		}
	}
	return repos
}

// ParseJSON finds the json on the output string
func ParseJSON(repoListString string) string {
	jsonString := ""
	repoStrings := strings.Split(repoListString, "\n")
	for _, repoStr := range repoStrings {
		if strings.HasPrefix(repoStr, "{") || strings.HasPrefix(repoStr, "[{") {
			jsonString = repoStr
			break
		}
	}
	return jsonString
}

// ParseRepoListJSON takes the json from 'appsody repo list -o json'
// and returns a RepositoryFile from the string.
func ParseRepoListJSON(jsonString string) (*RepositoryFile, error) {
	var repos *RepositoryFile
	e := json.Unmarshal([]byte(jsonString), &repos)
	if e != nil {
		return nil, e
	}
	return repos, nil
}

// ParseListJSON takes the json from 'appsody list -o json'
// and returns an array of Stack from the string.
func ParseListJSON(jsonString string) ([]Stack, error) {
	var stacks []Stack
	e := json.Unmarshal([]byte(jsonString), &stacks)
	if e != nil {
		return nil, e
	}
	return stacks, nil
}

// AddLocalFileRepo calls the repo add command with the repo index located
// at the local file path. The path may be relative to the current working
// directory.
// Returns the URL of the repo added.
// Returns a function which should be deferred by the caller to cleanup
// the repo list when finished.
func AddLocalFileRepo(repoName string, repoFilePath string) (string, func(), error) {
	absPath, err := filepath.Abs(repoFilePath)
	if err != nil {
		return "", nil, err
	}
	var repoURL string
	if runtime.GOOS == "windows" {
		// for windows, add a leading slash and convert to unix style slashes
		absPath = "/" + filepath.ToSlash(absPath)
	}
	repoURL = "file://" + absPath
	// add a new repo
	_, err = RunAppsodyCmdExec([]string{"repo", "add", repoName, repoURL}, ".")
	if err != nil {
		return "", nil, err
	}
	// cleanup whe finished
	cleanupFunc := func() {
		_, err = RunAppsodyCmdExec([]string{"repo", "remove", repoName}, ".")
		if err != nil {
			log.Fatalf("Error cleaning up with repo remove: %s", err)
		}
	}

	return repoURL, cleanupFunc, err
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
