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
 * @brief 主要实现mediator调度相关的功能。implements mediator scheduling related functions.
 */

package modules

import (
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
)

// Mediator调度顺序结构体
type MediatorSchedule struct {
	CurrentShuffledMediators []common.Address
}

func InitMediatorSchl(gp *GlobalProperty, dgp *DynamicGlobalProperty) *MediatorSchedule {
	log.Debug("initialize mediator schedule...")
	ms := NewMediatorSchl()

	aSize := uint64(len(gp.ActiveMediators))
	if aSize == 0 {
		log.Error("The current number of active mediators is 0!")
	}

	// Create witness scheduler
	ms.CurrentShuffledMediators = make([]common.Address, aSize, aSize)
	meds := gp.GetActiveMediators()
	for i, add := range meds {
		ms.CurrentShuffledMediators[i] = add
	}

	ms.UpdateMediatorSchedule(gp, dgp)

	return ms
}

func NewMediatorSchl() *MediatorSchedule {
	return &MediatorSchedule{
		CurrentShuffledMediators: []common.Address{},
	}
}

// 洗牌算法，更新mediator的调度顺序
func (ms *MediatorSchedule) UpdateMediatorSchedule(gp *GlobalProperty, dgp *DynamicGlobalProperty) {
	aSize := uint64(len(gp.ActiveMediators))
	if aSize == 0 {
		log.Error("The current number of active mediators is 0!")
		return
	}

	// 1. 判断是否到达洗牌时刻
	if dgp.HeadUnitNum%aSize != 0 {
		return
	}

	// 2. 清除CurrentShuffledMediators原来的空间，重新分配空间
	ms.CurrentShuffledMediators = make([]common.Address, aSize, aSize)

	// 3. 初始化数据
	meds := gp.GetActiveMediators()
	for i, add := range meds {
		ms.CurrentShuffledMediators[i] = add
	}

	// 4. 打乱证人的调度顺序
	nowHi := uint64(dgp.HeadUnitTime << 32)
	for i := uint64(0); i < aSize; i++ {
		// 高性能随机生成器(High performance random generator)
		// 原理请参考 http://xorshift.di.unimi.it/
		k := nowHi + uint64(i)*2685821657736338717
		k ^= k >> 12
		k ^= k << 25
		k ^= k >> 27
		k *= 2685821657736338717

		jmax := aSize - i
		j := i + k%jmax

		// 进行N次随机交换
		ms.CurrentShuffledMediators[i], ms.CurrentShuffledMediators[j] =
			ms.CurrentShuffledMediators[j], ms.CurrentShuffledMediators[i]
	}
}

/**
@brief 获取指定的未来slotNum对应的调度mediator来生产见证单元.
Get the mediator scheduled for uint verification in a slot.

slotNum总是对应于未来的时间。
slotNum always corresponds to a time in the future.

如果slotNum == 1，则返回下一个调度Mediator。
If slotNum == 1, return the next scheduled mediator.

如果slotNum == 2，则返回下下一个调度Mediator。
If slotNum == 2, return the next scheduled mediator after 1 uint gap.
*/
func (ms *MediatorSchedule) GetScheduledMediator(dgp *DynamicGlobalProperty, slotNum uint32) common.Address {
	currentASlot := dgp.CurrentASlot + uint64(slotNum)
	csmLen := len(ms.CurrentShuffledMediators)
	if csmLen == 0 {
		log.Error("The current number of shuffled mediators is 0!")
		return common.Address{}
	}

	// 由于创世单元不是有mediator生产，所以这里需要减1
	index := (currentASlot - 1) % uint64(csmLen)
	return ms.CurrentShuffledMediators[index]
}

/**
计算在过去的128个见证单元生产slots中miss的百分比，不包括当前验证单元。
Calculate the percent of verifiedUnit production slots that were missed in the past 128 verifiedUnits,
not including the current verifiedUnit.
*/
//func MediatorParticipationRate(dgp *d.DynamicGlobalProperty) float32 {
//	return dgp.RecentSlotsFilled / 128.0
//}

/**
@brief 获取给定的未来第slotNum个slot开始的时间。
Get the time at which the given slot occurs.

如果slotNum == 0，则返回time.Unix(0,0)。
If slotNum == 0, return time.Unix(0,0).

如果slotNum == N 且 N > 0，则返回大于UnitTime的第N个单元验证间隔的对齐时间
If slotNum == N for N > 0, return the Nth next unit-interval-aligned time greater than head_block_time().
*/
func GetSlotTime(gp *GlobalProperty, dgp *DynamicGlobalProperty, slotNum uint32) time.Time {
	if slotNum == 0 {
		return time.Unix(0, 0)
	}

	interval := gp.ChainParameters.MediatorInterval

	// 本条件是用来生产第一个unit
	if dgp.HeadUnitNum == 0 {
		/**
		注：第一个验证单元在genesisTime加上一个验证单元间隔
		n.b. first Unit is at genesisTime plus one UnitInterval
		*/
		genesisTime := dgp.HeadUnitTime
		return time.Unix(genesisTime+int64(slotNum)*int64(interval), 0)
	}

	// 最近的验证单元的绝对slot
	var unitAbsSlot = dgp.HeadUnitTime / int64(interval)
	// 最近的时间槽起始时间
	verifiedUnitSlotTime := time.Unix(unitAbsSlot*int64(interval), 0)

	// 在此处添加区块链网络参数修改维护的所需要的slot

	/**
	如果是维护周期的话，加上维护间隔时间
	如果不是，就直接加上验证单元的slot时间
	*/
	// "slot 1" is UnitSlotTime,
	// plus maintenance interval if last uint is a maintenance Unit
	// plus Unit interval if last uint is not a maintenance Unit
	return verifiedUnitSlotTime.Add(time.Second * time.Duration(slotNum) * time.Duration(interval))
}

/**
获取在给定时间或之前出现的最近一个slot。 Get the last slot which occurs AT or BEFORE the given time.
*/
func GetSlotAtTime(gp *GlobalProperty, dgp *DynamicGlobalProperty, when time.Time) uint32 {
	/**
	返回值是所有满足 GetSlotTime（N）<= when 中最大的N
	The return value is the greatest value N such that GetSlotTime( N ) <= when.
	如果都不满足，则返回 0
	If no such N exists, return 0.
	*/
	firstSlotTime := GetSlotTime(gp, dgp, 1)

	if when.Before(firstSlotTime) {
		return 0
	}

	diffSecs := when.Unix() - firstSlotTime.Unix()
	interval := int64(gp.ChainParameters.MediatorInterval)
	if interval == 0 {
		return 0
	}
	return uint32(diffSecs/interval) + 1
}

// UpdateDynGlobalProp, update global dynamic data
// @author Albert·Gou
func (dgp *DynamicGlobalProperty) UpdateDynGlobalProp(gp *GlobalProperty, unit *Unit, missedUnits uint64) {
	timestamp := unit.UnitHeader.Creationdate

	dgp.HeadUnitNum = unit.UnitHeader.Number.Index
	dgp.HeadUnitHash = unit.UnitHash
	dgp.HeadUnitTime = timestamp

	//	println(missedUnits)
	dgp.CurrentASlot += missedUnits + 1

}
