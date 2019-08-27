#!/bin/bash 
tar -czvf ./vm/baseimages/dev/palletone.tar . --exclude=.git --exclude=bdd --exclude=wallet --exclude=vm --exclude=ptnjson --exclude=test --exclude=ptnclient --exclude=ptnjson --exclude=ptn --exclude=light --exclude=internal --exclude=images --exclude=examples --exclude=consensus --exclude=cmd --exclude=build --exclude=adaptor > mr.log 2>&1
tar -czvf ./vm/baseimages/dev/adaptor.tar ../adaptor --exclude=.git  > mr.log 2>&1

rm mr.log
