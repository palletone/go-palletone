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
	"bytes"
	"encoding/hex"
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
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"

	"github.com/palletone/adaptor"
)

type PTNETH struct {
}

func (p *PTNETH) Init(stub shim.ChaincodeStubInterface) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(symbolsOwner, []byte(invokeAddr.String()))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}

	return shim.Success(nil)
}

func (p *PTNETH) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "initDepositAddr":
		return p.InitDepositAddr(stub)

	case "createGoToETHSigs":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNTransferTxID,[ethAddrInput])")
		}
		ethAddrInput := ""
		if len(args) > 1 {
			ethAddrInput = args[1]
		}
		return p.CreateGoToETHSigs(args[0], ethAddrInput, stub)

	case "getGoToETHSigs":
		if len(args) < 1 {
			return shim.Error("need 1 args (reqID (from withdrawPrepare))")
		}
		return p.GetGoToETHSigs(stub, args[0])

	case "payPTNByETHTxID":
		if len(args) < 1 {
			return shim.Error("need 1 args (ETHTxID)")
		}
		return p.PayPTNByETHTxID(args[0], stub)

	case "setGoToETHContract":
		if len(args) < 1 {
			return shim.Error("need 1 args (ETHContractAddr)")
		}
		return p.SetGoToETHContract(args[0], stub)
	case "setOwner":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNAddr)")
		}
		return p.SetOwner(args[0], stub)

	case "withdrawSubmit":
		if len(args) < 1 {
			return shim.Error("need 1 args (ETHTxID)")
		}
		return p.WithdrawSubmit(args[0], stub)
	case "withdrawFee": //todo 改为从本合约取款PTN
		if len(args) < 1 {
			return shim.Error("need 1 args (ethAddr)")
		}
		return p.WithdrawFee(args[0], stub)

	case "get":
		if len(args) < 1 {
			return shim.Error("need 1 args (Key)")
		}
		return p.Get(stub, args[0])

	case "withdrawAmount":
		if len(args) < 2 {
			return shim.Error("need 2  args (PTNAddress,PTNAmount)")
		}
		withdrawAddr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		amount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid amount:" + args[1])
		}
		return p.WithdrawAmount(stub, withdrawAddr, amount)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

type JuryMsgAddr struct {
	Address string
	Answer  []byte
}

//todo modify conforms 15
const Confirms = uint(1)

const symbolsJuryAddress = "juryAddress"
const symbolsJuryPubkeyAddress = "juryPubkeyAddress"

const symbolsGoToETHContract = "go_to_eth_contract"

const symbolsDeposit = "deposit_"
const symbolsSubmit = "submit_"

const symbolsWithdrawFee = "withdrawfee_"
const symbolsOwner = "owner_"

const symbolsGoToETHSigs = "goToETHSigs_"

const consultM = 3
const consultN = 4

const jsonResp1 = "{\"Error\":\"Failed to get contractAddr, need set contractAddr\"}"

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

type pubkeyAddr struct {
	Addr   string
	Pubkey []byte
}
type pubkeyAddrWrapper struct {
	pubAddr []pubkeyAddr
	by      func(p, q *pubkeyAddr) bool
}
type SortBy func(p, q *pubkeyAddr) bool

func (pw pubkeyAddrWrapper) Len() int { // 重写 Len() 方法
	return len(pw.pubAddr)
}
func (pw pubkeyAddrWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.pubAddr[i], pw.pubAddr[j] = pw.pubAddr[j], pw.pubAddr[i]
}
func (pw pubkeyAddrWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.by(&pw.pubAddr[i], &pw.pubAddr[j])
}

func sortPubAddr(thePubAddr []pubkeyAddr, by SortBy) { // sortPubAddr 方法
	sort.Sort(pubkeyAddrWrapper{thePubAddr, by})
}

func addrIncrease(p, q *pubkeyAddr) bool {
	return p.Addr < q.Addr // addr increase sort
}

