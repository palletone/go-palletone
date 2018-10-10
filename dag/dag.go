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
	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/tokenengine"
	"sync"
	"sync/atomic"

	//"github.com/ethereum/go-ethereum/params"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/memunit"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/txspool"
)

type Dag struct {
	Cache         *freecache.Cache
	currentUnit   atomic.Value
	Db            ptndb.Database
	unitRep       dagcommon.IUnitRepository
	dagdb         storage.IDagDb
	utxodb        storage.IUtxoDb
	statedb       storage.IStateDb
	propdb        storage.IPropertyDb
	utxoRep       dagcommon.IUtxoRepository
	propRep       dagcommon.IPropRepository
	validate      dagcommon.Validator
	ChainHeadFeed *event.Feed
	// GenesisUnit   *Unit  // comment by Albert·Gou
	Mutex sync.RWMutex
	logger log.ILogger
	Memdag *memunit.MemDag // memory unit
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
	uHeader, err := d.dagdb.GetHeader(hash, &height)
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
		//TODO xiaozhi
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
	//TODO xiaozhi
	return d.CurrentUnit()
}

func (d *Dag) GetCurrentMemUnit(assetId modules.IDType16) *modules.Unit {
	curUnit, err := d.Memdag.GetCurrentUnit(assetId)
	if err != nil {
		log.Error("GetCurrentMemUnit", "error", err.Error())
		return nil
	}
	return curUnit
}

func (d *Dag) GetUnit(hash common.Hash) *modules.Unit {
	return d.dagdb.GetUnit(hash)
}

func (d *Dag) HasUnit(hash common.Hash) bool {
	return d.dagdb.GetUnit(hash) != nil
}

func (d *Dag) GetUnitByHash(hash common.Hash) *modules.Unit {
	return d.dagdb.GetUnit(hash)
}

func (d *Dag) GetUnitByNumber(number modules.ChainIndex) *modules.Unit {
	return d.dagdb.GetUnitFormIndex(number)
}

func (d *Dag) GetHeaderByHash(hash common.Hash) *modules.Header {
	height, err := d.GetUnitNumber(hash)
	if err != nil {
		log.Error("GetHeaderByHash when GetUnitNumber", "error", err.Error())
	}
	// get unit header
	uHeader, err := d.dagdb.GetHeader(hash, &height)
	if err != nil {
		log.Error("Current unit when get unit header", "error", err.Error())
		return nil
	}
	return uHeader
}

func (d *Dag) GetHeaderByNumber(number modules.ChainIndex) *modules.Header {
	header, _ := d.dagdb.GetHeaderByHeight(number)
	return header
}

func (d *Dag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return d.ChainHeadFeed.Subscribe(ch)
}

// FastSyncCommitHead sets the current head block to the one defined by the hash
// irrelevant what the chain contents were prior.
func (d *Dag) FastSyncCommitHead(hash common.Hash) error {
	unit := d.GetUnit(hash)
	if unit == nil {
		return fmt.Errorf("non existent unit [%x...]", hash[:4])
	}
	// store current unit
	d.Mutex.Lock()
	d.currentUnit.Store(unit)
	d.Mutex.Unlock()

	return nil
}

func (d *Dag) SaveDag(unit modules.Unit, isGenesis bool) (int, error) {
	// step1. check exists
	if d.Memdag.Exists(unit.UnitHash) || d.Exists(unit.UnitHash) {
		return -2, fmt.Errorf("SaveDag, unit(%s) is already existing.", unit.UnitHash.String())
	}
	// step2. validate unit
	unitState := d.validate.ValidateUnitExceptGroupSig(&unit, isGenesis)
	if unitState != modules.UNIT_STATE_VALIDATED && unitState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
		return -1, fmt.Errorf("SaveDag, validate unit error, errno=%d", unitState)
	}
	if unitState == modules.UNIT_STATE_VALIDATED {
		// step3.1. pass and with group signature, put into leveldb
		if err := d.unitRep.SaveUnit(unit, false); err != nil {
			return 1, fmt.Errorf("SaveDag, save error when save unit to db: %s", err.Error())
		}
		// step3.2. if pass and with group signature, prune fork data
		if err := d.Memdag.Prune(unit.UnitHeader.Number.AssetID.String(), unit.UnitHash); err != nil {
			return 2, fmt.Errorf("SaveDag, save error when prune: %s", err.Error())
		}
	} else {
		// step4. pass but without group signature, put into memory( if the main fork longer than 15, should call prune)
		if err := d.Memdag.Save(&unit); err != nil {
			return 3, fmt.Errorf("SaveDag, save error: %s", err.Error())
		}
	}
	// step5. check if it is need to switch
	if err := d.Memdag.SwitchMainChain(); err != nil {
		return 4, fmt.Errorf("SaveDag, save error when switch chain: %s", err.Error())
	}
	return 0, nil
}

