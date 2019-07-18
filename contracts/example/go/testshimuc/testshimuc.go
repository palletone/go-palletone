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
	"github.com/palletone/go-palletone/common/math"
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
	case "testGetArgs":
		return t.test_GetArgs(stub, args)
	case "testGetStringArgs":
		return t.test_GetStringArgs(stub, args)
	case "testGetFunctionAndParameters":
		return t.test_GetFunctionAndParameters(stub, args)
	case "testGetArgsSlice":
		return t.test_GetArgsSlice(stub, args)
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
		return t.test_GetContractAllState(stub, args)
	case "testGetContractState":
		return t.test_GetContractState(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}

func (t *SimpleChaincode) test_GetArgs(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// GetArgs return args in ChaincodeStub
	// invokeInfo, funcName, function params
	newArgs := stub.GetArgs()
	params := make([]string, len(newArgs))
	for _, a := range newArgs {
		params = append(params, string(a))
	}
	res, err := json.Marshal(params)
	if err != nil {
		return shim.Error(err.Error())
	}
	if err := stub.PutState("GetArgs", res); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(res)
}

func (t *SimpleChaincode) test_GetStringArgs(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	params := stub.GetStringArgs()
	res, err := json.Marshal(params)
	if err != nil {
		return shim.Error(err.Error())
	}
	if err := stub.PutState("GetStringArgs", res); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(res)
}

func (t *SimpleChaincode) test_GetFunctionAndParameters(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	funcname, params := stub.GetFunctionAndParameters()
	data := struct {
		funcname string
		params   []string
	}{funcname: funcname, params: params}

	res, err := json.Marshal(data)
	if err != nil {
		return shim.Error(err.Error())
	}
	if err := stub.PutState("GetFunctionAndParameters", res); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(res)
}

func (t *SimpleChaincode) test_GetArgsSlice(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	argsSlice, err := stub.GetArgsSlice()
	if err != nil {
		return shim.Error(err.Error())
	}
	if err := stub.PutState("GetArgsSlice", argsSlice); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(argsSlice)
}

func (t *SimpleChaincode) test_GetInvokeParameters(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// GetArgs return args in ChaincodeStub
	// invokeInfo, funcName, function params
	invokeAddr, invokeTokens, invokeFees, funcName, params, err := stub.GetInvokeParameters()
	if err != nil {
		return shim.Error(err.Error())
	}
	data := struct {
		invokeAddr   string
		invokeTokens []*modules.InvokeTokens
		invokeFees   *modules.AmountAsset
		funcName     string
		params       []string
	}{invokeAddr.String(), invokeTokens, invokeFees, funcName, params}
	res, err := json.Marshal(data)
	if err != nil {
		return shim.Error(err.Error())
	}
	// put stats into it
	if err := stub.PutState("GetInvokeParameters", res); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_GetInvokeAddress(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// GetArgs return args in ChaincodeStub
	// invokeInfo, funcName, function params
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	// put stats into it
	if err := stub.PutState("GetInvokeAddress", []byte(invokeAddr.String())); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_GetInvokeFees(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// GetArgs return args in ChaincodeStub
	// invokeInfo, funcName, function params
	invokeFees, err := stub.GetInvokeFees()
	if err != nil {
		return shim.Error(err.Error())
	}
	res, err := json.Marshal(invokeFees)
	if err != nil {
		return shim.Error(err.Error())
	}
	// put stats into it
	if err := stub.PutState("GetInvokeFees", res); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_GetInvokeTokens(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// GetArgs return args in ChaincodeStub
	// invokeInfo, funcName, function params
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Error(err.Error())
	}
	res, err := json.Marshal(invokeTokens)
	if err != nil {
		return shim.Error(err.Error())
	}
	// put stats into it
	if err := stub.PutState("GetInvokeTokens", res); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_GetTxID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	txid := stub.GetTxID()
	return shim.Success([]byte(txid))
}

func (t *SimpleChaincode) test_GetChannelID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	chid := stub.GetChannelID()
	return shim.Success([]byte(chid))
}

func (t *SimpleChaincode) test_GetContractID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	contractId, _ := stub.GetContractID()
	return shim.Success(contractId)
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

func (t *SimpleChaincode) test_GetContractAllState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

func (t *SimpleChaincode) test_GetTxTimestamp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	ts, err := stub.GetTxTimestamp(10)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(ts.String()))
}

