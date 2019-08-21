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
	"github.com/palletone/go-palletone/common/crypto"
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
	case "setDepositAddr":
		return _setDepositAddr(args, stub)
	case "getBTCToken":
		return _getBTCToken(stub)

	case "withdrawPrepare":
		return _withdrawPrepare(args, stub)
	case "withdrawBTC":
		return _withdrawBTC(args, stub)

	case "send":
		return send(args, stub)

	case "get":
		return get(args, stub)
	case "getAsset":
		return getAsset(stub)
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

func creatMulti(userPubkey string, juryPubkeys []string, stub shim.ChaincodeStubInterface) ([]byte, error) {
	//
	a := sort.StringSlice(juryPubkeys[0:])
	sort.Sort(a)
	//
	var createMultiSigParams BTCAddress_createMultiSig
	createMultiSigParams.M = consultM     //mod
	createMultiSigParams.N = consultN + 1 //mod
	createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, userPubkey)
	for i := range juryPubkeys {
		createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, juryPubkeys[i])
	}
	creteMultiReqBytes, err := json.Marshal(createMultiSigParams)
	if err != nil {
		return nil, err
	}
	createMultiResult, err := stub.OutChainCall("btc", "CreateMultiSigAddress", creteMultiReqBytes)
	if err != nil {
		return nil, errors.New("OutChainCall CreateMultiSigAddress failed: " + err.Error())
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

const symbolsWithdrawPrepare = "withdrawPrepare_"
const symbolsWithdraw = "withdraw_"

const consultM = 3
const consultN = 4

func _initDepositAddr(stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsJuryPubkeys)
	if len(saveResult) != 0 {
		return shim.Error("DepositAddr has been init")
	}

	//Method:GetJuryBTCPubkey, return pubkey string
	result, err := stub.OutChainCall("btc", "GetJuryBTCPubkey", []byte(""))
	if err != nil {
		log.Debugf("OutChainCall GetJuryBTCPubkey err: %s", err.Error())
		return shim.Error("OutChainCall GetJuryBTCPubkey failed")
	}

	//
	recvResult, err := consult(stub, []byte("getPubkey"), result)
	if err != nil {
		return shim.Error("consult getPubkey failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Error("Unmarshal result failed: " + err.Error())
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) != consultN { //mod
		return shim.Error("RecvJury result's len not enough")
	}

	//
	pubkeys := make([]string, 0, len(juryMsg))
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

	//
	saveResult, _ := stub.GetState(symbolsBTCAsset)
	if len(saveResult) != 0 {
		return shim.Error("TokenAsset has been init")
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
		return shim.Error("You have set depsoitAddr")
	}
	log.Debugf("symbolsMultiAddr+%s need set", invokeAddr.String())
	//
	pubkeysJson, _ := stub.GetState(symbolsJuryPubkeys)
	if len(pubkeysJson) == 0 {
		log.Debugf("pubkeys is empty")
		return shim.Error("pubkeys is empty")
	}
	var pubkeys []string
	err = json.Unmarshal(pubkeysJson, &pubkeys)
	if err != nil {
		log.Debugf("pubkeys Unmarshal failed")
		return shim.Error("pubkeys Unmarshal failed")
	}
	if len(pubkeys) != consultN { //mod
		log.Debugf("Jury pubkeys' length is not enough")
		return shim.Error("Jury pubkeys' length is not enough")
	}

	createMultiResult, err := creatMulti(userPubkey, pubkeys, stub)
	if err != nil {
		log.Debugf("creatMulti failed: " + err.Error())
		return shim.Error("creatMulti failed: " + err.Error())
	}
	log.Debugf("createMultiResult: " + string(createMultiResult))

	var createResult CreateMultiSigResult
	err = json.Unmarshal(createMultiResult, &createResult)
	if err != nil {
		log.Debugf("creatMulti result Unmarshal failed: " + err.Error())
		return shim.Error("creatMulti result Unmarshal failed" + err.Error())
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

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

//refer to the struct GetUTXOHttpParams in "github.com/palletone/adaptor/AdaptorBTC.go",
type BTCQuery_GetUTXOHttp struct {
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
	getUTXO := BTCQuery_GetUTXOHttp{Address: btcAddr}
	reqBytes, err := json.Marshal(getUTXO)
	if err != nil {
		return nil, err
	}
	verifyResultByte, err := stub.OutChainCall("btc", "GetUTXOHttp", reqBytes)
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

func _getBTCToken(stub shim.ChaincodeStubInterface) pb.Response {
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
		return shim.Error(jsonResp)
	}
	multiAddr := string(multiAddrByte)

	//
	getUTXOResult, err := getAddrUTXO(multiAddr, stub)
	if err != nil {
		jsonResp := "{\"Error\":\"Get multiAddr's UTXO failed\"}"
		return shim.Error(jsonResp)
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
		err = stub.PutState(symbolsUnspend+multiAddr+sep+txIDVout, Int64ToBytes(int64(getUTXOResult.UTXOs[i].Amount*1e8)))
		if err != nil {
			log.Debugf("PutState txhash unspend failed err: %s", err.Error())
			return shim.Error("PutState txhash unspend failed")
		}
		btcAmount += uint64(getUTXOResult.UTXOs[i].Amount * 1e8)
	}

	if btcAmount == 0 {
		return shim.Error("You need deposit")
	}

	//
	btcTokenAsset := getBTCTokenAsset(stub)
	if btcTokenAsset == nil {
		return shim.Error("need call setBTCTokenAsset()")
	}
	log.Debugf("btcAmount: %d, asset: %s", btcAmount, btcTokenAsset.String())

	invokeTokens := new(dm.AmountAsset)
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
	var rawTxGen BTCTransaction_rawTransactionGen
	totalAmount := int64(0)
	for i := range unspends {
		rawTxGen.Inputs = append(rawTxGen.Inputs, Input{Txid: unspends[i].Txid, Vout: unspends[i].Vout})
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
	Complete        bool   `json:"complete"`
	TransactionHash string `json:"transactionhash"`
	TransactionHex  string `json:"transactionhex"`
}

func mergeTx(rawTx string, inputRedeemIndex []int, redeemHex []string, juryMsg []JuryMsgAddr, stub shim.ChaincodeStubInterface) (string, string, error) {
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
	//array := [][2]int{{1, 2}} //mod
	//num := 1

	for i := 0; i < num; i++ {
		mergeTx.MergeTransactionHexs = []string{}
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, answers[array[i][0]-1])
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, answers[array[i][1]-1])
		mergeTx.MergeTransactionHexs = append(mergeTx.MergeTransactionHexs, answers[array[i][2]-1]) //mod
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
		return result.TransactionHex, result.TransactionHash, nil
	}
	return "", "", errors.New("Not complete")
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
		err := stub.PutState(symbolsSpent+selUnspnds[i].MultiAddr+sep+selUnspnds[i].Txid+sep+strconv.Itoa(int(selUnspnds[i].Vout)),
			Int64ToBytes(selUnspnds[i].Value))
		if err != nil {
			log.Debugf("PutState txhash spent failed err: %s", err.Error())
			return errors.New("PutState txhash spent failed")
		}
	}

	if totalAmount > btcTokenAmount {
		err := stub.PutState(symbolsUnspend+selUnspnds[0].MultiAddr+sep+txHash+sep+strconv.Itoa(1), Int64ToBytes(totalAmount-btcTokenAmount))
		if err != nil {
			log.Debugf("PutState txhash unspend failed err: %s", err.Error())
			return errors.New("PutState txhash unspend failed")
		}
	}

	return nil
}

func deleteUtxos(selUnspnds []Unspend, stub shim.ChaincodeStubInterface) error {
	for i := range selUnspnds {
		err := stub.DelState(symbolsUnspend + selUnspnds[i].MultiAddr + sep + selUnspnds[i].Txid + sep + strconv.Itoa(int(selUnspnds[i].Vout)))
		if err != nil {
			log.Debugf("DelState txhash unspend failed err: %s", err.Error())
			return errors.New("DelState txhash unspend failed")
		}
	}

	return nil
}

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

type WithdrawPrepare struct {
	Unspends  []Unspend
	BtcAmount uint64
	RawTX     string
}

func _withdrawPrepare(args []string, stub shim.ChaincodeStubInterface) pb.Response {
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
		return shim.Error(jsonResp)
	}

	btcTokenAmount := uint64(0)
	log.Debugf("contractAddr %s", contractAddr)
	for i := 0; i < len(invokeTokens); i++ {
		log.Debugf("invokeTokens[i].Address %s", invokeTokens[i].Address)
		if invokeTokens[i].Address == contractAddr {
			if invokeTokens[i].Asset.AssetId == btcTokenAsset.AssetId {
				btcTokenAmount += invokeTokens[i].Amount
			}
		}
	}
	if btcTokenAmount == 0 {
		log.Debugf("You need send contractAddr btcToken")
		jsonResp := "{\"Error\":\"You need send contractAddr btcToken\"}"
		return shim.Error(jsonResp)
	}

	// 取未花费
	selUnspnds := getUnspends(int64(btcTokenAmount), stub)
	if len(selUnspnds) == 0 {
		jsonResp := "{\"Error\":\"getUnspends failed\"}"
		return shim.Error(jsonResp)
	}
	// 产生交易
	rawTx, err := genRawTx(int64(btcTokenAmount), btcFee, btcAddr, selUnspnds, stub)
	if err != nil {
		return shim.Error("genRawTx failed: " + err.Error())
	}
	log.Debugf("rawTx:%s", rawTx)

	tempHash := crypto.Keccak256([]byte(rawTx), []byte("prepare"))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)

	//协商交易
	recvResult, err := consult(stub, []byte(tempHashHex), []byte("rawTx"))
	if err != nil {
		return shim.Error("consult rawTx failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal rawTxSign result failed: " + err.Error())
		return shim.Error("Unmarshal rawTxSign result failed: " + err.Error())
	}
	if len(juryMsg) < consultM { //mod
		log.Debugf("RecvJury rawTxSign result's len not enough")
		return shim.Error("RecvJury rawTxSign result's len not enough")
	}

	// 记录Prepare
	var prepare WithdrawPrepare
	prepare.Unspends = selUnspnds
	prepare.BtcAmount = btcTokenAmount
	prepare.RawTX = rawTx
	prepareByte, err := json.Marshal(prepare)
	if err != nil {
		log.Debugf("Marshal selUnspnds failed: " + err.Error())
		return shim.Error("Marshal selUnspnds failed: " + err.Error())
	}
	err = stub.PutState(symbolsWithdrawPrepare+stub.GetTxID(), prepareByte)
	if err != nil {
		log.Debugf("save tx failed: " + err.Error())
		return shim.Error("save tx failed: " + err.Error())
	}
	// 删除UTXO，防止别的提现花费
	err = deleteUtxos(selUnspnds, stub)
	if err != nil {
		log.Debugf("deleteUtxos failed: " + err.Error())
		return shim.Error("deleteUtxos failed: " + err.Error())
	}

	return shim.Success([]byte("Withdraw is ready, please invoke withdrawBTC"))
}

