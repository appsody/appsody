// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// docs command is used to generate markdown file for all the appsody commands
var docsCmd = &cobra.Command{
	Use:    "docs",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {

		Debug.log("Running appsody docs command.")
		err := GenerateDoc(docFile)
		if err != nil {
			Error.log("appsody docs command failed with error: ", err)
			os.Exit(1)
		}
		Debug.log("appsody docs command completed successfully.")
	},
}

var docFile string

func init() {
	rootCmd.AddCommand(docsCmd)
	docFlags := flag.NewFlagSet("", flag.ContinueOnError)

	docFlags.StringVar(&docFile, "docFile", "", "Specify the file to contain the generated documentation.")
	docsCmd.PersistentFlags().AddFlagSet(docFlags)
}
