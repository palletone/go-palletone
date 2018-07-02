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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 * @brief 实现生产验证单元的功能。Implement the function of production verification unit.
 */

package verifyunit

import (
	"time"

	"github.com/palletone/go-palletone/common/log"

	d "github.com/palletone/go-palletone/consensus/dpos"
	s "github.com/palletone/go-palletone/consensus/dpos/mediators"
	a "github.com/palletone/go-palletone/core/application"
	v "github.com/palletone/go-palletone/dag/verifiedunit"
)

func GenerateVerifiedUnit(
	when time.Time,
	//	m *d.Mediator,
	signKey string,
	db *a.DataBase) v.VerifiedUnit {

	gp := &db.GlobalProp
	dgp := &db.DynGlobalProp

	// 1. 判断是否满足生产的若干条件

	// 2. 生产验证单元，添加交易集、时间戳、签名
//	println("\n正在生产验证单元...")
//	println("\nGenerating Verified Unit...")
	log.Info("Generating Verified Unit...")

	var vu v.VerifiedUnit
	vu.Timestamp = when
	vu.MediatorSig = signKey
	vu.PreVerifiedUnit = dgp.LastVerifiedUnit
	vu.VerifiedUnitNum = dgp.LastVerifiedUnitNum + 1

	// 3. 从未验证交易池中移除添加的交易

	// 3. 如果当前初生产的验证单元不在最长链条上，那么就切换到最长链分叉上。

	// 4. 将验证单元添加到本地DB
//	go println("将新验证单元添加到DB...")
//	go println("storing the new verified unit to database...")
	go log.Info("storing the new verified unit to database...")

	// 5. 更新全局动态属性值
//	println("更新全局动态属性值...")
//	println("Updating global dynamic property...")
	log.Info("Updating global dynamic property...")
	UpdateGlobalDynProp(gp, dgp, &vu)

	// 5. 判断是否到了维护周期，并维护

	// 6. 洗牌
//	println("尝试打乱mediators的调度顺序...")
//	println("shuffling the scheduling order of mediator...")
	log.Info("shuffling the scheduling order of mediator...")
	db.MediatorSchl.UpdateMediatorSchedule(gp, dgp)

	return vu
}

func UpdateGlobalDynProp(gp *d.GlobalProperty, dgp *d.DynamicGlobalProperty, vu *v.VerifiedUnit) {
	dgp.LastVerifiedUnitNum = vu.VerifiedUnitNum
	dgp.LastVerifiedUnit = vu
	dgp.LastVerifiedUnitTime = vu.Timestamp

	missedUnits := uint64(s.GetSlotAtTime(gp, dgp, vu.Timestamp))
	//	println(missedUnits)
	dgp.CurrentASlot += missedUnits + 1
}
