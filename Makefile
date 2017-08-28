CF_DIAL_TIMEOUT ?= 15
GINKGO_UNIT_TEST_NODES ?= 4
GINKGO_INTEGRATION_TEST_NODES ?= 4
LC_ALL = "en_US.UTF-8"

CF_BUILD_VERSION ?= $$(cat ci/VERSION)
CF_BUILD_SHA ?= $$(git rev-parse --short HEAD)
CF_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
GOSRC = $(shell find . -name "*.go")

all : test internationalization-binary build

build : out/cf

clean:
	rm $(wildcard out/*)

format :
	go fmt ./...

internationalization :
	$(PWD)/bin/i18n-checkup

internationalization-binary : internationalization
	$(PWD)/bin/generate-language-resources

integration-cleanup:
	$(PWD)/bin/cleanup-integration

integration-tests: integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/isolated integration/push
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 integration/global
	make integration-cleanup

integration-tests-full: integration-cleanup
	ginkgo -r -randomizeAllSpecs -slowSpecThreshold 60 -nodes $(GINKGO_INTEGRATION_TEST_NODES) integration/isolated integration/push integration/plugin
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

units : format vet internationalization build
	ginkgo -r -nodes $(GINKGO_UNIT_TEST_NODES) -randomizeAllSpecs -randomizeSuites \
		api actor command types util
	@echo "\nSWEET SUITE SUCCESS"

units-full : format vet internationalization
	@rm -f $(wildcard fixtures/plugins/*.exe)
	@ginkgo version
	CF_HOME=$(PWD)/fixtures ginkgo -r -nodes $(GINKGO_UNIT_TEST_NODES) -randomizeAllSpecs -randomizeSuites -skipPackage integration
	@echo "\nSWEET SUITE SUCCESS"

version :
	@echo $(CF_BUILD_VERSION)+$(CF_BUILD_SHA).$(CF_BUILD_DATE)

vet :
	@echo  "Vetting packages for potential issues..."
	@echo  "Be sure to do this prior to commiting!!!"
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


.PHONY: all build clean internationalization format version vet
.PHONY: test units units-full integration integration-tests-full integration-cleanup
