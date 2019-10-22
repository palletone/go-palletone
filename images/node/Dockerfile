# Build Gptn in a stock Go builder container
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

# Download the latest master branch go-palletone project
RUN git clone -b master https://github.com/palletone/go-palletone.git \
    && cd go-palletone \
    && make gptn

# Pull Gptn into a second stage deploy alpine container
FROM alpine:latest

# Maintainer information
MAINTAINER palletone "contract@pallet.one"

RUN apk add --no-cache ca-certificates

# Copy gptn from stock Go builder container
COPY --from=builder /go/go-palletone/build/bin/gptn /usr/local/bin/

# Working in the go-palletone directory
WORKDIR /go-palletone

# Exposing 8545 8546 30303 30303/udp ports
EXPOSE 8545 8546 30303 30303/udp

# Default start command
ENTRYPOINT ["gptn"]
