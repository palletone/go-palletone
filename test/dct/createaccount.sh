#!/usr/bin/expect
#!/bin/bash
set timeout 30
spawn ./gptn account new
expect "Passphrase:"
send "1\r"
expect "Repeat passphrase:"
send "1\r"  
interact

