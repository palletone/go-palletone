/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package common

import (
	"bytes"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Validate struct {
	dagdb   storage.DagDb
	utxodb  storage.UtxoDb
	statedb storage.StateDb
}

func NewValidate(dagdb storage.DagDb, utxodb storage.UtxoDb, statedb storage.StateDb) *Validate {
	return &Validate{dagdb: dagdb, utxodb: utxodb, statedb: statedb}
}

type Validator interface {
	ValidateTransactions(txs *modules.Transactions, isGenesis bool) (map[common.Hash]modules.TxValidationCode, bool, error)
	ValidateUnit(unit *modules.Unit, isGenesis bool) byte
	ValidateTx(tx *modules.Transaction, isCoinbase bool, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode
	ValidateUnitSignature(h *modules.Header, isGenesis bool) byte
}

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
func (validate *Validate) ValidateTransactions(txs *modules.Transactions, isGenesis bool) (map[common.Hash]modules.TxValidationCode, bool, error) {
	if txs == nil || txs.Len() < 1 {
		return nil, false, fmt.Errorf("Transactions should not be empty.")
	}

	fee := uint64(0)
	txFlags := map[common.Hash]modules.TxValidationCode{}
	isSuccess := bool(true)
	// all transactions' new worldState
	worldState := map[string]map[string]interface{}{}

	for txIndex, tx := range *txs {
		// validate transaction id duplication
		if _, ok := txFlags[tx.TxHash]; ok == true {
			isSuccess = false
			log.Info("ValidateTx", "txhash", tx.TxHash, "error validate code", modules.TxValidationCode_DUPLICATE_TXID)
			txFlags[tx.TxHash] = modules.TxValidationCode_DUPLICATE_TXID
			continue
		}
		// validate common property
		//The first Tx(txIdx==0) is a coinbase tx.
		txCode := validate.ValidateTx(tx, txIndex == 0, &worldState)
		if txCode != modules.TxValidationCode_VALID {
			log.Info("ValidateTx", "txhash", tx.TxHash, "error validate code", txCode)
			isSuccess = false
			txFlags[tx.TxHash] = txCode
			continue
		}
		// validate fee
		if isGenesis == false && txIndex != 0 {
			txFee := tx.Fee()
			if txFee.Cmp(modules.TXFEE) < 0 {
				isSuccess = false
				txFlags[tx.TxHash] = modules.TxValidationCode_INVALID_FEE
				log.Info("ValidateTx", "txhash", tx.TxHash, "error validate code", modules.TxValidationCode_INVALID_FEE)
				continue
			}
			fee += txFee.Uint64()
		}
		txFlags[tx.TxHash] = modules.TxValidationCode_VALID
	}

	// check coinbase fee and income
	if !isGenesis && isSuccess {
		if len((*txs)[0].TxMessages) != 1 {
			return nil, false, fmt.Errorf("Unit coinbase length is error.")
		}

		coinIn, ok := (*txs)[0].TxMessages[0].Payload.(*modules.PaymentPayload)
		if !ok {
			return nil, false, fmt.Errorf("Coinbase payload type error.")
		}
		if len(coinIn.Output) != 1 {
			return nil, false, fmt.Errorf("Coinbase outputs error0.")
		}
		income := uint64(fee) + ComputeInterest()
		if coinIn.Output[0].Value < income {
			return nil, false, fmt.Errorf("Coinbase outputs error: 1.%d", income)
		}
	}
	return txFlags, isSuccess, nil
}

/**
验证某个交易
To validate one transaction
*/
func (validate *Validate) ValidateTx(tx *modules.Transaction, isCoinbase bool, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode {
	if len(tx.TxMessages) == 0 {
		return modules.TxValidationCode_INVALID_MSG
	}
	if tx.TxMessages[0].App != modules.APP_PAYMENT { // 交易费
		fmt.Printf("-----------ValidateTx , %d\n", tx.TxMessages[0].App)
		return modules.TxValidationCode_INVALID_MSG
	}
	// validate transaction hash
	if !bytes.Equal(tx.TxHash.Bytes(), tx.Hash().Bytes()) {
		return modules.TxValidationCode_NIL_TXACTION
	}
	for _, msg := range tx.TxMessages {
		// check message type and payload
		if !validateMessageType(msg.App, msg.Payload) {
			return modules.TxValidationCode_UNKNOWN_TX_TYPE
		}
		// validate tx size
		if tx.Size().Float64() > float64(modules.TX_MAXSIZE) {
			log.Debug("Tx size is to big.")
			return modules.TxValidationCode_NOT_COMPARE_SIZE
		}

		// validate transaction signature
		if validateTxSignature(tx) == false {
			return modules.TxValidationCode_BAD_CREATOR_SIGNATURE
		}
		// validate every type payload
		switch msg.App {
		case modules.APP_PAYMENT:
			payment, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				return modules.TxValidationCode_INVALID_PAYMMENTLOAd
			}
			validateCode := validate.validatePaymentPayload(payment, isCoinbase)
			if validateCode != modules.TxValidationCode_VALID {
				return validateCode
			}
		case modules.APP_CONTRACT_TPL:
			payload, _ := msg.Payload.(*modules.ContractTplPayload)
			validateCode := validate.validateContractTplPayload(payload)
			if validateCode != modules.TxValidationCode_VALID {
				return validateCode
			}
		case modules.APP_CONTRACT_DEPLOY:
			payload, _ := msg.Payload.(*modules.ContractDeployPayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet, worldTmpState)
			if validateCode != modules.TxValidationCode_VALID {
				return validateCode
			}
		case modules.APP_CONTRACT_INVOKE:
			payload, _ := msg.Payload.(*modules.ContractInvokePayload)
			validateCode := validate.validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet, worldTmpState)
			if validateCode != modules.TxValidationCode_VALID {
				return validateCode
			}
		case modules.APP_CONFIG:
		case modules.APP_TEXT:

		default:
			return modules.TxValidationCode_UNKNOWN_TX_TYPE
		}
	}
	return modules.TxValidationCode_VALID
}

/**
检查message的app与payload是否一致
check messaage 'app' consistent with payload type
*/
func validateMessageType(app byte, payload interface{}) bool {
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
	case *modules.ContractInvokePayload:
		if app == modules.APP_CONTRACT_INVOKE {
			return true
		}
	case *modules.ConfigPayload:
		if app == modules.APP_CONFIG {
			return true
		}
	case *modules.TextPayload:
		if app == modules.APP_TEXT {
			return true
		}
	default:
		log.Info("The payload of message type is not expect. ", "payload_type", t)
		return false
	}
	return false
}

