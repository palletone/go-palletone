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

package validator

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

//验证一个Payment
//Validate a payment message
//1. Amount correct
//2. Asset must be equal
//3. Unlock correct
func (validate *Validate) validatePaymentPayload(payment *modules.PaymentPayload, isCoinbase bool) ValidationCode {
	// check locktime

	// TODO coinbase 交易的inputs是null.
	// if len(payment.Inputs) <= 0 {
	// 	log.Error("payment input is null.", "payment.input", payment.Inputs)
	// 	return TxValidationCode_INVALID_PAYMMENT_INPUT
	// }

	if !isCoinbase {
		for _, in := range payment.Inputs {
			// checkout input
			if in == nil || in.PreviousOutPoint == nil {
				log.Error("payment input is null.", "payment.input", payment.Inputs)
				return TxValidationCode_INVALID_PAYMMENT_INPUT
			}
			// 合约创币后同步到mediator的utxo验证不通过,在创币后需要先将创币的utxo同步到所有mediator节点。
			if utxo, err := validate.utxoquery.GetUtxoEntry(in.PreviousOutPoint); utxo == nil || err != nil {
				return TxValidationCode_INVALID_OUTPOINT
			}
			// check SignatureScript
		}
	}

	if len(payment.Outputs) <= 0 {
		log.Error("payment output is null.", "payment.output", payment.Outputs)
		return TxValidationCode_INVALID_PAYMMENT_OUTPUT
	}
	//Check coinbase payment
	//rule:
	//	1. all outputs have same asset
	asset0 := payment.Outputs[0].Asset
	for _, out := range payment.Outputs {
		if !asset0.IsSimilar(out.Asset) {
			return TxValidationCode_INVALID_ASSET
		}
	}

	for _, out := range payment.Outputs {
		// // checkout output
		// if i < 1 {
		// 	if !out.Asset.IsSimilar(modules.NewPTNAsset()) {
		// 		return TxValidationCode_INVALID_ASSET
		// 	}
		// 	// log.Debug("validation succeed！")
		// 	continue // asset = out.Asset
		// } else {
		// 	if out.Asset == nil {
		// 		return TxValidationCode_INVALID_ASSET
		// 	}
		// 	if !out.Asset.IsSimilar(payment.Outputs[i-1].Asset) {
		// 		return TxValidationCode_INVALID_ASSET
		// 	}
		// }
		if out.Value <= 0 || out.Value > 100000000000000000 {
			log.Debug("The OutPut value is :", "amount", out.Value)
			return TxValidationCode_INVALID_AMOUNT
		}
	}
	return TxValidationCode_VALID
}
