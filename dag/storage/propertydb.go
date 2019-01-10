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
	"github.com/palletone/go-palletone/common/ptndb"
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
	return StoreMediatorSchl(propdb.db, ms)
}

func (propdb *PropertyDb) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {
	return StoreDynGlobalProp(propdb.db, dgp)
}

func (propdb *PropertyDb) StoreGlobalProp(gp *modules.GlobalProperty) error {
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
