#!/usr/bin/expect
#!/bin/bash
set timeout 120
set ftppwd [lindex $argv 0]
set folder [lindex $argv 1]
spawn lftp travis:$ftppwd@47.74.209.46
expect "lftp"
send "cd ${folder}\r"
expect "cd"
send "mirror -R /home/travis/gopath/src/github.com/palletone/go-palletone/bdd/logs\r"  
expect "transferred"
send "exit\r"
interact