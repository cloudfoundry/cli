FROM golang

RUN go get golang.org/x/tools/cmd/vet
RUN go get golang.org/x/tools/cmd/cover

RUN apt-get update && apt-get -y install s3cmd

RUN apt-get update && apt-get -y install wine
