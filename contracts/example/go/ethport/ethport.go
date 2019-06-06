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
		return _initDepositAddr(args, stub)
	case "setETHTokenAsset":
		return _setETHTokenAsset(args, stub)
	case "getETHToken":
		return _getETHToken(args, stub)
	case "setETHContract":
		return _setETHContract(args, stub)
	case "withdrawPrepare":
		return _withdrawPrepare(args, stub)
	case "withdrawETH":
		return _withdrawETH(args, stub)

	case "put":
		return put(args, stub)
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

//Method:GetJuryETHAddr, return address string
type ETHAddress_GetJuryETHAddr struct {
	Method string `json:"method"`
}

const symbolsJuryAddress = "juryAddress"

const symbolsETHAsset = "eth_asset"
const symbolsETHContract = "eth_contract"

const symbolsDeposit = "deposit_"

const symbolsWithdrawPrepare = "withdrawPrepare_"
const symbolsWithdraw = "withdraw_"

const consultM = 3
const consultN = 4

// contractABI is same, but contractAddr is not
const contractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"reqid\",\"type\":\"string\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"reqid\",\"type\":\"string\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reqid\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"

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

func _initDepositAddr(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsJuryAddress)
	if len(saveResult) != 0 {
		return shim.Success([]byte("DepositAddr has been init"))
	}

	//
	getETHAddrParams := ETHAddress_GetJuryETHAddr{Method: "GetJuryETHAddr"}
	getETHAddrReqBytes, err := json.Marshal(getETHAddrParams)
	if err != nil {
		return shim.Error(err.Error())
	}
	result, err := stub.OutChainAddress("eth", getETHAddrReqBytes)
	if err != nil {
		log.Debugf("OutChainAddress GetJuryETHAddr err: %s", err.Error())
		return shim.Success([]byte("OutChainAddress GetJuryETHAddr failed"))
	}

	//
	recvResult, err := consult(stub, []byte("juryETHAddr"), []byte(result))
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Success([]byte("Unmarshal result failed: " + err.Error()))
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) != consultN {
		return shim.Success([]byte("RecvJury result's len not enough"))
	}

	//
	var address []string
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
	err := stub.PutState(symbolsETHContract, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsETHAsset failed: " + err.Error())
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
//add 'method' member.
type ETHQuery_GetBestHeader struct { //GetBestHeaderParams
	Method string `json:"method"`
	Number string `json:"Number"` //if empty, return the best header
}

type GetBestHeaderResult struct {
	Number string `json:"number"`
}

func getHight(stub shim.ChaincodeStubInterface) (string, string, error) {
	//
	getheader := ETHQuery_GetBestHeader{Method: "GetBestHeader"} //get best hight
	//
	reqBytes, err := json.Marshal(getheader)
	if err != nil {
		return "", "", err
	}
	//
	result, err := stub.OutChainQuery("eth", reqBytes)
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
	curHeight -= 6

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
	Txhashs   []string `json:"sigHashs"`
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
		return nil, err
	}
	//get doposit event log
	getevent := ETHTransaction_getevent{Method: "GetEventByAddress"} // GetJuryAddress
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
	result, err := stub.OutChainTransaction("eth", reqBytes)
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
	var depositInfo []DepositETHInfo
	for i, event := range geteventresult.Events {
		//Event example : ["0x0000000000000000000000000000000000000000","0x7d7116a8706ae08baa7f4909e26728fa7a5f0365",500000000000000000,"P1DXLJmJh9j3LFNUZ7MmfLVNWHoLzDUHM9A"]
		strArray := strings.Split(event, ",")
		if len(strArray) != 4 {
			continue
		}
		//confirm
		if geteventresult.Blocknums[i]+10 > endBlockNum {
			continue
		}
		//deposit amount, example : 500000000000000000
		str2 := strArray[2]
		bigInt := new(big.Int)
		bigInt.SetString(str2, 10)
		//
		depositInfo = append(depositInfo, DepositETHInfo{geteventresult.Txhashs[i], bigInt.Uint64()})
	}
	if len(depositInfo) == 0 {
		return nil, nil
	}

	return depositInfo, nil

}

