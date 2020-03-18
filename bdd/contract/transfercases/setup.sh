#!/bin/bash

cp ./createaccount_c.sh  ../../node

cd ../../node

./createaccount_c.sh $1

rm ./createaccount_c.sh

cd ../contract/transfercases

