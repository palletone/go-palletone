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
	"strconv"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

//PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf
type DebugChainCode struct {
}

func (d *DebugChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte("ok"))
}

func (d *DebugChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "putstate":
		if len(args) != 2 {
			return shim.Error("must input 2 args: key,value")
		}
		log.Debugf("put state key[%s],value[%s]", args[0], args[1])
		err := stub.PutState(args[0], []byte(args[1]))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "getstate":
		if len(args) != 1 {
			return shim.Error("must input 1 args: key")
		}
		value, err := stub.GetState(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		log.Debugf("get state key[%s],value[%s]", args[0], string(value))
		return shim.Success(value)
	case "add":
		if len(args) != 2 {
			return shim.Error("must input 2 args: a,b")
		}
		a, _ := strconv.Atoi(args[0])
		b, _ := strconv.Atoi(args[1])
		return d.Add(a, b)
	case "payout":
		if len(args) != 2 {
			return shim.Error("must input 2 args: amount,address")
		}
		amt, _ := strconv.Atoi(args[0])
		addr, _ := common.StringToAddress(args[1])
		return d.Payout(stub, uint64(amt), addr)
	case "testError":
		return d.Error(stub)
	case "testAddBalance":
		if len(args) != 2 {
			return shim.Error("must input 2 args: address,amount")
		}
		return d.AddBalance(stub, args[0], args[1])
	case "testGetBalance":
		if len(args) != 1 {
			return shim.Error("must input 1 args: address")
		}
		return d.GetBalance(stub, args[0])
	case "getbalance":
		if len(args) != 1 {
			return shim.Error("must input 1 args: address")
		}
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
func (d *DebugChainCode) Payout(stub shim.ChaincodeStubInterface, a uint64, addr common.Address) pb.Response {
	_, myAddr := stub.GetContractID()
	balance, _ := stub.GetTokenBalance(myAddr, nil)
	if len(balance) == 0 {
		return shim.Error("don't have token balance")
	}
	if a > balance[0].Amount {
		return shim.Error("balance not enough")
	}
	err := stub.PayOutToken(addr.String(), &modules.AmountAsset{
		Amount: a,
		Asset:  balance[0].Asset,
	}, 0)
	if err != nil {
		return shim.Error("payout error:" + err.Error())
	}
	key := "Payout-" + addr.String()
	paid := 0
	val, err := stub.GetState(key)
	if err == nil {
		paid, _ = strconv.Atoi(string(val))
	}
	paid += int(a)
	err = stub.PutState(key, []byte(strconv.Itoa(paid)))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
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
