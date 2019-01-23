#!/bin/bash

#read -p "Enter base image tag:" tag

docker build -t palletone/goimg .

echo "Download the base image: palletone/goimg successfully"

exit 0
