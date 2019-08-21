# -*- coding=utf-8 -*-
import requests
import json
import re
import unittest
from time import sleep
import time

PATH = lambda p: os.path.abspath(os.path.join(os.path.dirname(__file__), p))

class transferPtn(unittest.TestCase):
    host = 'http://localhost:8545/'

    def setUp(self):
        self.domain = transferPtn.host
        self.headers = {'Content-Type':'application/json'}
        self.genesisAddress = "P15efBXRNzWYTYK2twLvZTJVJ8jfAczY4PL"
        self.copygenesisAddress = "P15efBXRNzWYTYK2twLvZTJVJ8jfAczY4PL" #A

    def tearDown(self):
        pass

    def test_TransferPtn(self):
        print "Current URL:" + self.domain
        for i in range(100):
            self.transferPtn(self.copygenesisAddress,self.copygenesisAddress,1,0.0001,"remark","1")
            #sleep(1)
            self.getBalance(self.copygenesisAddress)

    def getBalance(self, address, domain=host):
        data = {
            "jsonrpc": "2.0",
            "method": "wallet_getBalance",
            "params": [address],
            "id": 1
        }
        data = json.dumps(data)

        response = requests.post(url=domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError, error:
            print "key " + error.message + " not found.\n"
        else:
            print 'Current Balance: ' + str(result) + '\n'
            return result

    def transferPtn(self,senderAddr,recieverAddr,senderAmount,poundage,remark,password):
        data = {
            "jsonrpc":"2.0",
            "method":"wallet_transferPtn",
            "params":
                [senderAddr,recieverAddr,senderAmount,poundage,remark,password],
            "id":1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        print 'CreateTrans Result: ' + result1['result'] + '\n'
        return result1['result'],recieverAddr

if __name__ == '__main__':
    unittest.main()