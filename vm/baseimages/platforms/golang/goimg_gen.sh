#!/bin/bash

rm ../../../../contracts/example/go/container/vendor/github.com/palletone/adaptor/*_mock.go

cp -r ../../../../contracts/example/go/container/vendor/ .

docker build -t palletone/goimg .

rm -rf vendor

rm -rf ../../../../contracts/example/go/container/vendor
