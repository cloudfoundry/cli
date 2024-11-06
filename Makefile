CF_DIAL_TIMEOUT ?= 15
NODES ?= 10
FLAKE_ATTEMPTS ?=5
PACKAGES ?= api actor command types util version integration/helpers
LC_ALL = "en_US.UTF-8"

CF_BUILD_VERSION ?= $$(git describe --tags --abbrev=0)
CF_BUILD_SHA ?= $$(git rev-parse --short HEAD)
CF_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
LD_FLAGS_COMMON=-w -s \
	-X code.cloudfoundry.org/cli/version.binarySHA=$(CF_BUILD_SHA) \
	-X code.cloudfoundry.org/cli/version.binaryBuildDate=$(CF_BUILD_DATE)
LD_FLAGS =$(LD_FLAGS_COMMON) \
	-X code.cloudfoundry.org/cli/version.binaryVersion=$(CF_BUILD_VERSION)
LD_FLAGS_LINUX = -extldflags \"-static\" $(LD_FLAGS)
REQUIRED_FOR_STATIC_BINARY =-a -tags "netgo" -installsuffix netgo
GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")
UNAME_S := $(shell uname -s)

POLL_PROGRESS_THRESHOLD=120s

GINKGO_FLAGS ?= -r -randomize-all -require-suite
GINKGO_INT_FLAGS = $(GINKGO_FLAGS) --poll-progress-after $(POLL_PROGRESS_THRESHOLD)
ginkgo_int = ginkgo $(GINKGO_INT_FLAGS)

GINKGO_UNITS_FLAGS = $(GINKGO_FLAGS) -randomize-suites
ginkgo_units = ginkgo $(GINKGO_UNITS_FLAGS)
GOFLAGS := -mod=mod

all: lint test build

build: out/cf ## Compile and build a new `cf` binary

check-target-env:
ifndef CF_INT_API
	$(error CF_INT_API is undefined)
endif
ifndef CF_INT_PASSWORD
	$(error CF_INT_PASSWORD is undefined)
endif

clean: ## Just remove all cf* files from the `out` directory
	rm -f $(wildcard out/cf*)

clear: clean  ## Make everyone happy

# TODO: update these fly-windows* to point at the correct CI repo
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

install-test-deps: ## Install "global" dependencies needed to run tests
# Running `go install <pkg>` without a version specifier will use version specified in go.mod
# See https://go.dev/ref/mod#go-install
	go install github.com/onsi/ginkgo/v2/ginkgo

integration-cleanup:
	$(CURDIR)/bin/cleanup-integration

ie: integration-experimental
integration-experimental: build integration-cleanup integration-shared-experimental integration-experimental-versioned ## Run all experimental integration tests, both versioned and shared across versions

ise: integration-shared-experimental
integration-experimental-shared: integration-shared-experimental
integration-shared-experimental: build install-test-deps integration-cleanup ## Run experimental integration tests that are shared between v6 and v7
	$(ginkgo_int) -nodes $(NODES) integration/shared/experimental

ive: integration-versioned-experimental
integration-experimental-versioned: integration-versioned-experimental
integration-versioned-experimental: build install-test-deps integration-cleanup ## Run experimental integration tests that are specific to your CLI version
	$(ginkgo_int) -nodes $(NODES) integration/v7/experimental

ig: integration-global
integration-global: build integration-cleanup integration-shared-global integration-global-versioned ## Run all unparallelizable integration tests that make cross-cutting changes to their test CF foundation

isg: integration-shared-global
integration-global-shared: integration-shared-global
integration-shared-global: build install-test-deps integration-cleanup ## Serially run integration tests that make cross-cutting changes to their test CF foundation and are shared between v6 and v7
	$(ginkgo_int) integration/shared/global

ivg: integration-versioned-global
integration-global-versioned: integration-versioned-global
integration-versioned-global: build install-test-deps integration-cleanup ## Serially run integration tests that make cross-cutting changes to their test CF foundation and are specific to your CLI version
	$(ginkgo_int) integration/v7/global

ii: integration-isolated
integration-isolated: build integration-cleanup integration-shared-isolated integration-isolated-versioned ## Run all parallel-enabled integration tests, both versioned and shared across versions

isi: integration-shared-isolated
integration-isolated-shared: integration-shared-isolated
integration-shared-isolated: build install-test-deps integration-cleanup ## Run all parallel-enabled integration tests that are shared between v6 and v7
	$(ginkgo_int) -nodes $(NODES) integration/shared/isolated

integration-performance: build integration-cleanup integration-shared-performance

isp: integration-shared-performance
integration-performance-shared: integration-shared-performance
integration-shared-performance: build install-test-deps integration-cleanup
	$(ginkgo_int) integration/shared/performance

ivi: integration-versioned-isolated
integration-isolated-versioned: integration-versioned-isolated
integration-versioned-isolated: build install-test-deps integration-cleanup ## Run all parallel-enabled integration tests, both versioned and shared across versions
	$(ginkgo_int) -nodes $(NODES) integration/v7/isolated

integration-plugin: build install-test-deps integration-cleanup ## Run all plugin-related integration tests
	$(ginkgo_int) -nodes $(NODES) integration/shared/plugin

