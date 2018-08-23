package common

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"strings"
)

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
func ValidateTransactions(txs *modules.Transactions, isGenesis bool) (map[common.Hash]modules.TxValidationCode, bool, error) {
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
			txFlags[tx.TxHash] = modules.TxValidationCode_DUPLICATE_TXID
			continue
		}
		// validate common property
		txCode := ValidateTx(tx, &worldState)
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

		coinIn, ok := (*txs)[0].TxMessages[0].Payload.(modules.PaymentPayload)
		if !ok {
			return nil, false, fmt.Errorf("Coinbase payload type error.")
		}
		if len(coinIn.Outputs) != 1 {
			return nil, false, fmt.Errorf("Coinbase outputs error.")
		}
		income := uint64(fee) + ComputeInterest()
		if coinIn.Outputs[0].Value != income {
			return nil, false, fmt.Errorf("Coinbase outputs error.")
		}
	}
	return txFlags, isSuccess, nil
}

/**
验证某个交易
To validate one transaction
*/
func ValidateTx(tx *modules.Transaction, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode {
	for _, msg := range tx.TxMessages {
		// check message type and payload
		if !validateMessageType(msg.App, msg.Payload) {
			return modules.TxValidationCode_UNKNOWN_TX_TYPE
		}
		// validate tx size
		if tx.Size().Float64() > float64(modules.TX_MAXSIZE) {
			log.Debug("Tx size is to big.\n")
			return modules.TxValidationCode_NOT_COMPARE_SIZE
		}
		// validate transaction hash
		if strings.Compare(tx.TxHash.String(), tx.Hash().String()) != 0 {
			fmt.Printf("tx.TxHash.String()=%s, tx.Hash()=%s\n", tx.TxHash.String(), tx.Hash().String())
			return modules.TxValidationCode_NIL_TXACTION
		}
		// validate transaction signature
		if validateTxSignature(tx) == false {
			return modules.TxValidationCode_BAD_CREATOR_SIGNATURE
		}
		// validate every type payload
		switch msg.App {
		case modules.APP_PAYMENT:

		case modules.APP_CONTRACT_TPL:
			payload, _ := msg.Payload.(modules.ContractTplPayload)
			validateCode := validateContractTplPayload(&payload)
			if validateCode != modules.TxValidationCode_VALID {
				return validateCode
			}
		case modules.APP_CONTRACT_DEPLOY:
			payload, _ := msg.Payload.(modules.ContractDeployPayload)
			validateCode := validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet, worldTmpState)
			if validateCode != modules.TxValidationCode_VALID {
				return validateCode
			}
		case modules.APP_CONTRACT_INVOKE:
			payload, _ := msg.Payload.(modules.ContractInvokePayload)
			validateCode := validateContractState(payload.ContractId, &payload.ReadSet, &payload.WriteSet, worldTmpState)
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
func validateMessageType(app string, payload interface{}) bool {
	switch payload.(type) {
	case modules.PaymentPayload:
		if app == modules.APP_PAYMENT {
			return true
		}
	case modules.ContractTplPayload:
		if app == modules.APP_CONTRACT_TPL {
			return true
		}
	case modules.ContractDeployPayload:
		if app == modules.APP_CONTRACT_DEPLOY {
			return true
		}
	case modules.ContractInvokePayload:
		if app == modules.APP_CONTRACT_INVOKE {
			return true
		}
	case modules.ConfigPayload:
		if app == modules.APP_CONFIG {
			return true
		}
	case modules.TextPayload:
		if app == modules.APP_TEXT {
			return true
		}
	default:
		return false
	}
	return false
}

/**
验证单元的签名，需要比对见证人列表
To validate unit's signature, and mediators' signature
*/
func ValidateUnitSignature(h *modules.Header, isGenesis bool) error {
	if h.Authors == nil || len(h.Authors.Address) <= 0 {
		return fmt.Errorf("No author info")
	}
	emptySigUnit := modules.Unit{}
	// copy unit's header
	emptySigUnit.UnitHeader = modules.CopyHeader(h)
	// signature does not contain authors and witness fields
	emptySigUnit.UnitHeader.Authors = nil
	emptySigUnit.UnitHeader.Witness = []*modules.Authentifier{}
	// recover signature
	sig := make([]byte, 65)
	copy(sig[32-len(h.Authors.R):32], h.Authors.R)
	copy(sig[64-len(h.Authors.S):64], h.Authors.S)
	copy(sig[64:], h.Authors.V)
	// recover pubkey
	hash := crypto.Keccak256Hash(util.RHashBytes(*emptySigUnit.UnitHeader))
	pubKey, err := modules.RSVtoPublicKey(hash[:], h.Authors.R[:], h.Authors.S[:], h.Authors.V[:])
	if err != nil {
		return fmt.Errorf("Verify unit signature when recover pubkey error:%s", err.Error())
	}
	//  pubKey to pubKey_bytes
	pubKey_bytes := crypto.FromECDSAPub(pubKey)
	if keystore.VerifyUnitWithPK(sig, *emptySigUnit.UnitHeader, pubKey_bytes) == false {
		return fmt.Errorf("Verify unit signature error.")
	}
	// if genesis unit just return
	if isGenesis == true {
		return nil
	}
	// todo group signature verify
	// get mediators
	data := GetConfig([]byte("MediatorCandidates"))
	var mList []string
	if err := rlp.DecodeBytes(data, &mList); err != nil {
		return fmt.Errorf("Check unit signature when get mediators list error: %s", err.Error())
	}
	bNum := GetConfig([]byte("ActiveMediators"))
	var mNum uint16
	if err := rlp.DecodeBytes(bNum, &mNum); err != nil {
		return fmt.Errorf("Check unit signature error: %s", err.Error())
	}
	if int(mNum) != len(mList) {
		return fmt.Errorf("Check unit signature error: mediators info error, pls update network")
	}

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

/**
对unit中某个交易的读写集进行验证
To validate read set and write set of one transaction in unit'
*/
func validateContractState(contractID []byte, readSet *[]modules.ContractReadSet, writeSet *[]modules.PayloadMapStruct, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode {
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
func validateContractTplPayload(contractTplPayload *modules.ContractTplPayload) modules.TxValidationCode {
	// to check template whether existing or not
	stateVersion, bytecode, name, path := storage.GetContractTpl(contractTplPayload.TemplateId.Bytes())
	if stateVersion == nil && bytecode == nil && name == "" && path == "" {
		return modules.TxValidationCode_VALID
	}
	return modules.TxValidationCode_INVALID_CONTRACT_TEMPLATE
}