func _withdrawBTC(args []string, stub shim.ChaincodeStubInterface) pb.Response {
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
	_ = json.Unmarshal(result, &prepare)
	if len(prepare.Unspends) == 0 {
		jsonResp := "check Unspends failed"
		return shim.Error(jsonResp)
	}
	if "" == prepare.RawTX {
		jsonResp := "check RawTx failed"
		return shim.Error(jsonResp)
	}
	log.Debugf("rawTx:%s", prepare.RawTX) //todo check utxo spent

	//
	inputRedeemIndex := []int{}
	redeemHex := []string{}
	mapRedeem := make(map[string]int)
	index := 0
	for i := range prepare.Unspends {
		if _, exist := mapRedeem[prepare.Unspends[i].MultiAddr]; exist {
			inputRedeemIndex = append(inputRedeemIndex, mapRedeem[prepare.Unspends[i].MultiAddr])
			continue
		}
		result, _ := stub.GetState(symbolsRedeem + prepare.Unspends[i].MultiAddr)
		if len(result) == 0 {
			return shim.Error("DepsoitRedeem is empty")
		}
		redeemHex = append(redeemHex, string(result))
		mapRedeem[prepare.Unspends[i].MultiAddr] = index
		inputRedeemIndex = append(inputRedeemIndex, index)
		index++
	}
	// 签名交易
	rawTxSign, err := signTx(prepare.RawTX, inputRedeemIndex, redeemHex, stub)
	if err != nil {
		return shim.Error("signTx failed: " + err.Error())
	}
	log.Debugf("rawTxSign:%s", rawTxSign)

	tempHash := crypto.Keccak256([]byte(prepare.RawTX))
	tempHashHex := fmt.Sprintf("%x", tempHash)
	log.Debugf("tempHashHex:%s", tempHashHex)

	//协商交易
	recvResult, err := consult(stub, []byte(tempHashHex), []byte(rawTxSign))
	if err != nil {
		return shim.Error("consult rawTxSign failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		log.Debugf("Unmarshal rawTxSign result failed: " + err.Error())
		return shim.Error("Unmarshal rawTxSign result failed: " + err.Error())
	}
	//stub.PutState("recvResult", recvResult)
	if len(juryMsg) < consultM { //mod
		log.Debugf("RecvJury rawTxSign result's len not enough")
		return shim.Error("RecvJury rawTxSign result's len not enough")
	}

	// 合并交易
	tx, txHash, err := mergeTx(prepare.RawTX, inputRedeemIndex, redeemHex, juryMsg, stub)
	if err != nil {
		log.Debugf("mergeTx failed:  %s", err.Error())
		return shim.Error("mergeTx failed: " + err.Error())
	}

	log.Debugf("start consult txHash %s", txHash)
	//协商 发送交易哈希
	txResult, err := consult(stub, []byte(txHash), []byte("txhash"))
	if err != nil {
		log.Debugf("consult txhash failed: " + err.Error())
		return shim.Error("consult txhash failed: " + err.Error())
	}
	var txJuryMsg []JuryMsgAddr
	err = json.Unmarshal(txResult, &txJuryMsg)
	if err != nil {
		log.Debugf("Unmarshal txhash result failed: " + err.Error())
		return shim.Error("Unmarshal txhash result failed: " + err.Error())
	}
	if len(txJuryMsg) < consultM { //mod
		log.Debugf("RecvJury txhash result's len not enough")
		return shim.Error("RecvJury txhash result's len not enough")
	}
	//协商 保证协商一致后才写入签名结果
	txResult2, err := consult(stub, []byte(txHash+"twice"), []byte("txhash2"))
	if err != nil {
		log.Debugf("consult txhash2 failed: " + err.Error())
		return shim.Error("consult txhash2 failed: " + err.Error())
	}
	var txJuryMsg2 []JuryMsgAddr
	err = json.Unmarshal(txResult2, &txJuryMsg2)
	if err != nil {
		log.Debugf("Unmarshal txhash2 result failed: " + err.Error())
		return shim.Error("Unmarshal txhash2 result failed: " + err.Error())
	}
	if len(txJuryMsg2) < consultM { //mod
		log.Debugf("RecvJury txhash2 result's len not enough")
		return shim.Error("RecvJury txhash2 result's len not enough")
	}

	//记录交易
	err = stub.PutState(symbolsWithdraw+stub.GetTxID(), []byte(tx))
	if err != nil {
		log.Debugf("save tx failed: " + err.Error())
		return shim.Error("save tx failed: " + err.Error())
	}
	// 记录花费
	err = saveUtxos(int64(prepare.BtcAmount), prepare.Unspends, txHash, stub)
	if err != nil {
		log.Debugf("saveUtxos failed: " + err.Error())
		return shim.Error("saveUtxos failed: " + err.Error())
	}
	//删除Prepare
	err = stub.DelState(symbolsWithdrawPrepare + reqid)
	if err != nil {
		log.Debugf("delete WithdrawPrepare failed: " + err.Error())
		return shim.Error("delete WithdrawPrepare failed: " + err.Error())
	}

	////发送交易
	//minSig := string(juryMsg[0].Answer)
	//for i := 1; i < len(juryMsg); i++ {
	//	if strings.Compare(minSig, string(juryMsg[i].Answer)) > 0 {
	//		minSig = string(juryMsg[i].Answer)
	//	}
	//}
	//if strings.Compare(minSig, rawTxSign) == 0 { //自己是执行jury
	//	// 发送交易
	//	txHash, err = sendTx(tx, stub)
	//	if err != nil {
	//		return shim.Error("sendTx failed: " + err.Error())
	//	}
	//} else { //自己不是执行jury
	//	time.Sleep(2 * time.Second)
	//}

	return shim.Success([]byte(txHash))
}

func send(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) == 0 {
		return shim.Error("need 1 args (reqid)")
	}

	reqid := args[0]
	if "0x" != reqid[0:2] {
		reqid = "0x" + reqid
	}

	//查询交易
	result, _ := stub.GetState(symbolsWithdraw + reqid)
	if len(result) == 0 {
		return shim.Error("No withdraw")
	}
	tx := string(result)
	// 发送交易
	txHash, err := sendTx(tx, stub)
	if err != nil {
		return shim.Error("sendTx failed: " + err.Error())
	}
	return shim.Success([]byte(txHash))
}

func get(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) > 0 {
		result, _ := stub.GetState(args[0])
		return shim.Success(result) //test
	}
	result, _ := stub.GetState("result")
	return shim.Success(result)
}

func getAsset(stub shim.ChaincodeStubInterface) pb.Response {
	asset := getBTCTokenAsset(stub)
	return shim.Success([]byte(asset.String()))
}

func main() {
	err := shim.Start(new(BTCPort))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
