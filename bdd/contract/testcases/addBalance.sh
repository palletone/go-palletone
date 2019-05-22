#!/usr/bin/expect
#!/bin/bash
set timeout 30
set account0 [lindex $argv 0]
set account1 [lindex $argv 1]
# python addbalance.py
python ./addcbl.py account0 account1