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
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
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
	return shim.Success(res)
}

func (t *SimpleChaincode) test_GetStringArgs(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	params := stub.GetStringArgs()
	res, err := json.Marshal(params)
	if err != nil {
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
	return shim.Success(res)
}

func (t *SimpleChaincode) test_GetArgsSlice(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	argsSlice, err := stub.GetArgsSlice()
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(argsSlice)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
