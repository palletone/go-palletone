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

package dag

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core/types"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/memunit"
	"github.com/palletone/go-palletone/dag/migration"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
)

type Dag struct {
	//Cache       palletcache.ICache
	Db          ptndb.Database
	currentUnit atomic.Value

	unstableUnitRep        dagcommon.IUnitRepository
	unstableUtxoRep        dagcommon.IUtxoRepository
	unstableStateRep       dagcommon.IStateRepository
	unstablePropRep        dagcommon.IPropRepository
	unstableUnitProduceRep dagcommon.IUnitProduceRepository

	stableUnitRep  dagcommon.IUnitRepository
	stableUtxoRep  dagcommon.IUtxoRepository
	stableStateRep dagcommon.IStateRepository
	stablePropRep  dagcommon.IPropRepository

	stableUnitProduceRep dagcommon.IUnitProduceRepository
	//validate             validator.Validator
	ChainHeadFeed *event.Feed

	Mutex           sync.RWMutex
	Memdag          memunit.IMemDag                     // memory unit
	PartitionMemDag map[modules.AssetId]memunit.IMemDag //其他分区的MemDag

	applyLock sync.Mutex

	//SPV
	rmLogsFeed    event.Feed
	chainFeed     event.Feed
	chainSideFeed event.Feed
	chainHeadFeed event.Feed
	logsFeed      event.Feed
	scope         event.SubscriptionScope
}

//type MemUtxos map[modules.OutPoint]*modules.Utxo
func cache() palletcache.ICache {
	return freecache.NewCache(1000 * 1024)
}
func (d *Dag) IsEmpty() bool {
	it := d.Db.NewIterator()
	return !it.Next()
}

// 返回稳定单元
func (d *Dag) CurrentUnit(token modules.AssetId) *modules.Unit {
	memdag, err := d.getMemDag(token)
	if err != nil {
		log.Errorf("Get CurrentUnit by token[%s] error:%s", token.String(), err.Error())
		return nil
	}
	stable_hash, _ := memdag.GetLastStableUnitInfo()
	unit, err := d.GetUnitByHash(stable_hash)
	if err != nil {
		return nil
	}
	return unit
}

func (d *Dag) GetMainCurrentUnit() *modules.Unit {
	return d.Memdag.GetLastMainChainUnit()
}

func (d *Dag) GetCurrentUnit(assetId modules.AssetId) *modules.Unit {
	memUnit := d.GetCurrentMemUnit(assetId, 0)
	curUnit := d.CurrentUnit(assetId)

	if memUnit == nil {
		return curUnit
	}
	if curUnit.NumberU64() >= memUnit.NumberU64() {
		return curUnit
	}
	return memUnit
}

func (d *Dag) GetCurrentMemUnit(assetId modules.AssetId, index uint64) *modules.Unit {
	memdag, err := d.getMemDag(assetId)
	if err != nil {
		log.Errorf("Get CurrentUnit by token[%s] error:%s", assetId.String(), err.Error())
		return nil
	}
	curUnit := memdag.GetLastMainChainUnit()

	return curUnit
}

func (d *Dag) HasUnit(hash common.Hash) bool {
	u, err := d.unstableUnitRep.GetUnit(hash)
	if err != nil {
		return false
	}
	return u != nil
}
func (d *Dag) IsTransactionExist(hash common.Hash) (bool, error) {
	return d.unstableUnitRep.IsTransactionExist(hash)
}

// confirm unit
func (d *Dag) UnitIsConfirmedByHash(hash common.Hash) bool {
	if d.HasUnit(hash) {
		return true
	}
	return false
}

//confirm unit's parent
func (d *Dag) ParentsIsConfirmByHash(hash common.Hash) bool {
	unit, err := d.GetUnitByHash(hash)
	if err != nil {
		return false
	}
	parents := unit.ParentHash()
	if len(parents) > 0 {
		par := parents[0]
		return d.HasUnit(par)
	}
	return false
}

func (d *Dag) GetUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error) {
	//return d.unstableUnitRep.GetUnitFormIndex(number)
	hash, err := d.unstableUnitRep.GetHashByNumber(number)
	if err != nil {
		log.Debug("GetUnitByNumber dagdb.GetHashByNumber err:", "error", err)
		return nil, err
	}
	//log.Debug("Dag", "GetUnitByNumber getChainUnit(hash):", hash)
	return d.unstableUnitRep.GetUnit(hash)
}

func (d *Dag) GetUnstableUnits() []*modules.Unit {
	units := d.Memdag.GetChainUnits()
	result := modules.Units{}
	for _, u := range units {
		result = append(result, u)
	}
	sort.Sort(result)
	return result
}
func (d *Dag) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	uHeader, err := d.unstableUnitRep.GetHeaderByHash(hash)
	if errors.IsNotFoundError(err) {
		uHeader, err = d.getHeaderByHashFromPMemDag(hash)
	}
	if err != nil {
		log.Debug("GetHeaderByHash failed", "error", err.Error())
		return nil, err
	}
	return uHeader, nil
}
func (d *Dag) getHeaderByHashFromPMemDag(hash common.Hash) (*modules.Header, error) {
	for _, memdag := range d.PartitionMemDag {
		h, e := memdag.GetHeaderByHash(hash)
		if e == nil {
			return h, e
		}
	}
	return nil, errors.ErrNotFound
}
func (d *Dag) getHeaderByNumberFromPMemDag(number *modules.ChainIndex) (*modules.Header, error) {
	for _, memdag := range d.PartitionMemDag {
		h, e := memdag.GetHeaderByNumber(number)
		if e == nil {
			return h, e
		}
	}
	return nil, errors.ErrNotFound
}
func (d *Dag) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	uHeader, err := d.unstableUnitRep.GetHeaderByNumber(number)
	if errors.IsNotFoundError(err) {
		uHeader, err = d.getHeaderByNumberFromPMemDag(number)
	}
	if err != nil {
		log.Info("GetHeaderByNumber failed ", "error:", err, "number", number.String())
		return nil, err
	}
	return uHeader, nil
}

