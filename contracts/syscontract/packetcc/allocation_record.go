/*
 *
 * 	This file is part of go-palletone.
 * 	go-palletone is free software: you can redistribute it and/or modify
 * 	it under the terms of the GNU General Public License as published by
 * 	the Free Software Foundation, either version 3 of the License, or
 * 	(at your option) any later version.
 * 	go-palletone is distributed in the hope that it will be useful,
 * 	but WITHOUT ANY WARRANTY; without even the implied warranty of
 * 	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * 	GNU General Public License for more details.
 * 	You should have received a copy of the GNU General Public License
 * 	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018-2020
 *
 */

package packetcc

import (
	"encoding/hex"
	"github.com/palletone/go-palletone/dag/modules"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
)

// 旧的红包领取记录
type OldPacketAllocationRecord struct {
	PubKey      []byte //红包公钥
	Message     string //领取红包用的消息，防止重复领取
	Amount      uint64 //领取的Token数量
	Token       *modules.Asset
	ToAddress   common.Address //领取人的地址
	RequestHash common.Hash    //领取请求的Hash
	Timestamp   uint64         //领取的时间戳，主要用于排序
}

//红包领取记录
type PacketAllocationRecord struct {
	PubKey      []byte //红包公钥
	Message     string //领取红包用的消息，防止重复领取
	Tokens      []*RecordTokens
	ToAddress   common.Address //领取人的地址
	RequestHash common.Hash    //领取请求的Hash
	Timestamp   uint64         //领取的时间戳，主要用于排序
}

// 兼容
func OldRecord2New(old *OldPacketAllocationRecord) *PacketAllocationRecord {
	return &PacketAllocationRecord{
		PubKey:  old.PubKey,
		Message: old.Message,
		Tokens: []*RecordTokens{
			{
				Amount: old.Amount,
				Asset:  old.Token,
			},
		},
		ToAddress:   old.ToAddress,
		RequestHash: old.RequestHash,
		Timestamp:   old.Timestamp,
	}
}

// 保存领取记录
func savePacketAllocationRecord(stub shim.ChaincodeStubInterface, record *PacketAllocationRecord) error {
	key := PacketAllocationRecordPrefix + hex.EncodeToString(record.PubKey) + "-" + record.Message
	value, err := rlp.EncodeToBytes(record)
	if err != nil {
		return err
	}
	return stub.PutState(key, value)
}

// 获取红包的所有领取记录
func getPacketAllocationHistory(stub shim.ChaincodeStubInterface, pubKey []byte) (
	[]*PacketAllocationRecord, error) {
	key := PacketAllocationRecordPrefix + hex.EncodeToString(pubKey) + "-"
	kvs, err := stub.GetStateByPrefix(key)
	if err != nil {
		return nil, err
	}
	result := make([]*PacketAllocationRecord, len(kvs))
	for i, kv := range kvs {
		p := PacketAllocationRecord{}
		err = rlp.DecodeBytes(kv.Value, &p)
		if err != nil {
			// 兼容
			or := OldPacketAllocationRecord{}
			err = rlp.DecodeBytes(kv.Value, &or)
			if err != nil {
				return nil, err
			}
			// 转换
			nr := OldRecord2New(&or)
			p = *nr
		}
		sort.Slice(p.Tokens, func(i, j int) bool {
			return p.Tokens[i].Amount > p.Tokens[j].Amount
		})
		result[i] = &p
	}

	return result, nil
}

//func getPacketAllocationRecord(stub shim.ChaincodeStubInterface, pubKey []byte, message string) (
//	*PacketAllocationRecord, error) {
//	key := PacketAllocationRecordPrefix + hex.EncodeToString(pubKey) + "-" + message
//	value, err := stub.GetState(key)
//	if err != nil {
//		return nil, err
//	}
//	p := PacketAllocationRecord{}
//	err = rlp.DecodeBytes(value, &p)
//	if err != nil {
//		return nil, err
//	}
//	return &p, nil
//}

type PacketAllocationRecordJson struct {
	PubKey      string //红包公钥
	Message     string //领取红包用的消息，防止重复领取
	Tokens      []RecordTokensJson
	ToAddress   common.Address //领取人的地址
	RequestHash string         //领取请求的Hash
	Timestamp   uint64         //领取的时间戳，主要用于排序
}

func convertAllocationRecord2Json(record *PacketAllocationRecord) *PacketAllocationRecordJson {
	p := &PacketAllocationRecordJson{
		PubKey:      hex.EncodeToString(record.PubKey),
		Message:     record.Message,
		ToAddress:   record.ToAddress,
		RequestHash: record.RequestHash.String(),
		Timestamp:   record.Timestamp,
	}
	p.Tokens = make([]RecordTokensJson, len(record.Tokens))
	for i, t := range record.Tokens {
		p.Tokens[i].Amount = t.Asset.DisplayAmount(t.Amount)
		p.Tokens[i].Asset = t.Asset.String()
	}
	return p
}
