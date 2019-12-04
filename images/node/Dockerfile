# Build Gptn in a stock Go builder container
FROM golang:1.13 as builder

# Download the latest master branch go-palletone project
#RUN git clone -b master https://github.com/palletone/go-palletone.git \
#    && cd go-palletone \
#    && make gptn

# Download the latest master branch go-palletone project
RUN mkdir -p src/github.com/palletone \
    && cd src/github.com/palletone \
    && git clone -b testnet https://github.com/palletone/go-palletone.git \
    && cd go-palletone/cmd/gptn \
    && go build -mod=vendor -tags mainnet

# Pull Gptn into a second stage deploy ubuntu container
FROM ubuntu:latest

# Maintainer information
MAINTAINER palletone "contract@pallet.one"

RUN mkdir /go-palletone

# Copy gptn from stock Go builder container
#COPY --from=builder /go/go-palletone/build/bin/gptn /usr/local/bin/
COPY --from=builder /go/src/github.com/palletone/go-palletone/cmd/gptn/gptn /usr/local/bin/

# Working in the go-palletone directory
WORKDIR /go-palletone

# Exposing 8545 8546 30303 30303/udp ports
EXPOSE 8545 8546 30303 30303/udp

# Default start command
ENTRYPOINT ["gptn"]
