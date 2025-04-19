ARG GO_VERSION=1.24.2
ARG ALPINE_VERSION=3.21

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION}

# Install packages
RUN apk add -q --update --progress --no-cache git sudo openssh-client make