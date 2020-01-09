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

package ptnjson

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

type UnitPropertyJson struct {
	Hash      common.Hash `json:"hash"`
	Number    uint64      `json:"number"`
	Timestamp string      `json:"timestamp"`
}

func UnitPropertyToJson(up *modules.UnitProperty) *UnitPropertyJson {
	return &UnitPropertyJson{
		Hash:      up.Hash,
		Number:    up.ChainIndex.Index,
		Timestamp: time.Unix(int64(up.Timestamp), 0).Format("2006-01-02 15:04:05 -0700 MST"),
	}
}

type DynamicGlobalPropertyJson struct {
	LastMediator        string `json:"lastMediator"`
	IsShuffledSchedule  bool   `json:"isShuffledSchedule"`
	NextMaintenanceTime string `json:"nextMaintenanceTime"`
	LastMaintenanceTime string `json:"lastMaintenanceTime"`
	CurrentAbsoluteSlot uint64 `json:"currentAbsoluteSlot"`
	RecentSlotsFilled   string `json:"recentSlotsFilled"`
	MaintenanceFlag     bool   `json:"maintenanceFlag"`
}

func DynGlobalPropToJson(dgp *modules.DynamicGlobalProperty) *DynamicGlobalPropertyJson {
	return &DynamicGlobalPropertyJson{
		LastMediator:        dgp.LastMediator.Str(),
		IsShuffledSchedule:  dgp.IsShuffledSchedule,
		NextMaintenanceTime: time.Unix(int64(dgp.NextMaintenanceTime), 0).Format("2006-01-02 15:04:05 -0700 MST"),
		LastMaintenanceTime: time.Unix(int64(dgp.LastMaintenanceTime), 0).Format("2006-01-02 15:04:05 -0700 MST"),
		CurrentAbsoluteSlot: dgp.CurrentASlot,
		RecentSlotsFilled:   dgp.RecentSlotsFilled.BinaryStr(),
		MaintenanceFlag:     dgp.MaintenanceFlag,
	}
}
