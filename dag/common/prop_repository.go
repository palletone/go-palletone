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

package common

import (
	"encoding/binary"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type PropRepository struct {
	db storage.IPropertyDb
}

type IPropRepository interface {
	StoreGlobalProp(gp *modules.GlobalProperty) error
	RetrieveGlobalProp() (*modules.GlobalProperty, error)
	StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error
	RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error)
	StoreMediatorSchl(ms *modules.MediatorSchedule) error
	RetrieveMediatorSchl() (*modules.MediatorSchedule, error)
	GetChainThreshold() (int, error)
	// SetLastStableUnit(hash common.Hash, index *modules.ChainIndex) error
	// GetLastStableUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error)
	SetNewestUnit(header *modules.Header) error
	GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error)
	GetNewestUnitTimestamp(token modules.AssetId) (int64, error)
	GetScheduledMediator(slotNum uint32) common.Address
	UpdateMediatorSchedule() bool
	GetSlotTime(slotNum uint32) time.Time
	GetSlotAtTime(when time.Time) uint32

	SaveChaincode(contractId common.Address, cc *list.CCInfo) error
	GetChaincode(contractId common.Address) (*list.CCInfo, error)
	RetrieveChaincodes() ([]*list.CCInfo, error)
	GetChainParameters() *core.ChainParameters
}

func (pRep *PropRepository) GetChainParameters() *core.ChainParameters {
	return pRep.db.GetChainParameters()
}

func (pRep *PropRepository) GetChaincode(contractId common.Address) (*list.CCInfo, error) {
	return pRep.db.GetChaincode(contractId)
}

func (pRep *PropRepository) SaveChaincode(contractId common.Address, cc *list.CCInfo) error {
	return pRep.db.SaveChaincode(contractId, cc)
}

func (pRep *PropRepository) RetrieveChaincodes() ([]*list.CCInfo, error) {
	return pRep.db.RetrieveChaincodes()
}
func NewPropRepository(db storage.IPropertyDb) *PropRepository {
	return &PropRepository{db: db}
}
func NewPropRepository4Db(db ptndb.Database) *PropRepository {
	pdb := storage.NewPropertyDb(db)
	return &PropRepository{db: pdb}
}
func (pRep *PropRepository) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	return pRep.db.RetrieveGlobalProp()
}
func (pRep *PropRepository) GetChainThreshold() (int, error) {
	gp, err := pRep.RetrieveGlobalProp()
	if err != nil {
		return 0, err
	}
	return gp.ChainThreshold(), nil
}

func (pRep *PropRepository) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	return pRep.db.RetrieveDynGlobalProp()
}

func (pRep *PropRepository) StoreGlobalProp(gp *modules.GlobalProperty) error {
	return pRep.db.StoreGlobalProp(gp)
}

func (pRep *PropRepository) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {
	return pRep.db.StoreDynGlobalProp(dgp)
}

func (pRep *PropRepository) StoreMediatorSchl(ms *modules.MediatorSchedule) error {
	return pRep.db.StoreMediatorSchl(ms)
}

func (pRep *PropRepository) RetrieveMediatorSchl() (*modules.MediatorSchedule, error) {
	return pRep.db.RetrieveMediatorSchl()
}

// func (pRep *PropRepository) SetLastStableUnit(hash common.Hash, index *modules.ChainIndex) error {
// 	return pRep.db.SetLastStableUnit(hash, index)
// }
// func (pRep *PropRepository) GetLastStableUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
// 	return pRep.db.GetLastStableUnit(token)
// }

func (pRep *PropRepository) SetNewestUnit(header *modules.Header) error {
	return pRep.db.SetNewestUnit(header)
}

func (pRep *PropRepository) GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
	hash, index, _, e := pRep.db.GetNewestUnit(token)
	return hash, index, e
}

func (pRep *PropRepository) GetNewestUnitTimestamp(token modules.AssetId) (int64, error) {
	_, _, t, e := pRep.db.GetNewestUnit(token)
	return t, e
}

// 洗牌算法，更新mediator的调度顺序
func (pRep *PropRepository) UpdateMediatorSchedule() bool {
	gp, err := pRep.RetrieveGlobalProp()
	if err != nil {
		log.Debugf("Retrieve Global Prop error: %v", err.Error())
		return false
	}
	ms, err := pRep.RetrieveMediatorSchl()
	if err != nil {
		log.Debugf("Retrieve MediatorSchl error: %v", err.Error())
		return false
	}

	token := dagconfig.DagConfig.GetGasToken()
	hash, idx, _, err := pRep.db.GetNewestUnit(token)
	if err != nil {
		log.Debugf("GetNewestUnit error:" + err.Error())
		return false
	}

	aSize := uint64(len(gp.ActiveMediators))
	if aSize == 0 {
		log.Debugf("the current number of active mediators is 0")
		return false
	}

	// 1. 判断是否到达洗牌时刻
	if idx.Index%aSize != 0 {
		return false
	}

	// 2. 清除CurrentShuffledMediators原来的空间，重新分配空间
	ms.CurrentShuffledMediators = make([]common.Address, aSize)

	// 3. 初始化数据
	meds := gp.GetActiveMediators()
	copy(ms.CurrentShuffledMediators, meds)
	//for i, add := range meds {
	//	ms.CurrentShuffledMediators[i] = add
	//}

	// 4. 打乱证人的调度顺序
	shuffleMediators(ms.CurrentShuffledMediators, binary.BigEndian.Uint64(hash[8:]))

	err = pRep.StoreMediatorSchl(ms)
	if err != nil {
		log.Debugf("StoreMediatorSchl error:" + err.Error())
		return false
	}

	return true
}

