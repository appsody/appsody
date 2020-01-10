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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/appsody/appsody/cmd"
	"gopkg.in/yaml.v2"
)

const CLEANUP = true

const TravisTesting = false

var TestDirPath = filepath.Join("..", "cmd", "testdata")

// Repository struct represents an appsody repository
type Repository struct {
	Name string
	URL  string
}

type TestSandbox struct {
	*testing.T
	ProjectDir   string
	TestDataPath string
	ProjectName  string
	ConfigDir    string
	ConfigFile   string
	Verbose      bool
}

func inArray(haystack []string, needle string) bool {
	for _, value := range haystack {
		if needle == value {
			return true
		}
	}
	return false
}

func TestSetup(t *testing.T, parallel bool) {
	if parallel {
		t.Parallel()
	}
}

func TestSetupWithSandbox(t *testing.T, parallel bool) (*TestSandbox, func()) {
	TestSetup(t, parallel)

	// default to verbose mode
	sandbox := &TestSandbox{T: t, Verbose: true}

	// create a temporary dir to create the project and run the test
	testDir, err := ioutil.TempDir("", "appsody-"+t.Name()+"-")
	if err != nil {
		t.Fatal("Error creating temporary directory: ", err)
	}
	// remove symlinks from the path
	// on mac, TMPDIR is set to /var which is a symlink to /private/var.
	//    Docker by default shares mounts with /private but not /var,
	//    so resolving the symlinks ensures docker can mount the temp dir
	testDir, err = filepath.EvalSymlinks(testDir)
	if err != nil {
		t.Fatal("Error evaluating symlinks: ", err)
	}
	sandbox.ProjectName = strings.ToLower(strings.Replace(filepath.Base(testDir), "appsody-", "", 1))
	sandbox.ProjectDir = filepath.Join(testDir, sandbox.ProjectName)
	sandbox.ConfigDir = filepath.Join(testDir, "config")
	sandbox.TestDataPath = filepath.Join(testDir, "testdata")
	err = os.MkdirAll(sandbox.ProjectDir, 0755)
	if err != nil {
		t.Fatal("Error creating project dir: ", err)
	}
	err = os.MkdirAll(sandbox.ConfigDir, 0755)
	if err != nil {
		t.Fatal("Error creating project dir: ", err)
	}

	t.Log("Created testing project dir: ", sandbox.ProjectDir)
	t.Log("Created testing config dir: ", sandbox.ConfigDir)

	// Create the config file if it does not already exist.
	sandbox.ConfigFile = filepath.Join(sandbox.ConfigDir, "config.yaml")
	data := []byte("home: " + sandbox.ConfigDir + "\n" + "generated-by-tests: Yes" + "\n")
	err = ioutil.WriteFile(sandbox.ConfigFile, data, 0644)
	if err != nil {
		t.Error("Error writing config file: ", err)
	}

	var outBuffer bytes.Buffer
	loggingConfig := &cmd.LoggingConfig{}
	loggingConfig.InitLogging(&outBuffer, &outBuffer)

	_, err = cmd.Exists(TestDirPath)
	if err != nil {
		t.Errorf("Cannot find source directory %s to copy", TestDirPath)
	}

	err = cmd.CopyDir(loggingConfig, TestDirPath, sandbox.TestDataPath)
	if err != nil {
		t.Errorf("Could not copy %s to %s - output of copy command %s", TestDirPath, testDir, err)
	}
	t.Logf("Directory copy of %s to %s was successful \n", TestDirPath, testDir)
	t.Log(outBuffer.String())

	configFolders := []string{"bad_format_repository_config", "default_repository_config", "empty_repository_config", "multiple_repository_config"}

	for _, config := range configFolders {
		file, err := ioutil.ReadFile(filepath.Join(sandbox.TestDataPath, config, "config.yaml"))

		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(string(file), "\n")

		for i, line := range lines {
			if strings.Contains(line, "home:") {
				t.Log("line = ", line)
				restoreLine := strings.SplitN(line, " ", -1)
				t.Log("rl ", restoreLine)
				lines[i] = "home: " + filepath.Join(testDir, restoreLine[1])
			}
		}

		output := strings.Join(lines, "\n")
		err = ioutil.WriteFile(filepath.Join(sandbox.TestDataPath, config, "config.yaml"), []byte(output), 0644)

		if err != nil {
			t.Fatal(err)
		}
	}

	cleanupFunc := func() {
		if CLEANUP {
			err := os.RemoveAll(testDir)
			if err != nil {
				t.Log("WARNING - ignoring error cleaning up test directory: ", err)
			}
		}
	}
	return sandbox, cleanupFunc
}

// RunAppsody runs the appsody CLI with the given args, using
// the sandbox for the project dir and config home.
// The stdout and stderr are captured, printed and returned
// args will be passed to the appsody command
func RunAppsody(t *TestSandbox, args ...string) (string, error) {

	if t.Verbose && !(inArray(args, "-v") || inArray(args, "--verbose")) {
		args = append(args, "-v")
	}

	if !inArray(args, "--config") {
		// Set appsody args to use custom home directory.
		args = append(args, "--config", t.ConfigFile)
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

	err := cmd.ExecuteE("vlatest", "latest", t.ProjectDir, outWriter, outWriter, args)

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

// AddLocalRepo calls the `appsody repo add` command with the repo index located
// at the local file path. The path may be relative to the current working directory.
// Returns the URL of the repo added.
func AddLocalRepo(t *TestSandbox, repoName string, repoFilePath string) (string, error) {
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
	_, err = RunAppsody(t, "repo", "add", repoName, repoURL)
	if err != nil {
		return "", err
	}

	return repoURL, nil
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
