#!/bin/bash

count=1;
while [ $count -le 5 ];
do 
  docker build -t palletone/mediator$count:$1 ./mediator$count/;
  let ++count;
done

docker build -t palletone/normalnode:$1 ./normalnode/;

count=1;
while [ $count -le 5 ];
do
  rm -rf mediator$count/node$count;
  let ++count;
done

rm -rf normalnode/node_test6;

rm gptn;
