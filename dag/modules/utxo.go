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

// Unspent Transaction Output module.
package modules

import (
	"github.com/palletone/go-palletone/common"
)

type Asset struct {
	AssertId IDType `json:"assert_id"` // 资产类别
	UniqueId IDType `json:"unique_id"` // every token has its unique id
	ChainId  IDType `json:"chain_id"`  // main chain id or sub-chain id
}

// key: utxo.hash(utxo+timestamp)
type Utxo struct {
	AccountId    common.Address `json:"account_id"`    // 所属人id
	TxID         common.Hash    `json:"unit_id"`       // transaction id
	MessageIndex uint32         `json:"message_index"` // message index in transaction
	Amount       uint64         `json:"amount"`        // 数量
	Asset        Asset          `json:"asset"`         // 资产类别
	PkScript     string         `json:"program"`       // 要执行的代码段
	Key          common.Hash         `json:"key"`           // 索引值
	IsCoinBase   int16          `json:"is_coinbase"`   //
	IsLocked     bool           `json:"is_locked"`
}

// OutPoint defines a bitcoin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash  common.Hash	// reference Utxo struct key field
	Index uint32
}

type Input struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
}

type Output struct {
	Value    int64
	Asset    Asset
	PkScript []byte
}

type SpendProof struct {
	Unit string `json:"unit"`
}

/**
根据用户地址、选择的资产类型、转账金额、手续费返回utxo
return utxo struct according to user's address, asset type, transaction amount and given gas.
*/
func GetUtxos(addr common.Address, asset Asset, amount uint64, gas uint64) []Utxo {
	vout := make([]Utxo, 0)
	return vout
}

/**
检查UTXO是否正确
check UTXO, if passed return true, else return false
*/
func CheckOutput(vout *Utxo) bool {
	return true
}

func WriteUtxo(tx *PaymentPayload) (common.Hash, error){
	return common.Hash{}, nil
}