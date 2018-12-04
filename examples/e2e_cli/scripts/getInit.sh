#!/usr/bin/expect

set timeout 30

#输入gptn绝对路径
spawn /opt/gopath/src/github.com/palletone/go-palletone/build/bin/gptn init

expect "Passphrase:"

send "1\r"

interact
