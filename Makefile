CF_DIAL_TIMEOUT ?= 15
NODES ?= 4
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

ginkgo_int = ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60

ifndef TARGET_V7
TARGET = v6
TAGS=''
else
TARGET = v7
ifdef RUN_V7_PUSH_TESTS
TAGS=''
else
TAGS='partialPush'
endif
endif

all : test build

build : out/cf

check-target-env :
ifndef CF_INT_API
	$(error CF_INT_API is undefined)
endif
ifndef CF_INT_PASSWORD
	$(error CF_INT_PASSWORD is undefined)
endif

clean :
	rm -f $(wildcard out/cf*)

format :
	go fmt ./...

fly-windows-experimental : check-target-env
	CF_TEST_SUITE=./integration/shared/experimental fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-isolated : check-target-env
	CF_TEST_SUITE=./integration/shared/isolated fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-plugin : check-target-env
	CF_TEST_SUITE=./integration/shared/plugin fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-push : check-target-env
	CF_TEST_SUITE=./integration/v6/push fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-global : check-target-env
	CF_TEST_SUITE=./integration/shared/global fly -t ci execute -c ci/cli/tasks/integration-windows-serial.yml -i cli=./

fly-windows-units :
	fly -t ci execute -c ci/cli/tasks/units-windows.yml -i cli=./ -i cli-ci=./

integration-cleanup :
	$(PWD)/bin/cleanup-integration

integration-experimental : build integration-cleanup integration-experimental-shared integration-experimental-versioned

integration-experimental-shared : build integration-cleanup
	$(ginkgo_int) -nodes $(NODES) -tags=$(TAGS) integration/shared/experimental

integration-experimental-versioned : build integration-cleanup
	$(ginkgo_int) -nodes $(NODES) integration/$(TARGET)/experimental

integration-global : build integration-cleanup integration-global-shared integration-global-versioned

integration-global-shared : build integration-cleanup
	$(ginkgo_int) -tags=$(TAGS) integration/shared/global

integration-global-versioned : build integration-cleanup
	$(ginkgo_int) -tags=$(TAGS) integration/$(TARGET)/global

integration-isolated : build integration-cleanup integration-isolated-shared integration-isolated-versioned

integration-isolated-shared : build integration-cleanup
	$(ginkgo_int) -nodes $(NODES) -tags=$(TAGS) integration/shared/isolated

integration-isolated-versioned : build integration-cleanup
	$(ginkgo_int) -nodes $(NODES) -tags=$(TAGS) integration/$(TARGET)/isolated

integration-plugin : build integration-cleanup
	$(ginkgo_int) -nodes $(NODES) -tags=$(TAGS) integration/shared/plugin

integration-push : build integration-cleanup
	$(ginkgo_int) -nodes $(NODES) -tags=$(TAGS) integration/$(TARGET)/push

integration-tests : build integration-cleanup integration-isolated integration-push integration-global

integration-tests-full : build integration-cleanup integration-isolated integration-push integration-experimental integration-plugin integration-global

lint :
	@echo "style linting files:" # this list will grow as we cleanup all the code
	@bash -c "go run bin/style/main.go api util/{configv3,manifest,randomword,sorting,ui}"
	@echo "No lint errors!"
	@echo

ifeq ($(TARGET),v6)
out/cf : out/cf6
	cp out/cf6 out/cf
else
out/cf : out/cf7
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

out/cf-cli_linux_i686 : $(GOSRC)
	CGO_ENABLED=0 GOARCH=386 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY) \
							-ldflags "$(LD_FLAGS_LINUX)" -o out/cf-cli_linux_i686 .

out/cf-cli_linux_x86-64 : $(GOSRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY) \
							-ldflags "$(LD_FLAGS_LINUX)" -o out/cf-cli_linux_x86-64 .

out/cf-cli_osx : $(GOSRC)
	GOARCH=amd64 GOOS=darwin go build \
				 -a -ldflags "$(LD_FLAGS)" -o out/cf-cli_osx .

out/cf-cli_win32.exe : $(GOSRC) rsrc.syso
	GOARCH=386 GOOS=windows go build -tags="forceposix" -o out/cf-cli_win32.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

out/cf-cli_winx64.exe : $(GOSRC) rsrc.syso
	GOARCH=amd64 GOOS=windows go build -tags="forceposix" -o out/cf-cli_winx64.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

out/cf7-cli_linux_i686 : $(GOSRC)
	CGO_ENABLED=0 GOARCH=386 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY_V7) \
							-ldflags "$(LD_FLAGS_LINUX_V7)" -o out/cf7-cli_linux_i686 .

out/cf7-cli_linux_x86-64 : $(GOSRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build \
							$(REQUIRED_FOR_STATIC_BINARY_V7) \
							-ldflags "$(LD_FLAGS_LINUX_V7)" -o out/cf7-cli_linux_x86-64 .

out/cf7-cli_osx : $(GOSRC)
	GOARCH=amd64 GOOS=darwin go build -tags="V7" \
				 -a -ldflags "$(LD_FLAGS_V7)" -o out/cf7-cli_osx .

out/cf7-cli_win32.exe : $(GOSRC) rsrc.syso
	GOARCH=386 GOOS=windows go build -tags="forceposix V7" -o out/cf7-cli_win32.exe -ldflags "$(LD_FLAGS_V7)" .
	rm rsrc.syso

out/cf7-cli_winx64.exe : $(GOSRC) rsrc.syso
	GOARCH=amd64 GOOS=windows go build -tags="forceposix V7" -o out/cf7-cli_winx64.exe -ldflags "$(LD_FLAGS_V7)" .
	rm rsrc.syso

rsrc.syso :
	@# Software for windows icon
	go get github.com/akavel/rsrc
	@# Generates icon file
	rsrc -ico ci/installers/windows/cf.ico

test : units

units : format vet lint build
	ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites \
		$(PACKAGES)
	@echo "\nSWEET SUITE SUCCESS"

units-plugin :
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes 1 -randomizeAllSpecs -randomizeSuites -flakeAttempts 2 -skipPackage integration ./**/plugin* ./plugin

units-non-plugin :
	@rm -f $(wildcard fixtures/plugins/*.exe)
	@ginkgo version
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites \
		-skipPackage integration,cf/ssh,plugin,cf/actors/plugin,cf/commands/plugin,cf/actors/plugin
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites -flakeAttempts 3 cf/ssh

units-full: format vet lint build units-plugin units-non-plugin
	@echo "\nSWEET SUITE SUCCESS"

version :
	@echo $(CF_BUILD_VERSION)+$(CF_BUILD_SHA).$(CF_BUILD_DATE)

vet :
	@echo  "Vetting packages for potential issues..."
	go tool vet -all -shadow=true ./api ./actor ./command ./integration ./types ./util ./version
	@echo

.PHONY : all build clean format version vet lint
.PHONY : test units units-full integration integration-tests-full integration-cleanup integration-experimental integration-plugin integration-isolated integration-push
.PHONY : check-target-env fly-windows-experimental fly-windows-isolated fly-windows-plugin fly-windows-push
