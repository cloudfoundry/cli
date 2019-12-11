GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")

LD_FLAGS_LINUX_V7 = -extldflags \"-static\"


build: out/test_plugin

clean:
	rm -f $(wildcard out/test_plugin*)

out/test_plugin: $(GOSRC)
	CGO_ENABLED=0 go build \
	  $(REQUIRED_FOR_STATIC_BINARY_V7) \
	  -ldflags "$(LD_FLAGS_LINUX_V7)" -o out/test_plugin .

install: build
	cf install-plugin -f out/test_plugin
	cf plugins
