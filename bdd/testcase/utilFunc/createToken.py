#-*- coding=utf-8 -*-
import requests
import json
import random
import re
import string
from time import sleep

class createToken(object):

    def __init__(self):
        self.domain = 'http://localhost:8545/'
        self.headers = {'Content-Type':'application/json'}
        self.tempToken = ''
        self.tempValue = ''
        self.vote_value = '0'

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
        tokenId,value = self.getTokenStarts(preToken,result)
        return tokenId,value

    def geneNickname(self):
        nickname = "qa"+str(random.randint(100,999))
        self.nickname = nickname
        print "TokenId: " +nickname+"\n"
        return nickname

    def getTokenId(self, nickname, dict):
        for key in dict.keys():
            if key.startswith(nickname):
                self.tempToken = key
        return self.tempToken

    def getTokenStarts(self, nickname, dict):
        for key in dict.keys():
            if key.startswith(nickname):
                self.tempToken = key
                self.tempValue = dict[key]
                #print self.tempToken,self.tempvalue
        return self.tempToken,self.tempValue

    def ccinvokeCreateToken(self,senderAddr,recieverAddr,senderAmount,poundage,contractId,method,evidence,nickname,decimalAccuracy,tokenPoundage):
        data = {
                "jsonrpc":"2.0",
                "method":"contract_ccinvoketxPass",
                "params":
                    [senderAddr,recieverAddr,str(senderAmount),str(poundage),contractId,[method,evidence,nickname,str(decimalAccuracy),str(tokenPoundage),senderAddr]],
                "id":1
            }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        return result1

    def ccinvokePass(self,senderAddr,recieverAddr,senderAmount,poundage,contractId,*params):
        print params[0]
        print "Request method is: "+ params[0][0]
        data = {
                "jsonrpc":"2.0",
                "method":"contract_ccinvoketxPass",
                "params":[
                    senderAddr, recieverAddr, senderAmount, poundage, contractId, params[0],"1",60000000,""],
                "id":1
            }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        print result1
        return result1

    def ccinvoketx_apply(self,senderAddr,recieverAddr,senderAmount,poundage,tokenAmount):
        print self.nickname
        data = {
            "jsonrpc": "2.0",
            "method": "contract_ccinvoketx",
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
                "method": "wallet_transferToken",
                "params": [
                    tokenId,senderAddr,recieverAddr,senderAmount,poundage,evidence,"1",unlocktime
                ],
                "id": 1
            }
            data=json.dumps(data)
            response = requests.post(url=self.domain, data=data, headers=self.headers)
            result1 = json.loads(response.content)
            try:
                return result1['result']
            except KeyError:
                print "Request transferToken failed.\n" + str(result1)
            else:
                print 'transferToken Result: ' + str(result1['error']) + '\n'
                return result1['error']

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

    def personalListAccounts(self):
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
        print result1['result']
        return result1['result']

    def personalUnlockAccount(self,addr):
        data = {
            "jsonrpc": "2.0",
            "method": "personal_listAccounts",
            "params":
                [addr,"1",6000000],
            "id": 1
        }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        print result1['result']
        return result1['result']

    def ccqueryById(self,contranctId,methodType,preTokenId):
        data = {
                "jsonrpc":"2.0",
                "method":"contract_ccquery",
                "params":
                    [contranctId,[methodType,preTokenId],0],
                "id":1
            }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        return result1

    def compareCquery(self,result):
        #result = '{\"Symbol\":\"QA001\",\"CreateAddr\":\"P1N6s8g9if8kSRL86ta4mkVSLq1yLvMr5Je\",\"TotalSupply\":25000,\"Decimals\":1,\"SupplyAddr\":\"P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw\",\"AssetID\":\"QA001+10BLAMF0JYO3UVBEDJM\"}'
        result = json.loads(result)
        for key in result:
            print key,result[key]
            try:
                assert result['Symbol'] == 'QA001'
            except AssertionError,ext:
                ext.message
                print "assert failed.\n"

    def getTxByReqId(self,reqId):
        data = {
                "jsonrpc":"2.0",
                "method":"dag_getTxByReqId",
                "params":
                    [reqId],
                "id":1
            }
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        result1 = json.loads(response.content)
        return result1

    def ccqueryVoteResult(self,result,CreateAddr,AssetID):
        #result = "{\"IsVoteEnd\":false,\"CreateAddr\":\"P14432ABS64C2qDxJSqs4xQ9ZdXL2Zc46a7\",\"TotalSupply\":2000,\"SupportResults\":[{\"TopicIndex\":1,\"TopicTitle\":\"vote your love blockchain\",\"VoteResults\":[{\"SelectOption\":\"ptn0\",\"Num\":1},{\"SelectOption\":\"btc0\",\"Num\":0},{\"SelectOption\":\"eth0\",\"Num\":0},{\"SelectOption\":\"eos0\",\"Num\":0}]},{\"TopicIndex\":2,\"TopicTitle\":\"vote your hate blockchain\",\"VoteResults\":[{\"SelectOption\":\"ptn1\",\"Num\":2},{\"SelectOption\":\"btc1\",\"Num\":2},{\"SelectOption\":\"eth1\",\"Num\":0},{\"SelectOption\":\"eos1\",\"Num\":0}]}],\"AssetID\":\"VOTE+0GF8VT5B9G1IAT6QRRX\"}"
        result = json.loads(result)
        expectList = [2000,
                      AssetID,
                      False,
                      [{u'VoteResults': [{u'Num': 1, u'SelectOption': u'ptn0'}, {u'Num': 0, u'SelectOption': u'btc0'}, {u'Num': 0, u'SelectOption': u'eth0'}, {u'Num': 0, u'SelectOption': u'eos0'}], u'TopicIndex': 1, u'TopicTitle': u'vote your love blockchain'}, {u'VoteResults': [{u'Num': 2, u'SelectOption': u'ptn1'}, {u'Num': 2, u'SelectOption': u'btc1'}, {u'Num': 0, u'SelectOption': u'eth1'}, {u'Num': 0, u'SelectOption': u'eos1'}], u'TopicIndex': 2, u'TopicTitle': u'vote your hate blockchain'}],
                      CreateAddr
                      ]
        keyList = []
        valueList = []
        print "There is key,value of dictionary:"
        for key in result:
            #print key,result[key]
            keyList.append(key)
            valueList.append(result[key])
        print valueList
        for n in range(len(expectList)):
            self.assertDict(result,keyList,expectList,n)
        return result

    def assertDict(self,result,keyList,expectList,n):
        keyword = keyList[n]
        try:
            assert result[keyword] == expectList[n]
        except AssertionError:
            print keyword+" wrong: Actual is "+str(result[keyword])+".Expect is "+str(expectList[n])

    def voteExist(self,voteId,dict):
        data = json.dumps(dict)
        if voteId in data:
            self.vote_value = dict['result'][voteId]
            print self.vote_value
        else:
            print self.vote_value
        return self.vote_value

    def jsonLoads(self,dict,*keys):
        data = json.loads(dict)
        #print data['Symbol']
        keysList = []
        for n in range(len(keys)):
            keysList.append(data[keys[n]])
        if n > 0:
            print keysList[0],keysList[1]
            return keysList[0],keysList[1]
        else:
            return keysList[0]

if __name__ == '__main__':
    pass
    #dict = '{"Symbol":"CA001","CreateAddr":"P1LQ8dg9EWWnZMp8dfCTAHJic1F55awfrGt","TokenType":1,"TotalSupply":10,"SupplyAddr":"P1LQ8dg9EWWnZMp8dfCTAHJic1F55awfrGt","AssetID":"CA001+09P56EF11ID5BRZ80W9","TokenIDs":["1","10","2","3","4","5","6","7","8","9"]}'
    #voteId = "VOTE+0G7N1MFOXE1DYRHFXUP"
    #dict = {"id":1,"jsonrpc":"2.0","result":{"PTN":"0.00003","VOTE+0G7N1MFOXE1DYRHFXUP":"1000"}}
    #createToken().jsonLoads(dict,'AssetID','TokenIDs')
    #createToken().ccinvokePass()