/**
验证单元的签名，需要比对见证人列表
To validate unit's signature, and mediators' signature
*/
func (validate *Validate) ValidateUnitSignature(h *modules.Header, isGenesis bool) byte {
	emptySigUnit := modules.Unit{}
	// copy unit's header
	emptySigUnit.UnitHeader = modules.CopyHeader(h)
	// signature does not contain authors and witness fields
	emptySigUnit.UnitHeader.Authors = nil
	emptySigUnit.UnitHeader.Witness = []*modules.Authentifier{}
	// recover signature
	if h.Authors == nil {
		log.Debug("Verify unit signature ,header's authors is nil.")
		return modules.UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	sig := make([]byte, 65)
	copy(sig[32-len(h.Authors.R):32], h.Authors.R)
	copy(sig[64-len(h.Authors.S):64], h.Authors.S)
	copy(sig[64:], h.Authors.V)
	// recover pubkey
	hash := crypto.Keccak256Hash(util.RHashBytes(*emptySigUnit.UnitHeader))
	pubKey, err := modules.RSVtoPublicKey(hash[:], h.Authors.R[:], h.Authors.S[:], h.Authors.V[:])
	if err != nil {
		log.Debug("Verify unit signature when recover pubkey", "error", err.Error())
		return modules.UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	//  pubKey to pubKey_bytes
	pubKey_bytes := crypto.FromECDSAPub(pubKey)
	if keystore.VerifyUnitWithPK(sig, *emptySigUnit.UnitHeader, pubKey_bytes) == false {
		log.Debug("Verify unit signature error.")
		return modules.UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	// if genesis unit just return
	if isGenesis == true {
		return modules.UNIT_STATE_VALIDATED
	}
	// get mediators
	data := validate.statedb.GetConfig([]byte("MediatorCandidates"))
	var mList []core.MediatorInfo
	if err := rlp.DecodeBytes(data, &mList); err != nil {
		log.Debug("Check unit signature when get mediators list", "error", err.Error())
		return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	}
	bNum := validate.statedb.GetConfig([]byte("ActiveMediators"))
	var mNum uint16
	if err := rlp.DecodeBytes(bNum, &mNum); err != nil {
		log.Debug("Check unit signature", "error", err.Error())
		return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	}
	if int(mNum) != len(mList) {
		log.Debug("Check unit signature", "error", "mediators info error, pls update network")
		return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	}
	// 这一步后续添加： 调用 mediator 模块校验见证人的接口

	//return modules.UNIT_STATE_VALIDATED
	return modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED
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

/**
对unit中某个交易的读写集进行验证
To validate read set and write set of one transaction in unit'
*/
func (validate *Validate) validateContractState(contractID []byte, readSet *[]modules.ContractReadSet, writeSet *[]modules.PayloadMapStruct, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode {
	// check read set, if read field in worldTmpState then the transaction is invalid
	contractState, cOk := (*worldTmpState)[hexutil.Encode(contractID[:])]
	if cOk && readSet != nil {
		for _, rs := range *readSet {
			if _, ok := contractState[rs.Key]; ok == true {
				return modules.TxValidationCode_CHAINCODE_VERSION_CONFLICT
			}
		}
	}
	// save write set to worldTmpState
	if !cOk && writeSet != nil {
		(*worldTmpState)[hexutil.Encode(contractID[:])] = map[string]interface{}{}
	}

	for _, ws := range *writeSet {
		(*worldTmpState)[hexutil.Encode(contractID[:])][ws.Key] = ws.Value
	}
	return modules.TxValidationCode_VALID
}

/**
验证合约模板交易
To validate contract template payload
*/
func (validate *Validate) validateContractTplPayload(contractTplPayload *modules.ContractTplPayload) modules.TxValidationCode {
	// to check template whether existing or not
	stateVersion, bytecode, name, path := validate.statedb.GetContractTpl(contractTplPayload.TemplateId)
	if stateVersion == nil && bytecode == nil && name == "" && path == "" {
		return modules.TxValidationCode_VALID
	}
	return modules.TxValidationCode_INVALID_CONTRACT_TEMPLATE
}

//验证一个Payment
//Validate a payment message
//1. Amount correct
//2. Asset must be equal
//3. Unlock correct
func (validate *Validate) validatePaymentPayload(payment *modules.PaymentPayload, isCoinbase bool) modules.TxValidationCode {
	// check locktime

	if len(payment.Input) <= 0 {
		return modules.TxValidationCode_INVALID_PAYMMENT_INPUT
	}
	if !isCoinbase {
		for _, in := range payment.Input {
			// checkout input
			if in.PreviousOutPoint == nil {
				return modules.TxValidationCode_INVALID_PAYMMENT_INPUT
			}
			if utxo, err := validate.utxodb.GetUtxoEntry(in.PreviousOutPoint.ToKey()); utxo == nil || err != nil {
				return modules.TxValidationCode_INVALID_OUTPOINT
			}
			// check SignatureScript
		}
	}

	if len(payment.Output) <= 0 {
		return modules.TxValidationCode_INVALID_PAYMMENT_OUTPUT
	}

	for i, out := range payment.Output {
		// checkout output
		if i < 1 {
			continue // asset = out.Asset
		} else {
			if out.Asset == nil {
				return modules.TxValidationCode_INVALID_ASSET
			}
			if !out.Asset.IsSimilar(payment.Output[i-1].Asset) {
				return modules.TxValidationCode_INVALID_ASSET
			}
		}
		if out.Value <= 0 || out.Value >= 100000000000000000 {
			return modules.TxValidationCode_INVALID_AMOUNT
		}
	}
	return modules.TxValidationCode_VALID
}

/**
验证Unit
Validate unit
*/
func (validate *Validate) ValidateUnit(unit *modules.Unit, isGenesis bool) byte {
	//  unit's size  should bigger than minimum.
	if unit.Size() < 125 {
		log.Debug("Validate size", "error", "size is invalid", "size", unit.Size())
		return modules.UNIT_STATE_INVALID_SIZE
	}

	// step1. check header.
	sigState := validate.validateHeader(unit.UnitHeader, isGenesis)
	if sigState != modules.UNIT_STATE_VALIDATED {
		log.Debug("Validate unit's header failed.", "error code", sigState)
		return sigState
	}
	// step2. check transactions in unit
	_, isSuccess, err := validate.ValidateTransactions(&unit.Txs, isGenesis)
	if isSuccess != true {
		msg := fmt.Sprintf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
		log.Debug(msg)
		return modules.UNIT_STATE_HAS_INVALID_TRANSACTIONS
	}
	return sigState
}

func (validate *Validate) validateHeader(header *modules.Header, isGenesis bool) byte {
	if header == nil {
		return modules.UNIT_STATE_INVALID_HEADER
	}

	if len(header.ParentsHash) == 0 {
		if !isGenesis {
			return modules.UNIT_STATE_INVALID_HEADER
		}
	}
	//  check header's extra data
	if uint64(len(header.Extra)) > configure.MaximumExtraDataSize {
		msg := fmt.Sprintf("extra-data too long: %d > %d", len(header.Extra), configure.MaximumExtraDataSize)
		log.Debug(msg)
		return modules.UNIT_STATE_INVALID_EXTRA_DATA
	}
	// check txroot
	if header.TxRoot == (common.Hash{}) {
		return modules.UNIT_STATE_INVALID_HEADER
	}

	// check creation_time
	if header.Creationdate <= modules.UNIT_CREATION_DATE_INITIAL_UINT64 {
		return modules.UNIT_STATE_INVALID_HEADER
	}

	// check header's number
	if header.Number == (modules.ChainIndex{}) {
		return modules.UNIT_STATE_INVALID_HEADER
	}
	if len(header.AssetIDs) == 0 {
		return modules.UNIT_STATE_INVALID_HEADER
	}

	if isGenesis {
		if len(header.AssetIDs) != 1 {
			return modules.UNIT_STATE_INVALID_HEADER
		}
		if header.AssetIDs[0] != modules.PTNCOIN || !header.Number.IsMain || header.Number.Index != 0 {
			return modules.UNIT_STATE_INVALID_HEADER
		}
		return modules.UNIT_STATE_CHECK_HEADER_PASSED
	}
	var isValidAssetId bool
	for _, asset := range header.AssetIDs {
		if asset == header.Number.AssetID {
			isValidAssetId = true
			break
		}
	}
	if !isValidAssetId {
		return modules.UNIT_STATE_INVALID_HEADER
	}

	// check authors
	if header.Authors == nil {
		return modules.UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	if len(header.Witness) > 21 || len(header.Witness) < 16 {
		return modules.UNIT_STATE_INVALID_HEADER_WITNESS
	}
	sigState := validate.ValidateUnitSignature(header, isGenesis)
	return sigState
}
