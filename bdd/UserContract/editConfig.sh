#!/bin/bash

local_host=`/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|grep -v 172.17.0.1|awk '{print $2}'|tr -d "addr:"`

echo $local_host

sed -i "s/8545/8645/g" node1/ptn-config.toml
sed -i "s/HTTPHost = \"localhost\"/HTTPHost = \"0.0.0.0\"/g" node1/ptn-config.toml
sed -i "s/ElectionNum = 2/ElectionNum = 3/g" node1/ptn-config.toml
#sed -i "s/ContractSigNum\s*=\s*2/ContractSigNum = 3/g" node1/ptn-config.toml
baseAddr="ContractAddress = \""
newAddr="$baseAddr$local_host"
echo $newAddr
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node1/ptn-config.toml

sed -i "s/8565/8665/g" node2/ptn-config.toml
sed -i "s/HTTPHost = \"localhost\"/HTTPHost = \"0.0.0.0\"/g" node2/ptn-config.toml
sed -i "s/ElectionNum = 2/ElectionNum = 3/g" node2/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node2/ptn-config.toml
#sed -i "s/ContractSigNum\s*=\s*2/ContractSigNum = 3/g" node2/ptn-config.toml

sed -i "s/8575/8675/g" node3/ptn-config.toml
sed -i "s/HTTPHost = \"localhost\"/HTTPHost = \"0.0.0.0\"/g" node3/ptn-config.toml
sed -i "s/ElectionNum = 2/ElectionNum = 3/g" node3/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node3/ptn-config.toml
#sed -i "s/ContractSigNum\s*=\s*2/ContractSigNum = 3/g" node3/ptn-config.toml

sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node_test4/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node_test5/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node_test6/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node_test7/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node_test8/ptn-config.toml