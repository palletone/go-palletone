#!/bin/bash

cd tardata

#下载主网gptn和leveldb
wget $2

cd ..

#构建镜像
docker build --no-cache -t palletone/mainnetnode:$1 .

rm -rf tardata/*
