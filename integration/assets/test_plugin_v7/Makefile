GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")

REQUIRED_FOR_STATIC_BINARY_V7 =-a -tags "V7"
LD_FLAGS_LINUX_V7 = -extldflags \"-static\"

export GOFLAGS = -tags=V7


build: out/test_plugin_v7

clean:
	rm -f $(wildcard out/test_plugin*)

out/test_plugin_v7: $(GOSRC)
	CGO_ENABLED=0 go build \
	  $(REQUIRED_FOR_STATIC_BINARY_V7) \
	  -ldflags "$(LD_FLAGS_LINUX_V7)" -o out/test_plugin_v7 .

