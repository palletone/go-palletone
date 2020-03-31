/*
 *  This file is part of go-palletone.
 *  go-palletone is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *  go-palletone is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *  You should have received a copy of the GNU General Public License
 *  along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 *
 *  @author PalletOne core developer <dev@pallet.one>
 *  @date 2018-2020
 */

package txpool2

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/txspool"
	"github.com/palletone/go-palletone/validator"
)

var (
	ErrNotFound    = errors.New("txpool: not found")
	ErrDuplicate   = errors.New("txpool: duplicate")
	ErrDoubleSpend = errors.New("txpool: double spend")
	ErrNotSupport  = errors.New("txpool: not support")
)

var Instance txspool.ITxPool

type TxPool struct {
	normals               *txList                                    //普通交易池
	orphans               map[common.Hash]*txspool.TxPoolTransaction //孤儿交易池
	userContractRequests  map[common.Hash]*txspool.TxPoolTransaction //用户合约请求，只参与utxo运算，不会被打包
	basedOnRequestOrphans map[common.Hash]*txspool.TxPoolTransaction //依赖于userContractRequests的孤儿交易池
	txValidator           txspool.IValidator
	dag                   txspool.IDag
	tokenengine           tokenengine.ITokenEngine
	sync.RWMutex
	txFeed event.Feed
	scope  event.SubscriptionScope
}

// NewTxPool creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func NewTxPool(config txspool.TxPoolConfig, cachedb palletcache.ICache, unit txspool.IDag) *TxPool {
	tokenEngine := tokenengine.Instance
	val := validator.NewValidate(unit, unit, unit, unit, nil, cachedb, false)
	pool := NewTxPool4DI(config, cachedb, unit, tokenEngine, val)
	//pool.startJournal(config)
	return pool
}

func NewTxPool4DI(config txspool.TxPoolConfig, cachedb palletcache.ICache, dag txspool.IDag,
	tokenEngine tokenengine.ITokenEngine, txValidator txspool.IValidator) *TxPool {
	return &TxPool{
		normals:               newTxList(),
		orphans:               make(map[common.Hash]*txspool.TxPoolTransaction),
		basedOnRequestOrphans: make(map[common.Hash]*txspool.TxPoolTransaction),
		userContractRequests:  make(map[common.Hash]*txspool.TxPoolTransaction),
		txValidator:           txValidator,
		dag:                   dag,
		tokenengine:           tokenEngine,
	}
}

