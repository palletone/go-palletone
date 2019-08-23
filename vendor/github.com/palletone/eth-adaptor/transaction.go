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
package ethadaptor

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/palletone/adaptor"
)

func httpGet(url string) (string, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	return string(body), resp.StatusCode, nil
}

//
//func httpPost(url string, params string) (string, int, error) {
//	resp, err := http.Post(url, "application/json", strings.NewReader(params))
//	if err != nil {
//		return "", 0, err
//	}
//	defer resp.Body.Close()
//
//	//fmt.Println(resp.StatusCode)
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", 0, err
//	}
//
//	return string(body), resp.StatusCode, nil
//}

type Tx struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	TransactionIndex  string `json:"transactionIndex"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	IsError           string `json:"isError"`
	TxreceiptStatus   string `json:"txreceipt_status"`
	Input             string `json:"input"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	GasUsed           string `json:"gasUsed"`
	Confirmations     string `json:"confirmations"`
}
type GetAddrTxHistoryResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []Tx   `json:"result"`
}

//https://api-ropsten.etherscan.io/api?module=account&action=txlist&address=0xddbd2b932c763ba5b1b7ae3b362eac3e8d40121a
// &startblock=0&endblock=99999999&page=1&offset=10&sort=asc&apikey=YourApiKeyToken
func GetAddrTxHistoryHTTP(apiURL string, input *adaptor.GetAddrTxHistoryInput) (*adaptor.GetAddrTxHistoryOutput, error) {
	request := apiURL
	request += "?module=account&action=tokentx&address=" + input.FromAddress + "&startblock=0&endblock=99999999"
	if input.PageIndex != 0 && input.PageSize != 0 {
		request += "&page=" + fmt.Sprintf("%d", input.PageIndex)
		request += "&offset=" + fmt.Sprintf("%d", input.PageSize)
	}
	if input.Asc {
		request += "&sort=asc"
	} else {
		request += "&sort=desc"
	}
	request += "&apikey=YourApiKeyToken"
	fmt.Println(request)
	//
	strRespose, _, err := httpGet(request)
	if err != nil {
		return nil, err
	}

	var txResult GetAddrTxHistoryResult
	err = json.Unmarshal([]byte(strRespose), &txResult)
	if err != nil {
		return nil, err
	}

	//result for return
	var result adaptor.GetAddrTxHistoryOutput
	if input.AddressLogicAndOr {
		for i := range txResult.Result {
			toAddr := strings.ToLower(input.ToAddress)
			if txResult.Result[i].To == toAddr || txResult.Result[i].ContractAddress == toAddr {
				tx := convertSimpleTx(&txResult.Result[i])
				result.Txs = append(result.Txs, tx)
			}
		}
	} else {
		for i := range txResult.Result {
			tx := convertSimpleTx(&txResult.Result[i])
			result.Txs = append(result.Txs, tx)
		}
	}
	result.Count = uint32(len(result.Txs))

	return &result, nil
}
func convertSimpleTx(txResult *Tx) *adaptor.SimpleTransferTokenTx {
	tx := &adaptor.SimpleTransferTokenTx{}
	tx.TxID = common.Hex2Bytes(txResult.Hash[2:])
	if len(txResult.Input) > 2 {
		tx.TxRawData = common.Hex2Bytes(txResult.Input[2:])
	}
	tx.CreatorAddress = txResult.From
	tx.TargetAddress = txResult.To
	tx.IsInBlock = true
	if txResult.IsError == "0" {
		tx.IsSuccess = true
	} else {
		tx.IsSuccess = false
	}
	confirms, _ := strconv.ParseUint(txResult.Confirmations, 10, 64)
	if confirms > 15 {
		tx.IsStable = true
	}
	tx.BlockID = common.Hex2Bytes(txResult.BlockHash[2:])
	blockNum, _ := strconv.ParseUint(txResult.BlockNumber, 10, 64)
	tx.BlockHeight = uint(blockNum)
	index, _ := strconv.ParseUint(txResult.TransactionIndex, 10, 64)
	tx.TxIndex = uint(index)
	timeStamp, _ := strconv.ParseUint(txResult.TimeStamp, 10, 64)
	tx.Timestamp = timeStamp
	tx.Amount = adaptor.NewAmountAssetString(txResult.Value, "ETH")
	//tx.Amount.Amount.SetString(txResult.Value, 10)
	tx.Fee = adaptor.NewAmountAssetString(txResult.GasUsed, "ETH")
	//tx.Fee.Amount.SetString(txResult.GasUsed, 10)
	tx.FromAddress = tx.CreatorAddress
	if txResult.To == "" {
		tx.ToAddress = txResult.ContractAddress
	} else {
		tx.ToAddress = txResult.To
	}
	tx.AttachData = tx.TxRawData //todo

	return tx
}

