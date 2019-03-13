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

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

type Validate struct {
	utxoquery  IUtxoQuery
	statequery IStateQuery
	dagquery   IDagQuery
}

const MAX_DATA_PAYLOAD_MAIN_DATA_SIZE = 128

func NewValidate(dagdb IDagQuery, utxoRep IUtxoQuery, statedb IStateQuery) *Validate {
	return &Validate{dagquery: dagdb, utxoquery: utxoRep, statequery: statedb}
}

type newUtxoQuery struct {
	oldUtxoQuery IUtxoQuery
	unitUtxo     sync.Map
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
func (validate *Validate) validateTransactions(txs modules.Transactions) ValidationCode {
	fee := uint64(0)
	needCheckCoinbase := false
	oldUtxoQuery := validate.utxoquery

	var unitUtxo sync.Map
	newUtxoQuery := &newUtxoQuery{oldUtxoQuery: oldUtxoQuery, unitUtxo: unitUtxo}
	validate.utxoquery = newUtxoQuery
	defer validate.setUtxoQuery(oldUtxoQuery)

	var coinbase *modules.Transaction
	for txIndex, tx := range txs {
		txHash := tx.Hash()
		if txIndex == 0 && tx.TxMessages[0].Payload.(*modules.PaymentPayload).IsCoinbase() {
			needCheckCoinbase = true
			coinbase = tx
			continue
			//每个单元的第一条交易比较特殊，是Coinbase交易，其包含增发和收集的手续费

		}
		txCode := validate.validateTx(tx, txIndex == 0)
		if txCode != TxValidationCode_VALID {
			log.Debug("ValidateTx", "txhash", txHash, "error validate code", txCode)

			return txCode
		}
		//getUtxoFromUnitAndDb := func(outpoint *modules.OutPoint) (*modules.Utxo, error) {
		//	if utxo, ok := unitUtxo[*outpoint]; ok {
		//		return utxo, nil
		//	}
		//	return validate.utxoquery.GetUtxoEntry(outpoint)
		//}
		//txFee, _ := tx.GetTxFee(getUtxoFromUnitAndDb)
		txFee, _ := tx.GetTxFee(validate.utxoquery.GetUtxoEntry)
		fee += txFee.Amount

		for outPoint, utxo := range tx.GetNewUtxos() {
			unitUtxo.Store(outPoint, utxo)
		}
		newUtxoQuery.unitUtxo = unitUtxo
		validate.utxoquery = newUtxoQuery
	}
	//验证第一条交易
	if needCheckCoinbase {
		//手续费应该与其他交易付出的手续费相等
		reward := ComputeRewards()
		//TODO PTN增发的情况

		mediatorIncome := coinbase.TxMessages[0].Payload.(*modules.PaymentPayload).Outputs[0].Value
		if mediatorIncome != fee+reward {
			log.Warnf("Unit has an incorrect coinbase, expect income=%d,actual=%d", fee+reward, mediatorIncome)
			return TxValidationCode_INVALID_FEE
		}
	}
	return TxValidationCode_VALID
}

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
func (validate *Validate) ValidateTransactions(txs modules.Transactions) error {
	code := validate.validateTransactions(txs)
	return NewValidateError(code)
	//if txs == nil || txs.Len() < 1 {
	//	if !dagconfig.DefaultConfig.IsRewardCoin {
	//		return nil
	//	}
	//	return fmt.Errorf("Transactions should not be empty.")
	//}
	//
	//// todo zxl overflow
	//txFlags := map[common.Hash]ValidationCode{}
	//isSuccess := bool(true)
	//// all transactions' new worldState
	////worldState := map[string]map[string]interface{}{}
	//
	//for txIndex, tx := range *txs {
	//	txHash := tx.Hash()
	//	// validate transaction id duplication
	//	if _, ok := txFlags[txHash]; ok == true {
	//		isSuccess = false
	//		log.Info("ValidateTx", "txhash", txHash, "error validate code", TxValidationCode_DUPLICATE_TXID)
	//		txFlags[txHash] = TxValidationCode_DUPLICATE_TXID
	//		continue
	//	}
	//	// validate common property
	//	//The first Tx(txIdx==0) is a coinbase tx.
	//
	//	txCode := validate.validateTx(tx, txIndex == 0)
	//	if txCode != TxValidationCode_VALID {
	//		log.Debug("ValidateTx", "txhash", txHash, "error validate code", txCode)
	//		isSuccess = false
	//		txFlags[txHash] = txCode
	//		continue
	//	}
	//	// validate total fee
	//	if isGenesis == false && txIndex != 0 {
	//		txFee, err := tx.GetTxFee(validate.utxoquery.GetUtxoEntry)
	//		if err != nil {
	//			log.Info("ValidateTx", "txhash", txHash, "error validate code", TxValidationCode_INVALID_FEE)
	//			return err
	//		}
	//		fee += txFee.Amount
	//	}
	//	txFlags[txHash] = TxValidationCode_VALID
	//}
	//
	//// check coinbase fee and income
	//if !isGenesis && isSuccess {
	//	if len((*txs)[0].TxMessages) != 1 {
	//		return fmt.Errorf("Unit coinbase length is error.")
	//	}
	//
	//	coinIn, ok := (*txs)[0].TxMessages[0].Payload.(*modules.PaymentPayload)
	//	if !ok {
	//		return fmt.Errorf("Coinbase payload type error.")
	//	}
	//	if len(coinIn.Outputs) != 1 {
	//		return fmt.Errorf("Coinbase outputs error0.")
	//	}
	//	income := uint64(fee) + ComputeRewards()
	//	if coinIn.Outputs[0].Value < income {
	//		return fmt.Errorf("Coinbase outputs error: 1.%d", income)
	//	}
	//}
	//return nil
}
func ComputeRewards() uint64 {
	var rewards uint64
	if dagconfig.DagConfig.IsRewardCoin {
		rewards = uint64(modules.DAO)
	}
	return rewards
}
func (validate *Validate) ValidateTx(tx *modules.Transaction, isCoinbase bool) error {
	code := validate.validateTx(tx, isCoinbase)
	if code == TxValidationCode_VALID {
		log.Debugf("Tx[%s] validate pass!", tx.Hash().String())
		return nil
	}
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
		if tx, err := validate.dagquery.GetTransactionOnly(txHash); err == nil && tx != nil {
			log.Debug("checkTxIsExist transactions exist in dag", "txHash", txHash.String())
			return true
		}
	}
	return false
}
