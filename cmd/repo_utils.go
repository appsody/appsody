// Copyright © 2020 IBM Corporation and others.
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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Stack struct {
	repoName    string
	ID          string     `yaml:"id,omitempty" json:"id,omitempty"`
	Version     string     `yaml:"version" json:"version"`
	Description string     `yaml:"description" json:"description"`
	Templates   []Template `yaml:"templates,omitempty" json:"templates,omitempty"`
}

// RepoIndex - Structure of a repo index YAML
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
	APIVersion      string           `yaml:"apiVersion"`
	ID              string           `yaml:"id,omitempty"`
	Created         time.Time        `yaml:"created"`
	Name            string           `yaml:"name"`
	Home            string           `yaml:"home"`
	Version         string           `yaml:"version"`
	Description     string           `yaml:"description"`
	Keywords        []string         `yaml:"keywords"`
	Maintainers     []interface{}    `yaml:"maintainers"`
	Requirements    StackRequirement `yaml:"requirements,omitempty"`
	Icon            string           `yaml:"icon"`
	Digest          string           `yaml:"digest"`
	URLs            []string         `yaml:"urls"` //V1
	Templates       []Template       `yaml:"templates,omitempty"`
	DefaultTemplate string           `yaml:"default-template"`
}

//StackRequirement - Structure that holds information for minimum requirements to use stack stack
type StackRequirement struct {
	Docker  string `yaml:"docker-version,omitempty"`
	Appsody string `yaml:"appsody-version,omitempty"`
	Buildah string `yaml:"buildah-version,omitempty"`
}

// RepositoryFile - Structure of the repository.yaml file in the .appsody directory
type RepositoryFile struct {
	APIVersion   string             `yaml:"apiVersion" json:"apiVersion"`
	Generated    time.Time          `yaml:"generated" json:"generated"`
	Repositories []*RepositoryEntry `yaml:"repositories" json:"repositories"`
}

//RepositoryEntry - Strcuture for how a repo entry appears in the RepositoryFile
type RepositoryEntry struct {
	Name      string `yaml:"name" json:"name"`
	URL       string `yaml:"url" json:"url"`
	IsDefault bool   `yaml:"default,omitempty" json:"default,omitempty"`
}

//Template -Structure of template entry in the Repository Index
type Template struct {
	ID        string `yaml:"id" json:"id"`
	URL       string `yaml:"url" json:"url"`
	IsDefault bool   `yaml:"default,omitempty" json:"default,omitempty"`
}

func findTemplateURL(projectVersion ProjectVersion, templateName string) string {
	templates := projectVersion.Templates

	for _, value := range templates {
		if value.ID == templateName {
			return value.URL
		}

	}
	return ""
}

type indexError struct {
	indexName string
	theError  error
}

type indexErrors struct {
	listOfErrors []indexError
}

func (e indexError) Error() string {
	return "- Repository: " + e.indexName + "\n  Reason: " + e.theError.Error()
}

func (e indexErrors) Error() string {
	var myerrors string
	for _, err := range e.listOfErrors {
		myerrors = myerrors + fmt.Sprintf("%v\n", err)
	}
	return myerrors
}

const (
	supportedIndexAPIVersion  = "v2"
	incubatorRepositoryURL    = "https://github.com/appsody/stacks/releases/latest/download/incubator-index.yaml"
	experimentalRepositoryURL = "https://github.com/appsody/stacks/releases/latest/download/experimental-index.yaml"
)

func getHome(rootConfig *RootCommandConfig) string {
	return rootConfig.CliConfig.GetString("home")
}

func getRepoDir(rootConfig *RootCommandConfig) string {
	return filepath.Join(getHome(rootConfig), "repository")
}

func getRepoFileLocation(rootConfig *RootCommandConfig) string {
	return filepath.Join(getRepoDir(rootConfig), "repository.yaml")
}

