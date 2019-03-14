#!/bin/bash

./test_case_setup.sh $1

listAccounts=`./node1/gptn --exec 'personal.listAccounts' attach node1/palletone/gptn.ipc`

key=`echo $listAccounts`
#echo $key

list=`echo $key | jq ''`;

length=`echo $key | jq 'length'`

num=$[$length - 1]

for index in `seq 0 $num`
do
	if [ $index -gt 1 ]
	then
	account0=`echo $list|jq ".[1]"`
	another=`echo $list|jq ".[$index]"`
	./transfertoken.sh $account0 $another
	sleep 1 
	fi
	#echo $list | jq ".[$index]";
done

mediatorAddr_01=`echo $list|jq ".[1]"`

foundationAddr=`echo $list|jq ".[2]"`

./node1/gptn --exec "personal.unlockAccount($mediatorAddr_01,'1',0)" attach node1/palletone/gptn.ipc

for index in `seq 0 $num`
do
	if [ $index -gt 1 ]
	then
	toAddr=`echo $list | jq ".[$index]"`
	./node1/gptn --exec "personal.unlockAccount($toAddr,'1',0)" attach node1/palletone/gptn.ipc
	fi
	#echo $list | jq ".[$index]";
done

one=`echo ${mediatorAddr_01//\"/}`

two=`echo ${foundationAddr//\"/}`
#echo $one

robot -d ./log -v mediatorAddr_01:$one -v foundationAddr:$two --test Business_01 ./pythonTest/depositContract/DepositContractTest.robot

sleep 1

./test_case_teardown.sh

killall gptn
