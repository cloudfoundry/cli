FROM golang:latest

RUN go get golang.org/x/tools/cmd/cover
RUN go get github.com/akavel/rsrc

RUN sed -i -e 's/httpredir.debian.org/ftp.us.debian.org/' /etc/apt/sources.list
RUN apt update
RUN apt install -y zip
