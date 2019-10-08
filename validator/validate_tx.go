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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
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
func (validate *Validate) validateTx(tx *modules.Transaction, isFullTx bool) (ValidationCode, []*modules.Addition) {
	if len(tx.TxMessages) == 0 {
		return TxValidationCode_INVALID_MSG, nil
	}
	isOrphanTx := false
	if tx.TxMessages[0].App != modules.APP_PAYMENT { // 交易费
		return TxValidationCode_INVALID_MSG, nil
	}
	txFeePass, txFee := validate.validateTxFee(tx)
	if !txFeePass {
		return TxValidationCode_INVALID_FEE, nil
	}
	hasRequestMsg := false
	requestMsgIndex := 9999
	isSysContractCall := false
	usedUtxo := make(map[string]bool) //Cached all used utxo in this tx
	for msgIdx, msg := range tx.TxMessages {
		// check message type and payload
		if !validateMessageType(msg.App, msg.Payload) {
			return TxValidationCode_UNKNOWN_TX_TYPE, txFee
		}
		// validate tx size
		if tx.Size().Float64() > float64(modules.TX_MAXSIZE) {
			log.Debug("Tx size is to big.")
			return TxValidationCode_NOT_COMPARE_SIZE, txFee
		}

		// validate every type payload
		switch msg.App {
		case modules.APP_PAYMENT:
			payment, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				return TxValidationCode_INVALID_PAYMMENTLOAD, txFee
			}
			//如果是合约执行结果中的Payment，只有是完整交易的情况下才检查解锁脚本
			if msgIdx > requestMsgIndex && !isFullTx {
				log.Debugf("Tx reqid[%s] is processing tx, don't need validate result payment", tx.RequestHash().String())
			} else {
				validateCode := validate.validatePaymentPayload(tx, msgIdx, payment, usedUtxo)
				if validateCode != TxValidationCode_VALID {
					if validateCode == TxValidationCode_ORPHAN {
						isOrphanTx = true
					} else {
						return validateCode, txFee
					}
				}
				//检查一个Tx是否包含了发币的Payment，如果有，那么检查是否是系统合约调用的结果
				if msgIdx != 0 && payment.IsCoinbase() && !isSysContractCall {
					log.Error("Invalid Coinbase message")
					return TxValidationCode_INVALID_COINBASE, txFee
				}
			}
		case modules.APP_CONTRACT_TPL:
			payload, _ := msg.Payload.(*modules.ContractTplPayload)
			validateCode := validate.validateContractTplPayload(payload)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_CONTRACT_DEPLOY:
			payload, _ := msg.Payload.(*modules.ContractDeployPayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_CONTRACT_INVOKE:
			payload, _ := msg.Payload.(*modules.ContractInvokePayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_CONTRACT_TPL_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, txFee
			}
			hasRequestMsg = true
			requestMsgIndex = msgIdx
			payload, _ := msg.Payload.(*modules.ContractInstallRequestPayload)
			if payload.TplName == "" || payload.Path == "" || payload.Version == "" {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}

		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, txFee
			}
			hasRequestMsg = true
			requestMsgIndex = msgIdx
			// 参数临界值验证
			payload, _ := msg.Payload.(*modules.ContractDeployRequestPayload)
			if len(payload.TemplateId) == 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}

			validateCode := validate.validateContractdeploy(payload.TemplateId)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}

		case modules.APP_CONTRACT_INVOKE_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, txFee
			}
			hasRequestMsg = true
			requestMsgIndex = msgIdx
			payload, _ := msg.Payload.(*modules.ContractInvokeRequestPayload)
			// 验证ContractId有效性
			if len(payload.ContractId) <= 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			contractId := payload.ContractId
			if common.IsSystemContractAddress(contractId) {
				isSysContractCall = true
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			payload, _ := msg.Payload.(*modules.ContractStopRequestPayload)
			if len(payload.ContractId) == 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			// 验证ContractId有效性
			if len(payload.ContractId) <= 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			requestMsgIndex = msgIdx
		case modules.APP_CONTRACT_STOP:
			payload, _ := msg.Payload.(*modules.ContractStopPayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_SIGNATURE:
			// 签名验证
			payload, _ := msg.Payload.(*modules.SignaturePayload)
			validateCode := validate.validateContractSignature(payload.Signatures[:], tx)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}

		case modules.APP_DATA:
			payload, _ := msg.Payload.(*modules.DataPayload)
			validateCode := validate.validateDataPayload(payload)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}

		case modules.APP_ACCOUNT_UPDATE:
			return validate.validateVoteMediatorTx(msg.Payload), txFee

		default:
			return TxValidationCode_UNKNOWN_TX_TYPE, txFee
		}
	}
	if isOrphanTx {
		return TxValidationCode_ORPHAN, txFee
	}
	return TxValidationCode_VALID, txFee
}

