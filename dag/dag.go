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
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/memunit"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/tokenengine"
	"strings"
)

type Dag struct {
	Cache         *freecache.Cache
	Db            ptndb.Database
	currentUnit   atomic.Value
	unitRep       dagcommon.IUnitRepository
	dagdb         storage.IDagDb
	utxodb        storage.IUtxoDb
	statedb       storage.IStateDb
	propdb        storage.IPropertyDb
	utxoRep       dagcommon.IUtxoRepository
	propRep       dagcommon.IPropRepository
	stateRep      dagcommon.IStateRepository
	validate      dagcommon.Validator
	ChainHeadFeed *event.Feed
	// GenesisUnit   *Unit  // comment by Albert·Gou
	Mutex  sync.RWMutex
	logger log.ILogger
	Memdag memunit.IMemDag // memory unit
	// memutxo
	// 按unit单元划分存储Utxo
	utxos_cache map[common.Hash]map[modules.OutPoint]*modules.Utxo
	// utxos_cache1 sync.Map

	// append by albert·gou 用于活跃mediator更新时的事件订阅
	activeMediatorsUpdatedFeed  event.Feed
	activeMediatorsUpdatedScope event.SubscriptionScope
}

type MemUtxos map[modules.OutPoint]*modules.Utxo

func (d *Dag) IsEmpty() bool {
	it := d.Db.NewIterator()
	return !it.Next()
}

func (d *Dag) CurrentUnit() *modules.Unit {
	// step1. get current unit hash
	hash, err := d.GetHeadUnitHash()
	//fmt.Println("d.GetHeadUnitHash()=", hash)
	if err != nil {
		log.Error("CurrentUnit when GetHeadUnitHash()", "error", err.Error())
		return nil
	}
	// step2. get unit height
	height, err := d.GetUnitNumber(hash)
	// get unit header
	uHeader, err := d.dagdb.GetHeader(hash, height)
	if err != nil {
		log.Error("Current unit when get unit header", "error", err.Error())
		return nil
	}
	// get unit hash
	uHash := common.Hash{}
	uHash.SetBytes(hash.Bytes())
	// get transaction list
	txs, err := d.dagdb.GetUnitTransactions(uHash)
	if err != nil {
		log.Error("Current unit when get transactions", "error", err.Error())
		return nil
	}
	// generate unit
	unit := modules.Unit{
		UnitHeader: uHeader,
		UnitHash:   uHash,
		Txs:        txs,
	}
	unit.UnitSize = unit.Size()
	return &unit
}

func (d *Dag) GetCurrentUnit(assetId modules.IDType16) *modules.Unit {
	memUnit := d.GetCurrentMemUnit(assetId, 0)
	curUnit := d.CurrentUnit()

	if memUnit == nil {
		return curUnit
	}
	if curUnit.NumberU64() >= memUnit.NumberU64() {
		return curUnit
	}
	return memUnit
}

func (d *Dag) GetCurrentMemUnit(assetId modules.IDType16, index uint64) *modules.Unit {
	curUnit, err := d.Memdag.GetCurrentUnit(assetId, index)
	if err != nil {
		log.Info("GetCurrentMemUnit", "error", err.Error())
		return nil
	}
	return curUnit
}

func (d *Dag) HasUnit(hash common.Hash) bool {
	u, err := d.dagdb.GetUnit(hash)
	if err != nil {
		return false
	}
	return u != nil
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

// GetMemUnitbyHash: get unit from memdag
func (d *Dag) GetMemUnitbyHash(hash common.Hash) (*modules.Unit, error) {

	unit, err := d.Memdag.GetUnit(hash)
	return unit, err
}

// GetUnitByHash: get unit from dagdb
func (d *Dag) GetUnitByHash(hash common.Hash) (*modules.Unit, error) {
	return d.dagdb.GetUnit(hash)
}

func (d *Dag) GetUnitByNumber(number modules.ChainIndex) (*modules.Unit, error) {
	//return d.dagdb.GetUnitFormIndex(number)
	hash, err := d.dagdb.GetHashByNumber(number)
	if err != nil {
		log.Debug("GetUnitByNumber dagdb.GetHashByNumber err:", "error", err)
		return nil, err
	}
	//log.Debug("Dag", "GetUnitByNumber GetUnit(hash):", hash)
	return d.dagdb.GetUnit(hash)
}

func (d *Dag) GetHeaderByHash(hash common.Hash) *modules.Header {
	height, err := d.GetUnitNumber(hash)
	if err != nil {
		log.Debug("GetHeaderByHash when GetUnitNumber", "error", err.Error())
		return nil
	}
	// get unit header
	uHeader, err := d.dagdb.GetHeader(hash, height)
	if err != nil {
		log.Debug("Current unit when get unit header", "error", err.Error())
		return nil
	}
	return uHeader
}

func (d *Dag) GetHeaderByNumber(number modules.ChainIndex) *modules.Header {
	//log.Debug("Dag", "GetHeaderByNumber ChainIndex:", number)
	hash, err := d.dagdb.GetHashByNumber(number)
	if err != nil {
		log.Debug("Dag", "GetHeaderByNumber dagdb.GetHashByNumber err:", err)
		return nil
	}

	uHeader, err1 := d.dagdb.GetHeader(hash, &number)
	if err1 != nil {
		log.Info("GetUnit when GetHeader failed ", "error:", err1, "hash", hash.String())
		//log.Info("index info:", "height", number, "index", number.Index, "asset", number.AssetID, "ismain", number.IsMain)
		return nil
	}
	return uHeader
}

func (d *Dag) GetPrefix(prefix string) map[string][]byte {
	return d.dagdb.GetPrefix(*(*[]byte)(unsafe.Pointer(&prefix)))
}

func (d *Dag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return d.ChainHeadFeed.Subscribe(ch)
}

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
	//TODO must recover，不连续的孤儿unit也应当存起来，以方便后面处理
	defer func(start time.Time) {
		log.Debug("Dag InsertDag", "elapsed", time.Since(start), "size:", len(units))
	}(time.Now())
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

		// append by albert·gou, 利用 unit 更新相关状态
		time := time.Unix(u.Timestamp(), 0)
		log.Info(fmt.Sprint("Received unit "+u.UnitHash.TerminalString()+" #", u.NumberU64(),
			" @ ", time.Format("2006-01-02 15:04:05"), " signed by ", u.UnitAuthor().Str()))
		d.ApplyUnit(u)

		// todo 应当和本地生产的unit统一接口，而不是直接存储
		// modified by albert·gou
		//if err := d.unitRep.SaveUnit(u, false); err != nil {
		if err := d.SaveUnit(u, txpool, false); err != nil {
			fmt.Errorf("Insert dag, save error: %s", err.Error())
			return count, err
		}
		log.Debug("Dag", "InsertDag ok index:", u.UnitHeader.Number.Index, "hash:", u.Hash())
		count += 1
	}

	return count, nil
}

