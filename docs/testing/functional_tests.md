# Appsody Functional Tests
## Install/Upgrade/Uninstall
#### Synopsis
- Verify Appsody can be installed, upgraded, uninstalled on each supported OS: Windows, macOS, RHEL, Ubuntu
#### Documentation
1. Steps are documented here: https://appsody.dev/docs/getting-started/installation
#### Good Path
1. 
#### Bad Path
1. Verify proper error message if Appsody is installed on an unsupported OS
1. Verify proper error message if the user messed with the Appsody home directory
#### Variations
1. Install Appsody on a different volume/disk than the Appsody project
1. Verify different OS versions of Windows, macOS, RHEL, Ubuntu
#### Automation
- 
#### Status
1. Terrence does an upgrade on each OS for each new CLI release and mixes in some installs and uninstalls from time to time
1.  We are not seeing many issues 
#### Needs/Gaps
1. Haven't done many scratch installs or uninstalls

## Appsody commands
### appsody operator
#### Synopsis
- Verify the Appsody operator can be installed and uninstalled and used to deploy an Appsody project to a Kubernetes cluster
#### Documentation
1. Testing on a local Kubernetes cluster is documented here: https://github.com/appsody/appsody/blob/master/docs/testing/operator_test_plan_v3.md
#### Good Path
1. 
#### Bad Path
1. The test plan has some bad path scenarios 
1. Verify proper error messages when required services are not running (knative, istio, docker, ...)
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Non-local Kubernetes clusters
1. Verify all supported commands and parameters (install, uninstall, --namespace, ...)
1. Testing with limited Kubernetes permissions was tested only once as the setup is non-trivial
#### Automation
-
#### Status
1. Terrence and Kim run the documented scenarios from time to time
#### Needs/Gaps
1. Need a non-local environment to test on. What are all the environments? (minikube, OKD, ...)
1. Haven't tested any bad path scenarios outside of what is in the test plan
1. Haven't tested all the supported command parameters 
1. Would be nice to automate

### appsody deploy
#### Synopsis
- Verify deployment of an Appsody project
#### Documentation
1. The operator test plan has some deploy scenarios: https://github.com/appsody/appsody/blob/master/docs/testing/operator_test_plan_v3.md
1. Tekton deploy is documented here: https://github.com/appsody/appsody/blob/master/docs/testing/tekton_test_notes.md
1. Minikube deploy is documented here: https://github.com/appsody/appsody/blob/master/docs/testing/deploy_knative_istio_test_notes.md
#### Good Path
1. Verify the provided endpoint is reachable once the project is deployed
#### Bad Path
1. Verify proper error message when deploying a project with a "bad" deployment yaml
2. Verify proper error message when deploying a project with a "bad" stack (e.g. a relevant file has a syntax error) 
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (delete, --push, --knative, --generate-only...)
#### Automation
-
#### Status
1. Terrence and Kim run the appsody operator tests from time to time which cover some deploy scenarios
#### Needs/Gaps
1. Similar to appsody operator command as we need to figure out what the environments are and how to automate
1. Haven't tested any bad path scenarios outside of what is in the test plan
1. Haven't tested all the supported command parameters 
1. Need a way to automate endpoint validation

### appsody init
#### Synopsis
- Verify the creation of a new Appsody project
#### Documentation
1. 
#### Good Path
1. 
#### Bad Path
1. Verify proper error message when initializing a "bad" stack (e.g. corrupt template, repo doesn't contain the stack, ...) 
1. Verify proper error message when running command in a non-empty directory
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (--no-template, --project-name, ...)
#### Automation
1. The `appsody stack validate` command performs an `appsody init`
2. There are existing unit tests which test `appsody init`
#### Status
1. Terrence typically runs `appsody stack validate` on all approved stacks on multiple operating systems with each CLI release
#### Needs/Gaps
1. Haven't done much bad path testing or tested many of the command parameters
1. Would be nice to automate testing against all "approved" stacks

### appsody run
#### Synopsis
- Verify the local Appsody build environment
#### Documentation
1. 
#### Good Path
1. Verify the provided endpoint is reachable once the project is deployed
1. Perform multiple `appsody run` instances
    1. Same stack
    1. Different stacks
#### Bad Path
1. Verify proper error message when running `appsody run` against a a "bad" stack (e.g. a relevant file has a syntax error) 
1. Verify proper error message and recompile while the environment is running and a relevant file is saved with syntax or compile error
    1. Try this using different editors (VS Code, vi, notepad, notepad++, ...)
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (--publish-all, --no-watcher ...)
#### Automation
1. The `appsody stack validate` command performs an `appsody run`
2. There are existing unit tests which test `appsody run`
#### Status
1. Terrence typically runs `appsody stack validate` on all approved stacks on multiple operating systems with each CLI release
#### Needs/Gaps
1. Haven't done much bad path testing or tested many of the command parameters
1. Would be nice to automate testing against all "approved" stacks
1. Need a way to automate endpoint validation

### appsody test
