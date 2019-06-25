.DEFAULT_GOAL := help

#### Constant variables
# Set a default VERSION only if it is not already set
VERSION ?= 0.0.0
COMMAND := appsody
BUILD_PATH := $(PWD)/build
PACKAGE_PATH := $(PWD)/package
GO_PATH := $(shell go env GOPATH)
GOLANGCI_LINT_BINARY := $(GO_PATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION := v1.16.0
BINARY_EXT_linux :=
BINARY_EXT_darwin :=
BINARY_EXT_windows := .exe
DOCKER_IMAGE_RPM := alectolytic/rpmbuilder
DOCKER_IMAGE_DEB := chilantim/debian-builder
CONTROLLER_BASE_URL := https://github.com/${GH_ORG}/controller/releases/download/0.2.5

#### Dynamic variables. These change depending on the target name.
# Gets the current os from the target name, e.g. the 'build-linux' target will result in os = 'linux'
# CAUTION: All targets that use these variables must have the OS after the first '-' in their name.
#          For example, these are all good: build-linux, tar-darwin, tar-darwin-new
os = $(word 2,$(subst -, ,$@))
build_name = $(COMMAND)-$(VERSION)-$(os)-amd64
build_binary = $(build_name)$(BINARY_EXT_$(os))
package_binary = $(COMMAND)$(BINARY_EXT_$(os))

.PHONY: all
all: lint test package ## Run lint, test, build, and package

.PHONY: test
test: ## Run the automated tests
# use -count=1 to disable cache
	go test -v -count=1 ./cmd

.PHONY: lint
lint: $(GOLANGCI_LINT_BINARY) ## Run the static code analyzers
# Configure the linter here. Helpful commands include `golangci-lint linters` and `golangci-lint run -h`
# Set exclude-use-default to true if this becomes to noisy.
	golangci-lint run -v --disable-all \
		--enable deadcode \
		--enable errcheck \
		--enable gosimple \
		--enable govet \
		--enable ineffassign \
		--enable staticcheck \
		--enable structcheck \
		--enable typecheck \
		--enable unused \
		--enable varcheck \
		--enable gofmt \
		--enable golint \
		--enable gofmt \
		--exclude-use-default=true \
		./...

# not PHONY, installs golangci-lint if it doesn't exist
$(GOLANGCI_LINT_BINARY):
	# see https://github.com/golangci/golangci-lint
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GO_PATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: clean
clean: ## Removes existing build artifacts in order to get a fresh build
	rm -rf $(BUILD_PATH)
	rm -rf $(PACKAGE_PATH)
	rm -f $(GOLANGCI_LINT_BINARY)
	go clean

.PHONY: build
build: build-linux build-darwin build-windows ## Build binaries for all operating systems and store them in the build/ dir

.PHONY: build-linux build-darwin build-windows
build-linux: ## Build the linux binary
.PHONY: build-darwin
build-darwin: ## Build the OSX binary
.PHONY: build-windows
build-windows: ## Build the windows binary
build-linux build-darwin build-windows: ## Build the binary of the respective operating system
	GOOS=$(os) GOARCH=amd64 go build -o $(BUILD_PATH)/$(build_binary) -ldflags "-X main.VERSION=$(VERSION)"

.PHONY: package
package: build-docs tar-linux deb-linux rpm-linux tar-darwin brew-darwin tar-windows ## Creates packages for all operating systems and store them in package/ dir

.PHONY: tar-linux
tar-linux: build-linux ## Build the linux binary and package it in a .tar file
.PHONY: tar-darwin
tar-darwin: build-darwin ## Build the OSX binary and package it in a .tar file
tar-linux tar-darwin:
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)
	tar cfz $(build_name).tar LICENSE README.md $(package_binary)
	mkdir -p $(PACKAGE_PATH)
	mv $(build_name).tar $(PACKAGE_PATH)/
	rm -f $(package_binary)
	
.PHONY: tar-windows
tar-windows: build-windows ## Build the windows binary and package it in a .tar file
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)	
	win-build/build-win.sh $(PACKAGE_PATH) $(package_binary) $(CONTROLLER_BASE_URL) $(VERSION)
	rm -f $(package_binary)	

.PHONY: brew-darwin
brew-darwin: build-darwin ## Build the OSX binary and package it for OSX brew install
	# brew script goes here
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)
	homebrew-build/build-darwin.sh $(PACKAGE_PATH) $(package_binary) $(CONTROLLER_BASE_URL) $(VERSION)
	rm -f $(package_binary)	
.PHONY: deb-linux
deb-linux: build-linux ## Build the linux binary and package it as a .deb for Debian apt-get install
	# deb script goes here
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)	
	deb-build/build-deb.sh $(package_binary) $(DOCKER_IMAGE_DEB) $(PACKAGE_PATH) $(CONTROLLER_BASE_URL) $(VERSION)
	rm -f $(package_binary)
.PHONY: rpm-linux
rpm-linux: build-linux ## Build the linux binary and package it as a .rpm for RedHat yum install
	# rpm script goes here
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)	
	rpm-build/build-rpm.sh $(package_binary) $(DOCKER_IMAGE_RPM) $(PACKAGE_PATH) $(CONTROLLER_BASE_URL) $(VERSION)
	rm -f $(package_binary)	


.PHONY: build-docs
build-docs:
		# make docs md file
		mkdir my-project
		cd my-project && go run ../main.go docs --docFile $(BUILD_PATH)/cli-commands.md
		cd ..
		rm -rf my-project
.PHONY: deploy
deploy: ## Publishes the formula
	./deploy-build/deploy.sh
	./docs-build/deploy.sh




# Auto documented help from http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Prints this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
