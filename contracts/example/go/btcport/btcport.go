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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	dm "github.com/palletone/go-palletone/dag/modules"
)

type BTCPort struct {
}

func (p *BTCPort) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *BTCPort) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "initDepositAddr":
		return _initDepositAddr(stub)
	case "setBTCTokenAsset":
		return _setBTCTokenAsset(args, stub)
	case "getDepositAddr":
		return _getDepositAddr(stub)
	case "_getBTCToken":
		return _getBTCToken(args, stub)
	case "withdrawBTC":
		return _withdrawBTC(args, stub)

	case "put":
		return put(stub)
	case "get":
		return get(stub)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

type JuryMsgAddr struct {
	Address string
	Answer  []byte
}

//refer to the struct CreateMultiSigParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCAddress_createMultiSig struct {
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

func creatMulti(juryMsg []JuryMsgAddr, stub shim.ChaincodeStubInterface) ([]byte, error) {
	//
	answers := make([]string, 0, len(juryMsg))
	for i := range juryMsg {
		answers = append(answers, string(juryMsg[i].Answer))
	}
	a := sort.StringSlice(answers[0:])
	sort.Sort(a)
	//
	var createMultiSigParams BTCAddress_createMultiSig
	createMultiSigParams.M = 3
	createMultiSigParams.N = 4
	for i := range answers {
		createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, answers[i])
	}
	creteMultiReqBytes, err := json.Marshal(createMultiSigParams)
	if err != nil {
		return nil, err
	}
	createMultiResult, err := stub.OutChainCall("btc", "CreateMultiSigAddress", creteMultiReqBytes)
	if err != nil {
		return nil, errors.New("OutChainCall CreateMultiSigAddress failed: " + err.Error())
	}
	log.Debugf("OutChainCall CreateMultiSigAddress createMultiResult ==== ===== %s", createMultiResult)

	return createMultiResult, nil
}

const symbolsDeposit = "createMultiResult"
const symbolsDepositAddr = "btc_multsigAddr"
const symbolsDepositRedeem = "btc_redeem"

const symbolsBTCAsset = "btc_asset"

const symbolsTx = "tx_"
const symbolsUnspend = "unspend_"
const symbolsSpent = "spent_"
const sep = "_"

func _initDepositAddr(stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsDeposit)
	if len(saveResult) != 0 {
		return shim.Success([]byte("DepositAddr has been init"))
	}

	//Method:GetJuryBTCPubkey, return pubkey string
	result, err := stub.OutChainCall("btc", "GetJuryBTCPubkey", []byte(""))
	if err != nil {
		log.Debugf("OutChainCall GetJuryBTCPubkey err: %s", err.Error())
		return shim.Success([]byte("OutChainCall GetJuryBTCPubkey failed"))
	}

	//
	sendResult, err := stub.SendJury(1, []byte("getPubkey"), result) //todo 封装重构
	if err != nil {
		log.Debugf("SendJury getPubkey err: %s", err.Error())
		return shim.Success([]byte("SendJury getPubkey failed"))
	}
	log.Debugf("sendResult: %s", common.Bytes2Hex(sendResult))
	recvResult, err := stub.RecvJury(1, []byte("getPubkey"), 2)
	if err != nil {
		log.Debugf("RecvJury getPubkey err: %s", err.Error())
		return shim.Success([]byte("RecvJury failed"))
	}
	log.Debugf("recvResult: %s", string(recvResult))
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Success([]byte("Unmarshal result failed: " + err.Error()))
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) != 4 {
		return shim.Success([]byte("RecvJury result's len not enough"))
	}

	//
	createMultiResult, err := creatMulti(juryMsg, stub)
	if err != nil {
		return shim.Success([]byte("creatMulti failed" + err.Error()))
	}

	var createResult CreateMultiSigResult
	err = json.Unmarshal(createMultiResult, &createResult)
	if err != nil {
		return shim.Success([]byte("creatMulti result Unmarshal failed" + err.Error()))
	}

	// Write the state to the ledger
	err = stub.PutState(symbolsDeposit, createMultiResult)
	if err != nil {
		return shim.Error("write " + symbolsDeposit + " failed: " + err.Error())
	}
	err = stub.PutState(symbolsDepositAddr, []byte(createResult.P2ShAddress))
	if err != nil {
		return shim.Error("write " + symbolsDepositAddr + " failed: " + err.Error())
	}
	err = stub.PutState(symbolsDepositRedeem, []byte(createResult.RedeemScript))
	if err != nil {
		return shim.Error("write " + symbolsDepositRedeem + " failed: " + err.Error())
	}
	return shim.Success(createMultiResult)
}

