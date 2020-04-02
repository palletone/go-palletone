#!/bin/bash

./test_case_setup.sh $1

sleep 1

listAccounts=`../node/gptn --exec 'personal.listAccounts' attach ../node/palletone/gptn.ipc`
accounts=`echo $listAccounts`
echo $accounts

accountsList=`echo $accounts | jq ''`
echo $accountsList


length=`echo $accountsList | jq 'length'`

num=$[$length - $1]
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

all=$[$addrs - 1]
for index in `seq 0 $all`
do
	another=`echo $accountsList|jq ".[$num + $index]"`
	./transfertoken.sh $tH $another
	sleep 3
#	#echo $list | jq ".[$index]";
done

for index in `seq 0 $all`
do
	toAddr=`echo $accountsList | jq ".[$num + $index]"`
	../node/gptn --exec "personal.unlockAccount($toAddr,'1',0)" attach ../node/palletone/gptn.ipc
done

pybot -d ../logs/blacklist -v foundation:$tokenHolder -v one:${one} -v two:${two} blacklist.robot
