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
	"sort"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	dm "github.com/palletone/go-palletone/dag/modules"
)

type ETHPort struct {
}

func (p *ETHPort) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *ETHPort) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "initDepositAddr":
		return _initDepositAddr(stub)
	case "setETHTokenAsset":
		return _setETHTokenAsset(args, stub)
	case "getETHToken":
		return _getETHToken(stub)
	case "setETHContract":
		return _setETHContract(args, stub)
	case "setOwner":
		return _setOwner(args, stub)
	case "withdrawPrepare":
		return _withdrawPrepare(args, stub)
	case "withdrawETH":
		return _withdrawETH(args, stub)
	case "withdrawFee":
		return _withdrawFee(args, stub)

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

const symbolsJuryAddress = "juryAddress"

const symbolsETHAsset = "eth_asset"
const symbolsETHContract = "eth_contract"

const symbolsDeposit = "deposit_"

const symbolsWithdrawPrepare = "withdrawPrepare_"

const symbolsWithdrawFee = "withdrawfee_"
const symbolsOwner = "owner_"

const symbolsWithdraw = "withdraw_"

const consultM = 3
const consultN = 4

const jsonResp1 = "{\"Error\":\"Failed to get contractAddr, need set contractAddr\"}"

// contractABI is same, but contractAddr is not
const contractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"reqid\",\"type\":\"string\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"name\":\"setaddrs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"reqid\",\"type\":\"string\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"},{\"name\":\"sigstr3\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reqid\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"

func consult(stub shim.ChaincodeStubInterface, content []byte, myAnswer []byte) ([]byte, error) {
	sendResult, err := stub.SendJury(2, content, myAnswer)
	if err != nil {
		log.Debugf("SendJury err: %s", err.Error())
		return nil, errors.New("SendJury failed")
	}
	log.Debugf("sendResult: %s", common.Bytes2Hex(sendResult))
	recvResult, err := stub.RecvJury(2, content, 2)
	if err != nil {
		recvResult, err = stub.RecvJury(2, content, 2)
		if err != nil {
			log.Debugf("RecvJury err: %s", err.Error())
			return nil, errors.New("RecvJury failed")
		}
	}
	log.Debugf("recvResult: %s", string(recvResult))
	return recvResult, nil
}

func _initDepositAddr(stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsJuryAddress)
	if len(saveResult) != 0 {
		return shim.Error("DepositAddr has been init")
	}

	//Method:GetJuryETHAddr, return address string
	result, err := stub.OutChainCall("eth", "GetJuryETHAddr", []byte(""))
	if err != nil {
		log.Debugf("OutChainCall GetJuryETHAddr err: %s", err.Error())
		return shim.Error("OutChainCall GetJuryETHAddr failed")
	}

	//
	recvResult, err := consult(stub, []byte("juryETHAddr"), result)
	if err != nil {
		log.Debugf("consult juryETHAddr failed: " + err.Error())
		return shim.Error("consult juryETHAddr failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Error("Unmarshal result failed: " + err.Error())
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) != consultN {
		return shim.Error("RecvJury result's len not enough")
	}

	//
	address := make([]string, 0, len(juryMsg))
	for i := range juryMsg {
		address = append(address, string(juryMsg[i].Answer))
	}
	a := sort.StringSlice(address[0:])
	sort.Sort(a)

	addressJson, err := json.Marshal(address)
	if err != nil {
		return shim.Error("address Marshal failed: " + err.Error())
	}

	// Write the state to the ledger
	err = stub.PutState(symbolsJuryAddress, addressJson)
	if err != nil {
		return shim.Error("write " + symbolsJuryAddress + " failed: " + err.Error())
	}
	return shim.Success(addressJson)
}

func getETHAddrs(stub shim.ChaincodeStubInterface) []string {
	result, _ := stub.GetState(symbolsJuryAddress)
	if len(result) == 0 {
		return []string{}
	}
	var address []string
	err := json.Unmarshal(result, &address)
	if err != nil {
		return []string{}
	}
	return address
}

func _setETHTokenAsset(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (AssetStr)")
	}
	err := stub.PutState(symbolsETHAsset, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsETHAsset failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getETHTokenAsset(stub shim.ChaincodeStubInterface) *dm.Asset {
	result, _ := stub.GetState(symbolsETHAsset)
	if len(result) == 0 {
		return nil
	}
	asset, _ := dm.StringToAsset(string(result))
	log.Debugf("resultHex %s, asset: %s", common.Bytes2Hex(result), asset.String())

	return asset
}

func _setETHContract(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (ETHContractAddr)")
	}
	//
	saveResult, _ := stub.GetState(symbolsETHContract)
	if len(saveResult) != 0 {
		return shim.Error("TokenAsset has been init")
	}

	err := stub.PutState(symbolsETHContract, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsETHContract failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func _setOwner(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (PTNAddr)")
	}
	err := stub.PutState(symbolsOwner, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}
func getETHContract(stub shim.ChaincodeStubInterface) string {
	result, _ := stub.GetState(symbolsETHContract)
	if len(result) == 0 {
		return ""
	}
	log.Debugf("contractAddr: %s", string(result))

	return string(result)
}

//refer to the struct GetBestHeaderParams in "github.com/palletone/adaptor/AdaptorETH.go",
type ETHQuery_GetBestHeader struct { //GetBestHeaderParams
	Number string `json:"Number"` //if empty, return the best header
}

type GetBestHeaderResult struct {
	Number string `json:"number"`
}

func getHight(stub shim.ChaincodeStubInterface) (string, string, error) {
	//
	getheader := ETHQuery_GetBestHeader{""} //get best hight
	//
	reqBytes, err := json.Marshal(getheader)
	if err != nil {
		return "", "", err
	}
	//
	result, err := stub.OutChainCall("eth", "GetBestHeader", reqBytes)
	if err != nil {
		return "", "", err
	}
	//
	var getheadresult GetBestHeaderResult
	err = json.Unmarshal(result, &getheadresult)
	if err != nil {
		return "", "", err
	}

	if getheadresult.Number == "" {
		return "", "", errors.New("{\"Error\":\"Failed to get eth height\"}")
	}

	curHeight, err := strconv.ParseUint(getheadresult.Number, 10, 64)
	if err != nil {
		return "", "", errors.New("{\"Error\":\"Failed to parse eth height\"}")
	}
	curBefore30d := curHeight - 172800 // 30 days
	//curHeight -= 6

	curBefore30dStr := strconv.FormatUint(curBefore30d, 10)
	curHeightStr := strconv.FormatUint(curHeight, 10)
	return curBefore30dStr, curHeightStr, nil
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
	Events    []string `json:"events"`
	Txhashs   []string `json:"txhashs"`
	Blocknums []uint64 `json:"blocknums"`
}

type DepositETHInfo struct {
	Txhash string
	Amount uint64
}

//need check confirms
func getDepositETHInfo(contractAddr, ptnAddr string, stub shim.ChaincodeStubInterface) ([]DepositETHInfo, error) {
	startHeight, endHeight, err := getHight(stub)
	if err != nil {
		log.Debugf("getHight failed %s", err.Error())
		return nil, err
	}
	log.Debugf("startHeight %s, endHeight %s", startHeight, endHeight)
	//get doposit event log
	var getevent ETHTransaction_getevent // GetJuryAddress
	getevent.ContractABI = contractABI
	getevent.ContractAddr = contractAddr
	getevent.ConcernAddr = ptnAddr
	getevent.EventName = "Deposit"
	getevent.StartHeight = startHeight
	getevent.EndHeight = endHeight
	//
	reqBytes, err := json.Marshal(getevent)
	if err != nil {
		return nil, err
	}
	//
	result, err := stub.OutChainCall("eth", "GetEventByAddress", reqBytes)
	if err != nil {
		return nil, err
	}
	//
	var geteventresult GetEventByAddressResult
	err = json.Unmarshal(result, &geteventresult)
	if err != nil {
		return nil, err
	}

	//event Deposit(address token, address user, uint amount, string ptnaddr);
	endBlockNum, _ := strconv.ParseUint(endHeight, 10, 64)
	depositInfo := make([]DepositETHInfo, 0, len(geteventresult.Events))
	for i, event := range geteventresult.Events {
		//Event example : ["0x0000000000000000000000000000000000000000","0x7d7116a8706ae08baa7f4909e26728fa7a5f0365",500000000000000000,"P1DXLJmJh9j3LFNUZ7MmfLVNWHoLzDUHM9A"]
		strArray := strings.Split(event, ",")
		if len(strArray) != 4 {
			log.Debugf("len(strArray) %d", len(strArray))
			continue
		}
		//confirm
		if geteventresult.Blocknums[i]+10 > endBlockNum {
			log.Debugf("geteventresult.Blocknums[i] %d, endBlockNum %d", geteventresult.Blocknums[i], endBlockNum)
			continue
		}
		//deposit amount, example : 500000000000000000
		str2 := strArray[2]
		bigInt := new(big.Int)
		bigInt.SetString(str2, 10)
		bigInt = bigInt.Div(bigInt, big.NewInt(10000000000)) //ethToken's decimal is 8
		//
		depositInfo = append(depositInfo, DepositETHInfo{geteventresult.Txhashs[i], bigInt.Uint64()})
	}
	if len(depositInfo) == 0 {
		log.Debugf("len(depositInfo) is 0")
		return nil, nil
	}

	return depositInfo, nil

}

func _getETHToken(stub shim.ChaincodeStubInterface) pb.Response {
	//
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}

	contractAddr := getETHContract(stub)
	if contractAddr == "" {
		return shim.Error(jsonResp1)
	}

	depositInfo, err := getDepositETHInfo(contractAddr, invokeAddr.String(), stub)
	if depositInfo == nil || err != nil {
		return shim.Error("You need deposit")
	}
	log.Debugf("len(depositInfo) is %d", len(depositInfo))
	//
	ethAmount := uint64(0)
	for i := range depositInfo {
		deposit, _ := stub.GetState(symbolsDeposit + depositInfo[i].Txhash)
		if len(deposit) != 0 {
			continue
		}
		//
		err = stub.PutState(symbolsDeposit+depositInfo[i].Txhash, []byte(invokeAddr.String()))
		if err != nil {
			log.Debugf("PutState sigHash failed err: %s", err.Error())
			return shim.Error("PutState sigHash failed")
		}
		ethAmount += depositInfo[i].Amount
	}

	if ethAmount == 0 {
		return shim.Error("You need deposit or need wait confirm")
	}
	//
	ethTokenAsset := getETHTokenAsset(stub)
	if ethTokenAsset == nil {
		return shim.Error("need call setETHTokenAsset()")
	}
	invokeTokens := new(dm.AmountAsset)
	invokeTokens.Amount = ethAmount
	invokeTokens.Asset = ethTokenAsset
	err = stub.PayOutToken(invokeAddr.String(), invokeTokens, 0)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.PayOutToken\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success([]byte("get success"))
}

type WithdrawPrepare struct {
	EthAddr   string
	EthAmount uint64
	EthFee    uint64
}

func _withdrawPrepare(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (ethAddr, [ethFee(>10000)])")
	}
	ethAddr := args[0]
	ethFee := uint64(0)
	if len(args) > 1 {
		ethFee, _ = strconv.ParseUint(args[1], 10, 64)
	}
	if ethFee <= 10000 { //0.0001eth
		ethFee = 10000
	}
	//
	ethTokenAsset := getETHTokenAsset(stub)
	if ethTokenAsset == nil {
		return shim.Error("need call setETHTokenAsset()")
	}
	//contractAddr
	_, contractAddr := stub.GetContractID()

	//check token
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		jsonResp := "{\"Error\":\"GetInvokeTokens failed\"}"
		return shim.Error(jsonResp)
	}

	ethTokenAmount := uint64(0)
	log.Debugf("contractAddr %s", contractAddr)
	for i := 0; i < len(invokeTokens); i++ {
		log.Debugf("invokeTokens[i].Address %s", invokeTokens[i].Address)
		if invokeTokens[i].Address == contractAddr {
			if invokeTokens[i].Asset.AssetId == ethTokenAsset.AssetId {
				ethTokenAmount += invokeTokens[i].Amount
			}
		}
	}
	if ethTokenAmount == 0 {
		log.Debugf("You need send contractAddr ethToken")
		jsonResp := "{\"Error\":\"You need send contractAddr ethToken\"}"
		return shim.Error(jsonResp)
	}
	log.Debugf("ethTokenAmount %d", ethTokenAmount)

	reqid := stub.GetTxID()
	// 产生交易
	rawTx := fmt.Sprintf("%s %d %s", ethAddr, ethTokenAmount, reqid)
	log.Debugf("rawTx:%s", rawTx)

	tempHash := crypto.Keccak256([]byte(rawTx), []byte("prepare"))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)

	//协商交易
	recvResult, err := consult(stub, []byte(tempHashHex), []byte("rawTx"))
	if err != nil {
		log.Debugf("consult rawTx failed: " + err.Error())
		return shim.Error("consult rawTx failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal rawTx result failed: " + err.Error())
		return shim.Error("Unmarshal rawTx result failed: " + err.Error())
	}
	if len(juryMsg) < consultM {
		log.Debugf("RecvJury rawTx result's len not enough")
		return shim.Error("RecvJury rawTx result's len not enough")
	}

	// 记录Prepare
	var prepare WithdrawPrepare
	prepare.EthAddr = ethAddr
	prepare.EthAmount = ethTokenAmount
	prepare.EthFee = ethFee
	prepareByte, err := json.Marshal(prepare)
	if err != nil {
		log.Debugf("Marshal prepare failed: " + err.Error())
		return shim.Error("Marshal prepare failed: " + err.Error())
	}
	err = stub.PutState(symbolsWithdrawPrepare+reqid, prepareByte)
	if err != nil {
		log.Debugf("save symbolsWithdrawPrepare failed: " + err.Error())
		return shim.Error("save symbolsWithdrawPrepare failed: " + err.Error())
	}

	updateFee(ethFee, stub)

	return shim.Success([]byte("Withdraw is ready, please invoke withdrawETH"))
}

