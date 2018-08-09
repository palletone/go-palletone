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
package common

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/txscript"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

/**
根据用户地址、选择的资产类型、转账金额、手续费返回utxo
Return utxo struct and his total amount according to user's address, asset type
*/
func ReadUtxos(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64) {
	if dagconfig.DefaultConfig.UtxoIndex {
		return readUtxosByIndex(addr, asset)
	} else {
		return readUtxosFrAll(addr, asset)
	}
}

/**
根据UTXO索引数据库读取UTXO信息
To get utxo info by utxo index db
*/
func readUtxosByIndex(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64) {
	// step1. read outpoint from utxo index
	utxoIndex := modules.UtxoIndex{
		AccountAddr: addr,
		Asset:       asset,
	}
	key := utxoIndex.AssetKey()
	data := storage.GetPrefix([]byte(key))
	if data == nil {
		return nil, 0
	}
	// step2. get utxo
	vout := map[modules.OutPoint]*modules.Utxo{}
	balance := uint64(0)
	for k := range data {
		if err := utxoIndex.QueryFields([]byte(k)); err != nil {
			continue
		}
		udata, err := storage.Get(utxoIndex.OutPoint.ToKey())
		if err != nil {
			log.Error("Get utxo error by outpoint", "error:", err.Error())
			continue
		}
		var utxo modules.Utxo
		if err := rlp.DecodeBytes([]byte(udata), &utxo); err != nil {
			log.Error("Decode utxo data :", "error:", err.Error())
			continue
		}
		vout[utxoIndex.OutPoint] = &utxo
		balance += utxo.Amount
	}
	return vout, balance
}

/**
扫描UTXO全表，获取对应账户的指定token可用的utxo列表
To get utxo info by scanning all utxos.
*/
func readUtxosFrAll(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64) {
	// key: [UTXO_PREFIX][addr]_[asset]_[msgindex]_[out index]
	key := fmt.Sprintf("%s", string(modules.UTXO_PREFIX))
	data := storage.GetPrefix([]byte(key))
	if data == nil {
		return nil, 0
	}

	// value: utxo rlp bytes
	vout := map[modules.OutPoint]*modules.Utxo{}
	var balance uint64
	balance = 0
	for k, v := range data {
		var utxo modules.Utxo
		if err := rlp.DecodeBytes([]byte(v), &utxo); err != nil {
			log.Error("Decode utxo data :", "error:", err.Error())
			continue
		}
		// check asset
		if strings.Compare(asset.AssertId.String(), utxo.Asset.AssertId.String()) != 0 ||
			strings.Compare(asset.UniqueId.String(), utxo.Asset.UniqueId.String()) != 0 ||
			asset.ChainId != utxo.Asset.ChainId {
			continue
		}
		// get addr
		scriptPubKey, err := txscript.ExtractPkScriptAddrs(utxo.PkScript)
		if err != nil {
			log.Error("Get address from utxo output script", "error", err.Error())
			continue
		}
		// check address
		if strings.Compare(scriptPubKey.Address.String(), addr.String()) != 0 {
			continue
		}
		outpoint := modules.KeyToOutpoint([]byte(k))
		vout[outpoint] = &utxo
		balance += utxo.Amount
	}

	return vout, balance
}

/**
获取某个input对应的utxo信息
To get utxo struct according to it's input information
*/
func GetUxto(txin modules.Input) modules.Utxo {
	data, err := storage.Get(txin.PreviousOutPoint.ToKey())
	if err != nil {
		return modules.Utxo{}
	}
	var utxo modules.Utxo
	rlp.DecodeBytes(data, &utxo)
	return utxo
}

/**
根据交易信息中的outputs创建UTXO， 根据交易信息中的inputs销毁UTXO
To create utxo according to outpus in transaction, and destory utxo according to inputs in transaction
*/
func UpdateUtxo(txHash common.Hash, msg *modules.Message, msgIndex uint32, lockTime uint32) error {
	var payload interface{}

	payload = msg.Payload
	payment, ok := payload.(modules.PaymentPayload)
	if ok == true {
		// create utxo
		errs := writeUtxo(txHash, msgIndex, payment.Outputs, lockTime)
		// destory utxo
		destoryUtxo(payment.Inputs)
		if len(errs) > 0 {
			return errors.New("error occurred on updated utxos, check the log file to find details.")
		}
	}
	return nil
}

