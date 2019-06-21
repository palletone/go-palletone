#-*- coding=utf-8 -*-
import requests
import json
import random
import re
import string
from time import sleep

class createToken(object):
    #host = 'http://192.168.0.128:8545/'
    host = 'http://localhost:8545/'
    def __init__(self):
        self.domain = createToken.host
        self.headers = {'Content-Type':'application/json'}
        self.tempToken = ''
        self.tempValue = ''
        self.vote_value = '0'
        self.status = -1
        self.supplyAddr = ''

    def runTest(self):
        pass

    def getBalance(self,address,domain=host):
        data = {
            "jsonrpc": "2.0",
            "method": "ptn_getBalance",
            "params": [address],
            "id": 1
        }
        data = json.dumps(data)
        print "Current URL:"+domain
        response = requests.post(url=domain, data=data, headers=self.headers)
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

    def getTokenStatus(self, tkInfo):
        dict = json.loads(tkInfo)
        if dict.has_key:
          self.status = int(dict['Status'])
        return self.status
    def getTokenSupplyAddr(self, tkInfo):
        dict = json.loads(tkInfo)
        if dict.has_key:
          self.supplyAddr = dict['SupplyAddr']
        return self.supplyAddr

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

    def getOneTokenInfo(self,tokenSymbol):
            data = {
                "jsonrpc": "2.0",
                "method": "wallet_getOneTokenInfo",
                "params": [tokenSymbol],
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

    def ccqueryVoteResult(self,result,CreateAddr,AssetID,TokenAmount):
        result = json.loads(result)
        expectList = [
                      int(TokenAmount),
                      AssetID,
                      False,
                      [{u'VoteResults': [{u'Num': 1, u'SelectOption': u'ptn0'}, {u'Num': 0, u'SelectOption': u'btc0'}, {u'Num': 0, u'SelectOption': u'eth0'}, {u'Num': 0, u'SelectOption': u'eos0'}], u'TopicIndex': 1, u'TopicTitle': u'vote your love blockchain'}, {u'VoteResults': [{u'Num': 1, u'SelectOption': u'ptn1'}, {u'Num': 1, u'SelectOption': u'btc1'}, {u'Num': 0, u'SelectOption': u'eth1'}, {u'Num': 0, u'SelectOption': u'eos1'}], u'TopicIndex': 2, u'TopicTitle': u'vote your hate blockchain'}],
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
        for k in range(len(expectList)):
            self.assertDict(result,keyList,expectList,k)
        return result

    def assertDict(self,result,keyList,expectList,key):
        keyword = keyList[key]
        try:
            assert result[keyword] == expectList[key]
        except AssertionError:
            print keyword+" false: Actual is "+str(result[keyword])+".Expect is "+str(expectList[key])

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

    def getTokenIdByNum(self, nickname, dict, num=1):
        num=int(num)
        calculate = 0
        for key in dict.keys():
            if key.startswith(nickname):
                calculate = calculate + 1
                if calculate==num:
                    print key
                    return key

if __name__ == '__main__':
    pass
    #createToken().assertDict()
    #dict = "{\"IsVoteEnd\":false,\"CreateAddr\":\"P1D4GXUcrcT7tAhMPvTxTQk7vQQSJkiad54\",\"TotalSupply\":60000,\"SupportResults\":[{\"TopicIndex\":1,\"TopicTitle\":\"vote your love blockchain\",\"VoteResults\":[{\"SelectOption\":\"ptn0\",\"Num\":1},{\"SelectOption\":\"btc0\",\"Num\":0},{\"SelectOption\":\"eth0\",\"Num\":0},{\"SelectOption\":\"eos0\",\"Num\":0}]},{\"TopicIndex\":2,\"TopicTitle\":\"vote your hate blockchain\",\"VoteResults\":[{\"SelectOption\":\"ptn1\",\"Num\":1},{\"SelectOption\":\"btc1\",\"Num\":1},{\"SelectOption\":\"eth1\",\"Num\":0},{\"SelectOption\":\"eos1\",\"Num\":0}]}],\"AssetID\":\"VOTE+0GWEXZZNEM91F1MJL3X\"}"
    #data = createToken().ccqueryVoteResult(dict,CreateAddr='P1D4GXUcrcT7tAhMPvTxTQk7vQQSJkiad54',AssetID='VOTE+0GWEXZZNEM91F1MJL3X',TokenAmount=60000)
    dict = '{"VOTE+0GB77ABWKPPOPZW3TVP": "8000", "PTN": "16000.002241", "VOTE+0GEXCOMQDBK5XK9W7XF": "60000", "VOTE+0GK301J75JEKVBJTAS4": "60000"}'
    dict = json.loads(dict)
    
    ctObj = createToken()
    
    print ctObj.getTokenIdByNum('VOTE', dict,3)
    
    result = ctObj.getOneTokenInfo('btc')
    result1 = json.loads(result)
    print result1['Status']
    
    print ctObj.getTokenStatus(result)
    