func updateFee(fee uint64, stub shim.ChaincodeStubInterface) {
	feeCur := uint64(0)
	result, _ := stub.GetState(symbolsWithdrawFee)
	if len(result) != 0 {
		log.Debugf("updateFee fee current : %s ", string(result))
		feeCur, _ = strconv.ParseUint(string(result), 10, 64)
	}
	fee += feeCur
	feeStr := fmt.Sprintf("%d", fee)
	err := stub.PutState(symbolsWithdrawFee, []byte(feeStr))
	if err != nil {
		log.Debugf("updateFee failed: " + err.Error())
	}
}

func getFee(stub shim.ChaincodeStubInterface) uint64 {
	feeCur := uint64(0)
	result, _ := stub.GetState(symbolsWithdrawFee)
	if len(result) != 0 {
		log.Debugf("getFee fee current : %s ", string(result))
		feeCur, _ = strconv.ParseUint(string(result), 10, 64)
	}
	return feeCur
}

//refer to the struct Keccak256HashPackedSigParams in "github.com/palletone/adaptor/AdaptorETH.go",
//Remove 'PrivateKeyHex', Jury will set it when sign.
type ETHTransaction_Keccak256HashPackedSig struct {
	ParamTypes string `json:"paramtypes"`
	Params     string `json:"params"`
}
type Keccak256HashPackedSigResult struct {
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
}

