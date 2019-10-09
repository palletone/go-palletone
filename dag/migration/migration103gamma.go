/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/uint128"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration103beta_103gamma struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration103beta_103gamma) FromVersion() string {
	return "1.0.3-beta"
}

func (m *Migration103beta_103gamma) ToVersion() string {
	return "1.0.3-gamma"
}

func (m *Migration103beta_103gamma) ExecuteUpgrade() error {
	//转换GLOBALPROPERTY结构体
	if err := m.upgradeDGP(); err != nil {
		return err
	}

	return nil
}

func (m *Migration103beta_103gamma) upgradeDGP() error {
	oldDgp := &DynamicGlobalProperty103beta{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.DYNAMIC_GLOBALPROPERTY_KEY, oldDgp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	newData := &modules.DynamicGlobalProperty{}
	newData.LastMediator = oldDgp.LastMediator
	newData.IsShuffledSchedule = oldDgp.IsShuffledSchedule
	newData.NextMaintenanceTime = oldDgp.NextMaintenanceTime
	newData.LastMaintenanceTime = oldDgp.LastMaintenanceTime
	newData.CurrentASlot = oldDgp.CurrentASlot
	newData.MaintenanceFlag = oldDgp.MaintenanceFlag

	newData.RecentSlotsFilled = uint128.New(oldDgp.RecentSlotsFilled, ^uint64(0))

	err = storage.StoreToRlpBytes(m.propdb, constants.DYNAMIC_GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type DynamicGlobalProperty103beta struct {
	LastMediator       common.Address // 最新单元的生产 mediator
	IsShuffledSchedule bool           // 标记 mediator 的调度顺序是否刚被打乱

	NextMaintenanceTime uint32 // 下一次系统维护时间
	LastMaintenanceTime uint32 // 上一次系统维护时间

	CurrentASlot uint64

	RecentSlotsFilled uint64

	MaintenanceFlag bool
}
