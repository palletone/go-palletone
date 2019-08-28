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
 *
 */

package modules

import (
	"sort"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
)

type GlobalPropBase struct {
	ImmutableParameters core.ImmutableChainParameters // 不可改变的区块链网络参数
	ChainParameters     core.ChainParameters          // 区块链网络参数
}

func NewGlobalPropBase() GlobalPropBase {
	return GlobalPropBase{
		ImmutableParameters: core.NewImmutChainParams(),
		ChainParameters:     core.NewChainParams(),
	}
}

// 全局属性的结构体定义
type GlobalProperty struct {
	GlobalPropBase

	// todo albert 待重构为数组，提高效率
	ActiveJuries       map[common.Address]bool // 当前活跃Jury集合
	ActiveMediators    map[common.Address]bool // 当前活跃 mediator 集合；每个维护间隔更新一次
	PrecedingMediators map[common.Address]bool // 上一届 mediator
}

func NewGlobalProp() *GlobalProperty {
	return &GlobalProperty{
		GlobalPropBase:     NewGlobalPropBase(),
		ActiveJuries:       make(map[common.Address]bool),
		ActiveMediators:    make(map[common.Address]bool),
		PrecedingMediators: make(map[common.Address]bool),
	}
}

type GlobalPropertyHistory struct {
	// unit生产之间的间隔时间，以秒为单元。 interval in seconds between Units
	MediatorInterval uint8 `json:"mediatorInterval"`

	// 区块链维护事件之间的间隔，以秒为单元。 interval in sections between unit maintenance events
	MaintenanceInterval uint32 `json:"maintenanceInterval"`

	// 在维护时跳过的MediatorInterval数量。 number of MediatorInterval to skip at maintenance time
	MaintenanceSkipSlots uint8 `json:"maintenanceSkipSlots"`

	ActiveJuries    []common.Address //当前活跃Jury集合
	ActiveMediators []common.Address // 当前活跃 mediator 集合；每个维护间隔更新一次
	EffectiveTime   uint64           //生效时间
	EffectiveHeight uint64           //生效高度
	ExpiredTime     uint64           //失效时间
}

// 动态全局属性的结构体定义
type DynamicGlobalProperty struct {
	// 防止同一个mediator连续生产单元导致分叉
	LastMediator       common.Address // 最新单元的生产 mediator
	IsShuffledSchedule bool           // 标记 mediator 的调度顺序是否刚被打乱

	NextMaintenanceTime uint32 // 下一次系统维护时间
	LastMaintenanceTime uint32 // 上一次系统维护时间

	// 当前的绝对时间槽数量，== 从创世开始所有的时间槽数量 == UnitNum + 丢失的槽数量
	CurrentASlot uint64

	// 记录每个生产slot的unit生产情况，用于计算mediator的参与率。
	// 每一位表示一个生产slot，mediator正常生产unit则值为1，否则为0。
	// 最低位表示最近一个slot， 初始值全为1。
	RecentSlotsFilled uint64

	// If MaintenanceFlag is true, then the head unit is a maintenance unit.
	// This means GetTimeSlot(1) - HeadBlockTime() will have a gap due to maintenance duration.
	//
	// This flag answers the question, "Was maintenance performed in the last call to ApplyUnit()?"
	MaintenanceFlag bool
}
type UnitProperty struct {
	Hash      common.Hash // 最近的单元hash
	Index     *ChainIndex // 最近的单元编号(数量)
	Timestamp uint32      // 最近的单元时间
}

func NewDynGlobalProp() *DynamicGlobalProperty {
	return &DynamicGlobalProperty{
		LastMediator:       common.Address{},
		IsShuffledSchedule: false,

		NextMaintenanceTime: 0,
		LastMaintenanceTime: 0,
		CurrentASlot:        0,

		RecentSlotsFilled: ^uint64(0),

		MaintenanceFlag: false,
	}
}

func (gp *GlobalProperty) ActiveMediatorsCount() int {
	return len(gp.ActiveMediators)
}

func (gp *GlobalProperty) PrecedingMediatorsCount() int {
	return len(gp.PrecedingMediators)
}

func (gp *GlobalProperty) ChainThreshold() int {
	return calcThreshold(gp.ActiveMediatorsCount())
}

func (gp *GlobalProperty) PrecedingThreshold() int {
	return calcThreshold(gp.PrecedingMediatorsCount())
}

func calcThreshold(aSize int) int {
	offset := (core.PalletOne100Percent - core.PalletOneIrreversibleThreshold) * aSize /
		core.PalletOne100Percent

	return aSize - offset
}

func (gp *GlobalProperty) IsActiveMediator(add common.Address) bool {
	return gp.ActiveMediators[add]
}

func (gp *GlobalProperty) IsPrecedingMediator(add common.Address) bool {
	return gp.PrecedingMediators[add]
}

func (gp *GlobalProperty) GetActiveMediatorAddr(index int) common.Address {
	if index < 0 || index > gp.ActiveMediatorsCount()-1 {
		log.Errorf("%v is out of the bounds of active mediator list!", index)
	}

	meds := gp.GetActiveMediators()

	return meds[index]
}

// GetActiveMediators, return the list of active mediators, and the order of the list from small to large
func (gp *GlobalProperty) GetActiveMediators() []common.Address {
	var mediators common.Addresses
	//mediators = make([]common.Address, 0, gp.ActiveMediatorsCount())

	for medAdd := range gp.ActiveMediators {
		mediators = append(mediators, medAdd)
	}

	sort.Sort(mediators)

	return mediators
}

func InitGlobalProp(genesis *core.Genesis) *GlobalProperty {
	log.Debug("initialize global property...")

	// Create global properties
	gp := NewGlobalProp()

	log.Debug("initialize chain parameters...")
	gp.ChainParameters = genesis.InitialParameters
	gp.ImmutableParameters = genesis.ImmutableParameters

	log.Debug("Set active mediators...")
	// Set active mediators
	for i := uint8(0); i < genesis.InitialParameters.ActiveMediatorCount; i++ {
		initMed := genesis.InitialMediatorCandidates[i]
		addr, err := core.StrToMedAdd(initMed.AddStr)
		if err != nil {
			panic(err.Error())
		}
		gp.ActiveMediators[addr] = true
	}

	return gp
}

func InitDynGlobalProp(genesis *Unit) *DynamicGlobalProperty {
	log.Debug("initialize dynamic global property...")
	dgp := NewDynGlobalProp()

	return dgp
}
