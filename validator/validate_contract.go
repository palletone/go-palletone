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
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/modules"
	"math"
)

/**
对unit中某个交易的读写集进行验证
To validate read set and write set of one transaction in unit'
*/
func (validate *Validate) validateContractState(contractID []byte, readSet []modules.ContractReadSet, writeSet []modules.ContractWriteSet) ValidationCode {
	if !validate.dagquery.CheckReadSetValid(contractID, readSet) {
		return TxValidationCode_CHAINCODE_VERSION_CONFLICT
	}

	return TxValidationCode_VALID
}

/**
验证合约模板交易
To validate contract template payload
*/
func (validate *Validate) validateContractTplPayload(contractTplPayload *modules.ContractTplPayload) ValidationCode {
	// to check template whether existing or not
	stateDb := validate.statequery
	if stateDb != nil {
		tpl, _ := validate.statequery.GetContractTpl(contractTplPayload.TemplateId)
		if tpl != nil {
			log.Debug("validateContractTplPayload", "Contract template already exist!", contractTplPayload.TemplateId)
			return TxValidationCode_INVALID_CONTRACT_TEMPLATE
		}
	}
	return TxValidationCode_VALID
}

func (validate *Validate) validateContractDeploy(tplId []byte) ValidationCode {
	return TxValidationCode_VALID
}

//验证陪审团签名是否有效
func (validate *Validate) validateContractSignature(signatures []modules.SignatureSet,
	tx *modules.Transaction, isFullTx bool) ValidationCode {
	//contractId := tx.GetContractId()
	txHash := tx.Hash().String()
	needSign := 1
	//如果是Deploy，那么Jury在DeployPayload里面
	var jury *modules.ElectionNode
	var err error
	var contractId []byte
	for _, msg := range tx.TxMessages() {
		if msg.App == modules.APP_CONTRACT_DEPLOY {
			deploy := msg.Payload.(*modules.ContractDeployPayload)
			jury = &deploy.EleNode
		} else if msg.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			invokeReq := msg.Payload.(*modules.ContractInvokeRequestPayload)
			contractId = invokeReq.ContractId
		} else if msg.App == modules.APP_CONTRACT_STOP_REQUEST {
			stopReq := msg.Payload.(*modules.ContractStopRequestPayload)
			contractId = stopReq.ContractId
		}
	}
	// 1.对于用户合约，确认签名者都是Jury
	if common.IsUserContractId(contractId) { // user contract
		jury, err = validate.statequery.GetContractJury(contractId)
		if err != nil {
			log.Errorf("GetContractJury by contractId[%x] throw an error:%s",
				contractId, err.Error())
			return TxValidationCode_INVALID_CONTRACT_SIGN
		}
	}
	if jury != nil { //有陪审团信息,判断公钥和陪审员是否匹配
		jurorCount := len(jury.EleList)
		needSign = int(math.Ceil((float64(jurorCount)*2 + 1) / 3))
		for _, s := range signatures {
			jAddr := crypto.PubkeyBytesToAddress(s.PubKey)
			jAddrHash := util.RlpHash(jAddr)
			find := false
			for _, node := range jury.EleList {
				if jAddrHash == node.AddrHash {
					find = true
					break
				}
			}
			if !find { //签名者不是合法的陪审员
				log.Warnf("Tx[%s] signature payload pubKey[%x] is not a valid juror", txHash, s.PubKey)
				return TxValidationCode_INVALID_CONTRACT_SIGN
			}
		}
	}

	//2.确认签名都验证通过
	tx4Sign := tx.GetResultRawTx()
	txBytes, _ := rlp.EncodeToBytes(tx4Sign)
	passCount := 0
	for _, s := range signatures {
		pass, err := crypto.MyCryptoLib.Verify(s.PubKey, s.Signature, txBytes)
		if err != nil {
			log.Error(err.Error())
			return TxValidationCode_INVALID_CONTRACT_SIGN
		}
		if pass {
			passCount++
		}
	}
	//3.确认签名数量满足系统要求
	if isFullTx && passCount < needSign {
		log.Errorf("Tx[%s] need signature count:%d, but current has %d", txHash, needSign, passCount)
		return TxValidationCode_INVALID_CONTRACT_SIGN
	}
	return TxValidationCode_VALID
}
