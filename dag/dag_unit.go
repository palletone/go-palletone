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
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

func (dag *Dag) setUnitHeader(pendingUnit *modules.Unit) {
	current_index, _ := dag.GetCurrentChainIndex(pendingUnit.UnitHeader.ChainIndex().AssetID)

	if len(pendingUnit.UnitHeader.AssetIDs) > 0 {

		curMemUnit := dag.GetCurrentMemUnit(pendingUnit.UnitHeader.AssetIDs[0], current_index.Index)
		curUnit := dag.GetCurrentUnit(pendingUnit.UnitHeader.AssetIDs[0])

		if curMemUnit != nil {

			if curMemUnit.UnitHeader.Index() > curUnit.UnitHeader.Index() {
				pendingUnit.UnitHeader.ParentsHash = append(pendingUnit.UnitHeader.ParentsHash, curMemUnit.UnitHash)
				pendingUnit.UnitHeader.Number = curMemUnit.UnitHeader.Number
				pendingUnit.UnitHeader.Number.Index += 1
			} else {
				pendingUnit.UnitHeader.ParentsHash = append(pendingUnit.UnitHeader.ParentsHash, curUnit.UnitHash)
				pendingUnit.UnitHeader.Number = curUnit.UnitHeader.Number
				pendingUnit.UnitHeader.Number.Index += 1
			}
		} else {
			pendingUnit.UnitHeader.ParentsHash = append(pendingUnit.UnitHeader.ParentsHash, curUnit.UnitHash)
			pendingUnit.UnitHeader.Number = curUnit.UnitHeader.Number
			pendingUnit.UnitHeader.Number.Index += 1
		}

	} else {
		pendingUnit.UnitHeader.Number = *current_index
		pendingUnit.UnitHeader.Number.Index = current_index.Index + 1

		pendingUnit.UnitHeader.ParentsHash =
			append(pendingUnit.UnitHeader.ParentsHash, dag.HeadUnitHash())
	}

	if pendingUnit.UnitHeader.Number == (modules.ChainIndex{}) {
		current_index.Index += 1
		pendingUnit.UnitHeader.Number = *current_index
	} else {
		go log.Info("the pending unit header number index info. ", "index", pendingUnit.UnitHeader.Number.Index,
			"hex", pendingUnit.UnitHeader.Number.AssetID.String())
	}
}

// GenerateUnit, generate unit
// @author Albert·Gou
func (dag *Dag) GenerateUnit(when time.Time, producer common.Address, groupPubKey []byte,
	ks *keystore.KeyStore, txpool txspool.ITxPool) *modules.Unit {
	defer func(start time.Time) {
		log.Debug("GenerateUnit unit elapsed", "elapsed", time.Since(start))
	}(time.Now())
	// 1. 判断是否满足生产的若干条件

	// 2. 生产验证单元，添加交易集、时间戳、签名
	newUnits, err := dag.CreateUnit(&producer, txpool, ks, when)
	if err != nil {
		log.Debug("GenerateUnit", "error", err.Error())
		return &modules.Unit{}
	}
	// added by yangyu, 2018.8.9
	if newUnits == nil || len(newUnits) == 0 || newUnits[0].IsEmpty() {
		log.Info("No unit need to be packaged for now.")
		return &modules.Unit{}
	}

	pendingUnit := &newUnits[0]
	pendingUnit.UnitHeader.Creationdate = when.Unix()

	dag.setUnitHeader(pendingUnit)

	pendingUnit.UnitHeader.ParentsHash[0] = dag.HeadUnitHash()
	pendingUnit.UnitHeader.Number.Index = dag.HeadUnitNum() + 1
	pendingUnit.UnitHeader.GroupPubKey = groupPubKey
	pendingUnit.Hash()

	sign_unit, err1 := dagcommon.GetUnitWithSig(pendingUnit, ks, producer)
	if err1 != nil {
		log.Debug(fmt.Sprintf("GetUnitWithSig error: %v", err))
	}

	sign_unit.UnitSize = sign_unit.Size()

	//go log.Debug("Dag", "GenerateUnit unit:", *sign_unit)

	dag.PushUnit(sign_unit, txpool)
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
	// 1. 如果当前初生产的验证单元不在最长链条上，那么就切换到最长链分叉上。

	// 2. 更新状态
	dag.ApplyUnit(newUnit)

	// 3. 将验证单元添加到本地DB
	//err := dag.SaveUnit(newUnit, false)
	//if err != nil {
	//	log.Debug("unit_production", "PushUnit err:", err)
	//	return false
	//}
	dag.SaveUnit(newUnit, txpool, false)

	return true
}

// ApplyUnit, 利用下一个 unit 更新整个区块链状态
func (dag *Dag) ApplyUnit(nextUnit *modules.Unit) {
	// 1. 下一个 unit 和本地 unit 连续性的判断
	if nextUnit.ParentHash()[0] != dag.HeadUnitHash() {
		// todo 出现分叉, 调用本方法之前未处理分叉
		return
	}

	// 2. 验证 unit 的 mediator 调度
	if !dag.validateMediatorSchedule(nextUnit) {
		return
	}

	// 5. 更新Unit中交易的状态

	// 3. 计算当前 unit 到上一个 unit 之间的缺失数量，并更新每个mediator的unit的缺失数量
	missed := dag.updateMediatorMissedUnits(nextUnit)

	// 4. 更新全局动态属性值
	dag.updateDynGlobalProp(nextUnit, missed)

	// 5. 更新 mediator 的相关数据
	dag.updateSigningMediator(nextUnit)

	// 6. 更新最新不可逆区块高度
	dag.updateLastIrreversibleUnit()

	// 7. 判断是否到了维护周期，并维护
	dag.performChainMaintenance(nextUnit)

	// 8. 洗牌
	dag.updateMediatorSchedule()
}
