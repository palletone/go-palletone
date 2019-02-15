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
	"github.com/palletone/go-palletone/tokenengine"
)

//Coinbase可以没有输入，就算有输入也没有Preoutpoint
func (validate *Validate) validateCoinbase(payment *modules.PaymentPayload) ValidationCode {
	return TxValidationCode_VALID
}

//验证一个Payment
//Validate a payment message
//1. Amount correct
//2. Asset must be equal
//3. Unlock correct
func (validate *Validate) validatePaymentPayload(tx *modules.Transaction, msgIdx int, payment *modules.PaymentPayload, isCoinbase bool) ValidationCode {

	if isCoinbase {
		return validate.validateCoinbase(payment)
	}
	if payment.LockTime > 0 {
		// TODO check locktime
	}

	var asset *modules.Asset
	totalInput := uint64(0)
	isInputnil := false
	if len(payment.Inputs) == 0 {
		if payment.Outputs[0].Asset.AssetId == modules.PTNCOIN {
			return TxValidationCode_INVALID_PAYMMENT_INPUT
		}
		isInputnil = true
	} else {
		invokeReqMsgIdx := tx.GetContractInvokeReqMsgIdx()
		txForSign := tx
		if msgIdx < invokeReqMsgIdx {
			txForSign = tx.GetRequestTx()
		}
		utxos := []*modules.Utxo{}
		for inputIdx, in := range payment.Inputs {
			// checkout input
			if in == nil || in.PreviousOutPoint == nil {
				log.Error("payment input is null.", "payment.input", payment.Inputs)
				return TxValidationCode_INVALID_PAYMMENT_INPUT
			}
			// 合约创币后同步到mediator的utxo验证不通过,在创币后需要先将创币的utxo同步到所有mediator节点。
			utxo, err := validate.utxoquery.GetUtxoEntry(in.PreviousOutPoint)
			if utxo == nil || err != nil {
				return TxValidationCode_INVALID_OUTPOINT
			}
			utxos = append(utxos, utxo)
			if asset == nil {
				asset = utxo.Asset
			} else {
				//input asset must be same
				if !asset.IsSimilar(utxo.Asset) {
					return TxValidationCode_INVALID_ASSET
				}
			}
			totalInput += utxo.Amount
			// check SignatureScript
			err = tokenengine.ScriptValidate(utxo.PkScript, nil, txForSign, msgIdx, inputIdx)
			if err != nil {
				log.Infof("Unlock script validate fail,tx[%s],MsgIdx[%d],In[%d],unlockScript:%x,utxoScript:%x",
					tx.Hash().String(), msgIdx, inputIdx, in.SignatureScript, utxo.PkScript)
				return TxValidationCode_INVALID_PAYMMENT_INPUT
			} else {
				log.Debugf("Unlock script validated! tx[%s],%d,%d", tx.Hash().String(), msgIdx, inputIdx)
			}
		}
	}

	if len(payment.Outputs) == 0 {
		log.Error("payment output is null.", "payment.output", payment.Outputs)
		return TxValidationCode_INVALID_PAYMMENT_OUTPUT
	}
	totalOutput := uint64(0)
	//Check payment
	//rule:
	//	1. all outputs have same asset
	asset0 := payment.Outputs[0].Asset
	for _, out := range payment.Outputs {
		if !asset0.IsSimilar(out.Asset) {
			return TxValidationCode_INVALID_ASSET
		}
		totalOutput += out.Value
		if totalOutput < out.Value || out.Value == 0 { //big number overflow
			return TxValidationCode_INVALID_AMOUNT
		}
	}
	if !isInputnil {
		//Input Output asset mustbe same
		if !asset.IsSimilar(asset0) {
			return TxValidationCode_INVALID_ASSET
		}
		if totalOutput > totalInput { //相当于手续费为负数
			return TxValidationCode_INVALID_AMOUNT
		}
	}
	return TxValidationCode_VALID
}
