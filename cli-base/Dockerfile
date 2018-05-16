FROM golang:1.9

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