//https://api-ropsten.etherscan.io/api?module=account&action=tokentx&address=0x588eb98f8814aedb056d549c0bafd5ef4963069c
// &startblock=0&endblock=99999999&sort=desc&apikey=YourApiKeyToken
func GetAddrErc20TxHistoryHTTP(apiURL string, input *adaptor.GetAddrTxHistoryInput) (*adaptor.GetAddrTxHistoryOutput,
	error) {
	request := apiURL
	request += "?module=account&action=tokentx&address=" + input.FromAddress + "&startblock=0&endblock=99999999"
	if input.PageIndex != 0 && input.PageSize != 0 {
		request += "&page=" + fmt.Sprintf("%d", input.PageIndex)
		request += "&offset=" + fmt.Sprintf("%d", input.PageSize)
	}
	if input.Asc {
		request += "&sort=asc"
	} else {
		request += "&sort=desc"
	}
	request += "&apikey=YourApiKeyToken"
	//fmt.Println(request)
	//
	strRespose, _, err := httpGet(request)
	if err != nil {
		return nil, err
	}

	var txResult GetAddrTxHistoryResult
	err = json.Unmarshal([]byte(strRespose), &txResult)
	if err != nil {
		return nil, err
	}

	//result for return
	var result adaptor.GetAddrTxHistoryOutput
	if input.AddressLogicAndOr {
		for i := range txResult.Result {
			toAddr := strings.ToLower(input.ToAddress)
			if len(input.Asset) != 0 && txResult.Result[i].ContractAddress != input.Asset {
				continue
			}
			if txResult.Result[i].To == toAddr || txResult.Result[i].ContractAddress == toAddr {
				tx := convertSimpleErc20Tx(&txResult.Result[i])
				result.Txs = append(result.Txs, tx)
			}
		}
	} else {
		for i := range txResult.Result {
			if len(input.Asset) != 0 && txResult.Result[i].ContractAddress != input.Asset {
				continue
			}
			tx := convertSimpleErc20Tx(&txResult.Result[i])
			result.Txs = append(result.Txs, tx)
		}
	}
	result.Count = uint32(len(result.Txs))

	return &result, nil
}
func convertSimpleErc20Tx(txResult *Tx) *adaptor.SimpleTransferTokenTx {
	tx := &adaptor.SimpleTransferTokenTx{}
	tx.TxID = common.Hex2Bytes(txResult.Hash[2:])
	if len(txResult.Input) > 2 {
		tx.TxRawData = common.Hex2Bytes(txResult.Input[2:])
	}
	tx.CreatorAddress = txResult.From
	tx.TargetAddress = txResult.To
	tx.IsInBlock = true
	tx.IsSuccess = true
	confirms, _ := strconv.ParseUint(txResult.Confirmations, 10, 64)
	if confirms > 15 {
		tx.IsStable = true
	}
	tx.BlockID = common.Hex2Bytes(txResult.BlockHash[2:])
	blockNum, _ := strconv.ParseUint(txResult.BlockNumber, 10, 64)
	tx.BlockHeight = uint(blockNum)
	index, _ := strconv.ParseUint(txResult.TransactionIndex, 10, 64)
	tx.TxIndex = uint(index)
	timeStamp, _ := strconv.ParseUint(txResult.TimeStamp, 10, 64)
	tx.Timestamp = timeStamp
	tx.Amount = adaptor.NewAmountAssetString(txResult.Value, txResult.ContractAddress)
	//tx.Amount = &adaptor.AmountAsset{Asset: txResult.ContractAddress}
	//tx.Amount.Amount.SetString(txResult.Value, 10)
	tx.Fee = adaptor.NewAmountAssetString(txResult.GasUsed, "ETH")
	//tx.Fee = &adaptor.AmountAsset{Asset: "ETH"}
	//tx.Fee.Amount.SetString(txResult.GasUsed, 10)
	tx.FromAddress = tx.CreatorAddress
	if txResult.To == "" {
		tx.ToAddress = txResult.ContractAddress
	} else {
		tx.ToAddress = txResult.To
	}
	//tx.AttachData = tx.TxRawData //

	return tx
}

