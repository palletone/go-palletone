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

package walletjson

type PaymentJson struct {
	Inputs  []InputJson  `json:"inputs"`
	Outputs []OutputJson `json:"outputs"`
}
type ProofJson struct {
	Inputs  []InputJson  `json:"inputs"`
	Outputs []OutputJson `json:"outputs"`
	Proof   string       `json:"proof"`
	Extra   string       `json:"extra"`
}
type TxJson struct {
	Payload       []PaymentJson         `json:"payload"`
	InvokeRequest ContractInvokeRequest `json:"invoke_request"`
}
type TxProofJson struct {
	Payload       []ProofJson           `json:"payload"`
	InvokeRequest ContractInvokeRequest `json:"invoke_request"`
}
type ContractInvokeRequest struct {
	ContractAddress string
	Args            []string
}
type InputJson struct {
	TxHash       string `json:"txid"`          // reference Utxo struct key field
	MessageIndex uint32 `json:"message_index"` // message index in transaction
	OutIndex     uint32 `json:"out_index"`
	HashForSign  string `json:"hash"`
	Signature    string `json:"signature"`
}

type OutputJson struct {
	Amount    uint64 `json:"amount"`
	Asset     string `json:"asset"`
	ToAddress string `json:"to_address"`
}
type RawTxjsonGenParams struct {
	Inputs []struct {
		TxHash       string `json:"txid"`
		OutIndex     uint32 `json:"outindex"`
		MessageIndex uint32 `json:"messageindex"`
		HashForSign  string `json:"hash"`
		Signature    string `json:"signature"`
	} `json:"inputs"`
	Outputs []struct {
		Address string `json:"address"`
		Amount  uint64 `json:"amount"`
		Asset   string `json:"asset"`
	} `json:"outputs"`
}

//
//func ConvertPayment2Json(payment *modules.PaymentPayload) PaymentJson {
//	json := PaymentJson{}
//	json.LockTime = payment.LockTime
//	json.Inputs = []InputJson{}
//	json.Outputs = []OutputJson{}
//	if len(payment.Inputs) > 0 {
//		for _, in := range payment.Inputs {
//			// @jay :genesis or coinbase unit occurred nil error.
//			var hstr string
//			var mindex uint32
//			var outindex uint32
//			if in.PreviousOutPoint != nil {
//				hstr = in.PreviousOutPoint.TxHash.String()
//				mindex = in.PreviousOutPoint.MessageIndex
//				outindex = in.PreviousOutPoint.OutIndex
//			}
//			input := InputJson{TxHash: hstr, MessageIndex: mindex, OutIndex: outindex}
//			json.Inputs = append(json.Inputs, input)
//
//		}
//	}
//
//	for _, out := range payment.Outputs {
//		addr, _ := tokenengine.GetAddressFromScript(out.PkScript)
//		output := OutputJson{Amount: out.Value, Asset: out.Asset.String(), ToAddress: addr.String()}
//		json.Outputs = append(json.Outputs, output)
//	}
//	return json
//}
//func ConvertJson2Payment(json *PaymentJson) modules.PaymentPayload {
//	payment := modules.PaymentPayload{}
//	payment.LockTime = json.LockTime
//	for _, in := range json.Inputs {
//		hash, _ := common.NewHashFromStr(in.TxHash)
//		outPoint := modules.NewOutPoint(hash, in.MessageIndex, in.OutIndex)
//		input := modules.NewTxIn(outPoint, []byte{})
//		payment.AddTxIn(input)
//	}
//	for _, out := range json.Outputs {
//		addr, _ := common.StringToAddress(out.ToAddress)
//		lockScript := tokenengine.GenerateLockScript(addr)
//		asset := modules.Asset{}
//		asset.SetString(out.Asset)
//		output := modules.NewTxOut(out.Amount, lockScript, &asset)
//		payment.AddTxOut(output)
//	}
//	return payment
//}
