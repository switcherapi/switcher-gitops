ARG GO_VERSION=1.24.3
ARG ALPINE_VERSION=3.21

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION}

# Install packages
RUN apk add -q --update --progress --no-cache git make openssh-client sudo