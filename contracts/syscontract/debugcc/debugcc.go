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
 *  * @date 2018
 *
 */

package debugcc

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"strconv"
)

type DebugChainCode struct {
}

func (d *DebugChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte("ok"))
}

func (d *DebugChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "add":
		a, _ := strconv.Atoi(args[0])
		b, _ := strconv.Atoi(args[1])
		return d.Add(a, b)
	case "testError":
		return d.Error(stub)
	case "testAddBalance":
		return d.AddBalance(stub, args[0], args[1])
	case "testGetBalance":
		return d.GetBalance(stub, args[0])
	case "getbalance":
		return d.Getbalance(stub, args[0])
	case "getRequesterCert":
		return d.GetRequesterCert(stub)
	case "checkRequesterCert":
		return d.CheckRequesterCert(stub)
	case "ForfeitureDeposit":
	case "getRootCABytes":
		return d.GetRootCABytes(stub)
	default:
		return shim.Error("debug cc Invoke error" + funcName)
	}
	return shim.Error("debug cc Invoke error" + funcName)
}
func (d *DebugChainCode) Error(stub shim.ChaincodeStubInterface) pb.Response {
	stub.PutState("CannotPut", []byte("Your error will ignore this put."))
	return shim.Error("Test Error")
}
func (d *DebugChainCode) Add(a, b int) pb.Response {

	rspStr := fmt.Sprintf("Value:%d", a+b)
	return shim.Success([]byte(rspStr))
}
func (d *DebugChainCode) Getbalance(stub shim.ChaincodeStubInterface, addr string) pb.Response {
	result, err := stub.GetTokenBalance(addr, nil)
	if err != nil {
		return shim.Error(err.Error())
	}
	log.Debugf("GetBalance result:%+v", result)
	b, e := json.Marshal(result)
	if e != nil {
		return shim.Error(e.Error())
	}
	return shim.Success(b)
}

func (d *DebugChainCode) GetRequesterCert(stub shim.ChaincodeStubInterface) pb.Response {
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

func (d *DebugChainCode) CheckRequesterCert(stub shim.ChaincodeStubInterface) pb.Response {
	isValid, err := stub.IsRequesterCertValid()
	//b := []byte{}
	if isValid {
		b, _ := json.Marshal(fmt.Sprintf("Requester cert is valid"))
		return shim.Success(b)
	} else {
		return shim.Error(fmt.Sprintf("Requester cert is invalid, because %s", err.Error()))
	}
}

func (d *DebugChainCode) GetRootCABytes(stub shim.ChaincodeStubInterface) pb.Response {
	val, err := stub.GetState("RootCABytes")
	if err != nil {
		return shim.Error(err.Error())
	}
	b, e := json.Marshal(val)
	if e != nil {
		return shim.Error(e.Error())
	}
	return shim.Success(b)
}
func (d *DebugChainCode) AddBalance(stub shim.ChaincodeStubInterface, account string, amount string) pb.Response {
	amt, _ := strconv.Atoi(amount)
	balanceB, _ := stub.GetState(account)
	balance, _ := strconv.Atoi(string(balanceB))
	balance = balance + amt
	str := strconv.Itoa(balance)
	stub.PutState(account, []byte(str))
	return shim.Success([]byte(str))
}
func (d *DebugChainCode) GetBalance(stub shim.ChaincodeStubInterface, account string) pb.Response {
	balanceB, _ := stub.GetState(account)
	return shim.Success(balanceB)
}
