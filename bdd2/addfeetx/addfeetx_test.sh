#!/bin/bash

./test_case_setup.sh 2

sleep 1

listAccounts=`../node/gptn --exec 'personal.listAccounts' attach ../node/palletone/gptn.ipc`
accounts=`echo $listAccounts`
echo $accounts

accountsList=`echo $accounts | jq ''`
echo $accountsList


length=`echo $accountsList | jq 'length'`

num=$[$length - 2]
echo "num=" $num

addrs=$[$length - $num]
echo "addrs=" $addrs

tH=`echo $accountsList | jq ".[0]"`
tokenHolder=`echo ${tH//\"/}`
echo "tokenHolder=" $tokenHolder

o=`echo $accountsList | jq ".[$num]"`
Alice=`echo ${o//\"/}`
echo "Alice=" $Alice

num2=$[$num + 1]
t=`echo $accountsList | jq ".[$num2]"`
Bob=`echo ${t//\"/}`
echo "Bob=" $Bob


pybot -d ../logs/exchange  -v foundation:$tokenHolder    -v Alice:${Alice} -v Bob:${Bob} -v AliceToken:${1} -v BobToken:${2} addfeetx.robot 