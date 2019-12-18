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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package btcadaptor

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/palletone/btc-adaptor/txscript"

	"github.com/palletone/adaptor"
)

type outputIndexValue struct {
	OutputIndex string
	Value       uint64
}

// A slice of outputIndexValue that implements sort.Interface to sort by Value.
type outputIndexValueList []outputIndexValue

func (p outputIndexValueList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p outputIndexValueList) Len() int           { return len(p) }
func (p outputIndexValueList) Less(i, j int) bool { return p[i].Value > p[j].Value }

// A function to turn a map into a outputIndexValueList, then sort and return it.
func sortByValue(tpl outputIndexValueList) outputIndexValueList {
	sort.Stable(tpl) //sort.Sort(tpl)
	return tpl
}

func selUnspends(outputIndexMap map[string]float64, btcAmout uint64) []outputIndexValue {
	var smlUnspends []outputIndexValue
	var bigUnspends []outputIndexValue
	var selUnspends []outputIndexValue
	for outputIndex, value := range outputIndexMap {
		amount := uint64(value * 1e8)
		if amount == btcAmout {
			selUnspends = append(selUnspends, outputIndexValue{outputIndex, amount})
			break
		} else if amount > btcAmout {
			bigUnspends = append(bigUnspends, outputIndexValue{outputIndex, amount})
		} else {
			smlUnspends = append(smlUnspends, outputIndexValue{outputIndex, amount})
		}
	}
	//
	if len(selUnspends) != 0 {
		return selUnspends
	}
	//
	selAmount := uint64(0)
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
	selUnspends = []outputIndexValue{}
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
func CreateTransferTokenTx(input *adaptor.CreateTransferTokenTxInput, rpcParams *RPCParams, netID int) (*adaptor.CreateTransferTokenTxOutput, error) {
	//chainnet
	realNet := GetNet(netID)

	//convert address from string
	addr, err := btcutil.DecodeAddress(input.FromAddress, realNet)
	if err != nil {
		return nil, fmt.Errorf("DecodeAddress FromAddress failed %s", err.Error())
	}
	if len(input.Extra)%33 != 0 {
		return nil, fmt.Errorf("input.Extra len invalid, txid:22+index:1")
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//check amount
	fee := +input.Fee.Amount.Uint64()
	if 0 == fee {
		return nil, fmt.Errorf("input.Fee invalid, must not be zero")
	}
	amount := input.Amount.Amount.Uint64()
	btcAmount := amount + fee

	//1.get all unspend
	outputIndexMap, err := getAllUnspend(client, addr)
	if err != nil {
		return nil, err
	}
	if len(outputIndexMap) == 0 {
		return nil, fmt.Errorf("getAllUnspend failed : no utxos")
	}
	//for outputIndex, value := range outputIndexMap {
	//	fmt.Println(outputIndex, value)
	//}

	//2.remove extra utxo
	for i := 0; i < len(input.Extra); i += 33 {
		idIndexHex := hex.EncodeToString(input.Extra[i:33])
		if _, exist := outputIndexMap[idIndexHex]; exist {
			delete(outputIndexMap, idIndexHex)
		}
	}

	//3.select greet
	outputIndexSel := selUnspends(outputIndexMap, btcAmount)
	if len(outputIndexSel) == 0 {
		return nil, fmt.Errorf("selUnspends failed : balance is not enough")
	}
	//for _, out := range outputIndexSel { //Debug
	//	fmt.Println(out.OutputIndex, out.Value)
	//}

	msgTx := wire.NewMsgTx(1)
	//transaction inputs
	allInputAmount := uint64(0)
	extra := []byte{}
	for _, outputIndexV := range outputIndexSel {
		//fmt.Println(outputIndexV.OutputIndex, outputIndexV.Value)
		voutByte, _ := hex.DecodeString(outputIndexV.OutputIndex[64:66])
		vout := uint64(voutByte[0])
		hash, err := chainhash.NewHashFromStr(outputIndexV.OutputIndex[0:64])
		if err != nil {
			return nil, fmt.Errorf("NewHashFromStr outputIndexSel failed")
		}
		input := &wire.TxIn{PreviousOutPoint: wire.OutPoint{*hash, uint32(vout)}}
		msgTx.AddTxIn(input)
		allInputAmount += outputIndexV.Value
		outputIndexByte, _ := hex.DecodeString(outputIndexV.OutputIndex)
		extra = append(extra, outputIndexByte...)
	}
	if len(msgTx.TxIn) == 0 {
		return nil, fmt.Errorf("Process TxIn error : NO Input.")
	}

	//transaction outputs
	addrTo, err := btcutil.DecodeAddress(input.ToAddress, realNet)
	if err != nil {
		amount = 0 //op_return set amount 0
		pkScript, _ := txscript.NullDataScript([]byte(input.ToAddress))
		txOut := wire.NewTxOut(int64(0), pkScript)
		msgTx.AddTxOut(txOut)
		//return nil, fmt.Errorf("DecodeAddress ToAddress failed %s", err.Error())
	} else {
		pkScript, _ := txscript.PayToAddrScript(addrTo)
		txOut := wire.NewTxOut(int64(amount), pkScript)
		msgTx.AddTxOut(txOut)
	}

	//change
	change := allInputAmount - amount - fee
	//fmt.Println(change, allInputAmount, btcAmount, fee) //Debug
	if change > 0 {
		pkScript, _ := txscript.PayToAddrScript(addr)
		txOut := wire.NewTxOut(int64(change), pkScript)
		msgTx.AddTxOut(txOut)
	}
	if len(msgTx.TxOut) == 0 {
		return nil, fmt.Errorf("Process TxOut error : NO Output.")
	}
	//for _, out := range msgTx.TxOut { //Debug
	//	fmt.Println(out.Value)
	//}

	//SerializeSize transaction to bytes
	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	if err := msgTx.Serialize(buf); err != nil {
		return nil, err
	}
	//result for return
	var output adaptor.CreateTransferTokenTxOutput
	output.Transaction = buf.Bytes()
	output.Extra = extra

	return &output, nil
}

func decodeRawTransaction(rawTxHex string, netID int) (interface{}, error) {
	//covert rawtransaction hexString to bytes
	rawTXBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return nil, err
	}

	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(rawTXBytes))
	if err != nil {
		return nil, err
	}

	realNet := GetNet(netID)

	//
	fmt.Println("Input : ")
	for i := range mtx.TxIn {
		fmt.Println(mtx.TxIn[i].PreviousOutPoint.Hash.String(), mtx.TxIn[i].PreviousOutPoint.Index)
	}
	fmt.Println("Output : ")
	for i := range mtx.TxOut {
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(mtx.TxOut[i].PkScript, realNet)
		if 0 != len(addrs) {
			fmt.Println(addrs[0].EncodeAddress(), btcutil.Amount(mtx.TxOut[i].Value).ToBTC())
		} else {
			disbuf, _ := txscript.DisasmString(mtx.TxOut[i].PkScript)
			dataHex := disbuf[len("OP_RETURN "):]
			data, _ := hex.DecodeString(dataHex)
			fmt.Println("data : ", string(data))
		}
	}

	return &mtx, nil
}

