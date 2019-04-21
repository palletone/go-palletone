#-*- coding=utf-8 -*-
import requests
import json
import random
import re

class createToken(object):

    def __init__(self):
        self.domain = 'http://localhost:8545/'
        self.headers = {'Content-Type':'application/json'}
        self.tempToken = ''
        self.tempValue = ''

    def runTest(self):
        pass

    def getBalance(self,address):
        data = {
            "jsonrpc": "2.0",
            "method": "ptn_getBalance",
            "params": [address],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result =result1['result']
        except KeyError,error:
            print "key " + error.message +" not found.\n"
        else:
            print 'Current Balance: ' + str(result) + '\n'
            return result

    def getTokenIdFromTokens(self,address,preToken):
        result = self.getBalance(address)
        tokenId = self.getTokenStarts(preToken,result)

        print result
        return str(tokenId)

    def geneNickname(self):
        nickname = "qa"+str(random.randint(100,999))
        self.nickname = nickname
        print "TokenId: " +nickname+"\n"
        return nickname

    def getTokenId(self, nickname, dict):
        for key in dict.keys():
            if re.search(r'^'+nickname, key):
                self.tempToken = key
        return self.tempToken

    def getTokenStarts(self, nickname, dict):
        for key in dict.keys():
            if key.startswith(nickname):
                self.tempToken = key
                self.tempValue = dict[key]
                #print self.tempToken,self.tempvalue
        return self.tempToken,self.tempValue

    def ccinvoketxPass(self,senderAddr,recieverAddr,senderAmount,poundage,contractId,method,evidence,nickname,decimalAccuracy,tokenPoundage):
        data = {
                "jsonrpc":"2.0",
                "method":"ptn_ccinvoketxPass",
                "params":
                    [senderAddr,recieverAddr,str(senderAmount),str(poundage),contractId,[method,evidence,nickname,str(decimalAccuracy),str(tokenPoundage),senderAddr]],
                "id":1
            }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        return result1


    def listAccounts(self):
        data = {
            "jsonrpc": "2.0",
            "method": "personal_listAccounts",
            "params":
                [],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        print result1['result'][0]
        return result1['result'][0]

    def paramGroups(self,dynamic,number,*params):
        if len(params) < number:
            paramsList = []
            paramsList.append(dynamic)
            print  paramsList
            for n in params:
                # paramsList.append(n)
                print n
            paramsList.extend(n)
            print paramsList
            return paramsList
        else:
            for n in params:
                # paramsList.append(n)
                print n
            return n

    def ccinvoketx(self,*params):
        geneAdd = self.listAccounts()
        n = self.paramGroups(geneAdd,13,*params)
        #'''
        data = {
                "jsonrpc":"2.0",
                "method":"ptn_ccinvoketxPass",
                "params":
                    n,
                "id":1
            }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        print result1
        return result1
        #'''

    def ccinvoketx_apply(self,senderAddr,recieverAddr,senderAmount,poundage,tokenAmount):
        print self.nickname
        data = {
            "jsonrpc": "2.0",
            "method": "ptn_ccinvoketx",
            "params":
                [senderAddr, recieverAddr, str(senderAmount), str(poundage), "PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43",
                 ["supplyToken", self.nickname, str(tokenAmount)]],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        try:
            result = result1['result']
        except KeyError:
            print "Request ApplyToken failed.\n" + str(result1)
        else:
            print 'ApplyToken Result: ' + str(result) + '\n'
            return result

    def transferToken(self,tokenId,senderAddr,recieverAddr,senderAmount,poundage,evidence,unlocktime):
            data = {
                "jsonrpc": "2.0",
                "method": "ptn_transferToken",
                "params": [
                    tokenId,senderAddr,recieverAddr,senderAmount,poundage,evidence,"1",unlocktime
                ],
                "id": 1
            }
            data=json.dumps(data)
            response = requests.post(url=self.domain, data=data, headers=self.headers)
            result1 = json.loads(response.content)
            try:
                result = result1['result']
            except KeyError:
                print "Request transferToken failed.\n" + str(result1)
            else:
                print 'transferToken Result: ' + str(result) + '\n'
                return result

if __name__ == '__main__':
    pass
    #createToken().getTokenIdFromTokens('P1HhWxfQLMgb5TfE56GASURCuitX2XL397G','QA003')