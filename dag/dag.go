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
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/core/types"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/memunit"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/parameter"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/validator"
)

type Dag struct {
	//Cache       palletcache.ICache
	Db          ptndb.Database
	currentUnit atomic.Value

	unstableUnitRep  dagcommon.IUnitRepository
	unstableUtxoRep  dagcommon.IUtxoRepository
	unstableStateRep dagcommon.IStateRepository

	stableUnitRep  dagcommon.IUnitRepository
	stableUtxoRep  dagcommon.IUtxoRepository
	stableStateRep dagcommon.IStateRepository

	propRep       dagcommon.IPropRepository
	validate      validator.Validator
	ChainHeadFeed *event.Feed

	Mutex           sync.RWMutex
	Memdag          memunit.IMemDag                     // memory unit
	PartitionMemDag map[modules.AssetId]memunit.IMemDag //其他分区的MemDag
	// memutxo
	// 按unit单元划分存储Utxo
	//utxos_cache map[common.Hash]map[modules.OutPoint]*modules.Utxo
	// utxos_cache1 sync.Map

	// append by albert·gou 用于活跃mediator更新时的事件订阅
	activeMediatorsUpdatedFeed  event.Feed
	activeMediatorsUpdatedScope event.SubscriptionScope

	// append by albert·gou 用于account 各种投票数据统计
	mediatorVoteTally voteTallys
	applyLock         sync.Mutex

	//SPV
	rmLogsFeed    event.Feed
	chainFeed     event.Feed
	chainSideFeed event.Feed
	chainHeadFeed event.Feed
	logsFeed      event.Feed
	scope         event.SubscriptionScope
}

//type MemUtxos map[modules.OutPoint]*modules.Utxo

func (d *Dag) IsEmpty() bool {
	it := d.Db.NewIterator()
	return !it.Next()
}

func (d *Dag) CurrentUnit(token modules.AssetId) *modules.Unit {
	return d.Memdag.GetLastMainchainUnit(token)
}

func (d *Dag) GetMainCurrentUnit() *modules.Unit {
	main_token := dagconfig.DagConfig.GetGasToken()
	return d.Memdag.GetLastMainchainUnit(main_token)
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
	curUnit := d.Memdag.GetLastMainchainUnit(assetId)

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
	//if d.Memdag.Exists(hash) {
	//	unit, err := d.Memdag.getChainUnit(hash)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return unit.UnitHeader, nil
	//}
	//height, err := d.GetUnitNumber(hash)
	//if err != nil {
	//	log.Debug("GetHeaderByHash when GetUnitNumber", "error", err.Error())
	//	return nil
	//}
	// get unit header
	uHeader, err := d.unstableUnitRep.GetHeaderByHash(hash)
	if err != nil {
		log.Debug("Current unit when get unit header", "error", err.Error())
		return nil, err
	}
	return uHeader, nil
}

