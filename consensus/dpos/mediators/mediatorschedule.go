/**
@version 0.1
@author albert·gou
@time June 6, 2018
@brief 主要实现mediator调度相关的功能
*/

package mediators

import (
	"time"

	d "github.com/palletone/go-palletone/consensus/dpos"
)

// Mediator调度顺序结构体
type MediatorSchedule struct {
	CurrentShuffledMediators []*d.Mediator
}

// 洗牌算法，更新mediator的调度顺序
func (ms *MediatorSchedule) UpdateMediatorSchedule(gp *d.GlobalProperty, dgp *d.DynamicGlobalProperty) {
	aSize := uint32(len(gp.ActiveMediators))

	// 1. 判断是否到达洗牌时刻
	if dgp.LastVerifiedUnitNum%aSize != 0 {
		return
	}

	// 2. 判断CurrentShuffledMediators的空间是否够大
	if uint32(len(ms.CurrentShuffledMediators)) < aSize {
		ms.CurrentShuffledMediators = make([]*d.Mediator, aSize, aSize)
	}

	// 3. 截取slice为所需大小
	ms.CurrentShuffledMediators = ms.CurrentShuffledMediators[:aSize]

	// 4. 初始化数据
	for i, m := range gp.ActiveMediators {
		ms.CurrentShuffledMediators[i] = m
	}

	// 5. 打乱证人的调度顺序
	nowHi := uint64(dgp.LastVerifiedUnitTime.Unix() << 32)
	for i := uint32(0); i < aSize; i++ {
		// 高性能随机生成器(High performance random generator)
		// 原理请参考 http://xorshift.di.unimi.it/
		k := nowHi + uint64(i)*2685821657736338717
		k ^= k >> 12
		k ^= k << 25
		k ^= k >> 27
		k *= 2685821657736338717

		jmax := uint32(aSize - i)
		j := i + uint32(k%uint64(jmax))

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
If slotNum == 2, return the next scheduled mediator after 1 verified uint gap.
*/
func (ms *MediatorSchedule) GetScheduledMediator(dgp *d.DynamicGlobalProperty, slotNum uint32) *d.Mediator {
	currentASlot := dgp.CurrentASlot + uint64(slotNum)
	return ms.CurrentShuffledMediators[currentASlot%uint64(len(ms.CurrentShuffledMediators))]
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

如果slotNum == N 且 N > 0，则返回大于verifiedUnitTime的第N个单元验证间隔的对齐时间
If slotNum == N for N > 0, return the Nth next unit-interval-aligned time greater than head_block_time().
*/
func GetSlotTime(gp *d.GlobalProperty, dgp *d.DynamicGlobalProperty, slotNum uint32) time.Time {
	if slotNum == 0 {
		return time.Unix(0, 0)
	}

	interval := gp.ChainParameters.VerifiedUnitInterval

	// 本条件是用来生产创世区块
	if dgp.LastVerifiedUnitNum == 0 {
		/**
		注：第一个验证单元在genesisTime加上一个验证单元间隔
		n.b. first verifiedUnit is at genesisTime plus one verifiedUnitInterval
		*/
		genesisTime := dgp.LastVerifiedUnitTime.Unix()
		return time.Unix(genesisTime + int64(slotNum) * int64(interval), 0)
	}

	// 最近的验证单元的绝对slot
	var verifiedUnitAbsSlot = dgp.LastVerifiedUnitTime.Unix() / int64(interval)
	// 最近的时间槽起始时间
	verifiedUnitSlotTime := time.Unix(verifiedUnitAbsSlot*int64(interval), 0)

	// 在此处添加区块链网络参数修改维护的所需要的slot

	/**
	如果是维护周期的话，加上维护间隔时间
	如果不是，就直接加上验证单元的slot时间
	*/
	// "slot 1" is verifiedUnitSlotTime,
	// plus maintenance interval if last uint is a maintenance verifiedUnit
	// plus verifiedUnit interval if last uint is not a maintenance verifiedUnit
	return verifiedUnitSlotTime.Add(time.Second * time.Duration(slotNum) * time.Duration(interval))
}

/**
获取在给定时间或之前出现的最近一个slot。 Get the last slot which occurs AT or BEFORE the given time.
*/
func GetSlotAtTime(gp *d.GlobalProperty, dgp *d.DynamicGlobalProperty, when time.Time) uint32 {
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
	interval := int64(gp.ChainParameters.VerifiedUnitInterval)

	return uint32(diffSecs/interval) + 1
}
