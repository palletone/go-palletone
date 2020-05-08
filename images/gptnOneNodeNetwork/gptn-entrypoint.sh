#!/bin/bash

# 判断一下是否已经存在了，避免 restart 的时候重复初始化
if [ ! -f "./ptn-config.toml" ]; then
  ./newgenesis.sh

  sleep 2

  ./init.sh

  sleep 2
fi

sed -i "s/HTTPHost = \"localhost\"/HTTPHost = \"0.0.0.0\"/g" ./ptn-config.toml

sed -i "s/HTTPVirtualHosts = \[\"localhost\"\]/HTTPVirtualHosts = \[\"*\"\]/g" ./ptn-config.toml

./gptn
