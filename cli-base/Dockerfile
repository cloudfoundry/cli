FROM golang:latest

# used by coverage utilities
RUN go get golang.org/x/tools/cmd/cover

# used by go-bindata
RUN go get github.com/jteeuwen/go-bindata 

# used to set the icon when build binaries for windows
RUN go get github.com/akavel/rsrc

RUN sed -i -e 's/httpredir.debian.org/ftp.us.debian.org/' /etc/apt/sources.list
RUN apt update

# used in CATs tests
RUN apt install -y jq zip

# used in internationalization tests
RUN apt install -y locales

# install bosh
RUN curl -L https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-2.0.35-linux-amd64 --output /usr/local/bin/bosh && chmod 0755 /usr/local/bin/bosh