func _setBTCTokenAsset(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (AssetStr)")
	}
	err := stub.PutState(symbolsBTCAsset, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsBTCAsset failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getBTCTokenAsset(stub shim.ChaincodeStubInterface) *dm.Asset {
	result, _ := stub.GetState(symbolsBTCAsset)
	if len(result) == 0 {
		return nil
	}
	asset, _ := dm.StringToAsset(string(result))
	log.Debugf("resultHex %s, asset: %s", common.Bytes2Hex(result), asset.String())

	return asset
}

func _getDepositAddr(stub shim.ChaincodeStubInterface) pb.Response {
	result, _ := stub.GetState(symbolsDepositAddr)
	if len(result) == 0 {
		return shim.Error("DepsoitAddr is empty")
	}
	return shim.Success(result)
}

//refer to the struct GetTransactionHttpParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCTransaction_getTxHTTP struct {
	//GetTransactionHttpParams
	TxHash string `json:"txhash"`
}

//refert to the struct GetTransactionHttpResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type GetTransactionHttpResult struct {
	Confirms uint64        `json:"confirms"`
	Inputs   []Input       `json:"inputs"`
	Outputs  []OutputIndex `json:"outputs"`
}
type Input struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
	Addr string `json:"addr"`
}
type OutputIndex struct {
	Index uint32 `json:"index"`
	Addr  string `json:"addr"`
	Value int64  `json:"value"` //satoshi
}

func getDepositBTCInfo(btcTxHash string, stub shim.ChaincodeStubInterface) (uint64, string, []OutputIndex, error) {
	//
	getTxHttp := BTCTransaction_getTxHTTP{btcTxHash}

	var outputs []OutputIndex
	//
	reqBytes, err := json.Marshal(getTxHttp)
	if err != nil {
		return 0, "", outputs, err
	}
	getTxHttpResult, err := stub.OutChainCall("btc", "GetTransactionHttp", reqBytes)
	if err != nil {
		return 0, "", outputs, err
	}

	//
	var getTxResult GetTransactionHttpResult
	err = json.Unmarshal(getTxHttpResult, &getTxResult)
	if err != nil {
		return 0, "", outputs, err
	}
	if getTxResult.Confirms < 6 {
		return 0, "", outputs, errors.New("Confirms is less than 6, please wait")
	}

	//
	result, _ := stub.GetState(symbolsDepositAddr)
	if len(result) == 0 {
		return 0, "", outputs, errors.New("DepsoitAddr is empty")
	}
	depositAddr := string(result)

	//
	depositAmount := int64(0)
	for i := range getTxResult.Outputs {
		if getTxResult.Outputs[i].Addr == depositAddr {
			depositAmount += getTxResult.Outputs[i].Value
			outputs = append(outputs, getTxResult.Outputs[i])
		}
	}
	if depositAmount == 0 {
		return 0, "", outputs, errors.New("Deposit amount is empty")
	}

	return uint64(depositAmount), getTxResult.Inputs[0].Addr, outputs, nil
}

//refer to the struct VerifyMessageParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCTransaction_VerifyMessage struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	Address   string `json:"address"`
}

//refert to the struct VerifyMessageResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type VerifyMessageResult struct {
	Valid bool `json:"valid"`
}

