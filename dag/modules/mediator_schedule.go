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

	//ms.UpdateMediatorSchedule(gp, dgp)

	return ms
}

func NewMediatorSchl() *MediatorSchedule {
	return &MediatorSchedule{
		CurrentShuffledMediators: []common.Address{},
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

// UpdateDynGlobalProp, update global dynamic data
// @author Albert·Gou
func (dgp *DynamicGlobalProperty) UpdateDynGlobalProp(unit *Unit, missedUnits uint64) {
	//dgp.HeadUnitNum = unit.NumberU64()
	//dgp.HeadUnitHash = unit.Hash()
	//dgp.HeadUnitTime = unit.Timestamp()
	//dgp.SetNewestUnit(unit.Header())
	dgp.CurrentASlot += missedUnits + 1
}