//支持合约Request，普通FullTx，用户合约FullTx的加入，不支持系统合约FullTx
func (pool *TxPool) AddLocal(tx *modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		data, _ := rlp.EncodeToBytes(tx)
		return fmt.Sprintf("[%s]try to add tx[%s] to txpool, tx hex:%x", tx.RequestHash().ShortStr(), tx.Hash().String(), data)
	})
	err := pool.addLocal(tx)
	if err != nil {
		return err
	}

	return nil
}
func (pool *TxPool) AddRemote(tx *modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		data, _ := rlp.EncodeToBytes(tx)
		return fmt.Sprintf("[%s]try to add tx[%s] to txpool, tx hex:%x", tx.RequestHash().ShortStr(), tx.Hash().String(), data)
	})
	err := pool.addLocal(tx)
	if err != nil {
		return err
	}

	return nil
}
func (pool *TxPool) checkDuplicateAdd(txHash common.Hash) error {
	if _, err := pool.normals.GetTx(txHash); err == nil { //found tx
		log.Infof("ignore add duplicate tx[%s] to tx pool", txHash.String())
		return ErrDuplicate
	}
	if _, ok := pool.orphans[txHash]; ok { //found in orphans
		log.Infof("ignore add duplicate orphan tx[%s] to tx pool", txHash.String())
		return ErrDuplicate
	}
	if _, ok := pool.userContractRequests[txHash]; ok { //found in userContractRequests
		log.Infof("ignore add duplicate user contract request[%s] to tx pool", txHash.String())
		return ErrDuplicate
	}
	if _, ok := pool.basedOnRequestOrphans[txHash]; ok { //found in basedOnRequestOrphans
		log.Infof("ignore add duplicate tx[%s] to tx pool", txHash.String())
		return ErrDuplicate
	}
	return nil
}
func (pool *TxPool) addLocal(tx *modules.Transaction) error {
	//check duplicate add
	txHash := tx.Hash()
	reqHash := tx.RequestHash()
	err := pool.checkDuplicateAdd(txHash)
	if err != nil {
		return nil //重复添加，不用报错
	}
	if tx.IsSystemContract() && !tx.IsOnlyContractRequest() {
		log.Infof("[%s]tx[%s] is a full system contract invoke tx, don't support", reqHash.ShortStr(), txHash.String())
		return ErrNotSupport
	}
	//0. if tx is a full user contract tx, delete request from pool first
	var deletedReq *txspool.TxPoolTransaction
	if tx.IsUserContract() && !tx.IsOnlyContractRequest() { //FullTx about user contract
		//delete request
		var ok bool
		deletedReq, ok = pool.userContractRequests[reqHash]
		if ok {
			delete(pool.userContractRequests, reqHash)
			log.Debugf("[%s]delete user contract request[%s] by hash:%s", reqHash.ShortStr(), reqHash.String(), txHash.String())
		}
	}
	reverseDeleteReq := func() {
		if deletedReq != nil {
			pool.userContractRequests[deletedReq.TxHash] = deletedReq
			log.Debugf("[%s]reverse delete request %s", reqHash.ShortStr(), deletedReq.TxHash.String())
		}
	}
	//1.validate tx
	pool.txValidator.SetUtxoQuery(pool)
	fee, vcode, err := pool.txValidator.ValidateTx(tx, !tx.IsOnlyContractRequest())
	//log.Debugf("[%s]validate tx[%s] get result:%v", reqHash.ShortStr(), txHash.String(), vcode)
	if err != nil && vcode != validator.TxValidationCode_ORPHAN {
		//验证不通过，而且也不是孤儿
		log.Warnf("[%s]validate tx[%s] get error:%s", reqHash.ShortStr(), txHash.String(), err.Error())
		reverseDeleteReq()
		return err
	}
	//2. process orphan
	if vcode == validator.TxValidationCode_ORPHAN {
		return pool.addOrphanTx(pool.convertBaseTx(tx))
	}

	tx2 := pool.convertTx(tx, fee)
	//如果是用户合约请求，则直接添加到RequestPool
	//如果是用户合约FullTx，那么需要判断依赖交易是否还是Request，是则认为是孤儿Tx
	//否则，增加到正常交易池。
	if tx.IsUserContract() && tx.IsOnlyContractRequest() {
		//user contract request
		log.Debugf("[%s]add tx[%s] to user contract request pool", reqHash.ShortStr(), txHash.String())
		pool.userContractRequests[tx2.TxHash] = tx2
		pool.txFeed.Send(modules.TxPreEvent{Tx: tx, IsOrphan: false})
	} else { //不是用户合约请求
		//有可能是连续的用户合约请求R1,R2，但是R2先被执行完，这个时候R1还在RequestPool里面，没办法被打包，所以R2应该被扔到basedOnReqOrphanPool
		//父交易还是Request，所以本Tx是Orphan
		if pool.isBasedOnRequestPool(tx2) {
			log.Debugf("Tx[%s]'s parent or ancestor is a request, not a full tx, add it to based on request pool",
				tx2.TxHash.String())
			if err = pool.addBasedOnReqOrphanTx(tx2); err != nil {
				log.Errorf("add tx[%s] to based on request pool error:%s", tx2.TxHash.String(), err.Error())
				reverseDeleteReq()
				return err
			}
		} else {
			//3. process normal tx
			log.Debugf("!%s!add tx[%s] to normal pool", reqHash.ShortStr(), tx2.TxHash.String())
			err = pool.normals.AddTx(tx2)
			if err != nil {
				log.Errorf("add tx[%s] to normal pool error:%s", tx2.TxHash.String(), err.Error())
				reverseDeleteReq()
				return err
			}
			pool.txFeed.Send(modules.TxPreEvent{Tx: tx, IsOrphan: false})
		}
	}
	//添加了该Tx后，会不会导致其他Orphan变成普通交易？所以需要检查一下
	//4. check orphan txpool
	err = pool.checkBasedOnReqOrphanTxToNormal()
	if err != nil {
		return err
	}
	return pool.checkOrphanTxToNormal(tx2.TxHash)
}
func (pool *TxPool) isBasedOnRequestPool(tx *txspool.TxPoolTransaction) bool {
	for h := range tx.DependOnTxs {
		if _, ok := pool.userContractRequests[h]; ok {
			return true
		}
		if _, ok := pool.basedOnRequestOrphans[h]; ok {
			return true
		}
		for _, tx := range pool.basedOnRequestOrphans {
			if tx.ReqHash == h {
				return true
			}
		}
	}

	return false
}

