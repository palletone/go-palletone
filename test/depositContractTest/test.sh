#!/bin/bash
listAccounts=`gptn --exec 'personal.listAccounts' attach palletone/gptn.ipc`
key=`echo $listAccounts`
#echo $key

list=`echo $key | jq ''`;
length=`echo $key | jq 'length'`
num=$[$length - 1]
for index in `seq 0 $num`
do
	if [ $index -gt 5 ]
	then
	account0=`echo $list|jq ".[0]"`
	another=`echo $list|jq ".[$index]"`
	./transfertoken.sh $account0 $another
	sleep 3
	fi
	#echo $list | jq ".[$index]";
done

mediatorAddr_01=`echo $list|jq ".[0]"`
foundationAddr=`echo $list|jq ".[6]"`
gptn --exec "personal.unlockAccount($mediatorAddr_01,'1',0)" attach palletone/gptn.ipc
for index in `seq 0 $num`
do
	if [ $index -gt 5 ]
	then
	sleep 3
	toAddr=`echo $list | jq ".[$index]"`
	gptn --exec "personal.unlockAccount($toAddr,'1',0)" attach palletone/gptn.ipc
	fi
	#echo $list | jq ".[$index]";
done
one=`echo ${mediatorAddr_01//\"/}`
two=`echo ${foundationAddr//\"/}`
#echo $one
robot -v mediatorAddr_01:$one -v foundationAddr:$two --test Business_01 ./pythonTest/depositContract/DepositContractTest.robot


