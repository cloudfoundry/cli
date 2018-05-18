CF_DIAL_TIMEOUT ?= 15
NODES ?= 4
LC_ALL = "en_US.UTF-8"

CF_BUILD_VERSION ?= $$(cat BUILD_VERSION)
CF_BUILD_SHA ?= $$(git rev-parse --short HEAD)
CF_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
LD_FLAGS =-w -s \
	-X code.cloudfoundry.org/cli/version.binaryVersion=$(CF_BUILD_VERSION) \
	-X code.cloudfoundry.org/cli/version.binarySHA=$(CF_BUILD_SHA) \
	-X code.cloudfoundry.org/cli/version.binaryBuildDate=$(CF_BUILD_DATE)
LD_FLAGS_LINUX = -extldflags \"-static\" $(LD_FLAGS)
REQUIRED_FOR_STATIC_BINARY =-a -tags netgo -installsuffix netgo
GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")

all : test build

build : out/cf-cli_linux_x86-64
	cp out/cf-cli_linux_x86-64 out/cf

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
	CF_CLI_EXPERIMENTAL=true CF_TEST_SUITE=./integration/experimental fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-isolated : check-target-env
	CF_TEST_SUITE=./integration/isolated fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-plugin : check-target-env
	CF_TEST_SUITE=./integration/plugin fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-push : check-target-env
	CF_TEST_SUITE=./integration/push fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./

fly-windows-units :
	fly -t ci execute -c ci/cli/tasks/units-windows.yml -i cli=./ -i cli-ci=./

integration-cleanup :
	$(PWD)/bin/cleanup-integration

integration-experimental : build integration-cleanup
	CF_CLI_EXPERIMENTAL=true ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(NODES) integration/experimental

integration-global : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 integration/global

integration-isolated : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(NODES) integration/isolated

integration-plugin : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(NODES) integration/plugin

integration-push : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(NODES) integration/push

integration-tests : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(NODES) integration/isolated integration/push
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 integration/global
	make integration-cleanup

integration-tests-full : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(NODES) \
		integration/isolated integration/push integration/plugin integration/experimental
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 integration/global
	make integration-cleanup

lint :
	@echo "style linting files:" # this list will grow as we cleanup all the code
	@bash -c "go run bin/style/main.go api util/{configv3,manifest,randomword,sorting,ui}"
	@echo "No lint errors!"
	@echo

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
				 $(REQUIRED_FOR_STATIC_BINARY) \
				 -ldflags "$(LD_FLAGS)" -o out/cf-cli_osx .

out/cf-cli_win32.exe : $(GOSRC) rsrc.syso
	GOARCH=386 GOOS=windows go build -tags="forceposix" -o out/cf-cli_win32.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

out/cf-cli_winx64.exe : $(GOSRC) rsrc.syso
	GOARCH=amd64 GOOS=windows go build -tags="forceposix" -o out/cf-cli_winx64.exe -ldflags "$(LD_FLAGS)" .
	rm rsrc.syso

rsrc.syso :
	@# Software for windows icon
	go get github.com/akavel/rsrc
	@# Generates icon file
	rsrc -ico ci/installers/windows/cf.ico

test : units

units : format vet lint build
	ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites \
		api actor command types util version
	@echo "\nSWEET SUITE SUCCESS"

units-full : format vet lint build
	@rm -f $(wildcard fixtures/plugins/*.exe)
	@ginkgo version
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites -skipPackage integration,cf/ssh
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites -flakeAttempts 3 cf/ssh
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
