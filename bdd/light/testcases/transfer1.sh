#!/bin/bash
function Transfer()
{
transfer=`curl -H "Content-Type:application/json" -X POST -d  "{\"jsonrpc\":\"2.0\",\"method\":\"wallet_transferToken\",\"params\":[\"PTN\",\"P1AUTVTAKRcouMGBsGfWEXaWUEnLYJ1G4QN\",\"P1JVifKvVZromtyTeEw7C7knZrp9AKA1SQx\",\"100\",\"1\",\"1\",\"1\"],\"id\":1}" http://127.0.0.1:8545`

info=`echo $transfer|jq '.result'`

echo $info
length=`echo ${#info}`
if [ $length -eq 68 ];then
    echo "============transfer ok============"
else
    echo "============transfer err"$info"============"
fi
}