func CalcTxHash(input *adaptor.CalcTxHashInput) (*adaptor.CalcTxHashOutput, error) {
	//deserialize to MsgTx
	var tx wire.MsgTx
	err := tx.Deserialize(bytes.NewReader(input.Transaction))
	if err != nil {
		return nil, fmt.Errorf("Deserialize tx failed : %s", err.Error())
	}

	//result for return
	var output adaptor.CalcTxHashOutput
	txHashHex := tx.TxHash().String()
	txHashByte, _ := hex.DecodeString(txHashHex)
	output.Hash = txHashByte

	return &output, nil
}

func GetBlockInfo(input *adaptor.GetBlockInfoInput, rpcParams *RPCParams) (*adaptor.GetBlockInfoOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//
	var blkHash *chainhash.Hash
	if input.Latest {
		blkHash, _, err = client.GetBestBlock() //BTCD API
		if err != nil {
			return nil, fmt.Errorf("GetBestBlock Latest failed : %s", err.Error())
		}
	} else if len(input.BlockID) != 0 {
		blkHash, err = chainhash.NewHashFromStr(hex.EncodeToString(input.BlockID))
		if err != nil {
			return nil, fmt.Errorf("NewHashFromStr BlockID failed : %s", err.Error())
		}
	} else {
		blkHash, err = client.GetBlockHash(int64(input.Height)) //BTCD API
		if err != nil {
			return nil, fmt.Errorf("GetBlockHash Height failed : %s", err.Error())
		}
	}

	blkResult, err := client.GetBlockVerbose(blkHash) //BTCD API
	if err != nil {
		return nil, fmt.Errorf("GetBlockVerbose failed : %s", err.Error())
	}
	blkHeader, err := client.GetBlockHeader(blkHash) //BTCD API
	buf := bytes.NewBuffer(make([]byte, 0, 80))
	if err := blkHeader.Serialize(buf); err != nil {
		return nil, fmt.Errorf("Serialize blkHeader failed : %s", err.Error())
	}

	//result for return
	var output adaptor.GetBlockInfoOutput
	blockID, _ := hex.DecodeString(blkResult.Hash)
	output.Block.BlockID = blockID
	output.Block.BlockHeight = uint(blkResult.Height) //GetBlockVerbose
	output.Block.Timestamp = uint64(blkResult.Time)
	blockIDPre, _ := hex.DecodeString(blkResult.PreviousHash)
	output.Block.ParentBlockID = blockIDPre

	output.Block.HeaderRawData = buf.Bytes()
	merkleRoot, _ := hex.DecodeString(blkResult.MerkleRoot)
	output.Block.TxsRoot = merkleRoot

	if len(blkResult.RawTx) > 0 {
		txCoinBase := blkResult.RawTx[0]
		if 0 != len(txCoinBase.Vout[0].ScriptPubKey.Addresses) {
			output.Block.ProducerAddress = txCoinBase.Vout[0].ScriptPubKey.Addresses[0]
		}
	} else if len(blkResult.Tx) > 0 {
		hash, err := chainhash.NewHashFromStr(blkResult.Tx[0])
		if err != nil {
			return nil, fmt.Errorf("NewHashFromStr tx failed : %s", err.Error())
		}
		txResult, err := client.GetRawTransactionVerbose(hash) //BTCD API
		if 0 != len(txResult.Vout[0].ScriptPubKey.Addresses) {
			output.Block.ProducerAddress = txResult.Vout[0].ScriptPubKey.Addresses[0]
		}
	}

	if blkResult.Confirmations >= MinConfirm { //GetBlockVerbose
		output.Block.IsStable = true
	} else {
		output.Block.IsStable = false
	}

	return &output, nil
}

