FROM golang:latest

# used by coverage utilities
RUN go get golang.org/x/tools/cmd/cover

# used to set the icon when build binaries for windows
RUN go get github.com/akavel/rsrc

RUN sed -i -e 's/httpredir.debian.org/ftp.us.debian.org/' /etc/apt/sources.list
RUN apt update

# used in CATs tests
RUN apt install -y zip
