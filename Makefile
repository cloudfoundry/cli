CF_DIAL_TIMEOUT ?= 15
GINKGO_UNIT_TEST_NODES ?= 4
GINKGO_INTEGRATION_TEST_NODES ?= 4
LC_ALL = "en_US.UTF-8"

CF_BUILD_VERSION ?= $$(cat ci/VERSION)
CF_BUILD_SHA ?= $$(git rev-parse --short HEAD)
CF_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*")

all : test i18n-binary build

build : out/cf

check_target_env :
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

fly-windows-experimental : check_target_env
	CF_CLI_EXPERIMENTAL=true CF_TEST_SUITE=./integration/experimental fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ -x

fly-windows-isolated : check_target_env
	CF_TEST_SUITE=./integration/isolated fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ -x

fly-windows-plugin : check_target_env
	CF_TEST_SUITE=./integration/plugin fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ -x

fly-windows-push : check_target_env
	CF_TEST_SUITE=./integration/push fly -t ci execute -c ci/cli/tasks/integration-windows-oneoff.yml -i cli=./ -x

i18n :
	$(PWD)/bin/i18n-checkup

i18n-binary : i18n
	$(PWD)/bin/generate-language-resources

integration-cleanup :
	$(PWD)/bin/cleanup-integration

integration-experimental : build integration-cleanup
	CF_CLI_EXPERIMENTAL=true ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/experimental

integration-isolated : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/isolated

integration-plugin : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/plugin

integration-push : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/push

integration-tests : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/isolated integration/push
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 integration/global
	make integration-cleanup

integration-tests-full : build integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/isolated integration/push integration/plugin integration/experimental
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
	git co cf/resources/i18n_resources.go

test : units

units : format vet i18n build
	ginkgo -r -nodes $(GINKGO_UNIT_TEST_NODES) -randomizeAllSpecs -randomizeSuites \
		api actor command types util
	@echo "\nSWEET SUITE SUCCESS"

units-full : format vet i18n
	@rm -f $(wildcard fixtures/plugins/*.exe)
	@ginkgo version
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes $(GINKGO_UNIT_TEST_NODES) -randomizeAllSpecs -randomizeSuites -skipPackage integration
	@echo "\nSWEET SUITE SUCCESS"

version :
	@echo $(CF_BUILD_VERSION)+$(CF_BUILD_SHA).$(CF_BUILD_DATE)

vet :
	@echo  "Vetting packages for potential issues..."
	@echo  "Be sure to do this prior to committing!!!"
	@git status -s \
		| grep -i -e "^ *N" -e "^ *M" \
		| grep -e api/ -e actor/ -e command -e cf/ -e plugin/ -e util/ \
		| grep -e .go \
		| awk '{print $$2}' \
		| xargs -r -L 1 -P 5 go tool vet -all -shadow=true
	@git status -s \
		| grep -i -e "^ *R" \
		| grep -e api/ -e actor/ -e command -e cf/ -e plugin/ -e util/ \
		| grep -e .go \
		| awk '{print $$4}' \
		| xargs -r -L 1 -P 5 go tool vet -all -shadow=true


.PHONY : all build clean i18n format version vet
.PHONY : test units units-full integration integration-tests-full integration-cleanup integration-experimental integration-plugin integration-isolated integration-push
.PHONY : fly-windows-experimental fly-windows-isolated fly-windows-plugin fly-windows-push
