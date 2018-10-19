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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type PropRepository struct {
	db     storage.IPropertyDb
	logger log.ILogger
}
type IPropRepository interface {
	UpdateGlobalDynProp(unit *modules.Unit)
	UpdateMediatorSchedule()
}

func NewPropRepository(db storage.IPropertyDb, l log.ILogger) *PropRepository {
	return &PropRepository{db: db, logger: l}
}

func (rep *PropRepository) UpdateGlobalDynProp(unit *modules.Unit) {
	gp, _ := rep.db.RetrieveGlobalProp()
	dgp, _ := rep.db.RetrieveDynGlobalProp()

	dgp.UpdateGlobalDynProp(gp, unit)
	rep.db.StoreDynGlobalProp(dgp)

	return
}

func (rep *PropRepository) UpdateMediatorSchedule() {
	gp, _ := rep.db.RetrieveGlobalProp()
	dgp, _ := rep.db.RetrieveDynGlobalProp()
	ms, _ := rep.db.RetrieveMediatorSchl()

	ms.UpdateMediatorSchedule(gp, dgp)
	rep.db.StoreMediatorSchl(ms)

	return
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
