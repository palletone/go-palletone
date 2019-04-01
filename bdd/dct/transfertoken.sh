#!/usr/bin/expect
#!/bin/bash
set timeout 30
set tokenHolder [lindex $argv 0]
set another [lindex $argv 1]
spawn ../node/gptn --exec "wallet.transferToken('PTN',$tokenHolder,$another,'4990','10')" attach ../node/palletone/gptn.ipc
expect "Passphrase:"
send "1\r"
interact
