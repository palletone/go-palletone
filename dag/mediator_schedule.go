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

	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

func (d *Dag) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := d.propdb.RetrieveGlobalProp()
	return gp
}

func (d *Dag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp, _ := d.propdb.RetrieveDynGlobalProp()
	return dgp
}

func (d *Dag) GetMediatorSchl() *modules.MediatorSchedule {
	ms, _ := d.propdb.RetrieveMediatorSchl()
	return ms
}

func (d *Dag) SaveGlobalProp(gp *modules.GlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propdb.StoreGlobalProp(gp)
	return
}

func (d *Dag) SaveDynGlobalProp(dgp *modules.DynamicGlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propdb.StoreDynGlobalProp(dgp)
	return
}

func (d *Dag) SaveMediatorSchl(ms *modules.MediatorSchedule, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propdb.StoreMediatorSchl(ms)
	return
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
func (d *Dag) GetActiveMediatorNodes() map[string]*discover.Node {
	return d.GetGlobalProp().GetActiveMediatorNodes()
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorInitPubs() []kyber.Point {
	return d.GetGlobalProp().GetActiveMediatorInitPubs()
}

// author Albert·Gou
func (d *Dag) GetCurThreshold() int {
	return d.GetGlobalProp().GetCurThreshold()
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorCount() int {
	return d.GetGlobalProp().GetActiveMediatorCount()
}

// author Albert·Gou
func (d *Dag) GetActiveMediators() []common.Address {
	return d.GetGlobalProp().GetActiveMediators()
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorAddr(index int) common.Address {
	return d.GetGlobalProp().GetActiveMediatorAddr(index)
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorNode(index int) *discover.Node {
	return d.GetGlobalProp().GetActiveMediatorNode(index)
}

// author Albert·Gou
func (d *Dag) GetActiveMediator(add common.Address) *core.Mediator {
	return d.GetGlobalProp().GetActiveMediator(add)
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

func (dag *Dag) GetSlotAtTime(when time.Time) uint32 {
	return modules.GetSlotAtTime(dag.GetGlobalProp(), dag.GetDynGlobalProp(), when)
}

func (dag *Dag) GetSlotTime(slotNum uint32) time.Time {
	return modules.GetSlotTime(dag.GetGlobalProp(), dag.GetDynGlobalProp(), slotNum)
}

func (dag *Dag) GetScheduledMediator(slotNum uint32) common.Address {
	return dag.GetMediatorSchl().GetScheduledMediator(dag.GetDynGlobalProp(), slotNum)
}

func (dag *Dag) HeadUnitTime() int64 {
	return dag.GetDynGlobalProp().HeadUnitTime
}

func (dag *Dag) HeadUnitNum() uint64 {
	return dag.GetDynGlobalProp().HeadUnitNum
}

func (dag *Dag) HeadUnitHash() common.Hash {
	return dag.GetDynGlobalProp().HeadUnitHash
}

// 根据最新 unit 计算出生产该 unit 的 mediator 缺失的 unit 个数，
// 并更新到 mediator的相应字段中，返回数量
func (dag *Dag) UpdateMediatorMissedUnits(unit *modules.Unit) uint64 {
	timestamp := unit.UnitHeader.Creationdate
	missedUnits := dag.GetSlotAtTime(time.Unix(timestamp, 0))
	// todo

	return uint64(missedUnits)
}

func (dag *Dag) UpdateDynGlobalProp(unit *modules.Unit, missedUnits uint64) {
	gp := dag.GetGlobalProp()
	dgp := dag.GetDynGlobalProp()

	dgp.UpdateDynGlobalProp(gp, unit, missedUnits)
	dag.SaveDynGlobalProp(dgp, false)

	return
}

func (dag *Dag) UpdateMediatorSchedule() {
	gp := dag.GetGlobalProp()
	dgp := dag.GetDynGlobalProp()
	ms := dag.GetMediatorSchl()

	ms.UpdateMediatorSchedule(gp, dgp)
	dag.SaveMediatorSchl(ms, false)

	return
}

func (dag *Dag) GetMediators() map[common.Address]bool {
	return dag.statedb.GetMediators()
}

func (dag *Dag) MediatorSchedule() []common.Address {
	return dag.GetMediatorSchl().CurrentShuffledMediators
}

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
	if nextUnit.UnitAuthor().Equal(scheduledMediator) {
		log.Error("Mediator produced unit at wrong time!")
		return false
	}

	return true
}