func calSig(contractAddr, reqid, ethAddr string, ethAmount uint64, stub shim.ChaincodeStubInterface) (string, string, error) {
	//keccak256(abi.encodePacked(address(this), recver, amount, reqid));
	paramTypesArray := []string{"Address", "Address", "Uint", "String"} //eth
	paramTypesJson, err := json.Marshal(paramTypesArray)
	if err != nil {
		return "", "", err
	}

	//
	var paramsArray []string
	paramsArray = append(paramsArray, contractAddr)
	paramsArray = append(paramsArray, ethAddr)
	ethAmountStr := fmt.Sprintf("%d", ethAmount)
	ethAmountStr = ethAmountStr + "0000000000"
	paramsArray = append(paramsArray, ethAmountStr)
	paramsArray = append(paramsArray, reqid)
	paramsJson, err := json.Marshal(paramsArray)
	if err != nil {
		return "", "", err
	}
	log.Debugf("paramsJson %s", string(paramsJson))

	ethTX := ETHTransaction_Keccak256HashPackedSig{string(paramTypesJson), string(paramsJson)}
	reqBytes, err := json.Marshal(ethTX)
	if err != nil {
		return "", "", err
	}
	//
	result, err := stub.OutChainCall("eth", "Keccak256HashPackedSig", reqBytes)
	if err != nil {
		return "", "", errors.New("Keccak256HashPackedSig error")
	}
	//
	var sigResult Keccak256HashPackedSigResult
	err = json.Unmarshal(result, &sigResult)
	if err != nil {
		return "", "", err
	}
	return sigResult.Hash, sigResult.Signature, nil
}