ip: integration-push
integration-push: build install-test-deps integration-cleanup  ## Run all push-related integration tests
	$(ginkgo_int) -nodes $(NODES) integration/v7/push

integration-selfcontained: build install-test-deps 
	$(ginkgo_int) -nodes $(NODES) integration/v7/selfcontained

integration-tests: build integration-cleanup integration-isolated integration-push integration-global integration-selfcontained ## Run all isolated, push, selfcontained, and global integration tests

integration-tests-ci-client-creds: build integration-cleanup integration-push integration-global integration-selfcontained

i: integration-tests-full
integration-full-tests: integration-tests-full
integration-tests-full: build integration-cleanup integration-isolated integration-push integration-experimental integration-plugin integration-global integration-selfcontained ## Run all isolated, push, experimental, plugin, selfcontained, and global integration tests

integration-tests-full-ci: install-test-deps integration-cleanup
	$(ginkgo_int) -nodes $(NODES)  -flake-attempts $(FLAKE_ATTEMPTS) \
		integration/shared/isolated integration/v7/isolated integration/shared/plugin integration/shared/experimental integration/v7/experimental integration/v7/push
	$(ginkgo_int) -flake-attempts $(FLAKE_ATTEMPTS) integration/shared/global integration/v7/global

lint: format ## Runs all linters and formatters
	@echo "Running linters..."
	golangci-lint run --exclude-dirs cf --exclude-dirs fixtures --exclude-dirs plugin --exclude-dirs command/plugin
	@echo "No lint errors!"

# TODO: version specific tagging for all these builds
# Build dynamic binary for Darwin
ifeq ($(UNAME_S),Darwin)
out/cf: $(GOSRC)
	go build -ldflags "$(LD_FLAGS)" -o out/cf .
else
out/cf: $(GOSRC)
	CGO_ENABLED=0 go build \
		$(REQUIRED_FOR_STATIC_BINARY) \
		-ldflags "$(LD_FLAGS_LINUX)" -o out/cf .
endif

out/cf-cli_linux_i686: $(GOSRC)
	CGO_ENABLED=0 GOARCH=386 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY) \
							-ldflags "$(LD_FLAGS_LINUX)" -o out/cf-cli_linux_i686 .

out/cf-cli_linux_x86-64: $(GOSRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY) \
							-ldflags "$(LD_FLAGS_LINUX)" -o out/cf-cli_linux_x86-64 .
							
out/cf-cli_linux_arm64: $(GOSRC)
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY) \
							-ldflags "$(LD_FLAGS_LINUX)" -o out/cf-cli_linux_arm64 .

out/cf-cli_osx: $(GOSRC)
	GOARCH=amd64 GOOS=darwin go build \
				 -a -ldflags "$(LD_FLAGS)" -o out/cf-cli_osx .

out/cf-cli_macosarm: $(GOSRC)
	GOARCH=arm64 GOOS=darwin go build \
				 -a -ldflags "$(LD_FLAGS)" -o out/cf-cli_macosarm .

out/cf-cli_win32.exe: $(GOSRC) rsrc.syso
	GOARCH=386 GOOS=windows go build -tags="forceposix" -o out/cf-cli_win32.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

out/cf-cli_winx64.exe: $(GOSRC) rsrc.syso
	GOARCH=amd64 GOOS=windows go build -tags="forceposix" -o out/cf-cli_winx64.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

rsrc.syso:
	rsrc -ico cf.ico -o rsrc.syso

test: units ## (synonym for units)

units: units-full ## (synonym for units-full)

units-plugin: install-test-deps
	CF_HOME=$(CURDIR)/fixtures $(ginkgo_units) -nodes 1 -flake-attempts 2 -skip-package integration ./**/plugin* ./plugin

ifeq ($(OS),Windows_NT)
units-non-plugin: install-test-deps
	@rm -f $(wildcard fixtures/plugins/*/*.exe)
	@ginkgo version
	CF_HOME=$(CURDIR)/fixtures CF_USERNAME="" CF_PASSWORD="" $(ginkgo_units) \
		-skip-package integration,plugin,cf\actors\plugin,cf\commands\plugin,cf\actors\plugin,util\randomword
else
units-non-plugin: install-test-deps
	@rm -f $(wildcard fixtures/plugins/*/*.exe)
	@ginkgo version
	CF_HOME=$(CURDIR)/fixtures CF_USERNAME="" CF_PASSWORD="" $(ginkgo_units) \
		-skip-package integration,plugin,cf/actors/plugin,cf/commands/plugin,cf/actors/plugin,util/randomword
endif

units-full: build units-plugin units-non-plugin
	@echo "\nSWEET SUITE SUCCESS"

version: ## Print the version number of what would be built
	@echo $(CF_BUILD_VERSION)+$(CF_BUILD_SHA).$(CF_BUILD_DATE)

.PHONY: all build clean format version lint
.PHONY: test units units-full install-test-deps integration integration-tests-full integration-cleanup integration-experimental integration-plugin integration-isolated integration-push
.PHONY: check-target-env fly-windows-experimental fly-windows-isolated fly-windows-plugin fly-windows-push
.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-34s\033[0m %s\n", $$1, $$2}'
