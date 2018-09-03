// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
package fortest

import (
	"fmt"
	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"sync"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"bytes"
)
type Dag struct {
	Cache         *freecache.Cache
	db            ptndb.Database
	ChainHeadFeed *event.Feed
	Mutex         sync.RWMutex
	GlobalProp    *modules.GlobalProperty
	DynGlobalProp *modules.DynamicGlobalProperty
	MediatorSchl  *modules.MediatorSchedule
}

func MakeDags(Memdb ptndb.Database,number int) (*Dag, error) {
	header := NewHeader([]common.Hash{}, []IDType16{PTNCOIN}, []byte{})
	header.Number.AssetID = PTNCOIN
	header.Number.IsMain = true
	header.Number.Index = 0
	header.Authors = &Authentifier{"P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ",[]byte{},[]byte{},[]byte{}}
	header.Witness = []*Authentifier{&Authentifier{"P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ",[]byte{},[]byte{},[]byte{}}}
	tx, _ := NewCoinbaseTransaction()
	txs := Transactions{tx}
	genesisUnit := NewUnit(header, txs)
	//fmt.Printf("--------这是最新块----unit.UnitHeader-----%#v\n", genesisUnit.UnitHeader)
	err := SaveGenesis(Memdb, genesisUnit)
	if err != nil {
		fmt.Println("SaveGenesis, err",err)
		return nil,err
	}
	fmt.Printf("--------这是最新块----unit-----%#v\n", genesisUnit)
	fmt.Printf("--------这是最新块----unit.UnitHeader-----%#v\n", genesisUnit.UnitHeader)
	fmt.Printf("--------这是最新块----unit.Txs-----%#v\n", genesisUnit.Txs[0].Hash())
	fmt.Printf("--------这是最新块----unit.UnitHash-----%#v\n", genesisUnit.UnitHash)
	fmt.Printf("--------这是最新块----unit.UnitHeader.ParentsHash-----%#v\n", genesisUnit.UnitHeader.ParentsHash)
	fmt.Printf("--------这是最新块----unit.UnitHeader.Number.Index-----%#v\n", genesisUnit.UnitHeader.Number.Index)
	units, _ := newDag(Memdb, genesisUnit, number)
	fmt.Println("len(units).........", len(units))
	for i, v := range units {
		fmt.Printf("%d====%#v\n", i, v)
	}
	dag := NewDag(Memdb)
	uu := dag.CurrentUnit()
	fmt.Printf("current===>>>%#v\n",uu)
	fmt.Printf("--------这是最新块----unit.UnitHeader-----%#v\n", uu.UnitHeader)
	fmt.Printf("--------这是最新块----unit.Txs-----%#v\n", uu.Txs[0].Hash())
	fmt.Printf("--------这是最新块----unit.UnitHash-----%#v\n", uu.UnitHash)
	fmt.Printf("--------这是最新块----unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	fmt.Printf("--------这是最新块----unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	//_, err := dag.InsertDag(units)
	//if err != nil {
	//	panic(err)
	//}
	return dag, nil
}
func (d *Dag) CurrentUnit() *Unit {
	// step1. get current unit hash
	hash, err := d.GetHeadUnitHash()

	if err != nil {
		return nil
	}

	// step2. get unit height
	height := d.GetUnitNumber(hash)
	fmt.Printf("height := d.GetUnitNumber(hash)%#v\n\n",height)
	//无错
	// get unit header
	uHeader, err := GetHeader(d.db, hash, &height)
	if err != nil {
		log.Error("Current unit when get unit header", "error", err.Error())
		return nil
	}
	// get unit hash
	uHash := common.Hash{}
	uHash.SetBytes(hash.Bytes())

	// get transaction list
	tx, err := GetUnitTransactions(d.db,uHash)
	if err != nil {
		log.Error("Current unit when get transactions", "error", err.Error())
		return nil
	}
	txs := Transactions{tx}
	// generate unit
	unit := Unit{
		UnitHeader: uHeader,
		UnitHash:   uHash,
		Txs:        txs,
	}
	unit.UnitSize = unit.Size()
	return &unit
}


func GetUnitTransactions(db ptndb.Database,unitHash common.Hash) (*Transaction, error) {
	tx_bytes,err := db.Get(append(storage.TRANSACTION_PREFIX, unitHash.Bytes()...))
	if err != nil {
		return nil, err
	}
	transaction := new(Transaction)
	if err := rlp.Decode(bytes.NewReader(tx_bytes), transaction); err != nil {
		log.Error("Invalid unit header rlp:", err)
		return nil, err
	}
	return transaction, nil
}
func (d *Dag) GetCurrentUnit(assetId IDType16) *Unit {
	return d.CurrentUnit()
}

func (d *Dag) GetUnit(hash common.Hash) *Unit {
	return GetUnit(d.db, hash)
}

func (d *Dag) HasUnit(hash common.Hash) bool {
	return GetUnit(d.db, hash) != nil
}

func (d *Dag) GetUnitByHash(hash common.Hash) *Unit {
	//TODO must modify
	return nil
	return d.CurrentUnit()
}

func (d *Dag) GetUnitByNumber(number ChainIndex) *Unit {
	return GetUnitFormIndex(d.db, number.Index, number.AssetID)
}

func (d *Dag) GetHeaderByHash(hash common.Hash) *Header {
	height := d.GetUnitNumber(hash)
	// get unit header
	uHeader, err := GetHeader(d.db, hash, &height)
	if err != nil {
		log.Error("Current unit when get unit header", "error", err.Error())
		return nil
	}
	return uHeader
}

func (d *Dag) GetHeaderByNumber(number ChainIndex) *Header {
	header, _ := GetHeaderByHeight(d.db, number)
	return header
}

// func (d *Dag) GetHeader(hash common.Hash, number uint64) *Header {
// 	return d.CurrentUnit().Header()
// }

//func (d *Dag) StateAt(common.Hash) (*ptndb.MemDatabase, error) {
//	return d.Mdb, nil
//}

//func (d *Dag) SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription {
//	return d.ChainHeadFeed.Subscribe(ch)
//}

// FastSyncCommitHead sets the current head block to the one defined by the hash
// irrelevant what the chain contents were prior.
func (d *Dag) FastSyncCommitHead(hash common.Hash) error {
	return nil
}

//func (d *Dag) SaveDag(unit Unit) (int, error) {
//	if err := SaveUnit(unit, false); err != nil {
//		fmt.Errorf("SaveDag, save error: %s", err.Error())
//		return -1, err
//	}
//	return 0, nil
//}

// InsertDag attempts to insert the given batch of blocks in to the canonical
// chain or, otherwise, create a fork. If an error is returned it will return
// the index number of the failing block as well an error describing what went
// wrong.
// After insertion is done, all accumulated events will be fired.
// reference : Eth InsertChain
func (d *Dag) InsertDag(db ptndb.Database,units Units) (int, error) {
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
		if err := SaveUnit(db,u); err != nil {
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

// need add:   assetId IDType16, onMain bool
func (d *Dag) HasHeader(hash common.Hash, number uint64) bool {
	index := new(ChainIndex)
	index.Index = number
	// copy(index.AssetID[:], assetId[:])
	// index.IsMain = onMain
	if h, err := GetHeader(d.db, hash, index); err == nil && h != nil {
		return true
	}
	return false
}

func (d *Dag) CurrentHeader() *Header {
	unit := d.CurrentUnit()
	if unit != nil {
		return unit.Header()
	}
	return nil
}

// GetBodyRLP retrieves a block body in RLP encoding from the database by hash,
// caching it if found.
//func (d *Dag) GetBodyRLP(hash common.Hash) rlp.RawValue {
//	return d.getBodyRLP(d.db, hash)
//}

func (d *Dag) GetTransactionsByHash(hash common.Hash) (Transactions, error) {
	tx, err := GetUnitTransactions(d.db,hash)
	if err != nil {
		log.Error("Get body rlp", "unit hash", hash.String(), "error", err.Error())
		return nil, err
	}
	txs := Transactions{tx}
	return txs, nil
}

func (d *Dag) getBodyRLP(db DatabaseReader, hash common.Hash) rlp.RawValue {
	// get hash list
	tx, err := GetUnitTransactions(d.db,hash)
	txs := Transactions{tx}
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

func (d *Dag) GetHeaderRLP(db ptndb.Database, hash common.Hash) rlp.RawValue {
	number, err := GetUnitNumber(db, hash)
	if err != nil {
		log.Error("Get header rlp ", "error", err.Error())
		return nil
	}
	return GetHeaderRlp(db, hash, number.Index)
}

// InsertHeaderDag attempts to insert the given header chain in to the local
// chain, possibly creating a reorg. If an error is returned, it will return the
// index number of the failing header as well an error describing what went wrong.
//
// The verify parameter can be used to fine tune whether nonce verification
// should be done or not. The reason behind the optional check is because some
// of the header retrieval mechanisms already need to verify nonces, as well as
// because nonces can be verified sparsely, not needing to check each.
func (d *Dag) InsertHeaderDag(headers []*Header, checkFreq int) (int, error) {
	return checkFreq, nil
}

//VerifyHeader checks whether a header conforms to the consensus rules of the stock
//Ethereum ethash engine.go
//
//func (d *Dag) VerifyHeader(header *Header, seal bool) error {
//	// step1. check unit signature, should be compare to mediator list
//	if err := ValidateUnitSignature(header, false); err != nil {
//		log.Info("Validate unit signature", "error", err.Error())
//		return err
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

//All leaf nodes for dag downloader.
//MUST have Priority.
func (d *Dag) GetAllLeafNodes() ([]*Header, error) {
	return []*Header{}, nil
}

/**
获取account address下面的token信息
To get account token list and tokens's information
*/
//func (d *Dag) WalletTokens(addr common.Address) (map[string]*AccountToken, error) {
//	return dagcommon.GetAccountTokens(addr)
//}

//func (d *Dag) WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
//	newAssetid := IDType16{}
//	newUnitqueid := IDType16{}
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
//	asset := Asset{
//		AssertId: newAssetid,
//		UniqueId: newUnitqueid,
//		ChainId:  chainid,
//	}
//
//	addr := common.Address{}
//	addr.SetString(address)
//	return WalletBalance(addr, asset), nil
//}

func NewDag(db ptndb.Database) *Dag {


	return &Dag{
		Cache: freecache.NewCache(200 * 1024 * 1024),
		db:    db,
		//    GenesisUnit:   genesis, // comment by Albert·Gou
		ChainHeadFeed: new(event.Feed),
		Mutex:         * new(sync.RWMutex),
		GlobalProp:    storage.RetrieveGlobalProp(db),
		DynGlobalProp: storage.RetrieveDynGlobalProp(db),
		MediatorSchl:  storage.RetrieveMediatorSchl(db),
	}

}

// Get Contract Api
//func (d *Dag) GetContract(id common.Hash) (*Contract, error) {
//	return GetContract(d.db, id)
//}

// Get Header
func (d *Dag) GetHeader(hash common.Hash, number uint64) (*Header, error) {
	index := d.GetUnitNumber(hash)
	//TODO compare index with number
	return GetHeader(d.db, hash, &index)
}

// Get UnitNumber
func (d *Dag) GetUnitNumber(hash common.Hash) ChainIndex {
	height, _ := GetUnitNumber(d.db, hash)
	//fmt.Printf("height====%#v",height)
	return height
}

// GetCanonicalHash
func (d *Dag) GetCanonicalHash(number uint64) (common.Hash, error) {
	return GetCanonicalHash(d.db, number)
}

// Get state
func (d *Dag) GetHeadHeaderHash() (common.Hash, error) {
	return GetHeadHeaderHash(d.db)
}

func (d *Dag) GetHeadUnitHash() (common.Hash, error) {
	return GetHeadUnitHash(d.db)
}

func (d *Dag) GetHeadFastUnitHash() (common.Hash, error) {
	return GetHeadFastUnitHash(d.db)
}

func (d *Dag) GetTrieSyncProgress() (uint64, error) {
	return GetTrieSyncProgress(d.db)
}

//func (d *Dag) GetUtxoEntry(key []byte) (*Utxo, error) {
//	return GetUtxoEntry(d.db, key)
//}

func (d *Dag) GetAddrOutput(addr string) ([]Output, error) {
	return GetAddrOutput(d.db, addr)
}

func (d *Dag) GetAddrTransactions(addr string) (Transactions, error) {
	return GetAddrTransactions(d.db, addr)
}

//// author Albert·Gou
//func (d *Dag) GetActiveMediatorNodes() []*discover.Node {
//	return d.GlobalProp.GetActiveMediatorNodes()
//}
