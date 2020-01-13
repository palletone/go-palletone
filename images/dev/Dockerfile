# Build palletone golang image
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
COPY ./vendor /gopath/src/
