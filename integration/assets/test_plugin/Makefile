GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")

LD_FLAGS_LINUX = -extldflags \"-static\"

build: out/test_plugin

clean:
	rm -rf out

out/test_plugin: $(GOSRC)
	CGO_ENABLED=0 go build \
	  -ldflags "$(LD_FLAGS_LINUX)" -o out/test_plugin .

install: build
	cf install-plugin -f out/test_plugin
	cf plugins
