FROM golang:1.13.15-alpine

RUN apk update && \
    apk add --upgrade git bash make && \
    mkdir workspace

RUN cd workspace && \
    git clone https://github.com/cloudfoundry/cli.git && \
    cd cli && \
    make build

ENV PATH /go/workspace/cli/out:$PATH

WORKDIR /

RUN cf --version