func GetPalletOneMappingAddress(input *adaptor.GetPalletOneMappingAddressInput, rpcParams *RPCParams) (*adaptor.GetPalletOneMappingAddressOutput, error) {
	//covert TxHash
	hash, err := chainhash.NewHashFromStr(input.MappingDataSource)
	if err != nil {
		return nil, fmt.Errorf("NewHashFromStr MappingDataSource failed : %s", err.Error())
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//rpc GetRawTransactionVerbose
	txResult, err := client.GetRawTransactionVerbose(hash) //BTCD API
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerbose tx failed : %s", err.Error())
	}

	//result for return
	var output adaptor.GetPalletOneMappingAddressOutput
	//get from address
	hashPre, err := chainhash.NewHashFromStr(txResult.Vin[0].Txid)
	if err != nil {
		return nil, fmt.Errorf("NewHashFromStr txPre failed : %s", err.Error())
	}
	txPreResult, err := client.GetRawTransactionVerbose(hashPre) //BTCD API
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerbose txPre 0 failed : %s", err.Error())
	}
	fromAddr := txPreResult.Vout[txResult.Vin[0].Vout].ScriptPubKey.Addresses[0]
	if input.ChainAddress == "" {
		output.ChainAddress = fromAddr
	} else if fromAddr != input.ChainAddress {
		return nil, fmt.Errorf("the ChainAddress is not match")
	}

	//get op_return data
	exist := false
	for _, out := range txResult.Vout {
		if out.ScriptPubKey.Type == "nulldata" { //todo: more op_return ?
			if strings.HasPrefix(out.ScriptPubKey.Asm, "OP_RETURN") {
				data, _ := hex.DecodeString(out.ScriptPubKey.Asm[len("OP_RETURN "):])
				output.PalletOneAddress = string(data)
				exist = true
				break
			}
		}
	}
	if !exist {
		return nil, fmt.Errorf("the PalletOneAddress not exist in MappingDataSource")
	}

	return &output, nil
}

