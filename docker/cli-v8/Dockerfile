FROM alpine:latest

RUN apk update && \
    apk add --upgrade git bash go make && \ 
    mkdir workspace

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN cd workspace && \
    git clone https://github.com/cloudfoundry/cli.git && \
    cd cli && \
    make out/cf-cli_linux_x86-64
