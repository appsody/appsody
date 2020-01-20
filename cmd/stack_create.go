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
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type stackCreateCommandConfig struct {
	*RootCommandConfig
	copy string
}

func newStackCreateCmd(rootConfig *RootCommandConfig) *cobra.Command {
	config := &stackCreateCommandConfig{RootCommandConfig: rootConfig}

	var stackCmd = &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Appsody stack.",
		Long: `Create a new Appsody stack, called <name>, in the current directory. You can use this stack as a starting point for developing your own Appsody stack.

By default, the new stack is based on the example stack: incubator/starter. If you want to use a different stack as the basis for your new stack, use the copy flag to specify the stack you want to use as the starting point. You can use 'appsody list' to see the available stacks.

The stack name must start with a lowercase letter, and can contain only lowercase letters, numbers, or dashes, and cannot end with a dash. The stack name cannot exceed 128 characters.`,
		Example: `  appsody stack create my-stack  
  Creates a stack called my-stack, based on the example stack “incubator/starter”.

  appsody stack create my-stack --copy incubator/nodejs-express  
  Creates a stack called my-stack, based on the Node.js Express stack.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentTime := time.Now().Format("20060102150405")

			if len(args) < 1 {
				return errors.New("Required parameter missing. You must specify a stack name")
			}

			stack := args[0]

			match, err := IsValidProjectName(stack)
			if !match {
				return err
			}

			exists, err := Exists(stack)
			if err != nil {
				return err
			}

			if exists {
				return errors.New("A stack named " + stack + " already exists in your directory. Specify a unique stack name")
			}

			extractFolderExists, err := Exists(filepath.Join(getHome(rootConfig), "extract"))
			if err != nil {
				return err
			}

			if !extractFolderExists {
				err = os.MkdirAll(filepath.Join(getHome(rootConfig), "extract"), os.ModePerm)
				if err != nil {
					return err
				}
			}

			repoID, stackID, err := parseProjectParm(config.copy, config.RootCommandConfig)
			if err != nil {
				return err
			}

			extractFilename := stackID + ".tar.gz"
			extractDir := filepath.Join(getHome(rootConfig), "extract")
			extractDirFile := filepath.Join(extractDir, extractFilename)

			// Get Repository directory and umarshal
			repoDir := getRepoDir(rootConfig)
			var repoFile RepositoryFile
			source, err := ioutil.ReadFile(filepath.Join(repoDir, "repository.yaml"))
			if err != nil {
				return errors.Errorf("Error trying to read: %v", err)
			}

			err = yaml.Unmarshal(source, &repoFile)
			if err != nil {
				return errors.Errorf("Error parsing the repository.yaml file: %v", err)
			}

			// get specificed repo and umarshal
			repoInfo := repoFile.GetRepo(repoID)

			// error if repo not found in repository.yaml
			if repoInfo == nil {
				//return errors.Errorf("Repository: %s not found in repository.yaml file", repoID)
				return oldCreateMethod(rootConfig.LoggingConfig, rootConfig, config, stack, currentTime)
			}
			repoIndexURL := repoInfo.URL
			var repoIndex IndexYaml
			tempRepoFile := filepath.Join(extractDir, "repo.yaml")
			err = downloadFileToDisk(rootConfig.LoggingConfig, repoIndexURL, tempRepoFile, config.Dryrun)
			if err != nil {
				return err
			}
			defer os.Remove(tempRepoFile)
			sourceIndex, err := ioutil.ReadFile(tempRepoFile)
			if err != nil {
				return errors.Errorf("Error trying to read: %v", err)
			}

			err = yaml.Unmarshal(sourceIndex, &repoIndex)
			if err != nil {
				return errors.Errorf("Error parsing the repository.yaml file: %v", err)
			}

			// get specified stack and get URL
			createStack := getStack(&repoIndex, stackID)
			if createStack == nil {
				return errors.New("Stack not found in index")
			}

			stackSource := createStack.SourceURL

			if stackSource == "" {
				//return errors.New("No source URL specified.  Use the add-to-repo command with the --release-url flag to your repo")
				return oldCreateMethod(rootConfig.LoggingConfig, rootConfig, config, stack, currentTime)
			}

			err = downloadFileToDisk(rootConfig.LoggingConfig, stackSource, extractDirFile, config.Dryrun)
			if err != nil {
				return err
			}

			extractFile, err := os.Open(extractDirFile)
			if err != nil {
				return err
			}

			untarErr := untarSource(rootConfig.LoggingConfig, stack, extractFile, config.Dryrun)
			if untarErr != nil {
				return untarErr
			}

			//deleting the stacks repo zip
			os.Remove(extractDirFile)

			if !config.Dryrun {
				rootConfig.Info.log("Stack created: ", stack)
			} else {
				rootConfig.Info.log("Dry run complete")
			}

			return nil
		},
	}
	stackCmd.PersistentFlags().StringVar(&config.copy, "copy", "incubator/starter", "Copy the specified stack. The format is <repository>/<stack>")
	return stackCmd
}

func getStack(r *IndexYaml, name string) *IndexYamlStack {
	for _, rf := range r.Stacks {
		if rf.ID == name {
			return &rf
		}
	}
	return nil
}

// taken from https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func untarSource(log *LoggingConfig, dst string, r io.Reader, dryrun bool) error {
	if !dryrun {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)

		for {
			header, err := tr.Next()

			switch {

			// if no more files are found return
			case err == io.EOF:
				return nil

			// return any other error
			case err != nil:
				return err

			// if the header is nil, just skip it (not sure how this happens)
			case header == nil:
				continue
			}

			// the target location where the dir/file should be created
			target := filepath.Join(dst, header.Name)

			// the following switch could also be done using fi.Mode(), not sure if there
			// a benefit of using one vs. the other.
			// fi := header.FileInfo()

			// check the file type
			switch header.Typeflag {

			// if its a dir and it doesn't exist create it
			case tar.TypeDir:
				if _, err := os.Stat(target); err != nil {
					if err := os.MkdirAll(target, 0755); err != nil {
						return err
					}
				}

			// if it's a file create it
			case tar.TypeReg:
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return err
				}

				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					return err
				}

				// manually close here after each file operation; defering would cause each file close
				// to wait until all operations have completed.
				f.Close()
			}
		}
	} else {
		log.Info.logf("Dry Run skipping -Untar of filee: %s to destination %s", r, dst)
		return nil
	}
}

// To be removed in next release
func oldCreateMethod(log *LoggingConfig, rootConfig *RootCommandConfig, config *stackCreateCommandConfig, stack string, currentTime string) error {
	err := downloadFileToDisk(rootConfig.LoggingConfig, "https://github.com/appsody/stacks/archive/master.zip", filepath.Join(getHome(rootConfig), "extract", "repo.zip"), config.Dryrun)
	if err != nil {
		return err
	}
	_, stackTempDir, err := parseProjectParm(config.copy, config.RootCommandConfig)
	if err != nil {
		return err
	}

	valid, unzipErr := unzip(rootConfig.LoggingConfig, filepath.Join(getHome(rootConfig), "extract", "repo.zip"), stack, config.copy, config.Dryrun)
	if unzipErr != nil {
		return unzipErr
	}

	if !valid {
		return errors.Errorf("Invalid stack name: " + config.copy + ". Stack name must be in the format <repo>/<stack>")
	}

	//deleting the stacks repo zip
	os.Remove(filepath.Join(getHome(rootConfig), "extract", "repo.zip"))

	//moving out the stack which we need
	if config.Dryrun {
		config.Info.logf("Dry Run -Skipping moving out of stack: %s from %s", stackTempDir, filepath.Join(stack, "stacks-master", config.copy))

	} else {
		stackTempDir = ".temp-" + stackTempDir + "-" + currentTime

		err = os.Rename(filepath.Join(stack, "stacks-master", config.copy), stackTempDir)
		if err != nil {
			return err
		}
	}

	//deleting the folder from which stack is extracted
	os.RemoveAll(stack)

	// rename the stack to the name which user want
	if config.Dryrun {
		config.Info.logf("Dry Run -Skipping renaming of stack from: %s to %s", stackTempDir, stack)

	} else {
		err = os.Rename(stackTempDir, stack)
		if err != nil {
			return err
		}
	}

	if !config.Dryrun {
		rootConfig.Info.log("Stack created: ", stack)
	} else {
		rootConfig.Info.log("Dry run complete")
	}

	return nil
}

// Unzip will decompress a zip archive - TO BE REMOVED NEXT RELEASE
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(log *LoggingConfig, src string, dest string, copy string, dryrun bool) (bool, error) {
	if dryrun {
		log.Info.logf("Dry Run -Skipping unzip of file: %s from %s", copy, src)

	} else {
		valid := false

		r, err := zip.OpenReader(src)
		if err != nil {
			return valid, err
		}
		defer r.Close()

		for _, f := range r.File {

			// Store filename/path for returning and using later on
			fpath := filepath.Join(dest, f.Name)

			// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
			if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
				return valid, errors.Errorf("%s: illegal file path", fpath)
			}

			if runtime.GOOS == "windows" {
				if !strings.HasPrefix(f.Name, "stacks-master/"+copy+"/") {
					continue
				} else {
					valid = true
				}
			} else {
				if !strings.HasPrefix(f.Name, filepath.Join("stacks-master", string(os.PathSeparator), copy)+string(os.PathSeparator)) {
					continue
				} else {
					valid = true
				}
			}

			if f.FileInfo().IsDir() {
				// Make Folder
				err := os.MkdirAll(fpath, os.ModePerm)
				if err != nil {
					return valid, err
				}
				continue
			}

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return valid, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return valid, err
			}

			rc, err := f.Open()
			if err != nil {
				return valid, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()
			rc.Close()

			if err != nil {
				return valid, err
			}
		}
		return valid, nil
	}
	return true, nil
}
