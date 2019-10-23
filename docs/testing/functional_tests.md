# Appsody Functional Tests
## Install/Uninstall/Upgrade
#### Synopsis
- Verify Appsody can be installed, upgraded, uninstalled on each supported OS: Windows, macOS, RHEL, Ubuntu
#### Documentation
1. Steps are documented here: https://appsody.dev/docs/getting-started/installation
#### Good Path
1. 
#### Bad Path
1. Verify proper error message if Appsody is installed on an unsupported OS
1. Verify proper error message if the user moved the Appsody home directory
#### Variations
1. Install Appsody on a different volume/disk than the Appsody project
1. Verify different OS versions of Windows, macOS, RHEL, Ubuntu
#### Automation
1. 
#### Status
1. Terrence does an upgrade on each OS for each new CLI release and mixes in some installs and uninstalls from time to time
#### Needs/Gaps
1. Haven't done many scratch installs or uninstalls

## MISC
### TTY
### Buildah

## Appsody commands
### appsody build
#### Synopsis
- Verify a docker image of the Appsody project can be built
#### Documentation
1. 
#### Good Path
1. 
#### Bad Path
1. Verify proper error message when deploying a project with a "bad" stack (e.g. a relevant file has a syntax error)
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (--docker-options, --dryrun, ...)
#### Automation
1. The `appsody stack validate` command performs an `appsody build`
1. There are existing unit tests which test `appsody build`
#### Status
1. Terrence typically runs `appsody stack validate` on all approved stacks on multiple operating systems with each CLI release
#### Needs/Gaps
1. Haven't done much bad path testing or tested many of the command parameters
1. Would be nice to automate testing against all "approved" stacks

### appsody completion
#### Synopsis
- Verify functionality of the bash tab completions
#### Documentation
1. 
#### Good Path
1. 
#### Bad Path
1. 
#### Variations
1. Verify command operation on all supported operating systems (Windows(?), macOS, RHEL, Ubuntu)
#### Automation
1. 
#### Status
1. I don't think there has been any test focus on this command outside of some people using the command
#### Needs/Gaps
1. Not sure if this needs much test focus

### appsody debug
#### Synopsis
- Verify the local Appsody build environment in debug mode
#### Documentation
1. 
#### Good Path
1. Verify proper debug messages are provided
1. Verify the provided endpoint is reachable once the project is deployed
#### Bad Path
1. Verify proper error message when running `appsody debug` against a a "bad" stack (e.g. a relevant file has a syntax error) 
1. Verify proper error message and recompile while the environment is running and a relevant file is saved with syntax or compile error
    1. Try this using different editors (VS Code, vi, notepad, notepad++, ...)
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (--publish-all, --no-watcher ...)
#### Automation
1. There are existing unit tests which test `appsody debug` but I'm not sure how much verification they do
#### Status
1. I don't think there has been any test focus on this command outside of the unit tests
#### Needs/Gaps
1. Not sure if this needs much test focus

### appsody deploy
#### Synopsis
- Verify deployment of an Appsody project
#### Documentation
1. The operator test plan has some deploy scenarios: https://github.com/appsody/appsody/blob/master/docs/testing/operator_test_plan_v3.md
1. Tekton deploy notes are documented here: https://github.com/appsody/appsody/blob/master/docs/testing/tekton_test_notes.md
1. Minikube deploy notes are documented here: https://github.com/appsody/appsody/blob/master/docs/testing/deploy_knative_istio_test_notes.md
#### Good Path
1. Verify the provided endpoint is reachable once the project is deployed
#### Bad Path
1. Verify proper error message when deploying a project with a "bad" deployment yaml
1. Verify proper error message when deploying a project with a "bad" stack (e.g. a relevant file has a syntax error) 
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (delete, --push, --knative, --generate-only...)
#### Automation
1. 
#### Status
1. Terrence and Kim run the appsody operator tests from time to time which cover some deploy scenarios
#### Needs/Gaps
1. Similar to appsody operator command as we need to figure out what the environments are and how to automate
1. Haven't tested any bad path scenarios outside of what is in the test plan
1. Haven't tested all the supported command parameters 
1. Need a way to automate endpoint validation

### appsody extract
#### Synopsis
- Verify the extraction of a stack and project to a local directory
#### Documentation
1. 
#### Good Path
1. 
#### Bad Path
1. 
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (--buildah, --target-dir, ...)
#### Automation
1. There are existing unit tests which test `appsody extract`
#### Status
1. I don't think there has been any test focus on this command outside of the unit tests
#### Needs/Gaps
1. Existing unit test is sufficient?
1. Need some bad path tests?

