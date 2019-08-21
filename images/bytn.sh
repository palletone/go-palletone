#!/bin/bash

cd ./scripts/

#生成5个超级节点和一个普通全节点
./deploy.sh 5 

count=1
while [ $count -le 5 ]
do
  #修改配置文件，主要包括配置jury的IsJury = true，以及jury节点监听的docker服务
  sed -i "s/IsJury = false/IsJury = true/g" node$count/ptn-config.toml
  sed -i "s/unix:\/\/\/var\/run\/docker.sock/tcp:\/\/0.0.0.0:2375/g" node$count/ptn-config.toml
  let ++count
done

cd ..
exit 0
docker pull palletone/mediator:$1

docker tag palletone/mediator:$1 palletone/mediator

docker-compose up -d