//refer to the struct RecoverParams in "github.com/palletone/adaptor/AdaptorETH.go",
type ETHTransaction_RecoverAddr struct {
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
}
type RecoverResult struct {
	Addr string `json:"addr"`
}

func recoverAddr(hash, sig string, stub shim.ChaincodeStubInterface) (string, error) {
	ethTX := ETHTransaction_RecoverAddr{hash, sig}
	reqBytes, err := json.Marshal(ethTX)
	if err != nil {
		return "", err
	}
	//
	result, err := stub.OutChainCall("eth", "RecoverAddr", reqBytes)
	if err != nil {
		return "", errors.New("RecoverAddr error")
	}
	//
	var recoverResult RecoverResult
	err = json.Unmarshal(result, &recoverResult)
	if err != nil {
		return "", err
	}
	return recoverResult.Addr, nil
}

func verifySigs(juryMsg []JuryMsgAddr, hash string, addrs []string, stub shim.ChaincodeStubInterface) []string {
	//
	var sigs []string
	for i := range juryMsg {
		addr, err := recoverAddr(hash, string(juryMsg[i].Answer), stub)
		if err != nil {
			continue
		}
		for j := range addrs {
			if addr == addrs[j] {
				sigs = append(sigs, string(juryMsg[i].Answer))
			}
		}
	}
	//sort
	a := sort.StringSlice(sigs[0:])
	sort.Sort(a)
	return sigs
}

