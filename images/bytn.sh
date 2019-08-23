#!/bin/bash

cd ./scripts/

#生成5个超级节点和一个普通全节点
./deploy.sh 5 

count=1
while [ $count -le 5 ]
do
  #修改配置文件，主要配置jury的IsJury = true
  sed -i "s/IsJury = false/IsJury = true/g" node$count/ptn-config.toml
  #修改容器gprc监听地址
  sed -i "s/127.0.0.1/mediator$count/g" node$count/ptn-config.toml
  let ++count
done

#如果本地没有gptn-net这个网络，则创建
docker network create gptn-net

cd ..

exit 0

#拉去官方提供的镜像文件
docker pull palletone/mediator:1.0.1

docker tag palletone/mediator:1.0.1 palletone/mediator

#需要手动修改各个节点的toml配置文件，将 VmEndpoint 的值修改为 宿主机ip:2375 
如： VmEndpoint = "192.168.152.128:2375"
#使用docker-compose up -d 启动节点对应的容器
docker-compose up -d
