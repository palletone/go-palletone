#!/bin/bash

cd ..

./deploy.sh 1

mv node1 mediator

cd -

docker build -t palletone/mediator .

rm -rf node1