func (d *Dag) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	//Query memdag first
	//hash, err := d.Memdag.GetHashByNumber(number)
	//if err == nil { //Exist
	//	unit, err := d.Memdag.getChainUnit(hash)
	//	if err != nil {
	//		log.Errorf("Number[%s] is exist in memdag, but cannot query unit by hash: %s", number.String(), hash.String())
	//		return nil, err
	//	}
	//	return unit.UnitHeader, nil
	//}
	uHeader, err1 := d.unstableUnitRep.GetHeaderByNumber(number)
	if err1 != nil {
		log.Info("getChainUnit when GetHeaderByNumber failed ", "error:", err1, "hash", number.String())
		//log.Info("index info:", "height", number, "index", number.Index, "asset", number.AssetID, "ismain", number.IsMain)
		return nil, err1
	}
	//uHeader, err1 := d.unstableUnitRep.GetHeaderByNumber(number)
	//if err1 != nil {
	//	log.Debug("GetUnit when GetHeaderByHash failed ", "error:", err1, "hash", number.String())
	//	//log.Info("index info:", "height", number, "index", number.Index, "asset", number.AssetID, "ismain", number.IsMain)
	//	return nil, err1
	//}
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
	// var (
	// 	events = make([]interface{}, 0, len(units))
	// 	coalescedLogs []*types.Log
	// )

	count := int(0)
	for i, u := range units {
		// todo 此处应判断第0个unit的父unit是否已存入本节点

		// all units must be continuous
		if i > 0 && units[i].UnitHeader.Number.Index == units[i-1].UnitHeader.Number.Index+1 {
			return count, fmt.Errorf("Insert dag error: child height are not continuous, "+
				"parent unit number=%d, hash=%s; "+
				"child unit number=%d, hash=%s",
				units[i-1].UnitHeader.Number.Index, units[i-1].UnitHash,
				units[i].UnitHeader.Number.Index, units[i].UnitHash)
		}
		if i > 0 && u.ContainsParent(units[i-1].UnitHash) == false {
			return count, fmt.Errorf("Insert dag error: child parents are not continuous, "+
				"parent unit number=%d, hash=%s; "+
				"child unit number=%d, hash=%s",
				units[i-1].UnitHeader.Number.Index, units[i-1].UnitHash,
				units[i].UnitHeader.Number.Index, units[i].UnitHash)
		}

		timestamp := time.Unix(u.Timestamp(), 0)
		log.Infof("InsertDag unit(%v) #%v parent(%v) @%v signed by %v", u.UnitHash.TerminalString(),
			u.NumberU64(), u.ParentHash()[0].TerminalString(), timestamp.Format("2006-01-02 15:04:05"),
			u.Author().Str())

		// append by albert·gou, 利用 unit 更新相关状态
		if err := d.ApplyUnit(u); err != nil {
			//return count, err
			return count, nil
		}

		// todo 应当和本地生产的unit统一接口，而不是直接存储
		//if err := d.unstableUnitRep.SaveUnit(u, false); err != nil {
		if err := d.SaveUnit(u, txpool, false); err != nil {
			return count, err
		}

		//d.updateLastIrreversibleUnitNum(u.Hash(), uint64(u.NumberU64()))
		//log.Debug("Dag", "InsertDag ok index:", u.UnitHeader.Number.Index, "hash:", u.Hash())
		count += 1
		// events = append(events, modules.ChainEvent{u, common.Hash{}, nil})
	}

	//TODO add PostChainEvents
	// d.PostChainEvents(events, coalescedLogs)

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
func (d *Dag) VerifyHeader(header *modules.Header) error {
	// step1. check unit signature, should be compare to mediator list
	unitState := d.validate.ValidateHeader(header)
	if unitState != nil {
		log.Errorf("Validate unit header error, errno=%s", unitState.Error())
		return unitState
	}

	// step2. check extra data
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > uint64(32) {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), configure.MaximumExtraDataSize)
	}

	return nil
}

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

