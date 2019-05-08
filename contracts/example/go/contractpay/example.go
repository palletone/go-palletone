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
	"fmt"
	"strconv"

	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Init callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
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

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "payout" {
		return t.payout(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
