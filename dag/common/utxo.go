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
	"fmt"
	"strings"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/txscript"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

func KeyToOutpoint(key []byte) modules.OutPoint {
	// key: u[addr]_[asset]_[index]
	data := strings.Split(string(key), "_")
	if len(data) != 3 {
		return modules.OutPoint{}
	}

	var vout modules.OutPoint
	vout.SetPrefix(modules.UTXO_PREFIX)

	if err := rlp.DecodeBytes([]byte(data[0][len(modules.UTXO_PREFIX):]), &vout.Addr); err != nil {
		vout.Addr = common.Address{}
	}

	if err := rlp.DecodeBytes([]byte(data[1]), &vout.Asset); err != nil {
		vout.Asset = modules.Asset{}
	}

	if err := rlp.DecodeBytes([]byte(data[2]), &vout.Hash); err != nil {
		vout.Hash = common.Hash{}
	}

	return vout
}

/**
根据用户地址、选择的资产类型、转账金额、手续费返回utxo
Return utxo struct and his total amount according to user's address, asset type, transaction amount and given gas.
*/
func ReadUtxos(addr common.Address, asset modules.Asset) (map[modules.OutPoint]*modules.Utxo, uint64) {
	// key: [UTXO_PREFIX][addr]_[asset]_[msgindex]_[out index]
	key := fmt.Sprintf("%s%s_%s_", string(modules.UTXO_PREFIX), addr.String(), asset.String())
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
			log.Error("Decode utxo data error:", err)
			continue
		}

		if utxo.IsLocked {
			continue
		}
		outpoint := KeyToOutpoint([]byte(k))
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
func UpdateUtxo(tx *modules.Transaction) {
	if len(tx.TxMessages) <= 0 {
		return
	}

	var payload interface{}

	for _, msg := range tx.TxMessages {
		payload = msg.Payload
		payment, ok := payload.(modules.PaymentPayload)
		if ok == true {
			var isCoinbase bool
			if len(payment.Inputs) <= 0 {
				isCoinbase = true
			} else {
				isCoinbase = false
			}

			// create utxo
			writeUtxo(tx.TxHash, msg.PayloadHash, payment.Outputs, isCoinbase)
			// destory utxo
			destoryUtxo(payment.Inputs)
		}
	}
}

/**
创建UTXO
*/
func writeUtxo(txHash common.Hash, msgIndex common.Hash, txouts []modules.Output, isCoinbase bool) {
	for outIndex, txout := range txouts {
		// get address
		spk, err := txscript.ExtractPkScriptAddrs(txout.PkScript)
		if err != nil {
			log.Error("Extract PkScript Address error.")
			continue
		}
		addr := spk.Address
		utxo := modules.Utxo{
			AccountAddr:  addr,
			TxID:         txHash,
			MessageIndex: msgIndex,
			OutIndex:     uint32(outIndex),
			Amount:       txout.Value,
			Asset:        txout.Asset,
			PkScript:     txout.PkScript,
			IsCoinBase:   isCoinbase,
			IsLocked:     false,
		}

		// write to database
		outpoint := modules.OutPoint{
			Addr:  addr,
			Asset: txout.Asset,
			Hash:  rlp.RlpHash(utxo),
		}
		outpoint.SetPrefix(modules.UTXO_PREFIX)

		v, err := rlp.EncodeToBytes(utxo)
		if err != nil {
			continue
		}
		if err = storage.Store(string(outpoint.ToKey()), v); err != nil {
			log.Error("Write utxo error: %s", err)
		}

	}
}

/**
销毁utxo
destory utxo, delete from UTXO database
*/
func destoryUtxo(txins []modules.Input) {
	for _, txin := range txins {
		outpoint := txin.PreviousOutPoint
		if err := storage.Delete(outpoint.ToKey()); err != nil {
			log.Error("Destory uxto error: %s", err)
		}
	}
}

/**
存储Asset的信息
write asset info to leveldb
*/
func SaveAssetInfo(assetInfo *modules.AssetInfo) error {
	assetID, err := rlp.EncodeToBytes(assetInfo.AssetID)
	if err != nil {
		return err
	}

	key := append(modules.ASSET_INFO_PREFIX, assetID...)

	data, err := rlp.EncodeToBytes(assetInfo)

	err = storage.Store(string(key), data)
	if err != nil {
		return err
	}

	return nil
}

/**
根据assetid从数据库中获取asset的信息
get asset infomation from leveldb by assetid ( Asset struct type )
*/
func GetAssetInfo(assetId *modules.Asset) (modules.AssetInfo, error) {
	assetID, err := rlp.EncodeToBytes(assetId)
	if err != nil {
		return modules.AssetInfo{}, err
	}

	key := append(modules.ASSET_INFO_PREFIX, assetID...)

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
	outpoint := modules.OutPoint{
		Addr:  addr,
		Asset: asset,
	}
	outpoint.SetPrefix(modules.UTXO_PREFIX)
	preKey := outpoint.ToPrefixKey()

	balance := uint64(0)
	if data := storage.GetPrefix(preKey); data != nil {
		for _, v := range data {
			var utxo modules.Utxo
			if err := rlp.DecodeBytes(v, &utxo); err != nil {
				log.Error("Decode utxo data error:", err)
				continue
			}

			if utxo.IsLocked {
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