// GetBlockHashesFromHash retrieves a number of block hashes starting at a given
// hash, fetching towards the genesis block.
func (d *Dag) GetUnitHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	header := d.GetHeaderByHash(hash)
	if header == nil {
		return nil
	}
	// Iterate the headers until enough is collected or the genesis reached
	chain := make([]common.Hash, 0, max)
	for i := uint64(0); i < max; i++ {
		if header.Index() == 0 {
			break
		}
		next := header.ParentsHash[0]
		h, err := d.GetHeader(next, header.Index()-1)
		if err != nil {
			break
		}
		header = h
		chain = append(chain, next)
	}
	return chain
}

// need add:   assetId modules.IDType16, onMain bool
func (d *Dag) HasHeader(hash common.Hash, number uint64) bool {
	return d.GetHeaderByHash(hash) != nil
}
func (d *Dag) Exists(hash common.Hash) bool {
	if unit, err := d.dagdb.GetUnit(hash); err == nil && unit != nil {
		log.Info("hash is exsit in leveldb ", "index:", unit.Header().Number.Index, "hash", hash.String())
		return true
	}
	return false
}
func (d *Dag) CurrentHeader() *modules.Header {
	unit := d.CurrentUnit()
	if unit != nil {
		return unit.Header()
	}
	return nil
}

// GetBodyRLP retrieves a block body in RLP encoding from the database by hash,
// caching it if found.
func (d *Dag) GetBodyRLP(hash common.Hash) rlp.RawValue {
	return d.getBodyRLP(hash)
}

// GetUnitTransactions is return unit's body, all transactions of unit.
func (d *Dag) GetUnitTransactions(hash common.Hash) (modules.Transactions, error) {
	return d.dagdb.GetUnitTransactions(hash)
}

// GetUnitTxsHash is return the unit's txs hash list.
func (d *Dag) GetUnitTxsHash(hash common.Hash) ([]common.Hash, error) {
	return d.dagdb.GetBody(hash)
}

// GetTransactionByHash is return the tx by tx's hash
func (d *Dag) GetTransactionByHash(hash common.Hash) (*modules.Transaction, common.Hash, error) {
	tx, uhash, _, _ := d.dagdb.GetTransaction(hash)
	if tx == nil {
		return nil, uhash, errors.New("get transaction by hash is failed,none the transaction.")
	}
	return tx, uhash, nil
}
func (d *Dag) GetTxSearchEntry(hash common.Hash) (*modules.TxLookupEntry, error) {
	unitHash, unitNumber, txIndex, err := d.dagdb.GetTxLookupEntry(hash)
	return &modules.TxLookupEntry{
		UnitHash:  unitHash,
		UnitIndex: unitNumber,
		Index:     txIndex,
	}, err
}

func (d *Dag) getBodyRLP(hash common.Hash) rlp.RawValue {
	txs := modules.Transactions{}
	// get hash list
	txs, err := d.dagdb.GetUnitTransactions(hash)
	if err != nil {
		log.Error("Get body rlp", "unit hash", hash.String(), "error", err.Error())
		return nil
	}

	data, err := rlp.EncodeToBytes(txs)
	if err != nil {
		log.Error("Get body rlp when rlp encode", "unit hash", hash.String(), "error", err.Error())
		return nil
	}
	// get hash data
	return data
}

func (d *Dag) GetHeaderRLP(db storage.DatabaseReader, hash common.Hash) rlp.RawValue {
	number, err := d.dagdb.GetNumberWithUnitHash(hash)
	if err != nil {
		log.Error("Get header rlp ", "error", err.Error())
		return nil
	}
	return d.dagdb.GetHeaderRlp(hash, number.Index)
}

