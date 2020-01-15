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
#./createtoken.sh $one 100 $1 1  5000000 
#sleep 5
#lgbalance=`../node/gptn --exec "wallet.getBalance(\"$one\")" attach ../node/palletone/gptn.ipc`
#onetoken=`echo $lgbalance`
#onetoken=`echo $lgbalance|grep $1 | awk -F "[ :]" '{print $2}'`
#echo $onetoken 

#./createtoken.sh $two 100 $2 1  5000000 
#sleep 5
#twbalance=`../node/gptn --exec "wallet.getBalance(\"$two\")" attach ../node/palletone/gptn.ipc`
#twotoken=`echo $lgbalance`
#twotoken=`echo $twbalance|grep $2 | awk -F "[ :]" '{print $2}'`
#echo $twotoken 


pybot -d ../logs/exchange  -v foundation:$tokenHolder    -v Alice:${Alice} -v Bob:${Bob} -v AliceToken:${1} -v BobToken:${2} exchange.robot
#pybot -d ../logs/exchange  -v foundation:$tokenHolder    -v Alice:${Alice} -v Bob:${Bob}  random_action.robot