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

package digitalidcc

import (
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"strconv"
)

type DigitalIdentityChainCode struct {
}

func (d *DigitalIdentityChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("*** SysConfig ChainCode system contract init ***")
	return shim.Success([]byte("ok"))
}

func (d *DigitalIdentityChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "add":
		return d.add(stub, args)
	case "update":
		return d.updateSysConfig(stub, args)
	case "ForfeitureDeposit":

	default:
		return shim.Error("Invoke error")
	}
	return shim.Error("Invoke error")
}

func (d *DigitalIdentityChainCode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	a, _ := strconv.Atoi(args[0])
	b, _ := strconv.Atoi(args[1])
	rspStr := fmt.Sprintf("Value:%d", a+b)
	return shim.Success([]byte(rspStr))
}
func (d *DigitalIdentityChainCode) updateSysConfig(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	key := args[0]
	value := args[1]
	invokeFromAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//基金会地址
	foundationAddress, _ := stub.GetSystemConfig("FoundationAddress")
	if invokeFromAddr != foundationAddress {
		return shim.Error("Only foundation can call this function")
	}
	stub.PutState(key, []byte(value))
	return shim.Success(nil)
}
