#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import pexpect

logPath = "/home/travis/gopath/src/github.com/palletone/go-palletone/bdd/logs"
os.chdir(logPath)
# login
child = pexpect.spawnu("ftp 39.105.191.26")
print child.after
child.expect(u'(?i)Name.*.')
child.sendline("ftpuser")
print child.after
child.expect(u"(?i)Password")
child.sendline("123456")
print child.after
# login succeed
child.expect(u"ftp>")
child.sendline("cd pub")
print child.after
# upload logs
child.expect(u"ftp>")
createTransLogName = "createTrans.log.html"
# createTransLogName = "log.txt"
# ccinvokeLogName = "ccinvoke.log.html"
# DigitalIdentityCertLogName = "DigitalIdentityCert.log.html"
putStr = "put "+logPath+"/"+createTransLogName+" "+createTransLogName
child.sendline(putStr)
try:
	child.expect(u"(?i).*complete.*")
	print "=== upload succed ==="
except:
	child.sendline('quote pasv')
	child.sendline('passive')
	child.after
	child.sendline(putStr)

try:
	child.expect(u"(?i).*complete.*")
	print "=== upload succed ==="
except:
	print "=== upload failed === "

child.expect(u"ftp>")
child.sendline("bye")
print child.after
child.close()
