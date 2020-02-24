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
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"unicode"

	"github.com/andrew-d/isbinary"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

type packageCommandConfig struct {
	*RootCommandConfig
	dockerBuildOptions  string
	buildahBuildOptions string
}

// structs for parsing the yaml files
type StackYaml struct {
	Name            string `yaml:"name"`
	Version         string `yaml:"version"`
	Description     string `yaml:"description"`
	License         string `yaml:"license"`
	Language        string `yaml:"language"`
	Maintainers     []Maintainer
	DefaultTemplate string            `yaml:"default-template"`
	TemplatingData  map[string]string `yaml:"templating-data"`
	Requirements    StackRequirement  `yaml:"requirements,omitempty"`
}
type Maintainer struct {
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	GithubID string `yaml:"github-id" mapstructure:"github-id"`
}

type IndexYaml struct {
	APIVersion string `yaml:"apiVersion"`
	Stacks     []IndexYamlStack
}
type IndexYamlStack struct {
	ID              string `yaml:"id"`
	Name            string `yaml:"name"`
	Version         string `yaml:"version"`
	Description     string `yaml:"description"`
	License         string `yaml:"license"`
	Language        string `yaml:"language"`
	Maintainers     []Maintainer
	DefaultTemplate string `yaml:"default-template"`
	SourceURL       string `yaml:"src"`
	Templates       []IndexYamlStackTemplate
	Requirements    StackRequirement `yaml:"requirements,omitempty"`
	Image           string           `yaml:"image"`
}
type IndexYamlStackTemplate struct {
	ID  string `yaml:"id"`
	URL string `yaml:"url"`
}

// struct to convert yaml to json files
type IndexJSONStack struct {
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	Language     string `json:"language"`
	ProjectType  string `json:"projectType"`
	ProjectStyle string `json:"projectStyle"`
	Location     string `json:"location"`
	Links
}

type Links struct {
	Self string `json:"self"`
}

func newStackPackageCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &packageCommandConfig{RootCommandConfig: rootConfig}

	// stack package is a tool for local stack developers to package their stack
	// the stack package command does the following...
	// 1. create/update a local index yaml
	// 2. create a tar for each stack template
	// 3. build a docker image
	// 4. create/update an appsody repo for the stack

	var imageNamespace string
	var imageRegistry string
	var namespaceAndRepo string

	log := rootConfig.LoggingConfig

	var stackPackageCmd = &cobra.Command{
		Use:   "package",
		Short: "Package your stack.",
		Long: `Package your stack in a local Appsody development environment. You must run this command from the root directory of your stack.

The packaging process builds the stack image, generates the "tar.gz" archive files for each template, and adds your stack to the "dev.local" repository in your Appsody configuration. You can see the list of your packaged stacks by running 'appsody list dev.local'.`,
		Example: `  appsody stack package
  Packages the stack in the current directory, tags the built image with the default registry and namespace, and adds the stack to the "dev.local" repository.
  
  appsody stack package --image-namespace my-namespace
  Packages the stack in the current directory, tags the built image with the default registry and "my-namespace" namespace, and adds the stack to the "dev.local" repository.
  
  appsody stack package --buildah --buildah-options "--format=docker"
  Packages the stack in the current directory, builds project using buildah primitives in Docker format, tags the built image with the default registry and namespace, and adds the stack to the "dev.local" repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("Unexpected argument. Use 'appsody [command] --help' for more information about a command")
			}

			log.Info.Log("******************************************")
			log.Info.Log("Running appsody stack package")
			log.Info.Log("******************************************")
			buildOptions := ""
			if config.buildahBuildOptions != "" {
				if !config.Buildah {
					return errors.New("Cannot specify --buildah-options flag without --buildah")
				}
				buildOptions = strings.TrimSpace(config.buildahBuildOptions)
			}

			if config.dockerBuildOptions != "" {
				if config.Buildah {
					return errors.New("Cannot specify --docker-options flag with --buildah")
				}
				buildOptions = strings.TrimSpace(config.dockerBuildOptions)
			}

			projectPath := rootConfig.ProjectDir

			// get the stack name from the stack path
			stackID := filepath.Base(projectPath)
			log.Debug.Log("stackID is: ", stackID)

			// sets stack path to be the copied folder
			stackPath := filepath.Join(getHome(rootConfig), "stacks", "packaging-"+stackID)
			log.Debug.Log("stackPath is: ", stackPath)

			// creates stacks dir if it doesn't exist
			err := os.MkdirAll(filepath.Dir(stackPath), 0777)

			if err != nil {
				return errors.Errorf("Error creating stacks directory: %v", err)
			}

			err = RemoveIfExists(stackPath)
			if err != nil {
				return errors.Errorf("Error removing packaging folder: %v", err)
			}

			// defer function to remove packaging folder created for stack variables
			defer func() {
				err = RemoveIfExists(stackPath)
				if err != nil {
					log.Info.logf("Error removing packaging folder: %v", err)
				}
			}()

			// make a copy of the folder to apply template to
			err = CopyDir(log, projectPath, stackPath)
			if err != nil {
				return errors.Errorf("Error trying to copy directory: %v", err)
			}

			// get the necessary data from the current stack.yaml
			var stackYaml StackYaml

			// get the necessary data from the current stack.yaml
			stackYaml, err = getStackData(stackPath)
			if err != nil {
				return errors.Errorf("Error parsing the stack.yaml file: %v", err)
			}

			// check for templates dir, error out if its not there
			check, err := Exists(filepath.Join(projectPath, "templates"))
			if err != nil {
				return errors.New("Error checking stack root directory: " + err.Error())
			}
			if !check {
				// if we can't find the templates directory then we are not starting from a valid root of the stack directory
				return errors.New("Unable to reach templates directory. Current directory must be the root of the stack")
			}

			appsodyHome := getHome(rootConfig)
			log.Debug.Log("appsodyHome is:", appsodyHome)

			devLocal := filepath.Join(appsodyHome, "stacks", "dev.local")
			log.Debug.Log("devLocal is: ", devLocal)

			// create the devLocal directory in appsody home
			err = os.MkdirAll(devLocal, os.FileMode(0755))
			if err != nil {
				return errors.Errorf("Error creating directory: %v", err)
			}

			indexFileLocal := filepath.Join(devLocal, "dev.local-index.yaml")
			log.Debug.Log("indexFileLocal is: ", indexFileLocal)

			// create IndexYaml struct and populate the APIVersion and Stacks header
			var indexYaml IndexYaml

			// check for existing index yaml file
			check, err = Exists(indexFileLocal)
			if err != nil {
				return errors.New("Error checking index file: " + err.Error())
			}
			if check {
				// index file exists already so see if it contains the stack data and remove it if found
				log.Debug.Log("Index file exists already")

				source, err := ioutil.ReadFile(indexFileLocal)
				if err != nil {
					return errors.Errorf("Error trying to read: %v", err)
				}

				err = yaml.Unmarshal(source, &indexYaml)
				if err != nil {
					return errors.Errorf("Error trying to unmarshall: %v", err)
				}
				indexYaml, _ = findStackAndRemove(log, stackID, indexYaml)

			} else {
				// create the beginning of the index yaml
				indexYaml = IndexYaml{}
				indexYaml.APIVersion = "v2"
				indexYaml.Stacks = make([]IndexYamlStack, 0, 1)
			}

			// docker build
			// create the image name to be used for the docker image
			namespaceAndRepo = imageRegistry + "/" + imageNamespace + "/" + stackID

			buildImage := namespaceAndRepo + ":" + stackYaml.Version

			imageDir := filepath.Join(stackPath, "image")
			log.Debug.Log("imageDir is: ", imageDir)

			dockerFile := filepath.Join(imageDir, "Dockerfile-stack")
			log.Debug.Log("dockerFile is: ", dockerFile)

			labels, err := GetLabelsForStackImage(stackID, buildImage, stackYaml, rootConfig)
			if err != nil {
				return err
			}

			// create the template metadata
			templateMetadata, err := CreateTemplateMap(labels, stackYaml, imageNamespace, imageRegistry)
			if err != nil {
				return errors.Errorf("Error creating templating mal: %v", err)
			}

			// apply templating to stack
			err = ApplyTemplating(stackPath, templateMetadata)
			if err != nil {
				return errors.Errorf("Error applying templating: %v", err)
			}

			// tag with the full version then majorminor, major, and latest
			cmdArgs := []string{"-t", buildImage}
			semver := templateMetadata["semver"].(map[string]string)
			cmdArgs = append(cmdArgs, "-t", namespaceAndRepo+":"+semver["majorminor"])
			cmdArgs = append(cmdArgs, "-t", namespaceAndRepo+":"+semver["major"])
			cmdArgs = append(cmdArgs, "-t", namespaceAndRepo)

			if buildOptions != "" {
				options := strings.Split(buildOptions, " ")
				err := checkBuildOptions(options)
				if err != nil {
					return err
				}
				cmdArgs = append(cmdArgs, options...)
			}

			labelPairs := CreateLabelPairs(labels)
			for _, label := range labelPairs {
				cmdArgs = append(cmdArgs, "--label", label)
			}

			cmdArgs = append(cmdArgs, "-f", dockerFile, imageDir)
			log.Debug.Log("cmdArgs is: ", cmdArgs)

			if !config.Buildah {
				log.Info.Log("Running docker build")
				err = DockerBuild(config.RootCommandConfig, cmdArgs, config.DockerLog)
			} else {
				log.Info.Log("Running buildah build")
				err = BuildahBuild(config.RootCommandConfig, cmdArgs, config.BuildahLog)
			}

			if err != nil {
				return err
			}

			stackImage := namespaceAndRepo + ":" + semver["majorminor"] + "." + semver["patch"]
			// build up stack struct for the new stack
			newStackStruct := initialiseStackData(stackID, stackImage, stackYaml)

			// get project directory
			sourceDir := projectPath
			log.Debug.Log("sourceDir is: ", sourceDir)

			// create name for the source tar file
			versionedArchive := filepath.Join(devLocal, stackID+".v"+stackYaml.Version+".")
			log.Debug.Log("versionedArchive is: ", versionedArchive)

			versionArchiveTar := versionedArchive + "source.tar.gz"
			log.Debug.Log("versionedArdhiveTar is: ", versionArchiveTar)

			// tar the files
			log.Info.Log("Creating tar for: " + stackID + " source")
			err = Targz(log, sourceDir, versionedArchive, "source")
			if err != nil {
				return errors.Errorf("Error trying to tar: %v", err)
			}

			if runtime.GOOS == "windows" {
				// for windows, add a leading slash and convert to unix style slashes
				versionArchiveTar = "/" + filepath.ToSlash(versionArchiveTar)
			}
			versionArchiveTar = "file://" + versionArchiveTar

			newStackStruct.SourceURL = versionArchiveTar

			// find and open the template path so we can loop through the templates
			templatePath := filepath.Join(stackPath, "templates")

			t, err := os.Open(templatePath)
			if err != nil {
				return errors.Errorf("Error opening directory: %v", err)
			}

			templates, err := t.Readdirnames(0)
			if err != nil {
				return errors.Errorf("Error reading directories: %v", err)
			}

			// loop through the template directories and create the id and url
			for i := range templates {
				log.Debug.Log("template is: ", templates[i])
				if strings.Contains(templates[i], ".DS_Store") {
					log.Debug.Log("Ignoring .DS_Store")
					continue
				}

				sourceDir := filepath.Join(stackPath, "templates", templates[i])
				log.Debug.Log("sourceDir is: ", sourceDir)

				// create name for the tar files
				versionedArchive := filepath.Join(devLocal, stackID+".v"+stackYaml.Version+".templates.")
				log.Debug.Log("versionedArchive is: ", versionedArchive)

				versionArchiveTar := versionedArchive + templates[i] + ".tar.gz"
				log.Debug.Log("versionedArdhiveTar is: ", versionArchiveTar)

				if runtime.GOOS == "windows" {
					// for windows, add a leading slash and convert to unix style slashes
					versionArchiveTar = "/" + filepath.ToSlash(versionArchiveTar)
				}
				versionArchiveTar = "file://" + versionArchiveTar

				// add the template data to the struct
				newTemplateStruct := IndexYamlStackTemplate{}
				newTemplateStruct.ID = templates[i]
				newTemplateStruct.URL = versionArchiveTar

				newStackStruct.Templates = append(newStackStruct.Templates, newTemplateStruct)

				// create a config yaml file for the tarball
				configYaml := filepath.Join(templatePath, templates[i], ".appsody-config.yaml")
				log.Debug.Log("configYaml is: ", configYaml)

				g, err := os.Create(configYaml)
				if err != nil {
					return errors.Errorf("Error trying to create file: %v", err)
				}

				// Only use major.minor version here
				_, err = g.WriteString("stack: " + namespaceAndRepo + ":" + semver["majorminor"])
				if err != nil {
					return errors.Errorf("Error trying to write: %v", err)
				}

				g.Close()

				// tar the files
				log.Info.Log("Creating tar for: " + templates[i])
				err = Targz(log, sourceDir, versionedArchive, templates[i])
				if err != nil {
					return errors.Errorf("Error trying to tar: %v", err)
				}

				// remove the config yaml file
				err = os.Remove(configYaml)
				if err != nil {
					return errors.Errorf("Error trying to remove file: %v", err)
				}
			}

			t.Close()

			// add the new stack struct to the existing struct
			indexYaml.Stacks = append(indexYaml.Stacks, newStackStruct)

			// write yaml data to the index yaml
			source, err := yaml.Marshal(&indexYaml)
			if err != nil {
				return errors.Errorf("Error trying to marshall: %v", err)
			}

			log.Info.Log("Writing: " + indexFileLocal)
			err = ioutil.WriteFile(indexFileLocal, source, 0644)
			if err != nil {
				return errors.Errorf("Error trying to read: %v", err)
			}

			// list repos
			var repoFile RepositoryFile
			repos, repoErr := repoFile.getRepos(rootConfig)
			if repoErr != nil {
				return repoErr
			}
			// See if a configured repo already points to dev.local, if so remove it
			repoName := "dev.local"

			repo := repos.GetRepo(repoName)
			if repo == nil || !strings.Contains(repo.URL, indexFileLocal) {
				// the repo is setup wrong, delete and recreate it
				if repo != nil {
					log.Info.logf("Appsody repo %s is configured with the wrong URL. Deleting and recreating it.", repoName)
					repos.Remove(repoName, rootConfig.LoggingConfig)
				}
				// check for a different repo with the same file url
				var repoNameToDelete string
				for _, repo := range repos.Repositories {
					if strings.Contains(repo.URL, indexFileLocal) {
						repoNameToDelete = repo.Name
						break
					}
				}
				if repoNameToDelete != "" {
					log.Info.logf("Appsody repo %s is configured with %s's URL. Deleting it to setup %s.", repoNameToDelete, repoName, repoName)
					repos.Remove(repoNameToDelete, rootConfig.LoggingConfig)
				}
				err = repos.WriteFile(getRepoFileLocation(rootConfig))
				if err != nil {
					return errors.Errorf("Error writing to repo file %s. %v", getRepoFileLocation(rootConfig), err)
				}
				log.Info.Logf("Creating %s repository", repoName)

				_, err = AddLocalFileRepo(repoName, indexFileLocal, rootConfig)
				if err != nil {
					return errors.Errorf("Error adding local repository. Your stack may not be available to appsody commands. %v", err)
				}
			}
			err = generateCodewindJSON(log, indexYaml, indexFileLocal, "Local")
			if err != nil {
				return errors.Errorf("Could not generate json file from yaml index: %v", err)
			}

			log.Info.log("Your local stack is available as part of repo ", repoName)

			return nil
		},
	}

	stackPackageCmd.PersistentFlags().StringVar(&imageNamespace, "image-namespace", "appsody", "Namespace used for creating the images.")
	stackPackageCmd.PersistentFlags().StringVar(&imageRegistry, "image-registry", "dev.local", "Registry used for creating the images.")
	stackPackageCmd.PersistentFlags().BoolVar(&rootConfig.Buildah, "buildah", false, "Build project using buildah primitives instead of Docker.")
	stackPackageCmd.PersistentFlags().StringVar(&config.dockerBuildOptions, "docker-options", "", "Specify the Docker build options to use. Value must be in \"\". The following Docker options are not supported: '--help','-t','--tag','-f','--file'.")
	stackPackageCmd.PersistentFlags().StringVar(&config.buildahBuildOptions, "buildah-options", "", "Specify the buildah build options to use. Value must be in \"\".")

	return stackPackageCmd
}

func initialiseStackData(stackID string, stackImage string, stackYaml StackYaml) IndexYamlStack {
	// build up stack struct for the new stack
	newStackStruct := IndexYamlStack{}
	// set the data in the new stack struct
	newStackStruct.ID = stackID
	newStackStruct.Name = stackYaml.Name
	newStackStruct.Version = stackYaml.Version
	newStackStruct.Description = stackYaml.Description
	newStackStruct.License = stackYaml.License
	newStackStruct.Language = stackYaml.Language
	newStackStruct.Maintainers = append(newStackStruct.Maintainers, stackYaml.Maintainers...)
	newStackStruct.DefaultTemplate = stackYaml.DefaultTemplate
	newStackStruct.Requirements = stackYaml.Requirements
	newStackStruct.Image = stackImage

	return newStackStruct
}

func findStackAndRemove(log *LoggingConfig, stackID string, indexYaml IndexYaml) (IndexYaml, bool) {
	// find the index of the stack
	stackExists := false
	foundStack := -1
	for i, stack := range indexYaml.Stacks {
		if stack.ID == stackID {
			log.Debug.Log("Existing stack: '" + stackID + "' found")
			foundStack = i
			break
		}
	}

	// delete index foundStack from indexYaml.Stacks as we will append the new stack later
	if foundStack != -1 {
		stackExists = true
		indexYaml.Stacks = indexYaml.Stacks[:foundStack+copy(indexYaml.Stacks[foundStack:], indexYaml.Stacks[foundStack+1:])]
	}
	return indexYaml, stackExists
}

// GetLabelsForStackImage - Gets labels associated with the stack image
func GetLabelsForStackImage(stackID string, buildImage string, stackYaml StackYaml, config *RootCommandConfig) (map[string]string, error) {
	var labels = make(map[string]string)

	gitLabels, err := getGitLabels(config)
	if err != nil {
		config.Warning.log("Not all labels will be set. ", err.Error())
	} else {
		if branchURL, ok := gitLabels[ociKeyPrefix+"source"]; ok {
			if contextDir, ok := gitLabels[appsodyImageCommitKeyPrefix+"contextDir"]; ok {
				branchURL += contextDir
				gitLabels[ociKeyPrefix+"url"] = branchURL
			}
			// These are enforced by the stack lint so they should exist
			gitLabels[ociKeyPrefix+"documentation"] = branchURL + "/README.md"
			gitLabels[ociKeyPrefix+"source"] = branchURL + "/image"
		}

		for key, value := range gitLabels {
			labels[key] = value
		}
	}

	// build a ProjectConfig struct from the stackyaml so we can reuse getConfigLabels() func
	projectConfig := ProjectConfig{
		ProjectName: stackYaml.Name,
		Version:     stackYaml.Version,
		Description: stackYaml.Description,
		License:     stackYaml.License,
		Maintainers: stackYaml.Maintainers,
	}
	configLabels, err := getConfigLabels(projectConfig, "stack.yaml", config.LoggingConfig)
	if err != nil {
		return labels, err
	}
	configLabels[appsodyStackKeyPrefix+"id"] = stackID
	configLabels[appsodyStackKeyPrefix+"tag"] = buildImage

	for key, value := range configLabels {
		labels[key] = value
	}

	return labels, nil
}

// CreateTemplateMap - uses the git labels, stack.yaml, stackID and imageNamespace to create a map
// with all the necessary data needed for the template
func CreateTemplateMap(labels map[string]string, stackYaml StackYaml, imageNamespace string, imageRegistry string) (map[string]interface{}, error) {

	// create stack variables and add to templateMetadata map
	var templateMetadata = make(map[string]interface{})

	// split version number into major, minor and patch strings
	var err error

	versionLabel := labels[ociKeyPrefix+"version"]
	versionFull := strings.Split(versionLabel, ".")

	if len(versionFull) != 3 {
		err = errors.Errorf("Version format incorrect")
		return templateMetadata, err
	}

	// create map that holds stack variables
	templateMetadata["id"] = labels[appsodyStackKeyPrefix+"id"]
	templateMetadata["name"] = labels[ociKeyPrefix+"title"]
	templateMetadata["version"] = versionLabel
	templateMetadata["description"] = labels[ociKeyPrefix+"description"]
	templateMetadata["created"] = labels[ociKeyPrefix+"created"]
	templateMetadata["tag"] = labels[appsodyStackKeyPrefix+"tag"]
	templateMetadata["maintainers"] = labels[ociKeyPrefix+"authors"]

	// create version map and add to templateMetadata map
	var semver = make(map[string]string)
	semver["major"] = versionFull[0]
	semver["minor"] = versionFull[1]
	semver["patch"] = versionFull[2]
	semver["majorminor"] = strings.Join(versionFull[0:2], ".")
	templateMetadata["semver"] = semver

	// create image map add to templateMetadata map
	var image = make(map[string]string)
	image["namespace"] = imageNamespace
	image["registry"] = imageRegistry
	templateMetadata["image"] = image

	// loop through user variables and add them to map, must begin with alphanumeric character
	for key, value := range stackYaml.TemplatingData {

		// validates that key starts with alphanumeric character
		runes := []rune(key)
		firstRune := runes[0]
		if unicode.IsLetter(firstRune) || unicode.IsNumber(firstRune) {
			templateMetadata[key] = value
		} else {
			return templateMetadata, errors.Errorf("Variable name didn't start with alphanumeric character")
		}
	}
	return templateMetadata, err

}

// ApplyTemplating -  walks through the copied folder directory and applies a template using the
// previously created templateMetada to all files in the target directory
func ApplyTemplating(stackPath string, templateMetadata interface{}) error {

	err := filepath.Walk(stackPath, func(path string, info os.FileInfo, err error) error {

		//Skip .git folder
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		} else if !info.IsDir() {

			//get file name
			file := filepath.Base(path)

			// get permission of file
			permission := info.Mode()

			binaryFile, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Errorf("Error reading file for binary test: %v", err)
			}

			// skip binary files
			if isbinary.Test(binaryFile) {
				return nil
			}

			// set file permission to writable to apply templating
			err = os.Chmod(path, 0666)
			if err != nil {
				return errors.Errorf("Error changing file permision: %v", err)
			}

			// create new template from parsing file
			tmpl, err := template.New(file).Delims("{{.stack", "}}").ParseFiles(path)
			if err != nil {
				return errors.Errorf("Error creating new template from file: %v", err)
			}

			// open file at path
			f, err := os.Create(path)
			if err != nil {
				return errors.Errorf("Error opening file: %v", err)
			}

			// apply template to file
			err = tmpl.ExecuteTemplate(f, file, templateMetadata)
			if err != nil {
				return errors.Errorf("Error executing template: %v", err)
			}

			// set old file permission to new file
			err = os.Chmod(path, permission)
			if err != nil {
				return errors.Errorf("Error reverting file permision: %v", err)
			}
			f.Close()
		}
		return nil
	})

	if err != nil {
		return errors.Errorf("Error walking through directory: %v", err)
	}

	return nil

}