// Locate or create config structure in $APPSODY_HOME
func ensureConfig(rootConfig *RootCommandConfig) error {
	directories := []string{
		getHome(rootConfig),
		getRepoDir(rootConfig),
	}

	for _, p := range directories {
		if fi, err := os.Stat(p); err != nil {

			if rootConfig.Dryrun {
				rootConfig.Info.log("Dry Run - Skipping create of directory ", p)
			} else {
				rootConfig.Debug.log("Creating ", p)
				if err := os.MkdirAll(p, 0755); err != nil {
					return errors.Errorf("Could not create %s: %s", p, err)

				}
			}

		} else if !fi.IsDir() {
			return errors.Errorf("%s must be a directory", p)

		}
	}

	// Repositories file
	var repoFileLocation = getRepoFileLocation(rootConfig)
	if file, err := os.Stat(repoFileLocation); err != nil {

		if rootConfig.Dryrun {
			rootConfig.Info.log("Dry Run - Skipping creation of incubator repo: ", incubatorRepositoryURL)
		} else {

			repo := NewRepoFile()
			repo.AddRepo(&RepositoryEntry{
				Name:      "incubator",
				URL:       incubatorRepositoryURL,
				IsDefault: true,
			})
			repo.AddRepo(&RepositoryEntry{
				Name: "experimental",
				URL:  experimentalRepositoryURL,
			})
			rootConfig.Debug.log("Creating ", repoFileLocation)
			if err := repo.WriteFile(repoFileLocation); err != nil {
				return errors.Errorf("Error writing %s file: %s ", repoFileLocation, err)
			}
		}
	} else if file.IsDir() {
		return errors.Errorf("%s must be a file, not a directory ", repoFileLocation)
	}

	defaultConfigFile := getDefaultConfigFile(rootConfig)
	if _, err := os.Stat(defaultConfigFile); err != nil {
		if rootConfig.Dryrun {
			rootConfig.Info.log("Dry Run - Skip creation of default config file ", defaultConfigFile)
		} else {
			rootConfig.Debug.log("Creating ", defaultConfigFile)
			if err := ioutil.WriteFile(defaultConfigFile, []byte{}, 0644); err != nil {
				return errors.Errorf("Error creating default config file %s", err)

			}
		}
	}

	// TEMPORARY CODE: if the user has the legacy index.docker.io value, update it to docker.io
	images := rootConfig.CliConfig.GetString("images")
	if images == "index.docker.io" {
		rootConfig.Debug.logf("Updating reference to 'index.docker.io' in %s", rootConfig.CliConfig.ConfigFileUsed())
		rootConfig.CliConfig.Set("images", "docker.io")
	}

	if rootConfig.Dryrun {
		rootConfig.Info.log("Dry Run - Skip writing config file ", defaultConfigFile)
	} else {
		rootConfig.Debug.log("Writing config file ", defaultConfigFile)
		if err := rootConfig.CliConfig.WriteConfig(); err != nil {
			return errors.Errorf("Writing default config file %s", err)
		}
	}
	return nil
}

func downloadIndex(log *LoggingConfig, url string) (*RepoIndex, error) {
	log.Debug.log("Downloading appsody repository index from ", url)
	indexBuffer := bytes.NewBuffer(nil)
	err := downloadFile(log, url, indexBuffer)
	if err != nil {
		return nil, err
	}

	yamlFile, err := ioutil.ReadAll(indexBuffer)
	if err != nil {
		return nil, fmt.Errorf("Could not read buffer into byte array")
	}
	var index RepoIndex
	err = yaml.Unmarshal(yamlFile, &index)
	if err != nil {
		log.Debug.logf("Contents of downloaded index from %s\n%s", url, yamlFile)
		return nil, fmt.Errorf("Repository index formatting error: %s", err)
	}
	return &index, nil
}

