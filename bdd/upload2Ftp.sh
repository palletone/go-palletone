#!/usr/bin/expect
#!/bin/bash
set timeout 120
set ftppwd [lindex $argv 0]
spawn lftp travis:$ftppwd@47.74.209.46
expect "lftp"
send "cd pub\r"
expect "ok"
send "mput /home/travis/gopath/src/github.com/palletone/go-palletone/bdd/node/log/all.log\r"  
expect "transferred"
send "mirror -R /home/travis/gopath/src/github.com/palletone/go-palletone/bdd/logs\r"  
expect "transferred"
send "exit\r"
interact