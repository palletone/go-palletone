#!/bin/bash

gptnVersion=`$GOPATH/src/github.com/palletone/go-palletone/build/bin/gptn version | grep ^Version | awk '{print $2}' | awk -F '-' '{print $1}'`

echo "================ gptn version: ${gptnVersion}   ================"

echo "================ palletone/goimg:${gptnVersion} ================"

docker pull palletone/goimg:${gptnVersion}
