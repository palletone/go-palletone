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

func ValidateTxSig(tx *modules.Transaction) bool {

	return true
}
