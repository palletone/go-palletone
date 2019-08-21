#!/bin/bash

cd ./scripts/

#生成5个超级节点和一个普通全节点
./deploy.sh 5 

local_host=`/sbin/ifconfig -a | grep inet | grep -v 127.0.0.1 | grep -v inet6 | grep -v 172.17.0.1 | grep -v 172 | awk '{print $2}' | tr -d "addr:"`

echo "local_host" $local_host

count=1
while [ $count -le 5 ]
do
  #修改配置文件，主要包括配置jury的IsJury = true，以及jury节点监听的docker服务
  sed -i "s/IsJury = false/IsJury = true/g" node$count/ptn-config.toml
  sed -i "s/unix:\/\/\/var\/run\/docker.sock/tcp:\/\/$local_host:2375/g" node$count/ptn-config.toml
  sed -i "s/127.0.0.1/172.11.0.$[$count+1]/g" node$count/ptn-config.toml
  let ++count
done

cd ..

docker-compose -f docker-compose-update.yml up -d

exit 0

cd ..

docker pull palletone/mediator:1.0.1

docker tag palletone/mediator:1.0.1 palletone/mediator

docker-compose -f docker-compose-update.yml up -d

