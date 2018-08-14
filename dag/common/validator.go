package common

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"strings"
)

/**
检查unit中所有交易的合法性，返回所有交易的交易费总和
check all transactions in one unit
return all transactions' fee
*/
func ValidateTransactions(txs *modules.Transactions, isGenesis bool) (map[common.Hash]modules.TxValidationCode, bool, error) {
	//fee := uint64(0)
	txFlages := map[common.Hash]modules.TxValidationCode{}
	isSuccess := bool(true)
	for _, tx := range *txs {
		// validate transaction id duplication
		if _, ok := txFlages[tx.TxHash]; ok == true {
			fmt.Println(">>>>> Duplicate transaction:", tx.TxHash)
			isSuccess = false
			txFlages[tx.TxHash] = modules.TxValidationCode_DUPLICATE_TXID
			continue
		}
		// validate common property
		txCode := ValidateTx(tx)
		if txCode != modules.TxValidationCode_VALID {
			fmt.Println(">>>>> ValidateTx error:", txCode)
			isSuccess = false
			txFlages[tx.TxHash] = txCode
			continue
		}
		//// validate fee
		//if isGenesis == false && txIndex != 0 {
		//	// check transaction fee
		//	if tx.TxFee.Cmp(modules.TXFEE) != 0 {
		//		fmt.Println(">>>>> Invalid fee")
		//		isSuccess = false
		//		txFlages[tx.TxHash] = modules.TxValidationCode_NOT_COMPARE_SIZE
		//		continue
		//	}
		//	fee += tx.TxFee.Uint64()
		//}
	}

	// to check total fee with coinbase tx
	return txFlages, isSuccess, nil
}

/**
验证某个交易
To validate one transaction
*/
func ValidateTx(tx *modules.Transaction) modules.TxValidationCode {
	for _, msg := range tx.TxMessages {
		// check message type and payload
		if !validateMessageType(msg.App, msg.Payload) {
			return modules.TxValidationCode_UNKNOWN_TX_TYPE
		}
		// validate tx size
		if tx.Size() != tx.Txsize {
			log.Debug("Txsize=%v, tx.Size()=%v\n", tx.Txsize, tx.Size())
			return modules.TxValidationCode_NOT_COMPARE_SIZE
		}
		// validate transaction hash
		if strings.Compare(tx.TxHash.String(), tx.Hash().String()) != 0 {
			fmt.Printf("tx.TxHash.String()=%s, tx.Hash()=%s\n", tx.TxHash.String(), tx.Hash().String())
			return modules.TxValidationCode_NIL_TXACTION
		}
		//// validate transaction signature
		//if validateTxSignature(tx.Hash(), tx.From) == false {
		//	return modules.TxValidationCode_BAD_CREATOR_SIGNATURE
		//}
		// validate every type payload
		switch msg.App {
		case modules.APP_PAYMENT:

		case modules.APP_CONTRACT_TPL:

		case modules.APP_CONTRACT_DEPLOY:

		case modules.APP_CONTRACT_INVOKE:

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
	//  pubKey to pubKey_bytes
	pubKey_bytes := crypto.FromECDSAPub(pubKey)
	if keystore.VerifyUnitWithPK(sig, *emptySigUnit.UnitHeader, pubKey_bytes) == false {
		return fmt.Errorf("Verify unit signature error.")
	}
	// if genesis unit just return
	if isGenesis == false {
		return nil
	}
	// todo group signature verify
	// get mediators
	data := GetConfig([]byte("MediatorCandidates"))
	bNum := GetConfig([]byte("ActiveMediators"))
	num, err := strconv.Atoi(string(bNum))
	if err != nil {
		return fmt.Errorf("Check unit signature error: %s", err)
	}
	if num != len(data) {
		return fmt.Errorf("Check unit signature error: mediators info error, pls update network")
	}
	// decode mediator list data
	var mediators []string
	if err := rlp.DecodeBytes(data, &mediators); err != nil {
		return fmt.Errorf("Check unit signature error: %s", err)
	}

	return nil
}

/**
验证交易签名
To validate transaction signature
*/
func validateTxSignature(txHash common.Hash, sig *modules.Authentifier) bool {
	// recover signature
	cpySig := make([]byte, 65)
	copy(cpySig[32-len(sig.R):32], sig.R)
	copy(cpySig[64-len(sig.S):64], sig.S)
	copy(cpySig[64:], sig.V)
	// recover pubkey
	hash := crypto.Keccak256Hash(util.RHashBytes(txHash))
	pubKey, err := modules.RSVtoPublicKey(hash[:], sig.R[:], sig.S[:], sig.V[:])
	if err != nil {
		log.Error("Validate transaction signature", "error", err.Error())
		return false
	}
	//  pubKey to pubKey_bytes
	pubKey_bytes := crypto.FromECDSAPub(pubKey)
	if keystore.VerifyUnitWithPK(cpySig, txHash, pubKey_bytes) == true {
		return true
	}
	return false
}

/**
对unit中所有交易的读写集进行验证
To validate read set and write set in all unit transactions'
*/
func validateState(txs *modules.Transactions) {

}