func (v *Validate) validateVoteMediatorTx(payload interface{}) ValidationCode {
	accountUpdate, ok := payload.(*modules.AccountStateUpdatePayload)
	if !ok {
		log.Errorf("tx payload do not match type")
		return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
	}

	for _, writeSet := range accountUpdate.WriteSet {
		if writeSet.Key != constants.VOTED_MEDIATORS {
			continue
		}

		var mediators map[string]bool
		err := json.Unmarshal(writeSet.Value, &mediators)
		if err != nil {
			log.Errorf("writeSet value do not match key")
			return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
		}

		maxMediatorCount := int(v.propquery.GetChainParameters().MaximumMediatorCount)
		mediatorCount := len(mediators)
		if mediatorCount > maxMediatorCount {
			log.Errorf("the total number(%v) of mediators voted exceeds the maximum limit: %v",
				mediatorCount, maxMediatorCount)
			return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
		}

		mp := v.statequery.GetMediators()
		for mediatorStr, ok := range mediators {
			if !ok {
				log.Errorf("the value of map can only be true")
				return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
			}

			mediator, err := common.StringToAddress(mediatorStr)
			if err != nil {
				log.Errorf("invalid account address: %v", mediatorStr)
				return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
			}

			if !mp[mediator] {
				log.Errorf("%v is not mediator", mediatorStr)
				return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
			}
		}
	}

	return TxValidationCode_VALID
}

