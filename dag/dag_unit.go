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
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	// "github.com/palletone/go-palletone/core/types"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
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

	gasToken := dagconfig.DagConfig.GetGasToken()

	// 1. 判断是否满足生产的若干条件

	//检查NewestUnit是否存在，不存在则从MemDag获取最新的Unit作为NewestUnit
	// todo 应当在其他地方其他时刻更新该值
	hash, chainIndex, _ := dag.propRep.GetNewestUnit(gasToken)
	if !dag.IsHeaderExist(hash) {
		log.Debugf("Newest unit[%s] not exist in dag, retrieve another from memdag and update NewestUnit.index [%d]", hash.String(), chainIndex.Index)
		newestUnit := dag.Memdag.GetLastMainchainUnit(gasToken)
		if nil != newestUnit {
			dag.propRep.SetNewestUnit(newestUnit.Header())
		}
	}
	log.Infof("Start generate unit... index:%d ", chainIndex.Index+1)
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
		log.Debug(fmt.Sprintf("GetUnitWithSig error: %v", err))
		return nil
	}

	sign_unit.UnitSize = sign_unit.Size()
	log.Debugf("Generate new unit[%s],size:%s, parent unit[%s], spent time: %s", sign_unit.UnitHash.String(), sign_unit.UnitSize.String(), newUnit.UnitHeader.ParentsHash[0].String(), time.Since(t0).String())

	//TODO add PostChainEvents
	// go func() {
	// 	var (
	// 		events        = make([]interface{}, 0, 1)
	// 		coalescedLogs []*types.Log
	// 	)
	// 	events = append(events, modules.ChainEvent{newUnit, common.Hash{}, nil})
	// 	dag.PostChainEvents(events, coalescedLogs)
	// }()

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
	// 1. 如果当前初生产的unit不在最长链条上，那么就切换到最长链分叉上。
	t0 := time.Now()
	// 2. 更新状态
	if !dag.ApplyUnit(newUnit) {
		return false
	}
	go dag.Memdag.AddUnit(newUnit, txpool)
	log.Debugf("save newest unit spent time: %s, parent hash:%s", time.Since(t0).String(), newUnit.ParentHash()[0].String())
	return true
}

// ApplyUnit, 运用下一个 unit 更新整个区块链状态
func (dag *Dag) ApplyUnit(nextUnit *modules.Unit) bool {
	defer func(start time.Time) {
		log.Debugf("ApplyUnit cost time: %v", time.Since(start))
	}(time.Now())

	// 1. 下一个 unit 和本地 unit 连续性的判断
	if !dag.validateUnitHeader(nextUnit) {
		return false
	}

	// 2. 验证 unit 的 mediator 调度
	if !dag.validateMediatorSchedule(nextUnit) {
		return false
	}

	// todo 5. 运用Unit中的交易

	// 3. 计算当前 unit 到上一个 unit 之间的缺失数量，并更新每个mediator的unit的缺失数量
	missed := dag.updateMediatorMissedUnits(nextUnit)

	// 4. 更新全局动态属性值
	dag.updateDynGlobalProp(nextUnit, missed)
	dag.propRep.SetNewestUnit(nextUnit.Header())

	// 5. 更新 mediator 的相关数据
	dag.updateSigningMediator(nextUnit)

	// 6. 更新最新不可逆区块高度
	dag.updateLastIrreversibleUnit()

	// 7. 判断是否到了维护周期，并维护
	dag.performChainMaintenance(nextUnit)

	// 8. 洗牌
	dag.updateMediatorSchedule()

	return true
}
