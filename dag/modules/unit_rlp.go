/* This file is part of go-palletone.
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

   @author PalletOne core developers <dev@pallet.one>
   @date 2018
*/
package modules

import (
	"io"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
)

type headerTemp struct {
	ParentsHash []common.Hash `json:"parents_hash"`
	Authors     Authentifier  `json:"mediator"`    // the unit creation authors
	GroupSign   []byte        `json:"groupSign"`   // 群签名, 用于加快单元确认速度
	GroupPubKey []byte        `json:"groupPubKey"` // 群公钥, 用于验证群签名
	TxRoot      common.Hash   `json:"root"`
	TxsIllegal  []uint16      `json:"txs_illegal"` //Unit中非法交易索引
	Number      *ChainIndex   `json:"index"`
	Extra       []byte        `json:"extra"`
	Time        uint32        `json:"creation_time"` // unit create time
	CryptoLib   []byte        `json:"crypto_lib"`    //该区块使用的加解密算法和哈希算法，0位表示非对称加密算法，1位表示Hash算法
}

func (input *Header) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}
	temp := &headerTemp{}
	err = rlp.DecodeBytes(raw, temp)
	if err != nil {
		return err
	}

	input.ParentsHash = temp.ParentsHash
	input.Authors = temp.Authors
	input.GroupSign = temp.GroupSign
	input.GroupPubKey = temp.GroupPubKey
	input.TxRoot = temp.TxRoot
	input.TxsIllegal = temp.TxsIllegal
	input.Number = temp.Number
	input.Extra = temp.Extra
	input.Time = int64(temp.Time)
	input.CryptoLib = temp.CryptoLib
	return nil
}
func (input *Header) EncodeRLP(w io.Writer) error {
	temp := &headerTemp{}
	temp.ParentsHash = input.ParentsHash
	temp.Authors = input.Authors
	temp.GroupSign = input.GroupSign
	temp.GroupPubKey = input.GroupPubKey
	temp.TxRoot = input.TxRoot
	temp.TxsIllegal = input.TxsIllegal
	temp.Number = input.Number
	temp.Extra = input.Extra
	temp.Time = uint32(input.Time)
	temp.CryptoLib = input.CryptoLib
	return rlp.Encode(w, temp)
}

type unitTemp struct {
	UnitHeader *Header      `json:"unit_header"`  // unit header
	Txs        Transactions `json:"transactions"` // transaction list
	UnitHash   common.Hash  `json:"unit_hash"`    // unit hash
	UnitSize   uint32       `json:"unit_size"`    // unit size
	// These fields are used by package ptn to track
	// inter-peer block relay.
	ReceivedAt   uint32
	ReceivedFrom []byte
}

func (input *Unit) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}
	temp := &unitTemp{}
	err = rlp.DecodeBytes(raw, temp)
	if err != nil {
		return err
	}

	input.UnitHeader = temp.UnitHeader
	input.Txs = temp.Txs
	input.UnitHash = temp.UnitHash
	input.UnitSize = common.StorageSize(temp.UnitSize)

	input.ReceivedAt = time.Unix(int64(temp.ReceivedAt), 0)
	//todo  ReceivedFrom
	return nil
}
func (input *Unit) EncodeRLP(w io.Writer) error {
	temp := &unitTemp{}
	temp.UnitHeader = input.UnitHeader
	temp.Txs = input.Txs
	temp.UnitHash = input.UnitHash
	temp.UnitSize = uint32(input.UnitSize)
	temp.ReceivedAt = uint32(input.ReceivedAt.Unix())

	return rlp.Encode(w, temp)
}
