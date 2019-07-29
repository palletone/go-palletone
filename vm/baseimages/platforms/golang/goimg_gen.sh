#!/bin/bash

cp -r ../../../../contracts/example/go/container/vendor/ .

docker build -t palletone/goimg .

rm -rf vendor