// InsertDag attempts to insert the given batch of blocks in to the canonical
// chain or, otherwise, create a fork. If an error is returned it will return
// the index number of the failing block as well an error describing what went
// wrong.
// After insertion is done, all accumulated events will be fired.
// reference : Eth InsertChain
func (d *Dag) InsertDag(units modules.Units) (int, error) {
	//TODO must recover
	log.Debug("===InsertDag===", "len(units):", len(units))
	count := int(0)
	for i, u := range units {
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
		if err := d.unitRep.SaveUnit(*u, false); err != nil {
			fmt.Errorf("Insert dag, save error: %s", err.Error())
			return count, err
		}
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
	index := new(modules.ChainIndex)
	index.Index = number
	//fmt.Println(hash)
	//fmt.Println(number)
	// copy(index.AssetID[:], assetId[:])
	// index.IsMain = onMain
	if h, err := d.dagdb.GetHeader(hash, index); err == nil && h != nil {
		return true
	}
	return false
}
func (d *Dag) Exists(hash common.Hash) bool {
	number, err := d.dagdb.GetNumberWithUnitHash(hash)
	if err == nil && (number != modules.ChainIndex{}) {
		log.Info("经检索，该hash已存储在leveldb中，", "hash", hash.String())
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
func (d *Dag) GetTransactionByHash(hash common.Hash) (*modules.Transaction, error) {
	tx, _, _, _ := d.dagdb.GetTransaction(hash)
	if tx == nil {
		return nil, fmt.Errorf("GetTransactionByHash: get none transaction")
	}
	return tx, nil
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
	return []*modules.Header{}, nil
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
		ChainId:  chainid,
	}

	return d.utxoRep.WalletBalance(address, asset), nil
}

func NewDag(db ptndb.Database,l log.ILogger) (*Dag, error) {
	mutex := new(sync.RWMutex)

	dagDb := storage.NewDagDb(db,l)
	utxoDb := storage.NewUtxoDb(db,l)
	stateDb := storage.NewStateDb(db,l)
	idxDb := storage.NewIndexDb(db,l)
	propDb, err := storage.NewPropertyDb(db,l)
	if err != nil {
		return nil, err
	}

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb,l)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb,l)
	validate := dagcommon.NewValidate(dagDb, utxoDb, stateDb,l)
	propRep := dagcommon.NewPropRepository(propDb,l)

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
		Memdag:        memunit.NewMemDag(dagDb, unitRep),
	}

	return dag, nil
}

func NewDag4GenesisInit(db ptndb.Database) (*Dag, error) {
	mutex := new(sync.RWMutex)
	logger:=log.New("Dag")
	dagDb := storage.NewDagDb(db,logger)
	utxoDb := storage.NewUtxoDb(db,logger)
	stateDb := storage.NewStateDb(db,logger)
	idxDb := storage.NewIndexDb(db,logger)
	propDb := storage.NewPropertyDb4GenesisInit(db)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb,logger)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb,logger)
	validate := dagcommon.NewValidate(dagDb, utxoDb, stateDb,logger)
	propRep := dagcommon.NewPropRepository(propDb,logger)

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
	}
	return dag, nil
}

func NewDagForTest(db ptndb.Database) (*Dag, error) {
	mutex := new(sync.RWMutex)
	logger:=log.New("Dag")
	dagDb := storage.NewDagDb(db,logger)
	utxoDb := storage.NewUtxoDb(db,logger)
	stateDb := storage.NewStateDb(db,logger)
	idxDb := storage.NewIndexDb(db,logger)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb,logger)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb,logger)
	validate := dagcommon.NewValidate(dagDb, utxoDb, stateDb,logger)

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

		//Memdag:        memunit.NewMemDag(dagDb, unitRep),
	}
	return dag, nil
}

// Get Contract Api
func (d *Dag) GetContract(id common.Hash) (*modules.Contract, error) {
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
		head, err := d.dagdb.GetHeader(hash, &index)
		if err != nil {
			fmt.Println("=============get unit header faled =============", err)
		}
		return head, err
	}
	return nil, err
}

// Get UnitNumber
func (d *Dag) GetUnitNumber(hash common.Hash) (modules.ChainIndex, error) {
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

func (d *Dag) GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error) {
	neededSet := make(map[modules.OutPoint]struct{})
	preout := modules.OutPoint{TxHash: tx.Hash()}
	for i, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				msgIdx := uint32(i)
				preout.MessageIndex = msgIdx
				for j := range msg.Output {
					txoutIdx := uint32(j)
					preout.OutIndex = txoutIdx
					neededSet[preout] = struct{}{}
				}
			}
		}

	}
	// if tx is Not CoinBase
	// add txIn previousoutpoint
	view := txspool.NewUtxoViewpoint()
	d.Mutex.RLock()
	err := view.FetchUtxos(d.utxodb, neededSet)
	d.Mutex.RUnlock()

	return view, err
}

