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
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/vote"
)

type Validate struct {
	dagdb   storage.IDagDb
	utxodb  storage.IUtxoDb
	utxoRep IUtxoRepository
	statedb storage.IStateDb
	logger  log.ILogger
}

func NewValidate(dagdb storage.IDagDb, utxodb storage.IUtxoDb, utxoRep IUtxoRepository, statedb storage.IStateDb, l log.ILogger) *Validate {
	return &Validate{dagdb: dagdb, utxodb: utxodb, utxoRep: utxoRep, statedb: statedb, logger: l}
}

type Validator interface {
	ValidateTransactions(txs *modules.Transactions, isGenesis bool) (map[common.Hash]modules.TxValidationCode, bool, error)
	ValidateUnitExceptGroupSig(unit *modules.Unit, isGenesis bool) byte
	ValidateTx(tx *modules.Transaction, isCoinbase bool, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode
	ValidateUnitSignature(h *modules.Header, isGenesis bool) byte
	ValidateUnitGroupSign(h *modules.Header, isGenesis bool) byte
}

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
func (validate *Validate) ValidateTransactions(txs *modules.Transactions, isGenesis bool) (
	map[common.Hash]modules.TxValidationCode, bool, error) {
	if txs == nil || txs.Len() < 1 {
		if !dagconfig.DefaultConfig.IsRewardCoin {
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("Transactions should not be empty.")
	}

	fee := uint64(0) // todo zxl overflow
	txFlags := map[common.Hash]modules.TxValidationCode{}
	isSuccess := bool(true)
	// all transactions' new worldState
	worldState := map[string]map[string]interface{}{}

	for txIndex, tx := range *txs {
		txHash := tx.Hash()
		// validate transaction id duplication
		if _, ok := txFlags[txHash]; ok == true {
			isSuccess = false
			log.Info("ValidateTx", "txhash", txHash, "error validate code", modules.TxValidationCode_DUPLICATE_TXID)
			txFlags[txHash] = modules.TxValidationCode_DUPLICATE_TXID
			continue
		}
		// validate common property
		//The first Tx(txIdx==0) is a coinbase tx.

		txCode := validate.ValidateTx(tx, txIndex == 0, &worldState)
		if txCode != modules.TxValidationCode_VALID {
			log.Debug("ValidateTx", "txhash", txHash, "error validate code", txCode)
			isSuccess = false
			txFlags[txHash] = txCode
			continue
		}
		// validate fee
		if isGenesis == false && txIndex != 0 {
			txFee, err := validate.utxoRep.ComputeTxFee(tx)
			if err != nil {
				log.Info("ValidateTx", "txhash", txHash, "error validate code", modules.TxValidationCode_INVALID_FEE)
				return nil, false, err
			}
			fee += txFee.Amount
		}
		txFlags[txHash] = modules.TxValidationCode_VALID
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
		if len(coinIn.Outputs) != 1 {
			return nil, false, fmt.Errorf("Coinbase outputs error0.")
		}
		income := uint64(fee) + ComputeRewards()
		if coinIn.Outputs[0].Value < income {
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

	if validate.checkTxIsExist(tx) {
		return modules.TxValidationCode_DUPLICATE_TXID
	}
	// validate transaction hash
	//if !bytes.Equal(tx.TxHash.Bytes(), tx.Hash().Bytes()) {
	//	return modules.TxValidationCode_NIL_TXACTION
	//}

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
				return modules.TxValidationCode_INVALID_PAYMMENTLOAD
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
		case modules.APP_CONTRACT_TPL_REQUEST:
			payload, _ := msg.Payload.(*modules.ContractInstallRequestPayload)
			if payload.TplName == "" || payload.Path == "" || payload.Version == "" {
				return modules.TxValidationCode_INVALID_CONTRACT
			}
			return modules.TxValidationCode_VALID

		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			// 参数临界值验证
			payload, _ := msg.Payload.(*modules.ContractDeployRequestPayload)
			if len(payload.TplId) == 0 || payload.TxId == "" || payload.Timeout < 0 {
				return modules.TxValidationCode_INVALID_CONTRACT
			}

			validateCode := validate.validateContractdeploy(payload.TplId, worldTmpState)
			return validateCode

		case modules.APP_CONTRACT_INVOKE_REQUEST:

			payload, _ := msg.Payload.(*modules.ContractInvokeRequestPayload)
			if len(payload.ContractId) == 0 {
				return modules.TxValidationCode_INVALID_CONTRACT
			}
			// 验证ContractId有效性
			if len(payload.ContractId) <= 0 {
				return modules.TxValidationCode_INVALID_CONTRACT
			}
			return modules.TxValidationCode_VALID

		case modules.APP_SIGNATURE:
			// 签名验证
			payload, _ := msg.Payload.(*modules.SignaturePayload)
			validateCode := validate.validateContractSignature(payload.Signatures[:], tx, worldTmpState)
			if validateCode != modules.TxValidationCode_VALID {
				return validateCode
			}

		case modules.APP_CONFIG:
		case modules.APP_TEXT:
		case modules.APP_VOTE:
		case modules.OP_MEDIATOR_CREATE:
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

	case *modules.ConfigPayload:
		if app == modules.APP_CONFIG {
			return true
		}
	case *modules.TextPayload:
		if app == modules.APP_TEXT {
			return true
		}
	case *vote.VoteInfo:
		if app == modules.APP_VOTE {
			return true
		}
	case *modules.MediatorCreateOperation:
		if app == modules.OP_MEDIATOR_CREATE {
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

	default:
		log.Debug("The payload of message type is not expect. ", "payload_type", t)
		return false
	}
	return false
}

// todo
// 验证群签名接口，需要验证群签的正确性和有效性
func (validate *Validate) ValidateUnitGroupSign(h *modules.Header, isGenesis bool) byte {

	return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
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
	//emptySigUnit.UnitHeader.Authors = nil
	emptySigUnit.UnitHeader.GroupSign = make([]byte, 0)
	// recover signature
	//if h.Authors == nil {
	if h.Authors.Empty() {
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
	//TODO Devin
	//data, _ := validate.statedb.GetCandidateMediatorAddrList() //.GetConfig([]byte("MediatorCandidates"))
	//var mList []core.MediatorInfo
	//if err := rlp.DecodeBytes(data, &mList); err != nil {
	//	log.Debug("Check unit signature when get mediators list", "error", err.Error())
	//	return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	//}
	//bNum, _ := validate.statedb.GetActiveMediatorAddrList()
	//var mNum uint16
	//if err := rlp.DecodeBytes(bNum, &mNum); err != nil {
	//	log.Debug("Check unit signature", "error", err.Error())
	//	return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	//}
	//if int(mNum) != len(mList) {
	//	log.Debug("Check unit signature", "error", "mediators info error, pls update network")
	//	return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	//}
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
func (validate *Validate) validateContractState(contractID []byte, readSet *[]modules.ContractReadSet, writeSet *[]modules.ContractWriteSet,
	worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode {
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

	// TODO coinbase 交易的inputs是null.
	// if len(payment.Inputs) <= 0 {
	// 	log.Error("payment input is null.", "payment.input", payment.Inputs)
	// 	return modules.TxValidationCode_INVALID_PAYMMENT_INPUT
	// }

	if !isCoinbase {
		for _, in := range payment.Inputs {
			// checkout input
			if in == nil || in.PreviousOutPoint == nil {
				log.Error("payment input is null.", "payment.input", payment.Inputs)
				return modules.TxValidationCode_INVALID_PAYMMENT_INPUT
			}
			// 合约创币后同步到mediator的utxo验证不通过,在创币后需要先将创币的utxo同步到所有mediator节点。
			if utxo, err := validate.utxodb.GetUtxoEntry(in.PreviousOutPoint); utxo == nil || err != nil {
				return modules.TxValidationCode_INVALID_OUTPOINT
			}
			// check SignatureScript
		}
	}

	if len(payment.Outputs) <= 0 {
		log.Error("payment output is null.", "payment.output", payment.Outputs)
		return modules.TxValidationCode_INVALID_PAYMMENT_OUTPUT
	}
	//Check coinbase payment
	//rule:
	//	1. all outputs have same asset
	asset0 := payment.Outputs[0].Asset
	for _, out := range payment.Outputs {
		if !asset0.IsSimilar(out.Asset) {
			return modules.TxValidationCode_INVALID_ASSET
		}
	}

	for _, out := range payment.Outputs {
		// // checkout output
		// if i < 1 {
		// 	if !out.Asset.IsSimilar(modules.NewPTNAsset()) {
		// 		return modules.TxValidationCode_INVALID_ASSET
		// 	}
		// 	// log.Debug("validation succeed！")
		// 	continue // asset = out.Asset
		// } else {
		// 	if out.Asset == nil {
		// 		return modules.TxValidationCode_INVALID_ASSET
		// 	}
		// 	if !out.Asset.IsSimilar(payment.Outputs[i-1].Asset) {
		// 		return modules.TxValidationCode_INVALID_ASSET
		// 	}
		// }
		if out.Value <= 0 || out.Value > 100000000000000000 {
			log.Debug("The OutPut value is :", "amount", out.Value)
			return modules.TxValidationCode_INVALID_AMOUNT
		}
	}
	return modules.TxValidationCode_VALID
}

/**
验证Unit
Validate unit
*/
// modified by Albert·Gou 新生产的unit暂时还没有群签名
//func (validate *Validate) ValidateUnit(unit *modules.Unit, isGenesis bool) byte {
func (validate *Validate) ValidateUnitExceptGroupSig(unit *modules.Unit, isGenesis bool) byte {
	//  unit's size  should bigger than minimum.
	if unit.Size() < 125 {
		log.Debug("Validate size", "error", "size is invalid", "size", unit.Size())
		return modules.UNIT_STATE_INVALID_SIZE
	}

	// step1. check header.New unit is no group signature yet
	//TODO must recover

	sigState := validate.validateHeaderExceptGroupSig(unit.UnitHeader, isGenesis)
	if sigState != modules.UNIT_STATE_VALIDATED &&
		sigState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED && sigState != modules.UNIT_STATE_CHECK_HEADER_PASSED {
		log.Debug("Validate unit's header failed.", "error code", sigState)
		return sigState
	}

	// step2. check transactions in unit
	//_, isSuccess, err := validate.ValidateTransactions(&unit.Txs, isGenesis)
	isSuccess := true //TODO test for sync
	var err error
	if isSuccess != true {
		msg := fmt.Sprintf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
		log.Debug(msg)
		return modules.UNIT_STATE_HAS_INVALID_TRANSACTIONS
	}
	return sigState
}

// modified by Albert·Gou 新生产的unit暂时还没有群签名
//func (validate *Validate) validateHeader(header *modules.Header, isGenesis bool) byte {
func (validate *Validate) validateHeaderExceptGroupSig(header *modules.Header, isGenesis bool) byte {
	// todo yangjie 应当错误返回前，打印验错误的具体消息
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
		//ptnAssetID, _ := modules.SetIdTypeByHex(dagconfig.DefaultConfig.PtnAssetHex)
		asset := modules.NewPTNAsset()
		ptnAssetID := asset.AssetId
		if header.AssetIDs[0] != ptnAssetID || !header.Number.IsMain || header.Number.Index != 0 {
			fmt.Println(6)
			fmt.Println(header.AssetIDs[0].String())
			fmt.Println(ptnAssetID.String())
			return modules.UNIT_STATE_INVALID_HEADER
		}

		// 	return modules.UNIT_STATE_CHECK_HEADER_PASSED
	}
	var isValidAssetId bool
	for _, asset := range header.AssetIDs {
		if asset == header.Number.AssetID {
			isValidAssetId = true
			break
		}
	}
	if !isValidAssetId {
		fmt.Println(7)
		return modules.UNIT_STATE_INVALID_HEADER
	}

	// check authors
	//TODO must recover
	//if header.Authors.Empty() {
	//	return modules.UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	//}

	// comment by Albert·Gou 新生产的unit暂时还没有群签名
	//if len(header.GroupSign) < 64 {
	//	return modules.UNIT_STATE_INVALID_HEADER_WITNESS
	//}

	// TODO 同步过来的unit 没有Authors ，因此无法验证签名有效性。
	var thisUnitIsNotTransmitted bool

	if thisUnitIsNotTransmitted {
		sigState := validate.ValidateUnitSignature(header, isGenesis)
		return sigState
	}
	return modules.UNIT_STATE_VALIDATED
}

func (validate *Validate) validateContractdeploy(tplId []byte, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode {
	return modules.TxValidationCode_VALID
}

func (validate *Validate) validateContractSignature(sinatures []modules.SignatureSet, tx *modules.Transaction, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode {
	return modules.TxValidationCode_VALID
}

func (validate *Validate) checkTxIsExist(tx *modules.Transaction) bool {
	if len(tx.TxMessages) > 2 {
		reqId := tx.RequestHash()
		if txHash, err := validate.dagdb.GetTxHashByReqId(reqId); err == nil && txHash != (common.Hash{}) {
			log.Debug("checkTxIsExist", "transactions exist in dag, reqId:", reqId.String())
			return true
		}
	}
	return false
}