//func (d *Dag) GetPrefix(prefix string) map[string][]byte {
//	return d.unstableUnitRep.GetPrefix(*(*[]byte)(unsafe.Pointer(&prefix)))
//}

//func (d *Dag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
//	return d.ChainHeadFeed.Subscribe(ch)
//}

// FastSyncCommitHead sets the current head block to the one defined by the hash
// irrelevant what the chain contents were prior.
func (d *Dag) FastSyncCommitHead(hash common.Hash) error {
	unit, err := d.GetUnitByHash(hash)
	if err != nil {
		return fmt.Errorf("non existent unit [%x...]", hash[:4])
	}
	// store current unit
	d.Mutex.Lock()
	d.currentUnit.Store(unit)
	d.Mutex.Unlock()

	return nil
}

// InsertDag attempts to insert the given batch of blocks in to the canonical
// chain or, otherwise, create a fork. If an error is returned it will return
// the index number of the failing block as well an error describing what went
// wrong.
// After insertion is done, all accumulated events will be fired.
// reference : Eth InsertChain
func (d *Dag) InsertDag(units modules.Units, txpool txspool.ITxPool) (int, error) {

	count := int(0)
	for i, u := range units {
		// todo 此处应判断第0个unit的父unit是否已存入本节点

		// all units must be continuous
		if i > 0 && units[i].UnitHeader.Number.Index == units[i-1].UnitHeader.Number.Index+1 {
			return count, fmt.Errorf("Insert dag error: child height are not continuous, "+
				"parent unit number=%d, hash=%s; "+"child unit number=%d, hash=%s",
				units[i-1].UnitHeader.Number.Index, units[i-1].UnitHash,
				units[i].UnitHeader.Number.Index, units[i].UnitHash)
		}
		if i > 0 && u.ContainsParent(units[i-1].UnitHash) == false {
			return count, fmt.Errorf("Insert dag error: child parents are not continuous, "+
				"parent unit number=%d, hash=%s; "+"child unit number=%d, hash=%s",
				units[i-1].UnitHeader.Number.Index, units[i-1].UnitHash,
				units[i].UnitHeader.Number.Index, units[i].UnitHash)
		}
		t1 := time.Now()
		timestamp := time.Unix(u.Timestamp(), 0)
		log.Debugf("Start InsertDag unit(%v) #%v parent(%v) @%v signed by %v", u.UnitHash.TerminalString(),
			u.NumberU64(), u.ParentHash()[0].TerminalString(), timestamp.Format("2006-01-02 15:04:05"),
			u.Author().Str())

		if a, b, c, dd, e, err := d.Memdag.AddUnit(u, txpool); err != nil {
			//return count, err
			log.Errorf("Memdag addUnit[%s] error:%s", u.UnitHash.String(), err.Error())
			return count, nil
		} else {
			if a != nil {
				d.unstableUnitRep = a
				d.unstableUtxoRep = b
				d.unstableStateRep = c
				d.unstablePropRep = dd
				d.unstableUnitProduceRep = e
			}
		}
		log.Debugf("InsertDag[%s] #%d spent time:%s", u.UnitHash.String(), u.NumberU64(), time.Since(t1))
		count += 1
	}

	return count, nil
}

// GetBlockHashesFromHash retrieves a number of block hashes starting at a given
// hash, fetching towards the genesis block.
func (d *Dag) GetUnitHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	header, err := d.GetHeaderByHash(hash)
	if err != nil {
		return nil
	}
	// Iterate the headers until enough is collected or the genesis reached
	chain := make([]common.Hash, 0, max)
	for i := uint64(0); i < max; i++ {
		if header.Index() == 0 {
			break
		}
		next := header.ParentsHash[0]
		h, err := d.unstableUnitRep.GetHeaderByHash(next)
		if err != nil {
			break
		}
		header = h
		chain = append(chain, next)
	}
	return chain
}

// need add:   assetId modules.AssetId, onMain bool
func (d *Dag) HasHeader(hash common.Hash, number uint64) bool {
	h, _ := d.GetHeaderByHash(hash)
	return h != nil
}

func (d *Dag) IsHeaderExist(hash common.Hash) bool {
	exist, _ := d.unstableUnitRep.IsHeaderExist(hash)
	return exist
}

func (d *Dag) CurrentHeader(token modules.AssetId) *modules.Header {
	unit := d.CurrentUnit(token)
	if unit != nil {
		return unit.Header()
	}
	return nil
}

// GetBodyRLP retrieves a block body in RLP encoding from the database by hash,
// caching it if found.
//func (d *Dag) GetBodyRLP(hash common.Hash) rlp.RawValue {
//	return d.getBodyRLP(hash)
//}

// GetUnitTransactions is return unit's body, all transactions of unit.
func (d *Dag) GetUnitTransactions(hash common.Hash) (modules.Transactions, error) {
	return d.unstableUnitRep.GetUnitTransactions(hash)
}

