import os
import subprocess
import pexpect
from time import sleep
import json

MAIN_PATH="../../../"

def edit_genesis_json():
    genesisPath = os.path.join(MAIN_PATH, 'GasToken/node/ptn-genesis.json')
    with open(genesisPath, "rw") as f:
        data_dict = json.load(f)
        data_dict['gasToken'] = 'WWW'
        json.dump(data_dict, f)
        f.close()

def new_genesis():
    gptnPath = os.path.join(MAIN_PATH, 'bdd/GasToken/node')
    os.chdir(gptnPath)
    child = pexpect.spawn(command='./gptn "" false')
    child.logfile=sys.stdout
    child.expect("Do you want to create a new account as the holder of the token")
    child.sendline("y")
    child.expect("Passphrase:")
    child.sendline("1")
    child.expect(["Repeat passphrase:"])
    child.sendline("1")
    child.expect(pexpect.EOF)
    sleep(2)

def init_gptn():
    gptnPath = os.path.join(MAIN_PATH, 'GasToken/node')
    os.chdir(gptnPath)
    child = pexpect.spawn('./gptn', ['init'])
    child.logfile=sys.stdout
    child.expect("Passphrase:")
    child.sendline("1")
    child.expect(pexpect.EOF)
    EOFLog = child.before
    print EOFLog

if __name__=='__main__':
    # new_genesis()
    edit_genesis_json()
    # init_gptn()