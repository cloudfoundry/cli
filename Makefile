CF_DIAL_TIMEOUT ?= 15
NODES ?= 10
PACKAGES ?= api actor command types util version integration/helpers
LC_ALL = "en_US.UTF-8"

CF_BUILD_VERSION ?= $$(cat BUILD_VERSION)
CF_BUILD_VERSION_V7 ?= $$(cat BUILD_VERSION_V7)
CF_BUILD_SHA ?= $$(git rev-parse --short HEAD)
CF_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
LD_FLAGS_COMMON=-w -s \
	-X code.cloudfoundry.org/cli/version.binarySHA=$(CF_BUILD_SHA) \
	-X code.cloudfoundry.org/cli/version.binaryBuildDate=$(CF_BUILD_DATE)
LD_FLAGS =$(LD_FLAGS_COMMON) \
	-X code.cloudfoundry.org/cli/version.binaryVersion=$(CF_BUILD_VERSION)
LD_FLAGS_V7 =$(LD_FLAGS_COMMON) \
	-X code.cloudfoundry.org/cli/version.binaryVersion=$(CF_BUILD_VERSION_V7)
LD_FLAGS_LINUX = -extldflags \"-static\" $(LD_FLAGS)
LD_FLAGS_LINUX_V7 = -extldflags \"-static\" $(LD_FLAGS_V7)
REQUIRED_FOR_STATIC_BINARY =-a -tags netgo -installsuffix netgo
REQUIRED_FOR_STATIC_BINARY_V7 =-a -tags "V7 netgo" -installsuffix netgo
GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")
UNAME_S := $(shell uname -s)


ifndef TARGET_V7
TARGET = v6
export GOFLAGS =
ginkgo_int = ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60
else
TARGET = v7
export GOFLAGS = '-tags=V7'
ginkgo_int = ginkgo -r -randomizeAllSpecs -slowSpecThreshold 120
endif

all: lint test build

build: out/cf ## Compile and build a new `cf` binary

check-target-env:
ifndef CF_INT_API
	$(error CF_INT_API is undefined)
endif
ifndef CF_INT_PASSWORD
	$(error CF_INT_PASSWORD is undefined)
endif

clean: ## Remove all files from the `out` directory
	rm -f $(wildcard out/cf*)

custom-lint: ## Run our custom linters
	@echo "style linting files:" # this list will grow as we cleanup all the code
	@bash -c "go run bin/style/main.go api util"
	@echo "No lint errors!"
	@echo

fly-windows-experimental: check-target-env
	CF_TEST_SUITE=./integration/shared/experimental fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ --tag "cli-windows"

fly-windows-isolated: check-target-env
	CF_TEST_SUITE=./integration/shared/isolated fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ --tag "cli-windows"

fly-windows-plugin: check-target-env
	CF_TEST_SUITE=./integration/shared/plugin fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ --tag "cli-windows"

fly-windows-push: check-target-env
	CF_TEST_SUITE=./integration/v6/push fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ --tag "cli-windows"

fly-windows-global: check-target-env
	CF_TEST_SUITE=./integration/shared/global fly -t ci execute -c ci/cli/tasks/integration-windows-serial.yml -i cli=./ --tag "cli-windows"

fly-windows-units:
	fly -t ci execute -c ci/cli/tasks/units-windows.yml -i cli=./ -i cli-ci=./ --tag "cli-windows"

format: ## Run go fmt
	go fmt ./...

golangci-lint: ## Run golangci-lint to validate code quality
	golangci-lint run

golangci-lint-fix: ## Run golangci-lint --fix to try to autofix issues
	golangci-lint run --fix

integration-cleanup:
	$(PWD)/bin/cleanup-integration

ie: integration-experimental
integration-experimental: build integration-cleanup integration-shared-experimental integration-experimental-versioned ## Run all experimental integration tests, both versioned and shared across versions

ise: integration-shared-experimental
integration-experimental-shared: integration-shared-experimental
integration-shared-experimental: build integration-cleanup ## Run experimental integration tests that are shared between v6 and v7
	$(ginkgo_int) -nodes $(NODES) integration/shared/experimental

