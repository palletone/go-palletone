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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

// SampleSysCC example simple Chaincode implementation
type SampleSysCC struct {
}

//
var mutex sync.Mutex

//
const (

	//test rate eth:btc = 2:1
	eth_btc_rate = float64(2.0)

	// ETH contract
	contractABI  = "[{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdrawtoken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"tokens\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposittoken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"admin_\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"redeem\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"
	contractAddr = "0x6817Cfb2c442693d850332c3B755B2342Ec4aFB2"
)

// Init initializes the sample system chaincode by storing the key and value
// arguments passed in as parameters
func (t *SampleSysCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//as system chaincodes do not take part in consensus and are part of the system,
	//best practice to do nothing (or very little) in Init.

	fmt.Println("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
	return shim.Success(nil)
}

// Invoke gets the supplied key and if it exists, updates the key with the newly
// supplied value.
func (t *SampleSysCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "multiSigAddrBTC":
		return multiSigAddrBTC(&args, &stub)
	case "withdrawBTC":
		return withdrawBTC(&args, &stub)

	case "multiSigAddrETH":
		return multiSigAddrETH(&args, &stub)
	case "calSigETH":
		return calSigETH(&args, &stub)

	case "putval":
		return putval(&args, &stub)

	case "getval":
		return getval(&args, &stub)

	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func putval(args *[]string, stub *shim.ChaincodeStubInterface) pb.Response {
	if len(*args) < 2 {
		return shim.Error("need 2 args (key and a value)")
	}
	key := (*args)[0]
	val := (*args)[1]
	// Get the state from the ledger
	valbytes, err := (*stub).GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get val for " + key + "\"}"
		return shim.Error(jsonResp)
	}
	fmt.Println("==== valOld demo ==== ", key, string(valbytes))
	// Write the state to the ledger
	err = (*stub).PutState(key, []byte(val))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func getval(args *[]string, stub *shim.ChaincodeStubInterface) pb.Response {
	if len(*args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key to query")
	}
	key := (*args)[0]
	// Get the state from the ledger
	valbytes, err := (*stub).GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}
	fmt.Println("==== valOld demo ==== ", key, string(valbytes))
	if valbytes == nil {
		jsonResp := "{\"Error\":\"Nil val for " + key + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valbytes)
}

func getBTCAddrByPubkey(pubkeyHex string, stub *shim.ChaincodeStubInterface) (string, error) {
	return "", nil
}

//refer to the struct CreateMultiSigParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member. When make a request, reserve one Pubkey for Jury.
//example, want 2/3 MultiSig Address, set alice and bob only, N=3, reserve one for Jury.
//If set 3 pubkeys and N=3, Jury not join in your contract, your contract is out of PalletOne.
type BTCAddress_createMultiSig struct {
	Method     string   `json:"method"`
	PublicKeys []string `json:"publicKeys"`
	N          int      `json:"n"`
	M          int      `json:"m"`
}

//result
type CreateMultiSigResult struct {
	P2ShAddress  string   `json:"p2sh_address"`
	RedeemScript string   `json:"redeem_script"`
	Addresses    []string `json:"addresses"`
}

func multiSigAddrBTC(args *[]string, stub *shim.ChaincodeStubInterface) pb.Response {
	if len(*args) < 2 {
		return shim.Error("need 3 args (chainName and two publicKeys)")
	}
	pubkeyAlice := (*args)[0]
	pubkeyBob := (*args)[1]

	createMultiSigParams := BTCAddress_createMultiSig{Method: "CreateMultiSigAddress"}
	createMultiSigParams.M = 2
	//set alice and bob. Jury set third pubkey, return redeem and address
	createMultiSigParams.N = 3
	createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, pubkeyAlice)
	createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, pubkeyBob)

	//
	reqBytes, err := json.Marshal(createMultiSigParams)
	if err != nil {
		return shim.Error(err.Error())
	}
	result, err := (*stub).OutChainAddress("btc", reqBytes)
	if err != nil {
		return shim.Error(string(result))
	}
	fmt.Println("multiSigAddrBTC Chaincode result ==== ===== ", string(result))

	var createResult CreateMultiSigResult
	err = json.Unmarshal(result, &createResult)
	if err != nil {
		return shim.Error(string(result))
	}

	// Write the state to the ledger
	err = (*stub).PutState("btc_alice", []byte(createResult.Addresses[0]))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = (*stub).PutState("btc_bob", []byte(createResult.Addresses[1]))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = (*stub).PutState("btc_multsigAddr", []byte(createResult.P2ShAddress))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = (*stub).PutState("btc_redeem", []byte(createResult.RedeemScript))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

