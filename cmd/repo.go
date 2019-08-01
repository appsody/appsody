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
	//"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type RepoIndex struct {
	APIVersion string                     `yaml:"apiVersion"`
	Generated  time.Time                  `yaml:"generated"`
	Projects   map[string]ProjectVersions `yaml:"projects"`
	Stacks     []ProjectVersion           `yaml:"stacks"`
}

// RepoIndices maps repos to their RepoIndex (i.e. the projects in a repo)
type RepoIndices map[string]*RepoIndex

type ProjectVersions []*ProjectVersion

type ProjectVersion struct {
	APIVersion  string        `yaml:"apiVersion"`
	Id          string        `yaml:"id,omitempty"`
	Created     time.Time     `yaml:"created"`
	Name        string        `yaml:"name"`
	Home        string        `yaml:"home"`
	Version     string        `yaml:"version"`
	Description string        `yaml:"description"`
	Keywords    []string      `yaml:"keywords"`
	Maintainers []interface{} `yaml:"maintainers"`
	Icon        string        `yaml:"icon"`
	Digest      string        `yaml:"digest"`
	URLs        []string      `yaml:"urls"`
}

type RepositoryFile struct {
	APIVersion   string             `yaml:"apiVersion"`
	Generated    time.Time          `yaml:"generated"`
	Repositories []*RepositoryEntry `yaml:"repositories"`
}

type RepositoryEntry struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	IsDefault bool   `yaml:"default,omitempty"`
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

var ensureConfigRun = false

// Locate or create config structure in $APPSODY_HOME
func ensureConfig() error {
	if ensureConfigRun {
		return nil
	}
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
					return errors.Errorf("Could not create %s: %s", p, err)

				}
			}

		} else if !fi.IsDir() {
			return errors.Errorf("%s must be a directory", p)

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
				Name:      "appsodyhub",
				URL:       appsodyHubURL,
				IsDefault: true,
			})

			Debug.log("Creating ", repoFileLocation)
			if err := repo.WriteFile(repoFileLocation); err != nil {
				return errors.Errorf("Error writing %s file: %s ", repoFileLocation, err)
			}
		}
	} else if file.IsDir() {
		return errors.Errorf("%s must be a file, not a directory ", repoFileLocation)
	}

	defaultConfigFile := getDefaultConfigFile()
	if _, err := os.Stat(defaultConfigFile); err != nil {
		if dryrun {
			Info.log("Dry Run - Skip creation of default config file ", defaultConfigFile)
		} else {
			Debug.log("Creating ", defaultConfigFile)
			if err := ioutil.WriteFile(defaultConfigFile, []byte{}, 0644); err != nil {
				return errors.Errorf("Error creating default config file %s", err)

			}
		}
	}

	if dryrun {
		Info.log("Dry Run - Skip writing config file ", defaultConfigFile)
	} else {
		Debug.log("Writing config file ", defaultConfigFile)
		if err := cliConfig.WriteConfig(); err != nil {
			return errors.Errorf("Writing default config file %s", err)

		}
	}
	ensureConfigRun = true
	return nil

}

func downloadFile(href string, writer io.Writer) error {

	// allow file:// scheme
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	Debug.log("Proxy function for HTTP transport set to: ", &t.Proxy)
	if runtime.GOOS == "windows" {
		// For Windows, remove the root url. It seems to work fine with an empty string.
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("")))
	} else {
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	}

	httpClient := &http.Client{Transport: t}

	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Debug.log("Could not read contents of response body: ", err)
		} else {
			Debug.logf("Contents http response:\n%s", buf)
		}
		resp.Body.Close()
		return fmt.Errorf("%s response trying to download %s", resp.Status, href)
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("Could not copy http response body to writer: %s", err)
	}
	resp.Body.Close()
	return nil
}

func downloadIndex(url string) (*RepoIndex, error) {
	Debug.log("Downloading appsody repository index from ", url)
	indexBuffer := bytes.NewBuffer(nil)
	err := downloadFile(url, indexBuffer)
	if err != nil {
		return nil, errors.Errorf("Failed to get repository index: %s", err)
	}

	yamlFile, err := ioutil.ReadAll(indexBuffer)
	if err != nil {
		return nil, fmt.Errorf("Could not read buffer into byte array")
	}
	var index RepoIndex
	err = yaml.Unmarshal(yamlFile, &index)
	if err != nil {
		Debug.logf("Contents of downloaded index from %s\n%s", url, yamlFile)
		return nil, fmt.Errorf("Repository index formatting error: %s", err)
	}
	return &index, nil
}

