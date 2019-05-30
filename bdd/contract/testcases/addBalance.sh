#!/bin/bash
set timeout 30
set account0 [lindex $argv 0]
set account1 [lindex $argv 1]
echo "addr: $account0"
python ./addcbl.py $account0 $account1
