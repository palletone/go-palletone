#!#!/bin/bash

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
one=`echo ${o//\"/}`
echo "one=" $one

num2=$[$num + 1]
t=`echo $accountsList | jq ".[$num2]"`
two=`echo ${t//\"/}`
echo "two=" $two

#all=$[$addrs - 1]
#for index in `seq 0 $all`
#do
#	another=`echo $accountsList|jq ".[$num + $index]"`
	#./transfertoken.sh $tH $another
#	sleep 3
#done
./transfertoken.sh $tH $one
sleep 3
./transfertoken.sh $tH $two 
sleep 3

for index in `seq 0 $all`
do
	toAddr=`echo $accountsList | jq ".[$num + $index]"`
	../node/gptn --exec "personal.unlockAccount($toAddr,'1',0)" attach ../node/palletone/gptn.ipc
done

./createtoken.sh $one 100 $1 1  5000000 
sleep 5
lgbalance=`../node/gptn --exec "wallet.getBalance(\"$one\")" attach ../node/palletone/gptn.ipc`
onetoken=`echo $lgbalance`
onetoken=`echo $lgbalance|grep $1 | awk -F "[ :]" '{print $2}'`
echo $onetoken 

./createtoken.sh $two 100 $2 1  5000000 
sleep 5
twbalance=`../node/gptn --exec "wallet.getBalance(\"$two\")" attach ../node/palletone/gptn.ipc`
twotoken=`echo $lgbalance`
twotoken=`echo $twbalance|grep $2 | awk -F "[ :]" '{print $2}'`
echo $twotoken 


pybot -d ../logs/exchange  -v one:${one} -v two:${two} -v onetoken:$onetoken -v twotoken:$twotoken exchange.robot