//检查如果将一个Tx加入Normal后，有没有后续的孤儿Tx需要连带加入
func (pool *TxPool) checkOrphanTxToNormal(txHash common.Hash) error {
	readyTx := []*modules.Transaction{}
	for hash, otx := range pool.orphans {
		if otx.IsFineToNormal(txHash) { //满足Normal的条件了
			log.Debugf("move tx[%s] from orphans to normals", otx.TxHash.String())
			delete(pool.orphans, hash) //从孤儿池删除
			readyTx = append(readyTx, otx.Tx)
		}
	}
	for _, tx := range readyTx {
		err := pool.addLocal(tx) //因为之前孤儿交易没有手续费，UTXO等，所以需要重新计算
		if err != nil {
			log.Warnf("add tx[%s] to pool fail:%s", tx.Hash().String(), err.Error())
		}
	}
	return nil
}
func (pool *TxPool) checkBasedOnReqOrphanTxToNormal() error {
	readyTx := []*modules.Transaction{}
	for hash, otx := range pool.basedOnRequestOrphans {
		if !pool.isBasedOnRequestPool(otx) { //满足Normal的条件了
			log.Debugf("move tx[%s] from based on request orphans to normals", otx.TxHash.String())
			delete(pool.basedOnRequestOrphans, hash) //从孤儿池删除
			readyTx = append(readyTx, otx.Tx)
		}
	}
	for _, tx := range readyTx {
		err := pool.addLocal(tx) //因为之前孤儿交易没有手续费，UTXO等，所以需要重新计算
		if err != nil {
			log.Warnf("add tx[%s] to pool fail:%s", tx.Hash().String(), err.Error())
		}
	}
	return nil
}

func (pool *TxPool) convertBaseTx(tx *modules.Transaction) *txspool.TxPoolTransaction {
	dependOnTxs := make(map[common.Hash]bool)
	for _, o := range tx.GetSpendOutpoints() {
		dependOnTxs[o.TxHash] = false
	}
	txAddr, _ := tx.GetToAddrs(pool.tokenengine.GetAddressFromScript)
	return &txspool.TxPoolTransaction{
		Tx:      tx,
		TxHash:  tx.Hash(),
		ReqHash: tx.RequestHash(),
		//TxFee:                fee,
		CreationDate: time.Now(),
		//FromAddr:             fromAddr,
		DependOnTxs:          dependOnTxs,
		From:                 tx.GetSpendOutpoints(),
		ToAddr:               txAddr,
		IsSysContractRequest: tx.IsOnlyContractRequest() && tx.IsSystemContract(),
		IsUserContractFullTx: tx.IsUserContract() && !tx.IsOnlyContractRequest(),
	}
}
func (pool *TxPool) convertTx(tx *modules.Transaction, fee []*modules.Addition) *txspool.TxPoolTransaction {
	fromAddr, _ := tx.GetFromAddrs(pool.GetUtxoEntry, pool.tokenengine.GetAddressFromScript)
	tx2 := pool.convertBaseTx(tx)
	tx2.TxFee = fee
	tx2.FromAddr = fromAddr
	return tx2
}

