/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"strings"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	if err := stub.PutState("paystate0", []byte("paystate0")); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "testGetInvokeInfo":
		return t.test_GetInvokeInfo(stub)
	case "testPutState":
		return t.test_PutState(stub, args)
	case "testGetState":
		return t.test_GetState(stub, args)
	case "testPutGlobalState":
		return t.test_PutGlobalState(stub, args)
	case "testGetGlobalState":
		return t.test_GetGlobalState(stub, args)
	case "testDelState":
		return t.test_DelState(stub, args)
	case "testDelGlobalState":
		return t.test_DelGlobalState(stub, args)
	case "testGetStateByPrefix":
		return t.test_GetStateByPrefix(stub, args)
	case "testGetContractAllState":
		return t.test_GetContractAllState(stub)
	case "testGetContractState":
		return t.test_GetContractState(stub, args)
	case "testDefineToken":
		return t.test_DefineToken(stub, args)
	case "testSupplyToken":
		return t.test_SupplyToken(stub, args)
	case "testPayOutToken":
		return t.test_PayOutToken(stub, args)
	case "testGetTokenBalance":
		return t.test_GetTokenBalance(stub, args)
	case "testUseCert":
		return t.test_UseCert(stub)
	case "testSendRecvJury":
		return t.test_SendRecvJury(stub)
	case "testSetEvent":
		return t.test_SetEvent(stub)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}

func (t *SimpleChaincode) test_GetInvokeInfo(stub shim.ChaincodeStubInterface) pb.Response {
	resMap := map[string]interface{}{}
	// getArgs，调用参数列表，byte类型
	newArgs := stub.GetArgs()
	params := make([]string, len(newArgs))
	for _, a := range newArgs {
		params = append(params, string(a))
	}
	resMap["GetArgs"] = params
	// GetStringArgs，调用参数类别，string类型
	strArgs := stub.GetStringArgs()
	strParams := make([]string, len(strArgs))
	//for _, s := range strArgs {
	strParams = append(strParams, strArgs...)
	//}
	resMap["GetStringArgs"] = strParams
	/// GetFunctionAndParameters，调用参数类别，string类型
	fn, fpArgs := stub.GetFunctionAndParameters()
	fpRes := map[string]interface{}{}
	fpParams := make([]string, len(strArgs))
	fpRes["functionName"] = fn
	//for _, s := range fpArgs {
	fpParams = append(fpParams, fpArgs...)
	//}
	fpRes["parameters"] = fpParams
	resMap["GetFunctionAndParameters"] = fpRes
	// GetArgsSlice，调用参数列表，byte类型
	sliceArgs, err := stub.GetArgsSlice()
	if err != nil {
		return shim.Error(err.Error())
	}
	sliceParams := make([]string, len(newArgs))
	for _, a := range sliceArgs {
		sliceParams = append(sliceParams, string(a))
	}
	resMap["GetArgsSlice"] = sliceParams
	// GetTxID
	txid := stub.GetTxID()
	resMap["GetTxID"] = txid
	// GetChannelID
	chid := stub.GetChannelID()
	resMap["GetChannelID"] = chid
	// GetTxTimestamp
	tt, err := stub.GetTxTimestamp(10)
	if err != nil {
		return shim.Error(err.Error())
	}
	resMap["GetTxTimestamp"] = tt.String()
	// GetInvokeAddress
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	resMap["GetInvokeAddress"] = invokeAddr.String()
	// GetInvokeTokens
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Error(err.Error())
	}
	sinvoketokens := make([]map[string]interface{}, len(invokeTokens))
	for _, tokenInfo := range invokeTokens {
		oneToken := map[string]interface{}{}
		oneToken["assetId"] = tokenInfo.Asset.AssetId.String()
		oneToken["uniqueId"] = tokenInfo.Asset.UniqueId.String()
		oneToken["amount"] = tokenInfo.Amount
		oneToken["address"] = tokenInfo.Address
		sinvoketokens = append(sinvoketokens, oneToken)
	}
	resMap["GetInvokeTokens"] = sinvoketokens
	// GetInvokeFees
	invokeFees, err := stub.GetInvokeFees()
	if err != nil {
		return shim.Error(err.Error())
	}
	oneFee := map[string]interface{}{}
	oneFee["assetId"] = invokeFees.Asset.AssetId.String()
	oneFee["uniqueId"] = invokeFees.Asset.UniqueId.String()
	oneFee["amount"] = invokeFees.Amount
	resMap["GetInvokeFees"] = oneFee
	// GetInvokeParameters
	invokeAddr, invokeTokens, invokeFees, funcName, params, err := stub.GetInvokeParameters()
	if err != nil {
		return shim.Error(err.Error())
	}
	GIP := map[string]interface{}{}
	GIP["invokeAddress"] = invokeAddr.String()
	GIP["funcName"] = funcName
	gipt := make([]map[string]interface{}, len(invokeTokens))
	for _, tokenInfo := range invokeTokens {
		oneToken := map[string]interface{}{}
		oneToken["assetId"] = tokenInfo.Asset.AssetId.String()
		oneToken["uniqueId"] = tokenInfo.Asset.UniqueId.String()
		oneToken["amount"] = tokenInfo.Amount
		oneToken["address"] = tokenInfo.Address
		gipt = append(gipt, oneToken)
	}
	GIP["invokeTokens"] = gipt
	gipf := map[string]interface{}{}
	gipf["assetId"] = invokeFees.Asset.AssetId.String()
	gipf["uniqueId"] = invokeFees.Asset.UniqueId.String()
	gipf["amount"] = invokeFees.Amount
	GIP["invokeFees"] = gipf
	GIP["invokeParams"] = params
	resMap["GetInvokeParameters"] = GIP
	// GetContractID
	_, scontractid := stub.GetContractID()
	resMap["GetContractID"] = scontractid

	res, err := json.Marshal(resMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(res)
}

