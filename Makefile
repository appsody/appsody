.DEFAULT_GOAL := help

#### Constant variables
export STACKSLIST ?= incubator/nodejs
# use -count=1 to disable cache and -p=1 to stream output live
GO_TEST_COMMAND := go test -v -count=1 -p=1 -parallel 1 -covermode=count -coverprofile=cover.out -coverpkg ./cmd
GO_TEST_LOGGING := | tee test.out | grep -E "^\s*(---|===)" ; tail -3 test.out ; awk '/--- FAIL/,/===/' test.out ; ! grep -E "(--- FAIL|^FAIL)" test.out
GO_TEST_COVER_VIEWER := go tool cover -func=cover.out && go tool cover -html=cover.out
# Set a default VERSION only if it is not already set
VERSION ?= 0.0.0
COMMAND := appsody
BUILD_PATH := $(PWD)/build
LOCAL_BIN_PATH := $(PWD)/bin
PACKAGE_PATH := $(PWD)/package
DOCS_PATH := $(PWD)/my-project
GO_PATH := $(shell go env GOPATH)
GOLANGCI_LINT_BINARY := $(GO_PATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION := v1.16.0
DEP_BINARY := $(GO_PATH)/bin/dep
DEP_RELEASE_TAG := v0.5.4
BINARY_EXT_linux :=
BINARY_EXT_darwin :=
BINARY_EXT_windows := .exe
DOCKER_IMAGE_RPM := alectolytic/rpmbuilder
DOCKER_IMAGE_DEB := appsody/debian-builder:0.1.0
GH_ORG ?= appsody
export APPSODY_CONTROLLER_VERSION ?=0.3.3

#### Dynamic variables. These change depending on the target name.
# Gets the current os and arch from the target name, e.g. the 'build-linux' target will result in os = 'linux'
# The arch is defaulted to amd64, otherwise it is taken from the part after the second '-', e.g. 'build-linux-ppc64le'
# CAUTION: All targets that use these variables must have the OS after the first '-' in their name and 
#          optionally arch after the second '-'
#          For example, these are all good: build-linux, tar-darwin, tar-darwin-ppc64le
os = $(word 2,$(subst -, ,$@))
arch = $(or $(word 3,$(subst -, ,$@)), amd64)
build_name = $(COMMAND)-$(VERSION)-$(os)-$(arch)
build_binary = $(build_name)$(BINARY_EXT_$(os))
package_binary = $(COMMAND)$(BINARY_EXT_$(os))

.PHONY: all
all: lint test package ## Run lint, test, build, and package

.PHONY: test
test: ## Run the all the automated tests
	$(GO_TEST_COMMAND) ./... $(GO_TEST_LOGGING)


.PHONY: unittest
unittest: ## Run the automated unit tests
	$(GO_TEST_COMMAND) ./cmd $(GO_TEST_LOGGING)

.PHONY: functest
functest: ## Run the automated functional tests
	$(GO_TEST_COMMAND) ./functest $(GO_TEST_LOGGING)


.PHONY: cover
cover: test ## Run all tests and open test coverage report
	$(GO_TEST_COVER_VIEWER)

.PHONY: unitcover
unitcover: unittest ## Run unit tests and open test coverage report
	$(GO_TEST_COVER_VIEWER)

.PHONY: funccover
funccover: functest ## Run functional tests and open test coverage report
	$(GO_TEST_COVER_VIEWER)

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

# not PHONY, installs deps if it doesn't exist
$(DEP_BINARY):
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | DEP_RELEASE_TAG=$(DEP_RELEASE_TAG) sh

.PHONY: ensure
ensure: $(DEP_BINARY) ## Runs `dep ensure` to make sure the Gopkg.lock and /vendor dir are in sync with dependencies
	$(DEP_BINARY) ensure

.PHONY: clean
clean: ## Removes existing build artifacts in order to get a fresh build
	rm -rf $(BUILD_PATH)
	rm -rf $(LOCAL_BIN_PATH)
	rm -rf $(PACKAGE_PATH)
	rm -f $(GOLANGCI_LINT_BINARY)
	rm -f $(DEP_BINARY)
	rm -rf $(DOCS_PATH)
	go clean

.PHONY: build
build: build-linux build-darwin build-windows build-linux-ppc64le ## Build binaries for all operating systems and store them in the build/ dir

.PHONY: build-linux
build-linux: ## Build the linux binary

.PHONY: build-linux-ppc64le
build-linux-ppc64le: ## Build the linux ppc64le binary

.PHONY: build-windows
build-windows: ## Build the windows binary

build-linux build-linux-ppc64le build-windows: ## Build the binary of the respective operating system
	GOOS=$(os) CGO_ENABLED=0 GOARCH=$(arch) go build -o $(BUILD_PATH)/$(build_binary) -ldflags "-X main.VERSION=$(VERSION) -X main.CONTROLLERVERSION=$(APPSODY_CONTROLLER_VERSION)"

.PHONY: build-darwin
build-darwin: ## Build the OSX binary
	GOOS=$(os) GOARCH=$(arch) go build -o $(BUILD_PATH)/$(build_binary) -ldflags "-X main.VERSION=$(VERSION) -X main.CONTROLLERVERSION=$(APPSODY_CONTROLLER_VERSION)"

.PHONY: localbin-darwin
localbin-darwin: ## copy the darwin binary to local bin

.PHONY: localbin-linux
localbin-linux: ## copy the linux binary to local bin

.PHONY: localbin-windows
localbin-windows: ## copy the windows binary to local bin

localbin-darwin localbin-linux localbin-windows:
	mkdir -p $(LOCAL_BIN_PATH)
	cp -p $(BUILD_PATH)/$(build_binary) $(LOCAL_BIN_PATH)/appsody

.PHONY: package
package: build-docs tar-linux tar-linux-ppc64le deb-linux rpm-linux tar-darwin brew-darwin tar-windows ## Creates packages for all operating systems and store them in package/ dir

.PHONY: tar-linux
tar-linux: build-linux ## Build the linux binary and package it in a .tar.gz file
.PHONY: tar-linux-ppc64le
tar-linux-ppc64le: build-linux-ppc64le ## Build the linux binary and package it in a .tar.gz file
.PHONY: tar-darwin
tar-darwin: build-darwin ## Build the OSX binary and package it in a .tar.gz file
tar-linux tar-linux-ppc64le tar-darwin:
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)
	tar cfz $(build_name).tar.gz LICENSE README.md $(package_binary)
	mkdir -p $(PACKAGE_PATH)
	mv $(build_name).tar.gz $(PACKAGE_PATH)/
	rm -f $(package_binary)