func GetTxBasicInfo(input *adaptor.GetTxBasicInfoInput, rpcParams *RPCParams) (*adaptor.GetTxBasicInfoOutput, error) {
	//covert TxHash
	hash, err := chainhash.NewHashFromStr(hex.EncodeToString(input.TxID))
	if err != nil {
		return nil, fmt.Errorf("NewHashFromStr tx failed : %s", err.Error())
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//rpc GetRawTransactionVerbose
	txResult, err := client.GetRawTransactionVerbose(hash) //BTCD API
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerbose tx failed : %s", err.Error())
	}

	//result for return
	var output adaptor.GetTxBasicInfoOutput
	//get from address
	hashPre, err := chainhash.NewHashFromStr(txResult.Vin[0].Txid)
	if err != nil {
		return nil, fmt.Errorf("NewHashFromStr hashPre failed : %s", err.Error())
	}
	txPreResult, err := client.GetRawTransactionVerbose(hashPre) //BTCD API
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerbose txPre 0 failed : %s", err.Error())
	}
	fromAddr := txPreResult.Vout[txResult.Vin[0].Vout].ScriptPubKey.Addresses[0]

	//get to address
	toAddr := ""
	for _, out := range txResult.Vout {
		if len(out.ScriptPubKey.Addresses) == 0 {
			continue
		}
		if fromAddr == out.ScriptPubKey.Addresses[0] {
			continue
		}
		if toAddr != "" && toAddr != out.ScriptPubKey.Addresses[0] {
			return nil, fmt.Errorf("Not support send 2+ tx ")
		}
		toAddr = out.ScriptPubKey.Addresses[0]
		break
	}

	output.Tx.TxID, _ = hex.DecodeString(txResult.Txid)
	txRaw, _ := hex.DecodeString(txResult.Hex)
	output.Tx.TxRawData = txRaw
	output.Tx.CreatorAddress = fromAddr
	output.Tx.TargetAddress = toAddr
	if txResult.BlockHash != "" { //GetRawTransactionVerbose
		output.Tx.IsInBlock = true
		output.Tx.IsSuccess = true
		blockID, _ := hex.DecodeString(txResult.BlockHash)
		output.Tx.BlockID = blockID
		blkHash, err := chainhash.NewHashFromStr(txResult.BlockHash)
		if err == nil {
			blkResult, err := client.GetBlockVerbose(blkHash) //BTCD API
			if err == nil {
				output.Tx.BlockHeight = uint(blkResult.Height)
			}
		}
	} else {
		output.Tx.IsInBlock = false
		output.Tx.IsSuccess = false
	}
	if txResult.Confirmations >= MinConfirm {
		output.Tx.IsStable = true
	} else {
		output.Tx.IsStable = false
	}
	output.Tx.TxIndex = 0 //todo
	output.Tx.Timestamp = uint64(txResult.Blocktime)

	return &output, nil
}

