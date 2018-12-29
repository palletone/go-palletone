#!/usr/bin/expect
#!/bin/bash

#产生账号信息
set timeout 30
#需要配置gptn的绝对路径地址
#spawn "gptn's dir" account new
spawn /opt/gopath/src/github.com/palletone/go-palletone/build/bin/gptn account new
expect "Passphrase:"
send "1\r"
expect "Repeat passphrase:"
send "1\r"
interact
