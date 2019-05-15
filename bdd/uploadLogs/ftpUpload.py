import os
import subprocess
import pexpect
from time import sleep

logPath = "/home/travis/gopath/src/github.com/palletone/go-palletone/bdd/logs"
os.chdir(logPath)
# 登录
child = pexpect.spawn(command="./ftp 39.105.191.26",maxread=3000)
child.expect("Name (39.105.191.26:root):")
child.sendline("ftpuser")
print child.after
child.expect("Password:")
child.sendline("123456")
print child.after
# 登录成功
child.expect("ftp>")
child.sendline("cd pub")
print child.after

# 上传日志
child.expect("ftp>")
createTransLogName = "createTrans.log.html"
ccinvokeLogName = "ccinvoke.log.html"
DigitalIdentityCertLogName = "DigitalIdentityCert.log.html"
child.sendline("put "+logPath+createTransLogName+" "+createTransLogName)
child.expect(pexpect.EOF)