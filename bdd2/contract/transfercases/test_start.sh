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
	fi
#	#echo $list | jq ".[$index]";
done
account2=`echo $"\"PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf"\"`
echo $account0
echo $account1
echo $account2
./transfertoken.sh $account0 $account1
./transferContractAddr.sh $account0 $account2
sleep 3
#调python 脚本，传入account0,account1
account0=`echo $account0 | sed 's/\"//g'`s
account1=`echo $account1 | sed 's/\"//g'`
#echo $account0
#echo $account1
python ./transfer.py  $account1

# killall gptn