// GetUnitTxsHash is return the unit's txs hash list.
func (d *Dag) GetUnitTxsHash(hash common.Hash) ([]common.Hash, error) {
	return d.unstableUnitRep.GetBody(hash)
}

// GetTransactionByHash is return the tx by tx's hash
//func (d *Dag) GetTransactionByHash(hash common.Hash) (*modules.Transaction, common.Hash, error) {
//	tx, uhash, _, _, err := d.unstableUnitRep.GetTransaction(hash)
//	if err != nil {
//		return nil, uhash, errors.New("get transaction by hash is failed,none the transaction.")
//	}
//	return tx, uhash, nil
//}
func (d *Dag) GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error) {
	return d.unstableUnitRep.GetTransaction(hash)
}
func (d *Dag) GetTxByReqId(reqid common.Hash) (*modules.TransactionWithUnitInfo, error) {
	hash, err := d.unstableUnitRep.GetTxHashByReqId(reqid)
	if err != nil {
		return nil, err
	}
	return d.unstableUnitRep.GetTransaction(hash)
}

func (d *Dag) GetTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	return d.unstableUnitRep.GetTransactionOnly(hash)
}
func (d *Dag) GetTxSearchEntry(hash common.Hash) (*modules.TxLookupEntry, error) {
	txlookup, err := d.unstableUnitRep.GetTxLookupEntry(hash)
	return txlookup, err
}

// InsertHeaderDag attempts to insert the given header chain in to the local
// chain, possibly creating a reorg. If an error is returned, it will return the
// index number of the failing header as well an error describing what went wrong.
//
// The verify parameter can be used to fine tune whether nonce verification
// should be done or not. The reason behind the optional check is because some
// of the header retrieval mechanisms already need to verify nonces, as well as
// because nonces can be verified sparsely, not needing to check each.
func (d *Dag) InsertHeaderDag(headers []*modules.Header) (int, error) {
	for i, header := range headers {
		err := d.saveHeader(header)

		if err != nil {
			return i, fmt.Errorf("InsertHeaderDag, on header:%d, at saveHeader Error:%s", i, err.Error())
		}

	}
	return len(headers), nil
}

//VerifyHeader checks whether a header conforms to the consensus rules of the stock
//Ethereum ethash engine.go
//func (d *Dag) VerifyHeader(header *modules.Header) error {
//	// step1. check unit signature, should be compare to mediator list
//	unitState := validator.NewValidateError(d.validate.ValidateHeader(header))
//	if unitState != nil {
//		log.Errorf("Validate unit header error, errno=%s", unitState.Error())
//		return unitState
//	}
//
//	// step2. check extra data
//	// Ensure that the header's extra-data section is of a reasonable size
//	if uint64(len(header.Extra)) > uint64(32) {
//		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), configure.MaximumExtraDataSize)
//	}
//
//	return nil
//}

/**
获取account address下面的token信息
To get account token list and tokens's information
*/
//func (d *Dag) WalletTokens(addr common.Address) (map[string]*modules.AccountToken, error) {
//	return d.unstableUtxoRep.GetAccountTokens(addr)
//}
//
//func (d *Dag) WalletBalance(address common.Address, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
//	newAssetid := modules.AssetId{}
//	newUnitqueid := modules.AssetId{}
//
//	if len(assetid) != cap(newAssetid) {
//		return 0, fmt.Errorf("Assetid lenth is wrong")
//	}
//	if len(uniqueid) != cap(newUnitqueid) {
//		return 0, fmt.Errorf("Uniqueid lenth is wrong")
//	}
//	if chainid == 0 {
//		return 0, fmt.Errorf("Chainid is invalid")
//	}
//
//	newAssetid.SetBytes(assetid)
//	newUnitqueid.SetBytes(uniqueid)
//
//	asset := modules.Asset{
//		AssetId:  newAssetid,
//		UniqueId: newUnitqueid,
//	}
//
//	return d.unstableUtxoRep.WalletBalance(address, asset), nil
//}

// Utxos : return mem utxos
//func (d *Dag) Utxos() map[common.Hash]map[modules.OutPoint]*modules.Utxo {
//	// result := d.utxos_cache1
//	// utxos := make(Utxos, 0)
//	// d.utxos_cache1.Range(func(key, v interface{}) bool {
//	// 	utxos[key] = v
//	// })
//	return d.utxos_cache
//}