func (p *PTNETH) InitDepositAddr(stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsJuryPubkeyAddress)
	if len(saveResult) != 0 {
		return shim.Error("DepositAddr has been init")
	}

	//Method:GetJuryETHAddr, return address string
	juryAddr, err := stub.OutChainCall("eth", "GetJuryAddr", []byte(""))
	if err != nil {
		log.Debugf("OutChainCall GetJuryETHAddr err: %s", err.Error())
		return shim.Error("OutChainCall GetJuryETHAddr failed " + err.Error())
	}
	log.Debugf("juryAddr: %s", string(juryAddr))

	result, err := stub.OutChainCall("eth", "GetJuryPubkey", []byte(""))
	if err != nil {
		log.Debugf("OutChainCall GetJuryPubkey err: %s", err.Error())
		return shim.Error("OutChainCall GetJuryPubkey failed " + err.Error())
	}
	var juryPubkey adaptor.GetPublicKeyOutput
	err = json.Unmarshal(result, &juryPubkey)
	if err != nil {
		log.Debugf("OutChainCall GetJuryPubkey Unmarshal err: %s", err.Error())
		return shim.Error("OutChainCall GetJuryPubkey Unmarshal failed " + err.Error())
	}
	log.Debugf("juryPubkey.PublicKey: %x", juryPubkey.PublicKey)

	myPubkeyAddr := pubkeyAddr{string(juryAddr), juryPubkey.PublicKey}
	myPubkeyAddrJSON, err := json.Marshal(myPubkeyAddr)
	if err != nil {
		log.Debugf("myPubkeyAddr Marshal failed: %s", err.Error())
		return shim.Error("myPubkeyAddr Marshal failed: " + err.Error())
	}
	log.Debugf("myPubkeyAddrJSON: %s", string(myPubkeyAddrJSON))

	//
	recvResult, err := consult(stub, []byte("juryETHPubkey"), myPubkeyAddrJSON)
	if err != nil {
		log.Debugf("consult juryETHPubkey failed: %s", err.Error())
		return shim.Error("consult juryETHPubkey failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Error("Unmarshal result failed: " + err.Error())
	}
	if len(juryMsg) != consultN {
		return shim.Error("RecvJury result's len not enough")
	}

	//
	pubkeyAddrs := make([]pubkeyAddr, 0, len(juryMsg))
	for i := range juryMsg {
		var onePubkeyAddr pubkeyAddr
		err := json.Unmarshal(juryMsg[i].Answer, &onePubkeyAddr)
		if err != nil {
			continue
		}
		pubkeyAddrs = append(pubkeyAddrs, onePubkeyAddr)
	}
	if len(pubkeyAddrs) != consultN {
		return shim.Error("pubkeyAddrs result's len not enough")
	}
	sortPubAddr(pubkeyAddrs, addrIncrease)

	address := make([]string, 0, len(pubkeyAddrs))
	for i := range pubkeyAddrs {
		address = append(address, pubkeyAddrs[i].Addr)
	}
	addressJSON, err := json.Marshal(address)
	if err != nil {
		return shim.Error("address Marshal failed: " + err.Error())
	}
	log.Debugf("addressJSON: %s", string(addressJSON))
	pubkeyAddrsJSON, err := json.Marshal(pubkeyAddrs)
	if err != nil {
		return shim.Error("pubkeyAddrs Marshal failed: " + err.Error())
	}
	log.Debugf("pubkeyAddrsJson: %s", string(pubkeyAddrsJSON))

	// Write the state to the ledger
	err = stub.PutState(symbolsJuryAddress, addressJSON)
	if err != nil {
		return shim.Error("write " + symbolsJuryAddress + " failed: " + err.Error())
	}
	err = stub.PutState(symbolsJuryPubkeyAddress, pubkeyAddrsJSON)
	if err != nil {
		return shim.Error("write " + symbolsJuryPubkeyAddress + " failed: " + err.Error())
	}

	return shim.Success(addressJSON)
}

func getETHAddrs(stub shim.ChaincodeStubInterface) []pubkeyAddr {
	result, _ := stub.GetState(symbolsJuryPubkeyAddress)
	if len(result) == 0 {
		return []pubkeyAddr{}
	}
	var pubkeyAddrs []pubkeyAddr
	err := json.Unmarshal(result, &pubkeyAddrs)
	if err != nil {
		return []pubkeyAddr{}
	}
	return pubkeyAddrs
}

func (p *PTNETH) SetGoToETHContract(ethContractAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsGoToETHContract)
	if len(saveResult) != 0 {
		return shim.Error("TokenAsset has been init")
	}

	err := stub.PutState(symbolsGoToETHContract, []byte(strings.ToLower(ethContractAddr)))
	if err != nil {
		return shim.Error("write symbolsGoToETHContract failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func (p *PTNETH) SetOwner(ptnAddr string, stub shim.ChaincodeStubInterface) pb.Response {
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
	err = stub.PutState(symbolsOwner, []byte(ptnAddr))
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

func getETHContract(stub shim.ChaincodeStubInterface) string {
	result, _ := stub.GetState(symbolsGoToETHContract)
	if len(result) == 0 {
		return ""
	}
	log.Debugf("contractAddr: %s", string(result))

	return string(result)
}

func getHeight(stub shim.ChaincodeStubInterface) (uint, error) {
	//
	input := adaptor.GetBlockInfoInput{Latest: true} //get best hight
	//
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return 0, err
	}
	//adaptor.
	result, err := stub.OutChainCall("eth", "GetBlockInfo", inputBytes)
	if err != nil {
		return 0, errors.New("GetBlockInfo error: " + err.Error())
	}
	//
	var output adaptor.GetBlockInfoOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return 0, err
	}

	if output.Block.BlockHeight == 0 {
		return 0, errors.New("{\"Error\":\"Failed to get eth height\"}")
	}

	return output.Block.BlockHeight, nil
}