type Withdraw struct {
	EthAddr   string
	EthAmount uint64
	EthFee    uint64
	Sigs      []string
}

func _withdrawETH(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (reqid)")
	}
	reqid := args[0]
	if "0x" != reqid[0:2] {
		reqid = "0x" + reqid
	}

	result, _ := stub.GetState(symbolsWithdrawPrepare + reqid)
	if len(result) == 0 {
		return shim.Error("Please invoke withdrawPrepare first")
	}

	// 检查交易
	var prepare WithdrawPrepare
	err := json.Unmarshal(result, &prepare)
	if nil != err {
		jsonResp := "Unmarshal WithdrawPrepare failed"
		return shim.Error(jsonResp)
	}

	contractAddr := getETHContract(stub)
	if contractAddr == "" {
		return shim.Error(jsonResp1)
	}

	// 计算签名
	hash, sig, err := calSig(contractAddr, reqid, prepare.EthAddr, prepare.EthAmount-prepare.EthFee, stub)
	if err != nil {
		return shim.Error("calSig failed: " + err.Error())
	}
	log.Debugf("hash: %s, sig: %s", hash, sig)

	//
	reqidNew := stub.GetTxID()
	rawTx := fmt.Sprintf("%s %d %s", prepare.EthAddr, prepare.EthAmount-prepare.EthFee, reqidNew)
	tempHash := crypto.Keccak256([]byte(rawTx))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)

	//协商交易
	recvResult, err := consult(stub, []byte(tempHashHex), []byte(sig))
	if err != nil {
		log.Debugf("consult sig failed: " + err.Error())
		return shim.Error("consult sig failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal sig result failed: " + err.Error())
		return shim.Error("Unmarshal sig result failed: " + err.Error())
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) < consultM {
		log.Debugf("RecvJury sig result's len not enough")
		return shim.Error("RecvJury sig result's len not enough")
	}

	addrs := getETHAddrs(stub)
	if len(addrs) != consultN {
		log.Debugf("getETHAddrs result's len not enough")
		return shim.Error("getETHAddrs result's len not enough")
	}

	sigs := verifySigs(juryMsg, hash, addrs, stub)
	if len(sigs) < consultM {
		log.Debugf("verifySigs result's len not enough")
		return shim.Error("verifySigs result's len not enough")
	}
	sigsStr := sigs[0]
	for i := 1; i < consultM; i++ {
		sigsStr = sigsStr + sigs[i]
	}
	sigHash := crypto.Keccak256([]byte(sigsStr))
	sigHashHex := fmt.Sprintf("%x", sigHash)
	log.Debugf("start consult sigHashHex %s", sigHashHex)

	//协商 发送交易哈希
	txResult, err := consult(stub, []byte(sigHashHex), []byte("sigHash"))
	if err != nil {
		log.Debugf("consult sigHash failed: " + err.Error())
		return shim.Error("consult sigHash failed: " + err.Error())
	}
	var txJuryMsg []JuryMsgAddr
	err = json.Unmarshal(txResult, &txJuryMsg)
	if err != nil {
		log.Debugf("Unmarshal sigHash result failed: " + err.Error())
		return shim.Error("Unmarshal sigHash result failed: " + err.Error())
	}
	if len(txJuryMsg) < consultM {
		log.Debugf("RecvJury sigHash result's len not enough")
		return shim.Error("RecvJury sigHash result's len not enough")
	}
	//协商 保证协商一致后才写入签名结果
	txResult2, err := consult(stub, []byte(sigHashHex+"twice"), []byte("sigHash2"))
	if err != nil {
		log.Debugf("consult sigHash2 failed: " + err.Error())
		return shim.Error("consult sigHash2 failed: " + err.Error())
	}
	var txJuryMsg2 []JuryMsgAddr
	err = json.Unmarshal(txResult2, &txJuryMsg2)
	if err != nil {
		log.Debugf("Unmarshal sigHash2 result failed: " + err.Error())
		return shim.Error("Unmarshal sigHash2 result failed: " + err.Error())
	}
	if len(txJuryMsg2) < consultM {
		log.Debugf("RecvJury sigHash2 result's len not enough")
		return shim.Error("RecvJury sigHash2 result's len not enough")
	}

	//记录签名
	var withdraw Withdraw
	withdraw.EthAddr = prepare.EthAddr
	withdraw.EthAmount = prepare.EthAmount
	withdraw.EthFee = prepare.EthFee
	withdraw.Sigs = append(withdraw.Sigs, sigs[0:consultM]...)
	withdrawBytes, err := json.Marshal(withdraw)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(symbolsWithdraw+reqidNew, withdrawBytes)
	if err != nil {
		log.Debugf("save withdraw failed: " + err.Error())
		return shim.Error("save withdraw failed: " + err.Error())
	}

	//删除Prepare
	err = stub.DelState(symbolsWithdrawPrepare + reqid)
	if err != nil {
		log.Debugf("delete WithdrawPrepare failed: " + err.Error())
		return shim.Error("delete WithdrawPrepare failed: " + err.Error())
	}

	return shim.Success(withdrawBytes)
}