func GetTransferTx(input *adaptor.GetTransferTxInput, rpcParams *RPCParams) (*adaptor.GetTransferTxOutput, error) {
	//covert TxHash
	hash, err := chainhash.NewHashFromStr(hex.EncodeToString(input.TxID))
	//hash, err := chainhash.NewHash(input.TxID)//hash.String() is not same
	if err != nil {
		return nil, fmt.Errorf("NewHashFromStr tx failed : %s", err.Error())
	}
	//fmt.Println(hash.String())
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//rpc GetRawTransactionVerbose
	txResult, err := client.GetRawTransactionVerbose(hash) //BTCD API
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerbose tx failed : %s", err.Error())
	}

	//result for return
	var output adaptor.GetTransferTxOutput
	//get from address
	hashPre, err := chainhash.NewHashFromStr(txResult.Vin[0].Txid)
	if err != nil {
		return nil, fmt.Errorf("NewHashFromStr hashPre failed : %s", err.Error())
	}
	txPreResult, err := client.GetRawTransactionVerbose(hashPre) //BTCD API
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerbose txPre 0 failed : %s", err.Error())
	}
	fromAddr := txPreResult.Vout[txResult.Vin[0].Vout].ScriptPubKey.Addresses[0]
	output.Tx.FromAddress = fromAddr

	//get input amount
	inputAmount := txPreResult.Vout[txResult.Vin[0].Vout].Value
	for i := 1; i < len(txResult.Vin); i++ {
		hashPre, err := chainhash.NewHashFromStr(txResult.Vin[i].Txid)
		if err != nil {
			return nil, fmt.Errorf("hashPre failed : %s", err.Error())
		}
		txPreResult, err := client.GetRawTransactionVerbose(hashPre) //BTCD API
		if err != nil {
			return nil, fmt.Errorf("GetRawTransactionVerbose txPre %d failed : %s", i, err.Error())
		}
		inputAmount += txPreResult.Vout[txResult.Vin[i].Vout].Value
	}

	//get to address and amount
	change := float64(0)
	amount := float64(0)
	for _, out := range txResult.Vout {
		if out.ScriptPubKey.Type == "nulldata" { //todo: more op_return ?
			if strings.HasPrefix(out.ScriptPubKey.Asm, "OP_RETURN") {
				data, _ := hex.DecodeString(out.ScriptPubKey.Asm[len("OP_RETURN "):])
				output.Tx.AttachData = data
			}
			continue
		}
		if len(out.ScriptPubKey.Addresses) == 0 {
			continue
		}
		if fromAddr == out.ScriptPubKey.Addresses[0] {
			change = out.Value
			continue
		}
		if output.Tx.ToAddress != "" && output.Tx.ToAddress != out.ScriptPubKey.Addresses[0] {
			return nil, fmt.Errorf("Not support send 2+ tx ")
		}
		output.Tx.ToAddress = out.ScriptPubKey.Addresses[0]
		amount = out.Value
		break
	}
	fee := inputAmount - change - amount

	//turn to big int
	bigIntAmount := new(big.Int)
	bigIntAmount.SetUint64(uint64(amount * 1e8))
	output.Tx.Amount = adaptor.NewAmountAsset(bigIntAmount, "BTC")
	bigIntFee := new(big.Int)
	bigIntFee.SetUint64(uint64(fee * 1e8))
	output.Tx.Fee = adaptor.NewAmountAsset(bigIntFee, "BTC")

	output.Tx.TxID, _ = hex.DecodeString(txResult.Txid)
	txRaw, _ := hex.DecodeString(txResult.Hex)
	output.Tx.TxRawData = txRaw
	output.Tx.CreatorAddress = fromAddr
	output.Tx.TargetAddress = output.Tx.ToAddress
	if txResult.BlockHash != "" { //GetRawTransactionVerbose
		output.Tx.IsInBlock = true
		output.Tx.IsSuccess = true
		blockID, _ := hex.DecodeString(txResult.BlockHash)
		output.Tx.BlockID = blockID
		blkHash, err := chainhash.NewHashFromStr(txResult.BlockHash)
		if err == nil {
			blkResult, err := client.GetBlockVerbose(blkHash) //BTCD API
			if err == nil {
				output.Tx.BlockHeight = uint(blkResult.Height)
			}
		}
	} else {
		output.Tx.IsInBlock = false
		output.Tx.IsSuccess = false
	}
	if txResult.Confirmations >= MinConfirm {
		output.Tx.IsStable = true
	} else {
		output.Tx.IsStable = false
	}
	output.Tx.TxIndex = 0 //todo
	output.Tx.Timestamp = uint64(txResult.Blocktime)

	return &output, nil
}

func httpGet(url string) (string, error, int) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err, 0
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err, 0
	}

	return string(body), nil, resp.StatusCode
}

