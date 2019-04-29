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
	fmt.Println("*** DebugChainCode system contract init ***")
	return shim.Success([]byte("ok"))
}

func (d *DebugChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "add":
		return d.add(stub, args)
	case "testAddBalance":
		return d.addBalance(stub, args)
	case "testGetBalance":
		return d.getBalance(stub, args)
	case "getbalance":
		return d.getbalance(stub, args)
	case "getRequesterCert":
		return d.getRequesterCert(stub, args)
	case "checkRequesterCert":
		return d.checkRequesterCert(stub, args)
	case "ForfeitureDeposit":
	case "getRootCABytes":
		return d.getRootCABytes(stub, args)
	default:
		return shim.Error("Invoke error")
	}
	return shim.Error("Invoke error")
}

func (d *DebugChainCode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	a, _ := strconv.Atoi(args[0])
	b, _ := strconv.Atoi(args[1])
	rspStr := fmt.Sprintf("Value:%d", a+b)
	return shim.Success([]byte(rspStr))
}
func (d *DebugChainCode) getbalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	addr := args[0]
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

func (d *DebugChainCode) getRequesterCert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

func (d *DebugChainCode) checkRequesterCert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	isValid, err := stub.IsRequesterCertValid()
	b := []byte{}
	if isValid {
		b, _ = json.Marshal(fmt.Sprintf("Requester cert is valid"))
	} else {
		b, _ = json.Marshal(fmt.Sprintf("Requester cert is invalid, because %s", err.Error()))
	}
	return shim.Success(b)
}

func (d *DebugChainCode) getRootCABytes(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
func (d *DebugChainCode) addBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	account := args[0]
	amt, _ := strconv.Atoi(args[1])
	balanceB, _ := stub.GetState(account)
	balance, _ := strconv.Atoi(string(balanceB))
	balance = balance + amt
	str := strconv.Itoa(balance)
	stub.PutState(account, []byte(str))
	return shim.Success([]byte(str))
}
func (d *DebugChainCode) getBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	account := args[0]

	balanceB, _ := stub.GetState(account)

	return shim.Success(balanceB)
}