func verifySig(btcTxHash, btcTxHashSig, btcAddr string, stub shim.ChaincodeStubInterface) (bool, error) {
	//
	var verifyMessage BTCTransaction_VerifyMessage
	verifyMessage.Message = btcTxHash
	verifyMessage.Signature = btcTxHashSig
	verifyMessage.Address = btcAddr

	//
	reqBytes, err := json.Marshal(verifyMessage)
	if err != nil {
		return false, err
	}
	verifyResultByte, err := stub.OutChainCall("btc", "VerifyMessage", reqBytes)
	if err != nil {
		return false, err
	}

	//
	var result VerifyMessageResult
	err = json.Unmarshal(verifyResultByte, &result)
	if err != nil {
		return false, err
	}

	return result.Valid, nil
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func _getBTCToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 2 {
		return shim.Error("need 2 args (btcTxHash, btcTxHashSig(signed by input btcAddr))")
	}

	//
	btcTxHash := args[0]
	unspendResult, _ := stub.GetState(symbolsTx + btcTxHash)
	if len(unspendResult) != 0 {
		jsonResp := "{\"Error\":\"The tx has been used\"}"
		return shim.Success([]byte(jsonResp))
	}

	//
	btcAmount, btcAddr, outputs, err := getDepositBTCInfo(btcTxHash, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"Have get token\"}"
		return shim.Success([]byte(jsonResp))
	}

	//verify message
	btcTxHashSig := args[1]
	valid, err := verifySig(btcTxHash, btcTxHashSig, btcAddr, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"verifySig failed," + err.Error() + "\"}"
		return shim.Success([]byte(jsonResp))
	}
	if !valid {
		jsonResp := "{\"Error\":\"You are not the Depositor\"}"
		return shim.Success([]byte(jsonResp))
	}

	//
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(symbolsTx+btcTxHash, []byte(invokeAddr.String()))
	if err != nil {
		log.Debugf("PutState txhash failed err: %s", err.Error())
		return shim.Error("PutState txhash failed")
	}
	for i := range outputs {
		err = stub.PutState(symbolsUnspend+btcTxHash+sep+strconv.Itoa(int(outputs[i].Index)),
			Int64ToBytes(outputs[i].Value))
		if err != nil {
			log.Debugf("PutState txhash unspend failed err: %s", err.Error())
			return shim.Error("PutState txhash unspend failed")
		}
	}

	//
	btcTokenAsset := getBTCTokenAsset(stub)
	if btcTokenAsset == nil {
		return shim.Error("need call setBTCTokenAsset()")
	}
	invokeTokens := new(dm.AmountAsset)
	invokeTokens.Amount = btcAmount
	invokeTokens.Asset = btcTokenAsset
	err = stub.PayOutToken(invokeAddr.String(), invokeTokens, 0)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.PayOutToken\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success([]byte("put failed"))
}

type Unspend struct {
	Txid  string `json:"txid"`
	Vout  uint32 `json:"vout"`
	Value int64  `json:"value"`
}

// A slice of Unspend that implements sort.Interface to sort by Value.
type UnspendList []Unspend

func (p UnspendList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p UnspendList) Len() int           { return len(p) }
func (p UnspendList) Less(i, j int) bool { return p[i].Value > p[j].Value }

// A function to turn a map into a UnspendList, then sort and return it.
func sortByValue(ul UnspendList) UnspendList {
	sort.Stable(ul) //sort.Sort(ul)
	return ul
}