//验证手续费是否合法，并返回手续费的分配情况
func (validate *Validate) validateTxFee(tx *modules.Transaction) (bool, []*modules.Addition) {
	if validate.utxoquery == nil {
		log.Warn("Cannot validate tx fee, your validate utxoquery not set")
		return true, nil
	}
	feeAllocate, err := tx.GetTxFeeAllocate(validate.utxoquery.GetUtxoEntry,
		validate.tokenEngine.GetScriptSigners, common.Address{})
	if err != nil {
		log.Warn("compute tx fee error: " + err.Error())
		return false, nil
	}
	assetId := dagconfig.DagConfig.GetGasToken()
	minFee := &modules.AmountAsset{Amount: 0, Asset: assetId.ToAsset()}
	if validate.statequery != nil {
		minFee, err = validate.statequery.GetMinFee()
		if err != nil {
			log.Errorf("GetMinFee throw an error:%s", err.Error())
			return true, feeAllocate
		}
	}
	if minFee.Amount > 0 { //需要验证最小手续费
		total := uint64(0)
		var feeAsset *modules.Asset
		for _, a := range feeAllocate {
			total += a.Amount
			feeAsset = a.Asset
		}
		if feeAsset.String() != minFee.Asset.String() || total < minFee.Amount {
			return false, feeAllocate
		}
	}
	return true, feeAllocate
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
func (validate *Validate) validateCoinbase(tx *modules.Transaction, ads []*modules.Addition) ValidationCode {
	contractId := syscontract.CoinbaseContractAddress.Bytes()
	if tx.TxMessages[0].App == modules.APP_PAYMENT { //到达一定高度，Account转UTXO

		//在Coinbase合约的StateDB中保存每个Mediator和Jury的奖励值，
		//key为奖励地址，Value为[]AmountAsset
		//读取之前的奖励统计值
		addrMap, err := validate.statequery.GetContractStatesByPrefix(contractId, constants.RewardAddressPrefix)
		if err != nil {
			return TxValidationCode_STATE_DATA_NOT_FOUND
		}
		rewards := map[common.Address][]modules.AmountAsset{}
		for key, v := range addrMap {
			addr := key[len(constants.RewardAddressPrefix):]
			incomeAddr, _ := common.StringToAddress(addr)
			aa := []modules.AmountAsset{}
			rlp.DecodeBytes(v.Value, &aa)
			if len(aa) > 0 {
				rewards[incomeAddr] = aa
			}
		}
		//附加最新的奖励
		for _, ad := range ads {

			reward, ok := rewards[ad.Addr]
			if !ok {
				reward = []modules.AmountAsset{}
			}
			reward = validate.addIncome(reward, ad.Amount, ad.Asset)
			rewards[ad.Addr] = reward
		}
		//Check payment output is correct
		payment := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
		if !validate.compareRewardAndOutput(rewards, payment.Outputs) {
			log.Errorf("Coinbase tx[%s] Output not match", tx.Hash().String())
			log.DebugDynamic(func() string {
				rjson, _ := json.Marshal(rewards)
				ojson, _ := json.Marshal(payment)
				return fmt.Sprintf("Data for help debug: \r\nRewards:%s \r\nPayment:%s", string(rjson), string(ojson))
			})
			// panic("Coinbase Output not match")
			return TxValidationCode_INVALID_COINBASE
		}
		//Check statedb should clear
		if len(addrMap) > 0 {
			clearStateInvoke := tx.TxMessages[1].Payload.(*modules.ContractInvokePayload)
			if !bytes.Equal(clearStateInvoke.ContractId, contractId) {
				log.Errorf("Coinbase tx[%s] contract id not correct", tx.Hash().String())
				return TxValidationCode_INVALID_COINBASE
			}
			if !validate.compareRewardAndStateClear(rewards, clearStateInvoke.WriteSet) {
				rjson, _ := json.Marshal(rewards)
				ojson, _ := json.Marshal(clearStateInvoke)
				data := fmt.Sprintf("Data for help debug: \r\nRewards:%s \r\nInvoke result:%s", string(rjson), string(ojson))
				log.Errorf("Coinbase tx[%s] Clear statedb not match, detail data:%s",
					tx.Hash().String(), data)
				return TxValidationCode_INVALID_COINBASE
			}
		}
		return TxValidationCode_VALID
	}
	if tx.TxMessages[0].App == modules.APP_CONTRACT_INVOKE { //Account模型记账
		//传入的ads,集合StateDB的历史，生成新的Reward记录
		rewards := map[common.Address][]modules.AmountAsset{}
		for _, v := range ads {
			key := constants.RewardAddressPrefix + v.Addr.String()
			data, _, err := validate.statequery.GetContractState(contractId, key)
			income := []modules.AmountAsset{}
			if err == nil { //之前有奖励
				rlp.DecodeBytes(data, &income)
			}
			v1 := *v
			log.DebugDynamic(func() string {
				data, _ := json.Marshal(income)
				return v1.Addr.String() + " Coinbase History reward:" + string(data)
			})
			log.Debugf("Add reward %d %s to %s", v.Amount, v.Asset.String(), v.Addr.String())
			newValue := validate.addIncome(income, v.Amount, v.Asset)
			rewards[v.Addr] = newValue
		}
		//比对reward和writeset是否一致
		invoke := tx.TxMessages[0].Payload.(*modules.ContractInvokePayload)
		if !bytes.Equal(invoke.ContractId, contractId) {
			log.Errorf("Coinbase tx[%s] contract id not correct", tx.Hash().String())
			return TxValidationCode_INVALID_COINBASE
		}
		if validate.compareRewardAndWriteset(rewards, invoke.WriteSet) {
			return TxValidationCode_VALID
		} else {
			rjson, _ := json.Marshal(rewards)
			ojson, _ := json.Marshal(invoke)
			debugData := fmt.Sprintf("Data for help debug: \r\nRewards:%s \r\nInvoke result:%s", string(rjson), string(ojson))

			log.Errorf("Coinbase tx[%s] contract write set not correct, %s",
				tx.Hash().String(), debugData)
			return TxValidationCode_INVALID_COINBASE
		}
	}
	return TxValidationCode_VALID
}

func (validate *Validate) compareRewardAndOutput(rewards map[common.Address][]modules.AmountAsset, outputs []*modules.Output) bool {
	comparedCount := 0
	for addr, reward := range rewards {
		if validate.rewardExistInOutputs(addr, reward, outputs) {
			comparedCount++
		} else {
			return false
		}

	}
	return comparedCount == len(outputs)
	// if comparedCount != len(outputs) {
	// 	return false
	// }
	// return true
}
func (validate *Validate) rewardExistInOutputs(addr common.Address, aa []modules.AmountAsset, outputs []*modules.Output) bool {
	for _, out := range outputs {
		outAddr, _ := validate.tokenEngine.GetAddressFromScript(out.PkScript)
		if outAddr.Equal(addr) {

			for _, a := range aa {

				if a.Asset.Equal(out.Asset) && a.Amount != out.Value {
					return false
				}

			}
		}
	}
	return true
}
func (validate *Validate) compareRewardAndStateClear(rewards map[common.Address][]modules.AmountAsset, writeset []modules.ContractWriteSet) bool {
	comparedCount := 0
	empty, _ := rlp.EncodeToBytes([]modules.AmountAsset{})
	for addr := range rewards {
		addrKey := constants.RewardAddressPrefix + addr.String()
		for _, w := range writeset {
			// if !w.IsDelete {
			// 	return false
			// }
			if w.Key == addrKey && bytes.Equal(w.Value, empty) {
				comparedCount++
			}
		}

	}
	//return comparedCount == len(writeset)
	if comparedCount != len(rewards) { //所有的Reward的状态数据库被清空
		log.Warnf("write set comparedCount:%d clean count:%d", comparedCount, len(rewards))
		return false
	}
	return true
}
func (validate *Validate) compareRewardAndWriteset(rewards map[common.Address][]modules.AmountAsset, writeset []modules.ContractWriteSet) bool {
	comparedCount := 0
	for addr, reward := range rewards {

		if validate.rewardExist(addr, reward, writeset) {
			comparedCount++
		} else {

			return false
		}

	}
	return comparedCount == len(rewards)
	// if comparedCount != len(rewards) { //所有的Reward的状态数据库被清空
	// 	return false
	// }
	// return true
}
func (validate *Validate) rewardExist(addr common.Address, aa []modules.AmountAsset, writeset []modules.ContractWriteSet) bool {
	for _, w := range writeset {
		if w.Key == constants.RewardAddressPrefix+addr.String() {
			dbAa := []modules.AmountAsset{}
			err := rlp.DecodeBytes(w.Value, &dbAa)
			if err != nil {
				log.Error("Decode rlp data to []modules.AmountAsset error")
				return false
			}
			for _, a := range aa {
				for _, b := range dbAa {
					if a.Asset.Equal(b.Asset) && a.Amount != b.Amount {
						a1 := a
						b1 := b
						log.DebugDynamic(func() string {
							data, _ := json.Marshal(dbAa)
							return fmt.Sprintf("Coinbase rewardExist false, a[%d] b[%d], db writeset:%s", a1.Amount, b1.Amount, string(data))
						})
						return false
					}
				}
			}
		}
	}
	return true
}

func (validate *Validate) addIncome(income []modules.AmountAsset, newAmount uint64, asset *modules.Asset) []modules.AmountAsset {
	newValue := []modules.AmountAsset{}
	hasOldValue := false
	for _, aa := range income {
		if aa.Asset.Equal(asset) {
			aa.Amount += newAmount
			hasOldValue = true
		}
		newValue = append(newValue, aa)
	}
	if !hasOldValue {
		newValue = append(newValue, modules.AmountAsset{Amount: newAmount, Asset: asset})
	}
	return newValue
}
