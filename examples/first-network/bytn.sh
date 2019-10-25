#!/bin/bash

cp /usr/local/bin/gptn /go-palletone/scripts

export DEBIAN_FRONTEND=noninteractive

apt-get update -q

apt-get install -y tcl tk jq expect

cd scripts

#生成 $1 个超级节点和 2 个普通全节点
./deploy.sh $MEDIATOR_COUNT

count=1

while [ $count -le $MEDIATOR_COUNT ]
do
  if [ $count == 1 ]; then
    newContractAddress="ContractAddress=\"mediator1:12345\""
    sed -i '/^ContractAddress/c'$newContractAddress'' node1/ptn-config.toml
  fi
  #修改容器gprc监听地址
  sed -i "s/127.0.0.1/mediator$count/g" node$count/ptn-config.toml
  let ++count
done