//refer to the struct DecodeRawTransactionParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member.
type BTCTransactionDecode struct {
	Method string `json:"method"`
	Rawtx  string `json:"rawtx"`
}

//result
type Input struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
}
type Output struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}
type DecodeRawTransactionResult struct {
	Inputs   []Input  `json:"inputs"`
	Outputs  []Output `json:"outputs"`
	Locktime uint32   `json:"locktime"`
}

//
type WithdrawBTCReqTX struct {
	ethAddr string
	btcAddr string
	Inputs  []Input
}

func getEthAddrByBTCTx(transactionhex string, stub *shim.ChaincodeStubInterface) (*WithdrawBTCReqTX, error) {
	//1.get 'btc' address from tx
	btcTXDecode := BTCTransactionDecode{"DecodeRawTransaction", transactionhex}
	reqBytes, err := json.Marshal(btcTXDecode)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Println("Chaincode params ==== ===== ", string(reqBytes))
	result, err := (*stub).OutChainTransaction("btc", reqBytes)
	if err != nil {
		fmt.Println("DecodeRawTransaction Chaincode result error ==== ===== ", err.Error())
		return nil, err
	}
	fmt.Println("DecodeRawTransaction Chaincode result ==== ===== ", string(result))
	var decodeResult DecodeRawTransactionResult
	err = json.Unmarshal(result, &decodeResult)
	if err != nil {
		fmt.Println("json.Unmarshal error ==== ===== ", err.Error())
		return nil, err
	}
	if len(decodeResult.Outputs) != 1 {
		return nil, errors.New("Only 1 output is valid.") //
	}
	btcAddr := decodeResult.Outputs[0].Address

	//2.judge alice or bob
	isFind := false
	ethAddr := ""
	for true {
		//3.get 'eth' address from palletone storage
		btc_alice, err := (*stub).GetState("btc_alice")
		if err != nil {
			return nil, errors.New("{\"Error\":\"Failed to get state for " + "btc_alice" + "\"}")
		}
		if strings.Compare(btcAddr, string(btc_alice)) == 0 { //judge
			eth_alice, err := (*stub).GetState("eth_alice")
			if err != nil {
				return nil, errors.New("{\"Error\":\"Failed to get state for " + "eth_alice" + "\"}")
			}
			ethAddr = string(eth_alice)
			isFind = true
			break
		}
		//3.get 'eth' address from palletone storage
		btc_bob, err := (*stub).GetState("btc_bob")
		if err != nil {
			return nil, errors.New("{\"Error\":\"Failed to get state for " + "btc_bob" + "\"}")
		}
		fmt.Println("btc_bob ==== ===== ", string(btc_bob))
		if strings.Compare(btcAddr, string(btc_bob)) == 0 { //judge
			eth_bob, err := (*stub).GetState("eth_bob")
			if err != nil {
				return nil, errors.New("{\"Error\":\"Failed to get state for " + "eth_bob" + "\"}")
			}
			ethAddr = string(eth_bob)
			isFind = true
			break
		}
		break
	}

	//
	if isFind {
		withdrawBTCReqTx := new(WithdrawBTCReqTX)
		withdrawBTCReqTx.btcAddr = btcAddr
		withdrawBTCReqTx.ethAddr = ethAddr
		withdrawBTCReqTx.Inputs = decodeResult.Inputs
		return withdrawBTCReqTx, nil
	} else {
		return nil, errors.New("BTC Transaction output address invalid.")
	}
}

//refer to the struct GetEventByAddressParams in "github.com/palletone/adaptor/AdaptorETH.go",
//add 'method' member.
type ETHTransaction_getevent struct { //GetEventByAddressParams
	Method       string `json:"method"`
	ContractABI  string `json:"contractABI"`
	ContractAddr string `json:"contractAddr"`
	ConcernAddr  string `json:"concernaddr"`
	StartHeight  string `json:"startheight"`
	EndHeight    string `json:"endheight"`
	EventName    string `json:"eventname"`
}

type GetEventByAddressResult struct {
	Events []string `json:"events"`
}