### appsody list
#### Synopsis
- Verify that correct repositories and stacks are listed
#### Documentation
1. 
#### Good Path
1. Long values are formatted correctly
1. Default repository is indicated 
1. Default template is indicated
#### Bad Path
1. 
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify all supported commands and parameters (--output, --dryrun, ...)
#### Automation
1. 
#### Status
1. I don't think there has been any test focus on this command outside of some people using the command
#### Needs/Gaps
1. 

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
1. There are existing unit tests which test `appsody init`
#### Status
1. Terrence typically runs `appsody stack validate` on all approved stacks on multiple operating systems with each CLI release
#### Needs/Gaps
1. Haven't done much bad path testing or tested many of the command parameters
1. Would be nice to automate testing against all "approved" stacks

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
1. 
#### Status
1. Terrence and Kim run the documented scenarios from time to time
#### Needs/Gaps
1. Need a non-local environment to test on. What are all the environments? (minikube, OKD, ...)
1. Haven't tested any bad path scenarios outside of what is in the test plan
1. Haven't tested all the supported command parameters 
1. Would be nice to automate

### appsody ps
#### Synopsis
- Verify running appsody containers are displayed
#### Documentation
1. 
#### Good Path
1. 
#### Bad Path
1. 
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify all supported commands and parameters (--verbose, --dryrun, ...)
#### Automation
1. There are existing unit tests which test `appsody ps`
#### Status
1. I don't think there has been any test focus on this command outside of the unit tests
#### Needs/Gaps
1. Existing unit test is sufficient?

### appsody repo
#### Synopsis
- Verify Appsody repositories are listed and managed correctly
#### Documentation
1. 
#### Good Path
1. `appsody repo add`
1. `appsody repo remove`
1. `appsody repo list`
1. `appsody repo set-default`
#### Bad Path
1. Verify proper error message when the repo commands are given "bad" values
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify all supported commands and parameters (add, remove, list, set-default, --verbose, --dryrun, ...)
#### Automation
1. 
#### Status
1. I don't think there has been any test focus on this command outside of some people using the commands
#### Needs
1. Would be nice to automate this
1. Need some bad path tests

### appsody run
#### Synopsis
- Verify the local Appsody build environment
#### Documentation
1. 
#### Good Path
1. Verify the provided endpoint is reachable
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
1. There are existing unit tests which test `appsody run`
#### Status
1. Terrence typically runs `appsody stack validate` on all approved stacks on multiple operating systems with each CLI release
#### Needs/Gaps
1. Haven't done much bad path testing or tested many of the command parameters
1. Would be nice to automate testing against all "approved" stacks
1. Need a way to automate endpoint validation

### appsody stack
#### Synopsis
- Verify appsody stack commands function properly
#### Documentation
1. 
#### Good Path
1. `appsody stack create`
1. `appsody stack lint`
1. `appsody stack package`
1. `appsody stack validate`
#### Bad Path
1. 
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (create, lint, package, validate, --dryrun, --verbose ...)
#### Automation
1. 
#### Status
1. Terrence typically runs `appsody stack validate` on all approved stacks on multiple operating systems with each CLI release
#### Needs/Gaps
1. Haven't done much bad path testing or tested many of the command parameters
1. Would be nice to automate testing against all "approved" stacks

### appsody stop
#### Synopsis
- Verify the running Appsody container stops
#### Documentation
1. 
#### Good Path
1. 
#### Bad Path
1. Verify proper error message when a "bad" name is given
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (--dryrun, --verbose ...)
1. Verify other ways to "stop" e.g. ctrl-c
#### Automation
1. There are existing unit tests which test `appsody stop`
#### Status
1. I don't think there has been any test focus on this command outside of the unit tests
#### Needs/Gaps
1. Existing unit test is sufficient?

### appsody test
#### Synopsis
- Verify the Appsody stack test 
#### Documentation
1. 
#### Good Path
1. Verify the stack test code works as designed
#### Bad Path
1. Verify proper error message when the stack test code is "bad" (e.g. syntax error or missing files)
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
1. Verify command operation on all "approved" stacks
1. Verify all supported commands and parameters (--network, --no-watcher ...)
#### Automation
1. The `appsody stack validate` command performs an `appsody test`
1. There are existing unit tests which test `appsody test`
#### Status
1. Terrence typically runs `appsody stack validate` on all approved stacks on multiple operating systems with each CLI release
#### Needs/Gaps
1. Haven't done much bad path testing or tested many of the command parameters
1. Would be nice to automate testing against all "approved" stacks

### appsody version
#### Synopsis
- Verify the installed Appsody CLI version is installed
#### Documentation
1. 
#### Good Path
1. 
#### Bad Path
1. 
#### Variations
1. Verify command operation on all supported operating systems (Windows, macOS, RHEL, Ubuntu)
#### Automation
1. 
#### Status
1. I don't think there has been any test focus on this command outside of some people using the command
#### Needs/Gaps
1. 





