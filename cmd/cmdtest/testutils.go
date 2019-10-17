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
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/appsody/appsody/cmd"
	"gopkg.in/yaml.v2"
)

var realStdout = os.Stdout
var realStderr = os.Stderr

// Repository struct represents an appsody repository
type Repository struct {
	Name string
	URL  string
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

	args = append(args, "-v")

	// setup pipes to capture stdout and stderr of the command
	stdoutReader, stdoutWriter, _ := os.Pipe()
	os.Stdout = stdoutWriter
	stderrReader, stderrWriter, _ := os.Pipe()
	os.Stderr = stderrWriter
	var outBuf bytes.Buffer
	// setup writers to both os out and the buffer
	stdoutMultiWriter := io.MultiWriter(realStdout, &outBuf)
	stderrMultiWriter := io.MultiWriter(realStderr, &outBuf)

	// in the background, copy the output to the multiwriters
	var wg sync.WaitGroup
	wg.Add(2)
	var ioCopyErr error
	go func() {
		_, ioCopyErr = io.Copy(stdoutMultiWriter, stdoutReader)
		wg.Done()
	}()
	go func() {
		_, ioCopyErr = io.Copy(stderrMultiWriter, stderrReader)
		wg.Done()
	}()

	err := cmd.ExecuteE("vlatest", workingDir, args)
	// set back the os output right away so output gets displayed
	os.Stdout = realStdout
	os.Stderr = realStderr

	// close the writers first
	stdoutWriter.Close()
	stderrWriter.Close()
	// now wait for the io.Copy threads to finish
	wg.Wait()
	// finally close the readers
	stdoutReader.Close()
	stderrReader.Close()

	if ioCopyErr != nil {
		return outBuf.String(), fmt.Errorf("Problem copying command output to the writers: %v", ioCopyErr)
	}

	return outBuf.String(), err
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

// ParseYAML finds the start of the yaml string
func ParseYAML(output string) string {
	var outputLines = strings.Split(output, "\n")
	var splitIndex int
	for index, line := range outputLines {
		if (strings.HasPrefix(line, "-") || strings.Contains(line, ":")) && !strings.HasPrefix(line, "[") {
			splitIndex = index
			break
		}
	}

	return strings.Join(outputLines[splitIndex:], "\n")
}

// ParseRepoListJSON takes the json from 'appsody repo list -o json'
// and returns a RepositoryFile from the string.
func ParseRepoListJSON(jsonString string) (cmd.RepositoryFile, error) {
	var repos cmd.RepositoryFile
	e := json.Unmarshal([]byte(jsonString), &repos)
	if e != nil {
		return repos, e
	}
	return repos, nil
}

// ParseRepoListYAML takes the yaml from 'appsody repo list -o yaml'
// and returns a RepositoryFile from the string.
func ParseRepoListYAML(yamlString string) (cmd.RepositoryFile, error) {
	var repos cmd.RepositoryFile
	yamlString = strings.Replace(yamlString, "\n\n", "\n", -1)
	e := yaml.Unmarshal([]byte(yamlString), &repos)
	if e != nil {
		return repos, e
	}
	return repos, nil
}

// ParseListJSON takes the json from 'appsody list -o json'
// and returns an array of IndexOutputFormat from the string.
func ParseListJSON(jsonString string) (cmd.IndexOutputFormat, error) {
	var index cmd.IndexOutputFormat
	err := json.Unmarshal([]byte(jsonString), &index)
	if err != nil {
		return index, err
	}
	return index, nil
}

// ParseListYAML takes the yaml from 'appsody list -o yaml'
// and returns an array of IndexOutputFormat from the string.
func ParseListYAML(yamlString string) (cmd.IndexOutputFormat, error) {
	var index cmd.IndexOutputFormat
	err := yaml.Unmarshal([]byte(yamlString), &index)
	if err != nil {
		return index, err
	}
	return index, nil
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
	_, err = RunAppsodyCmd([]string{"repo", "add", repoName, repoURL}, ".")
	if err != nil {
		return "", nil, err
	}
	// cleanup whe finished
	cleanupFunc := func() {
		_, err = RunAppsodyCmd([]string{"repo", "remove", repoName}, ".")
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

func GetTempProjectDir(t *testing.T) string {
	// create a temporary dir to create the project and run the test
	projectDir, err := ioutil.TempDir("", "appsody-test")
	if err != nil {
		t.Fatal(err)
	}
	// remove symlinks from the path
	// on mac, TMPDIR is set to /var which is a symlink to /private/var.
	//    Docker by default shares mounts with /private but not /var,
	//    so resolving the symlinks ensures docker can mount the temp dir
	projectDir, err = filepath.EvalSymlinks(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	return projectDir
}