func (pool *TxPool) addOrphanTx(tx *txspool.TxPoolTransaction) error {
	log.Debugf("add tx[%s] to orphan pool", tx.TxHash.String())
	tx.Status = txspool.TxPoolTxStatus_Orphan
	pool.orphans[tx.TxHash] = tx
	pool.txFeed.Send(modules.TxPreEvent{Tx: tx.Tx, IsOrphan: true})
	return nil
}
func (pool *TxPool) addBasedOnReqOrphanTx(tx *txspool.TxPoolTransaction) error {
	log.Debugf("add tx[%s] to based on request orphan pool", tx.TxHash.String())
	tx.Status = txspool.TxPoolTxStatus_Orphan
	pool.basedOnRequestOrphans[tx.TxHash] = tx
	pool.txFeed.Send(modules.TxPreEvent{Tx: tx.Tx, IsOrphan: false})
	return nil
}
func (pool *TxPool) GetSortedTxs() ([]*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	return pool.normals.GetSortedTxs()
}

//带锁的对外暴露的查询,只查询Pool的所有新UTXO，不查询DAG
func (pool *TxPool) GetUtxoFromAll(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	pool.RLock()
	defer pool.RUnlock()
	_, newUtxo, reqTxMapping := pool.getAllSpendAndNewUtxo()
	utxo, ok := newUtxo[*outpoint]
	if ok {
		return utxo, nil
	}
	if txHash, ok := reqTxMapping[outpoint.TxHash]; ok {
		o2 := modules.OutPoint{
			TxHash:       txHash,
			MessageIndex: outpoint.MessageIndex,
			OutIndex:     outpoint.OutIndex,
		}
		if utxo, ok := newUtxo[o2]; ok {
			return utxo, nil
		}
	}
	return nil, ErrNotFound

}
func parseTxUtxo(txs []*txspool.TxPoolTransaction, addr common.Address, token *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, map[common.Hash]common.Hash, map[modules.OutPoint]bool) {
	dbUtxos := make(map[modules.OutPoint]*modules.Utxo)
	spendUtxo := make(map[modules.OutPoint]bool)
	dbReqTxMapping := make(map[common.Hash]common.Hash)
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	for _, tx := range txs {
		for k, v := range tx.Tx.GetNewUtxos() {
			if !bytes.Equal(lockScript, v.PkScript) {
				continue
			}
			if token != nil && v.Asset.Equal(token) {
				dbUtxos[k] = v
			}
		}
		for _, so := range tx.Tx.GetSpendOutpoints() {
			spendUtxo[*so] = true
		}
		if tx.TxHash != tx.ReqHash {
			dbReqTxMapping[tx.ReqHash] = tx.TxHash
		}
	}
	return dbUtxos, dbReqTxMapping, spendUtxo
}

func (pool *TxPool) GetAddrUtxos(addr common.Address, token *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, error) {
	dbUtxos, dbReqTxMapping, err := pool.dag.GetAddrUtxoAndReqMapping(addr, token)
	if err != nil {
		return nil, err
	}
	log.DebugDynamic(func() string {
		utxoKeys := ""
		for o := range dbUtxos {
			utxoKeys += o.String() + ";"
		}
		mapping := ""
		for req, tx := range dbReqTxMapping {
			mapping += req.String() + ":" + tx.String() + ";"
		}
		return "db utxo outpoints:" + utxoKeys + " req:tx mapping :" + mapping
	})
	pool.RLock()
	defer pool.RUnlock()

	txs, err := pool.GetUnpackedTxsByAddr(addr)
	if err != nil {
		return nil, err
	}
	log.DebugDynamic(func() string {
		txHashs := ""
		for _, tx := range txs {
			txHashs += "[tx:" + tx.Tx.Hash().String() + "-req:" + tx.Tx.RequestHash().String() + "];"
		}
		return "txpool unpacked tx:" + txHashs
	})
	poolUtxo, poolReqTxMapping, poolSpend := parseTxUtxo(txs, addr, token)
	for k, v := range dbUtxos {
		poolUtxo[k] = v
	}
	for k, v := range dbReqTxMapping {
		poolReqTxMapping[k] = v
	}
	for spend := range poolSpend {
		delete(poolUtxo, spend)
		if txHash, ok := poolReqTxMapping[spend.TxHash]; ok {
			spend2 := modules.OutPoint{
				TxHash:       txHash,
				MessageIndex: spend.MessageIndex,
				OutIndex:     spend.OutIndex,
			}
			delete(poolUtxo, spend2)
		}
	}
	return poolUtxo, nil
}

