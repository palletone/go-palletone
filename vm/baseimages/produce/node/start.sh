#!/bin/bash

cd tardata

#下载测试网gptn及leveldb
wget $2

cd ..

#构建测试网节点镜像
docker build --no-cache -t palletone/gptnnode:$1 .


rm -rf tardata/*
