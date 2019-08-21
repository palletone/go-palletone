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

package samplesyscc2

import (
	"fmt"

	"strconv"

	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

//var instance SampleSysCC2

// SampleSysCC example simple Chaincode implementation
type SampleSysCC2 struct {
}

// Init initializes the sample system chaincode by storing the key and value
// arguments passed in as parameters
func (t *SampleSysCC2) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//as system chaincodes do not take part in consensus and are part of the system,
	//best practice to do nothing (or very little) in Init.

	//fmt.Println("***sample system contract init***")
	return shim.Success(nil)
}

// Invoke gets the supplied key and if it exists, updates the key with the newly
// supplied value.
func (t *SampleSysCC2) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "add":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value2:%d", a+b)
			return shim.Success([]byte(rspStr))
		}
	case "sub":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value2:%d", a-b)
			return shim.Success([]byte(rspStr))
		}
	case "mul":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value2:%d", a*b)
			return shim.Success([]byte(rspStr))
		}
	case "div":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value2:%d", a/b)
			return shim.Success([]byte(rspStr))
		}
	case "putval":
		if len(args) < 2 {
			return shim.Error("need 2 args (key and a value)")
		}

		// Initialize the chaincode
		key := args[0]
		val := args[1]

		_, err := stub.GetState(key)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get val for " + key + "\"}"
			return shim.Error(jsonResp)
		}

		// Write the state to the ledger
		err = stub.PutState(key, []byte(val))
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success(nil)
	case "getval":
		var err error
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting key to query")
		}
		key := args[0]

		// Get the state from the ledger
		valbytes, err := stub.GetState(key)
		//return shim.Success([]byte("abc"))
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
			return shim.Error(jsonResp)
		}

		if valbytes == nil {
			jsonResp := "{\"Error\":\"Nil val for " + key + "\"}"
			return shim.Error(jsonResp)
		}

		return shim.Success(valbytes)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}