func NewDag(db ptndb.Database) (*Dag, error) {
	mutex := new(sync.RWMutex)

	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb)
	//hash, idx, _ := propRep.GetLastStableUnit(modules.PTNCOIN)
	gasToken := dagconfig.DagConfig.GetGasToken()
	unstableChain := memunit.NewMemDag(gasToken, false, db, unitRep, propRep)
	tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()
	validate := validator.NewValidate(tunitRep, tutxoRep, tstateRep)
	partitionMemdag := make(map[modules.AssetId]memunit.IMemDag)
	for _, ptoken := range dagconfig.DagConfig.GeSyncPartitionTokens() {
		partitionMemdag[ptoken] = memunit.NewMemDag(ptoken, true, db, unitRep, propRep)
	}

	dag := &Dag{
		//Cache:            freecache.NewCache(200 * 1024 * 1024),
		Db:               db,
		unstableUnitRep:  tunitRep,
		unstableUtxoRep:  tutxoRep,
		propRep:          propRep,
		unstableStateRep: tstateRep,
		stableUnitRep:    unitRep,
		stableUtxoRep:    utxoRep,
		stableStateRep:   stateRep,
		validate:         validate,
		ChainHeadFeed:    new(event.Feed),
		Mutex:            *mutex,
		Memdag:           unstableChain,
		PartitionMemDag:  partitionMemdag,
	}

	// 检查NewestUnit是否存在，不存在则从MemDag获取最新的Unit作为NewestUnit
	hash, chainIndex, _ := dag.propRep.GetNewestUnit(gasToken)
	if !dag.IsHeaderExist(hash) {
		log.Debugf("Newest unit[%s] not exist in dag, retrieve another from memdag "+
			"and update NewestUnit.index [%d]", hash.String(), chainIndex.Index)
		newestUnit := dag.Memdag.GetLastMainchainUnit(gasToken)
		if nil != newestUnit {
			// todo 待删除 处理临时prop没有回滚的问题
			dgp := dag.GetDynGlobalProp()

			interval := dag.GetGlobalProp().ChainParameters.MediatorInterval
			time, _ := dag.propRep.GetNewestUnitTimestamp(gasToken)
			dgp.CurrentASlot -= uint64(uint8(time-newestUnit.Timestamp()) / interval)
			//dgp.CurrentASlot += newestUnit.NumberU64() - chainIndex.Index

			dag.SaveDynGlobalProp(dgp, false)
			//----------

			dag.propRep.SetNewestUnit(newestUnit.Header())
		}
	}

	return dag, nil
}

func NewDag4GenesisInit(db ptndb.Database) (*Dag, error) {
	mutex := new(sync.RWMutex)
	//logger := log.New("Dag")
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	validate := validator.NewValidate(dagDb, utxoRep, stateDb)
	propRep := dagcommon.NewPropRepository(propDb)

	dag := &Dag{
		//Cache:         freecache.NewCache(200 * 1024 * 1024),
		Db:            db,
		stableUnitRep: unitRep,
		stableUtxoRep: utxoRep,
		propRep:       propRep,
		validate:      validate,
		ChainHeadFeed: new(event.Feed),
		Mutex:         *mutex,
		//Memdag:        memunit.NewMemDag(dagDb, stateDb, unstableUnitRep),
		//utxos_cache: make(map[common.Hash]map[modules.OutPoint]*modules.Utxo),
	}

	return dag, nil
}

func NewDagForTest(db ptndb.Database, txpool txspool.ITxPool) (*Dag, error) {
	mutex := new(sync.RWMutex)
	//logger := log.New("Dag")
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)
	propRep := dagcommon.NewPropRepository(propDb)
	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb)
	validate := validator.NewValidate(dagDb, utxoRep, stateDb)
	unstableChain := memunit.NewMemDag(modules.PTNCOIN, false, db, unitRep, propRep)
	tunitRep, tutxoRep, tstateRep := unstableChain.GetUnstableRepositories()
	dag := &Dag{
		//Cache:            freecache.NewCache(200 * 1024 * 1024),
		Db:               db,
		stableUnitRep:    unitRep,
		stableUtxoRep:    utxoRep,
		validate:         validate,
		ChainHeadFeed:    new(event.Feed),
		Mutex:            *mutex,
		Memdag:           unstableChain,
		unstableUnitRep:  tunitRep,
		unstableUtxoRep:  tutxoRep,
		unstableStateRep: tstateRep,
		//utxos_cache:   make(map[common.Hash]map[modules.OutPoint]*modules.Utxo),
	}
	return dag, nil
}

func (d *Dag) GetChaincodes(contractId common.Address) (*list.CCInfo, error) {
	return d.propRep.GetChaincodes(contractId)
}