func (index *RepoIndex) listProjects(repoName string) string {
	table := uitable.New()
	table.MaxColWidth = 60
	table.AddRow("REPO", "ID", "VERSION", "DESCRIPTION")
	for id, value := range index.Projects {
		table.AddRow(repoName, id, value[0].Version, value[0].Description)
	}
	for _, value := range index.Stacks {
		table.AddRow(repoName, value.Id, value.Version, value.Description)
	}
	return table.String()
}
func (r *RepositoryFile) listProjects() (string, error) {
	table := uitable.New()
	table.MaxColWidth = 60
	//table.AddRow("REPO", "ID", "VERSION", "TEMPLATES", "DESCRIPTION")
	table.AddRow("REPO", "ID", "VERSION", "DESCRIPTION")
	indices, err := r.GetIndices()
	//rnd := rand.New(rand.NewSource(99))

	//err := index.getIndex()
	//templates := [8]string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"}
	if err != nil {
		return "", errors.Errorf("Could not read indices: %v", err)
	}
	if len(indices) != 0 {
		for repoName, index := range indices {
			//Info.log("\n", "Repository: ", repoName)
			for id, value := range index.Projects {
				//r1 := rnd.Intn(8)
				//r2 := rnd.Intn(8)
				//r3 := rnd.Intn(8)
				//rndTemplates := "*" + templates[r1] + ", " + templates[r2] + ", " + templates[r3]
				//table.AddRow(repoName, id, value[0].Version, rndTemplates, value[0].Description)
				table.AddRow(repoName, id, value[0].Version, value[0].Description)
			}
			for _, value := range index.Stacks {
				table.AddRow(repoName, value.Id, value.Version, value.Description)
			}
		}
		return table.String(), nil
	}
	return "", errors.New("there are no repositories in your configuration")

}
func (r *RepositoryFile) listRepoProjects(repoName string) (string, error) {
	if repo := r.GetRepo(repoName); repo != nil {
		url := repo.URL
		index, err := downloadIndex(url)
		if err != nil {
			return "", err
		}
		return index.listProjects(repoName), nil
	}
	return "", errors.New("cannot locate repository named " + repoName)
}

func (r *RepositoryFile) getRepos() (*RepositoryFile, error) {
	var repoFileLocation = getRepoFileLocation()
	repoReader, err := ioutil.ReadFile(repoFileLocation)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("Repository file does not exist %s. Check to make sure appsody init has been run. ", repoFileLocation)

		}
		return nil, errors.Errorf("Failed reading repository file %s", repoFileLocation)

	}
	err = yaml.Unmarshal(repoReader, r)
	if err != nil {
		return nil, errors.Errorf("Failed to parse repository file %v", err)

	}
	return r, nil
}

func (r *RepositoryFile) listRepos() string {
	table := uitable.New()
	table.MaxColWidth = 120
	table.AddRow("NAME", "URL")
	for _, value := range r.Repositories {
		var repoName string
		if repoName = value.Name; repoName == r.GetDefaultRepoName() {
			repoName = "*" + repoName
		}
		table.AddRow(repoName, value.URL)
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
func (r *RepositoryFile) GetRepo(name string) *RepositoryEntry {
	for _, rf := range r.Repositories {
		if rf.Name == name {
			return rf
		}
	}
	return nil
}

func (r *RepositoryFile) HasURL(url string) bool {
	for _, rf := range r.Repositories {
		if rf.URL == url {
			return true
		}
	}
	return false
}
func (r *RepositoryFile) GetDefaultRepoName() string {
	for _, rf := range r.Repositories {
		if rf.IsDefault {
			return rf.Name
		}
	}
	return ""
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

func (r *RepositoryFile) GetIndices() (RepoIndices, error) {

	indices := make(map[string]*RepoIndex)
	for _, rf := range r.Repositories {
		var index, err = downloadIndex(rf.URL)
		if err != nil {
			return indices, err
		}
		indices[rf.Name] = index
	}
	return indices, nil
}
