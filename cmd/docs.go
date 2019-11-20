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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

//generate Doc file (.md) for cmds in package

func generateDoc(log *LoggingConfig, commandDocFile string, rootCmd *cobra.Command) error {

	if commandDocFile == "" {
		return errors.New("no docFile specified")
	}
	dir := filepath.Dir(commandDocFile)

	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		mkdirErr := os.MkdirAll(dir, 0755)
		if mkdirErr != nil {
			log.Error.log("Could not create doc file directory: ", mkdirErr)
			return mkdirErr
		}
	}
	docFile, createErr := os.Create(commandDocFile)
	if createErr != nil {
		log.Error.log("Could not create doc file (.md): ", createErr)
		return createErr
	}

	defer docFile.Close()
	preAmble := "---\ntitle: CLI Reference\n---\n\n# Appsody CLI\n"
	preAmbleBytes := []byte(preAmble)
	_, preambleErr := docFile.Write(preAmbleBytes)
	if preambleErr != nil {
		log.Error.log("Could not write to markdown file:", preambleErr)
		return preambleErr
	}

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		newbase := strings.ReplaceAll(base, "_", "-")
		return "#" + newbase
	}

	var commandArray = []*cobra.Command{}
	commandArray = appendChildren(commandArray, rootCmd)
	for _, cmd := range commandArray {

		markdownGenErr := doc.GenMarkdownCustom(cmd, docFile, linkHandler)

		if markdownGenErr != nil {
			log.Error.log("Doc file generation failed: ", markdownGenErr)
			return markdownGenErr
		}
	}
	return nil

}

func newDocsCmd(log *LoggingConfig, rootCmd *cobra.Command) *cobra.Command {

	var docFile string
	// docs command is used to generate markdown file for all the appsody commands
	var docsCmd = &cobra.Command{
		Use:    "docs",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Debug.log("Running appsody docs command.")
			err := generateDoc(log, docFile, rootCmd)
			if err != nil {
				return errors.Errorf("appsody docs command failed with error: %v", err)

			}
			log.Debug.log("appsody docs command completed successfully.")
			return nil
		},
	}

	docsCmd.PersistentFlags().StringVar(&docFile, "docFile", "", "Specify the file to contain the generated documentation.")
	return docsCmd
}

func appendChildren(commandArray []*cobra.Command, cmd *cobra.Command) []*cobra.Command {

	if !cmd.Hidden && cmd.Name() != "help" {
		commandArray = append(commandArray, cmd)
		for _, value := range cmd.Commands() {

			if !value.Hidden && value.Name() != "help" {
				commandArray = append(commandArray, value)
			}

			for _, childValue := range value.Commands() {
				if !childValue.Hidden && childValue.Name() != "help" {
					commandArray = appendChildren(commandArray, childValue)
				}
			}

		}
	}
	return commandArray
}