//返回交易池中花费的UTXO，新产生的UTXO，Req-Tx Mapping
func (pool *TxPool) getAllSpendAndNewUtxo() (map[modules.OutPoint]common.Hash,
	map[modules.OutPoint]*modules.Utxo, map[common.Hash]common.Hash) {
	//查询NormalPool的已花费UTXO，新UTXO和Req-Tx Mapping
	spendUtxoes := make(map[modules.OutPoint]common.Hash)
	for k, v := range pool.normals.spendUtxo {
		spendUtxoes[k] = v
	}
	newUtxoes := make(map[modules.OutPoint]*modules.Utxo)
	for k, v := range pool.normals.newUtxo {
		newUtxoes[k] = v
	}
	reqTxMapping := make(map[common.Hash]common.Hash)
	for req, txHash := range pool.normals.reqTxMap {
		reqTxMapping[req] = txHash
	}
	//查询用户合约Request池的已花费UTXO，新UTXO，这里不会有Req-Tx Mapping
	s1, n1 := getUtxoFromTxs(pool.userContractRequests)
	for k, v := range s1 {
		spendUtxoes[k] = v
	}
	for k, v := range n1 {
		newUtxoes[k] = v
	}
	//查询based on request 池的已花费UTXO，新UTXO和Req-Tx Mapping
	s2, n2 := getUtxoFromTxs(pool.basedOnRequestOrphans)
	for k, v := range s2 {
		spendUtxoes[k] = v
	}
	for k, v := range n2 {
		newUtxoes[k] = v
	}
	for _, tx := range pool.basedOnRequestOrphans {
		if tx.TxHash != tx.ReqHash {
			reqTxMapping[tx.ReqHash] = tx.TxHash
		}
	}
	return spendUtxoes, newUtxoes, reqTxMapping
}

//主要用于Validator，不带锁,从Normal，Request和BasedOnReq三个池获取UTXO，而且禁止双花，如果Pool找不到，就去Dag找
func (pool *TxPool) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	spendUtxoes, newUtxoes, reqTxMapping := pool.getAllSpendAndNewUtxo()
	for o, hash := range spendUtxoes {
		if txHash, ok := reqTxMapping[o.TxHash]; ok {
			o2 := modules.OutPoint{
				TxHash:       txHash,
				MessageIndex: o.MessageIndex,
				OutIndex:     o.OutIndex,
			}
			spendUtxoes[o2] = hash
		}
	}
	if spendTxHash, ok := spendUtxoes[*outpoint]; ok {
		log.Warnf("Utxo(%s) already spend in pool tx[%s]", outpoint.String(), spendTxHash.String())
		return nil, ErrDoubleSpend
	}
	if utxo, ok := newUtxoes[*outpoint]; ok {
		return utxo, nil
	}
	if txHash, ok := reqTxMapping[outpoint.TxHash]; ok {
		o2 := modules.OutPoint{
			TxHash:       txHash,
			MessageIndex: outpoint.MessageIndex,
			OutIndex:     outpoint.OutIndex,
		}
		if utxo, ok := newUtxoes[o2]; ok {
			return utxo, nil
		}
	}
	log.DebugDynamic(func() string {
		return fmt.Sprintf("GetUtxoEntry(%s) not found in pool", outpoint.String())
	})
	return pool.dag.GetUtxoEntry(outpoint)
}

