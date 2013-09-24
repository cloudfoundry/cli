# Dockerfile for go-cf.
# Build the image as,
#   sudo docker build -t cf .
# Then run as,
#   sudo docker run cf

FROM ubuntu

# TODO: remove /cli (~6M) from the image to save on space.
ADD . /cli

ENV GOPATH /cli

# Use ppa:duh/golang as go-cf requires Go >=1.1

RUN apt-get -qy install python-software-properties && \
    add-apt-repository ppa:duh/golang && \
    apt-get -qy update && \
    apt-get -qy install golang git && \
    cd /cli && \
    git submodule update --init && \
    go get -v main && \
    go build -v -o /usr/bin/cf main && \
    apt-get -qy remove golang git python-software-properties && \
    apt-get -qy autoremove && \
    apt-get -qy clean

# Make the home directory a volume for persisting client config.
VOLUME ["/cf-home"]
ENV HOME /cf-home

ENTRYPOINT ["/usr/bin/cf"]