func (d *Dag) SaveChaincode(contractId common.Address, cc *list.CCInfo) error {
	return d.propRep.SaveChaincode(contractId, cc)
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

//// GetCanonicalHash
//func (d *Dag) GetCanonicalHash(number uint64) (common.Hash, error) {
//	return d.unstableUnitRep.GetCanonicalHash(number)
//}
//
//// Get state
//func (d *Dag) GetHeadHeaderHash() (common.Hash, error) {
//	return d.unstableUnitRep.GetHeadHeaderHash()
//}
//
//func (d *Dag) GetHeadUnitHash() (common.Hash, error) {
//	unit := new(modules.Unit)
//	var err0 error
//	var mem_hash common.Hash
//	if d.Memdag != nil {
//		unit, err0 = d.Memdag.GetCurrentUnit(modules.NewPTNIdType(), 0)
//		if err0 != nil {
//			log.Debug("get mem current unit info", "error", err0, "hash", unit.Hash().String())
//		}
//		mem_hash = unit.Hash()
//	}
//	head_hash, err := d.unstableUnitRep.GetHeadUnitHash()
//	head_unit, _ := d.GetUnitByHash(head_hash)
//	if head_unit != nil {
//		if unit.NumberU64() > head_unit.NumberU64() {
//			return mem_hash, err
//		}
//	}
//	return head_hash, err
//}
//
//func (d *Dag) GetHeadFastUnitHash() (common.Hash, error) {
//	return d.unstableUnitRep.GetHeadFastUnitHash()
//}

func (d *Dag) GetTrieSyncProgress() (uint64, error) {
	return d.unstableUnitRep.GetTrieSyncProgress()
}

func (d *Dag) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	return d.unstableUtxoRep.GetUtxoEntry(outpoint)
}

//func (d *Dag) GetUtxoPkScripHexByTxhash(txhash common.Hash, mindex, outindex uint32) (string, error) {
//	d.Mutex.RLock()
//	defer d.Mutex.RUnlock()
//	return d.utxodb.GetUtxoPkScripHexByTxhash(txhash, mindex, outindex)
//}

func (d *Dag) GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error) {
	neededSet := make(map[modules.OutPoint]struct{})
	//preout := modules.OutPoint{TxHash: tx.Hash()}
	//var isnot_coinbase bool
	//if !dagcommon.IsCoinBase(tx) {
	//	isnot_coinbase = true
	//}

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
	// get current hash
	// assetId 暂时默认为ptn的assetId
	//unit := d.GetCurrentUnit(modules.PTNCOIN)

	//if utxos, has := d.utxos_cache[unit.Hash()]; has {
	//	if utxos != nil {
	//		for out, utxo := range utxos {
	//			view.AddUtxo(out, utxo)
	//		}
	//	}
	//}

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
	deposit, _, _ := d.stableStateRep.GetConfig("DepositRate")
	depositYearRate, _ := strconv.ParseFloat(string(deposit), 64)
	parameter.CurrentSysParameters.DepositContractInterest = depositYearRate / 365
	log.Debugf("Load SysParameter DepositContractInterest value:%f",
		parameter.CurrentSysParameters.DepositContractInterest)

	txCoinYearRateStr, _, _ := d.stableStateRep.GetConfig("TxCoinYearRate")
	txCoinYearRate, _ := strconv.ParseFloat(string(txCoinYearRateStr), 64)
	parameter.CurrentSysParameters.TxCoinDayInterest = txCoinYearRate / 365
	log.Debugf("Load SysParameter TxCoinDayInterest value:%f", parameter.CurrentSysParameters.TxCoinDayInterest)

	generateUnitRewardStr, _, _ := d.stableStateRep.GetConfig("GenerateUnitReward")
	generateUnitReward, _ := strconv.ParseUint(string(generateUnitRewardStr), 10, 64)
	parameter.CurrentSysParameters.GenerateUnitReward = generateUnitReward
	log.Debugf("Load SysParameter GenerateUnitReward value:%d", parameter.CurrentSysParameters.GenerateUnitReward)

}

//func (d *Dag) SaveUtxoView(view *txspool.UtxoViewpoint) error {
//
//	return d.unstableUtxoRep.SaveUtxoView(view.Entries())
//}

