#!/bin/bash

cd $GOPATH/src/github.com/palletone/go-palletone/contracts/example/go/container/

govendor init

govendor add +e

cd -

rm ../../../../contracts/example/go/container/vendor/github.com/palletone/adaptor/*_mock.go

cp -r ../../../../contracts/example/go/container/vendor/ .

docker build --no-cache -t palletone/goimg:$1 .

rm -rf vendor

rm -rf ../../../../contracts/example/go/container/vendor
