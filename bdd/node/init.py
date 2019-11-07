import os
import subprocess
import pexpect
from time import sleep

# os.chdir("/home/travis/gopath/src/github.com/palletone/go-palletone/bdd/node")
child = pexpect.spawn(command="./gptn newgenesis \"\" false",maxread=3000)
child.expect("Do you want to create a new account as the holder of the token")
child.sendline("y")
print child.after
child.expect("Passphrase:")
child.sendline("1")
print child.after
child.expect(["Repeat passphrase:"])
child.sendline("1")
print child.before
print child.after
child.expect(pexpect.EOF)
sleep(2)

subprocess.call("sed -i 's/\"mediator_interval\": 3,/\"mediator_interval\": 2,/g' ptn-genesis.json",shell=True)
subprocess.call("sed -i 's/\"maintenance_skip_slots\": 2,/\"maintenance_skip_slots\": 0,/g' ptn-genesis.json",shell=True)
subprocess.call("sed -i 's/\"min_maint_skip_slots\": 2,/\"min_maint_skip_slots\": 0,/g' ptn-genesis.json",shell=True)
subprocess.call("sed -i 's/\"initialTimestamp\": [0-9]*,/\"initialTimestamp\": 1566269000,/g' ptn-genesis.json",shell=True)
subprocess.call("sed -i 's/\"contract_election_num\": 4,/\"contract_election_num\": 3,/g' ptn-genesis.json",shell=True)
subprocess.call("grep 'mediator_interval' ptn-genesis.json",shell=True)
subprocess.call("grep 'maintenance_skip_slots' ptn-genesis.json",shell=True)
subprocess.call("grep 'min_maint_skip_slots' ptn-genesis.json",shell=True)
subprocess.call("grep 'initialTimestamp' ptn-genesis.json",shell=True)

sleep(2)
child = pexpect.spawn('./gptn',['init'])
child.expect("Passphrase:")
child.sendline("1")
child.expect(pexpect.EOF)
EOFLog = child.before
print EOFLog



'''
child = pexpect.spawn(command="./gptn --exec 'personal.listAccounts' attach palletone/gptn.ipc")
child.expect(pexpect.EOF)
print child.before
print child.before.split('\"')[1]
GENEADD = child.before.split('\"')[1]
child = pexpect.spawn(command="./gptn account new")
# child.logfile = sys.stdout
# child.logfile = logFile
child.expect("Passphrase:")
child.sendline("1")
child.expect(["Repeat passphrase:"])
child.send("1\r\n")
child.expect(pexpect.EOF)
'''
