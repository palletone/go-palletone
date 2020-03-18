#!/usr/bin/expect -f
#!/bin/bash
set timeout 30
spawn ../node/gptn account new
expect "Passphrase:"
send "1\n"
expect "Repeat passphrase:"
send "1\n"
expect eof
