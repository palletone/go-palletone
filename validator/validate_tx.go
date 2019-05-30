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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

/**
验证某个交易，tx具有以下规则：
Tx的第一条Msg必须是Payment
如果有ContractInvokeRequest，那么要么：
	1.	ContractInvoke不存在（这是一个Request）
	2.	ContractInvoke必然在Request的下面，不可能在Request的上面
	3.  不是Coinbase的情况下，创币PaymentMessage必须在Request下面，并由系统合约创建
	4.  如果是系统合约的请求和结果，必须重新运行合约，保证结果一致
To validate one transaction
如果isFullTx为false，意味着这个Tx还没有被陪审团处理完，所以结果部分的Payment不验证
*/
func (validate *Validate) validateTx(tx *modules.Transaction, isFullTx bool, unitTime int64) (ValidationCode, *modules.AmountAsset) {
	if len(tx.TxMessages) == 0 {
		return TxValidationCode_INVALID_MSG, nil
	}
	isOrphanTx := false
	if tx.TxMessages[0].App != modules.APP_PAYMENT { // 交易费
		return TxValidationCode_INVALID_MSG, nil
	}
	txFeePass, txFee := validate.validateTxFee(tx, unitTime)
	if !txFeePass {
		return TxValidationCode_INVALID_FEE, nil
	}
	hasRequestMsg := false
	requestMsgIndex:=9999
	isSysContractCall := false
	usedUtxo := make(map[string]bool) //Cached all used utxo in this tx
	for msgIdx, msg := range tx.TxMessages {
		// check message type and payload
		if !validateMessageType(msg.App, msg.Payload) {
			return TxValidationCode_UNKNOWN_TX_TYPE, nil
		}
		// validate tx size
		if tx.Size().Float64() > float64(modules.TX_MAXSIZE) {
			log.Debug("Tx size is to big.")
			return TxValidationCode_NOT_COMPARE_SIZE, nil
		}

		// validate transaction signature
		if validateTxSignature(tx) == false {
			return TxValidationCode_BAD_CREATOR_SIGNATURE, nil
		}
		// validate every type payload
		switch msg.App {
		case modules.APP_PAYMENT:
			payment, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				return TxValidationCode_INVALID_PAYMMENTLOAD, nil
			}
			//如果是合约执行结果中的Payment，只有是完整交易的情况下才检查解锁脚本
			if msgIdx>requestMsgIndex && !isFullTx{
				log.Debugf("Tx reqid[%s] is processing tx, don't need validate result payment",tx.RequestHash().String())
			}else {
				validateCode := validate.validatePaymentPayload(tx, msgIdx, payment, usedUtxo)
				if validateCode != TxValidationCode_VALID {
					if validateCode == TxValidationCode_ORPHAN {
						isOrphanTx = true
					} else {
						return validateCode, nil
					}
				}
				//检查一个Tx是否包含了发币的Payment，如果有，那么检查是否是系统合约调用的结果
				if msgIdx != 0 && payment.IsCoinbase() && !isSysContractCall {
					return TxValidationCode_INVALID_COINBASE, nil
				}
			}
		case modules.APP_CONTRACT_TPL:
			payload, _ := msg.Payload.(*modules.ContractTplPayload)
			validateCode := validate.validateContractTplPayload(payload)
			if validateCode != TxValidationCode_VALID {
				return validateCode, nil
			}
		case modules.APP_CONTRACT_DEPLOY:
			payload, _ := msg.Payload.(*modules.ContractDeployPayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, nil
			}
		case modules.APP_CONTRACT_INVOKE:
			payload, _ := msg.Payload.(*modules.ContractInvokePayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, nil
			}
		case modules.APP_CONTRACT_TPL_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, nil
			}
			hasRequestMsg = true
			requestMsgIndex=msgIdx
			payload, _ := msg.Payload.(*modules.ContractInstallRequestPayload)
			if payload.TplName == "" || payload.Path == "" || payload.Version == "" {
				return TxValidationCode_INVALID_CONTRACT, nil
			}

		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, nil
			}
			hasRequestMsg = true
			requestMsgIndex=msgIdx
			// 参数临界值验证
			payload, _ := msg.Payload.(*modules.ContractDeployRequestPayload)
			if len(payload.TplId) == 0 || payload.Timeout < 0 {
				return TxValidationCode_INVALID_CONTRACT, nil
			}

			validateCode := validate.validateContractdeploy(payload.TplId)
			if validateCode != TxValidationCode_VALID {
				return validateCode, nil
			}

		case modules.APP_CONTRACT_INVOKE_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, nil
			}
			hasRequestMsg = true
			requestMsgIndex=msgIdx
			payload, _ := msg.Payload.(*modules.ContractInvokeRequestPayload)
			// 验证ContractId有效性
			if len(payload.ContractId) <= 0 {
				return TxValidationCode_INVALID_CONTRACT, nil
			}
			contractId := payload.ContractId
			if common.IsSystemContractAddress(contractId) {
				isSysContractCall = true
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			payload, _ := msg.Payload.(*modules.ContractStopRequestPayload)
			if len(payload.ContractId) == 0 {
				return TxValidationCode_INVALID_CONTRACT, nil
			}
			// 验证ContractId有效性
			if len(payload.ContractId) <= 0 {
				return TxValidationCode_INVALID_CONTRACT, nil
			}
			requestMsgIndex=msgIdx
		case modules.APP_CONTRACT_STOP:
			payload, _ := msg.Payload.(*modules.ContractStopPayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, nil
			}
		case modules.APP_SIGNATURE:
			// 签名验证
			payload, _ := msg.Payload.(*modules.SignaturePayload)
			validateCode := validate.validateContractSignature(payload.Signatures[:], tx)
			if validateCode != TxValidationCode_VALID {
				return validateCode, nil
			}

			//case modules.APP_CONFIG:
		case modules.APP_DATA:
			payload, _ := msg.Payload.(*modules.DataPayload)
			validateCode := validate.validateDataPayload(payload)
			if validateCode != TxValidationCode_VALID {
				return validateCode, nil
			}

		case modules.APP_ACCOUNT_UPDATE:

		default:
			return TxValidationCode_UNKNOWN_TX_TYPE, nil
		}
	}
	if isOrphanTx {
		return TxValidationCode_ORPHAN, nil
	}
	return TxValidationCode_VALID, txFee
}

