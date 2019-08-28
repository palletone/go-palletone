/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package adaptor

import (
	"encoding/hex"
	"encoding/json"
)

//BlockInfo 区块的基本信息
type BlockInfo struct {
	BlockID         []byte `json:"block_id"`         //交易被打包到了哪个区块ID
	BlockHeight     uint   `json:"block_height"`     //交易被打包到的区块的高度
	Timestamp       uint64 `json:"timestamp"`        //交易被打包的时间戳
	ParentBlockID   []byte `json:"parent_block_id"`  //父区块ID
	HeaderRawData   []byte `json:"header_raw_data"`  //区块头的原始信息
	TxsRoot         []byte `json:"txs_root"`         //默克尔根
	ProducerAddress string `json:"producer_address"` //生产者地址
	IsStable        bool   `json:"is_stable"`        //是否已经稳定不可逆
}

func (block *BlockInfo) String() string {
	d, _ := json.Marshal(block)
	return string(d)
}

type blockInfo4Json struct {
	BlockID         string `json:"block_id"`         //交易被打包到了哪个区块ID
	BlockHeight     uint   `json:"block_height"`     //交易被打包到的区块的高度
	Timestamp       uint64 `json:"timestamp"`        //交易被打包的时间戳
	ParentBlockID   string `json:"parent_block_id"`  //父区块ID
	HeaderRawData   string `json:"header_raw_data"`  //区块头的原始信息
	TxsRoot         string `json:"txs_root"`         //默克尔根
	ProducerAddress string `json:"producer_address"` //生产者地址
	IsStable        bool   `json:"is_stable"`        //是否已经稳定不可逆
}

func convertBlockInfo2Json(b BlockInfo) blockInfo4Json {
	return blockInfo4Json{
		BlockID:         hex.EncodeToString(b.BlockID),
		BlockHeight:     b.BlockHeight,
		Timestamp:       b.Timestamp,
		ParentBlockID:   hex.EncodeToString(b.ParentBlockID),
		HeaderRawData:   hex.EncodeToString(b.HeaderRawData),
		TxsRoot:         hex.EncodeToString(b.TxsRoot),
		ProducerAddress: b.ProducerAddress,
		IsStable:        b.IsStable,
	}
}
func setBlockInfoFromJson(b *BlockInfo, bjson blockInfo4Json) {
	b.BlockID, _ = hex.DecodeString(bjson.BlockID)
	b.BlockHeight = bjson.BlockHeight
	b.Timestamp = bjson.Timestamp
	b.ParentBlockID, _ = hex.DecodeString(bjson.ParentBlockID)
	b.HeaderRawData, _ = hex.DecodeString(bjson.HeaderRawData)
	b.TxsRoot, _ = hex.DecodeString(bjson.TxsRoot)
	b.ProducerAddress = bjson.ProducerAddress
	b.IsStable = bjson.IsStable
}
func (block *BlockInfo) MarshalJSON() ([]byte, error) {
	b4json := convertBlockInfo2Json(*block)
	return json.Marshal(b4json)
}
func (block *BlockInfo) UnmarshalJSON(input []byte) error {
	b4Json := blockInfo4Json{}
	err := json.Unmarshal(input, &b4Json)
	if err != nil {
		return err
	}
	setBlockInfoFromJson(block, b4Json)
	return nil
}
