#!/usr/bin/expect
#!/bin/bash
set timeout 30
set tokenHolder [lindex $argv 0]
set another [lindex $argv 1]
spawn ../../node/gptn --exec "personal.transferPtn($tokenHolder,$another,5990)" attach ../../node/palletone/gptn.ipc
expect "Passphrase:"
send "1\n"
interact