// InsertHeaderDag attempts to insert the given header chain in to the local
// chain, possibly creating a reorg. If an error is returned, it will return the
// index number of the failing header as well an error describing what went wrong.
//
// The verify parameter can be used to fine tune whether nonce verification
// should be done or not. The reason behind the optional check is because some
// of the header retrieval mechanisms already need to verify nonces, as well as
// because nonces can be verified sparsely, not needing to check each.
func (d *Dag) InsertHeaderDag(headers []*modules.Header, checkFreq int) (int, error) {
	for i, header := range headers {
		hash := header.Hash()
		number := header.Number
		index := header.Number.Index

		// ###save unit hash and chain index relation
		err := d.dagdb.SaveNumberByHash(hash, number)
		if err != nil {
			return i, fmt.Errorf("InsertHeaderDag, on header:%d, at SaveNumberByHash Error", i)
		}
		err = d.dagdb.SaveHashByNumber(hash, number)
		if err != nil {
			return i, fmt.Errorf("InsertHeaderDag, on header:%d, at SaveHashByNumber Error", i)
		}
		// ###save HeaderCanon & HeaderKey & HeadUnitKey & HeadFastKey
		err = d.dagdb.UpdateHeadByBatch(hash, index)
		if err != nil {
			return i, err
		}

	}
	return checkFreq, nil
}

//VerifyHeader checks whether a header conforms to the consensus rules of the stock
//Ethereum ethash engine.go
func (d *Dag) VerifyHeader(header *modules.Header, seal bool) error {
	// step1. check unit signature, should be compare to mediator list
	unitState := d.validate.ValidateUnitSignature(header, false)
	if unitState != modules.UNIT_STATE_VALIDATED && unitState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
		return fmt.Errorf("Validate unit signature error, errno=%d", unitState)
	}

	// step2. check extra data
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > uint64(32) {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), configure.MaximumExtraDataSize)
	}

	return nil
}

//All leaf nodes for dag downloader.
//MUST have Priority.
func (d *Dag) GetAllLeafNodes() ([]*modules.Header, error) {
	return d.dagdb.GetAllLeafNodes()
}

/**
获取account address下面的token信息
To get account token list and tokens's information
*/
func (d *Dag) WalletTokens(addr common.Address) (map[string]*modules.AccountToken, error) {
	return d.utxoRep.GetAccountTokens(addr)
}

func (d *Dag) WalletBalance(address common.Address, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
	newAssetid := modules.IDType16{}
	newUnitqueid := modules.IDType16{}

	if len(assetid) != cap(newAssetid) {
		return 0, fmt.Errorf("Assetid lenth is wrong")
	}
	if len(uniqueid) != cap(newUnitqueid) {
		return 0, fmt.Errorf("Uniqueid lenth is wrong")
	}
	if chainid == 0 {
		return 0, fmt.Errorf("Chainid is invalid")
	}

	newAssetid.SetBytes(assetid)
	newUnitqueid.SetBytes(uniqueid)

	asset := modules.Asset{
		AssetId:  newAssetid,
		UniqueId: newUnitqueid,
	}

	return d.utxoRep.WalletBalance(address, asset), nil
}

// Utxos : return mem utxos
func (d *Dag) Utxos() map[common.Hash]map[modules.OutPoint]*modules.Utxo {
	// result := d.utxos_cache1
	// utxos := make(Utxos, 0)
	// d.utxos_cache1.Range(func(key, v interface{}) bool {
	// 	utxos[key] = v
	// })
	return d.utxos_cache
}

func NewDag(db ptndb.Database, l log.ILogger) (*Dag, error) {
	mutex := new(sync.RWMutex)

	dagDb := storage.NewDagDb(db, l)
	utxoDb := storage.NewUtxoDb(db, l)
	stateDb := storage.NewStateDb(db, l)
	idxDb := storage.NewIndexDb(db, l)
	propDb := storage.NewPropertyDb(db, l)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, l)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, l)
	validate := dagcommon.NewValidate(dagDb, utxoDb, utxoRep, stateDb, l)
	propRep := dagcommon.NewPropRepository(propDb, l)
	stateRep := dagcommon.NewStateRepository(stateDb, l)
	dag := &Dag{
		Cache:         freecache.NewCache(200 * 1024 * 1024),
		Db:            db,
		unitRep:       unitRep,
		dagdb:         dagDb,
		utxodb:        utxoDb,
		statedb:       stateDb,
		propdb:        propDb,
		utxoRep:       utxoRep,
		propRep:       propRep,
		stateRep:      stateRep,
		validate:      validate,
		ChainHeadFeed: new(event.Feed),
		Mutex:         *mutex,
		Memdag:        memunit.NewMemDag(dagDb, stateDb, unitRep),
		utxos_cache:   make(map[common.Hash]map[modules.OutPoint]*modules.Utxo),
	}
	return dag, nil
}