func getPTNMapAddr(mapAddr, fromAddr string, stub shim.ChaincodeStubInterface) (string, error) {
	var input adaptor.GetPalletOneMappingAddressInput
	input.MappingDataSource = mapAddr
	input.ChainAddress = fromAddr

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	//
	result, err := stub.OutChainCall("eth", "GetPalletOneMappingAddress", inputBytes)
	if err != nil {
		return "", errors.New("GetPalletOneMappingAddress failed: " + err.Error())
	}
	//
	var output adaptor.GetPalletOneMappingAddressOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return "", err
	}
	if output.PalletOneAddress == "" {
		return "", errors.New("GetPalletOneMappingAddress result empty")
	}

	return output.PalletOneAddress, nil
}

func GetERC20Tx(txID []byte, stub shim.ChaincodeStubInterface) (*adaptor.GetTransferTxOutput, error) {
	input := adaptor.GetTransferTxInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	//todo 改成获取ERC20交易详情
	result, err := stub.OutChainCall("erc20", "GetTransferTx", inputBytes)
	if err != nil {
		return nil, errors.New("GetTransferTx error: " + err.Error())
	}
	log.Debugf("result : %s", string(result))

	//
	var output adaptor.GetTransferTxOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

const PTN_ERC20Addr = "0xfe76be9cec465ed3219a9972c21655d57d21aec6"

func (p *PTNETH) PayPTNByETHTxID(ethTxID string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	if "0x" == ethTxID[0:2] || "0X" == ethTxID[0:2] {
		ethTxID = ethTxID[2:]
	}
	result, _ := stub.GetState(symbolsDeposit + ethTxID)
	if len(result) != 0 {
		log.Debugf("The tx %s has been payout", ethTxID)
		return shim.Error("The tx has been payout")
	}

	//get sender receiver amount
	txIDByte, err := hex.DecodeString(ethTxID)
	if err != nil {
		log.Debugf("txid invalid: %s", err.Error())
		return shim.Error(fmt.Sprintf("txid invalid: %s", err.Error()))
	}

	mapAddr := getETHContract(stub)
	if mapAddr == "" {
		return shim.Error(jsonResp1)
	}
	txResult, err := GetERC20Tx(txIDByte, stub)
	if err != nil {
		log.Debugf("GetERC20Tx failed: %s", err.Error())
		return shim.Error(err.Error())
	}
	//check tx status
	if !txResult.Tx.IsSuccess {
		log.Debugf("The tx %s is failed", ethTxID)
		return shim.Error("The tx is failed")
	}

	//check contract address, must be ptn erc20 contract address
	if strings.ToLower(txResult.Tx.TargetAddress) != PTN_ERC20Addr {
		log.Debugf("The tx is't ERC20 contract transfer of PTN")
		return shim.Error("The tx is't ERC20 contract transfer of PTN")
	}
	//check receiver, must be ptnmap contract address
	if strings.ToLower(txResult.Tx.ToAddress) != mapAddr {
		log.Debugf("strings.ToLower(txResult.To): %s, mapAddr: %s ", strings.ToLower(txResult.Tx.ToAddress), mapAddr)
		return shim.Error("Not send token to the Map contract")
	}
	//check confirms
	curHeight, err := getHeight(stub)
	if curHeight == 0 || err != nil {
		return shim.Error("getHeight failed")
	}
	if curHeight-txResult.Tx.BlockHeight < Confirms {
		log.Debugf("The tx %s need more confirms", ethTxID)
		return shim.Error("Need more confirms")
	}

	//get the mapping ptnAddr
	ptnAddr, err := getPTNMapAddr(mapAddr, txResult.Tx.FromAddress, stub)
	if err != nil {
		log.Debugf("getPTNMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	bigIntAmount := txResult.Tx.Amount.Amount
	bigIntAmount = bigIntAmount.Div(bigIntAmount, big.NewInt(1e10)) //ethToken in PTN is decimal is 8
	ptnAmount := bigIntAmount.Uint64()
	if ptnAmount == 0 {
		return shim.Error("You need deposit bigger than 0")
	}

	//
	err = stub.PutState(symbolsDeposit+ethTxID, []byte(ptnAddr+"-"+bigIntAmount.String()))
	if err != nil {
		log.Debugf("PutState sigHash failed err: %s", err.Error())
		return shim.Error("PutState sigHash failed")
	}

	//
	PTNAsset := modules.NewPTNAsset() //todo 改成PTN Asset
	invokeTokens := new(modules.AmountAsset)
	invokeTokens.Amount = ptnAmount
	invokeTokens.Asset = PTNAsset
	err = stub.PayOutToken(ptnAddr, invokeTokens, 0)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.PayOutToken\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success([]byte("get success"))
}

type WithdrawPrepare struct {
	EthAddr   string
	PTNAmount uint64
	PTNFee    uint64
}

func updateFeeAdd(fee uint64, ptnAddr string, stub shim.ChaincodeStubInterface) error {
	feeCur := uint64(0)
	result, _ := stub.GetState(symbolsWithdrawFee + ptnAddr)
	if len(result) != 0 {
		log.Debugf("updateFee fee current : %s ", string(result))
		feeCur, _ = strconv.ParseUint(string(result), 10, 64)
	}
	feeCur += fee
	feeStr := fmt.Sprintf("%d", feeCur)
	err := stub.PutState(symbolsWithdrawFee+ptnAddr, []byte(feeStr))
	if err != nil {
		log.Debugf("updateFee failed: %s", err.Error())
		return fmt.Errorf("updateFee failed: " + err.Error())
	}
	return nil
}

//fee == 0, clear
func updateFeeSub(fee uint64, ptnAddr string, stub shim.ChaincodeStubInterface) error {
	feeCur := uint64(0)
	if fee != 0 {
		result, _ := stub.GetState(symbolsWithdrawFee + ptnAddr)
		if len(result) != 0 {
			log.Debugf("updateFee fee current : %s ", string(result))
			feeCur, _ = strconv.ParseUint(string(result), 10, 64)
		}
		if feeCur < fee {
			return fmt.Errorf("current fee is small")
		}
		feeCur -= fee
	}

	feeStr := fmt.Sprintf("%d", feeCur)
	err := stub.PutState(symbolsWithdrawFee+ptnAddr, []byte(feeStr))
	if err != nil {
		log.Debugf("updateFee failed: %s", err.Error())
		return fmt.Errorf("updateFee failed: " + err.Error())
	}
	return nil
}

func getFee(ptnAddr string, stub shim.ChaincodeStubInterface) uint64 {
	feeCur := uint64(0)
	result, _ := stub.GetState(symbolsWithdrawFee + ptnAddr)
	if len(result) != 0 {
		log.Debugf("getFee fee current : %s ", string(result))
		feeCur, _ = strconv.ParseUint(string(result), 10, 64)
	}
	return feeCur
}

func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20
)

// Address represents the 20 byte address of an Ethereum account.
type ETHAddress [AddressLength]byte

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) ETHAddress {
	var a ETHAddress
	a.SetBytes(b)
	return a
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *ETHAddress) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) ETHAddress { return BytesToAddress(FromHex(s)) }

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// Bytes gets the string representation of the underlying address.
func (a ETHAddress) Bytes() []byte { return a[:] }

