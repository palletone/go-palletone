# -*- coding=utf-8 -*-
import requests
import json
import threading
import time
import sys
from time import sleep

reqIds = []
txIds = []
addrs = []
for i in range(1, len(sys.argv)):
    arg = sys.argv[i]
    addrs.append(str(arg))
class createToken():
    def __init__(self):
        self.domain = 'http://localhost:8545/'
        self.headers = {'Content-Type': 'application/json'}
        self.genesisAddress = "P1FRZ2AVgCd2TwS5SYDy1ehe8YaXYn86J7U"
        self.copygenesisAddress = "P1ATS6kLVuktJWT6qvASRsA8zAQqYUCshU6"  # A
        self.recieverAddr1 = "P19eqechnFyLW4a2xtEMgxXPEGLjzuMg4od"  # B
        self.recieverAddr2 = "P1MCApH2y7KRkWkutViSAA7WbVoZdF3EFaR"  # C
        self.nickname = "QA580"
        self.senderAddress = "P12rKwCpT2DuqHL2tGAXCVUAEtanszzPwve"
        self.recieverAddress = "P12rKwCpT2DuqHL2tGAXCVUAEtanszzPwve"
        self.contractAddress = "PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf"
        self.funcName = "testAddBalance"
        self.assetId = "jay"

    def runTest(self):
        pass

    def tearDown(self):
        pass

    def getBalance(self, address):
        data = {
            "jsonrpc": "2.0",
            "method": "wallet_getBalance",
            "params": [address],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError, error:
            print "key " + error.message + " not found.\n"
        else:
            print 'Current Balance: ' + str(result) + '\n'
            return result

    def ccinvoketx_create(self, senderAddr, recieverAddr, contractAddr, tokenAmount):

        data = {
            "jsonrpc": "2.0",
            "method": "contract_ccinvoketxPass",
            "params":
                [senderAddr, recieverAddr, "10", "1", contractAddr, [self.funcName, self.assetId, str(tokenAmount)],"1", 6000000, ""],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError:
            print "Request addBalance failed. \naddr:" +str(senderAddr)+'\n' + str(result1)
        else:
            print 'testaddBalance Result: '+str(senderAddr) +'\n'+ str(result) + '\n'
            reqIds.append(str(result))
            return result

    def ccquery(self, address, funcName, assetId):
        data = {
            "jsonrpc": "2.0",
            "method": "contract_ccquery",
            "params":
                [address, [funcName, assetId],0],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError:
            print "Request testGetBalance failed.\n" + str(result1)
        else:
            print 'testGetBalance result: ' + str(result) + '\n'
            return result

    def transferToken(self, senderAddr, recieverAddr, senderAmount, poundage):
        tokenid = self.getTokenId(self.nickname)
        data = {
            "jsonrpc": "2.0",
            "method": "wallet_transferToken",
            "params": [
                tokenid, senderAddr, recieverAddr, senderAmount, poundage, "1"
            ],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError:
            print "Request transferToken failed.\n" + str(result1)
        else:
            print 'transferToken Result: ' + str(result) + '\n'
            return result

    def getTxHashByReqId(self, applyResult):
        data = {
            "jsonrpc": "2.0",
            "method": "dag_getTxHashByReqId",
            "params":
                [
                    applyResult
                ],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError:
            print "Request getTxHashByReqId failed.\n" + str(result1)
        else:
            print 'getTxHashByReqId Result: ' + str(result) + '\n'
            return result

    def getTxByHash(self, txHashInfo):
        txHashInfo = json.loads(txHashInfo)
        data = {
            "jsonrpc": "2.0",
            "method": "dag_getTxByHash",
            "params": [
                txHashInfo['info']
            ],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError:
            print "Request getTxByHash failed.\n" + str(result1)
        else:
            print 'getTxByHash Result: ' + str(result) + '\n'
            return result

    def BtoC(self, re_batchSign):
        print str(time.strftime("%Y-%m-%d %X")) + "  Begin\n"
        # @signResult = open(r'geneBatchResult.txt', 'a+', buffering=1)
        signResult = open(r'Transaction/geneSignResult.txt', 'a+', buffering=1)
        for batchSign in re_batchSign:
            signResult.write("".join(batchSign) + "\n")
            signResult.flush()
        signResult.close()
        print str(time.strftime("%Y-%m-%d %X")) + "  Finished!\n"


threads = []
for addr in addrs:
    print  "addr:" + str(addr)+ '\n'

    t1 = threading.Thread(target=createToken().ccinvoketx_create, args=(addr,addr,"PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf",100,))
    threads.append(t1)

if __name__ == '__main__':
    for t in threads:
        t.setDaemon(True)
        t.start()
    time.sleep(5)
    for id in reqIds:
        # print "reqid:" + str(id) + '\n'
        createToken().getTxByHash(createToken().getTxHashByReqId(id))
    createToken().ccquery("PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf", "testGetBalance", "jay")
