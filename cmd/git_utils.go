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
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

type CommitInfo struct {
	Author         string
	AuthorEmail    string
	Committer      string
	CommitterEmail string
	SHA            string
	Date           string
	URL            string
	Message        string
	contextDir     string
}

type GitInfo struct {
	Branch    string
	Upstream  string
	RemoteURL string

	ChangesMade bool
	Commit      CommitInfo
}

const trimChars = "' \r\n"

func stringBefore(value string, searchValue string) string {
	// Get substring before a string.

	gitURLElements := strings.Split(value, searchValue)
	if len(gitURLElements) == 0 {
		return ""
	}
	return gitURLElements[0]

}

func stringAfter(value string, searchValue string) string {
	// Get substring after a string.
	position := strings.LastIndex(value, searchValue)
	if position == -1 {
		return ""
	}
	adjustedPosition := position + len(searchValue)
	if adjustedPosition >= len(value) {
		return ""
	}
	return value[adjustedPosition:]
}
func stringBetween(value string, pre string, post string) string {
	// Get substring between two strings.
	positionBegin := strings.Index(value, pre)
	if positionBegin == -1 {
		return ""
	}
	positionEnd := strings.Index(value, post)
	if positionEnd == -1 {
		return ""
	}
	positionBeginAdjusted := positionBegin + len(pre)
	if positionBeginAdjusted >= positionEnd {
		return ""
	}
	return value[positionBeginAdjusted:positionEnd]
}

//RunGitFindBranc issues git status
func GetGitInfo(config *RootCommandConfig) (GitInfo, error) {
	var gitInfo GitInfo
	version, vErr := RunGitVersion(config.LoggingConfig, false)
	if vErr != nil {
		return gitInfo, vErr
	}
	if version == "" {
		return gitInfo, errors.Errorf("git does not appear to be available")
	}

	config.Debug.log("git version: ", version)

	kargs := []string{"status", "-sb"}

	output, gitErr := RunGit(config.LoggingConfig, kargs, config.Dryrun)
	if gitErr != nil {
		return gitInfo, gitErr
	}

	lineSeparator := "\n"
	if runtime.GOOS == "windows" {
		lineSeparator = "\r\n"
	}
	output = strings.Trim(output, trimChars)
	outputLines := strings.Split(output, lineSeparator)

	const noCommits = "## No commits yet on "
	const branchPrefix = "## "
	const branchSeparatorString = "..."

	value := strings.Trim(outputLines[0], trimChars)

	if strings.HasPrefix(value, branchPrefix) {
		if strings.Contains(value, branchSeparatorString) {
			gitInfo.Branch = strings.Trim(stringBetween(value, branchPrefix, branchSeparatorString), trimChars)
			gitInfo.Upstream = strings.Trim(stringAfter(value, branchSeparatorString), trimChars)
			gitInfo.Upstream = strings.Split(gitInfo.Upstream, " ")[0]
		} else {
			gitInfo.Branch = strings.Trim(stringAfter(value, branchPrefix), trimChars)
		}

	}
	if strings.Contains(value, noCommits) {
		gitInfo.Branch = stringAfter(value, noCommits)
	}
	changesMade := false
	outputLength := len(outputLines)

	if outputLength > 1 {
		changesMade = true

	}
	gitInfo.ChangesMade = changesMade

	if gitInfo.Upstream != "" {
		gitInfo.RemoteURL, gitErr = RunGitConfigLocalRemoteOriginURL(config.LoggingConfig, gitInfo.Upstream, config.Dryrun)
		if gitErr != nil {
			config.Info.Logf("Could not construct repository URL %v", gitErr)
		}

	} else {
		config.Info.log("Unable to determine origin to compute repository URL")
	}

	gitInfo.Commit, gitErr = RunGitGetLastCommit(gitInfo.RemoteURL, config)
	if gitErr != nil {
		config.Info.log("Received error getting current commit: ", gitErr)
	}

	return gitInfo, nil
}

//RunGitConfigLocalRemoteOriginURL
func RunGitConfigLocalRemoteOriginURL(log *LoggingConfig, upstream string, dryrun bool) (string, error) {
	log.Info.log("Attempting to perform git config --local remote.<origin>.url  ...")

	upstreamStart := strings.Split(upstream, "/")[0]
	kargs := []string{"config", "--local", "remote." + upstreamStart + ".url"}
	remote, err := RunGit(log, kargs, dryrun)
	if err != nil {
		return remote, err
	}

	// Convert ssh remote to https
	if strings.Contains(remote, "git@") {
		remote = strings.Replace(remote, ":", "/", 1)
		remote = strings.Replace(remote, "git@", "https://", 1)
	}

	remote = strings.Replace(remote, ".git", "", 1)

	return remote, err
}

//RunGitLog issues git log
func RunGitGetLastCommit(URL string, config *RootCommandConfig) (CommitInfo, error) {
	//git log -n 1 --pretty=format:"{"author":"%cn","sha":"%h","date":"%cd”,}”
	kargs := []string{"log", "-n", "1", "--pretty=format:'{\"author\":\"%an\", \"authoremail\":\"%ae\", \"sha\":\"%H\", \"date\":\"%cd\", \"committer\":\"%cn\", \"committeremail\":\"%ce\", \"message\":\"%s\"}'"}
	var commitInfo CommitInfo
	commitStringInfo, gitErr := RunGit(config.LoggingConfig, kargs, config.Dryrun)
	if gitErr != nil {
		return commitInfo, gitErr
	}
	err := json.Unmarshal([]byte(strings.Trim(commitStringInfo, trimChars)), &commitInfo)
	if err != nil {
		return commitInfo, errors.Errorf("JSON Unmarshall error: %v", err)
	}
	if URL != "" {
		commitInfo.URL = stringBefore(URL, ".git") + "/commit/" + commitInfo.SHA
	}

	gitLocation, gitErr := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if gitErr != nil {
		return commitInfo, gitErr
	}
	gitLocationString := strings.TrimSpace(string(gitLocation))

	projectDir, err := getProjectDir(config)
	if err != nil {
		if _, ok := err.(*NotAnAppsodyProject); ok {
			// ignore this, we don't care it it is not an appsody project here
		} else {
			return commitInfo, err
		}
	}

	commitInfo.contextDir = strings.Replace(projectDir, gitLocationString, "", 1)

	return commitInfo, nil
}

//RunGitVersion
func RunGitVersion(log *LoggingConfig, dryrun bool) (string, error) {
	kargs := []string{"version"}
	versionInfo, gitErr := RunGit(log, kargs, dryrun)
	if gitErr != nil {
		return "", gitErr
	}
	return strings.Trim(versionInfo, trimChars), nil
}

//RunGit runs a generic git
func RunGit(log *LoggingConfig, kargs []string, dryrun bool) (string, error) {
	kcmd := "git"
	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", strings.Join(kargs, " "))
		return "", nil
	}
	log.Info.log("Running git command: ", kcmd, " ", strings.Join(kargs, " "))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := execCmd.Output()

	if kerr != nil {
		return "", errors.Errorf("git command failed: %s", string(kout[:]))
	}
	log.Debug.log("Command successful...")
	result := string(kout[:])
	result = strings.TrimRight(result, "\n")
	return result, nil
}