func _getETHToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}

	contractAddr := getETHContract(stub)
	if contractAddr == "" {
		jsonResp := "{\"Error\":\"Failed to get contractAddr, need set contractAddr\"}"
		return shim.Error(jsonResp)
	}

	depositInfo, err := getDepositETHInfo(contractAddr, invokeAddr.String(), stub)
	if depositInfo == nil || err != nil {
		return shim.Success([]byte("You need deposit"))
	}
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
		ethAmount += uint64(depositInfo[i].Amount)
	}

	if ethAmount == 0 {
		return shim.Success([]byte("You need deposit"))
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
	if ethFee <= 100000 {
		ethFee = 100000
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
		return shim.Success([]byte(jsonResp))
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
		return shim.Success([]byte(jsonResp))
	}

	reqid := stub.GetTxID()
	// 产生交易
	rawTx := fmt.Sprintf("%s %d %s", ethAddr, ethTokenAmount, reqid)
	log.Debugf("rawTx:%s", rawTx)

	tempHash := crypto.Keccak256([]byte(rawTx), []byte("prepare"))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)

	//协商交易
	recvResult, err := consult(stub, []byte(tempHashHex), []byte("rawTx"))
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal rawTxSign result failed: " + err.Error())
		return shim.Success([]byte("Unmarshal rawTxSign result failed: " + err.Error()))
	}
	if len(juryMsg) < consultM {
		log.Debugf("RecvJury rawTxSign result's len not enough")
		return shim.Success([]byte("RecvJury rawTxSign result's len not enough"))
	}

	// 记录Prepare
	var prepare WithdrawPrepare
	prepare.EthAddr = ethAddr
	prepare.EthAmount = ethTokenAmount
	prepare.EthFee = ethFee
	prepareByte, err := json.Marshal(prepare)
	if err != nil {
		log.Debugf("Marshal prepare failed: " + err.Error())
		return shim.Success([]byte("Marshal prepare failed: " + err.Error()))
	}
	err = stub.PutState(symbolsWithdrawPrepare+stub.GetTxID(), prepareByte)
	if err != nil {
		log.Debugf("save symbolsWithdrawPrepare failed: " + err.Error())
		return shim.Success([]byte("save symbolsWithdrawPrepare failed: " + err.Error()))
	}

	return shim.Success([]byte("Withdraw is ready, please invoke withdrawETH"))
}

//refer to the struct Keccak256HashPackedSigParams in "github.com/palletone/adaptor/AdaptorETH.go",
//add 'method' member. Remove 'PrivateKeyHex', Jury will set itself when sign.
type ETHTransaction_calSig struct {
	Method     string `json:"method"`
	ParamTypes string `json:"paramtypes"`
	Params     string `json:"params"`
}
type Keccak256HashPackedSigResult struct {
	Signature string `json:"signature"`
}

func signTx(contractAddr, reqid, ethAddr string, ethAmount uint64, stub shim.ChaincodeStubInterface) (string, error) {
	//keccak256(abi.encodePacked(address(this), recver, amount, reqid));
	paramTypesArray := []string{"Address", "Address", "Uint", "String"} //eth
	paramTypesJson, err := json.Marshal(paramTypesArray)
	if err != nil {
		return "", err
	}
	//
	var paramsArray []string
	paramsArray = append(paramsArray, contractAddr)
	paramsArray = append(paramsArray, ethAddr)
	ethAmountStr := fmt.Sprintf("%d", ethAmount)
	paramsArray = append(paramsArray, ethAmountStr)
	paramsArray = append(paramsArray, reqid)

	paramsJson, err := json.Marshal(paramsArray)
	if err != nil {
		return "", err
	}
	ethTX := ETHTransaction_calSig{"Keccak256HashPackedSig", string(paramTypesJson), string(paramsJson)}
	reqBytes, err := json.Marshal(ethTX)
	if err != nil {
		return "", err
	}
	//
	result, err := stub.OutChainTransaction("eth", reqBytes)
	if err != nil {
		return "", errors.New("calSigETH error")
	}
	//
	var sigResult Keccak256HashPackedSigResult
	err = json.Unmarshal(result, &sigResult)
	if err != nil {
		return "", err
	}
	return sigResult.Signature, nil
}

