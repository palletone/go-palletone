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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"strings"
)

func KeyToOutpoint(key []byte) modules.OutPoint {
	// key: u[addr]_[asset]_[index]
	data := strings.Split(string(key), "_")
	if len(data) != 3 {
		return modules.OutPoint{}
	}

	var vout modules.OutPoint
	vout.Prefix = [1]byte{data[0][0]}

	if err := rlp.DecodeBytes([]byte(data[0][1:]), &vout.Addr); err != nil {
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
	// key: u[addr]_[asset]_[index]
	key := fmt.Sprintf("u%s_%s_", addr.String(), asset.String())
	data := storage.GetPrefix([]byte(key))
	if data == nil {
		return nil, 0
	}

	// value: utxo rlp bytes
	vout := make(map[modules.OutPoint]*modules.Utxo, len(data))
	for k, v := range data {
		var utxo modules.Utxo
		if err := rlp.DecodeBytes([]byte(v), &utxo); err != nil {
			continue
		}
		outpoint := KeyToOutpoint([]byte(k))
		vout[outpoint] = &utxo
	}

	return vout, 0
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
根据交易信息中的outpus创建UTXO， 根据交易信息中的inputs销毁UTXO
To create utxo according to outpus in transaction, and destory utxo according to inputs in transaction
*/
func UpdateUtxo(addr common.Address, tx modules.Transaction) {
	if len(tx.TxMessages) <= 0 {
		return
	}

	var payload interface{}

	for index, msg := range tx.TxMessages {
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
			writeUtxo(addr, tx.TxHash, uint32(index), payment.Outputs, isCoinbase)
			// destory utxo
			destoryUtxo(payment.Inputs)
		}
	}
}

/**
创建UTXO
*/
func writeUtxo(addr common.Address, txHash common.Hash, index uint32, txouts []modules.Output, isCoinbase bool) {
	for outIndex, txout := range txouts {
		utxo := modules.Utxo{
			AccountAddr:  addr,
			TxID:         txHash,
			MessageIndex: uint32(index),
			OutIndex:     uint32(outIndex),
			Amount:       txout.Value,
			Asset:        txout.Asset,
			PkScript:     txout.PkScript,
			IsCoinBase:   isCoinbase,
			IsLocked:     false,
		}

		// write to database
		outpoint := modules.OutPoint{
			Prefix: [1]byte{'u'},
			Addr:   addr,
			Asset:  txout.Asset,
			Hash:   rlp.RlpHash(utxo),
		}
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
