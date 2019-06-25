#!/bin/bash

#./preset.sh
#sleep 25

listcommand=`./gptn --exec 'personal.listAccounts'  attach node1/palletone/gptn.ipc`
list=`echo ${listcommand//^M/}`
moneyaccount=`echo "${list}" | jq -c '.[1]'`

newaccount5command=`./gptn --exec 'personal.newAccount("1")'  attach node_test5/palletone/gptn5.ipc`
a5=`echo ${newaccount5command//^M/}`
account5=`echo $a5`

newaccount6command=`./gptn --exec 'personal.newAccount("1")'  attach node_test6/palletone/gptn6.ipc`
a6=`echo ${newaccount6command//^M/}`
account6=`echo $a6`

newaccount7command=`./gptn --exec 'personal.newAccount("1")'  attach node_test7/palletone/gptn7.ipc`
a7=`echo ${newaccount7command//^M/}`
account7=`echo $a7`

echo $moneyaccount
echo $account5
echo $account6
echo $account7

startProduce=`./gptn --exec 'mediator.startProduce()'  attach node1/palletone/gptn.ipc`
echo $startProduce

#transfer to light account5
curl -H "Content-Type:application/json" -X POST -d  "{\"jsonrpc\":\"2.0\",\"method\":\"wallet_transferToken\",\"params\":[\"PTN\",$moneyaccount,$account5,\"100\",\"1\",\"1\",\"1\"],\"id\":1}" http://127.0.0.1:8545

sleep 5
balancecommand=`curl -H "Content-Type:application/json" -X POST -d "{\"jsonrpc\":\"2.0\",\"method\":\"wallet_getBalance\",\"params\":[$account5],\"id\":1}" http://127.0.0.1:8585`
balanceinfo=`echo $balancecommand`
balance=`echo $balanceinfo | sed -n '$p'`
b1=`echo $balance | jq '.result.PTN'`
b2=`echo $b1`
echo $b2
num="\"100\""
echo $num
if [ "$b2" = "$num" ];then 
    echo "============transferToken1 ok num"$b2"============"
else
    echo "============transferToken1 err num:"$b2"============"
fi

