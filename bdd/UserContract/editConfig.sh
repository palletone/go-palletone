#!/usr/bin/env bash

sed -i "s/ElectionNum = 2/ElectionNum = 3/g" node1/ptn-config.toml
sed -i "s/ElectionNum = 2/ElectionNum = 3/g" node2/ptn-config.toml
sed -i "s/ElectionNum = 2/ElectionNum = 3/g" node3/ptn-config.toml

baseAddr="ContractAddress = \""
newAddr="$baseAddr$local_host"
echo $newAddr
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node1/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node2/ptn-config.toml
sed -i "s/ContractAddress\s*=\s*\"127.0.0.1/$newAddr/g" node3/ptn-config.toml