//need check confirms
func getDepositETHAmount(concernAddr string, eth_redeem_base64 string, stub *shim.ChaincodeStubInterface) (*big.Int, error) {
	//get doposit event log
	getevent := ETHTransaction_getevent{Method: "GetEventByAddress"} // GetJuryAddress
	getevent.ContractABI = contractABI
	getevent.ContractAddr = contractAddr
	getevent.ConcernAddr = concernAddr
	getevent.EventName = "Deposit"
	//
	reqBytes, err := json.Marshal(getevent)
	if err != nil {
		return nil, err
	}
	//
	result, err := (*stub).OutChainTransaction("eth", reqBytes)
	if err != nil {
		return nil, err
	}
	//
	var geteventresult GetEventByAddressResult
	err = json.Unmarshal(result, &geteventresult)
	if err != nil {
		return nil, err
	}

	//event Deposit(address token, address user, uint amount, bytes redeem);
	bigIntAmount := big.NewInt(int64(0))
	for _, event := range geteventresult.Events {
		//example : ["0x0000000000000000000000000000000000000000","0x7d7116a8706ae08baa7f4909e26728fa7a5f0365",1000000000000000000,"fXEWqHBq4Iuqf0kJ4mco+npfA2WqqRmnxGW+mwU2c8Vn1zvoYDF5Y2xxEEgpIOCvFJqCGJJR8pKoQUioW3zXDQ=="]
		strArray := strings.Split(event, ",")
		//token 0x0 is ETH, example : ["0x0000000000000000000000000000000000000000"
		str0 := strArray[0][2 : len(strArray[0])-1]
		if strings.Compare(str0, "0x0000000000000000000000000000000000000000") != 0 {
			continue
		}

		//user is eth sender, example : "0x7d7116a8706ae08baa7f4909e26728fa7a5f0365"
		str1 := strArray[1][1 : len(strArray[1])-1]
		if strings.Compare(str1, concernAddr) != 0 {
			continue
		}
		//eth_redeem's base64, example : "fXEWqHBq4Iuqf0kJ4mco+npfA2WqqRmnxGW+mwU2c8Vn1zvoYDF5Y2xxEEgpIOCvFJqCGJJR8pKoQUioW3zXDQ=="]
		str3 := strArray[3][1 : len(strArray[3])-2]
		if strings.Compare(str3, eth_redeem_base64) != 0 {
			continue
		}

		//deposit amount, example : 1000000000000000000
		str2 := strArray[2]
		bigInt := new(big.Int)
		bigInt.SetString(str2, 10)
		bigIntAmount = bigIntAmount.Add(bigIntAmount, bigInt)
	}

	return bigIntAmount, nil

}

func getWithdrawETHAmount(concernAddr string, eth_redeem_base64 string, stub *shim.ChaincodeStubInterface) (*big.Int, error) {
	//get Withdraw event log
	getevent := ETHTransaction_getevent{Method: "GetEventByAddress"} // GetJuryAddress
	getevent.ContractABI = contractABI
	getevent.ContractAddr = contractAddr
	getevent.ConcernAddr = concernAddr
	getevent.EventName = "Withdraw"
	//
	reqBytes, err := json.Marshal(getevent)
	if err != nil {
		return nil, err
	}
	//
	result, err := (*stub).OutChainTransaction("eth", reqBytes)
	if err != nil {
		return nil, err
	}
	//
	var geteventresult GetEventByAddressResult
	err = json.Unmarshal(result, &geteventresult)
	if err != nil {
		return nil, err
	}

	//event Withdraw(address token, address user, bytes redeem, address recver, uint amount, uint confirmvalue, string state);
	bigIntAmount := new(big.Int)
	for _, event := range geteventresult.Events {
		//example : ["0x0000000000000000000000000000000000000000","0xaaa919a7c465be9b053673c567d73be860317963","fXEWqHBq4Iuqf0kJ4mco+npfA2WqqRmnxGW+mwU2c8Vn1zvoYDF5Y2xxEEgpIOCvFJqCGJJR8pKoQUioW3zXDQ==","0xaaa919a7c465be9b053673c567d73be860317963",1000000000000000000,2,"withdraw"]
		strArray := strings.Split(event, ",")
		//token 0x0 is ETH, example : ["0x0000000000000000000000000000000000000000"
		str0 := strArray[0][2 : len(strArray[0])-1]
		if strings.Compare(str0, "0x0000000000000000000000000000000000000000") != 0 {
			continue
		}

		//user is recver, example : "0xaaa919a7c465be9b053673c567d73be860317963"
		str3 := strArray[1][1 : len(strArray[3])-1]
		if strings.Compare(str3, concernAddr) != 0 {
			continue
		}
		//eth_redeem's base64, example : "fXEWqHBq4Iuqf0kJ4mco+npfA2WqqRmnxGW+mwU2c8Vn1zvoYDF5Y2xxEEgpIOCvFJqCGJJR8pKoQUioW3zXDQ=="
		str2 := strArray[2][1 : len(strArray[2])-1]
		if strings.Compare(str2, eth_redeem_base64) != 0 {
			continue
		}

		//deposit amount, example : 1000000000000000000
		str4 := strArray[4]
		fmt.Println(str4)
		bigInt := new(big.Int)
		bigInt.SetString(str4, 10)
		bigIntAmount = bigIntAmount.Add(bigIntAmount, bigInt)
	}

	return bigIntAmount, nil
}

