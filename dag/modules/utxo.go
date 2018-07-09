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
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
	"strings"
)

type Asset struct {
	AssertId IDType36 `json:"assert_id"` // 资产类别
	UniqueId IDType36 `json:"unique_id"` // every token has its unique id
	ChainId  IDType36 `json:"chain_id"`  // main chain id or sub-chain id
}

func (asset *Asset) String() string {
	data, err := rlp.EncodeToBytes(asset)
	if err != nil {
		return ""
	}
	return string(data)
}

// key: utxo.hash(utxo+timestamp)
type Utxo struct {
	AccountAddr  common.Address `json:"account_id"`    // 所属人id
	TxID         common.Hash    `json:"unit_id"`       // transaction id
	MessageIndex uint32         `json:"message_index"` // message index in transaction
	OutIndex     uint32         `json:"output_index"`
	Amount       uint64         `json:"amount"`      // 数量
	Asset        Asset          `json:"Asset"`       // 资产类别
	PkScript     []byte         `json:"program"`     // 要执行的代码段
	IsCoinBase   bool           `json:"is_coinbase"` //
	IsLocked     bool           `json:"is_locked"`
}

// utxo key
type OutPoint struct {
	Prefix [1]byte // default 'u'
	Addr   common.Address
	Asset  Asset
	Hash   common.Hash // reference Utxo struct key field
}

func (outpoint *OutPoint) ToPrefixKey() []byte {
	out := fmt.Sprintf("%s%s_%s",
		outpoint.Prefix,
		outpoint.Addr.String(),
		outpoint.Asset.String())
	return []byte(out)
}

func (outpoint *OutPoint) ToKey() []byte {
	out := fmt.Sprintf("%s%s_%s_%s",
		outpoint.Prefix,
		outpoint.Addr.String(),
		outpoint.Asset.String(),
		outpoint.Hash.String(),
	)
	return []byte(out)
}

func KeyToOutpoint(key []byte) OutPoint {
	// key: u[Addr]_[Asset]_[index]
	data := strings.Split(string(key), "_")
	if len(data) != 3 {
		return OutPoint{}
	}

	var vout OutPoint
	vout.Prefix = [1]byte{data[0][0]}

	if err := rlp.DecodeBytes([]byte(data[0][1:]), &vout.Addr); err != nil {
		vout.Addr = common.Address{}
	}

	if err := rlp.DecodeBytes([]byte(data[1]), &vout.Asset); err != nil {
		vout.Asset = Asset{}
	}

	if err := rlp.DecodeBytes([]byte(data[2]), &vout.Hash); err != nil {
		vout.Hash = common.Hash{}
	}

	return vout
}

func (outpoint *OutPoint) String() string {
	data, err := rlp.EncodeToBytes(outpoint)
	if err != nil {
		return ""
	}
	return string(data)
}

func (outpoint *OutPoint) Bytes() []byte {
	data, err := rlp.EncodeToBytes(outpoint)
	if err != nil {
		return nil
	}
	return data
}

type Input struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
}

type Output struct {
	Value    uint64
	Asset    Asset
	PkScript []byte
}

type SpendProof struct {
	Unit string `json:"unit"`
}