func getUnspends(btcAmout int64, stub shim.ChaincodeStubInterface) []Unspend {
	KVs, _ := stub.GetStateByPrefix(symbolsUnspend)
	var smlUnspends []Unspend
	var bigUnspends []Unspend
	var selUnspends []Unspend
	for _, oneKV := range KVs {
		unspend := Unspend{}
		keys := strings.Split(oneKV.Key, sep)
		if len(keys) < 3 {
			log.Debugf("invalid key: %s", oneKV.Key)
			continue
		}
		index, err := strconv.Atoi(keys[2])
		if err != nil {
			log.Debugf("invalid key index: %s", oneKV.Key)
			continue
		}
		unspend.Value = BytesToInt64(oneKV.Value)

		unspend.Txid = keys[1]
		unspend.Vout = uint32(index)
		if unspend.Value == btcAmout {
			selUnspends = append(selUnspends, unspend)
			break
		} else if unspend.Value > btcAmout {
			bigUnspends = append(bigUnspends, unspend)
		} else {
			smlUnspends = append(smlUnspends, unspend)
		}
	}
	//
	if len(selUnspends) != 0 {
		return selUnspends
	}
	//
	selAmount := int64(0)
	if len(smlUnspends) > 0 {
		smlUnspendsSort := sortByValue(smlUnspends)
		for i := range smlUnspendsSort {
			selAmount += smlUnspends[i].Value
			selUnspends = append(selUnspends, smlUnspends[i])
			if selAmount >= btcAmout {
				break
			}
		}
	}
	if selAmount >= btcAmout {
		return selUnspends
	}
	//
	if len(bigUnspends) == 0 {
		return bigUnspends
	}
	selUnspends = []Unspend{}
	minIndex := int64(0)
	minValue := bigUnspends[0].Value
	for i := range bigUnspends {
		if bigUnspends[i].Value < minValue {
			minIndex = int64(i)
			minValue = bigUnspends[i].Value
		}
	}
	selUnspends = append(selUnspends, bigUnspends[minIndex])
	return selUnspends
}

//refer to the struct RawTransactionGenParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCTransaction_rawTransactionGen struct {
	//GetTransactionHttpParams
	Inputs   []Input  `json:"inputs"`
	Outputs  []Output `json:"outputs"`
	Locktime int64    `json:"locktime"`
}
type Output struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"` //btc
}

//refert to the struct RawTransactionGenResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type RawTransactionGenResult struct {
	Rawtx string `json:"rawtx"`
}

func converAmount(a int64) float64 {
	return float64(a) / math.Pow10(int(8))
}
func genRawTx(btcAmout, btcFee int64, btcAddr string, unspends []Unspend, stub shim.ChaincodeStubInterface) (string, error) {
	//
	var rawTxGen BTCTransaction_rawTransactionGen
	totalAmount := int64(0)
	for i := range unspends {
		rawTxGen.Inputs = append(rawTxGen.Inputs, Input{Txid: unspends[i].Txid, Vout: unspends[i].Vout})
		totalAmount += unspends[i].Value
	}
	rawTxGen.Outputs = append(rawTxGen.Outputs, Output{btcAddr, converAmount(btcAmout - btcFee)})
	if totalAmount > btcAmout {
		//
		result, _ := stub.GetState(symbolsDepositAddr)
		if len(result) == 0 {
			return "", errors.New("DepsoitAddr is empty")
		}
		rawTxGen.Outputs = append(rawTxGen.Outputs, Output{string(result), converAmount(totalAmount - btcAmout)})
	}

	//
	reqBytes, err := json.Marshal(rawTxGen)
	if err != nil {
		return "", err
	}
	resultByte, err := stub.OutChainCall("btc", "RawTransactionGen", reqBytes)
	if err != nil {
		return "", err
	}

	//
	var result RawTransactionGenResult
	err = json.Unmarshal(resultByte, &result)
	if err != nil {
		return "", err
	}

	return result.Rawtx, nil
}

//refer to the struct SignTransactionParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCTransaction_signTransaction struct {
	//SignTransactionParams
	TransactionHex   string   `json:"transactionhex"`
	InputRedeemIndex []int    `json:"inputredeemindex"`
	RedeemHex        []string `json:"redeemhex"`
}

//refert to the struct SignTransactionResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type SignTransactionResult struct {
	Complete       bool   `json:"complete"`
	TransactionHex string `json:"transactionhex"`
}

func signTx(rawTx string, inputRedeemIndex []int, redeemHex []string, stub shim.ChaincodeStubInterface) (string, error) {
	//
	var signTx BTCTransaction_signTransaction
	signTx.TransactionHex = rawTx
	signTx.InputRedeemIndex = inputRedeemIndex
	signTx.RedeemHex = redeemHex

	//
	reqBytes, err := json.Marshal(signTx)
	if err != nil {
		return "", err
	}
	resultByte, err := stub.OutChainCall("btc", "SignTransaction", reqBytes)
	if err != nil {
		return "", err
	}

	//
	var signResult SignTransactionResult
	err = json.Unmarshal(resultByte, &signResult)
	if err != nil {
		return "", err
	}

	return signResult.TransactionHex, nil
}

