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
	"math/big"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

type PTNMain struct {
}

func (p *PTNMain) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *PTNMain) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "setOwner":
		return _setOwner(args, stub)
	case "setETHContract":
		return _setETHContract(args, stub)

	case "payoutPTN":
		return _payoutPTN(args, stub)

	case "put":
		return put(args, stub)
	case "get":
		return get(args, stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

const PTN_ERC20Addr = "0xa54880da9a63cdd2ddacf25af68daf31a1bcc0c9"

//const PTNMapABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ptnToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmap\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addrHex\",\"type\":\"address\"}],\"name\":\"encodeBase58\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_ptnhex\",\"type\":\"address\"},{\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getptnhex\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"ptnhex\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"
const PTNMapABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getptnhex\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

const symbolsOwner = "owner_"
const symbolsContractMap = "eth_map"

const symbolsPayout = "payout_"

func _setOwner(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (PTNAddr)")
	}
	_, err := getOwner(stub)
	if err == nil {
		return shim.Error("Owner has been set")
	}
	err = stub.PutState(symbolsOwner, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getOwner(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsOwner)
	if len(result) == 0 {
		return "", errors.New("Need set Owner")
	}

	return string(result), nil
}

func _setETHContract(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (MapContractAddr)")
	}

	//only owner can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can set")
	}

	err = stub.PutState(symbolsContractMap, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsContractMap failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func getMapAddr(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsContractMap)
	if len(result) == 0 {
		return "", errors.New("Need set MapContractAddr")
	}

	return string(result), nil
}

