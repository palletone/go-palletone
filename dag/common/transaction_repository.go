package common

import (
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
)

func GetTxSig(tx *modules.Transaction, ks *keystore.KeyStore, signer common.Address) ([]byte, error) {
	sign, err := ks.SigData(tx, signer)
	if err != nil {
		msg := fmt.Sprintf("Failed to singure transaction:%v", err)
		log.Error(msg)
		return nil, errors.New(msg)
	}

	return sign, nil
}

func ValidateTxSig(tx *modules.Transaction, ks *keystore.KeyStore) bool {
	var sigs []modules.SignatureSet
	if tx == nil {
		return false
	}

	tmpTx := modules.Transaction{}
	tmpTx.TxHash = tx.TxHash
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigs = msg.Payload.(modules.SignaturePayload).Signatures
		} else {
			tmpTx.TxMessages = append(tmpTx.TxMessages, msg)
		}
	}

	if len(sigs) > 0 {
		for i := 0; i < len(sigs); i++ {
			if keystore.VerifyTXWithPK(sigs[i].Signature, tmpTx, sigs[i].PubKey) != true {
				log.Error("ValidateTxSig", "VerifyTXWithPK sig[%d] fail", i)
				return false
			}
		}
	}

	return true
}
