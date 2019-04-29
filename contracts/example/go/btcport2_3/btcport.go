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
		return _initDepositAddr(args, stub)
	case "setBTCTokenAsset":
		return _setBTCTokenAsset(args, stub)
	case "setDepositAddr":
		return _setDepositAddr(args, stub)
	case "getBTCToken":
		return _getBTCToken(args, stub)
	case "withdrawBTC":
		return _withdrawBTC(args, stub)

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

//Method:GetPubkey, return pubkey string
type BTCAddress_GetPubkey struct {
	Method string `json:"method"`
}

//refer to the struct CreateMultiSigParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member.
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

func creatMulti(userPubkey string, juryPubkeys []string, stub shim.ChaincodeStubInterface) ([]byte, error) {
	//
	a := sort.StringSlice(juryPubkeys[0:])
	sort.Sort(a)
	//
	createMultiSigParams := BTCAddress_createMultiSig{Method: "CreateMultiSigAddress"}
	createMultiSigParams.M = 2 //mod
	createMultiSigParams.N = 3 //mod
	createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, userPubkey)
	for i := range juryPubkeys {
		createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, juryPubkeys[i])
	}
	creteMultiReqBytes, err := json.Marshal(createMultiSigParams)
	if err != nil {
		return nil, err
	}
	createMultiResult, err := stub.OutChainAddress("btc", creteMultiReqBytes)
	if err != nil {
		return nil, errors.New("OutChainAddress CreateMultiSigAddress failed: " + err.Error())
	}
	log.Debugf("creatMulti createMultiResult : %s", string(createMultiResult))

	return createMultiResult, nil
}

const symbolsBTCAsset = "btc_asset"

const symbolsJuryPubkeys = "juryPubkeys"
const symbolsCreateMulti = "createMulti_"
const symbolsMultiAddr = "addr_"
const symbolsRedeem = "redeem_"

const symbolsTx = "tx_"
const symbolsUnspend = "unspend_"
const symbolsSpent = "spent_"
const sep = "_"

func _initDepositAddr(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsJuryPubkeys)
	if len(saveResult) != 0 {
		return shim.Success([]byte("DepositAddr has been init"))
	}

	//
	getPubkeyParams := BTCAddress_GetPubkey{Method: "GetPubkey"}
	getPubkeyReqBytes, err := json.Marshal(getPubkeyParams)
	if err != nil {
		return shim.Error(err.Error())
	}
	result, err := stub.OutChainAddress("btc", getPubkeyReqBytes)
	if err != nil {
		log.Debugf("OutChainAddress GetPubkey err: %s", err.Error())
		return shim.Success([]byte("OutChainAddress GetPubkey failed"))
	}

	//
	sendResult, err := stub.SendJury(1, []byte("getPubkey"), []byte(result)) //todo 封装重构
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
	if len(juryMsg) != 2 { //mod
		return shim.Success([]byte("RecvJury result's len not enough"))
	}

	//
	var pubkeys []string
	for i := range juryMsg {
		pubkeys = append(pubkeys, string(juryMsg[i].Answer))
	}
	a := sort.StringSlice(pubkeys[0:])
	sort.Sort(a)

	pubkeysJson, err := json.Marshal(pubkeys)
	if err != nil {
		return shim.Error("pubkeys Marshal failed: " + err.Error())
	}

	// Write the state to the ledger
	err = stub.PutState(symbolsJuryPubkeys, pubkeysJson)
	if err != nil {
		return shim.Error("write " + symbolsJuryPubkeys + " failed: " + err.Error())
	}
	return shim.Success(pubkeysJson)
}

func _setBTCTokenAsset(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (AssetStr)")
	}
	asset, _, err := dm.String2AssetId(args[0])
	if err != nil {
		return shim.Success([]byte("AssetStr invalid"))
	}
	err = stub.PutState(symbolsBTCAsset, asset.Bytes())
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
	asset := new(dm.Asset)
	asset.SetBytes(result)
	return asset
}

func _setDepositAddr(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//params check
	if len(args) < 1 {
		return shim.Error("need 1 args (BTCPubkey)")
	}
	userPubkey := args[0]

	//
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	addrJson, _ := stub.GetState(symbolsMultiAddr + invokeAddr.String())
	if len(addrJson) != 0 {
		log.Debugf("symbolsMultiAddr+%s: %s", invokeAddr.String(), string(addrJson))
		return shim.Success([]byte("You have set depsoitAddr"))
	}
	log.Debugf("symbolsMultiAddr+%s need set", invokeAddr.String())
	//
	pubkeysJson, _ := stub.GetState(symbolsJuryPubkeys)
	if len(pubkeysJson) == 0 {
		log.Debugf("pubkeys is empty")
		return shim.Success([]byte("pubkeys is empty"))
	}
	var pubkeys []string
	err = json.Unmarshal(pubkeysJson, &pubkeys)
	if err != nil {
		log.Debugf("pubkeys Unmarshal failed")
		return shim.Success([]byte("pubkeys Unmarshal failed"))
	}
	if len(pubkeys) != 2 { //mod
		log.Debugf("pubkeys' length is not 2")
		return shim.Success([]byte("pubkeys' length is not 2")) //mod
	}

	createMultiResult, err := creatMulti(userPubkey, pubkeys, stub)
	if err != nil {
		log.Debugf("creatMulti failed: " + err.Error())
		return shim.Success([]byte("creatMulti failed: " + err.Error()))
	}
	log.Debugf("createMultiResult: " + string(createMultiResult))

	var createResult CreateMultiSigResult
	err = json.Unmarshal(createMultiResult, &createResult)
	if err != nil {
		log.Debugf("creatMulti result Unmarshal failed: " + err.Error())
		return shim.Success([]byte("creatMulti result Unmarshal failed" + err.Error()))
	}

	// Write the state to the ledger
	err = stub.PutState(symbolsCreateMulti+invokeAddr.String(), createMultiResult)
	if err != nil {
		log.Debugf("PutState symbolsCreateMulti failed err: %s", err.Error())
		return shim.Error("write symbolsCreateMulti failed: " + err.Error())
	}
	err = stub.PutState(symbolsMultiAddr+invokeAddr.String(), []byte(createResult.P2ShAddress))
	if err != nil {
		log.Debugf("PutState symbolsMultiAddr failed err: %s", err.Error())
		return shim.Error("PutState symbolsMultiAddr failed")
	}
	err = stub.PutState(symbolsRedeem+createResult.P2ShAddress, []byte(createResult.RedeemScript))
	if err != nil {
		log.Debugf("PutState symbolsRedeem failed err: %s", err.Error())
		return shim.Error("PutState symbolsRedeem failed")
	}
	log.Debugf("symbolsMultiAddr+invokeAddr.String(): %s", createResult.P2ShAddress)
	//
	return shim.Success(createMultiResult)
}

//refer to the struct VerifyMessageParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member.
type BTCTransaction_VerifyMessage struct {
	Method    string `json:"method"`
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
	verifyMessage := BTCTransaction_VerifyMessage{Method: "VerifyMessage"}
	verifyMessage.Message = btcTxHash
	verifyMessage.Signature = btcTxHashSig
	verifyMessage.Address = btcAddr

	//
	reqBytes, err := json.Marshal(verifyMessage)
	if err != nil {
		return false, err
	}
	verifyResultByte, err := stub.OutChainTransaction("btc", reqBytes)
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

//refer to the struct GetUTXOHttpParams in "github.com/palletone/adaptor/AdaptorBTC.go",
//add 'method' member.
type BTCTransaction_GetUTXOHttp struct {
	Method  string `json:"method"`
	Address string `json:"address"`
	Txid    string `json:"txid"`
}

//refert to the struct GetUTXOHttpResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type GetUTXOHttpResult struct {
	UTXOs []UTXO `json:"utxos"`
}
type UTXO struct {
	TxID     string  `json:"txid"`
	Vout     uint32  `json:"vout"`
	Amount   float64 `json:"amount"`
	Confirms uint64  `json:"confirms"`
}

