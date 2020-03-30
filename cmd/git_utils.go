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
	"fmt"
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
	Pushed         bool
}

type GitInfo struct {
	Branch    string
	Upstream  string
	RemoteURL string

	ChangesMade bool
	Commit      CommitInfo
}

const trimChars = "' \r\n"

func StringBefore(value string, searchValue string) string {
	// Get substring before a string.

	gitURLElements := strings.Split(value, searchValue)
	if len(gitURLElements) == 0 {
		return ""
	}
	return gitURLElements[0]

}

func StringAfter(value string, searchValue string) string {
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
func StringBetween(value string, pre string, post string) string {
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

//RunGitFindBranch issues git status
func GetGitInfo(config *RootCommandConfig) (GitInfo, error) {
	const noCommits = "## No commits yet on "
	const branchPrefix = "## "
	const branchSeparatorString = "..."
	var gitInfo GitInfo
	var gitErr error
	var noRemoteFound bool
	errMsg := ""

	version, vErr := RunGitVersion(config.LoggingConfig, config.ProjectDir, false)
	if vErr != nil {
		return gitInfo, vErr
	}
	if version == "" {
		return gitInfo, errors.Errorf("git does not appear to be available")
	}

	config.Debug.log("git version: ", version)

	gitInfo.Commit, gitErr = RunGitGetLastCommit(config)
	if gitErr != nil {
		errMsg += "Received error getting current commit: " + gitErr.Error()
	}
	gitInfo.Commit.Pushed = true

	lineSeparator := "\n"
	if runtime.GOOS == "windows" {
		lineSeparator = "\r\n"
	}

	kargs := []string{"status", "-sb"}
	statusOutput, gitErr := RunGit(config.LoggingConfig, config.ProjectDir, kargs, config.Dryrun)
	if gitErr != nil {
		return gitInfo, errors.Errorf("%v. Error running git status -sb. Full error: %v", errMsg, gitErr)
	}

	statusOutput = strings.Trim(statusOutput, trimChars)
	statusOutputLines := strings.Split(statusOutput, lineSeparator)

	value := strings.Trim(statusOutputLines[0], trimChars)

	if strings.HasPrefix(value, branchPrefix) {
		if strings.Contains(value, branchSeparatorString) {
			gitInfo.Branch = strings.Trim(StringBetween(value, branchPrefix, branchSeparatorString), trimChars)
			gitInfo.Upstream = strings.Trim(StringAfter(value, branchSeparatorString), trimChars)
			gitInfo.Upstream = strings.Split(gitInfo.Upstream, " ")[0]
		} else {
			gitInfo.Branch = strings.Trim(StringAfter(value, branchPrefix), trimChars)
		}

	}

	if strings.Contains(value, noCommits) {
		gitInfo.Branch = StringAfter(value, noCommits)
	}

	changesMade := false
	outputLength := len(statusOutputLines)

	if outputLength > 1 {
		changesMade = true
	}
	gitInfo.ChangesMade = changesMade

	outputLines, err := RunGitBranchContains(config.LoggingConfig, gitInfo.Commit.SHA, config.ProjectDir, lineSeparator, config.Dryrun)
	if err != nil {
		noRemoteFound = true
	} else {
		if gitInfo.Upstream != "" {
			for _, upstream := range outputLines {
				if gitInfo.Upstream == upstream {
					gitInfo.Commit.Pushed = true
					break
				}
			}
		} else {
			gitInfo.Upstream = outputLines[0]
			for _, upstream := range outputLines {
				if upstream == "origin" {
					gitInfo.Upstream = upstream
					break
				} else if upstream == "upstream" {
					gitInfo.Upstream = upstream
				}
			}
			gitInfo.Upstream = strings.TrimSpace(gitInfo.Upstream)
			config.Debug.log("Successfully retrieved upstream via git branch --contains")
		}

	}

	if gitInfo.Upstream == "" {
		gitInfo.Commit.Pushed = false
		gitRemoteOutput, err := RunGitRemote(config.LoggingConfig, config.ProjectDir, lineSeparator, config.Dryrun)
		if err != nil {
			return gitInfo, errors.Errorf("%v. Error running git remote. Full error: %v", errMsg, err)
		}
		gitInfo.Upstream = gitRemoteOutput[0]
		for _, remote := range gitRemoteOutput {
			if remote == "origin" {
				gitInfo.Upstream = remote
				break
			} else if remote == "upstream" {
				gitInfo.Upstream = remote
			}
		}
	}

	if gitInfo.Upstream != "" {
		config.Debug.log("Successfully retrieved remote name")
		noRemoteFound = false
		gitInfo.RemoteURL, gitErr = RunGitConfigLocalRemoteOriginURL(config.LoggingConfig, config.ProjectDir, gitInfo.Upstream, config.Dryrun)
		if gitErr != nil {
			errMsg += fmt.Sprintf("Could not construct repository URL %v ", gitErr)
		}

	} else {
		errMsg += "Unable to determine origin to compute repository URL "
	}

	if noRemoteFound {
		errMsg += "Unable to retrieve remote via git status or git branch --contains"
	}

	if gitInfo.RemoteURL != "" {
		gitInfo.Commit.setURL(gitInfo.RemoteURL)
	}

	if errMsg != "" {
		return gitInfo, errors.New(errMsg)
	}
	return gitInfo, nil
}

func RunGitBranchContains(log *LoggingConfig, commitSHA string, workDir string, lineSeparator string, dryrun bool) ([]string, error) {
	log.Debug.log("Attempting to run git branch -r --contains CommitSHA")

	kargs := []string{"branch", "-r", "--contains", commitSHA}

	output, gitErr := RunGit(log, workDir, kargs, dryrun)
	if gitErr != nil {
		return []string{}, gitErr
	}

	if output == "" {
		return []string{}, errors.New("No remotes returned from git branch command")
	}

	outputLines := strings.Split(output, lineSeparator)
	return outputLines, nil

}

func RunGitRemote(log *LoggingConfig, workDir string, lineSeparator string, dryrun bool) ([]string, error) {
	log.Debug.log("Attempting to run git remote")

	kargs := []string{"remote"}
	remoteOutput, err := RunGit(log, workDir, kargs, dryrun)
	if err != nil {
		return []string{}, err
	}

	if remoteOutput == "" {
		return []string{}, errors.New("No remotes returned from git remote command")
	}

	remoteOutputLines := strings.Split(remoteOutput, lineSeparator)
	return remoteOutputLines, nil
}

//RunGitConfigLocalRemoteOriginURL
func RunGitConfigLocalRemoteOriginURL(log *LoggingConfig, workDir string, upstream string, dryrun bool) (string, error) {
	log.Debug.log("Attempting to perform git config --local remote.<origin>.url  ...")

	upstreamStart := strings.Split(upstream, "/")[0]
	kargs := []string{"config", "--local", "remote." + upstreamStart + ".url"}
	remote, err := RunGit(log, workDir, kargs, dryrun)
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

func (commitInfo *CommitInfo) setURL(URL string) {
	if URL != "" {
		commitInfo.URL = StringBefore(URL, ".git") + "/commit/" + commitInfo.SHA
	}
}

//RunGitLog issues git log
func RunGitGetLastCommit(config *RootCommandConfig) (CommitInfo, error) {
	//git log -n 1 --pretty=format:"{"author":"%cn","sha":"%h","date":"%cd”,}”
	kargs := []string{"log", "-n", "1", "--pretty=format:'{\"author\":\"%an\", \"authoremail\":\"%ae\", \"sha\":\"%H\", \"date\":\"%cd\", \"committer\":\"%cn\", \"committeremail\":\"%ce\", \"message\":\"%s\"}'"}
	var commitInfo CommitInfo
	commitStringInfo, gitErr := RunGit(config.LoggingConfig, config.ProjectDir, kargs, config.Dryrun)
	if gitErr != nil {
		return commitInfo, gitErr
	}
	err := json.Unmarshal([]byte(strings.Trim(commitStringInfo, trimChars)), &commitInfo)
	if err != nil {
		return commitInfo, errors.Errorf("JSON Unmarshall error: %v", err)
	}
	gitLocation, gitErr := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if gitErr != nil {
		return commitInfo, gitErr
	}
	gitLocationString := strings.TrimSpace(string(gitLocation))

	projectDir, err := getProjectDir(config)
	if err != nil {
		if _, ok := err.(*NotAnAppsodyProject); !ok {
			return commitInfo, err
		}
	}

	commitInfo.contextDir = strings.Replace(projectDir, gitLocationString, "", 1)

	return commitInfo, nil
}

//RunGitVersion
func RunGitVersion(log *LoggingConfig, workDir string, dryrun bool) (string, error) {
	kargs := []string{"version"}
	versionInfo, gitErr := RunGit(log, workDir, kargs, dryrun)
	if gitErr != nil {
		return "", gitErr
	}
	return strings.Trim(versionInfo, trimChars), nil
}

//RunGit runs a generic git
func RunGit(log *LoggingConfig, workDir string, kargs []string, dryrun bool) (string, error) {
	kcmd := "git"
	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return "", nil
	}
	log.Debug.log("Running git command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)
	execCmd.Dir = workDir
	kout, kerr := SeparateOutput(execCmd)

	if kerr != nil {
		return "", errors.Errorf("git command failed: %s", string(kout[:]))
	}
	log.Debug.log("Command successful...")
	result := string(kout[:])
	result = strings.TrimRight(result, "\n")
	return result, nil
}
