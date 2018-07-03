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

// key: utxo.hash(utxo+timestamp)
type Utxo struct {
	AccountId  string `json:"account_id"`  // 所属人id
	UnitId     string `json:"unit_id"`     // unit id
	Amount     uint64 `json:"amount"`      // 数量
	Asset      Asset  `json:"asset"`       // 资产类别
	Alias      string `json:"alias"`       // 资产别名
	Program    string `json:"program"`     // 要执行的代码段
	Key        string `json:"key"`         // 索引值
	IsCoinBase int16  `json:"is_coinbase"` //
}

type Input struct {
	TxHash             common.Hash `json:"unit"`
	MessageIndex       uint16      `json:"message_index"`
	InputIndex         uint16      `json:"input_index"`
	Asset              Asset       `json:"asset"`
	Denomination       uint64      `json:"denomination"` // default 1
	IsUnique           int16       `json:"is_unique"`    //default 1
	TypeEnum           string      `json:"type_unum"`    //'transfer','headers_commission','witnessing','issue'
	SrcUnit            string      `json:"src_unit"`
	SrcMessageIndex    uint16      `json:"src_message_index"`
	SrcOutputIndex     uint16      `json:"src_output_index"`
	FromMainChainIndex uint64      `json:"from_main_chain_index"`
	ToMainChainIndex   uint64      `json:"to_main_chain_index"`
	SerialNumber       uint64      `json:"serial_number"`
	Amount             uint64      `json:"amount"`
	Address            string      `json:"address"`
}

type Output struct {
	OutputId     uint64      `json:"output_id"`
	TxHash       common.Hash `json:"unit"`
	MessageIndex uint16      `json:"message_index"`
	OutputIndex  uint16      `json:"output_index"`
	Asset        Asset       `json:"asset"`
	Denomination uint64      `json:"denomination"` // default 1
	Amount       uint64      `json:"amount"`
	Address      string      `json:"address"`
	Blinding     string      `json:"blinding"`
	OutputHash   string      `json:"output_hash"`
	IsSerial     int16       `json:"is_serial"`   // default 0 if not stable yet
	IsSpent      int16       `json:"is_spend"`    // default 0
	IsCoinBase   int16       `json:"is_coinbase"` // wether token generates from minner or not
}

type SpendProof struct {
	Unit string `json:"unit"`
}

type Asset struct {
	AssertId IDType `json:"assert_id"` // 资产类别
	UniqueId IDType `json:"unique_id"` // every token has its unique id
	ChainId  IDType `json:"chain_id"`  // main chain id or sub-chain id
}

/**
	根据用户地址、选择的资产类型、转账金额、手续费返回utxo
	return utxo struct according to user's address, asset type, transaction amount and given gas.
 */
func GetUtxo(addr common.Address, asset Asset, amount uint64, gas uint64) Utxo {
	vout := Utxo{}
	return vout
}


/**
	检查UTXO是否正确
	check UTXO, if passed return true, else return false
 */
func CheckUtxo(vout *Utxo)  bool{
	return true
}