//refer to the struct GetTransactionByHashParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member.
type BTCTransaction_getTxByHash struct { //GetTransactionByHashParams
	Method string `json:"method"`
	TxHash string `json:"txhash"`
}

//result
type GetTransactionByHashResult struct {
	Confirms uint64        `json:"confirms"`
	Outputs  []OutputIndex `json:"outputs"`
}

func getReqBTCAmountByInput(withdrawBTCReqTx *WithdrawBTCReqTX, stub *shim.ChaincodeStubInterface) (int64, error) {
	//
	btc_multsigAddr, err := (*stub).GetState("btc_multsigAddr")
	if err != nil {
		return 0, errors.New("{\"Error\":\"Failed to get state for " + "btc_multsigAddr" + "\"}")
	}

	int64Amount := int64(0)
	for i := range withdrawBTCReqTx.Inputs {
		getTxByHash := BTCTransaction_getTxByHash{Method: "GetTransactionByHash"}
		getTxByHash.TxHash = withdrawBTCReqTx.Inputs[i].Txid

		//
		reqBytes, err := json.Marshal(getTxByHash)
		if err != nil {
			return 0, err
		}
		result, err := (*stub).OutChainTransaction("btc", reqBytes)
		if err != nil {
			return 0, err
		}

		//
		var getTxResult GetTransactionByHashResult
		err = json.Unmarshal(result, &getTxResult)
		if err != nil {
			return 0, err
		}

		//utxo address check
		if withdrawBTCReqTx.Inputs[i].Vout < uint32(len(getTxResult.Outputs)) {
			if getTxResult.Outputs[withdrawBTCReqTx.Inputs[i].Vout].Addr != string(btc_multsigAddr) {
				return 0, errors.New("Only can spend multsigAddr's output.")
			}
			int64Amount += getTxResult.Outputs[withdrawBTCReqTx.Inputs[i].Vout].Value
		}

	}

	return int64Amount, nil

}

//refer to the struct GetTransactionsParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member.
type BTCTransaction_getTxs struct { //GetTransactionsParams
	Method  string `json:"method"`
	Account string `json:"account"`
	Count   int    `json:"count"`
	Skip    int    `json:"skip"`
}

type InputIndex struct {
	TxHash string `json:"txHash"`
	Index  uint32 `json:"index"`
	Addr   string `json:"addr"`
	Value  int64  `json:"value"`
}
type OutputIndex struct {
	Index uint32 `json:"index"`
	Addr  string `json:"addr"`
	Value int64  `json:"value"` //satoshi
}
type Transaction struct {
	TxHash        string        `json:"txHash"`
	Confirms      uint64        `json:"confirms"`
	BlanceChanged int64         `json:"blanceChanged"`
	Inputs        []InputIndex  `json:"inputs"`
	Outputs       []OutputIndex `json:"outputs"`
}

//result
type TransactionsResult struct {
	Transactions []Transaction `json:"transactions"`
}

