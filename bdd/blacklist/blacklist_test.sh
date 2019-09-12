#!/bin/bash

listAccounts=`../node/gptn --exec 'personal.listAccounts' attach ../node/palletone/gptn.ipc`
accounts=`echo $listAccounts`
echo $accounts

accountsList=`echo $accounts | jq ''`

echo $accountsList

tH=`echo $accountsList | jq ".[0]"`
tokenHolder=`echo ${tH//\"/}`
echo "tokenHolder=" $tokenHolder

o=`echo $accountsList | jq ".[9]"`
one=`echo ${o//\"/}`
echo "one=" $one

t=`echo $accountsList | jq ".[10]"`
two=`echo ${t//\"/}`
echo "two=" $two

pybot -d ../logs/blacklist -v foundation:$tokenHolder -v one:${one} -v two:${two} blacklist.robot
