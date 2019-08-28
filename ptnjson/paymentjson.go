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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

type PaymentJson struct {
	Inputs   []*InputJson  `json:"inputs"`
	Outputs  []*OutputJson `json:"outputs"`
	LockTime uint32        `json:"locktime"`
	Number   int           `json:"number"`
}
type InputJson struct {
	TxHash       string `json:"txid"`          // reference Utxo struct key field
	MessageIndex uint32 `json:"message_index"` // message index in transaction
	OutIndex     uint32 `json:"out_index"`
	UnlockScript string `json:"unlock_script"`
	//UnlockScriptHex string `json:"unlock_script_hex"`
	FromAddress string `json:"from_address"`
}
type OutputJson struct {
	Amount     uint64 `json:"amount"`
	Asset      string `json:"asset"`
	ToAddress  string `json:"to_address"`
	LockScript string `json:"lock_script"`
	//LockScriptHex string `json:"lock_script_hex"`
}
type OutPointJson struct {
	TxHashHex    string `json:"txhash"`        // reference Utxo struct key field
	MessageIndex uint32 `json:"message_index"` // message index in transaction
	OutIndex     uint32 `json:"out_index"`
}

func ConvertPayment2JsonIncludeFromAddr(payment *modules.PaymentPayload, utxoQuery modules.QueryUtxoFunc) *PaymentJson {
	paymentJson := ConvertPayment2Json(payment)
	for _, input := range paymentJson.Inputs {
		if input.TxHash != "" {
			utxo, err := utxoQuery(modules.NewOutPoint(common.HexToHash(input.TxHash), input.MessageIndex, input.OutIndex))
			if err != nil {
				log.Warnf("Query utxo error:%s", err.Error())
			} else {
				addr, _ := tokenengine.Instance.GetAddressFromScript(utxo.PkScript)
				input.FromAddress = addr.String()
			}
		}
	}
	return paymentJson
}
func ConvertPayment2Json(payment *modules.PaymentPayload) *PaymentJson {
	json := &PaymentJson{}
	json.LockTime = payment.LockTime
	json.Inputs = []*InputJson{}
	json.Outputs = []*OutputJson{}
	if len(payment.Inputs) > 0 {
		for _, in := range payment.Inputs {
			// @jay :genesis or coinbase unit occurred nil error.
			var hstr string
			var mindex uint32
			var outindex uint32
			if in.PreviousOutPoint != nil {
				hstr = in.PreviousOutPoint.TxHash.String()
				mindex = in.PreviousOutPoint.MessageIndex
				outindex = in.PreviousOutPoint.OutIndex
			}
			unlock := ""
			if in.SignatureScript != nil {
				unlock, _ = tokenengine.Instance.DisasmString(in.SignatureScript)
			}
			input := &InputJson{TxHash: hstr, MessageIndex: mindex, OutIndex: outindex, UnlockScript: unlock}
			json.Inputs = append(json.Inputs, input)

		}
	}

	for _, out := range payment.Outputs {
		addr, _ := tokenengine.Instance.GetAddressFromScript(out.PkScript)
		lock, _ := tokenengine.Instance.DisasmString(out.PkScript)
		output := &OutputJson{Amount: out.Value, Asset: out.Asset.String(), ToAddress: addr.String(), LockScript: lock}
		json.Outputs = append(json.Outputs, output)
	}
	return json
}
func ConvertJson2Payment(json *PaymentJson) *modules.PaymentPayload {
	payment := &modules.PaymentPayload{}
	payment.LockTime = json.LockTime
	for _, in := range json.Inputs {
		hash := common.HexToHash(in.TxHash)
		outPoint := modules.NewOutPoint(hash, in.MessageIndex, in.OutIndex)
		input := modules.NewTxIn(outPoint, []byte{})
		payment.AddTxIn(input)
	}
	for _, out := range json.Outputs {
		addr, _ := common.StringToAddress(out.ToAddress)
		lockScript := tokenengine.Instance.GenerateLockScript(addr)
		asset := modules.Asset{}
		asset.SetString(out.Asset)
		output := modules.NewTxOut(out.Amount, lockScript, &asset)
		payment.AddTxOut(output)
	}
	return payment
}

func ConvertOutPoint2Json(outpoint *modules.OutPoint) *OutPointJson {
	return &OutPointJson{
		TxHashHex:    outpoint.TxHash.String(),
		MessageIndex: outpoint.MessageIndex,
		OutIndex:     outpoint.OutIndex,
	}
}
