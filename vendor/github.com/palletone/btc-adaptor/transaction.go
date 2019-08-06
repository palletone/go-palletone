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
package adaptorbtc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"

	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	//"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/palletone/btc-adaptor/txscript"

	"github.com/palletone/adaptor"
)

func RawTransactionGen(rawTransactionGenParams *adaptor.RawTransactionGenParams, netID int) (string, error) {
	msgTx := wire.NewMsgTx(1)
	//transaction inputs
	for _, inputOne := range rawTransactionGenParams.Inputs {
		hash, err := chainhash.NewHashFromStr(inputOne.Txid)
		if err != nil {
			continue
		}
		input := &wire.TxIn{PreviousOutPoint: wire.OutPoint{*hash, inputOne.Vout}}
		msgTx.AddTxIn(input)
	}
	if len(msgTx.TxIn) == 0 {
		return "", errors.New("Params error : NO Input.")
	}

	//chainnet
	realNet := GetNet(netID)

	//transaction outputs
	for _, outOne := range rawTransactionGenParams.Outputs {
		if len(outOne.Address) == 0 || outOne.Amount <= 0 {
			continue
		}
		addr, err := btcutil.DecodeAddress(outOne.Address, realNet)
		if err != nil {
			return "", err
		}
		pkScript, _ := txscript.PayToAddrScript(addr)
		txOut := wire.NewTxOut(int64(outOne.Amount*1e8), pkScript)
		msgTx.AddTxOut(txOut)
	}
	if len(msgTx.TxOut) == 0 {
		return "", errors.New("Params error : NO Output.")
	}

	//SerializeSize transaction to bytes
	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	if err := msgTx.Serialize(buf); err != nil {
		return "", err
	}
	//result for return
	var rawTransactionGenResult adaptor.RawTransactionGenResult
	rawTransactionGenResult.Rawtx = hex.EncodeToString(buf.Bytes())

	jsonResult, err := json.Marshal(rawTransactionGenResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func DecodeRawTransaction(decodeRawTransactionParams *adaptor.DecodeRawTransactionParams, netID int) (string, error) {
	if "" == decodeRawTransactionParams.Rawtx {
		return "", errors.New("Params error : NO Rawtx.")
	}

	//covert rawtransaction hexString to bytes
	rawTXBytes, err := hex.DecodeString(decodeRawTransactionParams.Rawtx)
	if err != nil {
		return "", err
	}

	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(rawTXBytes))
	if err != nil {
		return "", err
	}

	realNet := GetNet(netID)

	//result for return
	var result adaptor.DecodeRawTransactionResult
	result.Locktime = mtx.LockTime
	for i, _ := range mtx.TxIn {
		result.Inputs = append(result.Inputs, adaptor.Input{Txid: mtx.TxIn[i].PreviousOutPoint.Hash.String(),
			Vout: mtx.TxIn[i].PreviousOutPoint.Index}) //todo Addr
	}
	for i, _ := range mtx.TxOut {
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(mtx.TxOut[i].PkScript, realNet)

		result.Outputs = append(result.Outputs, adaptor.Output{addrs[0].EncodeAddress(), btcutil.Amount(mtx.TxOut[i].Value).ToBTC()})
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func GetTransactionByHash(getTransactionByHashParams *adaptor.GetTransactionByHashParams, rpcParams *RPCParams) (string, error) {
	//covert TxHash
	hash, err := chainhash.NewHashFromStr(getTransactionByHashParams.TxHash)
	if err != nil {
		return "", err
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return "", err
	}
	defer client.Shutdown()

	//rpc GetRawTransactionVerbose
	txResult, err := client.GetRawTransactionVerbose(hash)
	if err != nil {
		return "", err
	}

	//result for return
	var getTransactionByHashResult adaptor.GetTransactionByHashResult
	for _, out := range txResult.Vout {
		getTransactionByHashResult.Outputs = append(getTransactionByHashResult.Outputs,
			adaptor.OutputIndex{out.N, out.ScriptPubKey.Addresses[0], out.Value})
	}
	for _, in := range txResult.Vin {
		getTransactionByHashResult.Inputs = append(getTransactionByHashResult.Inputs,
			adaptor.Input{Txid: in.Txid, Vout: in.Vout}) //todo Addr
	}
	getTransactionByHashResult.Txid = txResult.Txid
	getTransactionByHashResult.Confirms = txResult.Confirmations

	jsonResult, err := json.Marshal(getTransactionByHashResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
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

const base = "https://chain.so/api/v2/"

type GetTransactionHttpResponse struct {
	//Status string `json:"status"`
	Data struct {
		//Network       string `json:"network"`
		Txid string `json:"txid"`
		//Blockhash     string `json:"blockhash"`
		Confirmations int `json:"confirmations"`
		//Time          int    `json:"time"`
		Inputs []struct {
			//InputNo    int         `json:"input_no"`
			Value   string `json:"value"`
			Address string `json:"address"`
			//Type       string      `json:"type"`
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
			//Type     string `json:"type"`
			//Script   string `json:"script"`
		} `json:"outputs"`
		//TxHex    string `json:"tx_hex"`
		//Size     int    `json:"size"`
		//Version  int    `json:"version"`
		Locktime int `json:"locktime"`
	} `json:"data"`
}

func GetTransactionHttp(getTransactionByHashParams *adaptor.GetTransactionHttpParams, netID int) (string, error) {
	if "" == getTransactionByHashParams.TxHash {
		return "", errors.New("TxHash is empty")
	}
	var request string
	if netID == NETID_MAIN {
		request = base + "get_tx/BTC/"
	} else {
		request = base + "get_tx/BTCTEST/"
	}
	//
	strRespose, err, _ := httpGet(request + getTransactionByHashParams.TxHash)
	if err != nil {
		return "", err
	}

	var txResult GetTransactionHttpResponse
	err = json.Unmarshal([]byte(strRespose), &txResult)
	if err != nil {
		return "", err
	}

	//result for return
	var getTransactionByHashResult adaptor.GetTransactionHttpResult
	for _, out := range txResult.Data.Outputs {
		value, _ := strconv.ParseFloat(out.Value, 64)
		getTransactionByHashResult.Outputs = append(getTransactionByHashResult.Outputs,
			adaptor.OutputIndex{uint32(out.OutputNo), out.Address, value})
	}
	for _, in := range txResult.Data.Inputs {
		getTransactionByHashResult.Inputs = append(getTransactionByHashResult.Inputs,
			adaptor.Input{in.FromOutput.Txid, uint32(in.FromOutput.OutputNo), in.Address})
	}
	getTransactionByHashResult.Txid = txResult.Data.Txid
	getTransactionByHashResult.Confirms = uint64(txResult.Data.Confirmations)

	jsonResult, err := json.Marshal(getTransactionByHashResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}
