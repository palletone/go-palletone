#!/bin/bash

cd ./scripts/

#生成5个超级节点和一个普通全节点
./deploy.sh 5

count=1

while [ $count -le 5 ]
do
  if [ $count == 1 ]; then
    newContractAddress="ContractAddress=\"mediator1:12345\""
    sed -i '/^ContractAddress/c'$newContractAddress'' node1/ptn-config.toml
  fi
  #修改容器gprc监听地址
  sed -i "s/127.0.0.1/mediator$count/g" node$count/ptn-config.toml
  let ++count
done

cd ..

#拉取官方提供的镜像文件
docker pull palletone/gptn:latest