func NewDag4GenesisInit(db ptndb.Database) (*Dag, error) {
	mutex := new(sync.RWMutex)
	logger := log.New("Dag")
	dagDb := storage.NewDagDb(db, logger)
	utxoDb := storage.NewUtxoDb(db, logger)
	stateDb := storage.NewStateDb(db, logger)
	idxDb := storage.NewIndexDb(db, logger)
	propDb := storage.NewPropertyDb(db, logger)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, logger)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, logger)
	validate := dagcommon.NewValidate(dagDb, utxoDb, utxoRep, stateDb, logger)
	propRep := dagcommon.NewPropRepository(propDb, logger)

	dag := &Dag{
		Cache:         freecache.NewCache(200 * 1024 * 1024),
		Db:            db,
		unitRep:       unitRep,
		dagdb:         dagDb,
		utxodb:        utxoDb,
		statedb:       stateDb,
		propdb:        propDb,
		utxoRep:       utxoRep,
		propRep:       propRep,
		validate:      validate,
		ChainHeadFeed: new(event.Feed),
		Mutex:         *mutex,
		//Memdag:        memunit.NewMemDag(dagDb, stateDb, unitRep),
		utxos_cache: make(map[common.Hash]map[modules.OutPoint]*modules.Utxo),
	}

	return dag, nil
}

func NewDagForTest(db ptndb.Database, txpool txspool.ITxPool) (*Dag, error) {
	mutex := new(sync.RWMutex)
	logger := log.New("Dag")
	dagDb := storage.NewDagDb(db, logger)
	utxoDb := storage.NewUtxoDb(db, logger)
	stateDb := storage.NewStateDb(db, logger)
	idxDb := storage.NewIndexDb(db, logger)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, logger)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, logger)
	validate := dagcommon.NewValidate(dagDb, utxoDb, utxoRep, stateDb, logger)

	dag := &Dag{
		Cache:         freecache.NewCache(200 * 1024 * 1024),
		Db:            db,
		unitRep:       unitRep,
		dagdb:         dagDb,
		utxodb:        utxoDb,
		statedb:       stateDb,
		utxoRep:       utxoRep,
		validate:      validate,
		ChainHeadFeed: new(event.Feed),
		Mutex:         *mutex,
		Memdag:        memunit.NewMemDagForTest(dagDb, stateDb, unitRep, txpool),
		utxos_cache:   make(map[common.Hash]map[modules.OutPoint]*modules.Utxo),
	}
	return dag, nil
}

// Get Contract Api
func (d *Dag) GetContract(id []byte) (*modules.Contract, error) {
	return d.statedb.GetContract(id)
}

// Get Header
func (d *Dag) GetHeader(hash common.Hash, number uint64) (*modules.Header, error) {
	index, err := d.GetUnitNumber(hash)
	if err != nil {
		return nil, err
	}
	//TODO compare index with number
	if index.Index == number {
		head, err := d.dagdb.GetHeader(hash, index)
		if err != nil {
			log.Info("get unit header faled,", "error", err)
		}
		return head, err
	}
	return nil, err
}

// Get UnitNumber
func (d *Dag) GetUnitNumber(hash common.Hash) (*modules.ChainIndex, error) {
	return d.dagdb.GetNumberWithUnitHash(hash)
}

// GetCanonicalHash
func (d *Dag) GetCanonicalHash(number uint64) (common.Hash, error) {
	return d.dagdb.GetCanonicalHash(number)
}

// Get state
func (d *Dag) GetHeadHeaderHash() (common.Hash, error) {
	return d.dagdb.GetHeadHeaderHash()
}

func (d *Dag) GetHeadUnitHash() (common.Hash, error) {
	return d.dagdb.GetHeadUnitHash()
}

func (d *Dag) GetHeadFastUnitHash() (common.Hash, error) {
	return d.dagdb.GetHeadFastUnitHash()
}

func (d *Dag) GetTrieSyncProgress() (uint64, error) {
	return d.dagdb.GetTrieSyncProgress()
}

func (d *Dag) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	return d.utxodb.GetUtxoEntry(outpoint)
}
func (d *Dag) GetUtxoPkScripHexByTxhash(txhash common.Hash, mindex, outindex uint32) (string, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	return d.utxodb.GetUtxoPkScripHexByTxhash(txhash, mindex, outindex)
}
func (d *Dag) GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error) {
	neededSet := make(map[modules.OutPoint]struct{})
	//preout := modules.OutPoint{TxHash: tx.Hash()}
	var isnot_coinbase bool
	if !dagcommon.IsCoinBase(tx) {
		isnot_coinbase = true
	}

	for _, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				//msgIdx := uint32(i)
				//preout.MessageIndex = msgIdx
				//for j := range msg.Output {
				//	txoutIdx := uint32(j)
				//	preout.OutIndex = txoutIdx
				//	neededSet[preout] = struct{}{}
				//}
				// if tx is Not CoinBase
				// add txIn previousoutpoint
				if isnot_coinbase {
					for _, in := range msg.Inputs {
						neededSet[*in.PreviousOutPoint] = struct{}{}
					}
				}
			}
		}
	}

	view := txspool.NewUtxoViewpoint()
	d.Mutex.RLock()
	err := view.FetchUtxos(d.utxodb, neededSet)
	// get current hash
	// assetId 暂时默认为ptn的assetId
	unit := d.GetCurrentUnit(modules.PTNCOIN)

	if utxos, has := d.utxos_cache[unit.Hash()]; has {
		if utxos != nil {
			for out, utxo := range utxos {
				view.AddUtxo(out, utxo)
			}
		}
	}

	d.Mutex.RUnlock()

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
	items, err := d.utxodb.GetAllUtxos()
	// TODO---> merge dag.cache
	// if d.utxos_cache != nil {
	// 	for key, utxo := range d.utxos_cache {
	// 		if old, has := items[key]; has {
	// 			// merge
	// 			if old.IsSpent() {
	// 				delete(items, key)
	// 			}
	// 		}
	// 		items[key] = utxo
	// 	}
	// }
	if d.utxos_cache != nil {
		for _, utxos := range d.utxos_cache {
			if utxos != nil {
				for key, utxo := range utxos {
					if old, has := items[key]; has {
						// merge
						if old.IsSpent() {
							delete(items, key)
						}
					}
					items[key] = utxo
				}
			}
		}
	}
	d.Mutex.RUnlock()

	return items, err
}

