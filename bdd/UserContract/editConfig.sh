#!/bin/bash

local_host=`/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|grep -v 172|awk '{print $2}'|tr -d "addr:"`

echo $local_host

baseAddr="ContractAddress = \""
newAddr="$baseAddr$local_host"
echo $newAddr

NUM=$1
SIG=$2

for ((i=1; i<=$NUM; i++))
do
if [ $i == 1 ]
then
sed -i "s/8545/8645/g" node$i/ptn-config.toml
else
let "originPort=8545+$i*10"
let "newPort=8645+$i*10"
sed -i "s/$originPort/$newPort/g" node$i/ptn-config.toml
fi
sed -i "s/HTTPHost = \"localhost\"/HTTPHost = \"0.0.0.0\"/g" node$i/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node$i/ptn-config.toml
sed -i "s/IsJury = false/IsJury = true/g" node$i/ptn-config.toml
done