func (t *SimpleChaincode) test_GetSystemConfig(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	cp, err := stub.GetSystemConfig()
	if err != nil {
		return shim.Error(err.Error())
	}

	res, err := json.Marshal(cp)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(res)
}

func (t *SimpleChaincode) test_GetTokenBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error(fmt.Sprintf("input:<address><token name (option) >"))
	}

	asset := &modules.Asset{}
	if len(args) == 2 {
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

	amountAsset := &modules.AmountAsset{}
	if len(args) == 2 {
		asset := &modules.Asset{}
		if err := asset.SetString(args[1]); err != nil {
			return shim.Error(err.Error())
		}
		amountAsset.Asset = asset
		amount, err := strconv.Atoi(args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		amountAsset.Amount = uint64(amount)
	}
	if err := stub.PayOutToken(args[0], amountAsset, 0); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_SetEvent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	message := "Event send data is here!"
	err := stub.SetEvent("evtsender", []byte(message))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_DefineToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// stub.DefineToken(byte(dm.AssetType_FungibleToken), createJson, createAddr.String())
	if len(args) != 5 {
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
	if len(args) < 2 {
		return shim.Error("need 2 args (Symbol,SupplyAmout)")
	}

	//symbol
	symbol := strings.ToUpper(args[0])
	//check name is exist or not
	gTkInfo := modules.GlobalTokenInfo{}
	tkInfoBytes, err := stub.GetGlobalState(modules.GlobalPrefix + symbol)
	if len(tkInfoBytes) == 0 {
		return shim.Error(err.Error())
	}
	if err := json.Unmarshal(tkInfoBytes, &gTkInfo); err != nil {
		return shim.Error(err.Error())
	}

	//check status
	if gTkInfo.Status != 0 {
		jsonResp := "{\"Error\":\"Status is frozen\"}"
		return shim.Error(jsonResp)
	}

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
	if math.MaxInt64-gTkInfo.TotalSupply < supplyAmount {
		jsonResp := "{\"Error\":\"Too big, overflow\"}"
		return shim.Error(jsonResp)
	}

	//get invoke address
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	//check supply address
	if invokeAddr.String() != gTkInfo.SupplyAddr {
		jsonResp := "{\"Error\":\"Not the supply address\"}"
		return shim.Error(jsonResp)
	}

	//call SupplyToken
	assetID := gTkInfo.AssetID
	err = stub.SupplyToken(assetID.Bytes(),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, supplyAmount, gTkInfo.SupplyAddr)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.SupplyToken\"}"
		return shim.Error(jsonResp)
	}

	//add supply
	gTkInfo.TotalSupply += supplyAmount

	info := struct {
		Symbol      string
		CreateAddr  string
		TotalSupply uint64
		Decimals    uint64
		SupplyAddr  string
		AssetID     modules.AssetId
	}{gTkInfo.Symbol, gTkInfo.CreateAddr, gTkInfo.TotalSupply, uint64(assetID.GetDecimal()), gTkInfo.SupplyAddr, assetID}
	val, err := json.Marshal(&info)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState("symbol_"+info.Symbol, val)

	gti := modules.GlobalTokenInfo{Symbol: info.Symbol, TokenType: 1, Status: 0, CreateAddr: info.CreateAddr,
		TotalSupply: info.TotalSupply, SupplyAddr: info.SupplyAddr, AssetID: info.AssetID}
	val, err = json.Marshal(gti)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutGlobalState(modules.GlobalPrefix+gTkInfo.Symbol, val)

	return shim.Success([]byte(""))
}

func (t *SimpleChaincode) test_GetRequesterCert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("input:<certificate raw bytes>")
	}
	certBytes, err := stub.GetRequesterCert()
	if err != nil {
		return shim.Error(err.Error())
	}
	if string(certBytes) != args[0] {
		return shim.Error("Certificate bytes through GetRequesterCert was unexpected")
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) test_IsRequesterCertValid(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	v, err := stub.IsRequesterCertValid()
	if err != nil {
		return shim.Error(err.Error())
	}
	if v != true {
		return shim.Error("Certificate used is invalid.")
	}
	return shim.Success(nil)
}

type JuryMsgAddr struct {
	Address string
	Answer  []byte
}

func (t *SimpleChaincode) test_SendRecvJury(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
		return shim.Success([]byte("")) //test
	}
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}