func (d *Dag) GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error) {
	// TODO
	// merge dag.cache
	all, err := d.utxodb.GetAddrOutpoints(addr)
	if d.utxos_cache != nil {
		for hash, utxos := range d.utxos_cache {
			for key, utxo := range utxos {
				if utxo == nil {
					delete(utxos, key)
					continue
				} else {
					address, err := tokenengine.GetAddressFromScript(utxo.PkScript)
					if err == nil {
						if address.Equal(addr) {
							var exist bool
							for _, old := range all {
								if reflect.DeepEqual(key.ToKey(), old.ToKey()) {
									exist = true
									break
								}
							}
							if !exist {
								all = append(all, key)
							}
						}
					}
				}
			}
			d.utxos_cache[hash] = utxos
		}
	}
	return all, err
}

func (d *Dag) GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error) {
	utxo, err := d.utxodb.GetUtxoEntry(outPoint)
	if err != nil {
		return common.Address{}, err
	}
	return tokenengine.GetAddressFromScript(utxo.PkScript)
}

func (d *Dag) GetTxFee(pay *modules.Transaction) (*modules.InvokeFees, error) {
	return d.utxoRep.ComputeTxFee(pay)
}

func (d *Dag) GetAddrOutput(addr string) ([]modules.Output, error) {
	return d.dagdb.GetAddrOutput(addr)
}

func (d *Dag) GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error) {
	//TODO only get one token's UTXO
	all, err := d.utxodb.GetAddrUtxos(addr)
	if d.utxos_cache != nil {
		for hash, utxos := range d.utxos_cache {
			for key, utxo := range utxos {
				if utxo == nil {
					log.Info("------------------the utxo is nil  ----------------", "utxokey", key.String())
					delete(utxos, key)
					continue
				} else {
					address, err := tokenengine.GetAddressFromScript(utxo.PkScript)
					if err == nil {
						if address.Equal(addr) {
							if strings.Compare(utxo.Asset.String(), asset.String()) == 0 {
								if old, has := all[key]; has {
									// merge
									if old.IsSpent() {
										log.Warn("It is delete the spent utxo that I found the old utxo amount: ", "amount", old.Amount)
										// delete(all, key)
									}
								}
								log.Info("new utxo amount :", "amount", utxo.Amount)
								all[key] = utxo
							}
						}
					}
				}
			}
			d.utxos_cache[hash] = utxos
		}
	}
	return all, err
	//return map[modules.OutPoint]*modules.Utxo{}, nil
}
func (d *Dag) GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	// TODO
	// merge dag.cache
	all, err := d.utxodb.GetAddrUtxos(addr)
	if d.utxos_cache != nil {
		for hash, utxos := range d.utxos_cache {
			for key, utxo := range utxos {
				if utxo == nil {
					log.Info("------------------the utxo is nil  ----------------", "utxokey", key.String())
					delete(utxos, key)
					continue
				} else {
					address, err := tokenengine.GetAddressFromScript(utxo.PkScript)
					if err == nil {
						if address.Equal(addr) {
							if old, has := all[key]; has {
								// merge
								if old.IsSpent() {
									log.Warn("It is delete the spent utxo that I found the old utxo amount: ", "amount", old.Amount)
									// delete(all, key)
								}
							}
							log.Info("new utxo amount :", "amount", utxo.Amount)
							all[key] = utxo
						}
					}
				}
			}
			d.utxos_cache[hash] = utxos
		}
	}
	return all, err
}

func (d *Dag) SaveUtxoView(view *txspool.UtxoViewpoint) error {

	return d.utxodb.SaveUtxoView(view.Entries())
}

func (d *Dag) GetAddrTransactions(addr string) (modules.Transactions, error) {
	return d.dagdb.GetAddrTransactions(addr)
}

// get contract state
func (d *Dag) GetContractState(id []byte, field string) (*modules.StateVersion, []byte) {
	return d.statedb.GetContractState(id, field)
	//return d.statedb.GetContractState(common.HexToAddress(id), field)
}

//get contract all state
func (d *Dag) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	return d.statedb.GetContractStatesById(id)
}

func (d *Dag) CreateUnit(mAddr *common.Address, txpool txspool.ITxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error) {
	return d.unitRep.CreateUnit(mAddr, txpool, ks, t)
}

