FROM alpine:latest
RUN apk update && \
    apk add curl && \
    curl --location --silent "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=v8&source=github" | tar -zx -C /usr/local/bin && \
    cf --version