func (t *SimpleChaincode) test_PutState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("args:<state key><state value>")
	}
	if err := stub.PutState(args[0], []byte(args[1])); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_GetState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("args:<state key>")
	}
	val, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(val)
}

func (t *SimpleChaincode) test_PutGlobalState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("args:<state key><state value>")
	}
	if err := stub.PutGlobalState(args[0], []byte(args[1])); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_GetGlobalState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("args:<state key>")
	}
	val, err := stub.GetGlobalState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(val)
}

func (t *SimpleChaincode) test_GetContractState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("args:<contract address><state key>")
	}
	addr := common.Address{}
	if err := addr.SetString(args[0]); err != nil {
		return shim.Error(err.Error())
	}
	val, err := stub.GetContractState(addr, args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(val)
}

func (t *SimpleChaincode) test_GetStateByPrefix(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("args:<state key prefix>")
	}
	KVs, err := stub.GetStateByPrefix(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	res := map[string]string{}
	for _, kv := range KVs {
		//log.Debug("key:%s, value:%s", kv.Key, string(kv.Value))
		res[kv.Key] = string(kv.Value)
	}
	data, err := json.Marshal(res)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(data)
}

func (t *SimpleChaincode) test_GetContractAllState(stub shim.ChaincodeStubInterface) pb.Response {
	result, err := stub.GetContractAllState()
	if err != nil {
		return shim.Error(err.Error())
	}
	KVs := map[string]string{}
	for key, val := range result {
		KVs[key] = string(val.Value)
	}
	data, err := json.Marshal(KVs)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(data)
}

func (t *SimpleChaincode) test_DelState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("args:<state key>")
	}
	if err := stub.DelState(args[0]); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_DelGlobalState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("args:<golbal state key>")
	}
	if err := stub.DelGlobalState(args[0]); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//func (t *SimpleChaincode) test_GetSystemConfig(stub shim.ChaincodeStubInterface) pb.Response {
//	cp, err := stub.GetSystemConfig()
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//
//	res, err := json.Marshal(cp)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	return shim.Success(res)
//}

