#!/usr/bin/env bash
#cd /home/travis/gopath/src/github.com/palletone/go-palletone
cd ../../
go build ./cmd/gptn
cp ./cmd/gptn/gptn ./bdd/GasToken/node/
cd ./bdd/GasToken/node
chmod +x gptn

cd ../pylibs/
python init_chain.py

cd ../../../cmd/gptn
nohup ./gptn &