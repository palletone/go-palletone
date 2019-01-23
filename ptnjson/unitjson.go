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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

type UnitJson struct {
	UnitHeader *HeaderJson        `json:"unit_header"`  // unit header
	Txs        []*TxJson          `json:"transactions"` // transaction list
	UnitHash   common.Hash        `json:"unit_hash"`    // unit hash
	UnitSize   common.StorageSize `json:"unit_size"`    // unit size

}
type HeaderJson struct {
	ParentsHash []common.Hash `json:"parents_hash"`
	//AssetIDs      []string       `json:"assets"`
	AuthorAddress string         `json:"mediator_address"`
	AuthorSign    string         `json:"mediator_sign"` // the unit creation authors
	GroupSign     string         `json:"groupSign"`     // 群签名, 用于加快单元确认速度
	GroupPubKey   string         `json:"groupPubKey"`   // 群公钥, 用于验证群签名
	TxRoot        common.Hash    `json:"root"`
	Number        ChainIndexJson `json:"index"`
	Extra         string         `json:"extra"`
	CreationTime  time.Time      `json:"creation_time"` // unit create time
}
type ChainIndexJson struct {
	AssetID string `json:"asset_id"`
	IsMain  bool   `json:"is_main"`
	Index   uint64 `json:"index"`
}

func ConvertUnit2Json(unit *modules.Unit) *UnitJson {
	json := &UnitJson{
		UnitHash:   unit.Hash(),
		UnitSize:   unit.Size(),
		UnitHeader: convertUnitHeader2Json(unit.UnitHeader),
		Txs:        []*TxJson{},
	}

	for _, tx := range unit.Txs {
		txjson := ConvertTx2Json(tx)
		json.Txs = append(json.Txs, &txjson)
	}
	return json
}
func convertUnitHeader2Json(header *modules.Header) *HeaderJson {
	json := &HeaderJson{
		ParentsHash:   header.ParentsHash,
		AuthorAddress: header.Authors.Address.String(),
		AuthorSign:    hex.EncodeToString(append(header.Authors.R, header.Authors.S...)),
		GroupSign:     hex.EncodeToString(header.GroupSign),
		GroupPubKey:   hex.EncodeToString(header.GroupPubKey),
		TxRoot:        header.TxRoot,
		Extra:         hex.EncodeToString(header.Extra),
		CreationTime:  time.Now(), // TODO: header.Creationdate
	}
	json.Number = ChainIndexJson{
		AssetID: header.Number.AssetID.ToAssetId(),
		IsMain:  header.Number.IsMain,
		Index:   header.Number.Index,
	}
	return json
}
