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
	"fmt"
	"sort"
	"time"

	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"

	"github.com/palletone/go-palletone/dag/modules"
)

// 根据最新 unit 计算出生产该 unit 的 mediator 缺失的 unit 个数，
// 并更新到 mediator的相应字段中，返回数量
func (dag *Dag) updateMediatorMissedUnits(unit *modules.Unit) uint64 {
	missedUnits := dag.GetSlotAtTime(time.Unix(unit.Timestamp(), 0))
	if missedUnits == 0 {
		log.Debug("Trying to push double-produced unit onto current unit?!")
		return 0
	}

	missedUnits--
	log.Debug(fmt.Sprintf("the count of missed units: %v", missedUnits))

	aSize := dag.ActiveMediatorsCount()
	if missedUnits < uint32(aSize) {
		var i uint32
		for i = 0; i < missedUnits; i++ {
			mediatorMissed := dag.GetScheduledMediator(i + 1)

			med := dag.GetMediator(mediatorMissed)
			med.TotalMissed++
			dag.SaveMediator(med, false)
		}
	}

	return uint64(missedUnits)
}

func (dag *Dag) updateDynGlobalProp(unit *modules.Unit, missedUnits uint64) {
	dgp := dag.GetDynGlobalProp()

	dgp.UpdateDynGlobalProp(unit, missedUnits)
	dag.SaveDynGlobalProp(dgp, false)

	return
}

func (dag *Dag) updateMediatorSchedule() {
	gp := dag.GetGlobalProp()
	dgp := dag.GetDynGlobalProp()
	ms := dag.GetMediatorSchl()

	if dag.propRep.UpdateMediatorSchedule(ms, gp, dgp) {
		dag.SaveMediatorSchl(ms, false)
	}

	return
}

func (dag *Dag) updateSigningMediator(newUnit *modules.Unit) {
	// 1. 更新 签名mediator 的LastConfirmedUnitNum
	signingMediator := newUnit.Author()
	med := dag.GetMediator(signingMediator)

	med.LastConfirmedUnitNum = uint32(newUnit.NumberU64())
	dag.SaveMediator(med, false)
}

func (dag *Dag) updateLastIrreversibleUnit() {
	aSize := dag.ActiveMediatorsCount()
	lastConfirmedUnitNums := make([]int, 0, aSize)

	// 1. 获取所有活跃 mediator 最后确认unit编号
	meds := dag.GetActiveMediators()
	for _, add := range meds {
		med := dag.GetActiveMediator(add)
		lastConfirmedUnitNums = append(lastConfirmedUnitNums, int(med.LastConfirmedUnitNum))
	}

	// 2. 排序
	// todo 应当优化本排序方法，使用第n大元素的方法
	sort.Ints(lastConfirmedUnitNums)

	// 3. 获取倒数第 > 2/3 个确认unit编号
	//offset := aSize - dag.ChainThreshold()
	//var newLastIrreversibleUnitNum = uint64(lastConfirmedUnitNums[offset])
	//TODO Devin where is unit hash?
	// 4. 更新
	//dag.updateLastIrreversibleUnitNum( newLastIrreversibleUnitNum)
}

//func (dag *Dag) updateLastIrreversibleUnitNum(hash common.Hash, newLastIrreversibleUnitNum uint64) {
//	//dgp := dag.GetDynGlobalProp()
//	token := node.DefaultConfig.GetGasToken()
//	_, index, _ := dag.propRep.GetLastStableUnit(token)
//	if newLastIrreversibleUnitNum > index.Index {
//		dag.propRep.SetLastStableUnit(hash, &modules.ChainIndex{token, true, newLastIrreversibleUnitNum})
//		//dgp.s = newLastIrreversibleUnitNum
//		//dag.SaveDynGlobalProp(dgp, false)
//	}
//}

//func (dag *Dag) updateGlobalPropDependGroupSign(unitHash common.Hash) {
//	unit, err := dag.GetUnitByHash(unitHash)
//	if err != nil {
//		log.Debug(err.Error())
//
// 	return
//	}
//
//	// 1. 根据群签名更新不可逆unit高度
//	//dag.updateLastIrreversibleUnitNum(unitHash, uint64(unit.NumberU64()))
//}

// 活跃 mediators 更新事件
type ActiveMediatorsUpdatedEvent struct {
	IsChanged bool // 标记活跃 mediators 是否有改变
}

func (dag *Dag) SubscribeActiveMediatorsUpdatedEvent(ch chan<- ActiveMediatorsUpdatedEvent) event.Subscription {
	return dag.activeMediatorsUpdatedScope.Track(dag.activeMediatorsUpdatedFeed.Subscribe(ch))
}

func (dag *Dag) performChainMaintenance(nextUnit *modules.Unit) {
	dgp := dag.GetDynGlobalProp()

	// 1. 判断是否进入维护周期. Are we at the maintenance interval?
	if dgp.NextMaintenanceTime > nextUnit.Timestamp() {
		return
	}

	// 2. 对每个账户的各种投票信息进行初步统计
	dag.performAccountMaintenance()

	// 3. 统计投票并更新活跃 mediator 列表
	isChanged := dag.updateActiveMediators()

	// 4. 发送更新活跃 mediator 事件，以方便其他模块做相应处理
	go dag.activeMediatorsUpdatedFeed.Send(ActiveMediatorsUpdatedEvent{IsChanged: isChanged})

	// 5. 计算并更新下一次维护时间
	gp := dag.GetGlobalProp()
	nextMaintenanceTime := dgp.NextMaintenanceTime
	maintenanceInterval := int64(gp.ChainParameters.MaintenanceInterval)
	if nextUnit.NumberU64() == 1 {
		nextMaintenanceTime = (nextUnit.Timestamp()/maintenanceInterval + 1) * maintenanceInterval
	} else {
		// We want to find the smallest k such that nextMaintenanceTime + k * maintenanceInterval > HeadUnitTime()
		//  This implies k > ( HeadUnitTime() - nextMaintenanceTime ) / maintenanceInterval
		//
		// Let y be the right-hand side of this inequality, i.e.
		// y = ( HeadUnitTime() - nextMaintenanceTime ) / maintenanceInterval
		//
		// and let the fractional part f be y-floor(y).  Clearly 0 <= f < 1.
		// We can rewrite f = y-floor(y) as floor(y) = y-f.
		//
		// Clearly k = floor(y)+1 has k > y as desired.  Now we must
		// show that this is the least such k, i.e. k-1 <= y.
		//
		// But k-1 = floor(y)+1-1 = floor(y) = y-f <= y.
		// So this k suffices.
		//

		y := (dag.HeadUnitTime() - nextMaintenanceTime) / maintenanceInterval
		nextMaintenanceTime += (y + 1) * maintenanceInterval
	}

	dgp.LastMaintenanceTime = dgp.NextMaintenanceTime
	dgp.NextMaintenanceTime = nextMaintenanceTime
	dag.SaveDynGlobalProp(dgp, false)

	// 6. 清理中间处理缓存数据
	dag.mediatorVoteTally = nil
}
