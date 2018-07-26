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

package modules

import (
	"sync"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
)

type Dag struct {
	Cache *freecache.Cache
	Db    *palletdb.LDBDatabase

	Mdb           *palletdb.MemDatabase
	ChainHeadFeed *event.Feed
	GenesisUnit   *Unit

	Mutex sync.RWMutex

	GlobalProp    *GlobalProperty
	DynGlobalProp *DynamicGlobalProperty
	MediatorSchl  *MediatorSchedule
}

func (d *Dag) CurrentUnit() *Unit {
	return NewUnit(&Header{
		Extra: []byte("test pool"),
	}, nil)
}

func (d *Dag) GetUnit(hash common.Hash, number uint64) *Unit {
	return d.CurrentUnit()
}

func (d *Dag) StateAt(common.Hash) (*palletdb.MemDatabase, error) {
	return d.Mdb, nil
}

func (d *Dag) SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription {
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
func (bc *Dag) InsertDag(unit Unit) (int, error) {
	return 0, nil
}
