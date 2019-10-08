#!/usr/bin/expect
#!/bin/bash
set account [lindex $argv 0]
set timeout 30
spawn ./gptn account dumppubkey $account
expect "Passphrase:"
send "1\n"
interact