//refer to the struct MergeTransactionParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCTransaction_mergeTransaction struct {
	UserTransactionHex   string   `json:"usertransactionhex"`
	MergeTransactionHexs []string `json:"mergetransactionhexs"`
	InputRedeemIndex     []int    `json:"inputredeemindex"`
	RedeemHex            []string `json:"redeemhex"`
}

//refert to the struct MergeTransactionResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type MergeTransactionResult struct {
	Complete       bool   `json:"complete"`
	TransactionHex string `json:"transactionhex"`
}

func mergeTx(rawTx string, inputRedeemIndex []int, redeemHex []string, juryMsg []JuryMsgAddr, stub shim.ChaincodeStubInterface) (string, error) {
	//
	var mergeTx BTCTransaction_mergeTransaction
	mergeTx.UserTransactionHex = rawTx
	mergeTx.InputRedeemIndex = inputRedeemIndex
	mergeTx.RedeemHex = redeemHex

	//
	answers := make([]string, 0, len(juryMsg))
	for i := range juryMsg {
		answers = append(answers, string(juryMsg[i].Answer))
	}
	a := sort.StringSlice(answers[0:])
	sort.Sort(a)
	//
	var result MergeTransactionResult
	array := [][3]int{{1, 2, 3}, {1, 2, 4}, {1, 3, 4}, {2, 3, 4}}
	num := 4
	if len(answers) == 3 {
		num = 1
	}
	for i := 0; i < num; i++ {
		mergeTx.MergeTransactionHexs = []string{}
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, answers[array[i][0]])
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, answers[array[i][1]])
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, answers[array[i][2]])
		//
		reqBytes, err := json.Marshal(mergeTx)
		if err != nil {
			continue
		}
		resultByte, err := stub.OutChainCall("btc", "MergeTransaction", reqBytes)
		if err != nil {
			continue
		}
		err = json.Unmarshal(resultByte, &result)
		if err != nil {
			continue
		}
		if result.Complete {
			break
		}
	} //for

	if result.Complete {
		return result.TransactionHex, nil
	}
	return "", errors.New("Not complete")
}

//refer to the struct SendTransactionHttpParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCTransaction_sendTransactionHttp struct {
	TransactionHex string `json:"transactionhex"`
}

//refert to the struct type SendTransactionHttpResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type SendTransactionHttpResult struct {
	TransactionHah string `json:"transactionhash"`
}

func sendTx(tx string, stub shim.ChaincodeStubInterface) (string, error) {
	//
	rawTxGen := BTCTransaction_sendTransactionHttp{tx}

	//
	reqBytes, err := json.Marshal(rawTxGen)
	if err != nil {
		return "", err
	}
	resultByte, err := stub.OutChainCall("btc", "SendTransactionHttp", reqBytes)
	if err != nil {
		return "", err
	}

	//
	var result SendTransactionHttpResult
	err = json.Unmarshal(resultByte, &result)
	if err != nil {
		return "", err
	}

	return result.TransactionHah, nil
}

func saveUtxos(btcTokenAmount int64, selUnspnds []Unspend, txHash string, stub shim.ChaincodeStubInterface) error {
	totalAmount := int64(0)
	for i := range selUnspnds {
		totalAmount += selUnspnds[i].Value
		err := stub.DelState(symbolsUnspend + selUnspnds[i].Txid + sep + strconv.Itoa(int(selUnspnds[i].Vout)))
		if err != nil {
			log.Debugf("DelState txhash unspend failed err: %s", err.Error())
			return errors.New("DelState txhash unspend failed")
		}
		err = stub.PutState(symbolsSpent+selUnspnds[i].Txid+sep+strconv.Itoa(int(selUnspnds[i].Vout)),
			Int64ToBytes(selUnspnds[i].Value))
		if err != nil {
			log.Debugf("PutState txhash spent failed err: %s", err.Error())
			return errors.New("PutState txhash spent failed")
		}
	}

	if totalAmount > btcTokenAmount {
		err := stub.PutState(symbolsUnspend+txHash+sep+strconv.Itoa(1), Int64ToBytes(totalAmount-btcTokenAmount))
		if err != nil {
			log.Debugf("PutState txhash unspend failed err: %s", err.Error())
			return errors.New("PutState txhash unspend failed")
		}
	}

	return nil
}

