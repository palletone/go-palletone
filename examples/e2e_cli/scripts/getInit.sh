#!/usr/bin/expect

#(1)设置超时时间
set timeout 30
#(2)执行init
spawn ./gptn init
expect "Passphrase:"
send "1\n"
interact
