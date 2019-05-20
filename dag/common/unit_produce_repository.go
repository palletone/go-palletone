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
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

type IUnitProduceRepository interface {
	PushUnit(nextUnit *modules.Unit) error
	ApplyUnit(nextUnit *modules.Unit) error
}
type UnitProduceRepository struct {
	unitRep  IUnitRepository
	propRep  IPropRepository
	stateRep IStateRepository
}

func NewUnitProduceRepository(unitRep IUnitRepository, propRep IPropRepository, stateRep IStateRepository) *UnitProduceRepository {
	return &UnitProduceRepository{
		unitRep:  unitRep,
		propRep:  propRep,
		stateRep: stateRep,
	}
}

/**
 * Push unit "may fail" in which case every partial change is unwound.  After
 * push unit is successful the block is appended to the chain database on disk.
 *
 * 推块“可能会失败”，在这种情况下，每个部分地更改都会撤销。 推块成功后，该块将附加到磁盘上的链数据库。
 *
 * @return true if we switched forks as a result of this push.
 */
func (dag *UnitProduceRepository) PushUnit(newUnit *modules.Unit) error {

	// 2. 更新状态
	if err := dag.ApplyUnit(newUnit); err != nil {
		return err
	}
	//更新数据库
	return dag.unitRep.SaveUnit(newUnit, false)

}

// ApplyUnit, 运用下一个 unit 更新整个区块链状态
func (dag *UnitProduceRepository) ApplyUnit(nextUnit *modules.Unit) error {
	defer func(start time.Time) {
		log.Debugf("ApplyUnit cost time: %v", time.Since(start))
	}(time.Now())

	// todo 待删除 处理临时prop没有回滚的问题
	skip := false
	// 验证 unit 的 mediator 调度
	if err := dag.validateMediatorSchedule(nextUnit); err != nil {
		//return err
		skip = true
	}

	// todo 运用Unit中的交易

	// 计算当前 unit 到上一个 unit 之间的缺失数量，并更新每个mediator的unit的缺失数量
	missed := dag.updateMediatorMissedUnits(nextUnit)

	// 更新全局动态属性值
	dag.updateDynGlobalProp(nextUnit, missed)

	// 更新 mediator 的相关数据
	dag.updateSigningMediator(nextUnit)

	// 更新最新不可逆区块高度
	//dag.updateLastIrreversibleUnit()

	// 判断是否到了链维护周期，并维护
	//maintenanceNeeded := !(dag.GetDynGlobalProp().NextMaintenanceTime > uint32(nextUnit.Timestamp()))
	//if maintenanceNeeded {
	//	dag.performChainMaintenance(nextUnit)
	//}
	//
	//// 更新链维护周期标志
	//// n.b., updateMaintenanceFlag() happens this late because GetSlotTime() / GetSlotAtTime() is needed above
	//// 由于前面的操作需要调用 GetSlotTime() / GetSlotAtTime() 这两个方法，所以在最后才更新链维护周期标志
	//dag.updateMaintenanceFlag(maintenanceNeeded)

	if !skip {
		// 洗牌
		dag.updateMediatorSchedule()
	}

	return nil
}
func (dag *UnitProduceRepository) validateMediatorSchedule(nextUnit *modules.Unit) error {
	gasToken := dagconfig.DagConfig.GetGasToken()
	ts, _ := dag.propRep.GetNewestUnitTimestamp(gasToken)
	if ts >= nextUnit.Timestamp() {
		errStr := "invalidated unit's timestamp"
		log.Warnf("%s,db newest unit timestamp=%d,current unit[%s] timestamp=%d", errStr, ts, nextUnit.Hash().String(), nextUnit.Timestamp())
		return fmt.Errorf(errStr)
	}

	slotNum := dag.propRep.GetSlotAtTime(time.Unix(nextUnit.Timestamp(), 0))
	if slotNum <= 0 {
		errStr := "invalidated unit's slot"
		log.Debugf(errStr)
		return fmt.Errorf(errStr)
	}

	scheduledMediator := dag.propRep.GetScheduledMediator(slotNum)
	if !scheduledMediator.Equal(nextUnit.Author()) {
		errStr := fmt.Sprintf("mediator(%v) produced unit at wrong time", nextUnit.Author().Str())
		log.Debugf(errStr)
		return fmt.Errorf(errStr)
	}

	return nil
}