func getWithdrawBTCAmount(withdrawBTCReqTx *WithdrawBTCReqTX, stub *shim.ChaincodeStubInterface) (int64, error) {
	//
	btc_multsigAddr, err := (*stub).GetState("btc_multsigAddr")
	if err != nil {
		return 0, errors.New("{\"Error\":\"Failed to get state for " + "btc_multsigAddr" + "\"}")
	}

	//
	getTxs := BTCTransaction_getTxs{Method: "GetTransactions"}
	getTxs.Account = string(btc_multsigAddr)
	getTxs.Count = 10
	//
	reqBytes, err := json.Marshal(getTxs)
	if err != nil {
		return 0, err
	}
	result, err := (*stub).OutChainTransaction("btc", reqBytes)
	if err != nil {
		return 0, err
	}

	//
	var txsResult TransactionsResult
	err = json.Unmarshal(result, &txsResult)
	if err != nil {
		return 0, err
	}

	withdrawBtcAmount := int64(0)
	for _, tx := range txsResult.Transactions {
		if tx.BlanceChanged < 0 {
		} else { //withdraw tx
			oneAddr := true
			tempAmount := int64(0)
			for _, oneOutput := range tx.Outputs {
				if oneOutput.Addr == withdrawBTCReqTx.btcAddr { //A withdraw
					tempAmount += oneOutput.Value
				} else {
					oneAddr = false
				}
			}
			if oneAddr { //if (M -> A...), A withdraw, add BalanceChanged to Amount
				withdrawBtcAmount += tx.BlanceChanged
			} else { //if (M -> A... + B... + ...), only use A's values
				withdrawBtcAmount += tempAmount
			}
		}
	}
	return withdrawBtcAmount, nil
}

func checkBalanceForWithdrawBTC(withdrawBTCReqTx *WithdrawBTCReqTX, stub *shim.ChaincodeStubInterface) (bool, error) {
	//
	eth_redeem, err := (*stub).GetState("eth_redeem")
	if err != nil {
		return false, errors.New("{\"Error\":\"Failed to get state for " + "eth_redeem" + "\"}")
	}
	//eth_redeem base64
	eth_redeem_bytes, err := hex.DecodeString(string(eth_redeem))
	if err != nil {
		return false, err
	}
	eth_redeem_base64 := base64.StdEncoding.EncodeToString(eth_redeem_bytes)

	//2.1 get deposit and withdraw for calculate balance
	depositETHAmount, err := getDepositETHAmount(withdrawBTCReqTx.ethAddr, eth_redeem_base64, stub)
	if depositETHAmount == nil || err != nil {
		return false, err
	}
	withdrawETHAmount, err := getWithdrawETHAmount(withdrawBTCReqTx.ethAddr, eth_redeem_base64, stub)
	if withdrawETHAmount == nil || err != nil {
		return false, err
	}
	bigIntBalance := new(big.Int)
	bigIntBalance.Sub(depositETHAmount, withdrawETHAmount)
	//check
	if bigIntBalance.Cmp(big.NewInt(0)) <= 0 {
		return false, errors.New("You need doposit ETH. If you are Alice, you can invoke cancelTx func after 24 hours.")
	}

	//2.2 check withdraw inputs is multisig or not and return amount
	btcAmount, err := getReqBTCAmountByInput(withdrawBTCReqTx, stub)
	if err != nil {
		return false, err
	}
	//2.3 check user have been withdrawed other utxos by get multisig address's spend history
	withdrawBTCAmount, err := getWithdrawBTCAmount(withdrawBTCReqTx, stub)
	if err != nil {
		return false, err
	}
	//2.3 have withdrawed + 2.2 request
	weiBTCAmount := new(big.Float)
	weiBTCAmount.SetInt64(withdrawBTCAmount + btcAmount)
	weiBTCAmount.Mul(weiBTCAmount, big.NewFloat(1e10))         //(satoshi:e8 * e10 -> wei:e18)
	weiBTCAmount.Mul(weiBTCAmount, big.NewFloat(eth_btc_rate)) //rate

	//
	ethBalance := new(big.Float)
	ethBalance.SetInt(bigIntBalance)
	if ethBalance.Cmp(weiBTCAmount) >= 0 { //need bigger or equal than btc
		return true, nil
	} else {
		return false, errors.New("You need doposit ETH more.")
	}
}

//refer to the struct SignTxSendParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member.
type BTCTransaction_signTxSend struct { //SignTxSendParams
	Method         string `json:"method"`
	TransactionHex string `json:"transactionhex"`
	RedeemHex      string `json:"redeemhex"`
}

