FROM ubuntu

MAINTAINER palletone "contract@pallet.one"

RUN apt-get update \
    && apt-get install -y wget \
       git \
       gcc

#INSTALL Golang ENVIRONMENT
RUN wget https://studygolang.com/dl/golang/go1.10.linux-amd64.tar.gz \
    && tar -C /usr/local -zxvf go1.10.linux-amd64.tar.gz \
    && rm go1.10.linux-amd64.tar.gz \
    && mkdir -p $HOME/gopath/bin $HOME/gopath/src $HOME/gopath/pkg /chaincode/input

ENV GOPATH=/gopath  PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/bin:/usr/local/go/bin/

#INSTALL THE DEPENDENCY OF CHAINCODE RUNNING
RUN go get -v -u github.com/palletone/go-palletone

