#!/usr/bin/expect

set timeout 30

#输入gptn在Mediator0的绝对路径
spawn /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator0/gptn init

expect "Passphrase:"

send "1\r"

interact