func withdrawBTC(args *[]string, stub *shim.ChaincodeStubInterface) pb.Response {
	if len(*args) < 1 {
		return shim.Error("need 3 args (chainName, transactionhex and redeemhex)")
	}
	transactionhex := (*args)[0]

	//1.get 'eth' address for next step
	withdrawBTCReqTx, err := getEthAddrByBTCTx(transactionhex, stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	//
	mutex.Lock()
	defer mutex.Unlock()

	//2.check 'eth' address is deposited and have been withdrawed or not
	ok, err := checkBalanceForWithdrawBTC(withdrawBTCReqTx, stub)
	if !ok {
		return shim.Error(err.Error())
	}

	// Write the state to the ledger
	err = (*stub).PutState("btc_withdraw_by_eth", []byte("yes"))
	if err != nil {
		return shim.Error(err.Error())
	}

	//
	btc_redeem, err := (*stub).GetState("btc_redeem")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + "btc_redeem" + "\"}"
		return shim.Error(jsonResp)
	}

	//
	btcTX := BTCTransaction_signTxSend{"SignTxSend", transactionhex, string(btc_redeem)}
	reqBytes, err := json.Marshal(btcTX)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("Chaincode params ==== ===== ", string(reqBytes))
	result, err := (*stub).OutChainTransaction("btc", reqBytes)
	if err != nil {
		fmt.Println("withdrawBTC Chaincode result error ==== ===== ", err.Error())
		return shim.Error(err.Error())
	}
	fmt.Println("withdrawBTC Chaincode result ==== ===== ", string(result))

	// Write the state to the ledger
	err = (*stub).PutState("btc_withdraw_by_eth", []byte("success"))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

//Write a 2/3 MultiSig contract, set ETH MultiSigAddr fromat ' addr1 | addr2 | addrJury '
//refer to the struct CreateMultiSigAddressParams in "github.com/palletone/adaptor/AdaptorETH.go",
//add 'method' member. When make a request, reserve one Address for Jury.
//example, want 2/3 MultiSig Address, set alice and bob only, N=3, reserve one for Jury.
//If set 3 address and N=3, Jury not join in your contract, your contract is out of PalletOne.
type ETHAddress_createMultiSig struct {
	Method    string   `json:"method"`
	Addresses []string `json:"addresses"`
	N         int      `json:"n"`
	M         int      `json:"m"`
}
type CreateMultiSigAddressResult struct {
	RedeemHex string `json:"redeemhex"`
}

func multiSigAddrETH(args *[]string, stub *shim.ChaincodeStubInterface) pb.Response {
	if len(*args) < 2 {
		return shim.Error("need 3 args (chainName and two addresses)")
	}
	addrAlice := (*args)[0]
	addrBob := (*args)[1]

	addrAlice = strings.ToLower(addrAlice)
	addrBob = strings.ToLower(addrBob)

	createMultiSigParams := ETHAddress_createMultiSig{Method: "CreateMultiSigAddress"} // GetJuryAddress
	createMultiSigParams.N = 3
	createMultiSigParams.Addresses = append(createMultiSigParams.Addresses, addrAlice)
	createMultiSigParams.Addresses = append(createMultiSigParams.Addresses, addrBob)

	reqBytes, err := json.Marshal(createMultiSigParams)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Chaincode params ==== ===== ", reqBytes)
	result, err := (*stub).OutChainAddress("eth", reqBytes)
	if err != nil {
		fmt.Println("multiSigAddrETH Chaincode result error ==== ===== ", err.Error())
		return shim.Error(string(result))
	}
	fmt.Println("multiSigAddrETH Chaincode result ==== ===== ", string(result))

	var createResult CreateMultiSigAddressResult
	err = json.Unmarshal(result, &createResult)
	if err != nil {
		return shim.Error(string(result))
	}

	err = (*stub).PutState("eth_alice", []byte(addrAlice))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = (*stub).PutState("eth_bob", []byte(addrBob))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = (*stub).PutState("eth_redeem", []byte(createResult.RedeemHex))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

type CalETHSigReq struct {
	ethAddr string
	btcAddr string
	amount  string
}

func getBTCAddrByETHReq(ethRecvAddr string, stub *shim.ChaincodeStubInterface) (*CalETHSigReq, error) {
	//1.judge alice or bob
	isFind := false
	btcAddr := ""
	for true {
		//2.get 'eth' address from palletone storage
		eth_alice, err := (*stub).GetState("eth_alice")
		if err != nil {
			return nil, errors.New("{\"Error\":\"Failed to get state for " + "eth_alice" + "\"}")
		}
		if strings.Compare(ethRecvAddr, string(eth_alice)) == 0 { //judge
			btc_alice, err := (*stub).GetState("btc_alice")
			if err != nil {
				return nil, errors.New("{\"Error\":\"Failed to get state for " + "btc_alice" + "\"}")
			}
			btcAddr = string(btc_alice)
			isFind = true
			break
		}
		//2.get 'eth' address from palletone storage
		eth_bob, err := (*stub).GetState("eth_bob")
		if err != nil {
			return nil, errors.New("{\"Error\":\"Failed to get state for " + "eth_bob" + "\"}")
		}
		if strings.Compare(ethRecvAddr, string(eth_bob)) == 0 { //judge
			btc_bob, err := (*stub).GetState("btc_bob")
			if err != nil {
				return nil, errors.New("{\"Error\":\"Failed to get state for " + "btc_bob" + "\"}")
			}
			btcAddr = string(btc_bob)
			isFind = true
			break
		}
		break
	}

	//
	if isFind {
		calETHSigReq := new(CalETHSigReq)
		calETHSigReq.btcAddr = btcAddr
		calETHSigReq.ethAddr = ethRecvAddr
		return calETHSigReq, nil
	} else {
		return nil, errors.New("ETH address invalid.")
	}
}

func getBTCBalance(calETHSigReq *CalETHSigReq, stub *shim.ChaincodeStubInterface) (*big.Int, *big.Int, error) {
	//
	btc_multsigAddr, err := (*stub).GetState("btc_multsigAddr")
	if err != nil {
		return nil, nil, errors.New("{\"Error\":\"Failed to get state for " + "btc_multsigAddr" + "\"}")
	}

	//
	getTxs := BTCTransaction_getTxs{Method: "GetTransactions"}
	getTxs.Account = string(btc_multsigAddr)
	getTxs.Count = 10
	//
	reqBytes, err := json.Marshal(getTxs)
	if err != nil {
		return nil, nil, err
	}
	result, err := (*stub).OutChainTransaction("btc", reqBytes)
	if err != nil {
		return nil, nil, err
	}

	//
	var txsResult TransactionsResult
	err = json.Unmarshal(result, &txsResult)
	if err != nil {
		return nil, nil, err
	}

	depositBtcAmount := int64(0)
	withdrawBtcAmount := int64(0)
	for _, tx := range txsResult.Transactions {
		if tx.BlanceChanged > 0 { //deposit tx
			if tx.Confirms < 6 { //check confirms
				continue
			}
			oneAddr := true
			tempAmount := int64(0)
			for _, oneInput := range tx.Inputs {
				if oneInput.Addr == calETHSigReq.btcAddr { //A deposit
					tempAmount += oneInput.Value
				} else {
					oneAddr = false
				}
			}
			if oneAddr { //if (A... -> M), A deposit, add BalanceChanged to Amount
				depositBtcAmount += tx.BlanceChanged
			} else { //if (A... + B... + ... -> M), only use A's values
				depositBtcAmount += tempAmount
			}
		} else { //withdraw tx
			oneAddr := true
			tempAmount := int64(0)
			for _, oneOutput := range tx.Outputs {
				if oneOutput.Addr == calETHSigReq.btcAddr { //A withdraw
					tempAmount += oneOutput.Value
				} else {
					oneAddr = false
				}
			}
			if oneAddr { //if (M -> A...), A withdraw, add BalanceChanged to Amount
				withdrawBtcAmount += tx.BlanceChanged
			} else { //if (M -> A... + B... + ...), only use A's values
				withdrawBtcAmount += tempAmount
			}
		}
	}
	bigIntDepositAmount := big.NewInt(depositBtcAmount)
	bigIntWithdrawAmount := big.NewInt(withdrawBtcAmount)
	return bigIntDepositAmount, bigIntWithdrawAmount, nil
}

func checkBalanceForCalSigETH(calETHSigReq *CalETHSigReq, stub *shim.ChaincodeStubInterface) (bool, error) {
	//2.1 get deposit and withdraw for calculate balance
	depositBTCAmount, withdrawBTCAmount, err := getBTCBalance(calETHSigReq, stub)
	if depositBTCAmount == nil || withdrawBTCAmount == nil || err != nil {
		return false, err
	}
	btcBalance := new(big.Int)
	btcBalance.Sub(depositBTCAmount, withdrawBTCAmount)
	//check
	if btcBalance.Cmp(big.NewInt(0)) <= 0 {
		return false, errors.New("You need doposit BTC. If you are Bob, you can invoke cancelTx func after 24 hours.")
	}

	//
	eth_redeem, err := (*stub).GetState("eth_redeem")
	if err != nil {
		return false, errors.New("{\"Error\":\"Failed to get state for " + "eth_redeem" + "\"}")
	}
	//eth_redeem base64
	eth_redeem_bytes, err := hex.DecodeString(string(eth_redeem))
	if err != nil {
		return false, err
	}
	eth_redeem_base64 := base64.StdEncoding.EncodeToString(eth_redeem_bytes)
	//2.2 check user have been withdrawed other utxos by get multisig address's spend history
	withdrawETHAmount, err := getWithdrawETHAmount(calETHSigReq.ethAddr, eth_redeem_base64, stub)
	if withdrawETHAmount == nil || err != nil {
		return false, err
	}

	//2.3 have withdrawed + 2.2 request
	ethReqAmount := new(big.Int)
	_, ok := ethReqAmount.SetString(calETHSigReq.amount, 10)
	if !ok {
		return false, errors.New("Input amount is Invalid.")
	}
	weiETHSum := new(big.Int)
	weiETHSum.Add(withdrawETHAmount, ethReqAmount)
	//
	weiETHAmount := new(big.Float)
	weiETHAmount.SetInt(weiETHSum)

	//Satoshi -> Wei
	btcBalance.Mul(btcBalance, big.NewInt(1e10)) //(satoshi:e8 * e10 -> wei:e18)
	btcBalanceFloat := new(big.Float)
	btcBalanceFloat.SetInt(btcBalance)
	btcBalanceFloatNew := new(big.Float)
	btcBalanceFloatNew.Mul(btcBalanceFloat, big.NewFloat(eth_btc_rate)) //rate

	//
	if btcBalanceFloatNew.Cmp(weiETHAmount) >= 0 { //need bigger or equal than eth
		return true, nil
	} else {
		return false, errors.New("You need doposit BTC more.")
	}

}

//refer to the struct Keccak256HashPackedSigParams in "github.com/palletone/adaptor/AdaptorETH.go",
//add 'method' member. Remove 'PrivateKeyHex', Jury will set itself when sign.
type ETHTransaction_calSig struct {
	Method     string `json:"method"`
	ParamTypes string `json:"paramtypes"`
	Params     string `json:"params"`
}

func calSigETH(args *[]string, stub *shim.ChaincodeStubInterface) pb.Response {
	if len(*args) < 2 {
		return shim.Error("at least 2 args (ParamTypes and Parames)")
	}
	recver := (*args)[0] //example:0xaAA919a7c465be9b053673C567D73Be860317963
	amount := (*args)[1] //example(wei):1000000000000000000 ( 1 Ether )

	//1.get 'btc' address for next step
	calETHSigReq, err := getBTCAddrByETHReq(strings.ToLower(recver), stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	calETHSigReq.amount = amount
	//2.check 'eth' address is deposited and have been withdrawed or not
	ok, err := checkBalanceForCalSigETH(calETHSigReq, stub)
	if !ok {
		return shim.Error(err.Error())
	}

	//keccak256(abi.encodePacked(redeem, recver, address(this), amount, nonece));
	paramTypesArray := []string{"Bytes", "Address", "Address", "Uint", "Uint"} //eth
	paramTypesJson, err := json.Marshal(paramTypesArray)
	if err != nil {
		return shim.Error(string(err.Error()))
	}

	//
	eth_redeem, err := (*stub).GetState("eth_redeem")
	if err != nil {
		return shim.Error("{\"Error\":\"Failed to get state for " + "eth_redeem" + "\"}")
	}
	//
	var paramsArray []string
	paramsArray = append(paramsArray, string(eth_redeem))
	paramsArray = append(paramsArray, recver)
	paramsArray = append(paramsArray, contractAddr)
	paramsArray = append(paramsArray, amount)
	noneceStr := fmt.Sprintf("%d", 1)
	paramsArray = append(paramsArray, noneceStr)

	paramsJson, err := json.Marshal(paramsArray)
	if err != nil {
		return shim.Error(string(err.Error()))
	}

	ethTX := ETHTransaction_calSig{"Keccak256HashPackedSig", string(paramTypesJson), string(paramsJson)}
	reqBytes, err := json.Marshal(ethTX)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Chaincode params ==== ===== ", string(reqBytes))
	result, err := (*stub).OutChainTransaction("eth", reqBytes)
	if err != nil {
		fmt.Println("calSigETH Chaincode result error ==== ===== ", err.Error())
		return shim.Error(string(result))
	}
	fmt.Println("calSigETH Chaincode result ==== ===== ", string(result))
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(SampleSysCC))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
