# -*- coding=utf-8 -*-
import requests
import json
import threading
import time
import sys
from time import sleep

reqIds = []
addrs = []
for i in range(1, len(sys.argv)):
    arg = sys.argv[i]
    addrs.append(str(arg))
class createToken():
    def __init__(self):
        self.domain = 'http://localhost:8545/'
        self.headers = {'Content-Type': 'application/json'}
        self.contractAddress = "PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf"
        self.funcName = "payout"
        self.pwd = "1"

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
        except KeyError:
            print('Current Balance: ' + str(result) + '\n')
            return result

    def ccinvoketx_create(self, senderAddr, recieverAddr, contractAddr, tokenAmount,fee):

        data = {
            "jsonrpc": "2.0",
            "method": "contract_ccinvoketx",
            "params":
                [senderAddr, recieverAddr, tokenAmount, fee, contractAddr, [self.funcName, "1", recieverAddr],self.pwd,
                6000000],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError:
            print("Request transfer failed. \naddr:" +str(senderAddr)+'\n' + str(result1))
        else:
            print('test transfer Result: '+str(senderAddr) +'\n'+ str(result) + '\n')
            reqIds.append(str(result['request_id']))
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
            print("Request testGetBalance failed.\n" + str(result1))
        else:
            print('testGetBalance result: ' + str(result) + '\n')
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
            print("Request transferToken failed.\n" + str(result1))
        else:
            print('transferToken Result: ' + str(result) + '\n')
            return result

    def getTxByReqId(self, applyResult):
        data = {
            "jsonrpc": "2.0",
            "method": "dag_getTxByReqId",
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
            print("Request getTxByReqId failed.\n" + str(applyResult)+'\n'+ str(result1))
        else:
            print('getTxByReqId Result: ' + str(result) + '\n')
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
            print("Request getTxByHash failed.\n" + str(result1))
        else:
            print('getTxByHash Result: ' + str(result) + '\n')
            return result

    def BtoC(self, re_batchSign):
        print(str(time.strftime("%Y-%m-%d %X")) + "  Begin\n")
        # @signResult = open(r'geneBatchResult.txt', 'a+', buffering=1)
        signResult = open(r'Transaction/geneSignResult.txt', 'a+', buffering=1)
        for batchSign in re_batchSign:
            signResult.write("".join(batchSign) + "\n")
            signResult.flush()
        signResult.close()
        print(str(time.strftime("%Y-%m-%d %X")) + "  Finished!\n")

threads = []
for addr in addrs:
    print("addr:" + str(addr)+ '\n')
    for i in range(5):
        print("index:" + str(i)+ '\n')
        t1 = threading.Thread(target=createToken().ccinvoketx_create, args=(addr,addr,"PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf",100,1))
        threads.append(t1)

if __name__ == '__main__':
    createToken().getBalance("PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf")
    for t in threads:
        t.setDaemon(True)
        t.start()
    time.sleep(10)
    for id in reqIds:
        # print "reqid:" + str(id) + '\n'
        createToken().getTxByReqId(id)
    print("after transfer contract:"+ '\n')
    createToken().getBalance("PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf")