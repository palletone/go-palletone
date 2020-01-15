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

//TxBasicInfo 一个交易的基本信息
type TxBasicInfo struct {
	TxID           []byte `json:"tx_id"`           //交易的ID，Hash
	TxRawData      []byte `json:"tx_raw"`          //交易的二进制数据
	CreatorAddress string `json:"creator_address"` //交易的发起人
	TargetAddress  string `json:"target_address"`  //交易的目标地址（被调用的合约、收款人）
	IsInBlock      bool   `json:"is_in_block"`     //是否已经被打包到区块链中
	IsSuccess      bool   `json:"is_success"`      //是被标记为成功执行
	IsStable       bool   `json:"is_stable"`       //是否已经稳定不可逆
	BlockID        []byte `json:"block_id"`        //交易被打包到了哪个区块ID
	BlockHeight    uint   `json:"block_height"`    //交易被打包到的区块的高度
	TxIndex        uint   `json:"tx_index"`        //Tx在区块中的位置
	Timestamp      uint64 `json:"timestamp"`       //交易被打包的时间戳
}

func (tx *TxBasicInfo) String() string {
	d, _ := json.Marshal(tx)
	return string(d)
}

type txBasicInfo4Json struct {
	TxID           string `json:"tx_id"`           //交易的ID，Hash
	TxRawData      string `json:"tx_raw"`          //交易的二进制数据
	CreatorAddress string `json:"creator_address"` //交易的发起人
	TargetAddress  string `json:"target_address"`  //交易的目标地址（被调用的合约、收款人）
	IsInBlock      bool   `json:"is_in_block"`     //是否已经被打包到区块链中
	IsSuccess      bool   `json:"is_success"`      //是被标记为成功执行
	IsStable       bool   `json:"is_stable"`       //是否已经稳定不可逆
	BlockID        string `json:"block_id"`        //交易被打包到了哪个区块ID
	BlockHeight    uint   `json:"block_height"`    //交易被打包到的区块的高度
	TxIndex        uint   `json:"tx_index"`        //Tx在区块中的位置
	Timestamp      uint64 `json:"timestamp"`       //交易被打包的时间戳
}

func (tx *TxBasicInfo) UnmarshalJSON(input []byte) error {
	tx4json := txBasicInfo4Json{}
	err := json.Unmarshal(input, &tx4json)
	if err != nil {
		return err
	}
	setTxBasicInfoFromJson(tx, tx4json)
	return nil
}
func setTxBasicInfoFromJson(tx *TxBasicInfo, tx4json txBasicInfo4Json) {
	tx.TxID, _ = hex.DecodeString(tx4json.TxID)
	tx.TxRawData, _ = hex.DecodeString(tx4json.TxRawData)
	tx.CreatorAddress = tx4json.CreatorAddress
	tx.TargetAddress = tx4json.TargetAddress
	tx.IsInBlock = tx4json.IsInBlock
	tx.IsSuccess = tx4json.IsSuccess
	tx.IsStable = tx4json.IsStable
	tx.BlockID, _ = hex.DecodeString(tx4json.BlockID)
	tx.BlockHeight = tx4json.BlockHeight
	tx.TxIndex = tx4json.TxIndex
	tx.Timestamp = tx4json.Timestamp
}
func convertTxBasicInfo2Json(tx TxBasicInfo) txBasicInfo4Json {
	return txBasicInfo4Json{
		TxID:           hex.EncodeToString(tx.TxID),
		TxRawData:      hex.EncodeToString(tx.TxRawData),
		CreatorAddress: tx.CreatorAddress,
		TargetAddress:  tx.TargetAddress,
		IsInBlock:      tx.IsInBlock,
		IsSuccess:      tx.IsSuccess,
		IsStable:       tx.IsStable,
		BlockID:        hex.EncodeToString(tx.BlockID),
		BlockHeight:    tx.BlockHeight,
		TxIndex:        tx.TxIndex,
		Timestamp:      tx.Timestamp,
	}
}
func (tx *TxBasicInfo) MarshalJSON() ([]byte, error) {
	tx4json := convertTxBasicInfo2Json(*tx)
	return json.Marshal(tx4json)
}
