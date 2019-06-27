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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type RepoIndex struct {
	APIVersion string                     `yaml:"apiVersion"`
	Generated  time.Time                  `yaml:"generated"`
	Projects   map[string]ProjectVersions `yaml:"projects"`
}

type ProjectVersions []*ProjectVersion

type ProjectVersion struct {
	APIVersion  string    `yaml:"apiVersion"`
	Created     time.Time `yaml:"created"`
	Name        string    `yaml:"name"`
	Home        string    `yaml:"home"`
	Version     string    `yaml:"version"`
	Description string    `yaml:"description"`
	Keywords    []string  `yaml:"keywords"`
	Maintainers []string  `yaml:"maintainers"`
	Icon        string    `yaml:"icon"`
	Digest      string    `yaml:"digest"`
	URLs        []string  `yaml:"urls"`
}

type RepositoryFile struct {
	APIVersion   string             `yaml:"apiVersion"`
	Generated    time.Time          `yaml:"generated"`
	Repositories []*RepositoryEntry `yaml:"repositories"`
}

type RepositoryEntry struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

var (
	appsodyHubURL = "https://raw.githubusercontent.com/appsody/stacks/master/index.yaml"
)

// repoCmd represents the repo command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage your Appsody repositories",
	Long:  ``,
}

func getHome() string {
	return cliConfig.GetString("home")
}

func getRepoDir() string {
	return filepath.Join(getHome(), "repository")
}

func getRepoFileLocation() string {
	return filepath.Join(getRepoDir(), "repository.yaml")
}

func init() {
	rootCmd.AddCommand(repoCmd)
}

// Locate or create config structure in $APPSODY_HOME
func ensureConfig() {
	directories := []string{
		getHome(),
		getRepoDir(),
	}

	for _, p := range directories {
		if fi, err := os.Stat(p); err != nil {

			if dryrun {
				Info.log("Dry Run - Skipping create of directory ", p)
			} else {
				Debug.log("Creating ", p)
				if err := os.MkdirAll(p, 0755); err != nil {
					Error.logf("Could not create %s: %s", p, err)
					os.Exit(1)
				}
			}

		} else if !fi.IsDir() {
			Error.logf("%s must be a directory", p)
			os.Exit(1)
		}
	}

	// Repositories file
	var repoFileLocation = getRepoFileLocation()
	if file, err := os.Stat(repoFileLocation); err != nil {

		if dryrun {
			Info.log("Dry Run - Skipping creation of appsodyhub repo: ", appsodyHubURL)
		} else {

			repo := NewRepoFile()
			repo.Add(&RepositoryEntry{
				Name: "appsodyhub",
				URL:  appsodyHubURL,
			})
			Debug.log("Creating ", repoFileLocation)
			if err := repo.WriteFile(repoFileLocation); err != nil {
				Error.logf("Error writing %s file: %s ", repoFileLocation, err)
				os.Exit(1)
			}
		}
	} else if file.IsDir() {
		Error.logf("%s must be a file, not a directory ", repoFileLocation)
		os.Exit(1)
	}

	defaultConfigFile := getDefaultConfigFile()
	if _, err := os.Stat(defaultConfigFile); err != nil {
		if dryrun {
			Info.log("Dry Run - Skip creation of default config file ", defaultConfigFile)
		} else {
			Debug.log("Creating ", defaultConfigFile)
			if err := ioutil.WriteFile(defaultConfigFile, []byte{}, 0644); err != nil {
				Error.logf("Error creating default config file %s", err)
				os.Exit(1)
			}
		}
	}

	if dryrun {
		Info.log("Dry Run - Skip writing config file ", defaultConfigFile)
	} else {
		Debug.log("Writing config file ", defaultConfigFile)
		if err := cliConfig.WriteConfig(); err != nil {
			Error.logf("Writing default config file %s", err)
			os.Exit(1)
		}
	}

}

func downloadIndex(href string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	// allow file:// scheme
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	httpClient := &http.Client{Transport: t}

	Debug.log("Downloading appsody repository index from ", href)
	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return buf, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return buf, err
	}
	if resp.StatusCode != 200 {
		return buf, fmt.Errorf("Failed to fetch %s : %s", href, resp.Status)
	}

	_, err = io.Copy(buf, resp.Body)
	resp.Body.Close()
	return buf, err
}

func (index *RepoIndex) getIndex() *RepoIndex {
	var repos RepositoryFile
	repos.getRepos()

	var masterIndex *RepoIndex
	for _, value := range repos.Repositories {
		indexBuffer, err := downloadIndex(value.URL)
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
		}

		yamlFile, err := ioutil.ReadAll(indexBuffer)
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(yamlFile, index)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
		if masterIndex == nil {
			masterIndex = index
		} else {
			for name, project := range index.Projects {
				masterIndex.Projects[name] = project
			}
		}
	}

	return masterIndex
}

func (index *RepoIndex) listProjects() string {
	table := uitable.New()
	table.MaxColWidth = 60
	table.AddRow("NAME", "VERSION", "DESCRIPTION")
	for _, value := range index.Projects {
		table.AddRow(value[0].Name, value[0].Version, value[0].Description)
	}

	return table.String()
}

func (r *RepositoryFile) getRepos() *RepositoryFile {
	var repoFileLocation = getRepoFileLocation()
	repoReader, err := ioutil.ReadFile(repoFileLocation)
	if err != nil {
		if os.IsNotExist(err) {
			Error.logf("Repository file does not exist %s. Check to make sure appsody init has been run. ", repoFileLocation)
			os.Exit(1)
		} else {
			Error.log("Failed reading repository file ", repoFileLocation)
			os.Exit(1)
		}
	}
	err = yaml.Unmarshal(repoReader, r)
	if err != nil {
		Error.log("Failed to parse repository file ", err)
		os.Exit(1)
	}
	return r
}

func (r *RepositoryFile) listRepos() string {
	table := uitable.New()
	table.MaxColWidth = 120
	table.AddRow("NAME", "URL")
	for _, value := range r.Repositories {
		table.AddRow(value.Name, value.URL)
	}

	return table.String()
}

func NewRepoFile() *RepositoryFile {
	return &RepositoryFile{
		APIVersion:   APIVersionV1,
		Generated:    time.Now(),
		Repositories: []*RepositoryEntry{},
	}
}

func (r *RepositoryFile) Add(re ...*RepositoryEntry) {
	r.Repositories = append(r.Repositories, re...)
}

func (r *RepositoryFile) Has(name string) bool {
	for _, rf := range r.Repositories {
		if rf.Name == name {
			return true
		}
	}
	return false
}

func (r *RepositoryFile) Remove(name string) {
	for index, rf := range r.Repositories {
		if rf.Name == name {
			r.Repositories[index] = r.Repositories[0]
			r.Repositories = r.Repositories[1:]
			return
		}
	}
}

func (r *RepositoryFile) WriteFile(path string) error {
	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}
