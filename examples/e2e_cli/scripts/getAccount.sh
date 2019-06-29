#!/usr/bin/expect
#!/bin/bash

#产生账号信息
set timeout 30
#需要配置gptn在meditorX的绝对路径
#spawn "gptn's dir" account new
spawn ./gptn account new
expect "Passphrase:"
send "1\n"
expect "Repeat passphrase:"
send "1\n"
interact