func getPadderBytes(contractAddr, reqid, recvAddr string, ethAmount uint64) []byte {
	var allBytes []byte
	ethContractAddr := HexToAddress(contractAddr)
	allBytes = append(allBytes, ethContractAddr.Bytes()...)

	ethRecvAddr := HexToAddress(recvAddr)
	allBytes = append(allBytes, ethRecvAddr.Bytes()...)

	paramBigInt := new(big.Int)
	paramBigInt.SetUint64(ethAmount)
	paramBigInt.Mul(paramBigInt, big.NewInt(1e10)) //eth's decimal is 18, ethToken in PTN is decimal is 8
	paramBigIntBytes := LeftPadBytes(paramBigInt.Bytes(), 32)
	allBytes = append(allBytes, paramBigIntBytes...)

	reqHash := common.HexToHash(reqid)
	allBytes = append(allBytes, reqHash.Bytes()...)
	return allBytes
}
func calSig(msg []byte, stub shim.ChaincodeStubInterface) ([]byte, error) {
	//

	input := adaptor.SignMessageInput{Message: msg}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return []byte{}, err
	}

	//
	result, err := stub.OutChainCall("eth", "SignMessage", inputBytes)
	if err != nil {
		return []byte{}, errors.New("SignMessage error" + err.Error())
	}
	//
	var sigResult adaptor.SignMessageOutput
	err = json.Unmarshal(result, &sigResult)
	if err != nil {
		return []byte{}, err
	}
	return sigResult.Signature, nil
}

func recoverAddr(msg, pubkey, sig []byte, stub shim.ChaincodeStubInterface) (bool, error) {
	log.Debugf("recover %x-%x-%x", msg, pubkey, sig)

	ethTX := adaptor.VerifySignatureInput{Message: msg, Signature: sig, PublicKey: pubkey}
	reqBytes, err := json.Marshal(ethTX)
	if err != nil {
		return false, err
	}
	//
	result, err := stub.OutChainCall("eth", "VerifySignature", reqBytes)
	if err != nil {
		return false, errors.New("RecoverAddr error" + err.Error())
	}
	//
	var recoverResult adaptor.VerifySignatureOutput
	err = json.Unmarshal(result, &recoverResult)
	if err != nil {
		return false, err
	}
	return recoverResult.Pass, nil
}

