#!/usr/bin/expect
#!/bin/bash

set timeout 30

set tokenHolder [lindex $argv 0] 

set another [lindex $argv 1] 

spawn ./node1/gptn --exec "personal.transferPtn($tokenHolder,$another,5990)" attach ./node1/palletone/gptn.ipc

expect "Passphrase:"

send "1\r"

interact
