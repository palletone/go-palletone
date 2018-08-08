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

package mediatorplugin

import (
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

// GenerateVerifiedUnit, generate unit
// @author Albert·Gou
func GenerateUnit(dag *dag.Dag, when time.Time, producer common.Mediator, ks *keystore.KeyStore, txspool *txspool.TxPool) modules.Unit {
	dgp := dag.DynGlobalProp

	// 1. 判断是否满足生产的若干条件

	// 2. 生产验证单元，添加交易集、时间戳、签名
	log.Debug("Generating Verified Unit...")

	units, _ := dagcommon.CreateUnit(&producer.Address, txspool)
	pendingUnit := units[0]
	pendingUnit.UnitHeader.Creationdate = when.Unix()
	pendingUnit.UnitHeader.Number.Index = dgp.LastVerifiedUnitNum + 1
	pendingUnit.UnitHeader.ParentsHash =
		append(pendingUnit.UnitHeader.ParentsHash, dgp.LastVerifiedUnitHash)
	pendingUnit.UnitHash = pendingUnit.Hash()

	_, err := dagcommon.GetUnitWithSig(&pendingUnit, ks, producer.Address)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
	}

	pendingUnit.UnitSize = pendingUnit.Size()

	PushUnit(dag, &pendingUnit)

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
func PushUnit(dag *dag.Dag, newUnit *modules.Unit) bool {
	// 3. 如果当前初生产的验证单元不在最长链条上，那么就切换到最长链分叉上。

	ApplyUnit(dag, newUnit)

	// 4. 将验证单元添加到本地DB
	log.Debug("storing the new verified unit to database...")
	go dagcommon.SaveUnit(*newUnit, false)

	return false
}

func ApplyUnit(dag *dag.Dag, nextUnit *modules.Unit) {
	gp := dag.GlobalProp
	dgp := dag.DynGlobalProp

	// 4. 更新Unit中交易的状态

	// 5. 更新全局动态属性值
	log.Debug("Updating global dynamic property...")
	dagcommon.UpdateGlobalDynProp(gp, dgp, nextUnit)

	// 5. 判断是否到了维护周期，并维护

	// 6. 洗牌
	log.Debug("shuffling the scheduling order of mediator...")
	dag.MediatorSchl.UpdateMediatorSchedule(gp, dgp)
}