func verifySigs(msg []byte, juryMsg []JuryMsgAddr, pubkeyAddrs []pubkeyAddr, stub shim.ChaincodeStubInterface) []string {
	//
	var sigs []string
	for i := range juryMsg {
		var onePubkeySig pubkeySig
		err := json.Unmarshal(juryMsg[i].Answer, &onePubkeySig)
		if err != nil {
			continue
		}
		log.Debugf("verifySigs %x-%x", onePubkeySig.Pubkey, onePubkeySig.Sig)
		isJuryETHPubkey := false
		for j := range pubkeyAddrs {
			if bytes.Equal(pubkeyAddrs[j].Pubkey, onePubkeySig.Pubkey) {
				isJuryETHPubkey = true
			}
		}
		if !isJuryETHPubkey {
			continue
		}
		valid, err := recoverAddr(msg, onePubkeySig.Pubkey, onePubkeySig.Sig, stub)
		if err != nil {
			continue
		}
		if valid {
			sigs = append(sigs, fmt.Sprintf("%x", onePubkeySig.Sig))
		}
	}
	log.Debugf("sigs : %s", sigs)

	//sort
	a := sort.StringSlice(sigs[0:])
	sort.Sort(a)
	log.Debugf("sigs sort : %s", sigs)
	return sigs
}

type Withdraw struct {
	EthAddr   string
	PTNAmount uint64
	PTNFee    uint64
	Sigs      []string
}

type pubkeySig struct {
	Pubkey []byte
	Sig    []byte
}

func processWithdrawSig(txID, reqidNew, recvAddr string, PTNAmount uint64, stub shim.ChaincodeStubInterface) ([]string, error) {
	contractAddr := getETHContract(stub)
	if contractAddr == "" {
		return []string{}, fmt.Errorf(jsonResp1)
	}

	// 计算签名
	padderBytes := getPadderBytes(contractAddr, txID, recvAddr, PTNAmount)
	sig, err := calSig(padderBytes, stub)
	if err != nil {
		return []string{}, fmt.Errorf("calSig failed: " + err.Error())
	}
	log.Debugf("sig: %s", sig)

	//获取自己的eth公钥
	resultPubkey, err := stub.OutChainCall("eth", "GetJuryPubkey", []byte(""))
	if err != nil {
		log.Debugf("OutChainCall GetJuryPubkey err: %s", err.Error())
		return []string{}, fmt.Errorf("OutChainCall GetJuryPubkey failed " + err.Error())
	}
	var juryPubkey adaptor.GetPublicKeyOutput
	err = json.Unmarshal(resultPubkey, &juryPubkey)
	if err != nil {
		log.Debugf("OutChainCall GetJuryPubkey Unmarshal err: %s", err.Error())
		return []string{}, fmt.Errorf("OutChainCall GetJuryPubkey Unmarshal failed " + err.Error())
	}
	//计算交易哈希
	rawTx := fmt.Sprintf("%s %d %s", recvAddr, PTNAmount, reqidNew)
	tempHash := crypto.Keccak256([]byte(rawTx))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)
	//用交易哈希协商交易签名，作适当安全防护
	myPubkeySig := pubkeySig{Pubkey: juryPubkey.PublicKey, Sig: sig}
	myPubkeySigBytes, _ := json.Marshal(myPubkeySig)
	recvResult, err := consult(stub, []byte(tempHashHex), myPubkeySigBytes)
	if err != nil {
		log.Debugf("consult sig failed: %s", err.Error())
		return []string{}, fmt.Errorf("consult sig failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal sig result failed: %s", err.Error())
		return []string{}, fmt.Errorf("Unmarshal sig result failed: " + err.Error())
	}
	if len(juryMsg) < consultM {
		return []string{}, fmt.Errorf("RecvJury sig result's len not enough")
	}

	//验证收集到的所有eth签名
	pubkeyAddrs := getETHAddrs(stub)
	if len(pubkeyAddrs) != consultN {
		return []string{}, fmt.Errorf("getETHAddrs result's len not enough")
	}
	sigs := verifySigs(padderBytes, juryMsg, pubkeyAddrs, stub)
	if len(sigs) < consultM {
		return []string{}, fmt.Errorf("verifySigs result's len not enough")
	}
	sigsStr := sigs[0]
	for i := 1; i < consultM; i++ {
		sigsStr = sigsStr + sigs[i]
	}
	sigHash := crypto.Keccak256([]byte(sigsStr))
	sigHashHex := fmt.Sprintf("%x", sigHash)
	log.Debugf("start consult sigHashHex %s", sigHashHex)

	//用签名列表的哈希协商最终的3个交易签名，作适当安全防护
	txResult, err := consult(stub, []byte(sigHashHex), []byte("sigHash"))
	if err != nil {
		log.Debugf("consult sigHash failed: %s", err.Error())
		return []string{}, fmt.Errorf("consult sigHash failed: " + err.Error())
	}
	var txJuryMsg []JuryMsgAddr
	err = json.Unmarshal(txResult, &txJuryMsg)
	if err != nil {
		log.Debugf("Unmarshal sigHash result failed: %s", err.Error())
		return []string{}, fmt.Errorf("Unmarshal sigHash result failed: " + err.Error())
	}
	if len(txJuryMsg) < consultM {
		return []string{}, fmt.Errorf("RecvJury sigHash result's len not enough")
	}
	////协商两次 保证协商一致后才写入签名结果
	//txResult2, err := consult(stub, []byte(sigHashHex+"twice"), []byte("sigHash2"))
	//if err != nil {
	//	log.Debugf("consult sigHash2 failed: " + err.Error())
	//	return []string{}, fmt.Errorf("consult sigHash2 failed: " + err.Error())
	//}
	//var txJuryMsg2 []JuryMsgAddr
	//err = json.Unmarshal(txResult2, &txJuryMsg2)
	//if err != nil {
	//	log.Debugf("Unmarshal sigHash2 result failed: " + err.Error())
	//	return []string{}, fmt.Errorf("Unmarshal sigHash2 result failed: " + err.Error())
	//}
	//if len(txJuryMsg2) < consultM {
	//	log.Debugf("RecvJury sigHash2 result's len not enough")
	//	return []string{}, fmt.Errorf("RecvJury sigHash2 result's len not enough")
	//}
	return sigs, nil
}

