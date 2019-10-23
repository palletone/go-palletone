# Build Gptn in a stock Go builder container
FROM golang:1.13 as builder

RUN apt-get update -q && apt-get install -y jq expect

# Download the latest master branch go-palletone project
RUN mkdir -p src/github.com/palletone \
    && cd src/github.com/palletone \
    && git clone -b xiaozhi https://github.com/palletone/go-palletone.git \
    && cd go-palletone \
    && make gptn \
    && cp build/bin/gptn images/scripts

# Pull Gptn into a second stage deploy alpine container
FROM ubuntu:latest

# Maintainer information
MAINTAINER palletone "contract@pallet.one"

RUN apt-get update -q && apt-get install -y jq expect

# Copy gptn from stock Go builder container
COPY --from=builder /go/src/github.com/palletone/go-palletone /go-palletone

# Working in the go-palletone directory
WORKDIR /go-palletone