func (d *Dag) GetAddrTransactions(addr common.Address) ([]*modules.TransactionWithUnitInfo, error) {
	return d.unstableUnitRep.GetAddrTransactions(addr)
}

// get contract state
func (d *Dag) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return d.unstableStateRep.GetContractState(id, field)
	//return d.statedb.GetContractState(common.HexToAddress(id), field)
}

func (d *Dag) GetConfig(name string) ([]byte, *modules.StateVersion, error) {
	return d.unstableStateRep.GetConfig(name)
}

//get contract all state
func (d *Dag) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	return d.unstableStateRep.GetContractStatesById(id)
}

func (d *Dag) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return d.unstableStateRep.GetContractStatesByPrefix(id, prefix)
}

func (d *Dag) CreateUnit(mAddr *common.Address, txpool txspool.ITxPool, t time.Time) (*modules.Unit, error) {
	return d.unstableUnitRep.CreateUnit(mAddr, txpool, t)
}

func (d *Dag) saveHeader(header *modules.Header) error {
	unit := &modules.Unit{UnitHeader: header}
	asset := header.Number.AssetID
	var memdag memunit.IMemDag
	gasToken := dagconfig.DagConfig.GetGasToken()
	if asset == gasToken {
		memdag = d.Memdag
	} else {
		memdag = d.PartitionMemDag[asset]
	}
	if err := memdag.AddUnit(unit, nil); err != nil {
		return fmt.Errorf("Save MemDag, occurred error: %s", err.Error())
	} else {
		log.Debug("=============    save_memdag_unit header     =================", "save_memdag_unit_hex", unit.Hash().String(), "index", unit.UnitHeader.Index())
	}
	return nil
}

