#!/usr/bin/expect
#!/bin/bash
set timeout 30
#MUST FULL PATH
spawn ./gptn attach
send "personal.unlockAccount(personal.listAccounts\[0\])"
expect "Passphrase:"
send "1\n"  
expect "true"
send "contract.ccinvoketx(personal.listAccounts\[0\],personal.listAccounts[\0\],\"100\",\"1\",\"PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM\",\[\"HandlePledgeReward\"\])\n"
expect "0x"
send "exit\n"  
interact
