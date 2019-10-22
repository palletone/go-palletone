# Build goimage in a stock Go builder container
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

RUN go get -u -v github.com/palletone/adaptor

RUN mkdir -p src/github.com/palletone \
    && cd src/github.com/palletone \
    && git clone -b master https://github.com/palletone/go-palletone.git

RUN go get -u -v github.com/kardianos/govendor

RUN cd src/github.com/palletone/go-palletone/contracts/example/go/container \
    && govendor init \
    && govendor add +e \
    && rm vendor/github.com/palletone/adaptor/*_mock.go

FROM palletone/baseimg

MAINTAINER palletone "contract@pallet.one"

RUN wget -o download.log https://studygolang.com/dl/golang/go1.12.linux-amd64.tar.gz \
    && tar -C /usr/local -zxvf go1.12.linux-amd64.tar.gz >> download.log \
    && rm go1.12.linux-amd64.tar.gz download.log \
    && mkdir -p /gopath/bin /gopath/src /gopath/pkg /chaincode/input

COPY --from=builder /go/src/github.com/palletone/go-palletone/contracts/example/go/container/vendor /gopath/src/
