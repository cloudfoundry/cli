FROM golang:1.10

RUN go version

# used by coverage utilities
RUN go get golang.org/x/tools/cmd/cover

# used by go-bindata
RUN go get github.com/jteeuwen/go-bindata

# used to set the icon when build binaries for windows
RUN go get github.com/akavel/rsrc

RUN sed -i -e 's/httpredir.debian.org/ftp.us.debian.org/' /etc/apt/sources.list
RUN apt update

# used in CATs tests and building the binaries
RUN apt install -y jq zip make

# used in internationalization tests
RUN apt install -y locales

# install bosh
RUN curl -L https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-linux-amd64 --output /usr/local/bin/bosh && chmod 0755 /usr/local/bin/bosh

# install credhub
RUN curl -L https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/2.0.0/credhub-linux-2.0.0.tgz --output /tmp/credhub.tgz && tar -xzf /tmp/credhub.tgz -C /usr/local/bin && chmod 0755 /usr/local/bin/credhub

# install bbl
RUN curl -L https://github.com/cloudfoundry/bosh-bootloader/releases/download/v6.9.0/bbl-v6.9.0_linux_x86-64 --output /usr/local/bin/bbl && chmod 0755 /usr/local/bin/bbl