func (t *SimpleChaincode) test_GetTokenBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error(fmt.Sprintf("input:<address><token symbol (option) >"))
	}

	asset := &modules.Asset{}
	if len(args) == 2 {
		// 根据symbol查询asset
		if err := asset.SetString(args[1]); err != nil {
			return shim.Error(err.Error())
		}
	} else {
		asset = nil
	}
	tokens, err := stub.GetTokenBalance(args[0], asset)
	if err != nil {
		return shim.Error(err.Error())
	}

	res, err := json.Marshal(tokens)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(res)
}

func (t *SimpleChaincode) test_PayOutToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		return shim.Error(fmt.Sprintf("input:<address><token name><amount>"))
	}
	asset, err := modules.StringToAsset(args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error(err.Error())
	}
	amountAsset := &modules.AmountAsset{}
	amountAsset.Amount = uint64(amount)
	amountAsset.Asset = asset

	if err := stub.PayOutToken(args[0], amountAsset, 0); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_SetEvent(stub shim.ChaincodeStubInterface) pb.Response {
	message := "Event send data is here!"
	err := stub.SetEvent("evtsender", []byte(message))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_DefineToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// stub.DefineToken(byte(dm.AssetType_FungibleToken), createJson, createAddr.String())
	if len(args) != 4 {
		return shim.Error("input: <token name> <token symbol> <token dedimals> <token total>")
	}

	addr, err := stub.GetInvokeAddress()
	if err != nil {
		shim.Error(err.Error())
	}
	symbol := strings.ToUpper(args[1])
	decimal, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error(err.Error())
	}
	total, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error(err.Error())
	}
	fungible := modules.FungibleToken{
		Name:          args[0],
		Symbol:        symbol,
		Decimals:      byte(decimal),
		TotalSupply:   uint64(total),
		SupplyAddress: addr.String(),
	}
	createJson, err := json.Marshal(fungible)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate token Json\"}"
		return shim.Error(jsonResp)
	}

	if err := stub.DefineToken(byte(modules.AssetType_FungibleToken), createJson, addr.String()); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_SupplyToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		return shim.Error("need 3 args (Symbol,SupplyAmout,Decimals)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//supply amount
	supplyAmount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert supply amount\"}"
		return shim.Error(jsonResp)
	}
	if supplyAmount == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return shim.Error(jsonResp)
	}
	// supply decimals
	decimals, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert supply amount\"}"
		return shim.Error(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}

	//call SupplyToken
	txid := stub.GetTxID()
	assetID, _ := modules.NewAssetId(symbol, modules.AssetType_FungibleToken,
		byte(decimals), common.Hex2Bytes(txid[2:]), modules.UniqueIdType_Null)

	if err := stub.SupplyToken(assetID.Bytes(),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, supplyAmount, invokeAddr.String()); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) test_UseCert(stub shim.ChaincodeStubInterface) pb.Response {
	v, err := stub.IsRequesterCertValid()
	if err != nil {
		return shim.Error(err.Error())
	}
	if !v {
		return shim.Error("Certificate used is invalid.")
	}
	certBytes, err := stub.GetRequesterCert()
	if err != nil {
		return shim.Error(err.Error())
	}
	b, e := json.Marshal(certBytes)
	if e != nil {
		return shim.Error(e.Error())
	}
	return shim.Success(b)
}

type JuryMsgAddr struct {
	Address string
	Answer  []byte
}

func (t *SimpleChaincode) test_SendRecvJury(stub shim.ChaincodeStubInterface) pb.Response {
	_, err := stub.SendJury(1, []byte("hello"), []byte("result"))
	if err != nil {
		return shim.Error(fmt.Sprintf("sendresult err: %s", err.Error()))
	}
	result, err := stub.RecvJury(1, []byte("hello"), 2)
	if err != nil {
		err = stub.PutState("result", []byte(err.Error()))
		if err != nil {
			return shim.Error("PutState: " + string(result))
		}
		return shim.Error("RecvJury failed")
	} else {
		var juryMsg []JuryMsgAddr
		err := json.Unmarshal(result, &juryMsg)
		if err != nil {
			return shim.Error("Unmarshal result failed: " + string(result))
		}
		err = stub.PutState("result", result)
		if err != nil {
			return shim.Error("PutState: " + string(result))
		}
		return shim.Success(nil) //test
	}
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
