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
	"runtime"
	"time"
)

func getLastCheckTime(config *RootCommandConfig) string {
	return config.CliConfig.GetString("lastversioncheck")
}

func checkTime(config *RootCommandConfig) {
	var lastCheckTime = getLastCheckTime(config)

	lastTime, err := time.Parse("2006-01-02 15:04:05 -0700 MST", lastCheckTime)
	if err != nil {
		config.Debug.logf("Could not parse the config file's lastversioncheck: %v. Continuing with a new version check...", err)
		doVersionCheck(config)
	} else if time.Since(lastTime).Hours() > 24 {
		doVersionCheck(config)
	}

}

// TEMPORARY CODE: sets the old repo name "appsodyhub" to the new name "incubator"
// this code should be removed when we think everyone is using the new name.
func setNewRepoName(config *RootCommandConfig) {
	var repoFile RepositoryFile
	_, repoErr := repoFile.getRepoFile(config)
	if repoErr != nil {
		config.Warning.log("Unable to read repository file")
	}
	appsodyhubRepo := repoFile.GetRepo("appsodyhub")
	if appsodyhubRepo != nil && appsodyhubRepo.URL == incubatorRepositoryURL {
		config.Info.log("Migrating your repo name from 'appsodyhub' to 'incubator'")
		appsodyhubRepo.Name = "incubator"
		err := repoFile.WriteFile(getRepoFileLocation(config))
		if err != nil {
			config.Warning.logf("Failed to write file to repository location: %v", err)
		}
	}
}

func doVersionCheck(config *RootCommandConfig) {
	var latest = getLatestVersion(config.LoggingConfig)
	var currentTime = time.Now().Format("2006-01-02 15:04:05 -0700 MST")
	if latest != "" && VERSION != "vlatest" && VERSION != latest {
		updateString := GetUpdateString(runtime.GOOS, VERSION, latest)
		config.Warning.logf(updateString)
	}

	config.CliConfig.Set("lastversioncheck", currentTime)
	if err := config.CliConfig.WriteConfig(); err != nil {
		config.Error.logf("Writing default config file %s", err)

	}
}

// GetUpdateString Returns a format string to advise the user how to upgrade
func GetUpdateString(osName string, version string, latest string) string {
	var updateString string
	switch osName {
	case "darwin":
		updateString = "Please run `brew upgrade appsody` to upgrade"
	default:
		updateString = "Please go to https://appsody.dev/docs/getting-started/installation#upgrading-appsody and upgrade"
	}
	return fmt.Sprintf("\n*\n*\n*\n\nA new CLI update is available.\n%s from %s --> %s.\n\n*\n*\n*\n", updateString, version, latest)
}
