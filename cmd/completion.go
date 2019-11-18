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
	"strings"

	"github.com/spf13/cobra"
)

func newCompletionCmd(log *LoggingConfig, rootCmd *cobra.Command) *cobra.Command {
	// bash completions
	var bashCompletionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash tab completions",
		Long: `Outputs a bash completion script for appsody to stdout.  Bash completion is optionally available for your convenience. It helps you fill out appsody commands when you type the [TAB] key.

	To install on macOS
	1. brew install bash-completion
	2. Make sure to update your ~/.bash_profile as instructed
	3. appsody completion > /usr/local/etc/bash_completion.d/appsody

	To install on Linux
	1. On a current Linux OS (in a non-minimal installation), bash completion should be available.
	2. For Debian see the following link for more information:  https://debian-administration.org/article/316/An_introduction_to_bash_completion_part_1
	3. Make sure to copy the appsody completion file generated above into the appropriate directory for your Linux distribution e.g.
	appsody completion >  /etc/bash_completion.d/appsody`,

		Run: func(cmd *cobra.Command, args []string) {
			buf := new(bytes.Buffer)

			log.Debug.log("Running bash completion script")
			_ = rootCmd.GenBashCompletion(buf)
			output := buf.String()
			// We need to use real spaces because .GenBashCompletion does and formatting requires it
			spaceTab := "    "
			afterAppsodyDevInit := strings.Split(output, "_appsody_init()")
			header := "# Outputs a bash completion script for appsody to stdout. " +
				"Bash completion is optionally available for your convenience. It helps you fill out appsody commands when you type the [TAB] key.\n" +
				"# To install on Linux\n" +
				"# 1. On a current Linux OS (in a non-minimal installation), bash completion should be available.\n" +
				"# 2. Place the completion script generated above in your bash completions directory.\n" +
				"# 3. appsody completion > /usr/local/etc/bash_completion.d/appsody\n\n" +
				"# To install on macOS\n" +
				"# 1. brew install bash-completion\n" +
				"# 2. Make sure to update your ~/.bash_profile as instructed\n" +
				"# 3. appsody completion > /usr/local/etc/bash_completion.d/appsody\n"
			extra := "flags=()\n" + spaceTab +
				"kdl=\"$((appsody list) | awk -F ' ' '{print $1}{print $2}')\"\n" +
				spaceTab + "arr=($kdl)\n" + spaceTab + "len=${#arr[@]}\n" +
				spaceTab + "for (( i=2; i<$len; i=i+2 ))\n" + spaceTab + "do\n" +
				spaceTab + spaceTab + "j=i+1\n" + "thecmd=${arr[i]}/${arr[j]}\n" + "commands+=(\"${thecmd}\")\n" + spaceTab + "done\n"

			fmt.Println(header + afterAppsodyDevInit[0] + "_appsody_init()\n" + strings.Replace(afterAppsodyDevInit[1], "flags=()", extra, 1))

		},
	}

	return bashCompletionCmd
}
