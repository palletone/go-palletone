#!/usr/bin/expect
#!/bin/bash

spawn /data/deployment/gptn attach /data/deployment/node_test6/palletone/gptn6.ipc
send "personal.unlockAccount(\"P1GVi1wg1J9PXyBsJEPcQeQhcS4FDWrXMQk\")\n"
expect "Passphrase:"
send "1\n"  
expect "true"

set p_loop 100
while { $p_loop } {
    set timeout 20
    send "contract.ccquery(\"PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM\",\[\"IsFinishAllocated\"\])\n"
    expect {
        "true" {
            set p_loop 0
        }

        "false" {
            set p_loop [expr $p_loop-1]
	    send "contract.ccinvoketx(\"P1GVi1wg1J9PXyBsJEPcQeQhcS4FDWrXMQk\",\"P1GVi1wg1J9PXyBsJEPcQeQhcS4FDWrXMQk\",\"0\",\"0.001\",\"PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM\",\[\"HandlePledgeReward\"\])\n"
            sleep 7
        }
        timeout {
            set p_loop [expr $p_loop-1]
        }
    }
}
send "exit\n"
