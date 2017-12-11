CF_DIAL_TIMEOUT ?= 15
NODES ?= 4
LC_ALL = "en_US.UTF-8"

CF_BUILD_VERSION ?= $$(cat ci/VERSION)
CF_BUILD_SHA ?= $$(git rev-parse --short HEAD)
CF_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
LD_FLAGS = "-w -s \
	-X code.cloudfoundry.org/cli/version.binaryVersion=$(CF_BUILD_VERSION) \
	-X code.cloudfoundry.org/cli/version.binarySHA=$(CF_BUILD_SHA) \
	-X code.cloudfoundry.org/cli/version.binaryBuildDate=$(CF_BUILD_DATE)"
GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")

all : test build

build : out/cf

check-target-env :
ifndef CF_API
	$(error CF_API is undefined)
endif
ifndef CF_PASSWORD
	$(error CF_PASSWORD is undefined)
endif

clean :
	rm -r $(wildcard out/*)

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

i18n :
	$(PWD)/bin/i18n-checkup

i18n-extract-strings :
	$(PWD)/bin/i18n-extract-strings

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
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(NODES) integration/isolated integration/push integration/plugin integration/experimental
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 integration/global
	make integration-cleanup

out/cf : $(GOSRC)
	go build -o out/cf \
		-ldflags "-w \
							-s \
							-X code.cloudfoundry.org/cli/version.binaryVersion=$(CF_BUILD_VERSION) \
							-X code.cloudfoundry.org/cli/version.binarySHA=$(CF_BUILD_SHA) \
							-X code.cloudfoundry.org/cli/version.binaryBuildDate=$(CF_BUILD_DATE)" \
		.

out/cf-cli-_winx64.exe : $(GOSRC)
	go get github.com/akavel/rsrc
	rsrc -ico ci/installers/windows/cf.ico
	GOARCH=amd64 GOOS=windows go build -tags="forceposix" -o out/cf-cli_winx64.exe -ldflags $(LD_FLAGS) .
	rm rsrc.syso

out/cf-cli-_win32.exe : $(GOSRC)
	go get github.com/akavel/rsrc
	rsrc -ico ci/installers/windows/cf.ico
	GOARCH=386 GOOS=windows go build -tags="forceposix" -o out/cf-cli_win32.exe -ldflags $(LD_FLAGS) .
	rm rsrc.syso

test : units

units : format vet build
	ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites \
		api actor command types util version
	@echo "\nSWEET SUITE SUCCESS"

units-full : format vet build
	@rm -f $(wildcard fixtures/plugins/*.exe)
	@ginkgo version
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes $(NODES) -randomizeAllSpecs -randomizeSuites -skipPackage integration
	@echo "\nSWEET SUITE SUCCESS"

version :
	@echo $(CF_BUILD_VERSION)+$(CF_BUILD_SHA).$(CF_BUILD_DATE)

vet :
	@echo  "Vetting packages for potential issues..."
	go tool vet -all -shadow=true ./api ./actor ./command ./integration ./types ./util ./version

.PHONY : all build clean i18n i18n-extract-strings format version vet
.PHONY : test units units-full integration integration-tests-full integration-cleanup integration-experimental integration-plugin integration-isolated integration-push
.PHONY : check-target-env fly-windows-experimental fly-windows-isolated fly-windows-plugin fly-windows-push
