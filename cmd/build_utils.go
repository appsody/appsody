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
	"encoding/json"

	"github.com/pkg/errors"
)

func getStackLabels(config *RootCommandConfig) (map[string]string, error) {
	labels := make(map[string]string)
	var data []map[string]interface{}
	var buildahData map[string]interface{}
	var containerConfig map[string]interface{}
	projectConfig, projectConfigErr := getProjectConfig(config)
	if projectConfigErr != nil {
		return nil, projectConfigErr
	}
	imageName := projectConfig.Stack
	pullErrs := pullImage(imageName, config)
	if pullErrs != nil {
		return nil, pullErrs
	}
	inspectOut, err := inspectImage(imageName, config)
	if err != nil {
		return labels, err
	}
	if config.Buildah {
		err = json.Unmarshal([]byte(inspectOut), &buildahData)
		if err != nil {
			return labels, errors.Errorf("Error unmarshaling data from inspect command - exiting %v", err)
		}
		containerConfig = buildahData["config"].(map[string]interface{})
		config.Debug.Log("Config inspected by buildah: ", config)
	} else {
		err := json.Unmarshal([]byte(inspectOut), &data)
		if err != nil {
			return labels, errors.Errorf("Error unmarshaling data from inspect command - exiting %v", err)
		}
		containerConfig = data[0]["Config"].(map[string]interface{})
	}
	if containerConfig["Labels"] != nil {
		labelsMap := containerConfig["Labels"].(map[string]interface{})

		for key, value := range labelsMap {
			labels[key] = value.(string)
		}
	}

	return labels, nil
}