func getWithdrawPrepare(txID, ethAddrInput string, stub shim.ChaincodeStubInterface) (*WithdrawPrepare, error) {
	PTNAsset := modules.NewPTNAsset() //todo 改成PTN Asset

	// 1 get this tx
	tx, err := stub.GetStableTransactionByHash(txID)
	if nil != err {
		return nil, fmt.Errorf("GetStableTransactionByHash failed " + err.Error())
	}

	//2 sender address, get it from txPre output
	txMsgs := tx.TxMessages()
	payment := txMsgs[0].Payload.(*modules.PaymentPayload)
	outPoint := payment.Inputs[0].PreviousOutPoint
	txPre, err := stub.GetStableTransactionByHash(outPoint.TxHash.String())
	if nil != err {
		return nil, fmt.Errorf("GetStableTransactionByHash txPre failed " + err.Error())
	}
	txPreMsgs := txPre.TxMessages()
	paymentPre := txPreMsgs[outPoint.MessageIndex].Payload.(*modules.PaymentPayload)
	outputPre := paymentPre.Outputs[outPoint.OutIndex]
	toAddr, _ := tokenengine.Instance.GetAddressFromScript(outputPre.PkScript)

	//3 amount, get it from tx output
	_, contractAddr := stub.GetContractID()
	amount := uint64(0)
	newOutpointMapUTXOs := tx.GetNewUtxos()
	for _, utxo := range newOutpointMapUTXOs {
		recvAddr, _ := tokenengine.Instance.GetAddressFromScript(utxo.PkScript)
		if recvAddr.String() != contractAddr {
			continue
		}
		if !utxo.Asset.IsSameAssetId(PTNAsset) {
			continue
		}
		amount += utxo.Amount
	}

	//4 op_return
	ethAddr := ""
	for _, msg := range txMsgs {
		if msg.App == modules.APP_DATA {
			text := msg.Payload.(*modules.DataPayload)
			ethAddr = string(text.MainData)
			break
		}
	}
	if ethAddr == "" {
		if ethAddrInput == "" {
			return nil, fmt.Errorf("Get ethAddr failed")
		}
		ethAddr = ethAddrInput
	}

	//get all result
	var prepare WithdrawPrepare
	prepare.EthAddr = ethAddr
	prepare.PTNAmount = amount
	prepare.PTNFee = 500000000 // 5 ptn

	log.Debugf("%s-%d-%s", toAddr.String(), amount, ethAddr)
	return &prepare, nil
}

func (p *PTNETH) CreateGoToETHSigs(txID, ethAddrInput string, stub shim.ChaincodeStubInterface) pb.Response {
	if "0x" == txID[0:2] || "0X" == txID[0:2] {
		txID = txID[2:]
	}

	resultWithdraw, _ := stub.GetState(symbolsGoToETHSigs + txID)
	if len(resultWithdraw) != 0 {
		return shim.Error("The txID has been withdraw")
	}

	prepare, err := getWithdrawPrepare(txID, ethAddrInput, stub)
	if nil != err {
		return shim.Error(err.Error())
	}
	if prepare.PTNAmount <= prepare.PTNFee { //todo 收取PTN手续费作为调用ETH合约的补偿
		var withdraw Withdraw
		withdraw.EthAddr = prepare.EthAddr
		withdraw.PTNAmount = prepare.PTNAmount
		withdraw.PTNFee = prepare.PTNFee
		withdrawBytes, err := json.Marshal(withdraw)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(symbolsGoToETHSigs+txID, withdrawBytes)
		if err != nil {
			log.Debugf("save withdraw failed: %s", err.Error())
			return shim.Error("save withdraw failed: " + err.Error())
		}
		return shim.Success(withdrawBytes)
	}

	reqidNew := stub.GetTxID()
	sigs, err := processWithdrawSig(txID, reqidNew, prepare.EthAddr, prepare.PTNAmount-prepare.PTNFee, stub)
	if nil != err {
		jsonResp := "processWithdrawSig failed " + err.Error()
		return shim.Error(jsonResp)
	}

	//记录签名
	var withdraw Withdraw
	withdraw.EthAddr = prepare.EthAddr
	withdraw.PTNAmount = prepare.PTNAmount
	withdraw.PTNFee = prepare.PTNFee
	withdraw.Sigs = append(withdraw.Sigs, sigs[0:consultM]...)
	withdrawBytes, err := json.Marshal(withdraw)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(symbolsGoToETHSigs+txID, withdrawBytes)
	if err != nil {
		log.Debugf("save withdraw failed: %s", err.Error())
		return shim.Error("save withdraw failed: " + err.Error())
	}

	return shim.Success(withdrawBytes)
}