func (index *RepoIndex) listProjects(repoName string, config *RootCommandConfig) (string, error) {
	var Stacks []Stack
	table := uitable.New()
	table.MaxColWidth = 60
	table.Wrap = true
	if strings.Compare(index.APIVersion, supportedIndexAPIVersion) == 1 {
		config.Debug.log("Adding unsupported repository", repoName)
		config.UnsupportedRepos = append(config.UnsupportedRepos, repoName)
	}
	table.AddRow("REPO", "ID", "VERSION  ", "TEMPLATES", "DESCRIPTION")

	Stacks = index.buildStacksFromIndex(repoName, Stacks)

	for _, value := range Stacks {
		templatesListString := convertTemplatesArrayToString(value.Templates)
		table.AddRow(value.repoName, value.ID, value.Version, templatesListString, value.Description)
	}
	return table.String(), nil
}

func (r *RepositoryFile) listRepoProjects(repoName string, config *RootCommandConfig) (string, error) {
	if repo := r.GetRepo(repoName); repo != nil {
		url := repo.URL
		index, err := downloadIndex(config.LoggingConfig, url)
		if err != nil {
			return "", err
		}
		tableString, err := index.listProjects(repoName, config)
		if err != nil {
			return "", err
		}
		return tableString, nil
	}
	return "", errors.New("cannot locate repository named " + repoName)
}

