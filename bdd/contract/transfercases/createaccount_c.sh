#!/usr/bin/expect
#!/bin/bash
set timeout 30
spawn ./gptn --exec "personal.newAccount()" attach ./palletone/gptn.ipc
expect "Passphrase:"
send "1\n"
expect "Repeat passphrase:"
send "1\n"  
interact

