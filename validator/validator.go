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
	"sync"
	"time"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

type Validate struct {
	utxoquery  IUtxoQuery
	statequery IStateQuery
	dagquery   IDagQuery
	propquery  IPropQuery
}

const MAX_DATA_PAYLOAD_MAIN_DATA_SIZE = 128

func NewValidate(dagdb IDagQuery, utxoRep IUtxoQuery, statedb IStateQuery, propquery IPropQuery) *Validate {
	return &Validate{dagquery: dagdb, utxoquery: utxoRep, statequery: statedb, propquery: propquery}
}

type newUtxoQuery struct {
	oldUtxoQuery IUtxoQuery
	unitUtxo     *sync.Map
}

func (q *newUtxoQuery) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	utxo, ok := q.unitUtxo.Load(*outpoint)
	if ok {
		return utxo.(*modules.Utxo), nil
	}
	return q.oldUtxoQuery.GetUtxoEntry(outpoint)
}
func (validate *Validate) setUtxoQuery(q IUtxoQuery) {
	validate.utxoquery = q
}

//逐条验证每一个Tx，并返回总手续费
func (validate *Validate) validateTransactions(txs modules.Transactions, unitTime int64) ValidationCode {
	fee := uint64(0)

	oldUtxoQuery := validate.utxoquery

	unitUtxo := new(sync.Map)
	newUtxoQuery := &newUtxoQuery{oldUtxoQuery: oldUtxoQuery, unitUtxo: unitUtxo}
	validate.utxoquery = newUtxoQuery
	defer validate.setUtxoQuery(oldUtxoQuery)

	var coinbase *modules.Transaction
	ads:=[]*modules.Addition{}
	for txIndex, tx := range txs {
		//先检查普通交易并计算手续费，最后检查Coinbase
		txHash := tx.Hash()
		if validate.checkTxIsExist(tx) {
			return TxValidationCode_DUPLICATE_TXID
		}
		if txIndex == 0 {
			coinbase=tx
			continue
			//每个单元的第一条交易比较特殊，是Coinbase交易，其包含增发和收集的手续费

		}
		//TODO Devin
		txCode, txFee := validate.validateTx(tx,  true, unitTime)
		if txCode != TxValidationCode_VALID {
			log.Debug("ValidateTx", "txhash", txHash, "error validate code", txCode)

			return txCode
		}
		//txFee, _ := tx.GetTxFee(validate.utxoquery.GetUtxoEntry, unitTime)
		fee += txFee.Amount

		for outPoint, utxo := range tx.GetNewUtxos() {
			unitUtxo.Store(outPoint, utxo)
		}
		//newUtxoQuery.unitUtxo = unitUtxo
		//validate.utxoquery = newUtxoQuery
	}
	//验证第一条交易
	if len(txs)>0 {
		//手续费应该与其他交易付出的手续费相等
		return	validate.validateCoinbase(coinbase,ads)

		//allIncome := uint64(0)
		//outputs := coinbase.TxMessages[0].Payload.(*modules.PaymentPayload).Outputs
		//for _, output := range outputs {
		//	allIncome += output.Value
		//}
		//if allIncome != fee+reward {
		//	log.Warnf("Unit has an incorrect coinbase, expect income=%d,actual=%d", fee+reward, allIncome)
		//	return TxValidationCode_INVALID_FEE
		//}
	}
	return TxValidationCode_VALID
}

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
//func (validate *Validate) ValidateTransactions(txs modules.Transactions) error {
//	code := validate.validateTransactions(txs)
//	return NewValidateError(code)
//}
func (validate *Validate) validateCoinbase(tx *modules.Transaction,ads []*modules.Addition) ValidationCode{
	return TxValidationCode_VALID
}
func (validate *Validate) ValidateTx(tx *modules.Transaction,  isFullTx bool) error {
	code, _ := validate.validateTx(tx, isFullTx, time.Now().Unix())
	if code == TxValidationCode_VALID {
		return nil
	}
	log.Debugf("Tx[%s] validate not pass!", tx.Hash().String())
	return NewValidateError(code)
}

// todo
// 验证群签名接口，需要验证群签的正确性和有效性
func (validate *Validate) ValidateUnitGroupSign(h *modules.Header) error {
	return nil
}

/**
验证交易签名
To validate transaction signature
*/
func validateTxSignature(tx *modules.Transaction) bool {
	// recover signature
	//cpySig := make([]byte, 65)
	//copy(cpySig[32-len(sig.R):32], sig.R)
	//copy(cpySig[64-len(sig.S):64], sig.S)
	//copy(cpySig[64:], sig.V)
	//// recover pubkey
	//hash := crypto.Keccak256Hash(util.RHashBytes(txHash))
	//pubKey, err := modules.RSVtoPublicKey(hash[:], sig.R[:], sig.S[:], sig.V[:])
	//if err != nil {
	//	log.Error("Validate transaction signature", "error", err.Error())
	//	return false
	//}
	////  pubKey to pubKey_bytes
	//pubKey_bytes := crypto.FromECDSAPub(pubKey)
	//if keystore.VerifyUnitWithPK(cpySig, txHash, pubKey_bytes) == true {
	//	return true
	//}
	return true
}

//验证一个DataPayment
func (validate *Validate) validateDataPayload(payload *modules.DataPayload) ValidationCode {
	//验证 maindata是否存在
	//验证 maindata extradata大小 不可过大
	if len(payload.MainData) >= MAX_DATA_PAYLOAD_MAIN_DATA_SIZE || len(payload.MainData) == 0 {
		return TxValidationCode_INVALID_DATAPAYLOAD
	}
	//TODO 验证maindata其它属性
	return TxValidationCode_VALID
}
func (validate *Validate) CheckTxIsExist(tx *modules.Transaction) bool {
	return validate.checkTxIsExist(tx)
}
func (validate *Validate) checkTxIsExist(tx *modules.Transaction) bool {
	if len(tx.TxMessages) > 2 {
		txHash := tx.Hash()
		if validate.dagquery == nil {
			log.Warnf("Validate DagQuery doesn't set, cannot check tx[%s] is exist or not", txHash.String())
			return false
		}
		if has, _ := validate.dagquery.IsTransactionExist(txHash); has {
			log.Debug("checkTxIsExist transactions exist in dag", "txHash", txHash.String())
			return true
		}
	}
	return false
}
