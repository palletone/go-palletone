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

type StxoJson struct {
	TxHash         string    `json:"txid"`          // reference Utxo struct key field
	MessageIndex   uint32    `json:"message_index"` // message index in transaction
	OutIndex       uint32    `json:"out_index"`
	Amount         uint64    `json:"amount"`           // 数量
	Asset          string    `json:"asset"`            // 资产类别
	PkScriptHex    string    `json:"pk_script_hex"`    // 要执行的代码段
	PkScriptString string    `json:"pk_script_string"` // 要执行的代码段
	Time           time.Time `json:"create_time"`      //创建该UTXO的时间（打包到Unit的时间）
	LockTime       uint32    `json:"lock_time"`        //
	SpentByTxId    string    `json:"spent_by_tx_id"`   //
	SpentTime      time.Time `json:"spent_time"`       //
	//CoinDays       uint64    `json:"coin_days"`        //这个Utxo存在多少天了。用于计算利息
	//AmountWithInterest uint64    `json:"amount_with_interest"` //包含利息后的金额
}

func ConvertStxo2Json(outPoint *modules.OutPoint, stxo *modules.Stxo) *StxoJson {
	scriptStr, _ := tokenengine.Instance.DisasmString(stxo.PkScript)
	json := &StxoJson{
		TxHash:         outPoint.TxHash.String(),
		MessageIndex:   outPoint.MessageIndex,
		OutIndex:       outPoint.OutIndex,
		Amount:         stxo.Amount,
		Asset:          convertAsset2Json(stxo.Asset),
		PkScriptHex:    hexutil.Encode(stxo.PkScript),
		PkScriptString: scriptStr,

		LockTime:    stxo.LockTime,
		Time:        time.Unix(int64(stxo.Timestamp), 0),
		SpentByTxId: stxo.SpentByTxId.String(),
		SpentTime:   time.Unix(int64(stxo.SpentTime), 0),
	}

	return json
}