ive: integration-versioned-experimental
integration-experimental-versioned: integration-versioned-experimental
integration-versioned-experimental: build integration-cleanup ## Run experimental integration tests that are specific to your CLI version
	$(ginkgo_int) -nodes $(NODES) integration/$(TARGET)/experimental

ig: integration-global
integration-global: build integration-cleanup integration-shared-global integration-global-versioned ## Run all unparallelizable integration tests that make cross-cutting changes to their test CF foundation

isg: integration-shared-global
integration-global-shared: integration-shared-global
integration-shared-global: build integration-cleanup ## Serially run integration tests that make cross-cutting changes to their test CF foundation and are shared between v6 and v7
	$(ginkgo_int) integration/shared/global

ivg: integration-versioned-global
integration-global-versioned: integration-versioned-global
integration-versioned-global: build integration-cleanup ## Serially run integration tests that make cross-cutting changes to their test CF foundation and are specific to your CLI version
	$(ginkgo_int) integration/$(TARGET)/global

ii: integration-isolated
integration-isolated: build integration-cleanup integration-shared-isolated integration-isolated-versioned ## Run all parallel-enabled integration tests, both versioned and shared across versions

isi: integration-shared-isolated
integration-isolated-shared: integration-shared-isolated
integration-shared-isolated: build integration-cleanup ## Run all parallel-enabled integration tests that are shared between v6 and v7
	$(ginkgo_int) -nodes $(NODES) integration/shared/isolated

integration-performance: build integration-cleanup integration-shared-performance

isp: integration-shared-performance
integration-performance-shared: integration-shared-performance
integration-shared-performance: build integration-cleanup
	$(ginkgo_int) integration/shared/performance

ivi: integration-versioned-isolated
integration-isolated-versioned: integration-versioned-isolated
integration-versioned-isolated: build integration-cleanup ## Run all parallel-enabled integration tests, both versioned and shared across versions
	$(ginkgo_int) -nodes $(NODES) integration/$(TARGET)/isolated

integration-plugin: build integration-cleanup ## Run all plugin-related integration tests
	$(ginkgo_int) -nodes $(NODES) integration/shared/plugin

ip: integration-push
integration-push: build integration-cleanup  ## Run all push-related integration tests
	$(ginkgo_int) -nodes $(NODES) integration/$(TARGET)/push

integration-tests: build integration-cleanup integration-isolated integration-push integration-global ## Run all isolated, push, and global integration tests

i: integration-tests-full
integration-full-tests: integration-tests-full
integration-tests-full: build integration-cleanup integration-isolated integration-push integration-experimental integration-plugin integration-global  ## Run all isolated, push, experimental, plugin, and global integration tests

lint: format custom-lint golangci-lint ## Runs all linters and formatters

ifeq ($(TARGET),v6)
out/cf: out/cf6
	cp out/cf6 out/cf
else
out/cf: out/cf7
	cp out/cf7 out/cf
endif

# Build dynamic binary for Darwin
ifeq ($(UNAME_S),Darwin)
out/cf6: $(GOSRC)
	go build -ldflags "$(LD_FLAGS)" -o out/cf6 .
else
out/cf6: $(GOSRC)
	CGO_ENABLED=0 go build \
		$(REQUIRED_FOR_STATIC_BINARY) \
		-ldflags "$(LD_FLAGS_LINUX)" -o out/cf6 .
endif

# Build dynamic binary for Darwin
ifeq ($(UNAME_S),Darwin)
out/cf7: $(GOSRC)
	go build -tags="V7" -ldflags "$(LD_FLAGS_V7)" -o out/cf7 .
else
out/cf7: $(GOSRC)
	CGO_ENABLED=0 go build \
		$(REQUIRED_FOR_STATIC_BINARY_V7) \
		-ldflags "$(LD_FLAGS_LINUX_V7)" -o out/cf7 .
endif

