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
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package migration

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

type Migration0615_100 struct {
	//mdag  dag.IDag
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration0615_100) FromVersion() string {
	return "0.6.15"
}
func (m *Migration0615_100) ToVersion() string {
	return "1.0.0-beta"
}
func (m *Migration0615_100) ExecuteUpgrade() error {
	data, _ := m.statedb.Get([]byte("gpGlobalProperty"))
	gp := &GlobalProperty0615{}
	err := rlp.DecodeBytes(data, gp)

	if err != nil {
		log.Error("ExecuteUpgrade error:" + err.Error())
		return err
	}
	newGp := modules.NewGlobalProp()
	for _, m := range gp.ActiveMediators {
		newGp.ActiveMediators[m] = true
	}
	for _, j := range gp.ActiveJuries {
		newGp.ActiveJuries[j] = true
	}
	for _, p := range gp.PrecedingMediators {
		newGp.PrecedingMediators[p] = true
	}
	newData, err := rlp.EncodeToBytes(newGp)
	if err != nil {
		log.Error("ExecuteUpgrade error:" + err.Error())
		return err
	}
	m.statedb.Put([]byte("gpGlobalProperty"), newData)
	return nil
}

type GlobalProperty0615 struct {
	GlobalPropBase0615

	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}
type GlobalPropBase0615 struct {
	ImmutableParameters core.ImmutableChainParameters // 不可改变的区块链网络参数
	ChainParameters     ChainParameters0615           // 区块链网络参数
}
type ChainParameters0615 struct {
	// 目前的操作交易费，current schedule of fees
	CurrentFees FeeSchedule `json:"currentFees"`

	// unit生产之间的间隔时间，以秒为单元。 interval in seconds between Units
	MediatorInterval uint8 `json:"mediatorInterval"`

	// 区块链维护事件之间的间隔，以秒为单元。 interval in sections between unit maintenance events
	MaintenanceInterval uint32 `json:"maintenanceInterval"`

	// 在维护时跳过的MediatorInterval数量。 number of MediatorInterval to skip at maintenance time
	MaintenanceSkipSlots uint8 `json:"maintenanceSkipSlots"`

	// 活跃mediator的最大数量。maximum number of active mediators
	MaximumMediatorCount uint8 `json:"maxMediatorCount"`
}

// 操作交易费计划
type FeeSchedule struct {
	// mediator 创建费用
	MediatorCreateFee uint64                `json:"mediatorCreateFee"`
	AccountUpdateFee  uint64                `json:"accountUpdateFee"`
	TransferFee       TransferFeeParameters `json:"transferPtnFee"`
}
type TransferFeeParameters struct {
	BaseFee       uint64 `json:"baseFee"`
	PricePerKByte uint64 `json:"pricePerKByte"`
}