func sortSigs(juryMsg []JuryMsgAddr) []string { //todo verify
	//
	var answers []string
	for i := range juryMsg {
		answers = append(answers, string(juryMsg[i].Answer))
	}
	a := sort.StringSlice(answers[0:])
	sort.Sort(a)

	return answers
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
		return shim.Success([]byte(jsonResp))
	}

	contractAddr := getETHContract(stub)
	if contractAddr == "" {
		jsonResp := "{\"Error\":\"Failed to get contractAddr, need set contractAddr\"}"
		return shim.Error(jsonResp)
	}

	// 计算签名
	rawTxSign, err := signTx(contractAddr, reqid, prepare.EthAddr, prepare.EthAmount, stub)
	if err != nil {
		return shim.Success([]byte("signTx failed: " + err.Error()))
	}
	log.Debugf("rawTxSign:%s", rawTxSign)

	//
	reqidNew := stub.GetTxID()
	rawTx := fmt.Sprintf("%s %d %s", prepare.EthAddr, prepare.EthAmount, reqidNew)
	tempHash := crypto.Keccak256([]byte(rawTx))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)

	//协商交易
	recvResult, err := consult(stub, []byte(tempHashHex), []byte(rawTxSign))
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal rawTxSign result failed: " + err.Error())
		return shim.Success([]byte("Unmarshal rawTxSign result failed: " + err.Error()))
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) < consultM {
		log.Debugf("RecvJury rawTxSign result's len not enough")
		return shim.Success([]byte("RecvJury rawTxSign result's len not enough"))
	}

	// 合并交易
	sigs := sortSigs(juryMsg)
	sigHash := crypto.Keccak256([]byte(sigs[0] + sigs[1] + sigs[2]))
	sigHashHex := fmt.Sprintf("%x", sigHash)

	log.Debugf("start consult sigHashHex %s", sigHashHex)
	//协商 发送交易哈希
	txResult, err := consult(stub, []byte(sigHashHex), []byte("sigHash"))
	if err != nil {
		log.Debugf("consult sigHash failed: " + err.Error())
		return shim.Success([]byte("consult sigHash failed: " + err.Error()))
	}
	var txJuryMsg []JuryMsgAddr
	err = json.Unmarshal(txResult, &txJuryMsg)
	if err != nil {
		log.Debugf("Unmarshal sigHash result failed: " + err.Error())
		return shim.Success([]byte("Unmarshal sigHash result failed: " + err.Error()))
	}
	if len(txJuryMsg) < consultM {
		log.Debugf("RecvJury sigHash result's len not enough")
		return shim.Success([]byte("RecvJury sigHash result's len not enough"))
	}
	//协商 保证协商一致后才写入签名结果
	txResult2, err := consult(stub, []byte(sigHashHex+"twice"), []byte("sigHash2"))
	if err != nil {
		log.Debugf("consult sigHash2 failed: " + err.Error())
		return shim.Success([]byte("consult sigHash2 failed: " + err.Error()))
	}
	var txJuryMsg2 []JuryMsgAddr
	err = json.Unmarshal(txResult2, &txJuryMsg2)
	if err != nil {
		log.Debugf("Unmarshal sigHash2 result failed: " + err.Error())
		return shim.Success([]byte("Unmarshal sigHash2 result failed: " + err.Error()))
	}
	if len(txJuryMsg2) < consultM {
		log.Debugf("RecvJury sigHash2 result's len not enough")
		return shim.Success([]byte("RecvJury sigHash2 result's len not enough"))
	}

	//记录签名
	var withdraw Withdraw
	withdraw.EthAddr = prepare.EthAddr
	withdraw.EthAmount = prepare.EthAmount
	withdraw.EthFee = prepare.EthFee
	withdraw.Sigs = append(withdraw.Sigs, sigs[0:3]...)
	withdrawBytes, err := json.Marshal(withdraw)
	err = stub.PutState(symbolsWithdraw+stub.GetTxID(), withdrawBytes)
	if err != nil {
		log.Debugf("save withdraw failed: " + err.Error())
		return shim.Success([]byte("save withdraw failed: " + err.Error()))
	}

	//删除Prepare
	err = stub.DelState(symbolsWithdrawPrepare + reqid)
	if err != nil {
		log.Debugf("delete WithdrawPrepare failed: " + err.Error())
		return shim.Success([]byte("delete WithdrawPrepare failed: " + err.Error()))
	}

	return shim.Success([]byte(withdrawBytes))
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
	err := shim.Start(new(ETHPort))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
