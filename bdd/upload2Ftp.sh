#!/usr/bin/expect
#!/bin/bash
set timeout 120
set ftppwd [lindex $argv 0]
set folder [lindex $argv 1]
set number [lindex $argv 2]
spawn lftp travis:$ftppwd@47.74.209.46
expect "lftp"
send "cd ${folder}\n"
expect "cd"
send "mkdir ${number}\n"
expect "mkdir"
send "cd ${number}\n"
expect "cd"
send "mirror -R /home/travis/gopath/src/github.com/palletone/go-palletone/bdd/logs\n"  
expect "transferred"
send "exit\n"
interact