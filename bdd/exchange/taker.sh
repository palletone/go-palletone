#!/usr/bin/expect
#!/bin/bash

set taker [lindex $argv 0]
set saleassert [lindex $argv 1]
set saleamount [lindex $argv 2]
set exchangesn [lindex $argv 3]
sleep 1
spawn ../node/gptn --exec "personal.unlockAccount(\"$taker\",'1',0)" attach ../node/palletone/gptn.ipc
sleep 1
spawn ../node/gptn --exec "contract.ccinvokeToken(\"$taker\",'PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba',\"$saleassert\",\"$saleamount\",'0.1',\"PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba\",\[\"taker\",\"$exchangesn\"])"  attach ../node/palletone/gptn.ipc
#contract.ccinvokeToken("P13xtftvWYNmpB79KBMHJdqExBxsGZEo5oD","PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba","XXX+10WSCCMGA9T7EAAXRX4","100","0.1","PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba",["taker","0xfaf9690575e96d0e10d93e7508ea85636956058c1664d1ad8d964e9c1bc5e2e9"])
#spawn ../node/gptn --exec "wallet.getBalance(\"$taker\")" attach ../node/palletone/gptn.ipc
spawn ../node/gptn --exec "contract.ccquery('PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba', \['getActiveOrderList'])"  attach ../node/palletone/gptn.ipc
expect eof
catch wait result
exit [lindex $result 3]
