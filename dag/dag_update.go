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
	"sort"
	"time"

	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

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
	offset := aSize - dag.ChainThreshold()
	var newLastIrreversibleUnitNum = uint64(lastConfirmedUnitNums[offset])

	// 4. 更新
	dag.updateLastIrreversibleUnitNum(newLastIrreversibleUnitNum)
	log.Debugf("new last irreversible unit number is: %v", newLastIrreversibleUnitNum)
}

func (dag *Dag) updateLastIrreversibleUnitNum( /*hash common.Hash, */ newLastIrreversibleUnitNum uint64) {
	dgp := dag.GetDynGlobalProp()
	token := dagconfig.DagConfig.GetGasToken()
	_, index, _ := dag.stablePropRep.GetLastStableUnit(token)
	if newLastIrreversibleUnitNum > index.Index {
		//dag.stablePropRep.SetLastStableUnit(hash, &modules.ChainIndex{token, true, newLastIrreversibleUnitNum})
		dgp.LastIrreversibleUnitNum = newLastIrreversibleUnitNum
		dag.SaveDynGlobalProp(dgp, false)
	}
}

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
	log.Debugf("We are at the maintenance interval")

	// 更新要修改的区块链参数
	dag.updateChainParameters()

	// 对每个账户的各种投票信息进行初步统计
	dag.performAccountMaintenance()

	// 统计投票并更新活跃 mediator 列表
	isChanged := dag.updateActiveMediators()

	// 发送更新活跃 mediator 事件，以方便其他模块做相应处理
	go dag.activeMediatorsUpdatedFeed.Send(ActiveMediatorsUpdatedEvent{IsChanged: isChanged})

	// 计算并更新下一次维护时间
	dag.updateNextMaintenanceTime(nextUnit)

	// 清理中间处理缓存数据
	dag.mediatorVoteTally = nil
}

func (dag *Dag) updateNextMaintenanceTime(nextUnit *modules.Unit) {
	dgp := dag.GetDynGlobalProp()
	gp := dag.GetGlobalProp()

	nextMaintenanceTime := dgp.NextMaintenanceTime
	maintenanceInterval := int64(gp.ChainParameters.MaintenanceInterval)

	if nextUnit.NumberU64() == 1 {
		nextMaintenanceTime = uint32((nextUnit.Timestamp()/maintenanceInterval + 1) * maintenanceInterval)
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

		y := (dag.HeadUnitTime() - int64(nextMaintenanceTime)) / maintenanceInterval
		nextMaintenanceTime += uint32((y + 1) * maintenanceInterval)
	}

	dgp.LastMaintenanceTime = dgp.NextMaintenanceTime
	dgp.NextMaintenanceTime = nextMaintenanceTime
	dag.SaveDynGlobalProp(dgp, false)

	time := time.Unix(int64(nextMaintenanceTime), 0)
	log.Debugf("nextMaintenanceTime: %v", time.Format("2006-01-02 15:04:05"))

	return
}

func (dag *Dag) updateMaintenanceFlag(newMaintenanceFlag bool) {
	log.Debugf("update maintenance flag: %v", newMaintenanceFlag)

	dgp := dag.GetDynGlobalProp()
	dgp.MaintenanceFlag = newMaintenanceFlag
	dag.SaveDynGlobalProp(dgp, false)

	return
}
