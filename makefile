GODEPS = $(realpath ./Godeps/_workspace)
GOPATH := $(GODEPS):$(GOPATH)
PATH := $(GODEPS)/bin:$(PATH)

# target: all - Run tests and generate binary
all: test build

# target: help - Display targets
help:
	@egrep "^# target:" [Mm]akefile | sort -

# target: clean - Cleans build artifacts
clean:
	echo Cleaning build artifacts...
	go clean
	echo

# target: generate-language-resource - Creates language resource file
generate-language-resource:
	echo Generating i18n resource file
	go-bindata \
		-pkg resources \
		-ignore ".go" \
		-o cf/resources/i18n_resources.go \
		cf/i18n/resources/...
	echo

# target: fmt - Formats go code
fmt format:
	echo Formatting Packages...
	go fmt ./cf/... ./testhelpers/... ./generic/... ./main/... ./glob/... ./words/... 
	echo

# target: test - Runs CLI tests
test: clean generate-language-resource format
	echo Testing packages:
	LC_ALL="en_US.UTF-8" \
		go test ./cf/... ./generic/... -parallel 4 $(TEST_ARG)
	echo
	$(MAKE) vet

# target: ginkgo - Runs CLI tests with ginkgo command
ginkgo: clean generate-language-resource format
	echo Testing packages:
	LC_ALL="en_US.UTF-8" \
		ginkgo -p cf/* generic
	echo
	$(MAKE) vet

# target: vet - Vets CLI for issues
vet:
	echo Vetting packages for potential issues...
	go tool vet cf/.
	echo

# target: build - Build CLI binary
build: format generate-language-resource
	echo Generating Binary...
	mkdir -p out
	go build -o out/cf ./main
	echo

# target: install-dev-tools - Installs dev tools needed to work on the CLI
install-dev-tools:
	@echo Installing development tools into $(GODEPS)
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega
	go get golang.org/x/tools/cmd/vet
	go get github.com/jteeuwen/go-bindata/...

.PHONY: all help clean generate-language-resource fmt format test ginkgo vet build install-dev-tools
.SILENT: all help clean generate-language-resource fmt format test ginkgo vet build
