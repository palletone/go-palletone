#!/usr/bin/expect

set timeout 30
spawn ./gptn init
expect "Passphrase:"
send "1\n"
interact


