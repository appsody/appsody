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
	"bufio"
	"os/exec"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// StackContainer is our internal representation of the attributes of stack based container
type StackContainer struct {
	ID            string
	stackName     string
	status        string
	containerName string
}

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List the appsody containers running in the local docker environment",
	Long:  `This command lists all stack-based containers, that are currently running in the local docker envionment.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		setupErr := setupConfig()
		if setupErr != nil {
			return setupErr
		}

		containers, err := listContainers()
		if err != nil {
			return err
		}

		table, err := formatTable(containers)
		if err != nil {
			return errors.Errorf("%v", err)
		}
		if table != "" {
			Info.log(table)
		}
		return nil
	},
}

func listContainers() ([]StackContainer, error) {
	var containers = []StackContainer{}

	// We are going to do a 'docker ps' and parse the output into fields. At least one of these
	// fields can have white space in it (Status), so we need a way of splitting up the output.
	// To do this we use the --format option and include a string of illegal characters as a
	// seperator, which we then subsequently use to parse.
	strSep := "$!$!$!"
	cmdName := "docker"
	cmdArgs := []string{
		"ps",
		"--no-trunc",
		"--format",
		"{{.ID}}" + strSep + "{{.Image}}" + strSep + "{{.Status}}" +
			strSep + "{{.Names}}" + strSep + "{{.Command}}"}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		Error.log("Error creating StdoutPipe for Cmd", err)
		return nil, err
	}

	outScanner := bufio.NewScanner(cmdReader)
	outScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	go func() {
		for outScanner.Scan() {
			fields := strings.Split(outScanner.Text(), strSep)
			if strings.Contains(fields[4], "appsody-controller") {
				containers = append(containers, StackContainer{fields[0][0:12], fields[1], fields[2], fields[3]})
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		Error.log("Error running command", err)
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		Error.log("Error waiting for command", err)
		return nil, err
	}

	return containers, nil
}

func formatTable(containers []StackContainer) (string, error) {
	table := uitable.New()
	table.MaxColWidth = 60
	table.Wrap = true

	if len(containers) != 0 {
		table.AddRow("CONTAINER ID", "NAME", "IMAGE", "STATUS")

		for _, value := range containers {
			table.AddRow(value.ID, value.containerName, value.stackName, value.status)
		}
	} else {
		Info.log("There are no stack-based containers running in your docker environment")
		return "", nil
	}

	return table.String(), nil
}

func init() {
	rootCmd.AddCommand(psCmd)
}
