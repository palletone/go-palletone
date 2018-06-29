/**
@version 0.1
@author albert·gou
@time June 11, 2018
@brief 生产验证单元的功能
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
