#!/usr/bin/expect

spawn ./gptn newgenesis

expect "Do you want to create a new account as the holder of the token? \[y/N\]"

send "y\n"

expect "Passphrase:"

send "1\n"

expect "Repeat passphrase:"

send "1\n"

expect eof