func (r *RepositoryFile) getRepoFile(rootConfig *RootCommandConfig) (*RepositoryFile, error) {
	var repoFileLocation = getRepoFileLocation(rootConfig)
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

func (r *RepositoryFile) listRepos(rootConfig *RootCommandConfig) (string, error) {
	var entries = []RepositoryEntry{}
	table := uitable.New()
	table.MaxColWidth = 1024
	table.AddRow("NAME", "URL")
	for _, value := range r.Repositories {
		repoName := value.Name
		defaultRepoName, err := r.GetDefaultRepoName(rootConfig)
		if err != nil {
			return "", err
		}
		if repoName == defaultRepoName {
			repoName = "*" + repoName
		}
		entries = append(entries, RepositoryEntry{repoName, value.URL, value.IsDefault})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	for _, value := range entries {
		table.AddRow(value.Name, value.URL)
	}

	return table.String(), nil
}

// NewRepoFile - Creates new RepositoryFile
func NewRepoFile() *RepositoryFile {
	return &RepositoryFile{
		APIVersion:   APIVersionV1,
		Generated:    time.Now(),
		Repositories: []*RepositoryEntry{},
	}
}

//AddRepo - Adds RepositoryEntry to RepositoryFile
func (r *RepositoryFile) AddRepo(re ...*RepositoryEntry) {
	r.Repositories = append(r.Repositories, re...)
}

// HasRepo - Check if RepositoryFile has specified Repository
func (r *RepositoryFile) HasRepo(name string) bool {

	for _, rf := range r.Repositories {
		if rf.Name == name {
			return true
		}
	}
	return false
}

//GetRepo - Get RepoEntry from RepositoryFile, return nil if not found
func (r *RepositoryFile) GetRepo(name string) *RepositoryEntry {
	for _, rf := range r.Repositories {
		if rf.Name == name {
			return rf
		}
	}
	return nil
}

//HasURL - Check if RepositoryFile has specified URL
func (r *RepositoryFile) HasURL(url string) bool {
	for _, rf := range r.Repositories {
		if rf.URL == url {
			return true
		}
	}
	return false
}

//GetDefaultRepoName - Get default Repository name from RepositoryFile, return nil if not found
func (r *RepositoryFile) GetDefaultRepoName(rootConfig *RootCommandConfig) (string, error) {
	// Check if there are any repos first
	if len(r.Repositories) < 1 {
		return "", errors.New("your $HOME/.appsody/repository/repository.yaml contains no repositories")
	}
	for _, rf := range r.Repositories {
		if rf.IsDefault {
			return rf.Name, nil
		}
	}
	// If we got this far, no default repo was found - this is likely to be a 0.2.8 or prior
	// We'll set the default repo
	// If there's only one repo - set it as default
	// And if incubator isn't there set the first one as default
	var repoName string
	if len(r.Repositories) == 1 || !r.HasRepo("incubator") {
		r.Repositories[0].IsDefault = true
		repoName = r.Repositories[0].Name
	} else {
		// If there's more than one, let's search for incubator first
		repo := r.GetRepo("incubator")
		repo.IsDefault = true
		repoName = repo.Name
	}
	if err := r.WriteFile(getRepoFileLocation(rootConfig)); err != nil {
		return "", err
	}
	rootConfig.Info.log("Your default repository is now set to ", repoName)
	return repoName, nil
}

//RemoveRepo - Removes specified Repository from RepositoryFile
func (r *RepositoryFile) RemoveRepo(name string, log *LoggingConfig) {
	for index, rf := range r.Repositories {
		if rf.Name == name {
			r.Repositories[index] = r.Repositories[0]
			r.Repositories = r.Repositories[1:]
			log.Info.logf("The %v repository has been removed from your configured list of repositories.", name)
			return
		}
	}
}

//SetDefaultRepoName - Sets default repository to specified repo name
func (r *RepositoryFile) SetDefaultRepoName(name string, defaultRepoName string, rootConfig *RootCommandConfig) (string, error) {
	var repoName string
	for index, rf := range r.Repositories {
		//set current default repo to false
		if rf.Name == defaultRepoName {
			r.Repositories[index].IsDefault = false
		}
		//set new default repo
		if rf.Name == name {
			r.Repositories[index].IsDefault = true
			repoName = rf.Name
		}
	}
	if err := r.WriteFile(getRepoFileLocation(rootConfig)); err != nil {
		return "", err
	}
	rootConfig.Info.log("Your default repository is now set to ", repoName)
	return repoName, nil
}

//WriteFile - Writes file to RepositoryFile
func (r *RepositoryFile) WriteFile(path string) error {
	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

//GetIndices - Gets the indices of a RepositoryFile
func (r *RepositoryFile) GetIndices(log *LoggingConfig) (RepoIndices, error) {
	indices := make(map[string]*RepoIndex)
	brokenRepos := make([]indexError, 0)
	for _, rf := range r.Repositories {
		var index, err = downloadIndex(log, rf.URL)
		if err != nil {
			repoErr := indexError{rf.Name, err}
			brokenRepos = append(brokenRepos, repoErr)
		} else {
			indices[rf.Name] = index
		}
	}
	if len(brokenRepos) > 0 {
		return indices, &indexErrors{brokenRepos}
	}
	return indices, nil
}

func convertTemplatesArrayToString(Templates []Template) string {
	templatesListString := ""
	if len(Templates) > 0 {
		sort.Slice(Templates, func(i, j int) bool { return Templates[i].ID < Templates[j].ID })

		for _, template := range Templates {
			defaultMarker := ""
			if template.IsDefault {
				defaultMarker = "*"
			}
			if templatesListString != "" {
				templatesListString += ", " + defaultMarker + template.ID
			} else {
				templatesListString = defaultMarker + template.ID
			}

		}
	}

	return templatesListString
}

func setDefaultTemplate(Templates []Template, DefaultTemplate string) {
	for index, template := range Templates {
		if template.ID == DefaultTemplate {
			Templates[index].IsDefault = true
		}
	}
}
func (index *RepoIndex) buildStacksFromIndex(repoName string, Stacks []Stack) []Stack {

	for id, value := range index.Projects {
		setDefaultTemplate(value[0].Templates[:], value[0].DefaultTemplate)
		Stacks = append(Stacks, Stack{repoName, id, value[0].Version, value[0].Description, value[0].Templates})
	}
	for _, value := range index.Stacks {
		setDefaultTemplate(value.Templates[:], value.DefaultTemplate)
		Stacks = append(Stacks, Stack{repoName, value.ID, value.Version, value.Description, value.Templates})
	}

	sort.Slice(Stacks, func(i, j int) bool {
		if Stacks[i].repoName < Stacks[j].repoName {
			return true
		}
		if Stacks[i].repoName == Stacks[j].repoName && Stacks[i].ID < Stacks[j].ID {
			return true

		}
		return false
	})

	return Stacks
}

func (r *RepositoryFile) listProjects(config *RootCommandConfig) (string, error) {
	var Stacks []Stack
	table := uitable.New()
	table.MaxColWidth = 60
	table.Wrap = true

	table.AddRow("REPO", "ID", "VERSION  ", "TEMPLATES", "DESCRIPTION")
	indices, err := r.GetIndices(config.LoggingConfig)

	if err != nil {
		config.Error.logf("The following indices could not be read, skipping:\n%v", err)
	}
	if len(indices) != 0 {
		for repoName, index := range indices {

			if strings.Compare(index.APIVersion, supportedIndexAPIVersion) == 1 {
				config.Debug.log("Adding unsupported repository", repoName)
				config.UnsupportedRepos = append(config.UnsupportedRepos, repoName)
			}

			Stacks = index.buildStacksFromIndex(repoName, Stacks)

		}

	} else {
		return "", errors.New("there are no repositories in your configuration")
	}

	defaultRepoName, err := r.GetDefaultRepoName(config)
	if err != nil {
		return "", err
	}
	for _, value := range Stacks {

		if value.repoName == defaultRepoName {
			value.repoName = "*" + value.repoName
		}

		templatesListString := convertTemplatesArrayToString(value.Templates)
		table.AddRow(value.repoName, value.ID, value.Version, templatesListString, value.Description)
	}
	return table.String(), nil
}

//IndexOutputFormat - Type for outputting repositories of the RepositoryFile in JSON and YAML
type IndexOutputFormat struct {
	APIVersion   string                   `yaml:"apiVersion" json:"apiVersion"`
	Generated    time.Time                `yaml:"generated" json:"generated"`
	Repositories []RepositoryOutputFormat `yaml:"repositories" json:"repositories"`
}

//RepositoryOutputFormat - Type for outputting stacks of a repository in JSON and YAML
type RepositoryOutputFormat struct {
	Name   string  `yaml:"repositoryName" json:"repositoryName"`
	Stacks []Stack `yaml:"stacks" json:"stacks"`
}

func (r *RepositoryFile) getRepositories(log *LoggingConfig) (IndexOutputFormat, error) {
	var indexOutput IndexOutputFormat
	indexOutput.APIVersion = r.APIVersion
	indexOutput.Generated = r.Generated
	indices, err := r.GetIndices(log)
	if err != nil {
		return indexOutput, errors.Errorf("Could not read indices: %v", err)
	}

	if len(indices) != 0 {
		for repoName, index := range indices {
			var Stacks []Stack
			Stacks = index.buildStacksFromIndex(repoName, Stacks)

			indexOutput.Repositories = append(indexOutput.Repositories, RepositoryOutputFormat{Name: repoName, Stacks: Stacks})
		}
	}
	return indexOutput, nil
}

func (r *RepositoryFile) getRepository(log *LoggingConfig, repoName string) (IndexOutputFormat, error) {
	var indexOutput IndexOutputFormat
	indexOutput.APIVersion = r.APIVersion
	indexOutput.Generated = r.Generated
	indices, err := r.GetIndices(log)
	if err != nil {
		return indexOutput, errors.Errorf("Could not read indices: %v", err)
	}

	if indices[repoName] != nil {
		var Stacks []Stack
		Stacks = indices[repoName].buildStacksFromIndex(repoName, Stacks)

		indexOutput.Repositories = append(indexOutput.Repositories, RepositoryOutputFormat{Name: repoName, Stacks: Stacks})
	}
	return indexOutput, nil
}
