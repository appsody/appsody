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
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newStackCreateCmd(rootConfig *RootCommandConfig) *cobra.Command {
	var stackCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new stack as a copy of an existing stack",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {

			downloadFolderToDisk("https://github.com/appsody/stacks/archive/master.zip", getHome(rootConfig)+"/extract/hello.zip")

			unzip(getHome(rootConfig)+"/extract/hello.zip", "output-folder")
			os.Remove(getHome(rootConfig) + "/extract/hello.zip")

			err := os.Rename("output-folder/stacks-master/incubator/nodejs/", "nodejs")
			if err != nil {
				log.Fatal(err)
			}

			os.RemoveAll("output-folder")

			err = os.Rename("nodejs", "output-folder")
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}
	return stackCmd
}

func downloadFolderToDisk(url string, destFile string) error {
	outFile, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = downloadFile(url, outFile)
	if err != nil {
		return err
	}
	return nil
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		fileName := strings.Replace(f.Name, "/stacks-master", "", -1)
		if !strings.HasPrefix(fileName, "stacks-master/incubator/nodejs/") {
			continue
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
