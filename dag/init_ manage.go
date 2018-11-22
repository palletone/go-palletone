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

package dag

import (
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

// todo 待被调用
func (dag *Dag) validateMediatorSchedule(nextUnit *modules.Unit) bool {
	if dag.HeadUnitHash() != nextUnit.ParentHash()[0] {
		log.Error("invalidated unit's parent hash!")
		return false
	}

	if dag.HeadUnitTime() >= nextUnit.Timestamp() {
		log.Error("invalidated unit's timestamp!")
		return false
	}

	slotNum := dag.GetSlotAtTime(time.Unix(nextUnit.Timestamp(), 0))
	if slotNum <= 0 {
		log.Error("invalidated unit's slot!")
		return false
	}

	scheduledMediator := dag.GetScheduledMediator(slotNum)
	if scheduledMediator.Equal(nextUnit.UnitAuthor()) {
		log.Error("Mediator produced unit at wrong time!")
		return false
	}

	return true
}

// @author Albert·Gou
func (d *Dag) ValidateUnitExceptGroupSig(unit *modules.Unit, isGenesis bool) bool {
	unitState := d.validate.ValidateUnitExceptGroupSig(unit, isGenesis)
	if unitState != modules.UNIT_STATE_VALIDATED &&
		unitState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
		return false
	}
	return true
}

// author Albert·Gou
func (d *Dag) IsActiveMediator(add common.Address) bool {
	return d.GetGlobalProp().IsActiveMediator(add)
}

func (dag *Dag) InitPropertyDB(genesis *core.Genesis, genesisUnitHash common.Hash) error {
	//  全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	gp := modules.InitGlobalProp(genesis)
	if err := dag.propdb.StoreGlobalProp(gp); err != nil {
		return err
	}

	//  动态全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	dgp := modules.InitDynGlobalProp(genesis, genesisUnitHash)
	if err := dag.propdb.StoreDynGlobalProp(dgp); err != nil {
		return err
	}

	//  初始化mediator调度器，并存在数据库
	// @author Albert·Gou
	ms := modules.InitMediatorSchl(gp, dgp)
	if err := dag.propdb.StoreMediatorSchl(ms); err != nil {
		return err
	}

	return nil
}

func (dag *Dag) IsSynced() bool {
	gp := dag.GetGlobalProp()
	dgp := dag.GetDynGlobalProp()

	nowFine := time.Now()
	now := time.Unix(nowFine.Add(500*time.Millisecond).Unix(), 0)
	nextSlotTime := modules.GetSlotTime(gp, dgp, 1)

	if nextSlotTime.Before(now) {
		return false
	}

	return true
}