/**
创建UTXO
*/
func writeUtxo(txHash common.Hash, msgIndex uint32, txouts []modules.Output, lockTime uint32) []error {
	var errs []error
	for outIndex, txout := range txouts {
		utxo := modules.Utxo{
			TxID:         txHash,
			MessageIndex: msgIndex,
			OutIndex:     uint32(outIndex),
			Amount:       txout.Value,
			Asset:        txout.Asset,
			PkScript:     txout.PkScript,
			LockTime:     lockTime,
		}

		// write to database
		outpoint := modules.OutPoint{
			TxHash:       txHash,
			MessageIndex: msgIndex,
			OutIndex:     uint32(outIndex),
		}

		if err := storage.Store(storage.Dbconn, string(outpoint.ToKey()), utxo); err != nil {
			log.Error("Write utxo", "error", err.Error())
			errs = append(errs, err)
			continue
		}

		// write to utxo index db
		if dagconfig.DefaultConfig.UtxoIndex == false {
			continue
		}

		// get address
		spk, err := txscript.ExtractPkScriptAddrs(txout.PkScript)
		if err != nil {
			log.Error("Extract PkScript Address", "error", err.Error())
			errs = append(errs, err)
			continue
		}
		addr := spk.Address

		utxoIndex := modules.UtxoIndex{
			AccountAddr: addr,
			Asset:       txout.Asset,
			OutPoint:    outpoint,
		}
		utxoIndexVal := modules.UtxoIndexValue{
			Amount:   txout.Value,
			LockTime: lockTime,
		}
		if err := storage.Store(storage.Dbconn, string(utxoIndex.ToKey()), utxoIndexVal); err != nil {
			log.Error("Write utxo index error: %s", err.Error())
			errs = append(errs, err)
		}
	}
	return errs
}

/**
销毁utxo
destory utxo, delete from UTXO database
*/
func destoryUtxo(txins []modules.Input) {
	for _, txin := range txins {
		outpoint := txin.PreviousOutPoint
		if outpoint.IsEmpty() {
			if len(txin.Extra) > 0 {
				var assetInfo modules.AssetInfo
				if err := rlp.DecodeBytes(txin.Extra, &assetInfo); err == nil {
					// save asset info
					if err := SaveAssetInfo(&assetInfo); err != nil {
						log.Error("Save asset info error")
					}
				}
			}
			continue
		}
		// get utxo info
		data, err := storage.Get(outpoint.ToKey())
		if err != nil {
			log.Error("Query utxo when destory uxto", "error", err.Error())
			continue
		}
		var utxo modules.Utxo
		if err := rlp.DecodeBytes(data, &utxo); err != nil {
			log.Error("Decode utxo when destory uxto", "error", err.Error())
			continue
		}
		// delete utxo
		if err := storage.Delete(outpoint.ToKey()); err != nil {
			log.Error("Destory uxto", "error", err.Error())
			continue
		}
		// delete index data
		scriptPubKey, err := txscript.ExtractPkScriptAddrs(utxo.PkScript)
		if err != nil {
			log.Error("Extract address when destory uxto", "error", err.Error())
			continue
		}
		utxoIndex := modules.UtxoIndex{
			AccountAddr: scriptPubKey.Address,
			Asset:       utxo.Asset,
			OutPoint:    outpoint,
		}
		if err := storage.Delete(utxoIndex.ToKey()); err != nil {
			log.Error("Destory uxto index", "error", err.Error())
			continue
		}
	}
}

/**
存储Asset的信息
write asset info to leveldb
*/
func SaveAssetInfo(assetInfo *modules.AssetInfo) error {
	if err := storage.Store(storage.Dbconn, string(assetInfo.Tokey()), *assetInfo); err != nil {
		return err
	}

	return nil
}

