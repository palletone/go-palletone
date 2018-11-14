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

// GenerateVerifiedUnit, generate unit
// @author Albert·Gou
func (dag *Dag) GenerateUnit(when time.Time, producer common.Address,
	ks *keystore.KeyStore, txspool txspool.ITxPool) *modules.Unit {

	// 1. 判断是否满足生产的若干条件

	// 2. 生产验证单元，添加交易集、时间戳、签名
	newUnits, err := dag.CreateUnit(&producer, txspool, ks, when)
	if err != nil {
		log.Error("GenerateUnit", "error", err.Error())
		return &modules.Unit{}
	}
	// added by yangyu, 2018.8.9
	if newUnits == nil || len(newUnits) == 0 || newUnits[0].IsEmpty() {
		log.Info("No unit need to be packaged for now.")
		return &modules.Unit{}
	}

	pendingUnit := &newUnits[0]
	pendingUnit.UnitHeader.Creationdate = when.Unix()
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
		log.Info("the pending unit header number index info. ", "index", pendingUnit.UnitHeader.Number.Index,
			"hex", pendingUnit.UnitHeader.Number.AssetID.String())
	}

	pendingUnit.Hash()
	_, err = dagcommon.GetUnitWithSig(pendingUnit, ks, producer)
	if err != nil {
		log.Error(fmt.Sprintf("GetUnitWithSig error: %v", err))
	}

	pendingUnit.UnitSize = pendingUnit.Size()

	dag.PushUnit(pendingUnit)
	return pendingUnit
}

/**
 * Push unit "may fail" in which case every partial change is unwound.  After
 * push unit is successful the block is appended to the chain database on disk.
 *
 * 推块“可能会失败”，在这种情况下，每个部分地更改都会撤销。 推块成功后，该块将附加到磁盘上的链数据库。
 *
 * @return true if we switched forks as a result of this push.
 */
func (dag *Dag) PushUnit(newUnit *modules.Unit) bool {
	// 3. 如果当前初生产的验证单元不在最长链条上，那么就切换到最长链分叉上。

	dag.ApplyUnit(newUnit)

	// 4. 将验证单元添加到本地DB
	//err := dag.SaveUnit(newUnit, false)
	//if err != nil {
	//	log.Error("unit_production", "PushUnit err:", err)
	//	return false
	//}
	go dag.SaveUnit(newUnit, false)

	return true
}

func (dag *Dag) ApplyUnit(nextUnit *modules.Unit) {
	// 4. 更新Unit中交易的状态

	// 5. 更新全局动态属性值
	missed := dag.UpdateMediatorMissedUnits(nextUnit)
	dag.UpdateDynGlobalProp(nextUnit, missed)

	// 6. 更新最新不可逆区块高度
	dag.UpdateLastIrreversibleUnit()

	// 7. 判断是否到了维护周期，并维护

	// 8. 洗牌
	dag.UpdateMediatorSchedule()
}

func (dag *Dag) UpdateLastIrreversibleUnit() {
	// todo
}
