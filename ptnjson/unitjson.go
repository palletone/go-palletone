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
 *  * @date 2018
 *
 */

package ptnjson

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
)

type UnitJson struct {
	UnitHeader *HeaderJson        `json:"unit_header"`  // unit header
	Txs        []*TxJson          `json:"transactions"` // transaction list
	UnitHash   common.Hash        `json:"unit_hash"`    // unit hash
	UnitSize   common.StorageSize `json:"unit_size"`    // unit size
	Reward     decimal.Decimal    `json:"reward"`       //区块奖励
}

type ChainUnitPropertyJson struct {
	FastHash        common.Hash `json:"fast_hash"`
	FastIndex       uint64      `json:"fast_index"`
	StableHash      common.Hash `json:"stable_hash"`
	StableIndex     uint64      `json:"stable_index"`
	FastTimestamp   string      `json:"fast_timestamp"`
	StableTimestamp string      `json:"stable_timestamp"`
}

type HeaderJson struct {
	ParentsHash   []common.Hash  `json:"parents_hash"`
	Hash          string         `json:"hash"`
	AuthorAddress string         `json:"mediator_address"`
	AuthorPubKey  string         `json:"mediator_pubkey"`
	AuthorSign    string         `json:"mediator_sign"` // the unit creation authors
	GroupSign     string         `json:"group_sign"`    // 群签名, 用于加快单元确认速度
	GroupPubKey   string         `json:"group_pubKey"`  // 群公钥, 用于验证群签名
	TxRoot        common.Hash    `json:"root"`
	TxsIllegal    []string       `json:"txs_illegal"` //Unit中非法交易索引
	Number        ChainIndexJson `json:"index"`
	Extra         string         `json:"extra"`
	CreationTime  time.Time      `json:"creation_time"` // unit create time
}

type ChainIndexJson struct {
	AssetID string `json:"asset_id"`
	Index   uint64 `json:"index"`
}

func ConvertUnit2Json(unit *modules.Unit, utxoQuery modules.QueryUtxoFunc,
	versionFunc modules.QueryStateByVersionFunc) *UnitJson {
	json := &UnitJson{
		UnitHash:   unit.Hash(),
		UnitSize:   unit.Size(),
		UnitHeader: ConvertUnitHeader2Json(unit.UnitHeader),
		Txs:        []*TxJson{},
	}
	for _, tx := range unit.Txs {
		txjson := ConvertTx2FullJson(tx, utxoQuery)
		json.Txs = append(json.Txs, txjson)
	}
	reward, err := unit.Txs[0].GetCoinbaseReward(versionFunc, tokenengine.Instance.GetAddressFromScript)
	if err != nil {
		log.Error(err.Error())
		return json
	}
	if reward.Amount != 0 {
		json.Reward = reward.Asset.DisplayAmount(reward.Amount)
	}
	return json
}

func ConvertUnitHeader2Json(header *modules.Header) *HeaderJson {
	json := &HeaderJson{
		ParentsHash:   header.ParentHash(),
		Hash:          header.Hash().String(),
		AuthorAddress: header.Author().String(),
		AuthorPubKey:  hex.EncodeToString(header.GetAuthors().PubKey),
		AuthorSign:    hex.EncodeToString(header.GetAuthors().Signature),
		GroupSign:     hex.EncodeToString(header.GetGroupSign()),
		GroupPubKey:   hex.EncodeToString(header.GetGroupPubkey()),
		TxRoot:        header.TxRoot(),
		TxsIllegal:    make([]string, 0),
		Extra:         hex.EncodeToString(header.Extra()),
		CreationTime:  time.Unix(header.Timestamp(), 0),
	}
	for _, txI := range header.GetTxsIllegal() {
		json.TxsIllegal = append(json.TxsIllegal, strconv.Itoa(int(txI)))
	}
	json.Number = ChainIndexJson{
		AssetID: header.GetNumber().AssetID.String(),
		Index:   header.GetNumber().Index,
	}
	return json
}

type UnitSummaryJson struct {
	UnitHeader *HeaderJson        `json:"unit_header"`  // unit header
	Txs        []common.Hash      `json:"transactions"` // transaction list
	UnitHash   common.Hash        `json:"unit_hash"`    // unit hash
	UnitSize   common.StorageSize `json:"unit_size"`    // unit size
	TxCount    int                `json:"transaction_count"`
}

func ConvertUnit2SummaryJson(unit *modules.Unit) *UnitSummaryJson {
	json := &UnitSummaryJson{
		UnitHash:   unit.Hash(),
		UnitSize:   unit.Size(),
		UnitHeader: ConvertUnitHeader2Json(unit.UnitHeader),
		Txs:        []common.Hash{},
		TxCount:    len(unit.Txs),
	}
	for _, tx := range unit.Txs {

		json.Txs = append(json.Txs, tx.Hash())
	}
	return json
}
