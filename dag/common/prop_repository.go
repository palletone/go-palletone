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

package common

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type PropRepository struct {
	db storage.IPropertyDb
}
type IPropRepository interface {
	StoreGlobalProp(gp *modules.GlobalProperty) error
	RetrieveGlobalProp() (*modules.GlobalProperty, error)
	StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error
	RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error)
	StoreMediatorSchl(ms *modules.MediatorSchedule) error
	RetrieveMediatorSchl() (*modules.MediatorSchedule, error)

	SetLastStableUnit(hash common.Hash, index *modules.ChainIndex) error
	GetLastStableUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, error)
	SetNewestUnit(header *modules.Header) error
	GetNewestUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, error)
}

func NewPropRepository(db storage.IPropertyDb) *PropRepository {
	return &PropRepository{db: db}
}
func (pRep *PropRepository) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	return pRep.db.RetrieveGlobalProp()
}
func (pRep *PropRepository) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	return pRep.db.RetrieveDynGlobalProp()
}
func (pRep *PropRepository) StoreGlobalProp(gp *modules.GlobalProperty) error {
	return pRep.db.StoreGlobalProp(gp)
}
func (pRep *PropRepository) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {
	return pRep.db.StoreDynGlobalProp(dgp)
}
func (pRep *PropRepository) StoreMediatorSchl(ms *modules.MediatorSchedule) error {
	return pRep.db.StoreMediatorSchl(ms)
}
func (pRep *PropRepository) RetrieveMediatorSchl() (*modules.MediatorSchedule, error) {
	return pRep.db.RetrieveMediatorSchl()
}
func (pRep *PropRepository) SetLastStableUnit(hash common.Hash, index *modules.ChainIndex) error {
	return pRep.db.SetLastStableUnit(hash, index)
}
func (pRep *PropRepository) GetLastStableUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, error) {
	return pRep.db.GetLastStableUnit(token)
}
func (pRep *PropRepository) SetNewestUnit(header *modules.Header) error {
	return pRep.db.SetNewestUnit(header)
}
func (pRep *PropRepository) GetNewestUnit(token modules.IDType16) (common.Hash, *modules.ChainIndex, error) {
	return pRep.db.GetNewestUnit(token)
}

/**
mediator投票结果，返回区块高度
Method for getting mediator voting results
*/

//var lastStatisticalHeight = GenesisHeight()

//func MediatorVoteResult(db ptndb.Database,height modules.ChainIndex) (map[common.Address]uint64, error) {
//	var lastStatisticalHeight = GenesisHeight(db)
//	result := map[common.Address]uint64{}
//	// step1. check height
//	// check asset id
//	if strings.Compare(lastStatisticalHeight.AssetID.String(), height.AssetID.String()) != 0 {
//		return nil, fmt.Errorf("Mediator for different token comparing with last statistcal height.")
//	}
//	// check is main
//	if height.IsMain == false {
//		return nil, fmt.Errorf("Height must be the main height")
//	}
//	// step2. query vote db to get result
//	// step3. set lastStatisticalHeight
//	lastStatisticalHeight.AssetID.SetBytes(height.AssetID.Bytes())
//	lastStatisticalHeight.IsMain = height.IsMain
//	lastStatisticalHeight.Index = height.Index
//	return result, nil
//}
