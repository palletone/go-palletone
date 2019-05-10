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

/**
对unit中某个交易的读写集进行验证
To validate read set and write set of one transaction in unit'
*/
func (validate *Validate) validateContractState(contractID []byte, readSet *[]modules.ContractReadSet, writeSet *[]modules.ContractWriteSet) ValidationCode {
	// check read set, if read field in worldTmpState then the transaction is invalid
	//contractState, cOk := (*worldTmpState)[hexutil.Encode(contractID[:])]
	//if cOk && readSet != nil {
	//	for _, rs := range *readSet {
	//		if _, ok := contractState[rs.Key]; ok == true {
	//			return TxValidationCode_CHAINCODE_VERSION_CONFLICT
	//		}
	//	}
	//}
	//// save write set to worldTmpState
	//if !cOk && writeSet != nil {
	//	(*worldTmpState)[hexutil.Encode(contractID[:])] = map[string]interface{}{}
	//}
	//
	//for _, ws := range *writeSet {
	//	(*worldTmpState)[hexutil.Encode(contractID[:])][ws.Key] = ws.Value
	//}
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

func (validate *Validate) validateContractdeploy(tplId []byte) ValidationCode {
	return TxValidationCode_VALID
}

func (validate *Validate) validateContractSignature(sinatures []modules.SignatureSet, tx *modules.Transaction) ValidationCode {
	return TxValidationCode_VALID
}
