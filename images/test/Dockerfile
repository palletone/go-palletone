FROM middle1 as builder

RUN apt-get update -q \
#    && apt-get install -y jq tcl tk expect make gcc git \
#    && apt-get install wget -y \
    && wget https://studygolang.com/dl/golang/go1.12.linux-amd64.tar.gz \
    && tar -C /usr/local -zxvf go1.12.linux-amd64.tar.gz \
    && mkdir -p /gopath/bin /gopath/src /gopath/pkg \
    && export GOROOT=/usr/local/go \
    && export GOPATH=/gopath \
    && export PATH=$PATH:$GOROOT/bin:$GOPATH/bin \
    && mkdir -p /gopath/src/github.com/palletone \
    && cd /gopath/src/github.com/palletone \
    && git clone -b xiaozhi https://github.com/palletone/go-palletone.git \
    && cd go-palletone/cmd/gptn \
    && go build \
    && cp gptn ../../images/scripts \
    && cd ../../images \
    && ./bytn.sh \
    && cd scripts \
    && mkdir result \
    && mv node* result


## Build Gptn in a stock Go builder container
#FROM golang:1.13 as builder
#
#RUN apt-get update -q && apt-get install -y jq tcl tk expect
#
## Download the latest master branch go-palletone project
#RUN mkdir -p src/github.com/palletone \
#    && cd src/github.com/palletone \
#    && git clone -b xiaozhi https://github.com/palletone/go-palletone.git \
#    && cd go-palletone \
#    && make gptn \
#    && cd images \
#    && ./bytn.sh \
#    && cd scripts \
#    && mkdir result \
#    && mv node* result
#
## Pull Gptn into a second stage deploy alpine container
#FROM ubuntu:latest
#
## Maintainer information
#MAINTAINER palletone "contract@pallet.one"
#
## Copy gptn from stock Go builder container
##COPY --from=builder /go/go-palletone/build/bin/gptn /usr/local/bin/
#COPY --from=builder /go/src/github.com/palletone/go-palletone/images/scripts/result /go-palletone
#
## Working in the go-palletone directory
#WORKDIR /go-palletone
#
## Exposing 8545 8546 30303 30303/udp ports
##EXPOSE 8545 8546 30303 30303/udp
##
### Default start command
##ENTRYPOINT ["gptn"]
