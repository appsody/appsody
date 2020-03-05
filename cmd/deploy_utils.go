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
	"os/exec"

	"github.com/pkg/errors"
)

//KubeApply issues kubectl apply -f <filename>
func KubeApply(log *LoggingConfig, fileToApply string, namespace string, dryrun bool) error {
	log.Info.log("Attempting to apply resource in Kubernetes ...")
	kcmd := "kubectl"
	kargs := []string{"apply", "-f", fileToApply}
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return errors.Errorf("kubectl apply failed: %s", kout)
	}
	log.Debug.log("kubectl apply success: ", string(kout[:]))
	return kerr
}

//KubeDelete issues kubectl delete -f <filename>
func KubeDelete(log *LoggingConfig, fileToApply string, namespace string, dryrun bool) error {
	log.Info.log("Attempting to delete resource from Kubernetes...")
	kcmd := "kubectl"
	kargs := []string{"delete", "-f", fileToApply}
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)

	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return errors.Errorf("kubectl delete failed: %s", kout)
	}
	log.Debug.log("kubectl delete success: ", kout)
	return kerr
}

//KubeGetNodePortURL kubectl get svc <service> -o jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort} and prints the return URL
func KubeGetNodePortURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kargs := append([]string{"svc"}, service)
	kargs = append(kargs, "-o", "jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort}")
	out, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil {
		return "", errors.Errorf("Failed to find deployed service IP and Port: %s", err)
	}
	return out, nil
}

//KubeGetRouteURL issues kubectl get svc <service> -o jsonpath=http://{.status.loadBalancer.ingress[0].hostname}:{.spec.ports[0].nodePort} and prints the return URL
func KubeGetRouteURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kargs := append([]string{"route"}, service)
	kargs = append(kargs, "-o", "jsonpath={.status.ingress[0].host}")
	out, err := KubeGet(log, kargs, namespace, dryrun)
	// Performing the kubectl apply
	if err != nil {
		return "", errors.Errorf("Failed to find deployed service IP and Port: %s", err)
	}
	return out, nil
}

//KubeGetKnativeURL issues kubectl get rt <service> -o jsonpath="{.status.url}" and prints the return URL
func KubeGetKnativeURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	kcmd := "kubectl"
	kargs := append([]string{"get", "rt"}, service)
	kargs = append(kargs, "-o", "jsonpath=\"{.status.url}\"")
	if namespace != "" {
		kargs = append(kargs, "--namespace", namespace)
	}

	if dryrun {
		log.Info.log("Dry run - skipping execution of: ", kcmd, " ", ArgsToString(kargs))
		return "", nil
	}
	log.Info.log("Running command: ", kcmd, " ", ArgsToString(kargs))
	execCmd := exec.Command(kcmd, kargs...)
	kout, kerr := SeparateOutput(execCmd)
	if kerr != nil {
		return "", errors.Errorf("kubectl get failed: %s", kout)
	}
	return kout, kerr
}

//KubeGetDeploymentURL searches for an exposed hostname and port for the deployed service
func KubeGetDeploymentURL(log *LoggingConfig, service string, namespace string, dryrun bool) (url string, err error) {
	url, err = KubeGetKnativeURL(log, service, namespace, dryrun)
	if err == nil {
		return url, nil
	}
	url, err = KubeGetRouteURL(log, service, namespace, dryrun)
	if err == nil {
		return url, nil
	}
	url, err = KubeGetNodePortURL(log, service, namespace, dryrun)
	if err == nil {
		return url, nil
	}
	log.Error.log("Failed to get deployment hostname and port: ", err)
	return "", err
}
