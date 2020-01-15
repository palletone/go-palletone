/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

type ShimJury struct {
}

func (p *ShimJury) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *ShimJury) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "test":
		return test(stub)
	case "test1":
		return test1(stub)
	case "test2":
		return test2(stub)
	case "put":
		return put(stub)
	case "get":
		return get(args, stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

type JuryMsgAddr struct {
	Address string
	Answer  []byte
}

func test(stub shim.ChaincodeStubInterface) pb.Response {
	sendresult, err := stub.SendJury(1, []byte("hello"), []byte("result"))
	if err != nil {
		log.Debugf("sendresult err: %s", err.Error())
		return shim.Error("SendJury failed")
	}
	log.Debugf("sendresult: %s", common.Bytes2Hex(sendresult))
	result, err := stub.RecvJury(1, []byte("hello"), 2)
	if err != nil {
		log.Debugf("result err: %s", err.Error())
		err = stub.PutState("result", []byte(err.Error()))
		if err != nil {
			return shim.Error("PutState: " + string(result))
		}
	} else {
		log.Debugf("result: %s", string(result))
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
	return shim.Success([]byte("RecvJury failed"))
}
func test1(stub shim.ChaincodeStubInterface) pb.Response {
	sendresult, err := stub.SendJury(1, []byte("hello"), []byte("result"))
	if err != nil {
		log.Debugf("sendresult err: %s", err.Error())
		return shim.Error("SendJury failed")
	}
	stub.PutState("result1", sendresult)
	log.Debugf("sendresult: %s", common.Bytes2Hex(sendresult))
	return shim.Success([]byte("RecvJury failed"))
}
func test2( stub shim.ChaincodeStubInterface) pb.Response {
	result, err := stub.RecvJury(1, []byte("hello"), 2)
	if err != nil {
		log.Debugf("result err: %s", err.Error())
		stub.PutState("result2", []byte(err.Error()))
		return shim.Success([]byte("RecvJury failed"))
	} else {
		log.Debugf("result: #%v\n", result)
		var juryMsg []JuryMsgAddr
		err := json.Unmarshal(result, &juryMsg)
		if err != nil {
			return shim.Error("Unmarshal result failed: " + string(result))
		}
		err = stub.PutState("result2", result)
		if err != nil {
			return shim.Error("PutState: " + string(result))
		}

		return shim.Success([]byte("RecvJury OK"))
	}
}
func put(stub shim.ChaincodeStubInterface) pb.Response {
	err := stub.PutState("result", []byte("PutState put"))
	if err != nil {
		log.Debugf("PutState put err: %s", err.Error())
		return shim.Error("PutState put failed")
	}
	log.Debugf("PutState put ok")
	return shim.Success([]byte("PutState put ok"))
}

func get(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) > 0 {
		result, _ := stub.GetState(args[0])
		return shim.Success(result) //test
	}
	result, _ := stub.GetState("result")
	return shim.Success(result) //test
}

func main() {
	err := shim.Start(new(ShimJury))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
