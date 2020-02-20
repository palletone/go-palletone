package ptnapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
)

func unlockKS(b Backend, addr common.Address, password string, timeout *uint32) error {
	ks := b.GetKeyStore()
	if password != "" {
		if timeout == nil {
			return ks.Unlock(accounts.Account{Address: addr}, password)
		} else {
			d := time.Duration(*timeout) * time.Second
			return ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
		}
	}
	return nil
}

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

func buildRawTransferTx(b Backend, tokenId, fromStr, toStr string, amount, gasFee decimal.Decimal, password string, useMemoryDag bool) (
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
	gasAsset := dagconfig.DagConfig.GetGasToken()
	if tokenId == gasToken {
		ptnAmount = gasAsset.Uint64Amount(amount)
	}

	//构造转移PTN的Message0
	var dbUtxos map[modules.OutPoint]*modules.Utxo
	//if useMemoryDag && cacheTx != nil && cacheTx.mdag != nil {
	//	dbUtxos, err = getAddrUtxofrommDag(fromAddr)
	//} else {
	dbUtxos, err = b.Dag().GetAddrUtxos(fromAddr)
	//}
	if err != nil {
		return nil, nil, fmt.Errorf("GetAddrRawUtxos utxo err")
	}
	poolTxs, _ := b.GetUnpackedTxsByAddr(from)

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
	log.Debugf("gas token[%s], transfer token[%s], start build payment1", gasToken, tokenId)
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

func createPayment(fromAddr, toAddr common.Address, amountToken uint64, feePTN uint64,
	utxosPTN map[modules.OutPoint]*modules.Utxo) (*modules.PaymentPayload, []*modules.UtxoWithOutPoint, error) {

	if len(utxosPTN) == 0 {
		log.Errorf("No PTN Utxo or No Token Utxo for %s", fromAddr.String())
		return nil, nil, fmt.Errorf("No PTN Utxo or No Token Utxo")
	}

	//PTN
	utxoPTNView, asset := convertUtxoMap2Utxos(utxosPTN)

	utxosPTNTaken, change, err := core.Select_utxo_Greedy(utxoPTNView, amountToken+feePTN)
	if err != nil {
		return nil, nil, fmt.Errorf("createPayment Select_utxo_Greedy utxo err")
	}
	usedUtxo := []*modules.UtxoWithOutPoint{}
	//ptn payment
	payPTN := &modules.PaymentPayload{}
	//ptn inputs
	for _, u := range utxosPTNTaken {
		utxo := u.(*modules.UtxoWithOutPoint)
		usedUtxo = append(usedUtxo, utxo)
		prevOut := &utxo.OutPoint // modules.NewOutPoint(txHash, utxo.MessageIndex, utxo.OutIndex)
		txInput := modules.NewTxIn(prevOut, []byte{})
		payPTN.AddTxIn(txInput)
	}

	//ptn outputs
	if amountToken > 0 {
		payPTN.AddTxOut(modules.NewTxOut(amountToken, tokenengine.Instance.GenerateLockScript(toAddr), asset))
	}
	if change > 0 {
		payPTN.AddTxOut(modules.NewTxOut(change, tokenengine.Instance.GenerateLockScript(fromAddr), asset))
	}
	return payPTN, usedUtxo, nil
}

func signRawTransaction(b Backend, rawTx *modules.Transaction, fromStr, password string, timeout *uint32, hashType uint32,
	usedUtxo []*modules.UtxoWithOutPoint) error {
	ks := b.GetKeyStore()
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
	err = unlockKS(b, fromAddr, password, timeout)
	if err != nil {
		return err
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
	if tx.IsNewContractInvokeRequest() && !tx.IsSystemContract() {
		reqId, err := b.SendContractInvokeReqTx(tx)
		return reqId, err
	}
	//普通交易和系统合约交易，走交易池
	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

type synCacheTx struct {
	lock sync.RWMutex
	txs  map[common.Hash]bool //txHash or reqId,  contract or not
	mdag dag.IDag
}

var cacheTx *synCacheTx

func updateDag(b Backend, trs []*modules.Transaction) error {
	log.Debugf("updateDag enter, transaction num:%d, cacheTx num:%d", len(trs), len(cacheTx.txs))

	for idx, tx := range trs {
		if tx.IsContractTx() {
			log.Debugf("updateDag, is     contract Tx,index[%d]---tx[%s]", idx, tx.RequestHash().String())
		} else {
			log.Debugf("updateDag, is not contract Tx,index[%d]---tx[%s]", idx, tx.Hash().String())
		}
	}

	cacheTx.lock.Lock()
	defer cacheTx.lock.Unlock()

	//基于最新单元重新构建新的临时内存数据库
	log.Debug("create new tempdb and dag for api")
	tempdb, err := ptndb.NewTempdb(b.Dag().GetDb())
	if err != nil {
		msg := fmt.Sprintf("updateDag, NewTempdb error:%s", err.Error())
		log.Error(msg)
		return errors.New(msg)
	}
	newDag, err := dag.NewDagSimple(tempdb)
	if err != nil {
		msg := fmt.Sprintf("updateDag, NewDagSimple error:%s", err.Error())
		log.Error(msg)
		return errors.New(msg)
	}

	//将原来内存数据库中的未确认的交易添加到新的内存数据库中
	for _, tx := range trs {
		var txHash common.Hash
		if tx.IsContractTx() {
			txHash = tx.RequestHash()
		} else {
			txHash = tx.Hash()
		}
		if _, ok := cacheTx.txs[txHash]; ok {
			log.Debugf("updateDag, delete tx[%s] from cache", txHash.String())
			delete(cacheTx.txs, txHash)
		}
	}
	oldDag := cacheTx.mdag
	//var unitInfo *modules.TransactionWithUnitInfo
	//for txHash, isContract := range cacheTx.txs {
	//	if isContract {
	//		unitInfo, err = oldDag.GetTxByReqId(txHash)//GetTxByReqId , GetStableTransactionOnly
	//		if err != nil {
	//			log.Warnf("updateDag, GetTxByReqId[%s] err:%s", txHash.String(), err.Error())
	//		}
	//	} else {
	//		unitInfo, err = oldDag.GetTransaction(txHash)
	//		if err != nil {
	//			log.Warnf("updateDag, GetTransaction[%s] err:%s", txHash.String(), err.Error())
	//		}
	//	}
	//	if unitInfo != nil {
	//		log.Debugf("updateDag, SaveTransaction tx[%s]", txHash.String())
	//		err := newDag.SaveTransaction(unitInfo.Transaction, 0)
	//		if err != nil {
	//			log.Errorf("updateDag, SaveTransaction tx[%s] err:%s", txHash.String(), err.Error())
	//			//return err
	//			continue
	//		}
	//	}
	//}

	for txHash, _ := range cacheTx.txs {
		noPackTx, err := oldDag.GetStableTransactionOnly(txHash) //GetTxByReqId , GetStableTransactionOnly
		if err != nil {
			log.Warnf("updateDag, GetStableTransactionOnly tx[%s] err:%s",txHash.String(), err.Error())
			continue
		}
		err = newDag.SaveTransaction(noPackTx, 0)
		if err != nil {
			log.Errorf("updateDag, SaveTransaction tx[%s] err:%s", txHash.String(), err.Error())
			//return err
			continue
		}
	}

	//删除原来的内存数据库
	cacheTx.mdag = newDag
	//log.Debug("updateDag, ok")
	return nil
}

var synDagInited = false

func synDag(b Backend) error {
	log.Debug("synDag enter")
	if synDagInited {
		return nil
	}
	cacheTx = &synCacheTx{
		txs:  make(map[common.Hash]bool),
		mdag: b.Dag(),
	}

	rcvDag := b.Dag()
	headCh := make(chan modules.SaveUnitEvent, 10)
	defer close(headCh)
	headSub := rcvDag.SubscribeSaveUnitEvent(headCh)
	defer headSub.Unsubscribe()

	synDagInited = true
	for {
		select {
		case u := <-headCh:
			log.Debugf("synDag, SubscribeSaveUnitEvent received unit:%s, tx number:%d",
				u.Unit.DisplayId(), len(u.Unit.Txs))
			err := updateDag(b, u.Unit.Transactions())
			if err != nil {
				log.Warnf("synDag, SaveUnitEvent unit number[%s] updateDag err:%s", u.Unit.DisplayId(), err.Error())
			}
		case err := <-headSub.Err():
			return err
		}
	}
}

func saveTransaction2mDag(tx *modules.Transaction) error {
	if cacheTx != nil && cacheTx.mdag != nil {
		cacheTx.lock.Lock()
		defer cacheTx.lock.Unlock()

		if !checkTxContinuous(cacheTx.mdag, tx) {
			log.Errorf("saveTransaction2mDag, tx[%s] checkTxContinuous false", tx.RequestHash().String())
		}
		//var txHash common.Hash
		//if tx.IsContractTx() {
		//	txHash = tx.RequestHash()
		//	log.Debugf("saveTransaction2Mdag, save contract transaction:%s", txHash.String())
		//} else {
		//	txHash = tx.Hash()
		//	log.Debugf("saveTransaction2Mdag, save transaction:%s", txHash.String())
		//}
		txHash := tx.Hash()
		err := cacheTx.mdag.SaveTransaction(tx, 0)
		if err != nil {
			return fmt.Errorf("saveTransaction2Mdag,SaveTransaction[%s] err:%s", txHash.String(), err.Error())
		}
		cacheTx.txs[txHash] = tx.IsContractTx()

		//todo del
		_, err = cacheTx.mdag.GetStableTransactionOnly(tx.Hash())
		if err != nil {
			log.Warnf("saveTransaction2Mdag, GetStableTransactionOnly[%s] err:%s", txHash.String(), err.Error())
		}

		return nil
	}
	return errors.New("saveTransaction2Mdag, no mdag")
}
func checkTxContinuous(dag dag.IDag, tx *modules.Transaction) bool {
	msgs := tx.TxMessages()
	lenTxMsgs := len(msgs)
	if lenTxMsgs > 0 {
		msg0 := msgs[0].Payload.(*modules.PaymentPayload)
		//log.Debugf("checkTxContinuous tx[%s], PreviousOutPoint:%v", tx.RequestHash().String(), msg0.Inputs[0].PreviousOutPoint)
		invokeAddr, err := dag.GetAddrByOutPoint(msg0.Inputs[0].PreviousOutPoint)
		if err != nil {
			log.Debugf("checkTxContinuous tx[%s], err:%s", tx.RequestHash().String(), err.Error())
			return false
		}
		log.Debugf("checkTxContinuous tx[%s],invokeAddr[%s] ok", tx.RequestHash().String(), invokeAddr.String())
	}
	return true
}
func getAddrUtxofrommDag(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	if cacheTx != nil && cacheTx.mdag != nil {
		cacheTx.lock.Lock()
		defer cacheTx.lock.Unlock()
		return cacheTx.mdag.GetAddrUtxos(addr)
	}
	return nil, errors.New("no mdag")
}
