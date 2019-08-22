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
	"github.com/palletone/go-palletone/common/hexutil"
	// "github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	// "github.com/palletone/go-palletone/dag/parameter"
	"github.com/palletone/go-palletone/tokenengine"
	"time"
)

type UtxoJson struct {
	TxHash         string    `json:"txid"`          // reference Utxo struct key field
	MessageIndex   uint32    `json:"message_index"` // message index in transaction
	OutIndex       uint32    `json:"out_index"`
	Amount         uint64    `json:"amount"`           // 数量
	Asset          string    `json:"asset"`            // 资产类别
	PkScriptHex    string    `json:"pk_script_hex"`    // 要执行的代码段
	PkScriptString string    `json:"pk_script_string"` // 要执行的代码段
	Time           time.Time `json:"create_time"`      //创建该UTXO的时间（打包到Unit的时间）
	LockTime       uint32    `json:"lock_time"`        //
	FlagStatus     string    `json:"flag_status"`      // utxo状态
	CoinDays       uint64    `json:"coin_days"`        //这个Utxo存在多少天了。用于计算利息
	//AmountWithInterest uint64    `json:"amount_with_interest"` //包含利息后的金额
}

func (utxo *UtxoJson) GetAmount() uint64 {
	return utxo.Amount
	//return utxo.AmountWithInterest
}

//type AssetJson struct {
//	AssetId  string `json:"asset_id"`  // 资产类别
//	UniqueId string `json:"unique_id"` // every token has its unique id
//	ChainId  uint64 `json:"chain_id"`
//}

func ConvertUtxo2Json(outPoint *modules.OutPoint, utxo *modules.Utxo) *UtxoJson {
	scriptStr, _ := tokenengine.Instance.DisasmString(utxo.PkScript)
	json := &UtxoJson{
		TxHash:         outPoint.TxHash.String(),
		MessageIndex:   outPoint.MessageIndex,
		OutIndex:       outPoint.OutIndex,
		Amount:         utxo.Amount,
		Asset:          convertAsset2Json(utxo.Asset),
		PkScriptHex:    hexutil.Encode(utxo.PkScript),
		PkScriptString: scriptStr,
		FlagStatus:     utxo.Flag2Str(),
		LockTime:       utxo.LockTime,
		Time:           time.Unix(utxo.GetTimestamp(), 0),
	}

	json.CoinDays = utxo.GetCoinDays()
	//interest := uint64(0)
	//gasToken := dagconfig.DagConfig.GetGasToken()
	//if gasToken == utxo.Asset.AssetId {
	//	interest = uint64(float64(json.CoinDays) * parameter.CurrentSysParameters.TxCoinDayInterest)
	//}
	//json.AmountWithInterest = json.Amount + interest
	return json
}
func convertAsset2Json(asset *modules.Asset) string {
	return asset.String()
}
