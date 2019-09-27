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

// Simple test for appsody build command. A future enhancement would be to verify the image that gets built.
func TestTest(projectDir string) error {

	Info.log("******************************************")
	Info.log("Running appsody test")
	Info.log("******************************************")
	_, err := RunAppsodyCmdExec([]string{"test"}, projectDir)
	if err != nil {
		Error.log(err)
		return err
	}

	return nil
}