func httpPost(url string, params string) (string, error, int) {
	resp, err := http.Post(url, "application/json", strings.NewReader(params))
	if err != nil {
		return "", err, 0
	}
	defer resp.Body.Close()

	//fmt.Println(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err, 0
	}

	return string(body), nil, resp.StatusCode
}

const base = "https://block.io/api/v2/"

type GetTransactionHttpResponse struct {
	//Status string `json:"status"`
	Data struct {
		//Network       string `json:"network"`
		Txid          string `json:"txid"`
		Blockhash     string `json:"blockhash"`
		Confirmations int    `json:"confirmations"`
		Time          int    `json:"time"`
		Inputs        []struct {
			//InputNo    int         `json:"input_no"`
			Value   string `json:"value"`
			Address string `json:"address"`
			Type    string `json:"type"`
			//Script     string      `json:"script"`
			//Witness    interface{} `json:"witness"`
			FromOutput struct {
				Txid     string `json:"txid"`
				OutputNo int    `json:"output_no"`
			} `json:"from_output"`
		} `json:"inputs"`
		Outputs []struct {
			OutputNo int    `json:"output_no"`
			Value    string `json:"value"`
			Address  string `json:"address"`
			Type     string `json:"type"`
			Script   string `json:"script"`
		} `json:"outputs"`
		TxHex string `json:"tx_hex"`
		//Size     int    `json:"size"`
		//Version  int    `json:"version"`
		Locktime int `json:"locktime"`
	} `json:"data"`
}

func GetTxBasicInfoHttp(input *adaptor.GetTxBasicInfoInput, netID int) (*adaptor.GetTxBasicInfoOutput, error) {
	txHash := hex.EncodeToString(input.TxID)
	if "" == txHash {
		return nil, fmt.Errorf("TxHash is empty")
	}

	var request string
	//if netID == NETID_MAIN {
	//	request = base + "get_tx/BTC/"
	//} else {
	//	request = base + "get_tx/BTCTEST/"
	//}
	if netID == NETID_MAIN {
		request = base + "get_raw_transaction/?api_key=53ef-d8e9-1414-589f&txid="
	} else {
		request = base + "get_raw_transaction/?api_key=f13a-e0e6-614f-bbf8&txid="
	}
	//
	strRespose, err, _ := httpGet(request + txHash)
	if err != nil {
		return nil, err
	}

	var txResult GetTransactionHttpResponse
	err = json.Unmarshal([]byte(strRespose), &txResult)
	if err != nil {
		return nil, err
	}

	//result for return
	var output adaptor.GetTxBasicInfoOutput
	//get from address
	fromAddr := txResult.Data.Inputs[0].Address

	//get to address
	toAddr := ""
	for _, out := range txResult.Data.Outputs {
		if "" == out.Address {
			continue
		}
		if fromAddr == out.Address {
			continue
		}
		if toAddr != "" && toAddr != out.Address {
			return nil, fmt.Errorf("Not support send 2+ tx ")
		}
		toAddr = out.Address
		break
	}

	output.Tx.TxID, _ = hex.DecodeString(txResult.Data.Txid)
	txRaw, _ := hex.DecodeString(txResult.Data.TxHex)
	output.Tx.TxRawData = txRaw
	output.Tx.CreatorAddress = fromAddr
	output.Tx.TargetAddress = toAddr
	if txResult.Data.Blockhash != "" { //GetRawTransactionVerbose
		output.Tx.IsInBlock = true
		output.Tx.IsSuccess = true
		blockID, _ := hex.DecodeString(txResult.Data.Blockhash)
		output.Tx.BlockID = blockID
		output.Tx.BlockHeight = 0 //todo
	} else {
		output.Tx.IsInBlock = false
		output.Tx.IsSuccess = false
	}
	if txResult.Data.Confirmations >= MinConfirm {
		output.Tx.IsStable = true
	} else {
		output.Tx.IsStable = false
	}
	output.Tx.TxIndex = 0 //todo
	output.Tx.Timestamp = uint64(txResult.Data.Time)

	return &output, nil
}
