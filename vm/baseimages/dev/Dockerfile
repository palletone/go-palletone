FROM palletone/baseimg

MAINTAINER palletOne "contract@pallet.one"

#INSTALL Golang ENVIRONMENT
RUN wget -o download.log https://studygolang.com/dl/golang/go1.10.linux-amd64.tar.gz \
    && tar -C /usr/local -zxvf go1.10.linux-amd64.tar.gz >> download.log \
    && rm go1.10.linux-amd64.tar.gz download.log \
    && mkdir -p /gopath/bin /gopath/src /gopath/pkg /chaincode/input
ENV GOPATH=/gopath  PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/:$GOPATH/bin

#INSTALL THE DEPENDENCY OF CHAINCODE RUNNING
#RUN go get -u github.com/palletone/go-palletone
ADD palletone.tar $GOPATH/src/github.com/palletone/go-palletone/
ADD adaptor.tar $GOPATH/src/github.com/palletone/