func _payoutPTN(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1  args (transferTxID)")
	}
	log.Debugf("1")

	//
	txID := args[0]
	if "0x" != txID[0:2] {
		txID = "0x" + txID
	}
	result, _ := stub.GetState(symbolsPayout + txID)
	if len(result) != 0 {
		return shim.Error("The tx has been payout")
	}
	log.Debugf("2")

	//get sender receiver amount
	txResult, err := GetErc20Tx(args[0], stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	log.Debugf("3")

	//check contract address, must be ptn erc20 contract address
	if strings.ToLower(txResult.ContractAddr) != PTN_ERC20Addr {
		return shim.Error("The tx is't PTN contract")
	}
	//checke receiver, must be ptnmap contract address
	mapAddr, err := getMapAddr(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	log.Debugf("4")

	if strings.ToLower(txResult.To) != mapAddr {
		log.Debugf("strings.ToLower(txResult.To): %s, mapAddr: %s ", strings.ToLower(txResult.To), mapAddr)
		return shim.Error("Not send token to the Map contract")
	}
	//check token amount
	bigIntAmout := new(big.Int)
	bigIntAmout.SetString(txResult.Amount, 10)
	bigIntAmout = bigIntAmout.Div(bigIntAmout, big.NewInt(1e10)) //Token's decimal is 18, PTN's decimal is 8
	amt := bigIntAmout.Uint64()
	if amt == 0 {
		return shim.Error("Amount is 0")
	}
	log.Debugf("5")

	//check confirms
	curHeight, err := getHight(stub)
	if curHeight == 0 || err != nil {
		return shim.Error("getHeight failed")
	}
	blockNum, err := strconv.ParseUint(txResult.BlockNumber, 10, 64)
	if curHeight-blockNum < 1 {
		return shim.Error("Need more confirms")
	}
	log.Debugf("6")

	//query ptnmap contract for get ptnAddr
	ptnAddr, err := getPTNHex(mapAddr, txResult.From, stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	log.Debugf("7")

	//get addrPTN
	//ptnAddr := common.HexToAddress(ptnHex).String()
	//ptnAddr := "P" + base58.CheckEncode(common.FromHex(ptnHex), 0)
	if ptnAddr == "P14oLvT2" {
		return shim.Error("Need transfer 1 PTNMap for bind address")
	}
	//save payout history
	err = stub.PutState(symbolsPayout+txID, []byte(ptnAddr+"-"+bigIntAmout.String()))
	if err != nil {
		return shim.Error("write symbolsPayout failed: " + err.Error())
	}
	log.Debugf("8")

	//payout
	asset := modules.NewPTNAsset()
	amtToken := &modules.AmountAsset{Amount: amt, Asset: asset}
	err = stub.PayOutToken(ptnAddr, amtToken, 0)
	if err != nil {
		return shim.Error("PayOutToken failed: " + err.Error())
	}
	log.Debugf("9")

	return shim.Success([]byte("Success"))
}

//refer to the struct GetErc20TxByHashParams in "github.com/palletone/adaptor/AdaptorETH.go",
type ETH_GetErc20TxByHash struct {
	Hash string `json:"hash"`
}
type GetErc20TxByHashResult struct {
	Hash         string `json:"hash"`
	Status       string `json:"status"`
	BlockHash    string `json:"blockHash"`
	BlockNumber  string `json:"blockNumber"`
	ContractAddr string `json:"contractaddr"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
}

func GetErc20Tx(hash string, stub shim.ChaincodeStubInterface) (*GetErc20TxByHashResult, error) {
	ethGetTX := ETH_GetErc20TxByHash{hash}
	reqBytes, err := json.Marshal(ethGetTX)
	if err != nil {
		return nil, err
	}
	//
	result, err := stub.OutChainCall("eth", "GetErc20TxByHash", reqBytes)
	if err != nil {
		return nil, errors.New("GetErc20TxByHash error")
	}
	//
	var txResult GetErc20TxByHashResult
	err = json.Unmarshal(result, &txResult)
	if err != nil {
		return nil, err
	}
	return &txResult, nil
}

//refer to the struct GetBestHeaderParams in "github.com/palletone/adaptor/AdaptorETH.go",
type ETHQuery_GetBestHeader struct { //GetBestHeaderParams
	Number string `json:"Number"` //if empty, return the best header
}

type GetBestHeaderResult struct {
	Number string `json:"number"`
}

func getHight(stub shim.ChaincodeStubInterface) (uint64, error) {
	//
	getheader := ETHQuery_GetBestHeader{""} //get best hight
	//
	reqBytes, err := json.Marshal(getheader)
	if err != nil {
		return 0, err
	}
	//
	result, err := stub.OutChainCall("eth", "GetBestHeader", reqBytes)
	if err != nil {
		return 0, err
	}
	//
	var getheadresult GetBestHeaderResult
	err = json.Unmarshal(result, &getheadresult)
	if err != nil {
		return 0, err
	}

	if getheadresult.Number == "" {
		return 0, errors.New("{\"Error\":\"Failed to get eth height\"}")
	}

	curHeight, err := strconv.ParseUint(getheadresult.Number, 10, 64)
	if err != nil {
		return 0, errors.New("{\"Error\":\"Failed to parse eth height\"}")
	}

	return curHeight, nil
}

//refer to the struct QueryContractParams in "github.com/palletone/adaptor/AdaptorETH.go",
type ETH_QueryContract struct {
	ContractABI  string        `json:"contractABI"`
	ContractAddr string        `json:"contractAddr"`
	Method       string        `json:"method"`
	Params       string        `json:"params"`
	ParamsArray  []interface{} `json:"paramsarray"`
}
type QueryContractResult struct {
	Result string `json:"result"`
}

func getPTNHex(mapAddr, sender string, stub shim.ChaincodeStubInterface) (string, error) {
	var queryContract ETH_QueryContract
	queryContract.ContractAddr = mapAddr
	queryContract.ContractABI = PTNMapABI
	queryContract.Method = "getptnhex"
	//senderAddr := adaptoreth.HexToAddress(sender)
	//queryContract.ParamsArray = append(queryContract.ParamsArray, senderAddr)
	params := []string{"0x588eb98f8814aedb056d549c0bafd5ef4963069c"}
	reqBytesParams, _ := json.Marshal(params)
	queryContract.Params = string(reqBytesParams)

	reqBytes, err := json.Marshal(queryContract)
	if err != nil {
		return "", err
	}
	//
	result, err := stub.OutChainCall("eth", "QueryContract", reqBytes)
	if err != nil {
		return "", errors.New("QueryContract failed")
	}
	//
	var getResult QueryContractResult
	err = json.Unmarshal(result, &getResult)
	if err != nil {
		return "", err
	}
	if getResult.Result == "" {
		return "", errors.New("QueryContract result empty")
	}
	var addrs []string
	err = json.Unmarshal(result, &addrs)
	if err != nil || len(addrs) == 0 {
		return "", errors.New("QueryContract result Unmarshal failed")
	}

	return addrs[0], nil
}

func put(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) > 0 {
		err := stub.PutState(args[0], []byte("PutState put"))
		if err != nil {
			log.Debugf("PutState put %s err: %s", args[0], err.Error())
			return shim.Error("PutState put " + args[0] + " failed")
		}
		log.Debugf("PutState put " + args[0] + " ok")
		return shim.Success([]byte("PutState put " + args[0] + " ok"))
	}
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
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(PTNMain))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