func (d *Dag) refreshPartitionMemDag() {
	db := d.Db
	unitRep := d.stableUnitRep
	propRep := d.stablePropRep

	mainChain, err := d.stableStateRep.GetMainChain()
	if err == nil && mainChain != nil {
		//分区上要通过Memdag保留PTN的Header
		if d.PartitionMemDag == nil {
			d.PartitionMemDag = make(map[modules.AssetId]memunit.IMemDag)
		}
		mainChainMemDag, ok := d.PartitionMemDag[mainChain.GasToken]
		threshold := int(mainChain.StableThreshold)
		if !ok {
			d.initDataForMainChainHeader(mainChain)
			log.Debugf("Init main chain mem dag for:%s", mainChain.GasToken.String())
			pmemdag := memunit.NewMemDag(mainChain.GasToken, threshold, true, db, unitRep, propRep, d.stableStateRep, cache())
			//pmemdag.SetUnstableRepositories(d.unstableUnitRep, d.unstableUtxoRep, d.unstableStateRep, d.unstablePropRep, d.unstableUnitProduceRep)
			d.PartitionMemDag[mainChain.GasToken] = pmemdag
		} else {
			mainChainMemDag.SetStableThreshold(threshold) //可能更新了该数字
		}
	} else {
		log.Info("Don't have main chain config for partition")
	}
	partitions, err := d.stableStateRep.GetPartitionChains()
	if err != nil {
		log.Warnf("GetPartitionChains error:%s", err.Error())
		return
	}
	log.Debug("Start to refresh partition mem dag")
	//Init partition memdag
	if d.PartitionMemDag == nil {
		partitionMemdag := make(map[modules.AssetId]memunit.IMemDag)

		for _, partition := range partitions {
			ptoken := partition.GasToken
			threshold := int(partition.StableThreshold)
			d.initDataForPartition(partition)
			log.Debugf("Init partition mem dag for:%s", ptoken.String())
			pmemdag := memunit.NewMemDag(ptoken, threshold, true, db, unitRep, propRep, d.stableStateRep, cache())
			//pmemdag.SetUnstableRepositories(d.unstableUnitRep, d.unstableUtxoRep, d.unstableStateRep, d.unstablePropRep, d.unstableUnitProduceRep)
			partitionMemdag[ptoken] = pmemdag
		}

		d.PartitionMemDag = partitionMemdag
		return
	}
	//Exist! update
	for _, partition := range partitions {
		ptoken := partition.GasToken
		threshold := int(partition.StableThreshold)
		partitonMemDag, ok := d.PartitionMemDag[ptoken]
		if !ok {
			d.initDataForPartition(partition)
			log.Debugf("Init partition mem dag for:%s", ptoken.String())
			pmemdag := memunit.NewMemDag(ptoken, threshold, true, db, unitRep, propRep, d.stableStateRep, cache())
			//pmemdag.SetUnstableRepositories(d.unstableUnitRep, d.unstableUtxoRep, d.unstableStateRep, d.unstablePropRep, d.unstableUnitProduceRep)
			d.PartitionMemDag[ptoken] = pmemdag
		} else {
			partitonMemDag.SetStableThreshold(threshold) //可能更新了该数字
		}
	}

}

func (d *Dag) initDataForPartition(partition *modules.PartitionChain) {
	pHeader := partition.GetGenesisHeader()
	exist, _ := d.stableUnitRep.IsHeaderExist(pHeader.Hash())
	if !exist {
		log.Debugf("Init partition[%s] genesis header:%s", pHeader.ChainIndex().AssetID.String(), pHeader.Hash().String())
		d.stableUnitRep.SaveNewestHeader(pHeader)
	}
}
func (d *Dag) initDataForMainChainHeader(mainChain *modules.MainChain) {
	pHeader := mainChain.GetGenesisHeader()
	exist, _ := d.stableUnitRep.IsHeaderExist(pHeader.Hash())
	if !exist {
		log.Debugf("Init main chain[%s] genesis header:%s", pHeader.ChainIndex().AssetID.String(), pHeader.Hash().String())
		d.stableUnitRep.SaveNewestHeader(pHeader)
	}
}
func NewDag(db ptndb.Database, light bool) (*Dag, error) {
	mutex := new(sync.RWMutex)

	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	checkDbMigration(db, stateDb)
	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, propDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)
	stableUnitProduceRep := dagcommon.NewUnitProduceRepository(unitRep, propRep, stateRep)
	//hash, idx, _ := stablePropRep.GetLastStableUnit(modules.PTNCOIN)
	gasToken := dagconfig.DagConfig.GetGasToken()
	threshold, _ := propRep.GetChainThreshold()
	cache := freecache.NewCache(1000 * 1024)
	unstableChain := memunit.NewMemDag(gasToken, threshold, light /*false*/, db, unitRep, propRep, stateRep, cache)
	tunitRep, tutxoRep, tstateRep, tpropRep, tUnitProduceRep := unstableChain.GetUnstableRepositories()
	//validate := validator.NewValidate(tunitRep, tutxoRep, tstateRep, tpropRep)
	//partitionMemdag := make(map[modules.AssetId]memunit.IMemDag)
	//for _, ptoken := range dagconfig.DagConfig.GeSyncPartitionTokens() {
	//	partitionMemdag[ptoken] = memunit.NewMemDag(ptoken, true, db, unitRep, propRep, stateRep)
	//}

	dag := &Dag{
		//Cache:            freecache.NewCache(200 * 1024 * 1024),
		Db:                     db,
		unstableUnitRep:        tunitRep,
		unstableUtxoRep:        tutxoRep,
		unstableStateRep:       tstateRep,
		unstablePropRep:        tpropRep,
		unstableUnitProduceRep: tUnitProduceRep,
		stablePropRep:          propRep,
		stableUnitRep:          unitRep,
		stableUtxoRep:          utxoRep,
		stableStateRep:         stateRep,
		stableUnitProduceRep:   stableUnitProduceRep,
		//validate:               validate,
		ChainHeadFeed: new(event.Feed),
		Mutex:         *mutex,
		Memdag:        unstableChain,
		//PartitionMemDag:      partitionMemdag,
	}
	dag.stableUnitRep.SubscribeSysContractStateChangeEvent(dag.AfterSysContractStateChangeEvent)
	dag.stableUnitProduceRep.SubscribeChainMaintenanceEvent(dag.AfterChainMaintenanceEvent)
	// 检查NewestUnit是否存在，不存在则从MemDag获取最新的Unit作为NewestUnit
	hash, chainIndex, _ := dag.stablePropRep.GetNewestUnit(gasToken)
	if !dag.IsHeaderExist(hash) {
		log.Debugf("Newest unit[%s] not exist in dag, retrieve another from memdag "+
			"and update NewestUnit.index [%d]", hash.String(), chainIndex.Index)
		//TODO Devin query all unit,get newest one
		//newestUnit := dag.Memdag.GetLastMainchainUnit()
		//if nil != newestUnit {

		//	dgp := dag.GetDynGlobalProp()
		//
		//	interval := dag.GetGlobalProp().ChainParameters.MediatorInterval
		//	time, _ := dag.stablePropRep.GetNewestUnitTimestamp(gasToken)
		//	dgp.CurrentASlot -= uint64(uint8(time-newestUnit.Timestamp()) / interval)
		//	//dgp.CurrentASlot += newestUnit.NumberU64() - chainIndex.Index
		//
		//	dag.SaveDynGlobalProp(dgp, false)
		//	//----------
		//
		//	dag.stablePropRep.SetNewestUnit(newestUnit.Header())
		//}
	}
	dag.refreshPartitionMemDag()
	return dag, nil
}
func checkDbMigration(db ptndb.Database, stateDb storage.IStateDb) error {
	// 获取旧的gptn版本号
	t := time.Now()
	old_vertion, err := stateDb.GetDataVersion()
	if err != nil {
		old_vertion = &modules.DataVersion{Version: "0.6.15"}
	}
	log.Debugf("the database version is:%s", old_vertion.Version)
	// 获取当前gptn版本号
	now_version := configure.Version
	log.Debugf("the program version is:%s", now_version)
	next_version := old_vertion.Version
	if next_version != now_version {
		log.Infof("Start migration,upgrade gtpn vertion[%s] to [%s]", next_version, now_version)
		// migrations
		mig_versions := migration.NewMigrations(db)
		for {
			if mig, has := mig_versions[next_version]; has {
				if err := mig.ExecuteUpgrade(); err != nil {
					return err
				}
				next_version = mig.ToVersion()
				data_version := new(modules.DataVersion)
				data_version.Name = "gptn"
				data_version.Version = next_version
				stateDb.SaveDataVersion(data_version)
			}
			if next_version == now_version {
				break
			}
			// 版本升级超时处理
			if now := time.Now(); now.After(t.Add(1 * time.Minute)) {
				log.Infof("upgrade gptn failed. error: timeout[%s]", time.Since(t))
				break
			}
		}
		log.Infof("Complete migration, spent time:%s", time.Since(t))
		return nil
	}
	return nil
}

