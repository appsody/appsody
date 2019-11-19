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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd"
	"gopkg.in/yaml.v2"
)

// Repository struct represents an appsody repository
type Repository struct {
	Name string
	URL  string
}

func inArray(haystack []string, needle string) bool {
	for _, value := range haystack {
		if needle == value {
			return true
		}
	}
	return false
}

// RunAppsodyCmdExec runs the appsody CLI with the given args in a new process
// The stdout and stderr are captured, printed, and returned
// args will be passed to the appsody command
// workingDir will be the directory the command runs in
func RunAppsodyCmdExec(args []string, workingDir string, t *testing.T) (string, error) {
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
	t.Log(cmdArgs)

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
			t.Log(string(out))
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

// RunAppsodyCmd runs the appsody CLI with the given args, in a custom
// home directory named after the currently executing test.
// The stdout and stderr are captured, printed and returned
// args will be passed to the appsody command
// projectDir will be the directory the command acts upon
func RunAppsodyCmd(args []string, projectDir string, t *testing.T) (string, error) {

	args = append(args, "-v")

	// TODO: make sure test home dirs are purged before tests are run

	if !inArray(args, "--config") {
		// Set appsody args to use custom home directory. Create the directory
		// if it does not already exist.
		testHomeDir := filepath.Join(os.TempDir(), "AppsodyTests", t.Name())
		err := os.MkdirAll(testHomeDir, 0755)
		if err != nil {
			return "", err
		}
		configFile := filepath.Join(testHomeDir, "config.yaml")

		// Create the config file if it does not already exist.
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			data := []byte("home: " + testHomeDir + "\n" + "generated-by-tests: Yes" + "\n")
			err = ioutil.WriteFile(configFile, data, 0644)
			if err != nil {
				return "", err
			}
		}

		// Pass custom config file to appsody
		args = append(args, "--config", configFile)
	}

	// // Buffer cmd output, to be logged if there is a failure
	var outBuffer bytes.Buffer

	// Direct cmd console output to a buffer
	outReader, outWriter, _ := os.Pipe()

	// copy the output to the buffer, and also to the test log
	outScanner := bufio.NewScanner(outReader)
	go func() {
		for outScanner.Scan() {
			out := outScanner.Bytes()
			outBuffer.Write(out)
			outBuffer.WriteByte('\n')
			t.Log(string(out))
		}
	}()

	err := cmd.ExecuteE("vlatest", "latest", projectDir, outWriter, outWriter, args)

	// close the reader and writer
	outWriter.Close()
	outReader.Close()

	return outBuffer.String(), err

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
func AddLocalFileRepo(repoName string, repoFilePath string, t *testing.T) (string, func(), error) {
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
	_, err = RunAppsodyCmd([]string{"repo", "add", repoName, repoURL}, ".", t)
	if err != nil {
		return "", nil, err
	}
	// cleanup whe finished
	cleanupFunc := func() {
		_, err = RunAppsodyCmd([]string{"repo", "remove", repoName}, ".", t)
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
func RunDockerCmdExec(args []string, t *testing.T) (string, error) {

	cmdArgs := []string{"docker"}
	cmdArgs = append(cmdArgs, args...)
	t.Log(cmdArgs)

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
			t.Log(string(out))
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

// Checks whether an inode (it does not bother
// about file or folder) exists or not.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
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
