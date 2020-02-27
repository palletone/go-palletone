#!/usr/bin/expect
#!/bin/bash
set timeout 30
spawn ../node/gptn --exec "personal.newAccount()" attach ../node/palletone/gptn.ipc
expect "Passphrase:"
send "1\n"
expect "Repeat passphrase:"
send "1\n"  
interact

