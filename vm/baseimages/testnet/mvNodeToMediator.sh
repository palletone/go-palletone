#!/bin/bash

count=1;

rm -rf mediator*/node*;

rm -rf normalnode/node_test6

while [ $count -le 5 ];
do
  mv node$count mediator$count;
  let ++count;
done 

mv node_test6 normalnode;