func _withdrawFee(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (ethAddr)")
	}
	ethAddr := args[0]

	//
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	result, _ := stub.GetState(symbolsOwner)
	if len(result) == 0 {
		return shim.Error("Owner is not set")
	}
	if string(result) != invokeAddr.String() {
		return shim.Error("Must is the Owner")
	}

	//
	ethAmount := getFee(stub)
	if ethAmount == 0 {
		jsonResp := "{\"Error\":\"fee is 0\"}"
		return shim.Error(jsonResp)
	}
	contractAddr := getETHContract(stub)
	if contractAddr == "" {
		return shim.Error(jsonResp1)
	}

	//
	reqid := stub.GetTxID()
	// 计算签名
	hash, sig, err := calSig(contractAddr, reqid, ethAddr, ethAmount, stub)
	if err != nil {
		return shim.Error("calSig failed: " + err.Error())
	}
	log.Debugf("sig:%s", sig)

	rawTx := fmt.Sprintf("%s %d %s", ethAddr, ethAmount, reqid)
	tempHash := crypto.Keccak256([]byte(rawTx))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)

	//协商交易
	recvResult, err := consult(stub, []byte(tempHashHex), []byte(sig))
	if err != nil {
		log.Debugf("consult sig failed: " + err.Error())
		return shim.Error("consult sig failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal sig result failed: " + err.Error())
		return shim.Error("Unmarshal sig result failed: " + err.Error())
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) < consultM {
		log.Debugf("RecvJury sig result's len not enough")
		return shim.Error("RecvJury sig result's len not enough")
	}

	addrs := getETHAddrs(stub)
	if len(addrs) != consultN {
		log.Debugf("getETHAddrs result's len not enough")
		return shim.Error("getETHAddrs result's len not enough")
	}

	sigs := verifySigs(juryMsg, hash, addrs, stub)
	if len(sigs) < consultM {
		log.Debugf("verifySigs result's len not enough")
		return shim.Error("verifySigs result's len not enough")
	}
	sigsStr := sigs[0]
	for i := 1; i < consultM; i++ {
		sigsStr = sigsStr + sigs[i]
	}
	sigHash := crypto.Keccak256([]byte(sigsStr))
	sigHashHex := fmt.Sprintf("%x", sigHash)
	log.Debugf("start consult sigHashHex %s", sigHashHex)

	//协商 发送交易哈希
	txResult, err := consult(stub, []byte(sigHashHex), []byte("sigHash"))
	if err != nil {
		log.Debugf("consult sigHash failed: " + err.Error())
		return shim.Error("consult sigHash failed: " + err.Error())
	}
	var txJuryMsg []JuryMsgAddr
	err = json.Unmarshal(txResult, &txJuryMsg)
	if err != nil {
		log.Debugf("Unmarshal sigHash result failed: " + err.Error())
		return shim.Error("Unmarshal sigHash result failed: " + err.Error())
	}
	if len(txJuryMsg) < consultM {
		log.Debugf("RecvJury sigHash result's len not enough")
		return shim.Error("RecvJury sigHash result's len not enough")
	}
	//协商 保证协商一致后才写入签名结果
	txResult2, err := consult(stub, []byte(sigHashHex+"twice"), []byte("sigHash2"))
	if err != nil {
		log.Debugf("consult sigHash2 failed: " + err.Error())
		return shim.Error("consult sigHash2 failed: " + err.Error())
	}
	var txJuryMsg2 []JuryMsgAddr
	err = json.Unmarshal(txResult2, &txJuryMsg2)
	if err != nil {
		log.Debugf("Unmarshal sigHash2 result failed: " + err.Error())
		return shim.Error("Unmarshal sigHash2 result failed: " + err.Error())
	}
	if len(txJuryMsg2) < consultM {
		log.Debugf("RecvJury sigHash2 result's len not enough")
		return shim.Error("RecvJury sigHash2 result's len not enough")
	}

	//记录签名
	var withdraw Withdraw
	withdraw.EthAddr = ethAddr
	withdraw.EthAmount = ethAmount
	withdraw.EthFee = 0
	withdraw.Sigs = append(withdraw.Sigs, sigs[0:consultM]...)
	withdrawBytes, err := json.Marshal(withdraw)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(symbolsWithdraw+reqid, withdrawBytes)
	if err != nil {
		log.Debugf("save withdraw failed: " + err.Error())
		return shim.Error("save withdraw failed: " + err.Error())
	}

	return shim.Success(withdrawBytes)
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
	err := shim.Start(new(ETHPort))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
