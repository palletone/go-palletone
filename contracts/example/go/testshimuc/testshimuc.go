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
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	stub.PutState("paystate0", []byte("paystate0"))
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
	val, err := stub.GetStateByPrefix(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	data, err := json.Marshal(val)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(data)
}

func (t *SimpleChaincode) test_GetContractAllState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	val, err := stub.GetContractAllState()
	if err != nil {
		return shim.Error(err.Error())
	}
	data, err := json.Marshal(val)
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

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
