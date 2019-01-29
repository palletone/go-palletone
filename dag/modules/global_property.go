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
	"fmt"
	"sort"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
)

type GlobalPropBase struct {
	ChainParameters core.ChainParameters // 区块链网络参数
}

func NewGlobalPropBase() *GlobalPropBase {
	return &GlobalPropBase{
		ChainParameters: core.NewChainParams(),
	}
}

// 全局属性的结构体定义
type GlobalProperty struct {
	*GlobalPropBase

	ActiveJuries       map[common.Address]bool //当前活跃Jury集合
	ActiveMediators    map[common.Address]bool // 当前活跃 mediator 集合；每个维护间隔更新一次
	PrecedingMediators map[common.Address]bool // 上一届 mediator
}

func NewGlobalProp() *GlobalProperty {
	return &GlobalProperty{
		GlobalPropBase:     NewGlobalPropBase(),
		ActiveJuries:       make(map[common.Address]bool, 0),
		ActiveMediators:    make(map[common.Address]bool, 0),
		PrecedingMediators: make(map[common.Address]bool, 0),
	}
}

// 动态全局属性的结构体定义
type DynamicGlobalProperty struct {
	//HeadUnitNum  uint64      // 最新单元的编号(数量)
	//HeadUnitHash common.Hash // 最新单元的 hash
	//HeadUnitTime int64       // 最新单元的时间

	// 防止同一个mediator连续生产单元导致分叉
	LastMediator       common.Address // 最新单元的生产 mediator
	IsShuffledSchedule bool           // 标记 mediator 的调度顺序是否刚被打乱

	NextMaintenanceTime uint32 // 下一次系统维护时间
	LastMaintenanceTime uint32 // 上一次系统维护时间

	// 当前的绝对时间槽数量，== 从创世开始所有的时间槽数量 == UnitNum + 丢失的槽数量
	CurrentASlot uint64

	/**
	在过去的128个单元生产slots中miss的数量。
	The count of Unit production slots that were missed in the past 128 Units
	用于计算mediator的参与率。used to compute mediator participation.
	*/
	// RecentSlotsFilled float32

	//LastIrreversibleUnitNum uint32
	//NewestUnit     map[IDType16]*UnitProperty
	//LastStableUnit map[IDType16]*UnitProperty
}
type UnitProperty struct {
	Hash      common.Hash // 最近的单元hash
	Index     *ChainIndex // 最近的单元编号(数量)
	Timestamp uint32      // 最近的单元时间
}

func NewDynGlobalProp() *DynamicGlobalProperty {
	return &DynamicGlobalProperty{
		//HeadUnitNum:             0,
		//HeadUnitHash:            common.Hash{},

		LastMediator:       common.Address{},
		IsShuffledSchedule: false,

		NextMaintenanceTime: 0,
		LastMaintenanceTime: 0,
		CurrentASlot:        0,

		//LastIrreversibleUnitNum: 0,
		//NewestUnit:     map[IDType16]*UnitProperty{},
		//LastStableUnit: map[IDType16]*UnitProperty{},
	}
}

//func (gdp *DynamicGlobalProperty) SetNewestUnit(header *Header) {
//	gdp.NewestUnit[header.Number.AssetID] = &UnitProperty{header.Hash(), header.Number, header.Creationdate}
//}
//func (gdp *DynamicGlobalProperty) SetLastStableUnit(header *Header) {
//	gdp.LastStableUnit[header.Number.AssetID] = &UnitProperty{header.Hash(), header.Number, header.Creationdate}
//}

const TERMINTERVAL = 50 //DEBUG:50, DEPLOY:15000

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

func (gp *GlobalProperty) IsActiveJury(add common.Address) bool {
	return true //todo for test

	//return gp.ActiveJuries[add]
}

func (gp *GlobalProperty) GetActiveJuries() []common.Address {
	juries := make([]common.Address, 0, len(gp.ActiveJuries))
	for addr, _ := range gp.ActiveJuries {
		juries = append(juries, addr)
	}
	sortAddress(juries)

	return juries
}

func (gp *GlobalProperty) IsActiveMediator(add common.Address) bool {
	return gp.ActiveMediators[add]
}

func (gp *GlobalProperty) IsPrecedingMediator(add common.Address) bool {
	return gp.PrecedingMediators[add]
}

func (gp *GlobalProperty) GetActiveMediatorAddr(index int) common.Address {
	if index < 0 || index > gp.ActiveMediatorsCount()-1 {
		log.Error(fmt.Sprintf("%v is out of the bounds of active mediator list!", index))
	}

	meds := gp.GetActiveMediators()

	return meds[index]
}

// GetActiveMediators, return the list of active mediators, and the order of the list from small to large
func (gp *GlobalProperty) GetActiveMediators() []common.Address {
	mediators := make([]common.Address, 0, gp.ActiveMediatorsCount())

	for medAdd, _ := range gp.ActiveMediators {
		mediators = append(mediators, medAdd)
	}

	sortAddress(mediators)

	return mediators
}

func sortAddress(adds []common.Address) {
	aSize := len(adds)
	addStrs := make([]string, aSize, aSize)
	for i, add := range adds {
		addStrs[i] = add.Str()
	}

	sort.Strings(addStrs)

	for i, addStr := range addStrs {
		adds[i], _ = common.StringToAddress(addStr)
	}
}

func InitGlobalProp(genesis *core.Genesis) *GlobalProperty {
	log.Debug("initialize global property...")

	// Create global properties
	gp := NewGlobalProp()

	log.Debug("initialize chain parameters...")
	gp.ChainParameters = genesis.InitialParameters

	log.Debug("Set active mediators...")
	// Set active mediators
	for i := uint16(0); i < genesis.InitialActiveMediators; i++ {
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

	// Create dynamic global properties
	dgp := NewDynGlobalProp()
	//dgp.HeadUnitTime = genesis.InitialTimestamp
	//dgp.HeadUnitHash = genesisUnitHash
	//dgp.SetNewestUnit(genesis.Header())
	//dgp.SetLastStableUnit(genesis.Header())
	return dgp
}

// UpdateDynGlobalProp, update global dynamic data
// @author Albert·Gou
func (dgp *DynamicGlobalProperty) UpdateDynGlobalProp(unit *Unit, missedUnits uint64) {
	//dgp.HeadUnitNum = unit.NumberU64()
	//dgp.HeadUnitHash = unit.Hash()
	//dgp.HeadUnitTime = unit.Timestamp()
	//dgp.SetNewestUnit(unit.Header())

	dgp.LastMediator = unit.Author()
	dgp.IsShuffledSchedule = false

	dgp.CurrentASlot += missedUnits + 1
}
