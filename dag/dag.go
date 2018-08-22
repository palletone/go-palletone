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
	"sync"

	"github.com/coocood/freecache"
	//"github.com/ethereum/go-ethereum/params"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/configure"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Dag struct {
	Cache *freecache.Cache
	Db    *palletdb.LDBDatabase

	Mdb           *palletdb.MemDatabase
	ChainHeadFeed *event.Feed
	// GenesisUnit   *Unit  // comment by Albert·Gou

	Mutex sync.RWMutex

	GlobalProp    *modules.GlobalProperty
	DynGlobalProp *modules.DynamicGlobalProperty
	MediatorSchl  *modules.MediatorSchedule
}

func (d *Dag) CurrentUnit() *modules.Unit {
	// step1. get current unit hash
	hash, err := d.GetHeadUnitHash()
	if err != nil {
		return nil
	}
	// step2. get unit height
	height := d.GetUnitNumber(hash)
	// get unit header
	uHeader, err := storage.GetHeader(d.Db, hash, &height)
	if err != nil {
		log.Error("Current unit when get unit header", "error", err.Error())
		return nil
	}
	// get unit hash
	uHash := common.Hash{}
	uHash.SetBytes(hash.Bytes())

	// get transaction list
	txs, err := dagcommon.GetUnitTransactions(uHash)
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

func (d *Dag) GetUnit(hash common.Hash) *modules.Unit {
	return storage.GetUnit(d.Db, hash)
}

func (d *Dag) GetUnitByHash(hash common.Hash) *modules.Unit {
	return d.CurrentUnit()
}

func (d *Dag) GetUnitByNumber(number uint64) *modules.Unit {
	return d.CurrentUnit()
}

func (d *Dag) GetHeaderByHash(hash common.Hash) *modules.Header {
	unit := d.CurrentUnit()
	if unit != nil {
		return unit.Header()
	}
	return nil
}

func (d *Dag) GetHeaderByNumber(number uint64) *modules.Header {
	unit := d.CurrentUnit()
	if unit != nil {
		return unit.Header()
	}
	return nil
}

// func (d *Dag) GetHeader(hash common.Hash, number uint64) *modules.Header {
// 	return d.CurrentUnit().Header()
// }

func (d *Dag) StateAt(common.Hash) (*palletdb.MemDatabase, error) {
	return d.Mdb, nil
}

func (d *Dag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return d.ChainHeadFeed.Subscribe(ch)
}

// FastSyncCommitHead sets the current head block to the one defined by the hash
// irrelevant what the chain contents were prior.
func (d *Dag) FastSyncCommitHead(hash common.Hash) error {
	return nil
}

func (d *Dag) SaveDag(unit modules.Unit) (int, error) {
	if err := dagcommon.SaveUnit(unit, false); err != nil {
		fmt.Errorf("SaveDag, save error: %s", err.Error())
		return -1, err
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
	return len(units), nil
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
		if err := dagcommon.SaveUnit(*u, false); err != nil {
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
	//GetUnit()
	return []common.Hash{}
}

// need add:   assetId modules.IDType16, onMain bool
func (d *Dag) HasHeader(hash common.Hash, number uint64) bool {
	index := new(modules.ChainIndex)
	index.Index = number
	// copy(index.AssetID[:], assetId[:])
	// index.IsMain = onMain
	if h, err := storage.GetHeader(d.Db, hash, index); err == nil && h != nil {
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
func (d *Dag) GetBodyRLP(db storage.DatabaseReader, hash common.Hash) rlp.RawValue {
	txs := modules.Transactions{}
	// get hash list
	txs, err := dagcommon.GetUnitTransactions(hash)
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
	number, err := storage.GetUnitNumber(db, hash)
	if err != nil {
		log.Error("Get header rlp ", "error", err.Error())
		return nil
	}
	return storage.GetHeaderRlp(db, hash, number.Index)
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
	return checkFreq, nil
}

//VerifyHeader checks whether a header conforms to the consensus rules of the stock
//Ethereum ethash engine.go
func (d *Dag) VerifyHeader(header *modules.Header, seal bool) error {
	// step1. check unit signature, should be compare to mediator list
	if err := dagcommon.ValidateUnitSignature(header, false); err != nil {
		log.Info("Validate unit signature", "error", err.Error())
		return err
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
	return dagcommon.GetAccountTokens(addr)
}

func (d *Dag) WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
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
		AssertId: newAssetid,
		UniqueId: newUnitqueid,
		ChainId:  chainid,
	}

	addr := common.Address{}
	addr.SetString(address)
	return dagcommon.WalletBalance(addr, asset), nil
}

func NewDag() *Dag {
	// genesis, _ := NewGenesisUnit(nil) // comment by Albert·Gou
	db, _ := palletdb.NewMemDatabase()
	mutex := new(sync.RWMutex)
	return &Dag{
		Cache: freecache.NewCache(200 * 1024 * 1024),
		Db:    storage.Dbconn,
		Mdb:   db,
		//    GenesisUnit:   genesis, // comment by Albert·Gou
		ChainHeadFeed: new(event.Feed),
		Mutex:         *mutex,
		GlobalProp:    storage.RetrieveGlobalProp(),
		DynGlobalProp: storage.RetrieveDynGlobalProp(),
		MediatorSchl:  storage.RetrieveMediatorSchl(),
	}

}

// Get Contract Api
func (d *Dag) GetContract(id common.Hash) (*modules.Contract, error) {
	return storage.GetContract(d.Db, id)
}

// Get Header
func (d *Dag) GetHeader(hash common.Hash, number uint64) (*modules.Header, error) {
	index := d.GetUnitNumber(hash)
	return storage.GetHeader(d.Db, hash, &index)
}

// Get UnitNumber
func (d *Dag) GetUnitNumber(hash common.Hash) modules.ChainIndex {
	height, _ := storage.GetUnitNumber(d.Db, hash)
	return height
}

// GetCanonicalHash
func (d *Dag) GetCanonicalHash(number uint64) (common.Hash, error) {
	return storage.GetCanonicalHash(d.Db, number)
}

// Get state
func (d *Dag) GetHeadHeaderHash() (common.Hash, error) {
	return storage.GetHeadHeaderHash(d.Db)
}

func (d *Dag) GetHeadUnitHash() (common.Hash, error) {
	return storage.GetHeadUnitHash(d.Db)
}

func (d *Dag) GetHeadFastUnitHash() (common.Hash, error) {
	return storage.GetHeadFastUnitHash(d.Db)
}

func (d *Dag) GetTrieSyncProgress() (uint64, error) {
	return storage.GetTrieSyncProgress(d.Db)
}

func (d *Dag) GetUtxoEntry(key []byte) (*modules.Utxo, error) {
	return storage.GetUtxoEntry(d.Db, key)
}

func (d *Dag) GetAddrOutput(addr string) ([]modules.Output, error) {
	return storage.GetAddrOutput(d.Db, addr)
}

func (d *Dag) GetAddrTransactions(addr string) (modules.Transactions, error) {
	return storage.GetAddrTransactions(d.Db, addr)
}
