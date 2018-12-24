FROM golang:1.11

RUN go version

# used by coverage utilities
RUN go get golang.org/x/tools/cmd/cover

# used by go-bindata
RUN go get github.com/jteeuwen/go-bindata

# used to set the icon when build binaries for windows
RUN go get github.com/akavel/rsrc

RUN sed -i -e 's/httpredir.debian.org/ftp.us.debian.org/' /etc/apt/sources.list

# used in CATs tests, building the binaries, and internationalization
RUN apt update && apt install -y jq zip make vim locales

# install bosh
RUN curl -L https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.2.2-linux-amd64 --output /usr/local/bin/bosh && chmod 0755 /usr/local/bin/bosh

# install credhub
RUN curl -L https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/2.0.0/credhub-linux-2.0.0.tgz --output /tmp/credhub.tgz && tar -xzf /tmp/credhub.tgz -C /usr/local/bin && chmod 0755 /usr/local/bin/credhub

# install bbl
RUN curl -L https://github.com/cloudfoundry/bosh-bootloader/releases/download/v6.9.16/bbl-v6.9.16_linux_x86-64 --output /usr/local/bin/bbl && chmod 0755 /usr/local/bin/bbl

# install bbl v5.6.11 for legacy environment
RUN curl -L https://github.com/cloudfoundry/bosh-bootloader/releases/download/v5.11.6/bbl-v5.11.6_linux_x86-64 --output /usr/local/bin/bbl5 && chmod 0755 /usr/local/bin/bbl5
