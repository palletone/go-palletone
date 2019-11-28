# Build palletone golang image in a stock Go builder container
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

# Download the latest adaptor project
RUN go get -u -v github.com/palletone/adaptor

# Download the latest master branch go-palletone project
RUN mkdir -p src/github.com/palletone \
    && cd src/github.com/palletone \
    && git clone -b master https://github.com/palletone/go-palletone.git

# Download the latest govendor project
RUN go get -u -v github.com/kardianos/govendor

# Get the package dependencies needed in the palletone golang image
RUN cd src/github.com/palletone/go-palletone/contracts/example/go/container \
    && govendor init \
    && govendor add +e \
    && rm vendor/github.com/palletone/adaptor/*_mock.go

# Pull all package dependencies into a second stage deploy palletone baseimg container
FROM palletone/baseimg

# Maintainer information
MAINTAINER palletone "contract@pallet.one"

# Download and configure golang:1.12
RUN wget -o download.log https://studygolang.com/dl/golang/go1.12.linux-amd64.tar.gz \
    && tar -C /usr/local -zxvf go1.12.linux-amd64.tar.gz >> download.log \
    && rm go1.12.linux-amd64.tar.gz download.log \
    && mkdir -p /gopath/bin /gopath/src /gopath/pkg /chaincode/input

# Set ENV
ENV GOPATH=/gopath
ENV GOROOT=/usr/local/go
ENV PATH=$PATH:$GOPATH/bin:$GOROOT/bin

# Copy all package dependencies from stock Go builder container
COPY --from=builder /go/src/github.com/palletone/go-palletone/contracts/example/go/container/vendor /gopath/src/
