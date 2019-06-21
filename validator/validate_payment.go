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
	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"math"
)

//验证一个Payment
//Validate a payment message
//1. Amount correct
//2. Asset must be equal
//3. Unlock correct
func (validate *Validate) validatePaymentPayload(tx *modules.Transaction, msgIdx int,
	payment *modules.PaymentPayload, usedUtxo map[string]bool) ValidationCode {

	if payment.LockTime > 0 {
		// TODO check locktime
	}
	gasToken := dagconfig.DagConfig.GetGasToken()
	var asset *modules.Asset
	totalInput := uint64(0)
	isInputnil := false
	if len(payment.Inputs) == 0 {
		if payment.Outputs[0].Asset.AssetId.Equal(gasToken) {
			return TxValidationCode_INVALID_PAYMMENT_INPUT
		}
		isInputnil = true
	} else {
		invokeReqMsgIdx := tx.GetContractInvokeReqMsgIdx()
		txForSign := tx
		if msgIdx < invokeReqMsgIdx {
			txForSign = tx.GetRequestTx()
			log.Debugf("msgIdx %d, GetRequestTx 1", msgIdx)
		} else if invokeReqMsgIdx > 0 && msgIdx > invokeReqMsgIdx {
			txForSign = tx.GetResultTx()
			log.Debugf("msgIdx %d, GetResultTx 1", msgIdx)
		}

		statusValid := false
		for inputIdx, in := range payment.Inputs {
			// checkout input
			if in == nil || in.PreviousOutPoint == nil {
				log.Error("payment input is null.", "payment.input", payment.Inputs)
				return TxValidationCode_INVALID_PAYMMENT_INPUT
			}
			usedUtxoKey := in.PreviousOutPoint.String()
			if _, exist := usedUtxo[usedUtxoKey]; exist {
				log.Error("double spend utxo:", usedUtxoKey)
				return TxValidationCode_INVALID_DOUBLE_SPEND
			}
			usedUtxo[usedUtxoKey] = true
			// 合约创币后同步到mediator的utxo验证不通过,在创币后需要先将创币的utxo同步到所有mediator节点。
			utxo, err := validate.utxoquery.GetUtxoEntry(in.PreviousOutPoint)
			if utxo == nil || err != nil {
				//找不到对应的UTXO，应该是孤儿交易
				return TxValidationCode_ORPHAN
			}
			if utxo.IsSpent() {
				return TxValidationCode_INVALID_DOUBLE_SPEND
			}
			if asset == nil {
				asset = utxo.Asset
			} else {
				//input asset must be same
				if !asset.IsSimilar(utxo.Asset) {
					return TxValidationCode_INVALID_ASSET
				}
			}

			//check token status
			if msgIdx != 0 {
				if !statusValid && asset.AssetId != modules.PTNCOIN {
					ret := validate.checkTokenStatus(asset)
					if TxValidationCode_VALID != ret {
						return ret
					}
					statusValid = true
				}
			}
			totalInput += utxo.Amount
			// check SignatureScript

			pickJuryFn := func(contractAddr common.Address) ([]byte, error) {
				log.Debugf("Try to pickup jury for address:%s", contractAddr.String())
				var redeemScript []byte

				if !contractAddr.IsSystemContractAddress() {
					jury, err := validate.statequery.GetContractJury(contractAddr.Bytes())
					if err != nil {
						log.Errorf("Cannot get contract[%s] jury", contractAddr.String())
						return nil, errors.New("Cannot get contract jury")
					}
					redeemScript, _ = generateJuryRedeemScript(jury)
					log.DebugDynamic(func() string {
						redeemStr, _ := tokenengine.DisasmString(redeemScript)
						return "Generate RedeemScript: " + redeemStr
					})
				}

				return redeemScript, nil
			}
			err = tokenengine.ScriptValidate(utxo.PkScript, pickJuryFn, txForSign, msgIdx, inputIdx)
			if err != nil {

				log.Warnf("Unlock script validate fail,tx[%s],MsgIdx[%d],In[%d],unlockScript:%x,utxoScript:%x",
					tx.Hash().String(), msgIdx, inputIdx, in.SignatureScript, utxo.PkScript)
				txjson, _ := tx.MarshalJSON()
				rlpdata, _ := rlp.EncodeToBytes(tx)
				log.Debugf("Tx for help debug: json: %s ,rlp: %x", string(txjson), rlpdata)
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
	//	1. all outputs have same asset id
	asset0 := payment.Outputs[0].Asset
	for _, out := range payment.Outputs {
		if !asset0.IsSameAssetId(out.Asset) {
			return TxValidationCode_INVALID_ASSET
		}
		totalOutput += out.Value
		if totalOutput < out.Value || out.Value == 0 { //big number overflow
			return TxValidationCode_INVALID_AMOUNT
		}
	}
	if !isInputnil {
		//Input Output asset mustbe same
		if !asset.IsSameAssetId(asset0) {
			return TxValidationCode_INVALID_ASSET
		}
		//if msgIdx != 0 && totalOutput > totalInput { //相当于进行了增发
		//	return TxValidationCode_INVALID_AMOUNT
		//}
	}
	return TxValidationCode_VALID
}

func (validate *Validate) checkTokenStatus(asset *modules.Asset) ValidationCode {
	globalStateContractId := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	result, _, err := validate.statequery.GetContractState(globalStateContractId, modules.GlobalPrefix+asset.AssetId.GetSymbol())
	if nil != err {
		return TxValidationCode_INVALID_ASSET
	}
	var tokenInfo modules.GlobalTokenInfo
	err = json.Unmarshal(result, &tokenInfo)
	if nil != err {
		return TxValidationCode_INVALID_ASSET
	}
	if tokenInfo.Status != 0 {
		return TxValidationCode_INVALID_TOKEN_STATUS
	}
	return TxValidationCode_VALID
}

func generateJuryRedeemScript(jury []modules.ElectionInf) ([]byte, error) {
	count := len(jury)
	needed := byte(math.Ceil((float64(count)*2 + 1) / 3))
	pubKeys := [][]byte{}
	for _, jurior := range jury {
		pubKeys = append(pubKeys, jurior.PublicKey)
	}
	return tokenengine.GenerateRedeemScript(needed, pubKeys), nil
}
