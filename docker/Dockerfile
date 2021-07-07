FROM alpine:latest
ARG VERSION
RUN apk update && \
    apk add curl && \
    curl --location --silent "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=${VERSION}&source=github" | tar -zx -C /usr/local/bin && \
    cf --version
