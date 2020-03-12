#!/bin/bash

#logpath="/home/travis/gopath/src/github.com/palletone/go-palletone/bdd/logs/light"
logpath="../logs/light"

./preset.sh
sleep 40

listcommand=`./gptn --exec 'personal.listAccounts'  attach node1/palletone/gptn.ipc`
list=`echo ${listcommand//
/}`
moneyaccount=`echo "${list}" | jq -c '.[1]'`

newaccount5command=`./gptn --exec 'personal.newAccount("1")'  attach node_test5/palletone/gptn5.ipc`
a5=`echo ${newaccount5command//
/}`
account5=`echo $a5`

newaccount6command=`./gptn --exec 'personal.newAccount("1")'  attach node_test6/palletone/gptn6.ipc`
a6=`echo ${newaccount6command//
/}`
account6=`echo $a6`

newaccount7command=`./gptn --exec 'personal.newAccount("1")'  attach node_test7/palletone/gptn7.ipc`
a7=`echo ${newaccount7command//
/}`
account7=`echo $a7`

moneyaccount=`echo ${moneyaccount//\"/}`
account5=`echo ${account5//\"/}`
account6=`echo ${account6//\"/}`
account7=`echo ${account7//\"/}`

echo "tokenHolder: "$moneyaccount
echo "light account5: "$account5
echo "light account6: "$account6
echo "light account7: "$account7

startProduce=`./gptn --exec 'mediator.startProduce()'  attach node1/palletone/gptn.ipc`
echo $startProduce

python  -m robot.run -d $logpath -v moneyaccount:$moneyaccount -v account5:$account5 -v account6:$account6 -v account7:$account7 ./light.robot