out/cf-cli_linux_i686: $(GOSRC)
	CGO_ENABLED=0 GOARCH=386 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY) \
							-ldflags "$(LD_FLAGS_LINUX)" -o out/cf-cli_linux_i686 .

out/cf-cli_linux_x86-64: $(GOSRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY) \
							-ldflags "$(LD_FLAGS_LINUX)" -o out/cf-cli_linux_x86-64 .

out/cf-cli_osx: $(GOSRC)
	GOARCH=amd64 GOOS=darwin go build \
				 -a -ldflags "$(LD_FLAGS)" -o out/cf-cli_osx .

out/cf-cli_win32.exe: $(GOSRC) rsrc.syso
	GOARCH=386 GOOS=windows go build -tags="forceposix" -o out/cf-cli_win32.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

out/cf-cli_winx64.exe: $(GOSRC) rsrc.syso
	GOARCH=amd64 GOOS=windows go build -tags="forceposix" -o out/cf-cli_winx64.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

out/cf7-cli_linux_i686: $(GOSRC)
	CGO_ENABLED=0 GOARCH=386 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY_V7) \
							-ldflags "$(LD_FLAGS_LINUX_V7)" -o out/cf7-cli_linux_i686 .

out/cf7-cli_linux_x86-64: $(GOSRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY_V7) \
							-ldflags "$(LD_FLAGS_LINUX_V7)" -o out/cf7-cli_linux_x86-64 .

out/cf7-cli_osx: $(GOSRC)
	GOARCH=amd64 GOOS=darwin go build -tags="V7" \
				 -a -ldflags "$(LD_FLAGS_V7)" -o out/cf7-cli_osx .

out/cf7-cli_win32.exe: $(GOSRC) rsrc.syso
	GOARCH=386 GOOS=windows go build -tags="forceposix V7" -o out/cf7-cli_win32.exe -ldflags "$(LD_FLAGS_V7)" .
	rm rsrc.syso

out/cf7-cli_winx64.exe: $(GOSRC) rsrc.syso
	GOARCH=amd64 GOOS=windows go build -tags="forceposix V7" -o out/cf7-cli_winx64.exe -ldflags "$(LD_FLAGS_V7)" .
	rm rsrc.syso

rsrc.syso:
	@# Software for windows icon
	go get github.com/akavel/rsrc
	@# Generates icon file
	rsrc -ico ci/installers/windows/cf.ico

test: units ## (synonym for units)

units: build ## Ensure the code looks good, compiles, and passes unit tests
	ginkgo -r -p -randomizeAllSpecs -randomizeSuites \
		$(PACKAGES)
	@echo "\nSWEET SUITE SUCCESS"

units-plugin:
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes 1 -randomizeAllSpecs -randomizeSuites -flakeAttempts 2 -skipPackage integration ./**/plugin* ./plugin

units-non-plugin:
	@rm -f $(wildcard fixtures/plugins/*.exe)
	@ginkgo version
	CF_HOME=$(PWD)/fixtures ginkgo -r -p -randomizeAllSpecs -randomizeSuites \
		-skipPackage integration,cf/ssh,plugin,cf/actors/plugin,cf/commands/plugin,cf/actors/plugin
	CF_HOME=$(PWD)/fixtures ginkgo -r -p -randomizeAllSpecs -randomizeSuites -flakeAttempts 3 cf/ssh

units-full: build units-plugin units-non-plugin
	@echo "\nSWEET SUITE SUCCESS"

version: ## Print the version number of what would be built
	@echo $(CF_BUILD_VERSION)+$(CF_BUILD_SHA).$(CF_BUILD_DATE)

.PHONY: all build clean format version lint custom-lint golangci-lint golangci-lint-fix
.PHONY: test units units-full integration integration-tests-full integration-cleanup integration-experimental integration-plugin integration-isolated integration-push
.PHONY: check-target-env fly-windows-experimental fly-windows-isolated fly-windows-plugin fly-windows-push
.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-34s\033[0m %s\n", $$1, $$2}'
