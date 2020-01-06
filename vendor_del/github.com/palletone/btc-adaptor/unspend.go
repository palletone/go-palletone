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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/shopspring/decimal"

	"github.com/palletone/adaptor"
)

//==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ===

//var GHomeDir = btcutil.AppDataDir("btcwallet", false)
var GHomeDir = btcutil.AppDataDir("btcd", false)
var GCertPath = filepath.Join(GHomeDir, "rpc.cert")

func GetClient(rpcParams *RPCParams) (*rpcclient.Client, error) {
	//read cert from file
	var connCfg *rpcclient.ConnConfig
	if rpcParams.CertPath == "" {
		rpcParams.CertPath = GCertPath
	}
	if rpcParams.CertPath != "" {
		certs, err := ioutil.ReadFile(rpcParams.CertPath)
		if err != nil {
			return nil, err
		}

		// Connect to local bitcoin core RPC server using HTTP POST mode.
		connCfg = &rpcclient.ConnConfig{
			Host:         rpcParams.Host,
			Endpoint:     "ws",
			User:         rpcParams.RPCUser,
			Pass:         rpcParams.RPCPasswd,
			HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
			//DisableTLS:   true,  // Bitcoin core does not provide TLS by default
			Certificates: certs, // btcwallet provide TLS by default
		}
	} else {
		// Connect to local bitcoin core RPC server using HTTP POST mode.
		connCfg = &rpcclient.ConnConfig{
			Host:         rpcParams.Host,
			Endpoint:     "ws",
			User:         rpcParams.RPCUser,
			Pass:         rpcParams.RPCPasswd,
			HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   true, // Bitcoin core does not provide TLS by default
			//Certificates: certs, // btcwallet provide TLS by default
		}
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetNet(netID int) *chaincfg.Params {
	//chainnet
	var realNet *chaincfg.Params
	if netID == NETID_MAIN {
		realNet = &chaincfg.MainNetParams
	} else {
		realNet = &chaincfg.TestNet3Params
	}
	return realNet
}

func getAllUnspend(client *rpcclient.Client, addr btcutil.Address) (map[string]float64, error) {
	//get all raw transaction
	count := 999999
	msgTxs, err := client.SearchRawTransactionsVerbose(addr, 0, count, true, false, []string{}) //BTCD API
	if err != nil {
		return map[string]float64{}, fmt.Errorf("SearchRawTransactionsVerbose failed %s", err.Error())
	}

	addrStr := addr.String()
	//save utxo to map, check next one transanction is spend or not
	outputIndex := map[string]float64{}
	//the result for return
	for _, msgTx := range msgTxs {
		if int(msgTx.Confirmations) < MinConfirm {
			continue
		}
		//transaction inputs
		for _, in := range msgTx.Vin {
			//check is spend or not
			idIndex := in.Txid + fmt.Sprintf("%02x", in.Vout)
			_, exist := outputIndex[idIndex]
			if exist { //spend
				delete(outputIndex, idIndex)
			}
		}

		//transaction outputs
		for _, out := range msgTx.Vout {
			if 0 == len(out.ScriptPubKey.Addresses) {
				continue
			}
			if out.ScriptPubKey.Addresses[0] == addrStr {
				outputIndex[msgTx.Txid+fmt.Sprintf("%02x", out.N)] = out.Value
			}
		}
	}

	return outputIndex, nil
}
func GetBalance(input *adaptor.GetBalanceInput, rpcParams *RPCParams, netID int) (*adaptor.GetBalanceOutput, error) {
	if input.Address == "" {
		return nil, fmt.Errorf("the Address is empty")
	}

	//chainnet
	realNet := GetNet(netID)

	//convert address from string
	addr, err := btcutil.DecodeAddress(input.Address, realNet)
	if err != nil {
		return nil, fmt.Errorf("DecodeAddress address failed %s", err.Error())
	}
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	outputIndexMap, err := getAllUnspend(client, addr)
	if err != nil {
		return nil, err
	}

	//compute total Amount for balance
	var result adaptor.GetBalanceOutput
	var allAmount float64
	for _, value := range outputIndexMap {
		allAmount += value
	}

	//
	bigInt := new(big.Int)
	bigInt.SetUint64(uint64(uint64(decimal.NewFromFloat(allAmount).Mul(decimal.New(1, 8)).IntPart())))
	result.Balance.Amount = bigInt
	result.Balance.Asset = "BTC"

	return &result, nil
}

//
//type GetBalanceHttpResponse struct {
//	//Status string `json:"status"`
//	Data struct {
//		//Network            string `json:"network"`
//		Address            string `json:"address"`
//		ConfirmedBalance   string `json:"confirmed_balance"`
//		UnconfirmedBalance string `json:"unconfirmed_balance"`
//	} `json:"data"`
//}
//
//func GetBalanceHttp(params *adaptor.GetBalanceHttpParams, netID int) (*adaptor.GetBalanceHttpResult, error) {
//	if "" == params.Address {
//		return nil, errors.New("Address is empty")
//	}
//	var request string
//	if netID == NETID_MAIN {
//		request = base + "get_address_balance/BTC/"
//	} else {
//		request = base + "get_address_balance/BTCTEST/"
//	}
//	request += params.Address
//	if params.Minconf != 0 {
//		request += "/" + strconv.Itoa(params.Minconf)
//	}
//
//	strRespose, err, _ := httpGet(request)
//	if err != nil {
//		return nil, err
//	}
//
//	var balanceRes GetBalanceHttpResponse
//	err = json.Unmarshal([]byte(strRespose), &balanceRes)
//	if err != nil {
//		return nil, err
//	}
//	//compute total Amount for balance
//	var result adaptor.GetBalanceHttpResult
//	balance, _ := strconv.ParseFloat(balanceRes.Data.ConfirmedBalance, 64)
//	result.Value = balance
//
//	return &result, nil
//}

func GetTransactions(input *adaptor.GetAddrTxHistoryInput, rpcParams *RPCParams, netID int) (*adaptor.GetAddrTxHistoryOutput, error) {
	//chainnet
	realNet := GetNet(netID)

	//convert address from string
	addr, err := btcutil.DecodeAddress(input.FromAddress, realNet)
	if err != nil {
		return nil, fmt.Errorf("DecodeAddress FromAddress failed : %s", err.Error())
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//get all raw transaction
	var strs []string
	count := 999999
	msgTxs, err := client.SearchRawTransactionsVerbose(addr, 0, count, true, false, strs) //BTCD API
	if err != nil {
		return nil, err
	}

	isFilter := false
	if "" != input.ToAddress && input.AddressLogicAndOr {
		isFilter = true
	}

	//the result for return
	msgIndex := map[string]float64{}
	var output adaptor.GetAddrTxHistoryOutput
	for _, msgTx := range msgTxs {
		//one transaction result
		var tx adaptor.SimpleTransferTokenTx

		//
		change := float64(0)
		amount := float64(0)
		amountOther := float64(0)
		isTo := false
		for i, out := range msgTx.Vout {
			if out.ScriptPubKey.Type == "nulldata" { //todo: more op_return ?
				if strings.HasPrefix(out.ScriptPubKey.Asm, "OP_RETURN") {
					data, _ := hex.DecodeString(out.ScriptPubKey.Asm[len("OP_RETURN "):])
					tx.AttachData = data
				}
				continue
			}
			if len(out.ScriptPubKey.Addresses) == 0 {
				continue
			}
			idIndex := msgTx.Txid + fmt.Sprintf("%02x", i)
			msgIndex[idIndex] = out.Value //save in map

			if input.FromAddress == out.ScriptPubKey.Addresses[0] {
				change += out.Value
				continue
			}
			if "" != input.ToAddress {
				if input.ToAddress == out.ScriptPubKey.Addresses[0] {
					isTo = true
					tx.ToAddress = input.ToAddress
					amount += out.Value
					continue
				}
				amountOther += out.Value //todo no to address?empty
			} else {
				if "" == tx.ToAddress { //first recv address
					tx.ToAddress = out.ScriptPubKey.Addresses[0]
					amount = out.Value
					continue
				} else if tx.ToAddress == out.ScriptPubKey.Addresses[0] {
					amount += out.Value
					continue
				}
				amountOther += out.Value //todo amountOther?
			}
		}
		if isFilter && !isTo {
			continue
		}

		//get input amount
		inputAmount := float64(0)
		for i := 0; i < len(msgTx.Vin); i++ {
			idIndex := msgTx.Vin[i].Txid + fmt.Sprintf("%02x", i)
			if val, exist := msgIndex[idIndex]; exist { //have saved in map
				inputAmount += val
				continue
			}
			hashPre, err := chainhash.NewHashFromStr(msgTx.Vin[i].Txid)
			if err != nil {
				return nil, fmt.Errorf("hashPre failed : %s", err.Error())
			}
			txPreResult, err := client.GetRawTransactionVerbose(hashPre) //BTCD API
			if err != nil {
				return nil, fmt.Errorf("GetRawTransactionVerbose txPre %d failed : %s", i, err.Error())
			}
			for j, out := range txPreResult.Vout {
				if input.FromAddress != out.ScriptPubKey.Addresses[0] {
					continue
				}
				idIndex := txPreResult.Txid + fmt.Sprintf("%02x", j)
				msgIndex[idIndex] = out.Value //save in map
			}
			inputAmount += txPreResult.Vout[msgTx.Vin[i].Vout].Value
		}
		fee := inputAmount - change - amount - amountOther

		//turn to big int
		bigIntAmount := new(big.Int)
		bigIntAmount.SetUint64(uint64(uint64(decimal.NewFromFloat(amount).Mul(decimal.New(1, 8)).IntPart())))
		tx.Amount = adaptor.NewAmountAsset(bigIntAmount, "BTC")
		bigIntFee := new(big.Int)
		bigIntFee.SetUint64(uint64(uint64(decimal.NewFromFloat(fee).Mul(decimal.New(1, 8)).IntPart())))
		tx.Fee = adaptor.NewAmountAsset(bigIntFee, "BTC")

		tx.TxID, _ = hex.DecodeString(msgTx.Txid)
		txRaw, _ := hex.DecodeString(msgTx.Hex)
		tx.TxRawData = txRaw
		tx.CreatorAddress = input.FromAddress
		tx.TargetAddress = tx.ToAddress
		if msgTx.BlockHash != "" {
			tx.IsInBlock = true
			tx.IsSuccess = true
			blockID, _ := hex.DecodeString(msgTx.BlockHash)
			tx.BlockID = blockID
			blkHash, err := chainhash.NewHashFromStr(msgTx.BlockHash)
			if err == nil {
				blkResult, err := client.GetBlockVerbose(blkHash) //BTCD API
				if err == nil {
					tx.BlockHeight = uint(blkResult.Height) //GetBlockVerbose
				}
			}
		} else {
			tx.IsInBlock = false
			tx.IsSuccess = false
		}
		if msgTx.Confirmations >= MinConfirm {
			tx.IsStable = true
		} else {
			tx.IsStable = false
		}
		tx.TxIndex = 0 //todo
		tx.Timestamp = uint64(msgTx.Blocktime)

		//add to result for return
		output.Txs = append(output.Txs, &tx)
	}
	output.Count = uint32(len(output.Txs))

	return &output, nil
}
