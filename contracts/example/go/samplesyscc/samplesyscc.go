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

package samplesyscc

import (
	"encoding/json"
	"fmt"

	"strconv"

	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

var instance SampleSysCC

// SampleSysCC example simple Chaincode implementation
type SampleSysCC struct {
	BtcEthMultiAddr map[string]string
	EthBtcMultiAddr map[string]string
}

// Init initializes the sample system chaincode by storing the key and value
// arguments passed in as parameters
func (t *SampleSysCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//as system chaincodes do not take part in consensus and are part of the system,
	//best practice to do nothing (or very little) in Init.

	//fmt.Println("***sample system contract init***")
	stub.PutState("paystate1", []byte("paystate1"))
	return shim.Success(nil)
}

type BTCAddress struct {
	Method     string   `json:"method"`
	PublicKeys []string `json:"publicKeys"`
	N          int      `json:"n"`
	M          int      `json:"m"`
}

type BTCTransaction struct {
	Method         string `json:"method"`
	TransactionHex string `json:"transactionhex"`
	RedeemHex      string `json:"redeemhex"`
}

type BTCQuery struct {
	Method  string `json:"method"`
	Address string `json:"address"`
	Minconf int    `json:"minconf"`
}

type ETHAddress struct {
	Method    string   `json:"method"`
	Addresses []string `json:"addresses"`
	N         int      `json:"n"`
	M         int      `json:"m"`
}

// Invoke gets the supplied key and if it exists, updates the key with the newly
// supplied value.
func (t *SampleSysCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {

	case "addrBTC":
		if len(args) < 3 {
			return shim.Error("need 3 args (chainName and two publicKeys)")
		}
		chain1Name := args[0]
		chain1PubkeyAlice := args[1]
		chain1PubkeyBob := args[2]

		createMultiSigParams := BTCAddress{Method: "GetMultiAddr"}
		createMultiSigParams.M = 2
		createMultiSigParams.N = 3
		createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, chain1PubkeyAlice)
		createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, chain1PubkeyBob)

		reqBytes, err := json.Marshal(createMultiSigParams)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("Chaincode params ==== ===== ", reqBytes)
		result, err := stub.OutChainAddress(chain1Name, reqBytes)
		if err != nil {
			fmt.Println("Chaincode result ==== ===== ", err.Error())
			return shim.Error(string(result))
		}
		fmt.Println("Chaincode result ==== ===== ", string(result))
		return shim.Success(result)

	case "queryBTC":
		if len(args) < 3 {
			return shim.Error("need 3 args (chainName, address and minimum confirms)")
		}
		chain1 := args[0]
		addr := args[1]
		minconfStr := args[2]
		minconf, err := strconv.Atoi(minconfStr)
		if err != nil {
			minconf = 6
		}
		if minconf < 3 {
			minconf = 6
		}

		btcQuery := BTCQuery{"GetBalance", addr, minconf}
		reqBytes, err := json.Marshal(btcQuery)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("Chaincode params ==== ===== ", string(reqBytes))
		result, err := stub.OutChainQuery(chain1, reqBytes)
		if err != nil {
			fmt.Println("Chaincode result ==== ===== ", err.Error())
			return shim.Error(string(result))
		}
		fmt.Println("Chaincode result ==== ===== ", string(result))
		return shim.Success(result)

	case "transactionBTC":
		if len(args) < 2 {
			return shim.Error("need 3 args (chainName, transactionhex and redeemhex)")
		}
		chain1 := args[0]
		transactionhex := args[1]
		redeemhex := ""
		if len(args) > 2 {
			redeemhex = args[2]
		}

		btcTX := BTCTransaction{"SignTransaction", transactionhex, redeemhex}
		reqBytes, err := json.Marshal(btcTX)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("Chaincode params ==== ===== ", string(reqBytes))
		result, err := stub.OutChainTransaction(chain1, reqBytes)
		if err != nil {
			fmt.Println("Chaincode result ==== ===== ", err.Error())
			return shim.Error(string(result))
		}
		fmt.Println("Chaincode result ==== ===== ", string(result))
		return shim.Success(result)

	case "addrETH":
		if len(args) < 3 {
			return shim.Error("need 3 args (chainName and two addresses)")
		}
		chain1Name := args[0]
		chain1AddrAlice := args[1]
		chain1AddrBob := args[2]

		createMultiSigParams := ETHAddress{Method: "GetMultiAddr"} // GetJuryAddress
		createMultiSigParams.N = 3
		createMultiSigParams.Addresses = append(createMultiSigParams.Addresses, chain1AddrAlice)
		createMultiSigParams.Addresses = append(createMultiSigParams.Addresses, chain1AddrBob)

		reqBytes, err := json.Marshal(createMultiSigParams)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("Chaincode params ==== ===== ", reqBytes)
		result, err := stub.OutChainAddress(chain1Name, reqBytes)
		if err != nil {
			fmt.Println("Chaincode result ==== ===== ", err.Error())
			return shim.Error(string(result))
		}
		fmt.Println("Chaincode result ==== ===== ", string(result))
		return shim.Success(result)

	case "transaction":
		return shim.Success(nil)
	case "queryETH":
		return shim.Success(nil)

	case "txAddr":
		return shim.Success(nil)
	case "deposit":
		return shim.Success(nil)

	case "addrPTN":
		return shim.Success(nil)
	case "transactionPTN":
		return shim.Success(nil)
	case "queryPTN":
		return shim.Success(nil)

	case "add":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value:%d", a+b)
			return shim.Success([]byte(rspStr))
		}
	case "sub":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value:%d", a-b)
			return shim.Success([]byte(rspStr))
		}
	case "mul":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value:%d", a*b)
			return shim.Success([]byte(rspStr))
		}
	case "div":
		{
			a, _ := strconv.Atoi(args[0])
			b, _ := strconv.Atoi(args[1])
			rspStr := fmt.Sprintf("Value:%d", a/b)
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
