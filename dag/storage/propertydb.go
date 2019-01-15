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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package storage

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

// modified by Yiran
type PropertyDb struct {
	db ptndb.Database
	//GlobalProp    *modules.GlobalProperty
	//DynGlobalProp *modules.DynamicGlobalProperty
	//MediatorSchl  *modules.MediatorSchedule
}
type IPropertyDb interface {
	StoreGlobalProp(gp *modules.GlobalProperty) error
	RetrieveGlobalProp() (*modules.GlobalProperty, error)
	StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error
	RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error)
	StoreMediatorSchl(ms *modules.MediatorSchedule) error
	RetrieveMediatorSchl() (*modules.MediatorSchedule, error)

	//设置稳定单元的Hash
	SetLastStableUnit(hash common.Hash, index *modules.ChainIndex) error
	GetLastStableUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, error)
	SetNewestUnit(header *modules.Header) error
	GetNewestUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, int64, error)
}

// modified by Yiran
// initialize PropertyDB , and retrieve gp,dgp,mc from IPropertyDb.
func NewPropertyDb(db ptndb.Database) *PropertyDb {
	pdb := &PropertyDb{db: db}

	//gp, err := pdb.RetrieveGlobalProp()
	//if err != nil {
	//	logger.Error("RetrieveGlobalProp Error")
	//	return nil,err
	//}
	//
	//dgp, err := pdb.RetrieveDynGlobalProp()
	//if err != nil {
	//	logger.Error("RetrieveDynGlobalProp Error")
	//	return nil,err
	//}
	//
	//ms, err := pdb.RetrieveMediatorSchl()
	//if err != nil {
	//	logger.Error("RetrieveMediatorSchl Error")
	//	return nil,err
	//}
	//pdb.GlobalProp = gp
	//pdb.DynGlobalProp = dgp
	//pdb.MediatorSchl = ms

	return pdb
}

func (propdb *PropertyDb) StoreMediatorSchl(ms *modules.MediatorSchedule) error {
	log.Debug("Save mediator schedule to db.")
	return StoreMediatorSchl(propdb.db, ms)
}

func (propdb *PropertyDb) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {
	log.Debug("Save dynamic global property to db.")
	return StoreDynGlobalProp(propdb.db, dgp)
}

func (propdb *PropertyDb) StoreGlobalProp(gp *modules.GlobalProperty) error {
	log.Debug("Save global property to db.")
	return StoreGlobalProp(propdb.db, gp)
}

func (propdb *PropertyDb) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	return RetrieveGlobalProp(propdb.db)
}

func (propdb *PropertyDb) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	return RetrieveDynGlobalProp(propdb.db)
}

func (propdb *PropertyDb) RetrieveMediatorSchl() (*modules.MediatorSchedule, error) {
	return RetrieveMediatorSchl(propdb.db)
}

func (db *PropertyDb) SetLastStableUnit(hash common.Hash, index *modules.ChainIndex) error {
	data := &modules.UnitProperty{hash, index, 0}
	key := append(constants.LastStableUnitHash, index.AssetID.Bytes()...)
	log.Debugf("Save last stable unit %s,index:%s", hash.String(), index.String())
	return StoreBytes(db.db, key, data)
}
func (db *PropertyDb) GetLastStableUnit(asset modules.IDType16) (common.Hash, *modules.ChainIndex, error) {
	key := append(constants.LastStableUnitHash, asset.Bytes()...)
	data := &modules.UnitProperty{}
	err := retrieve(db.db, key, data)
	if err != nil {
		return common.Hash{}, nil, err
	}
	return data.Hash, data.Index, nil
}
func (db *PropertyDb) SetNewestUnit(header *modules.Header) error {
	hash := header.Hash()
	index := header.Number
	timestamp := header.Creationdate
	data := &modules.UnitProperty{hash, index, timestamp}
	key := append(constants.LastUnstableUnitHash, index.AssetID.Bytes()...)
	log.Debugf("Save newest unit %s,index:%s", hash.String(), index.String())

	return StoreBytes(db.db, key, data)
}
func (db *PropertyDb) GetNewestUnit(asset modules.IDType16) (common.Hash, *modules.ChainIndex, int64, error) {
	key := append(constants.LastUnstableUnitHash, asset.Bytes()...)
	data := &modules.UnitProperty{}
	err := retrieve(db.db, key, data)
	if err != nil {
		return common.Hash{}, nil, 0, err
	}
	return data.Hash, data.Index, data.Timestamp, nil
}