func (dag *Dag) AfterSysContractStateChangeEvent(arg *modules.SysContractStateChangeEvent) {
	log.Debug("Process AfterSysContractStateChangeEvent")
	if bytes.Equal(arg.ContractId, syscontract.PartitionContractAddress.Bytes()) {
		//分区合约进行了修改，刷新PartitionMemDag
		dag.refreshPartitionMemDag()
	}
}
func (dag *Dag) AfterChainMaintenanceEvent(arg *modules.ChainMaintenanceEvent) {
	log.Debug("Process AfterChainMaintenanceEvent")
	//换届完成，dag需要进行的操作：
	threshold, _ := dag.stablePropRep.GetChainThreshold()
	dag.Memdag.SetStableThreshold(threshold)
}
func NewDag4GenesisInit(db ptndb.Database) (*Dag, error) {
	mutex := new(sync.RWMutex)
	//logger := log.New("Dag")
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, propDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	//validate := validator.NewValidate(dagDb, utxoRep, stateDb, nil)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)

	statleUnitProduceRep := dagcommon.NewUnitProduceRepository(unitRep, propRep, stateRep)

	dag := &Dag{
		//Cache:         freecache.NewCache(200 * 1024 * 1024),
		Db:                   db,
		stableUnitRep:        unitRep,
		stableUtxoRep:        utxoRep,
		stablePropRep:        propRep,
		stableStateRep:       stateRep,
		stableUnitProduceRep: statleUnitProduceRep,
		//validate:             validate,
		ChainHeadFeed: new(event.Feed),
		Mutex:         *mutex,
		//Memdag:        memunit.NewMemDag(dagDb, stateDb, unstableUnitRep),
		//utxos_cache: make(map[common.Hash]map[modules.OutPoint]*modules.Utxo),
	}

	return dag, nil
}

func NewDagForTest(db ptndb.Database) (*Dag, error) {
	mutex := new(sync.RWMutex)
	//logger := log.New("Dag")
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)
	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, propDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	statleUnitProduceRep := dagcommon.NewUnitProduceRepository(unitRep, propRep, stateRep)

	threshold, _ := propRep.GetChainThreshold()
	//validate := validator.NewValidate(dagDb, utxoRep, stateDb, propRep)
	unstableChain := memunit.NewMemDag(modules.PTNCOIN, threshold, false, db, unitRep, propRep, stateRep, cache())
	tunitRep, tutxoRep, tstateRep, tpropRep, tUnitProduceRep := unstableChain.GetUnstableRepositories()

	dag := &Dag{
		//Cache:            freecache.NewCache(200 * 1024 * 1024),
		Db:                   db,
		stableUnitRep:        unitRep,
		stableUtxoRep:        utxoRep,
		stableStateRep:       stateRep,
		stablePropRep:        propRep,
		stableUnitProduceRep: statleUnitProduceRep,
		//validate:               validate,
		ChainHeadFeed:          new(event.Feed),
		Mutex:                  *mutex,
		Memdag:                 unstableChain,
		unstableUnitRep:        tunitRep,
		unstableUtxoRep:        tutxoRep,
		unstableStateRep:       tstateRep,
		unstablePropRep:        tpropRep,
		unstableUnitProduceRep: tUnitProduceRep,
		//utxos_cache:   make(map[common.Hash]map[modules.OutPoint]*modules.Utxo),
	}
	return dag, nil
}

