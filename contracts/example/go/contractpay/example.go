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
	"strconv"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Init callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	stub.PutState("paystate0", []byte("paystate0"))
	return shim.Success(nil)
}

//to_address,asset,amount
func (t *SimpleChaincode) payout(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	to_address := args[0]
	asset, _ := modules.StringToAsset(args[1])
	amt, _ := strconv.Atoi(args[2])
	amtToken := &modules.AmountAsset{Amount: uint64(amt), Asset: asset}
	stub.PayOutToken(to_address, amtToken, 0)
	fmt.Println("Payout token" + amtToken.String() + " to address " + to_address)
	return shim.Success(nil)
}

func (t *SimpleChaincode) paystate1(stub shim.ChaincodeStubInterface) pb.Response {
	stub.PutState("paystate1", []byte("paystate1"))
	return shim.Success(nil)
}

func (t *SimpleChaincode) paystate2(stub shim.ChaincodeStubInterface) pb.Response {
	stub.PutState("paystate2", []byte("paystate2"))
	return shim.Success(nil)
}

func (t *SimpleChaincode) payoutstate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	paystate1Byte, _ := stub.GetState("paystate1")
	if len(paystate1Byte) == 0 {
		jsonResp := "{\"Error\":\"You need call getDepositAddr for get your deposit address\"}"
		return shim.Error(jsonResp)
	}
	log.Debugf("paystate1Byte: %s", string(paystate1Byte))

	paystate2Byte, _ := stub.GetState("paystate2")
	if len(paystate2Byte) == 0 {
		jsonResp := "{\"Error\":\"You need call getDepositAddr for get your deposit address\"}"
		return shim.Error(jsonResp)
	}
	log.Debugf("paystate2Byte: %s", string(paystate2Byte))

	time.Sleep(1 * time.Second)

	stub.PutState("paystate", []byte("paystate"))

	to_address := args[0]
	asset, _ := modules.StringToAsset(args[1])
	amt, _ := strconv.Atoi(args[2])
	amtToken := &modules.AmountAsset{Amount: uint64(amt), Asset: asset}
	stub.PayOutToken(to_address, amtToken, 0)
	fmt.Println("Payout token" + amtToken.String() + " to address " + to_address)
	return shim.Success(nil)
}

func (t *SimpleChaincode) balance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	address := args[0]
	result, err := stub.GetTokenBalance(address, nil)
	if err != nil {
		return shim.Error(err.Error())
	}
	data, err := json.Marshal(result)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(data)
}
func (t *SimpleChaincode) time(stub shim.ChaincodeStubInterface) pb.Response {
	time, err := stub.GetTxTimestamp(10)
	if err != nil {
		return shim.Error(err.Error())
	}
	tStr := strconv.Itoa(int(time.Seconds))
	stub.PutState("time", []byte(tStr))

	return shim.Success([]byte(tStr))
}
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "payout" {
		return t.payout(stub, args)
	}

	if function == "paystate1" {
		return t.paystate1(stub)
	}
	if function == "paystate2" {
		return t.paystate2(stub)
	}
	if function == "payoutstate" {
		return t.payoutstate(stub, args)
	}
	if function == "addr" {
		return t.addr(stub)
	}
	if function == "balance" {
		return t.balance(stub, args)
	}
	if function == "time" {
		return t.time(stub)
	}
	if function == "put" {
		return t.put(stub)
	}
	if function == "get" {
		return t.get(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}
func (t *SimpleChaincode) put(stub shim.ChaincodeStubInterface) pb.Response {

	stub.PutState("b", []byte("b"))
	stub.PutState("a", []byte("a"))
	stub.PutState("c", []byte("c"))
	stub.PutState("a", []byte("aa"))
	return shim.Success(nil)
}
func (t *SimpleChaincode) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) > 0 {
		result, _ := stub.GetState(args[0])
		return shim.Success(result)
	}

	result, err := stub.GetContractAllState()
	if err != nil {
		return shim.Error(err.Error())
	}
	str := ""
	for key, v := range result {
		str += fmt.Sprintf("Key:%s,Value:%#v", key, v)
	}
	return shim.Success([]byte(str))
}

func (t *SimpleChaincode) addr(stub shim.ChaincodeStubInterface) pb.Response {
	contractBuf, contractAddr := stub.GetContractID()
	result := fmt.Sprintf("%x-%s", contractBuf, contractAddr)
	log.Debugf(result)
	return shim.Success([]byte(result))
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