func _withdrawBTC(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (btcAddr, [btcFee(>10000)])")
	}
	btcAddr := args[0]
	btcFee := int64(0)
	if len(args) > 1 {
		btcFee, _ = strconv.ParseInt(args[1], 10, 64)
	}
	if btcFee <= 100000 {
		btcFee = 100000
	}
	//
	btcTokenAsset := getBTCTokenAsset(stub)
	if btcTokenAsset == nil {
		return shim.Error("need call setBTCTokenAsset()")
	}
	//contractAddr
	_, contractAddr := stub.GetContractID()

	//check token
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		jsonResp := "{\"Error\":\"GetInvokeTokens failed\"}"
		return shim.Success([]byte(jsonResp))
	}
	btcTokenAmount := uint64(0)
	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Address == contractAddr {
			if invokeTokens[i].Asset.AssetId == btcTokenAsset.AssetId {
				btcTokenAmount += invokeTokens[i].Amount
			}
		}
	}

	// 取未花费
	selUnspnds := getUnspends(int64(btcTokenAmount), stub)
	if len(selUnspnds) == 0 {
		jsonResp := "{\"Error\":\"getUnspends failed\"}"
		return shim.Success([]byte(jsonResp))
	}
	// 产生交易
	rawTx, _ := genRawTx(int64(btcTokenAmount), btcFee, btcAddr, selUnspnds, stub)

	//
	result, _ := stub.GetState(symbolsDepositRedeem)
	if len(result) == 0 {
		return shim.Error("DepsoitRedeem is empty")
	}
	redeemHex := string(result)

	inputRedeemIndex := []int{}
	for i := len(selUnspnds); i > 0; i-- {
		inputRedeemIndex = append(inputRedeemIndex, 0)
	}
	// 签名交易
	rawTxSign, _ := signTx(rawTx, inputRedeemIndex, []string{redeemHex}, stub)
	//协商交易
	sendResult, err := stub.SendJury(2, []byte(rawTx), []byte(rawTxSign)) //todo 封装重构
	if err != nil {
		log.Debugf("SendJury rawTx err: %s", err.Error())
		return shim.Success([]byte("SendJury rawTx failed"))
	}
	log.Debugf("sendResult: %s", common.Bytes2Hex(sendResult))
	recvResult, err := stub.RecvJury(2, []byte("getPubkey"), 2)
	if err != nil {
		log.Debugf("RecvJury rawTx err: %s", err.Error())
		return shim.Success([]byte("RecvJury rawTx failed"))
	}
	log.Debugf("recvResult: %s", string(recvResult))
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Success([]byte("Unmarshal result failed: " + err.Error()))
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) < 3 {
		return shim.Success([]byte("RecvJury result's len not enough"))
	}

	// 合并交易
	tx, err := mergeTx(rawTx, inputRedeemIndex, []string{redeemHex}, juryMsg, stub)
	if err != nil {
		return shim.Success([]byte("mergeTx failed: " + err.Error()))
	}
	// 发送交易
	txHash, err := sendTx(tx, stub)
	if err != nil {
		return shim.Success([]byte("sendTx failed: " + err.Error()))
	}
	_ = txHash

	// 记录花费
	err = saveUtxos(int64(btcTokenAmount), selUnspnds, txHash, stub)
	if err != nil {
		return shim.Success([]byte("sendTx failed: " + err.Error()))
	}
	return shim.Success([]byte(txHash))
}
func put(stub shim.ChaincodeStubInterface) pb.Response {
	err := stub.PutState("result", []byte("put"))
	if err != nil {
		log.Debugf("PutState err: %s", err.Error())
		return shim.Error("PutState failed")
	}
	log.Debugf("ok")
	return shim.Success([]byte("PutState OK"))
}

func get(stub shim.ChaincodeStubInterface) pb.Response {
	result, _ := stub.GetState("result")
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(BTCPort))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