func (d *Dag) SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool) error {
	// todo 应当根据新的unit判断哪条链作为主链
	// step1. check exists

	if !isGenesis {
		if d.IsHeaderExist(unit.Hash()) {
			log.Debug("dag:the unit is already exist in leveldb. ", "unit_hash", unit.Hash().String())
			return errors.ErrUnitExist
		}
		// step2. validate unit
		err := d.validateUnit(unit)
		if err != nil {
			return fmt.Errorf("SaveDag, validate unit error, err=%s", err.Error())
		}
	}

	//	// step3.1. pass and with group signature, put into leveldb
	//	// todo 应当先判断是否切换，再保存，并更新状态
	//	if err := d.unstableUnitRep.SaveUnit(unit, txpool, false, false); err != nil {
	//		log.Debug("Dag", "SaveDag, save error when save unit to db err:", err)
	//		return fmt.Errorf("SaveDag, save error when save unit to db: %s", err.Error())
	//	}
	//	// step3.2. if pass and with group signature, prune fork data
	//	// if err := d.Memdag.Prune(unit.UnitHeader.Number.AssetID.String(), unit.Hash()); err != nil {
	//	// 	return fmt.Errorf("SaveDag, save error when prune: %s", err.Error())
	//	// }
	//} else {
	// step4. pass but without group signature, put into memory( if the main fork longer than 15, should call prune)
	if isGenesis {
		d.stableUnitRep.SaveUnit(unit, true)
		return nil
	}

	if err := d.Memdag.AddUnit(unit, txpool); err != nil {
		return fmt.Errorf("Save MemDag, occurred error: %s", err.Error())
	} else {
		log.Debug("=============    save_memdag_unit     =================", "save_memdag_unit_hex", unit.Hash().String(), "index", unit.UnitHeader.Index())
		//d.updateLastIrreversibleUnitNum(unit.Hash(), uint64(unit.NumberU64()))
	}

	//// todo 应当先判断是否切换，再保存，并更新状态
	//// step5. check if it is need to switch
	//// if err := d.Memdag.SwitchMainChain(); err != nil {
	//// 	return fmt.Errorf("SaveDag, save error when switch chain: %s", err.Error())
	//// }
	//// TODO
	//// update  utxo
	//go func(unit *modules.Unit) {
	//	view := txspool.NewUtxoViewpoint()
	//	if unitState == modules.UNIT_STATE_VALIDATED {
	//		view.FetchUnitUtxos(d.unstableUtxoRep, unit)
	//		// update leveldb
	//		if view != nil {
	//			needSet := make(map[modules.OutPoint]struct{})
	//			for key := range view.Entries() {
	//				needSet[key] = struct{}{}
	//			}
	//
	//			if err := view.SpentUtxo(d.unstableUtxoRep, needSet); err != nil {
	//				log.Error("update utxo failed", "error", err)
	//				// TODO
	//				// 回滚 view utxo  ，回滚world_state
	//			}
	//		}
	//		// fetch output utxo, and save
	//		//view.FetchOutputUtxos(db, unit)
	//		view2 := d.GetUtxosOutViewbyUnit(unit)
	//		for key, utxo := range view2.Entries() {
	//			if err := d.unstableUtxoRep.SaveUtxoEntity(&key, utxo); err != nil {
	//				log.Error("update output utxo failed", "error", err)
	//				// TODO
	//				// add  d.cache
	//			}
	//		}
	//
	//	} else {
	//		// get input utxos
	//		view.FetchUnitUtxos(d.unstableUtxoRep, unit)
	//		// update  cache
	//		utxos := make(map[modules.OutPoint]*modules.Utxo)
	//		var exist bool
	//		if view != nil {
	//			if utxos, exist = d.utxos_cache[parent_hash]; exist {
	//				for key, utxo := range view.Entries() {
	//					if d.utxos_cache != nil {
	//
	//						if old, has := utxos[key]; has {
	//							old.Spend()
	//							utxos[key] = old
	//							//delete(utxos, key)
	//						} else {
	//							utxo.Spend()
	//							utxos[key] = utxo
	//						}
	//					}
	//				}
	//				d.utxos_cache[parent_hash] = utxos
	//			} else {
	//				// 获取当前最新区块的utxo列表
	//				// TODO
	//				curUnit, _ := d.Memdag.GetCurrentUnit(unit.UnitHeader.Number.AssetID, unit.UnitHeader.Index()-1)
	//				utxos, _ = d.utxos_cache[curUnit.Hash()]
	//				for key, utxo := range view.Entries() {
	//					if old, has := utxos[key]; has {
	//						old.Spend()
	//						utxos[key] = old
	//						//delete(utxos, key)
	//					} else {
	//						utxo.Spend()
	//						utxos[key] = utxo
	//					}
	//					d.utxos_cache[curUnit.Hash()] = utxos
	//				}
	//			}
	//		}
	//		// get output utxos
	//		view2 := d.GetUtxosOutViewbyUnit(unit)
	//		// add d.utxo_cache
	//
	//		for key, utxo := range view2.Entries() {
	//			if utxos == nil {
	//				fmt.Println("init utxos:")
	//				utxos = make(map[modules.OutPoint]*modules.Utxo)
	//			}
	//			utxos[key] = utxo
	//		}
	//		//d.utxos_cache[unit.Hash()] = utxos
	//		//log.Info("=================saved Memdag and dag's utxo cache:  key-value ===============", "keyinfo", outpoint.String(), "utxoinfo", d.utxos_cache[unit.Hash()][outpoint])
	//	}
	//}(unit)

	return nil
}

// ValidateUnitGroupSig
//func (d *Dag) ValidateUnitGroupSig(hash common.Hash) (bool, error) {
//	unit, err := d.GetUnitByHash(hash)
//	if err != nil {
//		return false, err
//	}
//
//	//unitState := d.validate.ValidateUnitExceptGroupSig(unit, dagcommon.IsGenesis(hash))
//	unitState := d.validate.ValidateUnitExceptGroupSig(unit, d.unstableUnitRep.IsGenesis(hash))
//	if unitState != modules.UNIT_STATE_VALIDATED && unitState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
//		return false, fmt.Errorf("validate unit's groupSig failed, statecode:%d", unitState)
//	}
//	return true, nil
//}

