#!/bin/bash

cp ./createaccount.sh ./createaccounts.sh ../node

cd ../node

./createaccounts.sh $1

rm ./createaccount.sh ./createaccounts.sh

cd ../dct
