#!/bin/bash

cp ./createaccount.sh  ../../node

cd ../../node

./createaccount.sh $1

rm ./createaccount.sh

cd ../contract/testcases