//获得交易列表的消耗的UTXO和新产生的UTXO
func getUtxoFromTxs(txs map[common.Hash]*txspool.TxPoolTransaction) (map[modules.OutPoint]common.Hash, map[modules.OutPoint]*modules.Utxo) {
	newUtxo := make(map[modules.OutPoint]*modules.Utxo)
	spendUtxo := make(map[modules.OutPoint]common.Hash)
	for _, tx := range txs {
		for _, o := range tx.Tx.GetSpendOutpoints() {
			spendUtxo[*o] = tx.TxHash
		}
		for o, u := range tx.Tx.GetNewUtxos() {
			newUtxo[o] = u
		}
	}
	return spendUtxo, newUtxo
}

func (pool *TxPool) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	spendUtxoes, newUtxoes, _ := pool.getAllSpendAndNewUtxo()
	if spendTxHash, ok := spendUtxoes[*outpoint]; ok {
		var utxo *modules.Utxo
		var ok2 bool
		var err error
		if utxo, ok2 = newUtxoes[*outpoint]; !ok2 {
			utxo, err = pool.dag.GetUtxoEntry(outpoint)
			if err != nil {
				return nil, err
			}
		}
		stxo := &modules.Stxo{
			Amount:      utxo.Amount,
			Asset:       utxo.Asset,
			PkScript:    utxo.PkScript,
			LockTime:    utxo.LockTime,
			Timestamp:   utxo.Timestamp,
			SpentByTxId: spendTxHash,
			SpentTime:   0,
		}
		return stxo, nil
	}

	return pool.dag.GetStxoEntry(outpoint)
}

//从交易池删除指定的交易
func (pool *TxPool) DiscardTxs(txs []*modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		hashes := ""
		for _, tx := range txs {
			hashes += tx.Hash().String() + ";"
		}
		return fmt.Sprintf("discard txs: %s", hashes)
	})
	if pool.normals.Count() == 0 {
		return nil
	}
	for _, tx := range txs {
		requestHash := tx.RequestHash()
		if tx.IsContractTx() {
			if tx.IsSystemContract() {
				err := pool.normals.DiscardTx(requestHash)
				if err != nil && err != ErrNotFound {
					log.Warnf("Req[%s] discard error:%s", requestHash.String(), err.Error())
				}
			}
			delete(pool.orphans, requestHash)
			//删除对应的Request,可能有后续Tx在孤儿池，添加回来
			if _, ok := pool.userContractRequests[requestHash]; ok {
				log.Debugf("Request[%s] already packed into unit, delete it from request pool", requestHash.String())
				delete(pool.userContractRequests, requestHash)
				pool.checkBasedOnReqOrphanTxToNormal()
			}
		}
		err := pool.normals.DiscardTx(tx.Hash())
		if err != nil && err != ErrNotFound {
			log.Warnf("Tx[%s] discard error:%s", tx.Hash().String(), err.Error())
		}
		delete(pool.orphans, tx.Hash())
	}
	return nil
}

func (pool *TxPool) GetUnpackedTxsByAddr(addr common.Address) ([]*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	txs, err := pool.normals.GetTxsByStatus(txspool.TxPoolTxStatus_Unpacked)
	if err != nil {
		return nil, err
	}
	for h, tx := range pool.userContractRequests {
		txs[h] = tx
	}
	for h, tx := range pool.basedOnRequestOrphans {
		txs[h] = tx
	}
	result := []*txspool.TxPoolTransaction{}
	for _, tx := range txs {
		if tx.IsFrom(addr) || tx.IsTo(addr) {
			result = append(result, tx)
		}
	}
	return result, nil
}

//func (pool *TxPool) GetUnpackedTxs() (map[common.Hash]*txspool.TxPoolTransaction, error) {
//	return pool.normals.GetTxsByStatus(txspool.TxPoolTxStatus_Unpacked)
//}
func (pool *TxPool) Pending() (map[common.Hash][]*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	packedTxs, err := pool.normals.GetTxsByStatus(txspool.TxPoolTxStatus_Packed)
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash][]*txspool.TxPoolTransaction)
	for _, tx := range packedTxs {
		if txs, ok := result[tx.UnitHash]; ok {
			result[tx.UnitHash] = append(txs, tx)
		} else {
			result[tx.UnitHash] = []*txspool.TxPoolTransaction{tx}
		}
	}
	return result, nil
}
func (pool *TxPool) Queued() ([]*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	result := []*txspool.TxPoolTransaction{}
	for _, tx := range pool.orphans {
		result = append(result, tx)
	}
	return result, nil
}
func (pool *TxPool) Stop() {
	pool.scope.Close()
	log.Info("Transaction pool stopped")
}

