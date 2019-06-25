#!/bin/bash


function CheckBalance()
{
# ./gptn --exec 'wallet.transferPTN(personal.listAccounts[1],'$1',1000,2)'  attach node1/palletone/gptn.ipc
    node=$1
    account=$2
    balance=$3    
    balancecommand='./gptn --exec 'ptn.getBalance("'$2'")'  attach node_test'$1'/palletone/gptn'$1'.ipc'
    `echo $balancecommand`
    #balance=`echo $balancecommand | jq .`
    #account=${account:0:35}
    #account=`echo ${account//^M/}`

}
CheckBalance 4 'P1JVifKvVZromtyTeEw7C7knZrp9AKA1SQx' 1000 
