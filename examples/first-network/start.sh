#!/bin/bash

docker run -it --rm -e MEDIATOR_COUNT=$1 -v $PWD:/go-palletone --entrypoint="./bytn.sh" palletone/gptn

sleep 1

docker-compose up -d
