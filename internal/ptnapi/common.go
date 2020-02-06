package ptnapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
)

func buildRawTransferTx(b Backend, tokenId, from, to string, amount, gasFee decimal.Decimal) (
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
	fromAddr, err := common.StringToAddress(from)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	toAddr, err := common.StringToAddress(to)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	ptnAmount := uint64(0)
	ptn := dagconfig.DagConfig.GasToken
	if tokenId == ptn {
		ptnAmount = ptnjson.Ptn2Dao(amount)
	}

	//构造转移PTN的Message0
	dbUtxos, err := b.GetAddrRawUtxos(from)
	if err != nil {
		return nil, nil, fmt.Errorf("GetAddrRawUtxos utxo err")
	}
	poolTxs, _ := b.GetPoolTxsByAddr(from)
	//if len(poolTxs) == 0 {
	//       return nil, nil, fmt.Errorf("GetPoolTxsByAddr utxo err")
	//}

	utxosPTN, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, ptn)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool utxo err")
	}
	feeAmount := ptnjson.Ptn2Dao(gasFee)
	pay1, usedUtxo1, err := createPayment(fromAddr, toAddr, ptnAmount, feeAmount, utxosPTN)
	if err != nil {
		return nil, nil, err
	}
	tx := modules.NewTransaction([]*modules.Message{modules.NewMessage(modules.APP_PAYMENT, pay1)})
	if tokenId == ptn {
		return tx, usedUtxo1, nil
	}
	//构造转移Token的Message1
	utxosToken, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, tokenId)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool token utxo err")
	}
	tokenAmount := ptnjson.JsonAmt2AssetAmt(tokenAsset, amount)
	pay2, usedUtxo2, err := createPayment(fromAddr, toAddr, tokenAmount, 0, utxosToken)
	if err != nil {
		return nil, nil, err
	}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay2))
	//for _, u := range usedUtxo2 {
	usedUtxo1 = append(usedUtxo1, usedUtxo2...)
	//}
	return tx, usedUtxo1, nil
}
func signRawTransaction(rawTx *modules.Transaction,
	ks *keystore.KeyStore, from, password string, duration time.Duration, hashType uint32,
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
	fromAddr, err := common.StringToAddress(from)
	if err != nil {
		return err
	}
	if password != "" {
		err = ks.TimedUnlock(accounts.Account{Address: fromAddr}, password, duration)
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
