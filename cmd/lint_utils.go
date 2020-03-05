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
	"regexp"

	"github.com/mitchellh/go-spdx"
	"github.com/pkg/errors"
)

func CheckValidSemver(version string) error {
	versionRegex := regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	checkVersionNo := versionRegex.FindString(version)

	if checkVersionNo == "" {
		return errors.Errorf("Version must be formatted in accordance to semver - Please see: https://semver.org/ for valid versions.")
	}

	return nil
}

func checkValidLicense(log *LoggingConfig, license string) error {
	// Get the list of all known licenses
	list, _ := spdx.List()
	if list != nil {
		for _, spdx := range list.Licenses {
			if spdx.ID == license {
				return nil
			}
		}
	} else {
		log.Warning.log("Unable to check if license ID is valid.... continuing.")
		return nil
	}
	return errors.New("file must have a valid license ID, see https://spdx.org/licenses/ for the list of valid licenses")
}

func lintMountPathForSingleFile(path string, log *LoggingConfig) {

	file, err := os.Stat(path)
	if err != nil {
		log.Warning.logf("Could not stat mount path: %s", path)

	} else {
		if file.Mode().IsDir() {
			log.Debug.logf("Path %s for mount is a directory", path)
		} else {

			log.Warning.logf("Path %s for mount points to a single file.  Single file Docker mount paths cause unexpected behavior and will be deprecated in the future.", path)
		}

	}
}
