#!/bin/bash


count=1;

rm -rf mediator*/node*;

rm -rf normalnode/node_test6

while [ $count -le 5 ];
do
  sed -i "s/IsJury = false/IsJury = true/g" node$count/ptn-config.toml
  sed -i "s/unix:\/\/\/var\/run\/docker.sock/tcp:\/\/0.0.0.0:2375/g" node$count/ptn-config.toml
  mv node$count mediator$count;
  let ++count;
done 

mv node_test6 normalnode;
