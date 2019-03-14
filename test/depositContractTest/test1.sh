#!/bin/bash
listAccounts=`./gptn --exec 'personal.listAccounts' attach node1/palletone/gptn.ipc`
key=`echo $listAccounts`
#echo $key

list=`echo $key | jq ''`;
length=`echo $key | jq 'length'`
num=$[$length - 1]
#for index in `seq 0 $num`
#do
#	echo $list | jq ".[$index]";
#done


account0=`echo $list|jq ".[1]"`
#echo $account0
account1=`echo $list|jq ".[0]"`
./gptn --exec "wallet.transferToken('PTN',$account0,$account1,'5000','10')" attach node1/palletone/gptn.ipc