func GetETHContractTx(txID []byte, stub shim.ChaincodeStubInterface) (*adaptor.GetTxBasicInfoOutput, error) {
	input := adaptor.GetTxBasicInfoInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	//todo fixme 目的是GoToETH合约调用判断
	result, err := stub.OutChainCall("eth", "GetTxBasicInfo", inputBytes)
	if err != nil {
		return nil, errors.New("GetTransferTx error: " + err.Error())
	}
	log.Debugf("result : %s", string(result))

	//
	var output adaptor.GetTxBasicInfoOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (p *PTNETH) WithdrawSubmit(ethTxID string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	if "0x" == ethTxID[0:2] || "0X" == ethTxID[0:2] {
		ethTxID = ethTxID[2:]
	}
	result, _ := stub.GetState(symbolsSubmit + ethTxID)
	if len(result) != 0 {
		log.Debugf("The tx %s fee has been payout", ethTxID)
		return shim.Error("The fee has been payout")
	}

	//get sender receiver amount
	txIDByte, err := hex.DecodeString(ethTxID)
	if err != nil {
		log.Debugf("txid invalid: %s", err.Error())
		return shim.Error(fmt.Sprintf("txid invalid: %s", err.Error()))
	}

	mapAddr := getETHContract(stub)
	if mapAddr == "" {
		return shim.Error(jsonResp1)
	}
	txResult, err := GetETHContractTx(txIDByte, stub)
	if err != nil {
		log.Debugf("GetETHContractTx failed: %s", err.Error())
		return shim.Error(err.Error())
	}
	//check tx status
	if !txResult.Tx.IsSuccess {
		log.Debugf("The tx %s is failed", ethTxID)
		return shim.Error("The tx is failed")
	}
	//check contract address, must be ptn eth port contract address
	if strings.ToLower(txResult.Tx.TargetAddress) != mapAddr {
		log.Debugf("The tx %s is't transfer to eth port contract", ethTxID)
		return shim.Error("The tx is't transfer to eth port contract")
	}
	withdrawMethodId := Hex2Bytes("73432d0a") //todo withdraw 非标准ERC20调用
	if !bytes.HasPrefix(txResult.Tx.TxRawData, withdrawMethodId) {
		log.Debugf("The tx %s is't call withdraw", ethTxID)
		return shim.Error("The tx is't call withdraw")
	}

	//get the mapping ptnAddr
	ptnAddr, err := getPTNMapAddr(mapAddr, txResult.Tx.CreatorAddress, stub)
	if err != nil {
		log.Debugf("getPTNMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	reqid := hex.EncodeToString(txResult.Tx.TxRawData[68:100]) //4method+32recvAddr+32amount+32reqid
	resultWithdraw, _ := stub.GetState(symbolsGoToETHSigs + reqid)
	if len(resultWithdraw) == 0 {
		return shim.Error("Not exist withdraw of reqid : " + reqid)
	}
	// 检查交易
	var withdraw Withdraw
	err = json.Unmarshal(resultWithdraw, &withdraw)
	if nil != err {
		jsonResp := "Unmarshal withdraw failed"
		return shim.Error(jsonResp)
	}
	//
	err = stub.PutState(symbolsSubmit+ethTxID, []byte(ptnAddr+"-"+fmt.Sprintf("%d", withdraw.PTNFee)))
	if err != nil {
		log.Debugf("PutState symbolsSubmit failed err: %s", err.Error())
		return shim.Error("PutState symbolsSubmit failed " + err.Error())
	}

	//
	err = updateFeeAdd(withdraw.PTNFee, ptnAddr, stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("Success"))
}

func (p *PTNETH) WithdrawFee(ptnRecvAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	ptnAddr := invokeAddr.String()

	//
	PTNAmount := getFee(ptnAddr, stub)
	if PTNAmount == 0 {
		jsonResp := "{\"Error\":\"fee is 0\"}"
		return shim.Error(jsonResp)
	}

	//
	PTNAsset := modules.NewPTNAsset() //todo 改成PTN Asset
	if PTNAsset == nil {
		return shim.Error("need call setETHTokenAsset()")
	}
	invokeTokens := new(modules.AmountAsset)
	invokeTokens.Amount = PTNAmount
	invokeTokens.Asset = PTNAsset
	err = stub.PayOutToken(ptnRecvAddr, invokeTokens, 0)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.PayOutToken\"}"
		return shim.Error(jsonResp)
	}

	err = updateFeeSub(0, ptnAddr, stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getInputData(contractAddr, reqid, recvAddr string, ethAmount uint64, sig1, sig2, sig3 string, stub shim.ChaincodeStubInterface) (string, error) {
	const withdrawABI = `[{
		"constant": false,
		"inputs": [{
			"name": "recver",
			"type": "address"
		}, {
			"name": "amount",
			"type": "uint256"
		}, {
			"name": "reqid",
			"type": "bytes32"
		}, {
			"name": "sigstr1",
			"type": "bytes"
		}, {
			"name": "sigstr2",
			"type": "bytes"
		}, {
			"name": "sigstr3",
			"type": "bytes"
		}],
		"name": "withdraw",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}]`

	callerAddr := "0x7D7116A8706Ae08bAA7F4909e26728fa7A5f0365"

	//
	var invokeInput adaptor.CreateContractInvokeTxInput
	invokeInput.Address = callerAddr
	invokeInput.ContractAddress = contractAddr
	amt := new(big.Int)
	amt.SetString("21000000000000000", 10) //10000000000 10gwei*2100000
	invokeInput.Fee = adaptor.NewAmountAsset(amt, "ETH")
	invokeInput.Function = "withdraw"
	invokeInput.Extra = []byte(withdrawABI)
	invokeInput.Args = append(invokeInput.Args, []byte(recvAddr))
	amountBigInt := new(big.Int)
	amountBigInt.SetUint64(ethAmount)
	amountBigInt.Mul(amountBigInt, big.NewInt(1e10)) //eth's decimal is 18, ethToken in PTN is decimal is 8
	invokeInput.Args = append(invokeInput.Args, []byte(amountBigInt.String()))
	invokeInput.Args = append(invokeInput.Args, []byte(reqid))
	invokeInput.Args = append(invokeInput.Args, []byte(sig1))
	invokeInput.Args = append(invokeInput.Args, []byte(sig2))
	invokeInput.Args = append(invokeInput.Args, []byte(sig3))

	invokeInputJSON, _ := json.Marshal(invokeInput)
	invokeTxJSON, err := stub.OutChainCall("eth", "CreateContractInvokeTx", invokeInputJSON)
	if err != nil {
		log.Debugf("OutChainCall CreateContractInvokeTx err: %s", err.Error())
		return "", fmt.Errorf("OutChainCall CreateContractInvokeTx failed " + err.Error())
	}

	invokeOutput := adaptor.CreateContractInvokeTxOutput{}
	json.Unmarshal(invokeTxJSON, &invokeOutput)

	return fmt.Sprintf("%x", invokeOutput.Extra), nil
}

func (p *PTNETH) GetGoToETHSigs(stub shim.ChaincodeStubInterface, reqid string) pb.Response {
	if "0x" == reqid[0:2] || "0X" == reqid[0:2] {
		reqid = reqid[2:]
	}
	result, _ := stub.GetState(symbolsGoToETHSigs + reqid)
	if len(result) == 0 {
		return shim.Success([]byte{})
	}

	mapAddr := getETHContract(stub)
	if mapAddr == "" {
		return shim.Error(jsonResp1)
	}

	var withdraw Withdraw
	err := json.Unmarshal(result, &withdraw)
	if err != nil {
		return shim.Success([]byte{})
	}
	data, err := getInputData(mapAddr, reqid, withdraw.EthAddr, withdraw.PTNAmount-withdraw.PTNFee, withdraw.Sigs[0],
		withdraw.Sigs[1], withdraw.Sigs[2], stub)
	if err != nil {
		return shim.Success([]byte{})
	}
	return shim.Success([]byte(data))
}

func (p *PTNETH) Get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	result, _ := stub.GetState(key)
	return shim.Success(result)
}

func (p *PTNETH) WithdrawAmount(stub shim.ChaincodeStubInterface, withdrawAddr common.Address, amount decimal.Decimal) pb.Response {
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
		return shim.Error("Only owner can withdraw")
	}

	//contractAddr
	amount = amount.Mul(decimal.New(100000000, 0))
	amtToken := &modules.AmountAsset{Amount: uint64(amount.IntPart()), Asset: modules.NewPTNAsset()}
	err = stub.PayOutToken(withdrawAddr.String(), amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func main() {
	err := shim.Start(new(PTNETH))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
