#!/bin/bash

cd $GOPATH/src/github.com/palletone/go-palletone/contracts/example/go/container/

govendor init

govendor add +e

cd $GOPATH/src/github.com/palletone/go-palletone/images/dev/

rm $GOPATH/src/github.com/palletone/go-palletone/contracts/example/go/container/vendor/github.com/palletone/adaptor/*_mock.go

cp -r $GOPATH/src/github.com/palletone/go-palletone/contracts/example/go/container/vendor/ .

rm -rf $GOPATH/src/github.com/palletone/go-palletone/contracts/example/go/container/vendor

gptnVersion=`$GOPATH/src/github.com/palletone/go-palletone/build/bin/gptn version | grep ^Version | awk '{print $2}' | awk -F '-' '{print $1}'`

docker build --no-cache -t palletone/goimg:${gptnVersion} .

rm -rf vendor