//modified by Albert·Gou
func (d *Dag) SaveUnit4GenesisInit(unit *modules.Unit, txpool txspool.ITxPool) error {
	return d.unitRep.SaveUnit(unit, txpool, true, false)
}

func (d *Dag) SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool) error {
	// todo 应当根据新的unit判断哪条链作为主链
	// step1. check exists
	var parent_hash common.Hash
	if !isGenesis {
		parent_hash = unit.ParentHash()[0]
	} else {
		parent_hash = unit.Hash()
	}

	//log.Debug("start save dag", "index", unit.UnitHeader.Index(), "hash", unit.Hash())

	if !isGenesis {
		if d.Memdag.Exists(unit.Hash()) || d.Exists(unit.Hash()) {
			log.Info("dag:the unit is already exist in leveldb. ", "unit_hash", unit.Hash().String())
			return errors.ErrUnitExist //fmt.Errorf("SaveDag, unit(%s) is already existing.", unit.Hash().String())
		}
	}
	// step2. validate unit

	unitState := d.validate.ValidateUnitExceptGroupSig(unit, isGenesis)

	if unitState != modules.UNIT_STATE_VALIDATED && unitState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED && unitState != modules.UNIT_STATE_CHECK_HEADER_PASSED {
		return fmt.Errorf("SaveDag, validate unit error, err_no=%d", unitState)
	}

	if unitState == modules.UNIT_STATE_VALIDATED {
		// step3.1. pass and with group signature, put into leveldb
		// todo 应当先判断是否切换，再保存，并更新状态
		if err := d.unitRep.SaveUnit(unit, txpool, false, false); err != nil {
			log.Info("Dag", "SaveDag, save error when save unit to db err:", err)
			return fmt.Errorf("SaveDag, save error when save unit to db: %s", err.Error())
		}
		// step3.2. if pass and with group signature, prune fork data
		// if err := d.Memdag.Prune(unit.UnitHeader.Number.AssetID.String(), unit.Hash()); err != nil {
		// 	return fmt.Errorf("SaveDag, save error when prune: %s", err.Error())
		// }
	} else {
		// step4. pass but without group signature, put into memory( if the main fork longer than 15, should call prune)
		if err := d.Memdag.Save(unit, txpool); err != nil {
			return fmt.Errorf("Save MemDag, occurred error: %s", err.Error())
		} else {
			log.Info("=============    save_memdag_unit     =================", "save_memdag_unit_hex", unit.Hash().String(), "index", unit.UnitHeader.Index())
		}
	}

	// todo 应当先判断是否切换，再保存，并更新状态
	// step5. check if it is need to switch
	// if err := d.Memdag.SwitchMainChain(); err != nil {
	// 	return fmt.Errorf("SaveDag, save error when switch chain: %s", err.Error())
	// }
	// TODO
	// update  utxo
	go func(unit *modules.Unit) {
		view := txspool.NewUtxoViewpoint()
		if unitState == modules.UNIT_STATE_VALIDATED {
			view.FetchUnitUtxos(d.utxodb, unit)
			// update leveldb
			if view != nil {
				needSet := make(map[modules.OutPoint]struct{})
				for key, _ := range view.Entries() {
					needSet[key] = struct{}{}
				}

				if err := view.SpentUtxo(d.utxodb, needSet); err != nil {
					log.Error("update utxo failed", "error", err)
					// TODO
					// 回滚 view utxo  ，回滚world_state
				}
			}
			// fetch output utxo, and save
			//view.FetchOutputUtxos(db, unit)
			view2 := d.GetUtxosOutViewbyUnit(unit)
			for key, utxo := range view2.Entries() {
				if err := d.utxodb.SaveUtxoEntity(&key, utxo); err != nil {
					log.Error("update output utxo failed", "error", err)
					// TODO
					// add  d.cache
				}
			}

		} else {
			// get input utxos
			view.FetchUnitUtxos(d.utxodb, unit)
			// update  cache
			utxos := make(map[modules.OutPoint]*modules.Utxo)
			var exist bool
			if view != nil {
				if utxos, exist = d.utxos_cache[parent_hash]; exist {
					for key, utxo := range view.Entries() {
						if d.utxos_cache != nil {

							if old, has := utxos[key]; has {
								old.Spend()
								utxos[key] = old
								//delete(utxos, key)
							} else {
								utxo.Spend()
								utxos[key] = utxo
							}
						}
					}
					d.utxos_cache[parent_hash] = utxos
				} else {
					// 获取当前最新区块的utxo列表
					// TODO
					curUnit, _ := d.Memdag.GetCurrentUnit(unit.UnitHeader.Number.AssetID, unit.UnitHeader.Index()-1)
					utxos, _ = d.utxos_cache[curUnit.Hash()]
					for key, utxo := range view.Entries() {
						if old, has := utxos[key]; has {
							old.Spend()
							utxos[key] = old
							//delete(utxos, key)
						} else {
							utxo.Spend()
							utxos[key] = utxo
						}
						d.utxos_cache[curUnit.Hash()] = utxos
					}
				}
			}
			// get output utxos
			view2 := d.GetUtxosOutViewbyUnit(unit)
			// add d.utxo_cache

			for key, utxo := range view2.Entries() {
				if utxos == nil {
					fmt.Println("init utxos:")
					utxos = make(map[modules.OutPoint]*modules.Utxo)
				}
				utxos[key] = utxo
			}
			d.utxos_cache[unit.Hash()] = utxos
			//log.Info("=================saved Memdag and dag's utxo cache:  key-value ===============", "keyinfo", outpoint.String(), "utxoinfo", d.utxos_cache[unit.Hash()][outpoint])
		}
	}(unit)

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
//	unitState := d.validate.ValidateUnitExceptGroupSig(unit, d.unitRep.IsGenesis(hash))
//	if unitState != modules.UNIT_STATE_VALIDATED && unitState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
//		return false, fmt.Errorf("validate unit's groupSig failed, statecode:%d", unitState)
//	}
//	return true, nil
//}

//func (d *Dag) GetAccountMediatorVote(address common.Address) []common.Address {
//	// todo
//	bAddress := d.statedb.GetAccountVoteInfo(address, vote.TYPE_MEDIATOR)
//	res := []common.Address{}
//	for _, b := range bAddress {
//		res = append(res, common.BytesToAddress(b))
//	}
//	return res
//}

func (d *Dag) CreateUnitForTest(txs modules.Transactions) (*modules.Unit, error) {
	// get current unit
	currentUnit := d.CurrentUnit()
	if currentUnit == nil {
		return nil, fmt.Errorf("CreateUnitForTest ERROR: genesis unit is null")
	}
	// compute height
	height := modules.ChainIndex{
		AssetID: currentUnit.UnitHeader.Number.AssetID,
		IsMain:  currentUnit.UnitHeader.Number.IsMain,
		Index:   currentUnit.UnitHeader.Number.Index + 1,
	}
	//
	unitHeader := modules.Header{
		ParentsHash: []common.Hash{currentUnit.UnitHash},
		AssetIDs:    []modules.IDType16{currentUnit.UnitHeader.Number.AssetID},
		//Authors:      nil,
		GroupSign:    make([]byte, 0),
		GroupPubKey:  make([]byte, 0),
		Number:       height,
		Creationdate: time.Now().Unix(),
	}

	sAddr := "P1NsG3kiKJc87M6Di6YriqHxqfPhdvxVj2B"
	addr, err := common.StringToAddress(sAddr)
	if err != nil {

	}
	bAsset, _, _ := d.statedb.GetConfig([]byte("GenesisAsset"))
	if len(bAsset) <= 0 {
		return nil, fmt.Errorf("Create unit error: query asset info empty")
	}
	var asset modules.Asset
	if err := rlp.DecodeBytes(bAsset, &asset); err != nil {
		return nil, fmt.Errorf("Create unit: %s", err.Error())
	}
	coinbase, err := dagcommon.CreateCoinbase(&addr, 0, nil, &asset, time.Now())
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
func (d *Dag) GetGenesisUnit(index uint64) (*modules.Unit, error) {
	return d.unitRep.GetGenesisUnit(index)
}
func (d *Dag) GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string) {
	return d.statedb.GetContractTpl(templateID)
}

// save token info
func (d *Dag) SaveTokenInfo(token_info *modules.TokenInfo) (*modules.TokenInfo, error) { // return key's hex
	return d.dagdb.SaveTokenInfo(token_info)
}

// Get token info
func (d *Dag) GetTokenInfo(key string) (*modules.TokenInfo, error) {
	return d.dagdb.GetTokenInfo(key)
}

// Get all token info
func (d *Dag) GetAllTokenInfo() (*modules.AllTokenInfo, error) {
	return d.dagdb.GetAllTokenInfo()
}

//@Yiran
func (d *Dag) GetCurrentUnitIndex() (*modules.ChainIndex, error) {
	currentUnitHash := d.CurrentUnit().UnitHash
	return d.GetUnitNumber(currentUnitHash)
}

//@Yiran save utxo snapshot when new mediator cycle begin
// unit index MUST to be  integer multiples of  termInterval.
//func (d *Dag) SaveUtxoSnapshot() error {
//	currentUnitIndex, err := d.GetCurrentUnitIndex()
//	if err != nil {
//		return err
//	}
//	return d.utxodb.SaveUtxoSnapshot(currentUnitIndex)
//}

//@Yiran Get last utxo snapshot
// must calling after SaveUtxoSnapshot call , before this mediator cycle end.
// called by GenerateVoteResult
//func (d *Dag) GetUtxoSnapshot() (*[]modules.Utxo, error) {
//	unitIndex, err := d.GetCurrentUnitIndex()
//	if err != nil {
//		return nil, err
//	}
//	unitIndex.Index -= unitIndex.Index % modules.TERMINTERVAL
//	return d.utxodb.GetUtxoEntities(unitIndex)
//}

////@Yiran
//func (d *Dag) GenerateVoteResult() (*[]storage.AddressVote, error) {
//	AddressVoteBox := storage.NewAddressVoteBox()
//
//	utxos, err := d.utxodb.GetAllUtxos()
//	if err != nil {
//		return nil, err
//	}
//	for _, utxo := range utxos {
//		if utxo.Asset.AssetId == modules.PTNCOIN {
//			utxoHolder, err := tokenengine.GetAddressFromScript(utxo.PkScript)
//			if err != nil {
//				return nil, err
//			}
//			AddressVoteBox.AddToBoxIfNotVoted(utxoHolder, utxo.VoteResult)
//		}
//	}
//	AddressVoteBox.Sort()
//	return &AddressVoteBox.Candidates, nil
//}

func UtxoFilter(utxos map[modules.OutPoint]*modules.Utxo, assetId modules.IDType16) []*modules.Utxo {
	res := make([]*modules.Utxo, 0)
	for _, utxo := range utxos {
		if utxo.Asset.AssetId == assetId {
			res = append(res, utxo)
		}
	}
	return res
}

////@Yiran
//func (d *Dag) UpdateActiveMediators() error {
//	var TermInterval uint64 = 50
//	MediatorNumber := d.GetActiveMediatorCount()
//	// <1> Get election unit
//	hash := d.CurrentUnit().UnitHash
//	index, err := d.GetUnitNumber(hash)
//	if err != nil {
//		return err
//	}
//	if index.Index <= TermInterval {
//		return errors.New("first election must wait until first term period end")
//		//adjust TermInterval to fit the unit number
//		//TermInterval = index.Index
//	}
//	index.Index -= index.Index % TermInterval
//	d.GetUnitByNumber(index).
//
//	//// <2> Get all votes belonged to this election period
//	//voteBox := storage.AddressVoteBox{}
//	//for i := TermInterval; i > 0; i-- { // for each unit in period.
//	//	for _, Tx := range d.GetUnitByNumber(index).Txs { //for each transaction in unit
//	//		voter := Tx.TxMessages.GetInputAddress()
//	//		voteTo := Tx.TxMessages.GetVoteResult()
//	//		voteBox.AddToBoxIfNotVoted(voter, voteTo)
//	//	}
//	//}
//
//	// <3> calculate vote result
//	addresses := voteBox.Head(MediatorNumber) //sort by candidates vote number & return the addresses of the top n account
//
//	// <4> create active mediators from addresses & update globalProperty
//	activeMediators := make(map[common.Address]core.Mediator, 0)
//	for _, addr := range (addresses) {
//		newmediator := *d.GetGlobalProp().GetActiveMediator(addr)
//		activeMediators[addr] = newmediator
//	}
//
//	return nil
//}

//GetElectedMediatorsAddress YiRan@
//func (dag *Dag) GetElectedMediatorsAddress() (map[string]uint64, error) {
//	//gp, err := dag.propdb.RetrieveGlobalProp()
//	//if err != nil {
//	//	return nil, err
//	//}
//	//MediatorNumber := gp.GetActiveMediatorCount()
//	return dag.statedb.GetSortedMediatorVote(0)
//}

// UpdateMediator
//func (d *Dag) UpdateMediator() error {
//	mas, err := d.GetElectedMediatorsAddress()
//	if err != nil {
//		return err
//	}
//	fmt.Println(mas)
//	//TODO
//	return nil
//}

// dag's common geter
func (d *Dag) GetCommon(key []byte) ([]byte, error) {
	return d.dagdb.GetCommon(key)
}

// GetCommonByPrefix  return the prefix's all key && value.
func (d *Dag) GetCommonByPrefix(prefix []byte) map[string][]byte {
	return d.dagdb.GetCommonByPrefix(prefix)
}

func (d *Dag) GetCurrentChainIndex(assetId modules.IDType16) (*modules.ChainIndex, error) {
	return d.statedb.GetCurrentChainIndex(assetId)
}

func (d *Dag) SaveChainIndex(index *modules.ChainIndex) error {
	return d.statedb.SaveChainIndex(index)
}

func (d *Dag) SetUnitGroupSign(unitHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error {
	if groupSign == nil {
		return errors.New("group sign is null.")
	}
	// 验证群签名：
	err := d.VerifyUnitGroupSign(unitHash, groupSign)
	if err != nil {
		return err
	}

	// 群签之后， 更新memdag，将该unit和它的父单元们稳定存储。
	go d.Memdag.UpdateMemDag(unitHash, groupSign[:], txpool)

	// 将缓存池utxo更新到utxodb中
	go d.UpdateUtxosByUnit(unitHash)
	// 更新utxo缓存池
	go d.RefreshCacheUtxos()

	// 状态更新
	go d.updateGlobalPropDependGroupSign(unitHash)

	return nil
}

func (d *Dag) RefreshCacheUtxos() error {
	timeout := time.NewTimer(time.Microsecond * 500)
	var err error
	for {
		select {
		case hash := <-d.Memdag.GetDelhashs():
			// delete hash
			log.Debug("want to delete hash :", "hash", hash.String())
			delete(d.utxos_cache, hash)

		case <-timeout.C:
			err = errors.New("read hash time out.")
			goto ENDLINE
		}
	}
ENDLINE:
	return err
}

func (d *Dag) UpdateUtxosByUnit(hash common.Hash) error {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	utxos, has := d.utxos_cache[hash]
	if !has {
		return errors.New("the hash is not exist in utxoscache.")
	}
	return d.utxodb.SaveUtxoView(utxos)
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
