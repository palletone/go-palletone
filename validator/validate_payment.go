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
	"math"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

//验证一个Payment
//Validate a payment message
//1. Amount correct
//2. Asset must be equal
//3. Unlock correct
//4.Blacklist check, fromAddr toAddr must not in blacklist
func (validate *Validate) validatePaymentPayload(tx *modules.Transaction, msgIdx int,
	payment *modules.PaymentPayload, usedUtxo map[string]bool) ValidationCode {
	txId := tx.Hash()
	//if payment.LockTime > 0 {
	//	// TODO check locktime
	//}
	gasToken := dagconfig.DagConfig.GetGasToken()
	blacklistAddress := validate.getBlacklistAddress()
	log.DebugDynamic(func() string {
		data, _ := json.Marshal(blacklistAddress)
		return "Blacklist:" + string(data)
	})
	var asset *modules.Asset
	totalInput := uint64(0)
	isInputnil := false
	if len(payment.Inputs) > 1000 {
		//太多的Input会导致签名验证超时
		return TxValidationCode_INVALID_PAYMMENT_INPUT_COUNT
	}
	if len(payment.Inputs) == 0 {
		// if payment.Outputs[0].Asset.AssetId.Equal(gasToken) {
		// 	return TxValidationCode_INVALID_PAYMMENT_INPUT
		// }
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
		utxoScriptMap := make(map[string][]byte)
		for _, in := range payment.Inputs {
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
			var utxo *modules.Utxo
			var err error
			if in.PreviousOutPoint.TxHash.IsSelfHash() {
				output := tx.TxMessages[in.PreviousOutPoint.MessageIndex].Payload.(*modules.PaymentPayload).Outputs[in.PreviousOutPoint.OutIndex]
				utxo = &modules.Utxo{
					Amount:    output.Value,
					Asset:     output.Asset,
					PkScript:  output.PkScript,
					LockTime:  0,
					Timestamp: 0,
				}
			} else {

				utxo, err = validate.utxoquery.GetUtxoEntry(in.PreviousOutPoint)
				if utxo == nil || err != nil {
					//找不到对应的UTXO，去IsSpent再找一下
					stxo, _ := validate.utxoquery.GetStxoEntry(in.PreviousOutPoint)
					if stxo != nil && stxo.SpentByTxId != txId {
						log.Errorf("Utxo[%s] spent by tx[%s]", in.PreviousOutPoint.String(), stxo.SpentByTxId.String())
						return TxValidationCode_INVALID_DOUBLE_SPEND
					}
					//IsSpent找不到，说明是孤儿
					return TxValidationCode_ORPHAN
				}
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
				if !statusValid && asset.AssetId != gasToken {
					ret := validate.checkTokenStatus(asset)
					if TxValidationCode_VALID != ret {
						return ret
					}
					statusValid = true
				}
			}
			fromAddr, _ := validate.tokenEngine.GetAddressFromScript(utxo.PkScript)
			if _, isIn := blacklistAddress[fromAddr]; isIn {
				log.Infof("address[%s] is in blacklist", fromAddr.String())
				return TxValidationCode_ADDRESS_IN_BLACKLIST
			}

			totalInput += utxo.Amount
			// check SignatureScript
			utxoScriptMap[in.PreviousOutPoint.String()] = utxo.PkScript

		}
		t1 := time.Now()
		err := validate.tokenEngine.ScriptValidate1Msg(utxoScriptMap, validate.pickJuryFn, txForSign, msgIdx)
		if err != nil {
			return TxValidationCode_INVALID_PAYMMENT_INPUT
		} else {
			log.Debugf("Unlock script validated! tx[%s],%d, spend time:%s", tx.Hash().String(), msgIdx, time.Since(t1))
		}
	}

	totalOutput := uint64(0)
	//Check payment
	//rule:
	//	1. all outputs have same asset id
	if len(payment.Outputs) > 0 {
		asset0 := payment.Outputs[0].Asset
		for _, out := range payment.Outputs {
			if isInputnil { //Input为空，可能是721的创币，所以只检查AssetId相同，不检查UniqueId
				if !asset0.IsSameAssetId(out.Asset) {
					return TxValidationCode_INVALID_ASSET
				}
			} else { //Input不为空，则Input和Output必须是同样的Asset
				if !asset.IsSimilar(out.Asset) { //Input Output asset mustbe same
					return TxValidationCode_INVALID_ASSET
				}
			}
			totalOutput += out.Value
			if totalOutput < out.Value || out.Value == 0 { //big number overflow
				return TxValidationCode_INVALID_AMOUNT
			}
			toAddr, _ := validate.tokenEngine.GetAddressFromScript(out.PkScript)
			if _, isIn := blacklistAddress[toAddr]; isIn {
				log.Infof("address[%s] is in blacklist", toAddr.String())
				return TxValidationCode_ADDRESS_IN_BLACKLIST
			}
		}

		if !isInputnil {
			if msgIdx != 0 && totalOutput > totalInput { //相当于进行了增发
				return TxValidationCode_INVALID_AMOUNT
			}
		}
	}
	return TxValidationCode_VALID
}
func (validate *Validate) pickJuryFn(contractAddr common.Address) ([]byte, error) {
	log.Debugf("Try to pickup jury for address:%s", contractAddr.String())
	var redeemScript []byte

	if !contractAddr.IsSystemContractAddress() {
		jury, err := validate.statequery.GetContractJury(contractAddr.Bytes())
		if err != nil {
			log.Errorf("Cannot get contract[%s] jury", contractAddr.String())
			return nil, errors.New("Cannot get contract jury")
		}
		redeemScript = validate.generateJuryRedeemScript(jury)
		log.DebugDynamic(func() string {
			redeemStr, _ := validate.tokenEngine.DisasmString(redeemScript)
			return "Generate RedeemScript: " + redeemStr
		})
	}

	return redeemScript, nil
}

//检查转移的Token是否已经冻结，冻结的Token不能再转移
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

//var BlacklistAddress=[]byte("BlacklistAddress")
func (validate *Validate) getBlacklistAddress() map[common.Address]bool {
	result := make(map[common.Address]bool)
	if validate.statequery==nil{
		log.Warn("don't set statequery, blacklist is empty")
		return result
	}

	//data,err:= validate.cache.cache.Get(BlacklistAddress)
	//if err==nil{
	// 	addresses,_,_:=	validate.statequery.GetBlacklistAddress()
	// 	data,_=rlp.EncodeToBytes(addresses)
	// 	validate.cache.cache.Set(BlacklistAddress,data,60)
	//}
	//addresses:=[]common.Address{}
	//rlp.DecodeBytes(data,&addresses)
	addresses, _, _ := validate.statequery.GetBlacklistAddress()

	for _, addr := range addresses {
		result[addr] = true
	}
	return result
}

func (validate *Validate) generateJuryRedeemScript(jury *modules.ElectionNode) []byte {
	if jury == nil {
		return nil
	}
	count := len(jury.EleList)
	needed := byte(math.Ceil((float64(count)*2 + 1) / 3))
	pubKeys := [][]byte{}
	for _, jurior := range jury.EleList {
		pubKeys = append(pubKeys, jurior.PublicKey)
	}
	return validate.tokenEngine.GenerateRedeemScript(needed, pubKeys)
}
