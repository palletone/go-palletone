#!/bin/bash

set timeout 30
#/usr/bin/expect <<EOF

./preset.sh
#spawn ./start.sh
#expect "Please input the numbers of nodes you want:"
#send "3\r"
#interact
#expect eof
#EOF

sleep 20
startProduce1=`./gptn --exec 'mediator.startProduce()' attach node1/palletone/gptn.ipc`
nodeInfo1=`./gptn --exec 'admin.nodeInfo' attach node1/palletone/gptn.ipc`
nodeInfo2=`./gptn --exec 'admin.nodeInfo' attach node2/palletone/gptn2.ipc`
nodeInfo3=`./gptn --exec 'admin.nodeInfo' attach node3/palletone/gptn3.ipc`
echo `echo $startProduce1`
echo `echo $nodeInfo1`
echo `echo $nodeInfo2`
echo `echo $nodeInfo3`

