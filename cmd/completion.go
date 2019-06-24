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
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// bash completions
var bashCompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash tab completions",
	Long: `Running the completion command:
	Command line completion automatically fills in partially typed commands.  This functionality can be added for appsody commands by doing the following:

	appsody completion > "your-directory"/appsody  
	
	To install on Linux
	1. On a current Linux OS (in a non-minimal installation), bash completion should be available. 
	2. For Debian see the following link for more information:  https://debian-administration.org/article/316/An_introduction_to_bash_completion_part_1 
	3. Make sure to copy the appsody completion file generated above into the appropriate directory for your Linux distribution e.g.
	appsody completion > <your-directory>/appsody
	The <your-directory> could be /etc/bash_completion.d or the output file can be saved in user local directory and sourced from ~/.bashrc file as well.
	
	To install on macOs
	1. Install bash completions if need be with brew or MacPorts  
	2. Make sure to update your ~/.bash_profile as instructed  
	3. Put the output of 'appsody completion' into your bash completions directory e.g. /usr/local/etc/bash_completion.d/  
	`,
	Run: func(cmd *cobra.Command, args []string) {
		buf := new(bytes.Buffer)

		Debug.log("Running bash completion script")
		_ = rootCmd.GenBashCompletion(buf)
		output := buf.String()
		// We need to use real spaces because .GenBashCompletion does and formatting requires it
		spaceTab := "    "
		afterAppsodyDevInit := strings.Split(output, "_appsody_init()")
		header := "# Running the completion command\n #appsody completion > \"your-directory\"/appsody\n\n" +
			"# Linux\n" +
			"# 1. On a current Linux OS (in a non-minimal installation), bash completion should be available.\n" +
			"# 2. Place the completion script generated above in your bash completions directory.\n\n" +
			"# Mac\n" +
			"# 1. Install bash completions if need be with brew or MacPorts\n" +
			"# 2. Make sure to update your ~/.bash_profile as instructed\n" +
			"# 3. Put the output of 'appsody completion' into your bash completions directory e.g. /usr/local/etc/bash_completion.d/\n"

		extra := "flags=()\n" + spaceTab +
			"kdl=\"$((appsody list) | awk -F ' ' '{print $1}')\"\n" +
			spaceTab + "arr=($kdl)\n" + spaceTab + "len=${#arr[@]}\n" +
			spaceTab + "for (( i=1; i<$len; i++ ))\n" + spaceTab + "do\n" +
			spaceTab + spaceTab + "commands+=(\"${arr[$i]}\")\n" + spaceTab + "done\n"

		fmt.Println(header + afterAppsodyDevInit[0] + "_appsody_init()\n" + strings.Replace(afterAppsodyDevInit[1], "flags=()", extra, 1))

	},
}

func init() {
	rootCmd.AddCommand(bashCompletionCmd)

}
