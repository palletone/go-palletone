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
	"reflect"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

type PropertyDb struct {
	db ptndb.Database
}
type IPropertyDb interface {
	StoreGlobalProp(gp *modules.GlobalProperty) error
	RetrieveGlobalProp() (*modules.GlobalProperty, error)
	StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error
	RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error)
	StoreMediatorSchl(ms *modules.MediatorSchedule) error
	RetrieveMediatorSchl() (*modules.MediatorSchedule, error)

	//StoreGlobalPropHistory(gp *modules.GlobalPropertyHistory) error
	//RetrieveGlobalPropHistories() ([]*modules.GlobalPropertyHistory, error)

	SetNewestUnit(header *modules.Header) error
	//GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, int64, error)
	GetNewestUnit(token modules.AssetId) (*modules.UnitProperty, error)

	GetChainParameters() *core.ChainParameters
}

func (propdb *PropertyDb) GetChainParameters() *core.ChainParameters {
	gp, err := propdb.RetrieveGlobalProp()
	if err != nil {
		return nil
	}

	return &gp.ChainParameters
}

// modified by Yiran
// initialize PropertyDB , and retrieve gp,dgp,mc from IPropertyDb.
func NewPropertyDb(db ptndb.Database) *PropertyDb {
	pdb := &PropertyDb{db: db}
	return pdb
}

func (propdb *PropertyDb) StoreMediatorSchl(ms *modules.MediatorSchedule) error {
	err := StoreToRlpBytes(propdb.db, constants.MEDIATOR_SCHEDULE_KEY, ms)
	if err != nil {
		log.Errorf("Store mediator schedule error: %v", err.Error())
	}

	return err
}

func (propdb *PropertyDb) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {
	log.Debugf("DB[%s] Save dynamic global property to db.", reflect.TypeOf(propdb.db).String())
	err := StoreToRlpBytes(propdb.db, constants.DYNAMIC_GLOBALPROPERTY_KEY, dgp)
	if err != nil {
		log.Debugf("Store dynamic global properties error: %v", err.Error())
	}

	return err
}

func (propdb *PropertyDb) StoreGlobalProp(gp *modules.GlobalProperty) error {
	log.Debugf("DB[%s] Save global property to db.", reflect.TypeOf(propdb.db).String())
	err := StoreToRlpBytes(propdb.db, constants.GLOBALPROPERTY_KEY, gp)
	if err != nil {
		log.Debugf("Store global properties error: %v", err.Error())
	}

	return err
}

func (propdb *PropertyDb) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	gp := modules.NewGlobalProp()
	err := RetrieveFromRlpBytes(propdb.db, constants.GLOBALPROPERTY_KEY, gp)
	if err != nil {
		log.Errorf("Retrieve global properties error: %v", err.Error())
	}

	return gp, err
}

func (propdb *PropertyDb) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	dgp := modules.NewDynGlobalProp()
	err := RetrieveFromRlpBytes(propdb.db, constants.DYNAMIC_GLOBALPROPERTY_KEY, dgp)
	if err != nil {
		log.Errorf("Retrieve dynamic global properties error: %v", err.Error())
	}

	return dgp, err
}

func (propdb *PropertyDb) RetrieveMediatorSchl() (*modules.MediatorSchedule, error) {
	ms := modules.NewMediatorSchl()
	err := RetrieveFromRlpBytes(propdb.db, constants.MEDIATOR_SCHEDULE_KEY, ms)
	if err != nil {
		log.Errorf("Retrieve mediator schedule error: %v", err.Error())
	}

	return ms, err
}

//func makeGlobalPropHistoryKey(gp *modules.GlobalPropertyHistory) []byte {
//	b := make([]byte, 8)
//	binary.LittleEndian.PutUint64(b, gp.EffectiveTime)
//	return append(constants.GLOBAL_PROPERTY_HISTORY_PREFIX, b...)
//}
//
//func (propdb *PropertyDb) StoreGlobalPropHistory(gp *modules.GlobalPropertyHistory) error {
//	log.Debugf("DB[%s] Save global property history to db.", reflect.TypeOf(propdb.db).String())
//	key := makeGlobalPropHistoryKey(gp)
//	err := StoreToRlpBytes(propdb.db, key, gp)
//	if err != nil {
//		log.Debugf("Store global properties history error: %v", err.Error())
//	}
//
//	return err
//}
//
//func (propdb *PropertyDb) RetrieveGlobalPropHistories() ([]*modules.GlobalPropertyHistory, error) {
//	kv := getprefix(propdb.db, constants.GLOBAL_PROPERTY_HISTORY_PREFIX)
//	result := make([]*modules.GlobalPropertyHistory, 0)
//	for _, v := range kv {
//		gp := &modules.GlobalPropertyHistory{}
//		rlp.DecodeBytes(v, gp)
//		result = append(result, gp)
//	}
//	return result, nil
//}

func (db *PropertyDb) SetNewestUnit(header *modules.Header) error {
	hash := header.Hash()
	index := header.GetNumber()
	timestamp := uint32(header.Timestamp())

	data := &modules.UnitProperty{Hash: hash, ChainIndex: index, Timestamp: timestamp}
	//key := append(constants.LastUnitInfo, index.AssetID.Bytes()...)
	log.Debugf("DB[%s]Save newest unit %s,index:%s", reflect.TypeOf(db.db).String(), hash.String(), index.String())

	//return StoreToRlpBytes(db.db, key, data)
	return StoreToRlpBytes(db.db, getNewestUnit(index.AssetID), data)
}

func getNewestUnit(asset modules.AssetId) []byte {
	return append(constants.LastUnitInfo, asset.Bytes()...)
}

//func (db *PropertyDb) GetNewestUnit(asset modules.AssetId) (common.Hash, *modules.ChainIndex, int64, error) {
func (db *PropertyDb) GetNewestUnit(asset modules.AssetId) (*modules.UnitProperty, error) {
	data := &modules.UnitProperty{}
	//key := append(constants.LastUnitInfo, asset.Bytes()...)
	//err := RetrieveFromRlpBytes(db.db, key, data)

	err := RetrieveFromRlpBytes(db.db, getNewestUnit(asset), data)
	if err != nil {
		//return common.Hash{}, nil, 0, err
		return nil, err
	}

	log.Debugf("DB[%s] GetNewestUnit: %s,Index:%s,timestamp:%d", reflect.TypeOf(db.db).String(),
		data.Hash.String(), data.ChainIndex.String(), data.Timestamp)
	//return data.Hash, data.Index, int64(data.Timestamp), nil
	return data, nil
}