func (d *Dag) GetChaincodes(contractId common.Address) (*list.CCInfo, error) {
	return d.stablePropRep.GetChaincodes(contractId)
}

func (d *Dag) SaveChaincode(contractId common.Address, cc *list.CCInfo) error {
	return d.stablePropRep.SaveChaincode(contractId, cc)
}

// Get Contract Api
func (d *Dag) GetContract(id []byte) (*modules.Contract, error) {
	return d.unstableStateRep.GetContract(id)
}

func (d *Dag) GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error) {
	return d.unstableStateRep.GetContractDeploy(tempId, contractId, name)
}

// Get UnitNumber
func (d *Dag) GetUnitNumber(hash common.Hash) (*modules.ChainIndex, error) {
	return d.unstableUnitRep.GetNumberWithUnitHash(hash)
}

func (d *Dag) GetTrieSyncProgress() (uint64, error) {
	return d.unstableUnitRep.GetTrieSyncProgress()
}

func (d *Dag) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	return d.unstableUtxoRep.GetUtxoEntry(outpoint)
}
func (d *Dag) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	return d.unstableUtxoRep.GetStxoEntry(outpoint)
}
func (d *Dag) GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error) {
	neededSet := make(map[modules.OutPoint]struct{})

	for _, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				if !msg.IsCoinbase() {
					for _, in := range msg.Inputs {
						neededSet[*in.PreviousOutPoint] = struct{}{}
					}
				}
			}
		}
	}

	view := txspool.NewUtxoViewpoint()
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	err := view.FetchUtxos(d.unstableUtxoRep, neededSet)
	return view, err
}

func (d *Dag) GetUtxosOutViewbyTx(tx *modules.Transaction) *txspool.UtxoViewpoint {
	view := txspool.NewUtxoViewpoint()
	view.AddTxOuts(tx)
	return view
}

func (d *Dag) GetUtxosOutViewbyUnit(unit *modules.Unit) *txspool.UtxoViewpoint {
	txs := unit.Transactions()
	view := txspool.NewUtxoViewpoint()
	for _, tx := range txs {
		vi := d.GetUtxosOutViewbyTx(tx)
		for key, utxo := range vi.Entries() {
			view.AddUtxo(key, utxo)
		}
	}
	return view
}
func (d *Dag) IsUtxoSpent(outpoint *modules.OutPoint) (bool, error) {
	return d.unstableUtxoRep.IsUtxoSpent(outpoint)
}

// GetAllUtxos is return all utxo.
func (d *Dag) GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error) {
	d.Mutex.RLock()
	items, err := d.unstableUtxoRep.GetAllUtxos()
	d.Mutex.RUnlock()

	return items, err
}

func (d *Dag) GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error) {
	all, err := d.unstableUtxoRep.GetAddrOutpoints(addr)

	return all, err
}

func (d *Dag) GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error) {
	utxo, err := d.unstableUtxoRep.GetUtxoEntry(outPoint)
	if err != nil {
		return common.Address{}, err
	}
	return tokenengine.GetAddressFromScript(utxo.PkScript)
}

func (d *Dag) GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error) {
	return d.unstableUtxoRep.ComputeTxFee(pay)
}

func (d *Dag) GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error) {
	return d.unstableUnitRep.GetTxFromAddress(tx)
}

func (d *Dag) GetAssetTxHistory(asset *modules.Asset) ([]*modules.TransactionWithUnitInfo, error) {
	return d.unstableUnitRep.GetAssetTxHistory(asset)
}

func (d *Dag) GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error) {
	all, err := d.unstableUtxoRep.GetAddrUtxos(addr, asset)
	return all, err
}

func (d *Dag) GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	all, err := d.unstableUtxoRep.GetAddrUtxos(addr, nil)

	return all, err
}

func (d *Dag) RefreshSysParameters() {
	d.unstableUnitProduceRep.RefreshSysParameters()
}

func (d *Dag) GetAddrTransactions(addr common.Address) ([]*modules.TransactionWithUnitInfo, error) {
	return d.unstableUnitRep.GetAddrTransactions(addr)
}

// get contract state
func (d *Dag) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return d.unstableStateRep.GetContractState(id, field)
	//return d.statedb.GetContractState(common.HexToAddress(id), field)
}

//get contract all state
func (d *Dag) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	return d.unstableStateRep.GetContractStatesById(id)
}

func (d *Dag) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return d.unstableStateRep.GetContractStatesByPrefix(id, prefix)
}
func (d *Dag) GetContractJury(contractId []byte) ([]modules.ElectionInf, error) {
	return d.unstableStateRep.GetContractJury(contractId)

}
func (d *Dag) CreateUnit(mAddr common.Address, txpool txspool.ITxPool, t time.Time) (*modules.Unit, error) {
	_, _, _, rep, _ := d.Memdag.GetUnstableRepositories()

	return d.unstableUnitRep.CreateUnit(mAddr, txpool, rep, t)
}

func (d *Dag) saveHeader(header *modules.Header) error {
	if header == nil {
		return errors.ErrNullPoint
	}
	unit := &modules.Unit{UnitHeader: header}
	asset := header.Number.AssetID
	memdag, err := d.getMemDag(asset)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if a, b, c, dd, e, err := memdag.AddUnit(unit, nil); err != nil {
		return fmt.Errorf("Save MemDag, occurred error: %s", err.Error())
	} else {
		if a != nil {
			d.unstableUnitRep = a
			d.unstableUtxoRep = b
			d.unstableStateRep = c
			d.unstablePropRep = dd
			d.unstableUnitProduceRep = e
		}
	}

	return nil
}

