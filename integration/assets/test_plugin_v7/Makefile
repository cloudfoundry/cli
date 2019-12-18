GOSRC = $(shell find . -name "*.go" ! -name "*test.go" ! -name "*fake*" ! -path "./integration/*")

LD_FLAGS_LINUX_V7 = -extldflags \"-static\"

export GOFLAGS = -tags=V7

build: out/test_plugin_v7

clean:
	rm -rf out

out/test_plugin_v7: $(GOSRC)
	CGO_ENABLED=0 go build \
		-tags "V7" \
		-ldflags "$(LD_FLAGS_LINUX_V7)" -o out/test_plugin_v7 .

install: build
	cf install-plugin -f out/test_plugin_v7
	cf plugins
