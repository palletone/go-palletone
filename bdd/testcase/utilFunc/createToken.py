#-*- coding=utf-8 -*-
import requests
import json
import random
import re
import string
from time import sleep

class createToken(object):
    #host = 'http://192.168.0.105:8545/'
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
            "method": "wallet_getBalance",
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
            print "Post data:"+str(data)
            response = requests.post(url=self.domain, data=data, headers=self.headers)
            result1 = json.loads(response.content)
            print "Response:"+str(result1)
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
        print reqId
        data = json.dumps(data)
        response = requests.post(url=self.domain, data=data, headers=self.headers)
        print response.content
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

    def getAssetFromDict(self,dict):
        dict = json.loads(dict)
        print "Asset is: " + dict['info']['payment'][1]['outputs'][0]['asset']
        return dict['info']['payment'][1]['outputs'][0]['asset']

if __name__ == '__main__':
    pass
    # createToken().assertDict()
    # dict1 = "{\"IsVoteEnd\":false,\"CreateAddr\":\"P1D4GXUcrcT7tAhMPvTxTQk7vQQSJkiad54\",\"TotalSupply\":60000,\"SupportResults\":[{\"TopicIndex\":1,\"TopicTitle\":\"vote your love blockchain\",\"VoteResults\":[{\"SelectOption\":\"ptn0\",\"Num\":1},{\"SelectOption\":\"btc0\",\"Num\":0},{\"SelectOption\":\"eth0\",\"Num\":0},{\"SelectOption\":\"eos0\",\"Num\":0}]},{\"TopicIndex\":2,\"TopicTitle\":\"vote your hate blockchain\",\"VoteResults\":[{\"SelectOption\":\"ptn1\",\"Num\":1},{\"SelectOption\":\"btc1\",\"Num\":1},{\"SelectOption\":\"eth1\",\"Num\":0},{\"SelectOption\":\"eos1\",\"Num\":0}]}],\"AssetID\":\"VOTE+0GWEXZZNEM91F1MJL3X\"}"
    # data = createToken().ccqueryVoteResult(dict,CreateAddr='P1D4GXUcrcT7tAhMPvTxTQk7vQQSJkiad54',AssetID='VOTE+0GWEXZZNEM91F1MJL3X',TokenAmount=60000)
    # dict2 = '{"VOTE+0GB77ABWKPPOPZW3TVP": "8000", "PTN": "16000.002241", "VOTE+0GEXCOMQDBK5XK9W7XF": "60000", "VOTE+0GK301J75JEKVBJTAS4": "60000"}'
    dict = "{\"item\":\"transaction_info\",\"info\":{\"tx_hash\":\"0x1d264c6c6be069df692f59a2ff1f57a5d07dcedb249fe952ade91fbd4a95eea0\",\"tx_size\":2812,\"payment\":[{\"inputs\":[{\"txid\":\"0x8806301b14ecff7d28c04407262f158a1cbd61579ef19c42e3785dbc0b8c8353\",\"message_index\":0,\"out_index\":1,\"unlock_script\":\"304402203605c14e6515c2f5b21801c328c96caacd65950768cc112b9571f314fe35943a02207815eb379de90dd4dbd28ce7e07cc48fe90e57432fc453d8fe6814b850276afe01 03561de31fb449c459fe3112f0cab84447b4731c36a20a589a7a19d9d5361ad64a\",\"from_address\":\"P16Q7ZwJUh9ujMWk6xoK8pHQJJeoBktT7jZ\"}],\"outputs\":[{\"amount\":25000,\"asset\":\"PTN\",\"to_address\":\"P1QL7vY6tMUXEuqrHqtBktiZdyiRRwic7Qc\",\"lock_script\":\"OP_DUP OP_HASH160 ffe8934af2ea8a410fa8c46753f2943f1e74d66b OP_EQUALVERIFY OP_CHECKSIG\"},{\"amount\":99915996199808454,\"asset\":\"PTN\",\"to_address\":\"P16Q7ZwJUh9ujMWk6xoK8pHQJJeoBktT7jZ\",\"lock_script\":\"OP_DUP OP_HASH160 3b37adb9f13d85631d7a528e8ac6c43a04c28d06 OP_EQUALVERIFY OP_CHECKSIG\"}],\"locktime\":0,\"number\":0},{\"inputs\":[],\"outputs\":[{\"amount\":2000,\"asset\":\"VOTE+0GCM5TH22D23AZB2WW6\",\"to_address\":\"P16Q7ZwJUh9ujMWk6xoK8pHQJJeoBktT7jZ\",\"lock_script\":\"OP_DUP OP_HASH160 3b37adb9f13d85631d7a528e8ac6c43a04c28d06 OP_EQUALVERIFY OP_CHECKSIG\"}],\"locktime\":0,\"number\":3}],\"fee\":1,\"account_state_update\":null,\"data\":[],\"contract_tpl\":null,\"contract_deploy\":null,\"contract_invoke\":{\"row_number\":2,\"contract_id\":\"PCGTta3M4t3yXu8uRgkKvaWd2d8DRLGbeyd\",\"args\":[\"{\\\"invoke_address\\\":\\\"P16Q7ZwJUh9ujMWk6xoK8pHQJJeoBktT7jZ\\\",\\\"invoke_tokens\\\":[{\\\"amount\\\":25000,\\\"asset\\\":\\\"PTN\\\",\\\"address\\\":\\\"P1QL7vY6tMUXEuqrHqtBktiZdyiRRwic7Qc\\\"}],\\\"invoke_fees\\\":{\\\"amount\\\":1,\\\"asset\\\":\\\"PTN\\\"}}\",\"\",\"createToken\",\"zxl test description\",\"1\",\"2000\",\"2022-02-16 20:00:00\",\"[{\\\"TopicTitle\\\":\\\"vote your love blockchain\\\",\\\"SelectOptions\\\":[\\\"ptn0\\\",\\\"btc0\\\",\\\"eth0\\\",\\\"eos0\\\"],\\\"SelectMax\\\":2},{\\\"TopicTitle\\\":\\\"vote your hate blockchain\\\",\\\"SelectOptions\\\":[\\\"ptn1\\\",\\\"btc1\\\",\\\"eth1\\\",\\\"eos1\\\"],\\\"SelectMax\\\":1}]\"],\"read_set\":\"[]\",\"write_set\":\"[{\\\"is_delete\\\":false,\\\"key\\\":\\\"symbol_VOTE+0GCM5TH22D23AZB2WW6\\\",\\\"value\\\":\\\"eyJOYW1lIjoienhsIHRlc3QgZGVzY3JpcHRpb24iLCJTeW1ib2wiOiJWT1RFIiwiQ3JlYXRlQWRkciI6IlAxNlE3WndKVWg5dWpNV2s2eG9LOHBIUUpKZW9Ca3RUN2paIiwiVm90ZVR5cGUiOjEsIlRvdGFsU3VwcGx5IjoyMDAwLCJWb3RlRW5kVGltZSI6IjIwMjItMDItMTZUMjA6MDA6MDBaIiwiVm90ZUNvbnRlbnQiOiJXM3NpVkc5d2FXTlVhWFJzWlNJNkluWnZkR1VnZVc5MWNpQnNiM1psSUdKc2IyTnJZMmhoYVc0aUxDSldiM1JsVW1WemRXeDBjeUk2VzNzaVUyVnNaV04wVDNCMGFXOXVJam9pY0hSdU1DSXNJazUxYlNJNk1IMHNleUpUWld4bFkzUlBjSFJwYjI0aU9pSmlkR013SWl3aVRuVnRJam93ZlN4N0lsTmxiR1ZqZEU5d2RHbHZiaUk2SW1WMGFEQWlMQ0pPZFcwaU9qQjlMSHNpVTJWc1pXTjBUM0IwYVc5dUlqb2laVzl6TUNJc0lrNTFiU0k2TUgxZExDSlRaV3hsWTNSTllYZ2lPako5TEhzaVZHOXdhV05VYVhSc1pTSTZJblp2ZEdVZ2VXOTFjaUJvWVhSbElHSnNiMk5yWTJoaGFXNGlMQ0pXYjNSbFVtVnpkV3gwY3lJNlczc2lVMlZzWldOMFQzQjBhVzl1SWpvaWNIUnVNU0lzSWs1MWJTSTZNSDBzZXlKVFpXeGxZM1JQY0hScGIyNGlPaUppZEdNeElpd2lUblZ0SWpvd2ZTeDdJbE5sYkdWamRFOXdkR2x2YmlJNkltVjBhREVpTENKT2RXMGlPakI5TEhzaVUyVnNaV04wVDNCMGFXOXVJam9pWlc5ek1TSXNJazUxYlNJNk1IMWRMQ0pUWld4bFkzUk5ZWGdpT2pGOVhRPT0iLCJBc3NldElEIjoiVk9URSswR0NNNVRIMjJEMjNBWkIyV1c2In0=\\\"}]\",\"payload\":\"{\\\"name\\\":\\\"zxl test description\\\",\\\"symbol\\\":\\\"VOTE\\\",\\\"vote_type\\\":1,\\\"vote_end_time\\\":\\\"2022-02-16T20:00:00Z\\\",\\\"vote_content\\\":\\\"W3siVG9waWNUaXRsZSI6InZvdGUgeW91ciBsb3ZlIGJsb2NrY2hhaW4iLCJWb3RlUmVzdWx0cyI6W3siU2VsZWN0T3B0aW9uIjoicHRuMCIsIk51bSI6MH0seyJTZWxlY3RPcHRpb24iOiJidGMwIiwiTnVtIjowfSx7IlNlbGVjdE9wdGlvbiI6ImV0aDAiLCJOdW0iOjB9LHsiU2VsZWN0T3B0aW9uIjoiZW9zMCIsIk51bSI6MH1dLCJTZWxlY3RNYXgiOjJ9LHsiVG9waWNUaXRsZSI6InZvdGUgeW91ciBoYXRlIGJsb2NrY2hhaW4iLCJWb3RlUmVzdWx0cyI6W3siU2VsZWN0T3B0aW9uIjoicHRuMSIsIk51bSI6MH0seyJTZWxlY3RPcHRpb24iOiJidGMxIiwiTnVtIjowfSx7IlNlbGVjdE9wdGlvbiI6ImV0aDEiLCJOdW0iOjB9LHsiU2VsZWN0T3B0aW9uIjoiZW9zMSIsIk51bSI6MH1dLCJTZWxlY3RNYXgiOjF9XQ==\\\",\\\"total_supply\\\":2000,\\\"supply_address\\\":\\\"\\\"}\",\"error_code\":0,\"error_message\":\"\"},\"contract_stop\":null,\"signature\":{\"row_number\":4,\"signature_set\":[\"pubkey:03d531b99f854031c4bb026f405a92f8c6be4abc2981d4e18c189c67d70ab2a513,signature:3045022100d164c748dadd22eb6e0f1aab941d54bb4b26360ad81c36964aa685a3dee8c48302204a4108e4d261a5bf88e4993a9f09f0e69c434adfd2be5e8a921178af23e181e5\"]},\"install_request\":null,\"deploy_request\":null,\"invoke_request\":{\"row_number\":1,\"contract_addr\":\"PCGTta3M4t3yXu8uRgkKvaWd2d8DRLGbeyd\",\"Args\":[\"createToken\",\"zxl test description\",\"1\",\"2000\",\"2022-02-16 20:00:00\",\"[{\\\"TopicTitle\\\":\\\"vote your love blockchain\\\",\\\"SelectOptions\\\":[\\\"ptn0\\\",\\\"btc0\\\",\\\"eth0\\\",\\\"eos0\\\"],\\\"SelectMax\\\":2},{\\\"TopicTitle\\\":\\\"vote your hate blockchain\\\",\\\"SelectOptions\\\":[\\\"ptn1\\\",\\\"btc1\\\",\\\"eth1\\\",\\\"eos1\\\"],\\\"SelectMax\\\":1}]\"],\"timeout\":0},\"stop_request\":null,\"unit_hash\":\"0x6b2d7d0023cea89f317772dd4573c881281af3e2e2f9dbe08915e70bf14fdc63\",\"unit_height\":2327,\"timestamp\":\"2019-06-25T18:07:39+08:00\",\"tx_index\":1},\"hex\":\"\"}"
    createToken().getAssetFromDict(dict)