package ptnapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
)

func parseAddressStr(addr string, ks *keystore.KeyStore, password string) (common.Address, error) {
	addrArray := strings.Split(addr, ":")
	addrString := addrArray[0]
	if len(addrArray) == 2 { //HD wallet address format
		toAccountIndex, err := strconv.Atoi(addrArray[1])
		if err != nil {
			return common.Address{}, errors.New("invalid addrString address format")
		}
		hdBaseAddr, err := common.StringToAddress(addrArray[0])
		if err != nil {
			return common.Address{}, errors.New("invalid addrString address format")
		}
		var toAccount accounts.Account
		if ks.IsUnlock(hdBaseAddr) {
			toAccount, err = ks.GetHdAccount(accounts.Account{Address: hdBaseAddr}, uint32(toAccountIndex))
		} else {
			toAccount, err = ks.GetHdAccountWithPassphrase(accounts.Account{Address: hdBaseAddr}, password, uint32(toAccountIndex))
		}
		if err != nil {
			return common.Address{}, errors.New("GetHdAccountWithPassphrase error:" + err.Error())
		}
		addrString = toAccount.Address.String()
	}
	return common.StringToAddress(addrString)
}

func buildRawTransferTx(b Backend, tokenId, fromStr, toStr string, amount, gasFee decimal.Decimal, password string) (
	*modules.Transaction, []*modules.UtxoWithOutPoint, error) {
	//参数检查
	tokenAsset, err := modules.StringToAsset(tokenId)
	if err != nil {
		return nil, nil, err
	}
	if !gasFee.IsPositive() {
		return nil, nil, fmt.Errorf("fee is ZERO ")
	}
	//
	fromAddr, err := parseAddressStr(fromStr, b.GetKeyStore(), password)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	from := fromAddr.String()
	toAddr, err := parseAddressStr(toStr, b.GetKeyStore(), password)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	//to:=toAddr.String()
	ptnAmount := uint64(0)
	gasToken := dagconfig.DagConfig.GasToken
	gasAsset := dagconfig.DefaultConfig.GetGasToken()
	if tokenId == gasToken {
		ptnAmount = gasAsset.Uint64Amount(amount)
	}
	//构造转移PTN的Message0
	dbUtxos, err := b.GetAddrRawUtxos(from)
	if err != nil {
		return nil, nil, fmt.Errorf("GetAddrRawUtxos utxo err")
	}
	poolTxs, _ := b.GetPoolTxsByAddr(from)

	utxosPTN, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, gasToken)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool utxo err")
	}
	feeAmount := gasAsset.Uint64Amount(gasFee)
	pay1, usedUtxo1, err := createPayment(fromAddr, toAddr, ptnAmount, feeAmount, utxosPTN)
	if err != nil {
		return nil, nil, err
	}
	tx := modules.NewTransaction([]*modules.Message{modules.NewMessage(modules.APP_PAYMENT, pay1)})
	if tokenId == gasToken {
		return tx, usedUtxo1, nil
	}
	//构造转移Token的Message1
	utxosToken, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, tokenId)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool token utxo err")
	}
	tokenAmount := tokenAsset.Uint64Amount(amount)
	pay2, usedUtxo2, err := createPayment(fromAddr, toAddr, tokenAmount, 0, utxosToken)
	if err != nil {
		return nil, nil, err
	}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay2))
	usedUtxo1 = append(usedUtxo1, usedUtxo2...)

	return tx, usedUtxo1, nil
}
func signRawTransaction(rawTx *modules.Transaction,
	ks *keystore.KeyStore, fromStr, password string, duration *time.Duration, hashType uint32,
	usedUtxo []*modules.UtxoWithOutPoint) error {
	//lockscript
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		return ks.GetPublicKey(addr)
	}
	//sign tx
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		return ks.SignMessage(addr, msg)
	}
	utxoLockScripts := make(map[modules.OutPoint][]byte)
	for _, utxo := range usedUtxo {
		utxoLockScripts[utxo.OutPoint] = utxo.PkScript
	}
	fromAddr, err := parseAddressStr(fromStr, ks, password)
	if err != nil {
		return err
	}
	if password != "" {
		if duration != nil {
			err = ks.TimedUnlock(accounts.Account{Address: fromAddr}, password, *duration)
		} else {
			err = ks.Unlock(accounts.Account{Address: fromAddr}, password)
		}
		if err != nil {
			return err
		}
	}
	//Sign tx
	_, err = tokenengine.Instance.SignTxAllPaymentInput(rawTx, hashType, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return err
	}
	//log for debug
	log.DebugDynamic(func() string {
		txJson, _ := json.Marshal(rawTx)
		return "SignedTx:" + string(txJson)
	})
	return nil
}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *modules.Transaction) (common.Hash, error) {
	if tx.IsNewContractInvokeRequest() {
		reqId, err := b.SendContractInvokeReqTx(tx)
		return reqId, err
	}

	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}