func (d *Dag) CreateUnitForTest(txs modules.Transactions) (*modules.Unit, error) {
	// get current unit
	token := modules.PTNCOIN
	currentUnit := d.CurrentUnit(token)
	if currentUnit == nil {
		return nil, fmt.Errorf("CreateUnitForTest ERROR: genesis unit is null")
	}
	// compute height
	height := &modules.ChainIndex{
		AssetID: currentUnit.UnitHeader.Number.AssetID,
		//IsMain:  currentUnit.UnitHeader.Number.IsMain,
		Index: currentUnit.UnitHeader.Number.Index + 1,
	}
	//
	unitHeader := modules.Header{
		ParentsHash: []common.Hash{currentUnit.UnitHash},
		//Authors:      nil,
		GroupSign:   make([]byte, 0),
		GroupPubKey: make([]byte, 0),
		Number:      height,
		Time:        time.Now().Unix(),
	}

	sAddr := "P1NsG3kiKJc87M6Di6YriqHxqfPhdvxVj2B"
	addr, err := common.StringToAddress(sAddr)
	if err != nil {

	}
	bAsset, _, _ := d.unstableStateRep.GetConfig("GenesisAsset")
	if len(bAsset) <= 0 {
		return nil, fmt.Errorf("Create unit error: query asset info empty")
	}
	var asset modules.Asset
	if err := rlp.DecodeBytes(bAsset, &asset); err != nil {
		return nil, fmt.Errorf("Create unit: %s", err.Error())
	}
	ad := &modules.Addition{
		Addr:   addr,
		Asset:  asset,
		Amount: 0,
	}
	ads := make([]*modules.Addition, 0)
	ads = append(ads, ad)
	coinbase, _, err := dagcommon.CreateCoinbase(ads, time.Now())
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	newTxs := modules.Transactions{coinbase}
	if len(txs) > 0 {
		for _, tx := range txs {
			txs = append(txs, tx)
		}
	}

	unit := modules.Unit{
		UnitHeader: &unitHeader,
		Txs:        newTxs,
	}
	unit.UnitHash = unit.Hash()
	unit.UnitSize = unit.Size()
	return &unit, nil
}
func (d *Dag) GetGenesisUnit() (*modules.Unit, error) {
	return d.stableUnitRep.GetGenesisUnit()
}
func (d *Dag) GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string, tplVersion string) {
	return d.unstableStateRep.GetContractTpl(templateID)
}

func (d *Dag) GetCurrentUnitIndex(token modules.AssetId) (*modules.ChainIndex, error) {
	currentUnit := d.CurrentUnit(token)
	//	return d.GetUnitNumber(currentUnitHash)
	return currentUnit.Number(), nil
}

//func UtxoFilter(utxos map[modules.OutPoint]*modules.Utxo, assetId modules.AssetId) []*modules.Utxo {
//	res := make([]*modules.Utxo, 0)
//	for _, utxo := range utxos {
//		if utxo.Asset.AssetId == assetId {
//			res = append(res, utxo)
//		}
//	}
//	return res
//}

// dag's common geter
func (d *Dag) GetCommon(key []byte) ([]byte, error) {
	return d.unstableUnitRep.GetCommon(key)
}

// GetCommonByPrefix  return the prefix's all key && value.
func (d *Dag) GetCommonByPrefix(prefix []byte) map[string][]byte {
	return d.unstableUnitRep.GetCommonByPrefix(prefix)
}

//func (d *Dag) GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error) {
//	return d.unstableStateRep.GetCurrentChainIndex(assetId)
//}

//
//func (d *Dag) SaveChainIndex(index *modules.ChainIndex) error {
//	return d.unstableStateRep.SaveChainIndex(index)
//}