// GetAllUtxos is return all utxo.
func (d *Dag) GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error) {
	d.Mutex.RLock()
	items, err := d.utxodb.GetAllUtxos()
	d.Mutex.RUnlock()

	return items, err
}

func (d *Dag) SaveUtxoView(view *txspool.UtxoViewpoint) error {
	//return txspool.SaveUtxoView(db, view)
	return d.utxodb.SaveUtxoView(view.Entries())
}
func (d *Dag) GetAddrOutpoints(addr string) ([]modules.OutPoint, error) {
	return d.utxodb.GetAddrOutpoints(addr)
}

func (d *Dag) GetAddrOutput(addr string) ([]modules.Output, error) {
	return d.dagdb.GetAddrOutput(addr)
}

func (d *Dag) GetAddrUtxos(addr string) ([]modules.Utxo, error) {
	return d.utxodb.GetAddrUtxos(addr)
}

func (d *Dag) GetAddrTransactions(addr string) (modules.Transactions, error) {
	return d.dagdb.GetAddrTransactions(addr)
}

// get contract state
func (d *Dag) GetContractState(id string, field string) (*modules.StateVersion, []byte) {
	return d.GetContractState(id, field)
}

func (d *Dag) CreateUnit(mAddr *common.Address, txpool *txspool.TxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error) {
	return d.unitRep.CreateUnit(mAddr, txpool, ks, t)
}
func (d *Dag) SaveUnit(unit modules.Unit, isGenesis bool) error {
	return d.unitRep.SaveUnit(unit, isGenesis)
}
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
		ParentsHash:  []common.Hash{currentUnit.UnitHash},
		AssetIDs:     []modules.IDType16{currentUnit.UnitHeader.Number.AssetID},
		Authors:      nil,
		GroupSign:    make([]byte, 0),
		Number:       height,
		Creationdate: time.Now().Unix(),
	}

	sAddr := "P1NsG3kiKJc87M6Di6YriqHxqfPhdvxVj2B"
	addr, err := common.StringToAddress(sAddr)
	if err != nil {

	}
	bAsset := d.statedb.GetConfig([]byte("GenesisAsset"))
	if len(bAsset) <= 0 {
		return nil, fmt.Errorf("Create unit error: query asset info empty")
	}
	var asset modules.Asset
	if err := rlp.DecodeBytes(bAsset, &asset); err != nil {
		return nil, fmt.Errorf("Create unit: %s", err.Error())
	}
	coinbase, err := dagcommon.CreateCoinbase(&addr, 0, &asset, time.Now())
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
func (d *Dag) UpdateGlobalDynProp(gp *modules.GlobalProperty, dgp *modules.DynamicGlobalProperty, unit *modules.Unit) {
	d.propRep.UpdateGlobalDynProp(gp, dgp, unit)
}

//@Yiran
func (d *Dag) GetCurrentUnitIndex() (modules.ChainIndex, error) {
	currentUnitHash := d.CurrentUnit().UnitHash
	return d.GetUnitNumber(currentUnitHash)
}

//@Yiran save utxo snapshot when new mediator cycle begin
// unit index MUST to be  integer multiples of  termInterval.
func (d *Dag) SaveUtxoSnapshot() error {
	currentUnitIndex, err := d.GetCurrentUnitIndex()
	if err != nil {
		return err
	}
	return d.utxodb.SaveUtxoSnapshot(currentUnitIndex)
}

//@Yiran Get last utxo snapshot
// must calling after SaveUtxoSnapshot call , before this mediator cycle end.
// called by GenerateVoteResult
func (d *Dag) GetUtxoSnapshot() (*[]modules.Utxo, error) {
	unitIndex, err := d.GetCurrentUnitIndex()
	if err != nil {
		return nil, err
	}
	unitIndex.Index -= unitIndex.Index % modules.TERMINTERVAL
	return d.utxodb.GetUtxoEntities(unitIndex)
}

//@Yiran
func (d *Dag) GenerateVoteResult() (*[]storage.Candidate, error) {
	VoteBox := storage.NewVoteBox()

	utxos, err := d.utxodb.GetAllUtxos()
	if err != nil {
		return nil, err
	}
	for _, utxo := range utxos {
		if utxo.Asset.AssetId == modules.PTNCOIN {
			utxoHolder, err := tokenengine.GetAddressFromScript(utxo.PkScript)
			if err != nil {
				return nil, err
			}
			VoteBox.AddToBoxIfNotVoted(utxoHolder, utxo.VoteResult)
		}
	}
	VoteBox.Sort()
	return &VoteBox.Candidates, nil
}

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
//	//voteBox := storage.VoteBox{}
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
