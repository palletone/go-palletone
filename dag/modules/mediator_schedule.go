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
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
)

// Mediator调度顺序结构体
type MediatorSchedule struct {
	CurrentShuffledMediators []common.Address
}

func (ms *MediatorSchedule) String() string {
	data, _ := json.Marshal(ms.CurrentShuffledMediators)
	return string(data)
}
func InitMediatorSchl(gp *GlobalProperty, dgp *DynamicGlobalProperty) *MediatorSchedule {
	log.Debug("initialize mediator schedule...")
	ms := NewMediatorSchl()

	aSize := uint64(len(gp.ActiveMediators))
	if aSize == 0 {
		log.Error("The current number of active mediators is 0!")
	}

	// Create witness scheduler
	ms.CurrentShuffledMediators = make([]common.Address, aSize)
	copy(ms.CurrentShuffledMediators, gp.GetActiveMediators())

	return ms
}

func NewMediatorSchl() *MediatorSchedule {
	return &MediatorSchedule{
		CurrentShuffledMediators: []common.Address{},
	}
}
