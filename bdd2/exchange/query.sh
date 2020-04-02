#!/bin/bash
echo $1
pybot -d ../logs/exchange  -v addr:${1} -v twotoken:$2 query.robot