func GetTxBasicInfo(input *adaptor.GetTxBasicInfoInput, rpcParams *RPCParams, netID int) (
	*adaptor.GetTxBasicInfoOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	//call eth method
	hash := common.BytesToHash(input.TxID)
	tx, blockNumber, blockHash, err := client.TransactionsByHash(context.Background(), hash)
	if err != nil {
		//fmt.Println("0")//pending not found
		return nil, err
	}

	//conver to msg for from address
	bigIntBlockNum := new(big.Int)
	bigIntBlockNum.SetString(blockNumber, 0)

	var signer types.Signer
	if netID == NETID_MAIN {
		signer = types.MakeSigner(params.MainnetChainConfig, bigIntBlockNum)
	} else {
		signer = types.MakeSigner(params.TestnetChainConfig, bigIntBlockNum)
	}

	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.GetTxBasicInfoOutput
	result.Tx.TxID = tx.Hash().Bytes()
	result.Tx.TxRawData = tx.Data()
	result.Tx.CreatorAddress = msg.From().String()
	toAddr := msg.To()
	if toAddr != nil {
		result.Tx.TargetAddress = msg.To().String()
	}
	result.Tx.IsInBlock = true
	if receipt.Status > 0 {
		result.Tx.IsSuccess = true
	} else {
		result.Tx.IsSuccess = false
	}
	result.Tx.IsStable = true //todo delete
	if "0x" == blockHash[:2] || "0X" == blockHash[:2] {
		result.Tx.BlockID = Hex2Bytes(blockHash[2:])
	} else {
		result.Tx.BlockID = Hex2Bytes(blockHash)
	}
	result.Tx.BlockHeight = uint(bigIntBlockNum.Uint64())
	result.Tx.TxIndex = 0   //receipt.Logs[0].TxIndex //todo delete
	result.Tx.Timestamp = 0 //todo delete

	return &result, nil
}

func GetTransferTx(input *adaptor.GetTransferTxInput, rpcParams *RPCParams, netID int) (
	*adaptor.GetTransferTxOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	//call eth method
	hash := common.BytesToHash(input.TxID)
	tx, blockNumber, blockHash, err := client.TransactionsByHash(context.Background(), hash)
	if err != nil {
		//fmt.Println("0")//pending not found
		return nil, err
	}

	//conver to msg for from address
	bigIntBlockNum := new(big.Int)
	bigIntBlockNum.SetString(blockNumber, 0)

	var signer types.Signer
	if netID == NETID_MAIN {
		signer = types.MakeSigner(params.MainnetChainConfig, bigIntBlockNum)
	} else {
		signer = types.MakeSigner(params.TestnetChainConfig, bigIntBlockNum)
	}

	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.GetTransferTxOutput
	result.Tx.TxID = tx.Hash().Bytes()
	result.Tx.TxRawData = tx.Data()
	result.Tx.CreatorAddress = msg.From().String()
	toAddr := msg.To()
	if toAddr != nil {
		result.Tx.TargetAddress = msg.To().String()
	}
	result.Tx.IsInBlock = true
	if receipt.Status > 0 {
		result.Tx.IsSuccess = true
	} else {
		result.Tx.IsSuccess = false
	}
	result.Tx.IsStable = true //todo delete
	if "0x" == blockHash[:2] || "0X" == blockHash[:2] {
		result.Tx.BlockID = Hex2Bytes(blockHash[2:])
	} else {
		result.Tx.BlockID = Hex2Bytes(blockHash)
	}
	result.Tx.BlockHeight = uint(bigIntBlockNum.Uint64())
	result.Tx.TxIndex = 0   //receipt.Logs[0].TxIndex //todo delete
	result.Tx.Timestamp = 0 //todo delete

	if len(receipt.Logs) > 0 && len(receipt.Logs[0].Topics) > 2 {
		result.Tx.FromAddress = common.BytesToAddress(receipt.Logs[0].Topics[1].Bytes()).String()
		result.Tx.ToAddress = common.BytesToAddress(receipt.Logs[0].Topics[2].Bytes()).String()

		//result.Tx.Amount = &adaptor.AmountAsset{}
		//result.Tx.Amount.Amount.SetBytes(receipt.Logs[0].Data)
		amt := new(big.Int)
		amt.SetBytes(receipt.Logs[0].Data)
		result.Tx.Amount = adaptor.NewAmountAsset(amt, "ETH")
	} else {
		result.Tx.FromAddress = result.Tx.CreatorAddress
		receiptAddr := receipt.ContractAddress.String()
		if receiptAddr == "0x0000000000000000000000000000000000000000" {
			result.Tx.ToAddress = result.Tx.TargetAddress
		} else {
			result.Tx.ToAddress = receiptAddr
		}
		result.Tx.Amount = adaptor.NewAmountAsset(msg.Value(), "ETH")
		//result.Tx.Amount.Amount.Set(msg.Value())
	}

	result.Tx.Fee = adaptor.NewAmountAssetUint64(msg.Gas(), "ETH")
	//result.Tx.Fee.Amount.SetUint64(msg.Gas())
	result.Tx.AttachData = msg.Data()

	return &result, nil
}

