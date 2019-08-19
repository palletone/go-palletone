#!/bin/bash

cd ..

#构建一个节点
./deploy.sh 1

#移动到镜像构建目录下
mv node1 mediator

cd -

#构建镜像
docker build -t palletone/mediator:$1 .

rm -rf node1

rm ../gptn
