# Build Gptn in a stock Go builder container
FROM golang:1.13 as builder

# Download the latest master branch go-palletone project
#RUN git clone -b master https://github.com/palletone/go-palletone.git \
#    && cd go-palletone \
#    && make gptn

# Download the latest master branch go-palletone project
RUN mkdir -p src/github.com/palletone \
    && cd src/github.com/palletone \
    && git clone -b master https://github.com/palletone/go-palletone.git \
    && cd go-palletone/cmd/gptn \
    && go build -mod=vendor

# Pull Gptn into a second stage deploy ubuntu container
FROM ubuntu:latest

# Maintainer information
MAINTAINER palletone "contract@pallet.one"

RUN mkdir /go-palletone \
    && apt-get -y update \
    && apt-get install -yqq expect

# Working in the go-palletone directory
WORKDIR /go-palletone

COPY --from=builder /go/src/github.com/palletone/go-palletone/cmd/gptn/gptn .
COPY init.sh .
COPY newgenesis.sh .
COPY gptn-entrypoint.sh .

RUN chmod +x *.sh

# Exposing 8545 8546 30303 30303/udp ports
EXPOSE 8545

# Default start command
ENTRYPOINT ["./gptn-entrypoint.sh"]

