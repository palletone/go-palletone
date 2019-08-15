#!/bin/bash

cd tardata

rm -rf *

wget $1

cd ..

docker build --no-cache -t palletone/jurynode:1.0.1 .

rm -rf tardata/*