func (d *Dag) getMemDag(asset modules.AssetId) (memunit.IMemDag, error) {
	var memdag memunit.IMemDag
	gasToken := dagconfig.DagConfig.GetGasToken()
	if asset == gasToken {
		memdag = d.Memdag
	} else {
		memdag = d.PartitionMemDag[asset]
		if memdag == nil {
			return nil, errors.New("Don't have partition mem dag for token:" + asset.String())
		}
	}
	return memdag, nil
}

// TODO 待优化, 目前只用来存创世unit
func (d *Dag) SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool) error {
	// todo 应当根据新的unit判断哪条链作为主链
	// step1. check exists
	if !isGenesis {
		if d.IsHeaderExist(unit.Hash()) {
			log.Debug("dag:the unit is already exist in leveldb. ", "unit_hash", unit.Hash().String())
			return errors.ErrUnitExist
		}
		// step2. validate unit
		//Genesis unit don't need validate
		//err := d.validate.ValidateUnitExceptGroupSig(unit)
		//if err != nil {
		//	return fmt.Errorf("SaveDag, validate unit error, err=%s", err.Error())
		//}
	}
	if isGenesis {
		d.stableUnitRep.SaveUnit(unit, true)
		return nil
	}

	if a, b, c, dd, e, err := d.Memdag.AddUnit(unit, txpool); err != nil {
		return fmt.Errorf("Save MemDag, occurred error: %s", err.Error())
	} else {
		if a != nil {
			d.unstableUnitRep = a
			d.unstableUtxoRep = b
			d.unstableStateRep = c
			d.unstablePropRep = dd
			d.unstableUnitProduceRep = e
		}
	}

	return nil
}

func (d *Dag) GetGenesisUnit() (*modules.Unit, error) {
	return d.stableUnitRep.GetGenesisUnit()
}
func (d *Dag) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return d.unstableStateRep.GetContractTpl(tplId)
}
func (d *Dag) GetAllContractTpl() ([]*modules.ContractTemplate, error) {
	return d.unstableStateRep.GetAllContractTpl()
}

func (d *Dag) GetContractTplCode(tplId []byte) ([]byte, error) {
	return d.unstableStateRep.GetContractTplCode(tplId)
}

func (d *Dag) GetCurrentUnitIndex(token modules.AssetId) (*modules.ChainIndex, error) {
	currentUnit := d.GetCurrentUnit(token)
	//	return d.GetUnitNumber(currentUnitHash)
	return currentUnit.Number(), nil
}

// dag's common geter
func (d *Dag) GetCommon(key []byte) ([]byte, error) {
	return d.unstableUnitRep.GetCommon(key)
}

// GetCommonByPrefix  return the prefix's all key && value.
func (d *Dag) GetCommonByPrefix(prefix []byte) map[string][]byte {
	return d.unstableUnitRep.GetCommonByPrefix(prefix)
}

func (d *Dag) SaveCommon(key, val []byte) error {
	return d.stableUnitRep.SaveCommon(key, val)
}

func (d *Dag) SetUnitGroupSign(unitHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error {
	if groupSign == nil {
		err := fmt.Errorf("this unit(%v)'s group sign is null", unitHash.TerminalString())
		log.Debug(err.Error())
		return err
	}

	// 判断本节点是否正在同步数据
	if !d.IsSynced() {
		err := "this node is syncing"
		log.Debugf(err)
		return fmt.Errorf(err)
	}

	if d.IsIrreversibleUnit(unitHash) {
		log.Debugf("this unit(%v) is irreversible", unitHash.TerminalString())
		return nil
	}

	// 验证群签名：
	err := d.VerifyUnitGroupSign(unitHash, groupSign)
	if err != nil {
		return err
	}
	// 群签之后， 更新memdag，将该unit和它的父单元们稳定存储。
	//go d.Memdag.SetStableUnit(unitHash, groupSign[:], txpool)
	log.Debugf("Try to update unit[%s] group sign", unitHash.String())
	d.Memdag.SetUnitGroupSign(unitHash /*, nil*/, groupSign, txpool)

	//TODO albert 待合并
	// 状态更新
	//go d.updateGlobalPropDependGroupSign(unitHash)
	return nil
}

func (d *Dag) QueryDbByKey(key []byte) ([]byte, error) {
	return d.Db.Get(key)
}

func (d *Dag) QueryDbByPrefix(prefix []byte) ([]*modules.DbRow, error) {

	iter := d.Db.NewIteratorWithPrefix(prefix)
	result := []*modules.DbRow{}
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		result = append(result, &modules.DbRow{Key: key, Value: value})
	}
	return result, nil
}

// GetTxHashByReqId
func (d *Dag) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return d.unstableUnitRep.GetTxHashByReqId(reqid)
}

// GetReqIdByTxHash
//func (d *Dag) GetReqIdByTxHash(hash common.Hash) (common.Hash, error) {
//	return d.unstableUnitRep.GetReqIdByTxHash(hash)
//}

// GetFileInfo
func (d *Dag) GetFileInfo(filehash []byte) ([]*modules.FileInfo, error) {
	return d.unstableUnitRep.GetFileInfo(filehash)
}