func (validate *Validate) validateTxFee(tx *modules.Transaction, unitTime int64) (bool, *modules.AmountAsset) {
	if validate.utxoquery == nil {
		log.Warn("Cannot validate tx fee, your validate utxoquery not set")
		return true, nil
	}
	fee, err := tx.GetTxFee(validate.utxoquery.GetUtxoEntry, unitTime)
	if err != nil {
		log.Warn("compute tx fee error: " + err.Error())
		return false, nil
	}
	assetId := dagconfig.DagConfig.GetGasToken()
	minFee := &modules.AmountAsset{Amount: 0, Asset: assetId.ToAsset()}
	if validate.statequery != nil {
		minFee, err = validate.statequery.GetMinFee()
	}
	if minFee.Amount > 0 { //需要验证最小手续费
		if fee.Asset.String() != minFee.Asset.String() || fee.Amount < minFee.Amount {
			return false, fee
		}
	}
	return true, fee
}

/**
检查message的app与payload是否一致
check messaage 'app' consistent with payload type
*/
func validateMessageType(app modules.MessageType, payload interface{}) bool {
	switch t := payload.(type) {
	case *modules.PaymentPayload:
		if app == modules.APP_PAYMENT {
			return true
		}
	case *modules.ContractTplPayload:
		if app == modules.APP_CONTRACT_TPL {
			return true
		}
	case *modules.ContractDeployPayload:
		if app == modules.APP_CONTRACT_DEPLOY {
			return true
		}
	case *modules.ContractInvokeRequestPayload:
		if app == modules.APP_CONTRACT_INVOKE_REQUEST {
			return true
		}
	case *modules.ContractInvokePayload:
		if app == modules.APP_CONTRACT_INVOKE {
			return true
		}
	case *modules.SignaturePayload:
		if app == modules.APP_SIGNATURE {
			return true
		}
	case *modules.DataPayload:
		if app == modules.APP_DATA {
			return true
		}
	case *modules.AccountStateUpdatePayload:
		if app == modules.APP_ACCOUNT_UPDATE {
			return true
		}
	case *modules.ContractDeployRequestPayload:
		if app == modules.APP_CONTRACT_DEPLOY_REQUEST {
			return true
		}
	case *modules.ContractInstallRequestPayload:
		if app == modules.APP_CONTRACT_TPL_REQUEST {
			return true
		}
	case *modules.ContractStopRequestPayload:
		if app == modules.APP_CONTRACT_STOP_REQUEST {
			return true
		}
	case *modules.ContractStopPayload:
		if app == modules.APP_CONTRACT_STOP {
			return true
		}

	default:
		log.Debug("The payload of message type is unexpected. ", "payload_type", t, "app type", app)
		return false
	}
	return false
}
