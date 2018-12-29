package common

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
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
	if tx == nil {
		return false
	}
	var sigs []modules.SignatureSet
	tmpTx := &modules.Transaction{}

	//tmpTx.TxId = tx.TxId
	//if !bytes.Equal(tx.TxHash.Bytes(), tx.Hash().Bytes()){
	//log.Info("ValidateTxSig", "transaction hash :", tx.Hash().Bytes())
	//	return false
	//}
	//todo 检查msg的有效性

	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigs = msg.Payload.(*modules.SignaturePayload).Signatures
		} else {
			tmpTx.TxMessages = append(tmpTx.TxMessages, msg)
		}
	}
	//printTxInfo(tmpTx)
	if len(sigs) > 0 {
		for i := 0; i < len(sigs); i++ {
			fmt.Printf("ValidateTxSig sig[%v]-pubkey[%v]\n", sigs[i].Signature, sigs[i].PubKey)
			//if keystore.VerifyTXWithPK(sigs[i].Signature, tmpTx, sigs[i].PubKey) != true {
			//	log.Error("ValidateTxSig", "VerifyTXWithPK sig fail", tmpTx.RequestHash().String())
			//	return false
			//}
		}
	}

	return true
}

func printTxInfo(tx *modules.Transaction) {
	if tx == nil {
		return
	}

	log.Info("=========tx info============hash:", tx.Hash().String())
	for i := 0; i < len(tx.TxMessages); i++ {
		log.Info("---------")
		app := tx.TxMessages[i].App
		pay := tx.TxMessages[i].Payload
		log.Info("", "app:", app)
		if app == modules.APP_PAYMENT {
			p := pay.(*modules.PaymentPayload)
			fmt.Println(p.LockTime)
		} else if app == modules.APP_CONTRACT_INVOKE_REQUEST {
			p := pay.(*modules.ContractInvokeRequestPayload)
			fmt.Println(p.ContractId)
		} else if app == modules.APP_CONTRACT_INVOKE {
			p := pay.(*modules.ContractInvokePayload)
			fmt.Println(p.Args)
			for idx, v := range p.WriteSet {
				fmt.Printf("WriteSet:idx[%d], k[%v]-v[%v]\n", idx, v.Key, v.Value)
			}
			for idx, v := range p.ReadSet {
				fmt.Printf("ReadSet:idx[%d], k[%v]-v[%v]\n", idx, v.Key, v.Value)
			}
		} else if app == modules.APP_SIGNATURE {
			p := pay.(*modules.SignaturePayload)
			fmt.Printf("Signatures:[%v]\n", p.Signatures)
		}
	}
}