//Light Palletone Subprotocal
func (d *Dag) GetLightHeaderByHash(headerHash common.Hash) (*modules.Header, error) {
	return nil, nil
}
func (d *Dag) GetLightChainHeight(assetId modules.AssetId) uint64 {
	header := d.CurrentHeader(assetId)
	if header != nil {
		return header.Number.Index
	}
	return uint64(0)
}
func (d *Dag) InsertLightHeader(headers []*modules.Header) (int, error) {
	log.Debug("Dag InsertLightHeader numbers", "", len(headers))
	for _, header := range headers {
		log.Debug("Dag InsertLightHeader info", "header index:", header.Index(), "assetid", header.Number.AssetID)
	}
	count, err := d.InsertHeaderDag(headers)

	return count, err
}

//All leaf nodes for dag downloader.
//MUST have Priority.
//根据资产Id返回所有链的header。
func (d *Dag) GetAllLeafNodes() ([]*modules.Header, error) {
	// step1: get all AssetId
	partitions, _ := d.unstableStateRep.GetPartitionChains()
	leafs := []*modules.Header{}
	for _, partition := range partitions {
		tokenId := partition.GasToken
		pMemdag, ok := d.PartitionMemDag[tokenId]
		if ok {
			unit := pMemdag.GetLastMainChainUnit()
			leafs = append(leafs, unit.UnitHeader)
		} else {
			log.Warnf("Token[%s] is a patition, but not in dag.PartitionMemDag", tokenId.String())
		}
	}
	return leafs, nil
}

//SPV
// SubscribeRemovedLogsEvent registers a subscription of RemovedLogsEvent.
func (bc *Dag) SubscribeRemovedLogsEvent(ch chan<- modules.RemovedLogsEvent) event.Subscription {
	return bc.scope.Track(bc.rmLogsFeed.Subscribe(ch))
}

// SubscribeChainHeadEvent registers a subscription of ChainHeadEvent.
func (bc *Dag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}

// SubscribeLogsEvent registers a subscription of []*types.Log.
func (bc *Dag) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return bc.scope.Track(bc.logsFeed.Subscribe(ch))
}

// SubscribeChainEvent registers a subscription of ChainEvent.
func (bc *Dag) SubscribeChainEvent(ch chan<- modules.ChainEvent) event.Subscription {
	return bc.scope.Track(bc.chainFeed.Subscribe(ch))
}

// PostChainEvents iterates over the events generated by a chain insertion and
// posts them into the event feed.
// TODO: Should not expose PostChainEvents. The chain events should be posted in WriteBlock.
func (bc *Dag) PostChainEvents(events []interface{}) {
	log.Debug("enter PostChainEvents")

	for _, event := range events {
		switch ev := event.(type) {
		case modules.ChainEvent:
			bc.chainFeed.Send(ev)

		case modules.ChainHeadEvent:
			bc.chainHeadFeed.Send(ev)

			//case modules.ChainSideEvent:
			//	bc.chainSideFeed.Send(ev)
		}
	}
}
func (bc *Dag) GetPartitionChains() ([]*modules.PartitionChain, error) {
	return bc.unstableStateRep.GetPartitionChains()
}
func (bc *Dag) GetMainChain() (*modules.MainChain, error) {
	return bc.unstableStateRep.GetMainChain()
}

// SubscribeChainSideEvent registers a subscription of ChainSideEvent.
//func (bc *Dag) SubscribeChainSideEvent(ch chan<- ChainSideEvent) event.Subscription {
//	return bc.scope.Track(bc.chainSideFeed.Subscribe(ch))
//}

func (d *Dag) GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	return d.unstableUnitRep.GetTxRequesterAddress(tx)
}
func (d *Dag) RefreshAddrTxIndex() error {
	return d.stableUnitRep.RefreshAddrTxIndex()
}
func (d *Dag) GetAllContracts() ([]*modules.Contract, error) {
	return d.unstableStateRep.GetAllContracts()
}
func (d *Dag) GetContractsByTpl(tplId []byte) ([]*modules.Contract, error) {
	return d.unstableStateRep.GetContractsByTpl(tplId)
}

func (d *Dag) GetMinFee() (*modules.AmountAsset, error) {
	return d.unstableStateRep.GetMinFee()
}

func (d *Dag) SubscribeActiveMediatorsUpdatedEvent(ch chan<- modules.ActiveMediatorsUpdatedEvent) event.Subscription {
	return d.unstableUnitProduceRep.SubscribeActiveMediatorsUpdatedEvent(ch)
}

func (d *Dag) Close() {
	d.unstableUnitProduceRep.Close()
	d.Memdag.Close()

	for _, pmg := range d.PartitionMemDag {
		pmg.Close()
	}

	d.Db.Close()
	log.Debug("Close all dag database connections")
}

func (dag *Dag) MediatorVotedResults() (map[string]uint64, error) {
	return dag.unstableStateRep.GetMediatorVotedResults()
}

func (dag *Dag) StoreDataVersion(dv *modules.DataVersion) error {
	return dag.stableStateRep.StoreDataVersion(dv)
}

func (dag *Dag) GetDataVersion() (*modules.DataVersion, error) {
	return dag.stableStateRep.GetDataVersion()
}

func (dag *Dag) QueryProofOfExistenceByReference(ref []byte) ([]*modules.ProofOfExistence, error) {
	return dag.stableUnitRep.QueryProofOfExistenceByReference(ref)
}

func (dag *Dag) GetAssetReference(asset []byte) ([]*modules.ProofOfExistence, error) {
	return dag.stableUnitRep.GetAssetReference(asset)
}
