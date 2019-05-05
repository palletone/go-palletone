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
	"github.com/palletone/go-palletone/core/accounts/keystore"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

// GenerateUnit, generate unit
// @author Albert·Gou
func (dag *Dag) GenerateUnit(when time.Time, producer common.Address, groupPubKey []byte,
	ks *keystore.KeyStore, txpool txspool.ITxPool) *modules.Unit {
	t0 := time.Now()
	defer func(start time.Time) {
		log.Debugf("GenerateUnit cost time: %v", time.Since(start))
	}(t0)

	// 1. 判断是否满足生产的若干条件

	// 2. 生产unit，添加交易集、时间戳、签名
	newUnit, err := dag.CreateUnit(&producer, txpool, when)
	if err != nil {
		log.Debug("GenerateUnit", "error", err.Error())
		return nil
	}
	// added by yangyu, 2018.8.9
	if newUnit == nil || newUnit.IsEmpty() {
		log.Info("No unit need to be packaged for now.", "unit", newUnit)
		return nil
	}

	newUnit.UnitHeader.Time = when.Unix()
	newUnit.UnitHeader.ParentsHash[0] = dag.HeadUnitHash()
	newUnit.UnitHeader.Number.Index = dag.HeadUnitNum() + 1
	newUnit.UnitHeader.GroupPubKey = groupPubKey
	newUnit.Hash()

	sign_unit, err1 := dagcommon.GetUnitWithSig(newUnit, ks, producer)
	if err1 != nil {
		log.Debugf("GetUnitWithSig error: %v", err)
		return nil
	}

	sign_unit.UnitSize = sign_unit.Size()
	log.Debugf("Generate new unit index:[%d],hash:[%s],size:%s, parent unit[%s], spent time: %s",
		sign_unit.NumberU64(), sign_unit.Hash().String(), sign_unit.UnitSize.String(),
		newUnit.UnitHeader.ParentsHash[0].String(), time.Since(t0).String())

	//TODO add PostChainEvents
	go func() {
		var (
			events = make([]interface{}, 0, 1)
		)
		events = append(events, modules.ChainHeadEvent{newUnit})
		dag.PostChainEvents(events)
	}()

	if !dag.PushUnit(sign_unit, txpool) {
		return nil
	}

	return sign_unit
}

/**
 * Push unit "may fail" in which case every partial change is unwound.  After
 * push unit is successful the block is appended to the chain database on disk.
 *
 * 推块“可能会失败”，在这种情况下，每个部分地更改都会撤销。 推块成功后，该块将附加到磁盘上的链数据库。
 *
 * @return true if we switched forks as a result of this push.
 */
func (dag *Dag) PushUnit(newUnit *modules.Unit, txpool txspool.ITxPool) bool {
	t0 := time.Now()
	// 1. 如果当前初生产的unit不在最长链条上，那么就切换到最长链分叉上。

	// 2. 更新状态
	if err := dag.ApplyUnit(newUnit); err != nil {
		return false
	}

	dag.Memdag.AddUnit(newUnit, txpool)
	log.Debugf("save newest unit spent time: %s, index: %d , hash:%s", time.Since(t0).String(), newUnit.NumberU64(), newUnit.UnitHash.String())
	return true
}

// ApplyUnit, 运用下一个 unit 更新整个区块链状态
func (dag *Dag) ApplyUnit(nextUnit *modules.Unit) error {
	defer func(start time.Time) {
		log.Debugf("ApplyUnit cost time: %v", time.Since(start))
	}(time.Now())

	dag.applyLock.Lock()
	defer dag.applyLock.Unlock()

	// 下一个 unit 和本地 unit 连续性的判断
	if err := dag.validateUnitHeader(nextUnit); err != nil {
		return err
	}

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
	dag.updateLastIrreversibleUnit()

	// 判断是否到了链维护周期，并维护
	maintenanceNeeded := !(dag.GetDynGlobalProp().NextMaintenanceTime > uint32(nextUnit.Timestamp()))
	if maintenanceNeeded {
		dag.performChainMaintenance(nextUnit)
	}

	// 更新链维护周期标志
	// n.b., updateMaintenanceFlag() happens this late because GetSlotTime() / GetSlotAtTime() is needed above
	// 由于前面的操作需要调用 GetSlotTime() / GetSlotAtTime() 这两个方法，所以在最后才更新链维护周期标志
	dag.updateMaintenanceFlag(maintenanceNeeded)

	if !skip {
		// 洗牌
		dag.updateMediatorSchedule()
	}

	return nil
}
