#!/bin/bash

./tarPro.sh

docker build -t palletone/goimg:dev .

rm palletone.tar
rm adaptor.tar
