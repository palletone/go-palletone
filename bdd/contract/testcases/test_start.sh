#!/bin/bash

./setup.sh $1

sleep 1

listAccounts=`../../node/gptn --exec 'personal.listAccounts' attach ../../node/palletone/gptn.ipc`

key=`echo $listAccounts`
echo $key

list=`echo $key | jq ''`;

length=`echo $key | jq 'length'`

num=$[$length - 1]

for index in `seq 0 $num`
do
	if [ $index -gt 5 ]
	then
	account0=`echo $list|jq ".[0]"`
	account1=`echo $list|jq ".[$index]"`
	./transfertoken.sh $account0 $account1
	sleep 3
	fi
#	#echo $list | jq ".[$index]";
done

echo $account0
echo $account1
#调python 脚本，传入account0,account1
./addBalance.sh $account0 $account1

killall gptn
