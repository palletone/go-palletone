#!/bin/bash

./preset.sh
sleep 40

listcommand=`./gptn --exec 'personal.listAccounts'  attach node1/palletone/gptn.ipc`
list=`echo ${listcommand///}`
moneyaccount=`echo "${list}" | jq -c '.[1]'`

newaccount5command=`./gptn --exec 'personal.newAccount("1")'  attach node_test5/palletone/gptn5.ipc`
a5=`echo ${newaccount5command///}`
account5=`echo $a5`

newaccount6command=`./gptn --exec 'personal.newAccount("1")'  attach node_test6/palletone/gptn6.ipc`
a6=`echo ${newaccount6command///}`
account6=`echo $a6`

newaccount7command=`./gptn --exec 'personal.newAccount("1")'  attach node_test7/palletone/gptn7.ipc`
a7=`echo ${newaccount7command///}`
account7=`echo $a7`

echo $moneyaccount
echo $account5
echo $account6
echo $account7

startProduce=`./gptn --exec 'mediator.startProduce()'  attach node1/palletone/gptn.ipc`
echo $startProduce

#full account transferPTN to light account5
curl -H "Content-Type:application/json" -X POST -d  "{\"jsonrpc\":\"2.0\",\"method\":\"wallet_transferToken\",\"params\":[\"PTN\",$moneyaccount,$account5,\"100\",\"1\",\"1\",\"1\"],\"id\":1}" http://127.0.0.1:8545

sleep 5


#ptn.syncUTXOByAddr("P19wzjSAfVKRY84pPQMsqJSxeVK7oTYEiXt") in light node5
syncutxocommand=`./gptn --exec "ptn.syncUTXOByAddr($account5)"  attach node_test5/palletone/gptn5.ipc`
syncutxoinfo=`echo $syncutxocommand`
value="\"ok\""
echo $syncutxoinfo
if [ $syncutxoinfo = $value ];then
    echo "============syncUTXOByAddr account5 ok============"
    python  -m robot.run -d ./logs --test success ./light.robot
else
    echo "============syncUTXOByAddr account5 err:"$syncutxoinfo
    python  -m robot.run -d ./logs --test fail ./light.robot
fi

#ptn.getBalance in light node5
balancecommand=`./gptn --exec "wallet.getBalance($account5)"  attach node_test5/palletone/gptn5.ipc`
balanceinfo=`echo $balancecommand`
temp=`echo ${balanceinfo:7}`
length=`echo ${#temp}`
num=$[$lenght-2]
t1=`echo ${temp:0:$num} | sed 's/ //g' | sed 's/"//g'`
if [ $t1 = 100 ];then
    echo "============getBalance account5 ok============"
    python  -m robot.run -d ./logs --test success ./light.robot
else
    echo "============getBalance account5 err:"$t1
    python  -m robot.run -d ./logs --test fail ./light.robot
fi



# transferPTN in light node5 node6
curl -H "Content-Type:application/json" -X POST -d  "{\"jsonrpc\":\"2.0\",\"method\":\"wallet_transferToken\",\"params\":[\"PTN\",$account5,$account6,\"80\",\"1\",\"1\",\"1\"],\"id\":1}" http://127.0.0.1:8595

sleep 5

syncutxocommand=`./gptn --exec "ptn.syncUTXOByAddr($account6)"  attach node_test6/palletone/gptn6.ipc`
syncutxoinfo=`echo $syncutxocommand`
value="\"OK\""
if [ $syncutxoinfo = $value ];then
    echo "============syncUTXOByAddr account6 ok============"
    python  -m robot.run -d ./logs --test success ./light.robot
else
    echo "============syncUTXOByAddr account6 err:"$syncutxoinfo
    python  -m robot.run -d ./logs --test fail ./light.robot
fi


balancecommand=`./gptn --exec "wallet.getBalance($account6)"  attach node_test6/palletone/gptn6.ipc`
balanceinfo=`echo $balancecommand`
temp=`echo ${balanceinfo:7}`
length=`echo ${#temp}`
num=$[$lenght-2]
t1=`echo ${temp:0:$num} | sed 's/ //g' | sed 's/"//g'`
if [ $t1 = 80 ];then
    echo "============getBalance account6 ok============"
    python  -m robot.run -d ./logs --test success ./light.robot
else
    echo "============getBalance account6 err:"$t1
    python  -m robot.run -d ./logs --test fail ./light.robot
fi




# transferPTN in light node6 node7
curl -H "Content-Type:application/json" -X POST -d  "{\"jsonrpc\":\"2.0\",\"method\":\"wallet_transferToken\",\"params\":[\"PTN\",$account6,$account7,\"50\",\"1\",\"1\",\"1\"],\"id\":1}" http://127.0.0.1:8605

sleep 5

syncutxocommand=`./gptn --exec "ptn.syncUTXOByAddr($account7)"  attach node_test7/palletone/gptn7.ipc`
syncutxoinfo=`echo $syncutxocommand`
value="\"OK\""
if [ $syncutxoinfo = $value ];then
    echo "============syncUTXOByAddr account7 ok============"
    python  -m robot.run -d ./logs --test success ./light.robot
else
    echo "============syncUTXOByAddr account7 err:"$syncutxoinfo
    python  -m robot.run -d ./logs --test fail ./light.robot
fi


balancecommand=`./gptn --exec "wallet.getBalance($account7)"  attach node_test7/palletone/gptn7.ipc`
balanceinfo=`echo $balancecommand`
temp=`echo ${balanceinfo:7}`
length=`echo ${#temp}`
num=$[$lenght-2]
t1=`echo ${temp:0:$num} | sed 's/ //g' | sed 's/"//g'`
if [ $t1 = 50 ];then
    echo "============getBalance account7 ok============"
    python  -m robot.run -d ./logs --test success ./light.robot
else
    echo "============getBalance account7 err:"$t1
    python  -m robot.run -d ./logs --test fail ./light.robot
fi