func getAddrUTXO(btcAddr string, stub shim.ChaincodeStubInterface) (*GetUTXOHttpResult, error) {
	//
	getUTXO := BTCTransaction_GetUTXOHttp{Method: "GetUTXOHttp"}
	getUTXO.Address = btcAddr

	//
	reqBytes, err := json.Marshal(getUTXO)
	if err != nil {
		return nil, err
	}
	verifyResultByte, err := stub.OutChainTransaction("btc", reqBytes)
	if err != nil {
		return nil, err
	}

	//
	var result GetUTXOHttpResult
	err = json.Unmarshal(verifyResultByte, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func _getBTCToken(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	//
	multiAddrByte, _ := stub.GetState(symbolsMultiAddr + invokeAddr.String())
	if len(multiAddrByte) == 0 {
		jsonResp := "{\"Error\":\"You need call getDepositAddr for get your deposit address\"}"
		return shim.Success([]byte(jsonResp))
	}
	multiAddr := string(multiAddrByte)

	//
	getUTXOResult, err := getAddrUTXO(multiAddr, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"Get multiAddr's UTXO failed\"}"
		return shim.Success([]byte(jsonResp))
	}
	btcAmount := uint64(0)
	for i := range getUTXOResult.UTXOs {
		if getUTXOResult.UTXOs[i].Confirms < 1 { //
			continue
		}
		txIDVout := getUTXOResult.UTXOs[i].TxID + sep + strconv.Itoa(int(getUTXOResult.UTXOs[i].Vout))
		//
		unspendResult, _ := stub.GetState(symbolsTx + txIDVout)
		if len(unspendResult) != 0 {
			continue
		}
		err = stub.PutState(symbolsTx+txIDVout, []byte(invokeAddr.String()))
		if err != nil {
			log.Debugf("PutState txhash failed err: %s", err.Error())
			return shim.Error("PutState txhash failed")
		}
		//
		err = stub.PutState(symbolsUnspend+multiAddr+sep+txIDVout, []byte(Int64ToBytes(int64(getUTXOResult.UTXOs[i].Amount*1e8))))
		if err != nil {
			log.Debugf("PutState txhash unspend failed err: %s", err.Error())
			return shim.Error("PutState txhash unspend failed")
		}
		btcAmount += uint64(getUTXOResult.UTXOs[i].Amount * 1e8)
	}

	if btcAmount == 0 {
		return shim.Success([]byte("You need deposit"))
	}

	//
	btcTokenAsset := getBTCTokenAsset(stub)
	if btcTokenAsset == nil {
		return shim.Error("need call setBTCTokenAsset()")
	}
	invokeTokens := new(dm.InvokeTokens)
	invokeTokens.Amount = btcAmount
	invokeTokens.Asset = btcTokenAsset
	err = stub.PayOutToken(invokeAddr.String(), invokeTokens, 0)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.PayOutToken\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success([]byte("get success"))
}

type Unspend struct {
	MultiAddr string `json:"multiaddr"`
	Txid      string `json:"txid"`
	Vout      uint32 `json:"vout"`
	Value     int64  `json:"value"`
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
		keys := strings.Split(oneKV.Key, sep)
		if len(keys) < 4 {
			log.Debugf("invalid key: %s", oneKV.Key)
			continue
		}
		unspend := Unspend{}
		index, err := strconv.Atoi(keys[3])
		if err != nil {
			log.Debugf("invalid key index: %s", oneKV.Key)
			continue
		}
		unspend.Value = BytesToInt64(oneKV.Value)

		unspend.MultiAddr = keys[1]
		unspend.Txid = keys[2]
		unspend.Vout = uint32(index)
		if unspend.Value == btcAmout {
			selUnspends = append(selUnspends, unspend)
			break
		}
		smlUnspends = append(smlUnspends, unspend)
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
//add 'method' member.
type BTCTransaction_rawTransactionGen struct { //GetTransactionHttpParams
	Method   string   `json:"method"`
	Inputs   []Input  `json:"inputs"`
	Outputs  []Output `json:"outputs"`
	Locktime int64    `json:"locktime"`
}
type Input struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
	Addr string `json:"addr"`
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
	rawTxGen := BTCTransaction_rawTransactionGen{Method: "RawTransactionGe"}
	totalAmount := int64(0)
	for i := range unspends {
		rawTxGen.Inputs = append(rawTxGen.Inputs, Input{Txid: unspends[i].Txid, Vout: uint32(unspends[i].Vout)})
		totalAmount += unspends[i].Value
	}
	rawTxGen.Outputs = append(rawTxGen.Outputs, Output{btcAddr, converAmount(btcAmout - btcFee)})
	if totalAmount > btcAmout {
		rawTxGen.Outputs = append(rawTxGen.Outputs, Output{unspends[0].MultiAddr, converAmount(totalAmount - btcAmout)})
	}

	//
	reqBytes, err := json.Marshal(rawTxGen)
	if err != nil {
		return "", err
	}
	resultByte, err := stub.OutChainTransaction("btc", reqBytes)
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
//add 'method' member.
type BTCTransaction_signTransaction struct { //SignTransactionParams
	Method           string   `json:"method"`
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
	signTx := BTCTransaction_signTransaction{Method: "SignTransaction"}
	signTx.TransactionHex = rawTx
	signTx.InputRedeemIndex = inputRedeemIndex
	signTx.RedeemHex = redeemHex

	//
	reqBytes, err := json.Marshal(signTx)
	if err != nil {
		return "", err
	}
	resultByte, err := stub.OutChainTransaction("btc", reqBytes)
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
//add 'method' member.
type BTCTransaction_mergeTransaction struct {
	Method               string   `json:"method"`
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
	mergeTx := BTCTransaction_mergeTransaction{Method: "MergeTransaction"}
	mergeTx.UserTransactionHex = rawTx
	mergeTx.InputRedeemIndex = inputRedeemIndex
	mergeTx.RedeemHex = redeemHex

	//
	var answers []string
	for i := range juryMsg {
		answers = append(answers, string(juryMsg[i].Answer))
	}
	a := sort.StringSlice(answers[0:])
	sort.Sort(a)
	//
	var result MergeTransactionResult
	array := [][3]int{{1, 2}}
	num := 1
	for i := 0; i < num; i++ {
		mergeTx.MergeTransactionHexs = []string{}
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, string(answers[array[i][0]]))
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, string(answers[array[i][1]]))
		//
		reqBytes, err := json.Marshal(mergeTx)
		if err != nil {
			continue
		}
		resultByte, err := stub.OutChainTransaction("btc", reqBytes)
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
//add 'method' member.
type BTCTransaction_sendTransactionHttp struct {
	Method         string `json:"method"`
	TransactionHex string `json:"transactionhex"`
}

//refert to the struct type SendTransactionHttpResult in "github.com/palletone/adaptor/AdaptorBTC.go",
type SendTransactionHttpResult struct {
	TransactionHah string `json:"transactionhash"`
}

func sendTx(tx string, stub shim.ChaincodeStubInterface) (string, error) {
	//
	rawTxGen := BTCTransaction_sendTransactionHttp{Method: "SendTransactionHttp"}
	rawTxGen.TransactionHex = tx

	//
	reqBytes, err := json.Marshal(rawTxGen)
	if err != nil {
		return "", err
	}
	resultByte, err := stub.OutChainTransaction("btc", reqBytes)
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
		err := stub.DelState(symbolsUnspend + selUnspnds[i].MultiAddr + sep + selUnspnds[i].Txid + sep + strconv.Itoa(int(selUnspnds[i].Vout)))
		if err != nil {
			log.Debugf("DelState txhash unspend failed err: %s", err.Error())
			return errors.New("DelState txhash unspend failed")
		}
		err = stub.PutState(symbolsSpent+selUnspnds[i].MultiAddr+sep+selUnspnds[i].Txid+sep+strconv.Itoa(int(selUnspnds[i].Vout)),
			[]byte(Int64ToBytes(selUnspnds[i].Value)))
		if err != nil {
			log.Debugf("PutState txhash spent failed err: %s", err.Error())
			return errors.New("PutState txhash spent failed")
		}
	}

	if totalAmount > btcTokenAmount {
		err := stub.PutState(symbolsUnspend+selUnspnds[0].MultiAddr+sep+txHash+sep+strconv.Itoa(1), []byte(Int64ToBytes(totalAmount-btcTokenAmount)))
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
	rawTx, err := genRawTx(int64(btcTokenAmount), btcFee, btcAddr, selUnspnds, stub)

	//
	inputRedeemIndex := []int{}
	redeemHex := []string{}
	mapRedeem := make(map[string]int)
	index := 0
	for i := range selUnspnds {
		if _, exist := mapRedeem[selUnspnds[i].MultiAddr]; exist {
			inputRedeemIndex = append(inputRedeemIndex, mapRedeem[selUnspnds[i].MultiAddr])
			continue
		}
		result, _ := stub.GetState(symbolsRedeem + selUnspnds[i].MultiAddr)
		if len(result) == 0 {
			return shim.Error("DepsoitRedeem is empty")
		}
		redeemHex = append(redeemHex, string(result))
		mapRedeem[selUnspnds[i].MultiAddr] = index
		inputRedeemIndex = append(inputRedeemIndex, index)
		index++
	}
	// 签名交易
	rawTxSign, err := signTx(rawTx, inputRedeemIndex, redeemHex, stub)
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
	if len(juryMsg) < 2 { //mod
		return shim.Success([]byte("RecvJury result's len not enough"))
	}

	// 合并交易
	tx, err := mergeTx(rawTx, inputRedeemIndex, redeemHex, juryMsg, stub)
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
	err := shim.Start(new(BTCPort))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
