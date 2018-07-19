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