func GetContractInitialTx(input *adaptor.GetContractInitialTxInput, rpcParams *RPCParams, netID int) (
	*adaptor.GetContractInitialTxOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	//call eth method
	hash := common.BytesToHash(input.TxID)
	tx, blockNumber, blockHash, err := client.TransactionsByHash(context.Background(), hash)
	if err != nil {
		//fmt.Println("0")//pending not found
		return nil, err
	}

	//conver to msg for from address
	bigIntBlockNum := new(big.Int)
	bigIntBlockNum.SetString(blockNumber, 0)

	var signer types.Signer
	if netID == NETID_MAIN {
		signer = types.MakeSigner(params.MainnetChainConfig, bigIntBlockNum)
	} else {
		signer = types.MakeSigner(params.TestnetChainConfig, bigIntBlockNum)
	}

	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.GetContractInitialTxOutput
	result.TxID = tx.Hash().Bytes()
	result.TxRawData = tx.Data()
	result.CreatorAddress = msg.From().String()
	toAddr := msg.To()
	if toAddr != nil {
		result.TargetAddress = msg.To().String()
	}
	result.IsInBlock = true
	if receipt.Status > 0 {
		result.IsSuccess = true
	} else {
		result.IsSuccess = false
	}
	result.IsStable = true //todo delete
	if "0x" == blockHash[:2] || "0X" == blockHash[:2] {
		result.BlockID = Hex2Bytes(blockHash[2:])
	} else {
		result.BlockID = Hex2Bytes(blockHash)
	}
	result.BlockHeight = uint(bigIntBlockNum.Uint64())
	result.TxIndex = 0   //receipt.Logs[0].TxIndex //todo delete
	result.Timestamp = 0 //todo delete
	result.ContractAddress = receipt.ContractAddress.String()

	return &result, nil
}

func GetBlockInfo(input *adaptor.GetBlockInfoInput, rpcParams *RPCParams) (*adaptor.GetBlockInfoOutput,
	error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	//call eth rpc method
	var heder *types.Header
	if input.Latest {
		heder, err = client.HeaderByNumber(context.Background(), nil)
	} else if len(input.BlockID) > 0 {
		hash := common.BytesToHash(input.BlockID)
		heder, err = client.HeaderByHash(context.Background(), hash)
	} else {
		number := new(big.Int)
		number.SetUint64(input.Height)
		heder, err = client.HeaderByNumber(context.Background(), number)
	}
	if err != nil {
		return nil, err
	}

	//
	var result adaptor.GetBlockInfoOutput
	result.Block.BlockID = heder.TxHash.Bytes()
	result.Block.BlockHeight = uint(heder.Number.Uint64())
	result.Block.Timestamp = heder.Time
	result.Block.ParentBlockID = heder.ParentHash.Bytes()
	result.Block.HeaderRawData = heder.Extra
	//result.Block.TxsRoot = //todo delete
	//result.Block.ProducerAddress=//todo delete
	//result.Block.IsStable=//todo delete

	return &result, nil
}
