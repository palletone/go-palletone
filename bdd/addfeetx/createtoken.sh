#!/usr/bin/expect
#!/bin/bash
#./test_case_setup.sh $1
set tokenHolder [lindex $argv 0]
set ptnamount [lindex $argv 1]
set tokenname [lindex $argv 2]
set decimalpoint [lindex $argv 3]
set tokenamount [lindex $argv 4]
sleep 1


spawn ../node/gptn --exec "personal.unlockAccount(\"$tokenHolder\",'1',600000000)"  attach ../node/palletone/gptn.ipc

spawn ../node/gptn --exec "contract.ccinvoketx(\"$tokenHolder\",\"$tokenHolder\",\"$ptnamount\",'1','PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43',\['createToken', \"hahahah\",\"$tokenname\",\"$decimalpoint\",\"$tokenamount\",\"$tokenHolder\"\])" attach ../node/palletone/gptn.ipc
sleep 1
#expect eof
#catch wait result
#exit [lindex $result 3]
spawn ../node/gptn --exec "wallet.getBalance(\"$tokenHolder\")" attach ../node/palletone/gptn.ipc
sleep 3
expect eof
catch wait result
exit [lindex $result 3]