//基本状态(未打包，已打包，孤儿)
func (pool *TxPool) Status() (int, int, int) {
	pool.RLock()
	defer pool.RUnlock()
	normals := pool.normals.GetAllTxs()
	packed := 0
	unpacked := 0
	for _, tx := range normals {
		if tx.Status == txspool.TxPoolTxStatus_Packed {
			packed++
		}
		if tx.Status == txspool.TxPoolTxStatus_Unpacked {
			unpacked++
		}
	}
	return unpacked, packed, len(pool.orphans)
}
func (pool *TxPool) Content() (map[common.Hash]*txspool.TxPoolTransaction, map[common.Hash]*txspool.TxPoolTransaction) {
	pool.RLock()
	defer pool.RUnlock()
	return pool.normals.GetAllTxs(), pool.orphans
}

//将交易状态改为已打包
func (pool *TxPool) SetPendingTxs(unitHash common.Hash, num uint64, txs []*modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		hashes := ""
		for _, tx := range txs {
			hashes += tx.Hash().String() + ";"
		}
		return fmt.Sprintf("update status to packed txs: %s", hashes)
	})

	for _, tx := range txs {
		//如果是系统合约，那么需要按RequestHash去查找并改变状态
		if tx.IsSystemContract() {
			if _, err := pool.normals.GetTx(tx.RequestHash()); err != nil {
				//如果有交易没有出现在交易池中，则直接补充
				e := pool.addLocal(tx.GetRequestTx())
				if e != nil {
					return e
				}
			}
			err := pool.normals.UpdateTxStatusPacked(tx.RequestHash(), unitHash, num)
			if err != nil && err != ErrNotFound {
				return err
			}
		} else {
			if _, err := pool.normals.GetTx(tx.Hash()); err != nil {
				//如果有交易没有出现在交易池中，则直接补充
				e := pool.addLocal(tx)
				if e != nil {
					return e
				}
			}
			err := pool.normals.UpdateTxStatusPacked(tx.Hash(), unitHash, num)
			if err != nil && err != ErrNotFound {
				return err
			}
		}
	}
	return nil
}

//将交易状态改为未打包
func (pool *TxPool) ResetPendingTxs(txs []*modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		hashes := ""
		for _, tx := range txs {
			hashes += tx.Hash().String() + ";"
		}
		return fmt.Sprintf("update status to unpacked txs: %s", hashes)
	})
	if pool.normals.Count() == 0 {
		return nil
	}
	for _, tx := range txs {
		if tx.IsSystemContract() {
			err := pool.normals.UpdateTxStatusUnpacked(tx.RequestHash())
			if err != nil && err != ErrNotFound {
				return err
			}
		} else {
			err := pool.normals.UpdateTxStatusUnpacked(tx.Hash())
			if err != nil && err != ErrNotFound {
				return err
			}
		}
	}
	return nil
}
func (pool *TxPool) GetTx(hash common.Hash) (*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	if tx, err := pool.normals.GetTx(hash); err == nil {
		return tx, nil
	}
	if tx, ok := pool.userContractRequests[hash]; ok {
		return tx, nil
	}
	if tx, ok := pool.basedOnRequestOrphans[hash]; ok {
		return tx, nil
	}
	if tx, ok := pool.orphans[hash]; ok {
		return tx, nil
	}
	//4个池都找不到
	return nil, ErrNotFound

}

// SubscribeTxPreEvent registers a subscription of TxPreEvent and
// starts sending event to the given channel.
func (pool *TxPool) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	//return pool.txFeed.Subscribe(ch)
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}