func shuffleMediators(mediators []common.Address, seed uint64) {
	nowHi := seed << 32
	aSize := len(mediators)
	for i := 0; i < aSize; i++ {
		// 高性能随机生成器(High performance random generator)
		// 原理请参考 http://xorshift.di.unimi.it/
		k := nowHi + uint64(i)*2685821657736338717
		k ^= k >> 12
		k ^= k << 25
		k ^= k >> 27
		k *= 2685821657736338717

		jmax := uint64(aSize - i)
		j := uint64(i) + k%jmax

		// 进行N次随机交换
		mediators[i], mediators[j] = mediators[j], mediators[i]
	}
}

/**
@brief 获取给定的未来第slotNum个slot开始的时间。
Get the time at which the given slot occurs.

如果slotNum == 0，则返回time.Unix(0,0)。
If slotNum == 0, return time.Unix(0,0).

如果slotNum == N 且 N > 0，则返回大于UnitTime的第N个单元验证间隔的对齐时间
If slotNum == N for N > 0, return the Nth next unit-interval-aligned time greater than head_block_time().
*/
func (pRep *PropRepository) GetSlotTime(slotNum uint32) time.Time {
	if slotNum == 0 {
		return time.Unix(0, 0)
	}

	gp, err := pRep.RetrieveGlobalProp()
	if err != nil {
		log.Debugf("Retrieve Global Prop error: %v", err.Error())
		return time.Unix(0, 0)
	}
	dgp, err := pRep.RetrieveDynGlobalProp()
	if err != nil {
		log.Debugf("Retrieve Dyn Global Prop error: %v", err.Error())
		return time.Unix(0, 0)
	}

	interval := gp.ChainParameters.MediatorInterval
	gasToken := dagconfig.DagConfig.GetGasToken()
	_, idx, ts, _ := pRep.db.GetNewestUnit(gasToken)
	// 本条件是用来生产第一个unit
	if idx.Index == 0 {
		/**
		注：第一个unit的生产时间是在genesisTime加上一个unit生产间隔
		n.b. first Unit is at genesisTime plus one UnitInterval
		*/
		genesisTime := ts
		return time.Unix(genesisTime+int64(slotNum)*int64(interval), 0)
	}

	// 最近的unit的绝对slot
	var unitAbsSlot = ts / int64(interval)
	// 最近的时间槽起始时间
	headSlotTime := time.Unix(unitAbsSlot*int64(interval), 0)

	// 添加区块链网络参数修改维护时所需跳过的slot
	if dgp.MaintenanceFlag {
		slotNum += uint32(gp.ChainParameters.MaintenanceSkipSlots)
	}

	// 如果是维护周期的话，加上维护间隔时间
	// 如果不是，就直接加上unit的slot时间
	// "slot 0" is headSlotTime
	// "slot 1" is headSlotTime,
	// plus maintenance interval if head uint is a maintenance Unit
	// plus Unit interval if head uint is not a maintenance Unit
	return headSlotTime.Add(time.Second * time.Duration(slotNum) * time.Duration(interval))
}

/**
获取在给定时间或之前出现的最近一个slot。 Get the last slot which occurs AT or BEFORE the given time.
*/
func (pRep *PropRepository) GetSlotAtTime(when time.Time) uint32 {
	gp, err := pRep.RetrieveGlobalProp()
	if err != nil {
		log.Debugf("Retrieve Global Prop error: %v", err.Error())
		return 0
	}
	//dgp, err := pRep.RetrieveDynGlobalProp()
	//if err != nil {
	//	log.Debugf("Retrieve Dyn Global Prop error: %v", err.Error())
	//	return 0
	//}

	/**
	返回值是所有满足 GetSlotTime（N）<= when 中最大的N
	The return value is the greatest value N such that GetSlotTime( N ) <= when.
	如果都不满足，则返回 0
	If no such N exists, return 0.
	*/
	firstSlotTime := pRep.GetSlotTime(1)

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
func (pRep *PropRepository) GetScheduledMediator(slotNum uint32) common.Address {
	ms, _ := pRep.RetrieveMediatorSchl()
	dgp, _ := pRep.RetrieveDynGlobalProp()
	currentASlot := dgp.CurrentASlot + uint64(slotNum)
	csmLen := len(ms.CurrentShuffledMediators)
	if csmLen == 0 {
		log.Debug("The current number of shuffled mediators is 0!")
		return common.Address{}
	}

	// 由于创世单元不是有mediator生产，所以这里需要减1
	index := (currentASlot - 1) % uint64(csmLen)
	return ms.CurrentShuffledMediators[index]
}