func (d *Dag) SetUnitGroupSign(unitHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error {
	if groupSign == nil {
		err := fmt.Errorf("group sign is null")
		log.Debug(err.Error())
		return err
	}

	// 验证群签名：
	err := d.VerifyUnitGroupSign(unitHash, groupSign)
	if err != nil {
		return err
	}

	// 群签之后， 更新memdag，将该unit和它的父单元们稳定存储。
	//go d.Memdag.SetStableUnit(unitHash, groupSign[:], txpool)
	d.Memdag.SetUnitGroupSign(unitHash, nil, groupSign, txpool)
	log.Debugf("Update unit[%s] group sign", unitHash.String())
	//TODO Group pub key????
	// 将缓存池utxo更新到utxodb中
	//go d.UpdateUtxosByUnit(unitHash)
	//// 更新utxo缓存池
	//go d.RefreshCacheUtxos()

	// 状态更新
	//go d.updateGlobalPropDependGroupSign(unitHash)

	return nil
}

//func (d *Dag) RefreshCacheUtxos() error {
//	timeout := time.NewTimer(time.Microsecond * 500)
//	var err error
//	for {
//		select {
//		case hash := <-d.Memdag.GetDelhashs():
//			// delete hash
//			log.Debug("want to delete hash :", "hash", hash.String())
//			delete(d.utxos_cache, hash)
//
//		case <-timeout.C:
//			err = errors.New("read hash time out.")
//			goto ENDLINE
//		}
//	}
//ENDLINE:
//	return err
//}
//
//func (d *Dag) UpdateUtxosByUnit(hash common.Hash) error {
//	d.Mutex.Lock()
//	defer d.Mutex.Unlock()
//	utxos, has := d.utxos_cache[hash]
//	if !has {
//		return errors.New("the hash is not exist in utxoscache.")
//	}
//	return d.unstableUtxoRep.SaveUtxoView(utxos)
//}
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

// SaveReqIdByTx
//func (d *Dag) SaveReqIdByTx(tx *modules.Transaction) error {
//	return d.unstableUnitRep.SaveReqIdByTx(tx)
//}

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
	log.Debug("===InsertLightHeader===", "numbers:", len(headers))
	for _, header := range headers {
		log.Debug("===InsertLightHeader===", "header index:", header.Index())
	}
	return d.InsertHeaderDag(headers)
}

//All leaf nodes for dag downloader.
//MUST have Priority.
//根据资产Id返回所有链的header。
func (d *Dag) GetAllLeafNodes() ([]*modules.Header, error) {
	// step1: get all AssetId

	return []*modules.Header{}, nil
}

func (d *Dag) UpdateSysParams() error {
	version := &modules.StateVersion{}
	//Height: &modules.ChainIndex{Index: 123, IsMain: true}, TxIndex: 1
	unit := d.GetMainCurrentUnit()
	version.Height = unit.UnitHeader.Number
	version.TxIndex = 0
	return d.stableStateRep.UpdateSysParams(version)
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
	// post event logs for further processing
	//if logs != nil {
	//	bc.logsFeed.Send(logs)
	//}
	for _, event := range events {
		switch ev := event.(type) {
		case modules.ChainEvent:
			log.Debug("======PostChainEvents======", "ev", ev)
			bc.chainFeed.Send(ev)

		case modules.ChainHeadEvent:
			log.Debug("======PostChainHeadEvent======", "ev", ev)
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
func (d *Dag) GetCoinYearRate() float64 {
	data, _, err := d.GetConfig("TxCoinYearRate")
	if err != nil {
		log.Warn("Cannot read system config by key :TxCoinYearRate")
		return 0
	}
	rate, _ := strconv.ParseFloat(string(data), 64)
	return rate
}

// SubscribeChainSideEvent registers a subscription of ChainSideEvent.
//func (bc *Dag) SubscribeChainSideEvent(ch chan<- ChainSideEvent) event.Subscription {
//	return bc.scope.Track(bc.chainSideFeed.Subscribe(ch))
//}

func (d *Dag) GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	return d.stableUnitRep.GetTxRequesterAddress(tx)
}
