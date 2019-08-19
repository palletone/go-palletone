#!/bin/bash

cd ./scripts/

./deploy.sh 5

count=1
while [ $count -le 5 ]
do
  sed -i "s/IsJury = false/IsJury = true/g" node$count/ptn-config.toml
  sed -i "s/unix:\/\/\/var\/run\/docker.sock/tcp:\/\/0.0.0.0:2375/g" node$count/ptn-config.toml
  let ++count
done

cd ..

docker pull palletone/mediator:$1

docker tag palletone/mediator:$1 palletone/mediator

docker-compose up -d

