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
	"sync"

	"fmt"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
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
	return modules.NewUnit(&modules.Header{
		Extra: []byte("test pool"),
	}, nil)
}

func (d *Dag) GetUnit(hash common.Hash) *modules.Unit {
	return d.CurrentUnit()
}

func (d *Dag) GetUnitByHash(hash common.Hash) *modules.Unit {
	return d.CurrentUnit()
}

func (d *Dag) GetUnitByNumber(number uint64) *modules.Unit {
	return d.CurrentUnit()
}

func (d *Dag) GetHeaderByHash(hash common.Hash) *modules.Header {
	return d.CurrentUnit().Header()
}

func (d *Dag) GetHeaderByNumber(number uint64) *modules.Header {
	return d.CurrentUnit().Header()
}

func (d *Dag) GetHeader(hash common.Hash, number uint64) *modules.Header {
	return d.CurrentUnit().Header()
}

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

// InsertDag attempts to insert the given batch of blocks in to the canonical
// chain or, otherwise, create a fork. If an error is returned it will return
// the index number of the failing block as well an error describing what went
// wrong.
// After insertion is done, all accumulated events will be fired.
// reference : Eth InsertChain
func (d *Dag) SaveDag(unit modules.Unit) (int, error) {
	return 0, nil
	//SaveUnit
}

/**
将连续的单元保存到DAG中
To save some continuous units to dag storage
*/
func (d *Dag) InsertDag(units modules.Units) (int, error) {
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

func (d *Dag) HasHeader(hash common.Hash, number uint64) bool {
	return d.CurrentUnit().Header() != nil
}

func (d *Dag) CurrentHeader() *modules.Header {
	return d.CurrentUnit().Header()
}

// GetBodyRLP retrieves a block body in RLP encoding from the database by hash,
// caching it if found.
func (d *Dag) GetBodyRLP(hash common.Hash) rlp.RawValue {
	return rlp.RawValue{}
}

func (d *Dag) GetHeaerRLP(hash common.Hash) rlp.RawValue {
	return rlp.RawValue{}
}

// InsertHeaderChain attempts to insert the given header chain in to the local
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
