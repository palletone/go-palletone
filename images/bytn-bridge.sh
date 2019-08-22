#!/bin/bash

cd ./scripts/

#生成5个超级节点和一个普通全节点
./deploy.sh 3

count=1
while [ $count -le 3 ]
do
  #修改配置文件，主要包括配置jury的IsJury = true
  sed -i "s/IsJury = false/IsJury = true/g" node$count/ptn-config.toml
  #配置容器grpc监听地址
  sed -i "s/127.0.0.1/mediator$count/g" node$count/ptn-config.toml
  let ++count
done

