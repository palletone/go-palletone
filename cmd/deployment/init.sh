#!/usr/bin/expect
#!/bin/bash
set timeout 30
spawn ./gptn init
expect "Passphrase:"
send "1\r"
interact











