#!/usr/bin/expect
#!/bin/bash

set maker [lindex $argv 0]
set saleassert [lindex $argv 1]
set saleamount [lindex $argv 2]
set wantassert [lindex $argv 3]
set wantamount [lindex $argv 4]
sleep 1

spawn ../node/gptn --exec "contract.ccinvokeToken(\"$maker\",'PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba',\"$saleassert\",\"$saleamount\",\"0.1\",\"PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba\",\[\"maker\", \"$wantassert\",\"$wantamount\"\])"    attach ../node/palletone/gptn.ipc

spawn ../node/gptn --exec "wallet.getBalance(\"$maker\")" attach ../node/palletone/gptn.ipc
expect eof
catch wait result
exit [lindex $result 3]