.PHONY: tar-windows
tar-windows: build-windows ## Build the windows binary and package it in a .tar.gz file
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)	
	win-build/build-win.sh $(PACKAGE_PATH) $(package_binary) $(VERSION)
	rm -f $(package_binary)

.PHONY: brew-darwin
brew-darwin: build-darwin ## Build the OSX binary and package it for OSX brew install
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)
	homebrew-build/build-darwin.sh $(PACKAGE_PATH) $(package_binary) $(VERSION)
	rm -f $(package_binary)

.PHONY: deb-linux
deb-linux: build-linux ## Build the linux binary and package it as a .deb for Debian apt-get install
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)
	deb-build/build-deb.sh $(package_binary) $(DOCKER_IMAGE_DEB) $(PACKAGE_PATH) $(VERSION)
	rm -f $(package_binary)

.PHONY: rpm-linux
rpm-linux: build-linux ## Build the linux binary and package it as a .rpm for RedHat yum install
	cp -p $(BUILD_PATH)/$(build_binary) $(package_binary)
	rpm-build/build-rpm.sh $(package_binary) $(DOCKER_IMAGE_RPM) $(PACKAGE_PATH) $(VERSION)
	rm -f $(package_binary)

.PHONY: build-docs
build-docs: ## Create the CLI mardown reference doc
	mkdir -p $(DOCS_PATH)
	cd $(DOCS_PATH) && go run ../main.go docs --docFile $(BUILD_PATH)/cli-commands.md && sed -i.bak '/###### Auto generated by spf13/d' $(BUILD_PATH)/cli-commands.md && rm $(BUILD_PATH)/cli-commands.md.bak
	rm -rf $(DOCS_PATH)

.PHONY: deploy
deploy: ## Creates branches in the homebrew and website repos and commits the changes made here
	./deploy-build/deploy.sh
	./docs-build/deploy.sh

# Auto documented help from http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Prints this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
