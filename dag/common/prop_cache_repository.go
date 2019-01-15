/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package common

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/dag/storage"
	"time"
)

//带缓存的Prop存取仓库，用于加快读写的速度
type PropCacheRepository struct {
	dbRep *PropRepository
	cache palletcache.ICache
}

var (
	GlobalProperty        = []byte("GlobalProperty")
	DynamicGlobalProperty = []byte("DynamicGlobalProperty")
	MediatorSchedule      = []byte("MediatorSchedule")
)

func NewPropCacheRepository(db storage.IPropertyDb, cache palletcache.ICache) *PropCacheRepository {
	dbRep := NewPropRepository(db)
	return &PropCacheRepository{dbRep: dbRep, cache: cache}
}
func (pRep *PropCacheRepository) storeToCache(key []byte, value interface{}) {
	data, _ := rlp.EncodeToBytes(value)
	pRep.cache.Set(key, data, 0)
}
func (pRep *PropCacheRepository) retrieveFromCache(key []byte, value interface{}) bool {
	data, err := pRep.cache.Get(key)
	if err != nil {
		return false
	}

	if err = rlp.DecodeBytes(data, value); err != nil {
		return false
	}
	return true
}

func (pRep *PropCacheRepository) StoreGlobalProp(gp *modules.GlobalProperty) error {
	pRep.storeToCache(GlobalProperty, gp)
	return pRep.dbRep.StoreGlobalProp(gp)
}
func (pRep *PropCacheRepository) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	gp := &modules.GlobalProperty{}
	if !pRep.retrieveFromCache(GlobalProperty, gp) {
		gp, err := pRep.dbRep.RetrieveGlobalProp()
		if err != nil {
			return nil, err
		}
		pRep.storeToCache(GlobalProperty, gp)
	}
	return gp, nil
}
func (pRep *PropCacheRepository) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {
	pRep.storeToCache(DynamicGlobalProperty, dgp)
	return pRep.dbRep.StoreDynGlobalProp(dgp)
}
func (pRep *PropCacheRepository) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	gp := &modules.DynamicGlobalProperty{}
	pRep.retrieveFromCache(DynamicGlobalProperty, gp)
	//if !pRep.retrieveFromCache(DynamicGlobalProperty, gp) {
	//	gp, err := pRep.dbRep.RetrieveDynGlobalProp()
	//	if err != nil {
	//		return nil, err
	//	}
	//	pRep.storeToCache(DynamicGlobalProperty, gp)
	//}
	return gp, nil
}
func (pRep *PropCacheRepository) StoreMediatorSchl(ms *modules.MediatorSchedule) error {
	pRep.storeToCache(MediatorSchedule, ms)
	return pRep.dbRep.StoreMediatorSchl(ms)
}
func (pRep *PropCacheRepository) RetrieveMediatorSchl() (*modules.MediatorSchedule, error) {
	gp := &modules.MediatorSchedule{}
	if !pRep.retrieveFromCache(MediatorSchedule, gp) {
		gp, err := pRep.dbRep.RetrieveMediatorSchl()
		if err != nil {
			return nil, err
		}
		pRep.storeToCache(MediatorSchedule, gp)
	}
	return gp, nil
}
func (pRep *PropCacheRepository) SetLastStableUnit(hash common.Hash, index *modules.ChainIndex) error {
	return pRep.dbRep.SetLastStableUnit(hash, index)
}
func (pRep *PropCacheRepository) GetLastStableUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, error) {
	return pRep.dbRep.GetLastStableUnit(token)
}
func (pRep *PropCacheRepository) SetNewestUnit(header *modules.Header) error {
	return pRep.dbRep.SetNewestUnit(header)
}
func (pRep *PropCacheRepository) GetNewestUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, error) {
	return pRep.dbRep.GetNewestUnit(token)
}
func (pRep *PropCacheRepository) GetNewestUnitTimestamp(token modules.IDType16) (int64, error) {
	return pRep.dbRep.GetNewestUnitTimestamp(token)
}
func (pRep *PropCacheRepository) UpdateMediatorSchedule(ms *modules.MediatorSchedule, gp *modules.GlobalProperty, dgp *modules.DynamicGlobalProperty) bool {
	return pRep.dbRep.UpdateMediatorSchedule(ms, gp, dgp)
}
func (pRep *PropCacheRepository) GetSlotTime(gp *modules.GlobalProperty, dgp *modules.DynamicGlobalProperty, slotNum uint32) time.Time {
	return pRep.dbRep.GetSlotTime(gp, dgp, slotNum)
}
func (pRep *PropCacheRepository) GetSlotAtTime(gp *modules.GlobalProperty, dgp *modules.DynamicGlobalProperty, when time.Time) uint32 {
	return pRep.dbRep.GetSlotAtTime(gp, dgp, when)
}