/**
根据assetid从数据库中获取asset的信息
get asset infomation from leveldb by assetid ( Asset struct type )
*/
func GetAssetInfo(assetId *modules.Asset) (modules.AssetInfo, error) {
	key := append(modules.ASSET_INFO_PREFIX, assetId.AssertId.String()...)
	data, err := storage.Get(key)
	if err != nil {
		return modules.AssetInfo{}, err
	}

	var assetInfo modules.AssetInfo
	err = rlp.DecodeBytes(data, &assetInfo)

	if err != nil {
		return assetInfo, err
	}
	return assetInfo, nil
}

/**
获得某个账户下面的余额信息
To get balance by wallet address and his/her chosen asset type
*/
func WalletBalance(addr common.Address, asset modules.Asset) uint64 {
	if dagconfig.DefaultConfig.UtxoIndex {
		return walletBalanceByIndex(addr, asset)
	} else {
		return walletBalanceFrAll(addr, asset)
	}
}

/**
通过索引数据库获得账户余额
To get balance by utxo index db
*/
func walletBalanceByIndex(addr common.Address, asset modules.Asset) uint64 {
	balance := uint64(0)

	utxoIndex := modules.UtxoIndex{
		AccountAddr: addr,
		Asset:       asset,
	}
	preKey := utxoIndex.AssetKey()

	if data := storage.GetPrefix(preKey); data != nil {
		for _, v := range data {
			var utxoIndexVal modules.UtxoIndexValue
			if err := rlp.DecodeBytes(v, &utxoIndexVal); err != nil {
				log.Error("Decode utxo data error:", err)
				continue
			}
			balance += utxoIndexVal.Amount
		}
	}

	return balance

}

/**
通过查询全表获得账户余额
To get balance by query all utxo table
*/
func walletBalanceFrAll(addr common.Address, asset modules.Asset) uint64 {
	balance := uint64(0)

	preKey := fmt.Sprintf("%s", modules.UTXO_PREFIX)

	if data := storage.GetPrefix([]byte(preKey)); data != nil {
		for _, v := range data {
			var utxo modules.Utxo
			if err := rlp.DecodeBytes(v, &utxo); err != nil {
				log.Error("Decode utxo data error:", err)
				continue
			}
			if !checkUtxo(&addr, &asset, &utxo) {
				continue
			}
			balance += utxo.Amount
		}
	}

	return balance
}

/**
根据payload中的inputs获得对应的UTXO map
*/
func GetUxtoSetByInputs(txins []modules.Input) (map[modules.OutPoint]*modules.Utxo, uint64) {
	utxos := map[modules.OutPoint]*modules.Utxo{}
	total := uint64(0)
	for _, in := range txins {
		utxo := GetUxto(in)
		if unsafe.Sizeof(utxo) == 0 {
			continue
		}
		utxos[in.PreviousOutPoint] = &utxo
		total += utxo.Amount
	}
	return utxos, total
}

/**
获得某个账户下的token名称和assetid信息
To get someone account's list of tokens and those assetids
*/
func GetAccountTokens(addr common.Address) (map[modules.Asset]*modules.AccountToken, error) {
	if dagconfig.DefaultConfig.UtxoIndex {
		return getAccountTokensByIndex(addr)
	} else {
		return getAccountTokensWhole(addr)
	}
}

/**
通过索引数据库获得用户的token信息
To get account's token info by utxo index db
*/
func getAccountTokensByIndex(addr common.Address) (map[modules.Asset]*modules.AccountToken, error) {
	tokens := map[modules.Asset]*modules.AccountToken{}
	utxoIndex := modules.UtxoIndex{AccountAddr: addr}
	data := storage.GetPrefix(utxoIndex.AccountKey())
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	for k, v := range data {
		if err := utxoIndex.QueryFields([]byte(k)); err != nil {
			return nil, fmt.Errorf("Get account tokens error: data key is invalid")
		}
		var utxoIndexVal modules.UtxoIndexValue
		if err := rlp.DecodeBytes([]byte(v), &utxoIndexVal); err != nil {
			return nil, fmt.Errorf("Get account tokens error: data value is invalid")
		}
		val, ok := tokens[utxoIndex.Asset]
		if ok {
			val.Balance += utxoIndexVal.Amount
		} else {
			// get asset info
			assetInfo, err := GetAssetInfo(&utxoIndex.Asset)
			if err != nil {
				return nil, fmt.Errorf("Get acount tokens error: asset info does not exist")
			}
			tokens[utxoIndex.Asset] = &modules.AccountToken{
				Alias:   assetInfo.Alias,
				AssetID: utxoIndex.Asset,
				Balance: utxoIndexVal.Amount,
			}
		}
	}
	return tokens, nil
}

