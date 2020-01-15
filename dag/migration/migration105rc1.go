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
 *  * @author PalletOne core developer albert <dev@pallet.one>
 *  * @date 2019-2020
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

type Migration105delta_105rc1 struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration105delta_105rc1) FromVersion() string {
	return "1.0.5-delta"
}

func (m *Migration105delta_105rc1) ToVersion() string {
	return "1.0.5-rc1"
}

func (m *Migration105delta_105rc1) ExecuteUpgrade() error {
	if err := m.upgradeGP(); err != nil {
		return err
	}

	if err := m.upgradeDGP(); err != nil {
		return err
	}

	return nil
}

func (m *Migration105delta_105rc1) upgradeGP() error {
	oldGp := &GlobalProperty105delta{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, oldGp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	newData := &modules.GlobalPropertyTemp{}
	newData.GlobalPropBaseTemp = oldGp.GlobalPropBaseTemp
	newData.ActiveMediators = oldGp.ActiveMediators
	newData.PrecedingMediators = oldGp.PrecedingMediators

	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type GlobalProperty105delta struct {
	modules.GlobalPropBaseTemp
	GlobalPropExtra105delta
}

type GlobalPropExtra105delta struct {
	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

func (m *Migration105delta_105rc1) upgradeDGP() error {
	oldDgp := &DynamicGlobalProperty105delta{}
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
	newData.CurrentAbsoluteSlot = oldDgp.CurrentASlot
	newData.MaintenanceFlag = oldDgp.MaintenanceFlag

	// 由于之前版本的 Uint128 没有实现 rlp方法，取出来的值一直是最大值
	newData.RecentSlotsFilled = uint128.MaxValue

	err = storage.StoreToRlpBytes(m.propdb, constants.DYNAMIC_GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type Uint128_105delta struct {
	lo, hi uint64
}

type DynamicGlobalProperty105delta struct {
	LastMediator        common.Address
	IsShuffledSchedule  bool
	NextMaintenanceTime uint32
	LastMaintenanceTime uint32
	CurrentASlot        uint64
	RecentSlotsFilled   Uint128_105delta
	MaintenanceFlag     bool
}
