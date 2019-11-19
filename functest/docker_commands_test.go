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
package functest

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	cmd "github.com/appsody/appsody/cmd"
)

var invalidDockerCmdsTest = []struct {
	file     string
	expected string
}{
	{"imageName", "invalid reference format: repository name must be lowercase"},
	{"imagename", "No such image: imagename"},
}

func TestDockerInspect(t *testing.T) {

	for _, test := range invalidDockerCmdsTest {

		t.Run(fmt.Sprintf("Test Invalid DockerInspect"), func(t *testing.T) {
			var outBuffer bytes.Buffer
			config := &cmd.LoggingConfig{}
			config.InitLogging(&outBuffer, &outBuffer)

			out, err := cmd.RunDockerInspect(config, test.file)
			t.Log(outBuffer.String())

			if err == nil {
				t.Error("Expected an error from '", test.file, "' name but it did not return one.")
			} else if !strings.Contains(out, test.expected) {
				t.Error("Expected the stdout to contain '" + test.expected + "'. It actually contains: " + out)
			}
		})
	}
}