/**
遍历全局utxo，获取账户token信息
To get account token info by query the whole utxo table
*/
func getAccountTokensWhole(addr common.Address) (map[modules.Asset]*modules.AccountToken, error) {
	tokens := map[modules.Asset]*modules.AccountToken{}

	key := fmt.Sprintf("%s", string(modules.UTXO_PREFIX))
	data := storage.GetPrefix([]byte(key))
	if data == nil {
		fmt.Println("11111111111111111111")
		return nil, nil
	}

	for _, v := range data {
		var utxo modules.Utxo
		if err := rlp.DecodeBytes([]byte(v), &utxo); err != nil {
			return nil, err
		}
		if !checkUtxo(&addr, nil, &utxo) {
			fmt.Println("2222222222222222")
			continue
		}

		val, ok := tokens[utxo.Asset]
		if ok {
			val.Balance += utxo.Amount
		} else {
			// get asset info
			assetInfo, err := GetAssetInfo(&utxo.Asset)
			if err != nil {
				return nil, fmt.Errorf("Get acount tokens error: asset info does not exist")
			}
			tokens[utxo.Asset] = &modules.AccountToken{
				Alias:   assetInfo.Alias,
				AssetID: utxo.Asset,
				Balance: utxo.Amount,
			}
		}
	}
	return tokens, nil
}

/**
检查该utxo是否是需要的utxo
*/
func checkUtxo(addr *common.Address, asset *modules.Asset, utxo *modules.Utxo) bool {
	// check asset
	if asset != nil && (strings.Compare(asset.AssertId.String(), utxo.Asset.AssertId.String()) != 0 ||
		strings.Compare(asset.UniqueId.String(), utxo.Asset.UniqueId.String()) != 0 ||
		asset.ChainId != utxo.Asset.ChainId) {
		return false
	}
	// get addr
	scriptPubKey, err := txscript.ExtractPkScriptAddrs(utxo.PkScript)
	if err != nil {
		log.Error("Get address from utxo output script", "error", err.Error())
		return false
	}
	// check address
	if strings.Compare(scriptPubKey.Address.String(), addr.String()) != 0 {
		return false
	}
	return true
}

/**
根据交易列表计算交易费总和
To compute transactions' fees
*/
func ComputeFees(txs modules.Transactions) (uint64, error) {
	fees := uint64(0)
	for _, tx := range txs {
		for _, msg := range tx.TxMessages {
			payload, ok := msg.Payload.(modules.PaymentPayload)
			if ok == false {
				continue
			}
			inAmount := uint64(0)
			outAmount := uint64(0)
			for _, txin := range payload.Inputs {
				utxo := GetUxto(txin)
				if utxo.IsEmpty() {
					return 0, fmt.Errorf("Txin(txhash=%s, msgindex=%v, outindex=%v)'s utxo is empty:",
						txin.PreviousOutPoint.TxHash.String(),
						txin.PreviousOutPoint.MessageIndex,
						txin.PreviousOutPoint.OutIndex)
				}
				// check overflow
				if inAmount+utxo.Amount > 1<<64-1 {
					return 0, fmt.Errorf("Compute fees: txin total overflow")
				}
				inAmount += utxo.Amount
			}

			for _, txout := range payload.Outputs {
				// check overflow
				if outAmount+txout.Value > 1<<64-1 {
					return 0, fmt.Errorf("Compute fees: txout total overflow")
				}
				outAmount += txout.Value
			}
			if inAmount < outAmount {
				return 0, fmt.Errorf("Compute fees: tx %s txin amount less than txout amount.", tx.TxHash.String())
			}
			fees += inAmount - outAmount
		}
	}
	return fees, nil
}
