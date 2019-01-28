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
	"fmt"

	"github.com/palletone/go-palletone/common"
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

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
func (validate *Validate) ValidateTransactions(txs *modules.Transactions, isGenesis bool) error {
	if txs == nil || txs.Len() < 1 {
		if !dagconfig.DefaultConfig.IsRewardCoin {
			return nil
		}
		return fmt.Errorf("Transactions should not be empty.")
	}

	fee := uint64(0) // todo zxl overflow
	txFlags := map[common.Hash]ValidationCode{}
	isSuccess := bool(true)
	// all transactions' new worldState
	//worldState := map[string]map[string]interface{}{}

	for txIndex, tx := range *txs {
		txHash := tx.Hash()
		// validate transaction id duplication
		if _, ok := txFlags[txHash]; ok == true {
			isSuccess = false
			log.Info("ValidateTx", "txhash", txHash, "error validate code", TxValidationCode_DUPLICATE_TXID)
			txFlags[txHash] = TxValidationCode_DUPLICATE_TXID
			continue
		}
		// validate common property
		//The first Tx(txIdx==0) is a coinbase tx.

		txCode := validate.validateTx(tx, txIndex == 0)
		if txCode != TxValidationCode_VALID {
			log.Debug("ValidateTx", "txhash", txHash, "error validate code", txCode)
			isSuccess = false
			txFlags[txHash] = txCode
			continue
		}
		// validate total fee
		if isGenesis == false && txIndex != 0 {
			txFee, err := tx.GetTxFee(validate.utxoquery.GetUtxoEntry)
			if err != nil {
				log.Info("ValidateTx", "txhash", txHash, "error validate code", TxValidationCode_INVALID_FEE)
				return err
			}
			fee += txFee.Amount
		}
		txFlags[txHash] = TxValidationCode_VALID
	}

	// check coinbase fee and income
	if !isGenesis && isSuccess {
		if len((*txs)[0].TxMessages) != 1 {
			return fmt.Errorf("Unit coinbase length is error.")
		}

		coinIn, ok := (*txs)[0].TxMessages[0].Payload.(*modules.PaymentPayload)
		if !ok {
			return fmt.Errorf("Coinbase payload type error.")
		}
		if len(coinIn.Outputs) != 1 {
			return fmt.Errorf("Coinbase outputs error0.")
		}
		income := uint64(fee) + ComputeRewards()
		if coinIn.Outputs[0].Value < income {
			return fmt.Errorf("Coinbase outputs error: 1.%d", income)
		}
	}
	return nil
}
func ComputeRewards() uint64 {
	var rewards uint64
	if dagconfig.DefaultConfig.IsRewardCoin {
		rewards = uint64(modules.DAO)
	}
	return rewards
}
func (validate *Validate) ValidateTx(tx *modules.Transaction, isCoinbase bool) error {
	code := validate.validateTx(tx, isCoinbase)
	if code == TxValidationCode_VALID {
		log.Debugf("Tx[%s] validate pass!",tx.Hash().String())
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
func (validate *Validate) checkTxIsExist(tx *modules.Transaction) bool {
	if len(tx.TxMessages) > 2 {
		reqId := tx.RequestHash()
		if txHash, err := validate.dagquery.GetTxHashByReqId(reqId); err == nil && txHash != (common.Hash{}) {
			log.Debug("checkTxIsExist", "transactions exist in dag, reqId:", reqId.String())
			return true
		}
	}
	return false
}