// 根据最新 unit 计算出生产该 unit 的 mediator 缺失的 unit 个数，
// 并更新到 mediator的相应字段中，返回数量
func (dag *UnitProduceRepository) updateMediatorMissedUnits(unit *modules.Unit) uint64 {
	missedUnits := dag.propRep.GetSlotAtTime(time.Unix(unit.Timestamp(), 0))
	if missedUnits == 0 {
		log.Errorf("Trying to push double-produced unit onto current unit?!")
		return 0
	}

	missedUnits--
	log.Debugf("the count of missed units: %v", missedUnits)

	aSize := dag.GetGlobalProp().ActiveMediatorsCount()
	if missedUnits < uint32(aSize) {
		var i uint32
		for i = 0; i < missedUnits; i++ {
			mediatorMissed := dag.propRep.GetScheduledMediator(i + 1)

			med := dag.GetMediator(mediatorMissed)
			med.TotalMissed++
			dag.SaveMediator(med, false)
		}
	}

	return uint64(missedUnits)
}

// UpdateDynGlobalProp, update global dynamic data
func (dag *UnitProduceRepository) updateDynGlobalProp(unit *modules.Unit, missedUnits uint64) {
	log.Debugf("update global dynamic data")
	dgp := dag.GetDynGlobalProp()

	//dgp.HeadUnitNum = unit.NumberU64()
	//dgp.HeadUnitHash = unit.Hash()
	//dgp.HeadUnitTime = unit.Timestamp()
	dag.propRep.SetNewestUnit(unit.Header())

	dgp.LastMediator = unit.Author()
	dgp.IsShuffledSchedule = false
	dgp.RecentSlotsFilled = (dgp.RecentSlotsFilled << (missedUnits + 1)) + 1
	dgp.CurrentASlot += missedUnits + 1

	dag.SaveDynGlobalProp(dgp, false)

	return
}

func (dag *UnitProduceRepository) updateMediatorSchedule() {
	gp := dag.GetGlobalProp()
	dgp := dag.GetDynGlobalProp()
	ms := dag.GetMediatorSchl()

	if dag.propRep.UpdateMediatorSchedule(ms, gp, dgp) {
		log.Debugf("shuffle the scheduling order of mediators")
		dag.SaveMediatorSchl(ms, false)

		dgp.IsShuffledSchedule = true
		dag.SaveDynGlobalProp(dgp, false)
	}

	return
}

func (dag *UnitProduceRepository) updateSigningMediator(newUnit *modules.Unit) {
	// 1. 更新 签名mediator 的LastConfirmedUnitNum
	signingMediator := newUnit.Author()
	med := dag.GetMediator(signingMediator)

	lastConfirmedUnitNum := uint32(newUnit.NumberU64())
	med.LastConfirmedUnitNum = lastConfirmedUnitNum
	dag.SaveMediator(med, false)

	log.Debugf("the LastConfirmedUnitNum of mediator(%v) is: %v", med.Address.Str(), lastConfirmedUnitNum)
}

func (d *UnitProduceRepository) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := d.propRep.RetrieveGlobalProp()
	return gp
}

func (d *UnitProduceRepository) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp, _ := d.propRep.RetrieveDynGlobalProp()
	return dgp
}

func (d *UnitProduceRepository) GetMediatorSchl() *modules.MediatorSchedule {
	ms, _ := d.propRep.RetrieveMediatorSchl()
	return ms
}

func (d *UnitProduceRepository) SaveGlobalProp(gp *modules.GlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propRep.StoreGlobalProp(gp)
	return
}

func (d *UnitProduceRepository) SaveDynGlobalProp(dgp *modules.DynamicGlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propRep.StoreDynGlobalProp(dgp)
	return
}

func (d *UnitProduceRepository) SaveMediatorSchl(ms *modules.MediatorSchedule, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propRep.StoreMediatorSchl(ms)
	return
}
func (d *UnitProduceRepository) GetMediator(add common.Address) *core.Mediator {
	med, err := d.stateRep.RetrieveMediator(add)
	if err != nil {
		log.Error("dag", "GetMediator RetrieveMediator err:", err, "address:", add)
		return nil
	}
	return med
}

func (d *UnitProduceRepository) SaveMediator(med *core.Mediator, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.stateRep.StoreMediator(med)
	return
}
