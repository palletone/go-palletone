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

package memunit

import (
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	common2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

type MemDag struct {
	token             modules.AssetId
	stableUnitHash    common.Hash
	stableUnitHeight  uint64
	lastMainChainUnit *modules.Unit
	//threshold          int
	orphanUnits        map[common.Hash]*modules.Unit
	chainUnits         map[common.Hash]*modules.Unit
	tempdbunitRep      common2.IUnitRepository
	tempUtxoRep        common2.IUtxoRepository
	tempStateRep       common2.IStateRepository
	tempPropRep        common2.IPropRepository
	tempUnitProduceRep common2.IUnitProduceRepository

	ldbunitRep        common2.IUnitRepository
	ldbPropRep        common2.IPropRepository
	ldbUnitProduceRep common2.IUnitProduceRepository
	tempdb            *Tempdb
	saveHeaderOnly    bool
	lock              sync.RWMutex
}

type PartitionMemDag struct {
	*MemDag
	threshold int
}

func (pmg *PartitionMemDag) SetStableThreshold(count int) {
	pmg.threshold = count
}

func NewPartitionMemDag(token modules.AssetId, threshold int, saveHeaderOnly bool, db ptndb.Database,
	stableUnitRep common2.IUnitRepository, propRep common2.IPropRepository,
	stableStateRep common2.IStateRepository) *PartitionMemDag {
	return &PartitionMemDag{
		MemDag:    NewMemDag(token, saveHeaderOnly, db, stableUnitRep, propRep, stableStateRep),
		threshold: threshold,
	}
}

func NewMemDag(token modules.AssetId /*, threshold int*/, saveHeaderOnly bool, db ptndb.Database,
	stableUnitRep common2.IUnitRepository, propRep common2.IPropRepository,
	stableStateRep common2.IStateRepository) *MemDag {
	tempdb, _ := NewTempdb(db)
	trep := common2.NewUnitRepository4Db(tempdb)
	tutxoRep := common2.NewUtxoRepository4Db(tempdb)
	tstateRep := common2.NewStateRepository4Db(tempdb)
	tpropRep := common2.NewPropRepository4Db(tempdb)
	tempUnitProduceRep := common2.NewUnitProduceRepository(trep, tpropRep, tstateRep)
	ldbUnitProduceRep := common2.NewUnitProduceRepository(stableUnitRep, propRep, stableStateRep)
	stablehash, stbIndex, err := propRep.GetNewestUnit(token)
	if err != nil {
		log.Errorf("Cannot retrieve last stable unit from db for token:%s, you forget 'gptn init'??", token.String())
		return nil
	}
	var stableUnit *modules.Unit
	if saveHeaderOnly {
		header, err := stableUnitRep.GetHeaderByHash(stablehash)
		if err != nil {
			log.Errorf("Cannot retrieve last stable unit from db by hash[%s]", stablehash.String())
			return nil
		}
		stableUnit = modules.NewUnit(header, nil)
	} else {
		stableUnit, err = stableUnitRep.GetUnit(stablehash)
		if err != nil {
			log.Errorf("Cannot retrieve last stable unit from db by hash[%s]", stablehash.String())
			return nil
		}
	}
	log.Debugf("Init MemDag, get last stable unit[%s] to set lastMainChainUnit", stablehash.String())

	return &MemDag{
		token: token,
		//threshold:         threshold,
		ldbunitRep:        stableUnitRep,
		ldbPropRep:        propRep,
		tempdbunitRep:     trep,
		tempUtxoRep:       tutxoRep,
		tempStateRep:      tstateRep,
		tempPropRep:       tpropRep,
		tempdb:            tempdb,
		orphanUnits:       make(map[common.Hash]*modules.Unit),
		chainUnits:        make(map[common.Hash]*modules.Unit),
		stableUnitHash:    stablehash,
		stableUnitHeight:  stbIndex.Index,
		lastMainChainUnit: stableUnit,
		saveHeaderOnly:    saveHeaderOnly,

		ldbUnitProduceRep: ldbUnitProduceRep,

		tempUnitProduceRep: tempUnitProduceRep,
	}
}

func (chain *MemDag) GetUnstableRepositories() (common2.IUnitRepository, common2.IUtxoRepository, common2.IStateRepository, common2.IPropRepository, common2.IUnitProduceRepository) {
	return chain.tempdbunitRep, chain.tempUtxoRep, chain.tempStateRep, chain.tempPropRep, chain.tempUnitProduceRep
}
func (chain *MemDag) SetUnitGroupSign(uHash common.Hash, groupPubKey []byte, groupSign []byte, txpool txspool.ITxPool) error {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	//1. Set this unit as stable
	unit, err := chain.getChainUnit(uHash)
	if err != nil {
		return err
	}
	chain.SetStableUnit(uHash, unit.NumberU64(), txpool)
	//2. Update unit.groupSign
	header := unit.Header()
	header.GroupPubKey = groupPubKey
	header.GroupSign = groupSign
	log.Debugf("Try to update unit[%s] header group sign", uHash.String())
	return chain.ldbunitRep.SaveHeader(header)
}

//设置某个单元和高度为稳定单元。设置后会更新当前的稳定单元，并将所有稳定单元写入到StableDB中，并且将ChainUnit中的稳定单元删除。
//然后基于最新的稳定单元，重建Tempdb数据库
func (chain *MemDag) SetStableUnit(hash common.Hash, height uint64, txpool txspool.ITxPool) {
	//oldStableHash := chain.stableUnitHash
	//u, has := chain.chainUnits[hash]
	//if !has {
	//	return
	//}
	//token := u.Number().AssetID
	log.Debugf("Set stable unit to %s,height:%d", hash.String(), height)
	stableCount := int(height - chain.stableUnitHeight)
	newStableUnits := make([]*modules.Unit, stableCount)
	stbHash := hash
	for i := 0; i < stableCount; i++ {
		if u, has := chain.chainUnits[stbHash]; has {
			newStableUnits[stableCount-i-1] = u
			stbHash = u.ParentHash()[0]
		}
	}
	//Save stable unit and it's parent
	for _, unit := range newStableUnits {
		chain.setNextStableUnit(unit, txpool)
	}
	// Rebuild temp db
	chain.rebuildTempdb()
}

//设置当前稳定单元的指定子单元为稳定单元
func (chain *MemDag) setNextStableUnit(unit *modules.Unit, txpool txspool.ITxPool) {
	hash := unit.Hash()
	height := unit.NumberU64()
	//remove fork units
	for _, funit := range chain.chainUnits {
		if funit.NumberU64() <= height && funit.Hash() != hash {
			chain.removeUnitAndChildren(funit.Hash())
		}
	}
	if chain.saveHeaderOnly {
		chain.ldbunitRep.SaveHeader(unit.UnitHeader)
		chain.ldbPropRep.SetNewestUnit(unit.UnitHeader)
	} else {
		//Save stable unit to ldb
		//chain.ldbunitRep.SaveUnit(unit, false)
		chain.ldbUnitProduceRep.PushUnit(unit)
		//txpool flag tx as packaged
		if len(unit.Txs) > 0 {
			log.Debugf("Set tx[%x] status to confirm", unit.Txs.GetTxIds())
			txpool.SendStoredTxs(unit.Txs.GetTxIds())
		}
	}
	log.Debugf("Remove unit[%s] from chainUnits", hash.String())
	//remove new stable unit
	delete(chain.chainUnits, hash)
	//Set stable unit
	chain.stableUnitHash = hash
	chain.stableUnitHeight = height
	//remove too low orphan unit
	chain.removeLowOrphanUnit(unit.NumberU64())
}

//判断当前主链上的单元是否有满足稳定单元的确认数，如果有，则更新稳定单元，并重建Temp数据库，返回True
// 如果没有，则不进行任何操作，返回False
func (chain *MemDag) checkStableCondition(txpool txspool.ITxPool) bool {
	unit := chain.lastMainChainUnit
	unstableCount := int(unit.NumberU64() - chain.stableUnitHeight)
	//每个单元被多少个地址确认过(包括自己)
	unstableCofirmAddrs := make(map[common.Hash]map[common.Address]bool)
	childrenCofirmAddrs := make(map[common.Address]bool)
	ustbHash := unit.Hash()
	childrenCofirmAddrs[unit.Author()] = true
	for i := 0; i < unstableCount; i++ {
		u := chain.chainUnits[ustbHash]
		hs := unstableCofirmAddrs[ustbHash]
		if hs == nil {
			hs = make(map[common.Address]bool)
			unstableCofirmAddrs[ustbHash] = hs
		}
		hs[u.Author()] = true
		for addr := range childrenCofirmAddrs {
			hs[addr] = true
		}
		childrenCofirmAddrs[u.Author()] = true
		// todo threshold 换届后可能会变
		threshold, _ := chain.ldbPropRep.GetChainThreshold()
		if len(hs) >= threshold {
			log.Debugf("Unit[%s] has enough confirm address count=%d, make it to stable.", ustbHash.String(), len(hs))
			chain.SetStableUnit(ustbHash, u.NumberU64(), txpool)

			return true
		}
		//log.Debugf("Unstable unit[%s] has confirm address count: %d / %d", ustbHash.TerminalString(), len(hs), needAddrCount)

		ustbHash = u.ParentHash()[0]
	}

	return false
}

//清空Tempdb，然后基于稳定单元到最新主链单元的路径，构建新的Tempdb
func (chain *MemDag) rebuildTempdb() {
	log.Debugf("Clear tempdb and rebuild data")
	chain.tempdb.Clear()
	unstableUnits := chain.getMainChainUnits()
	for _, unit := range unstableUnits {
		chain.saveUnit2TempDb(unit)
	}
}

//获得从稳定单元到最新单元的主链上的单元列表，从久到新排列
// todo 按assetid 返回
func (chain *MemDag) getMainChainUnits() []*modules.Unit {
	unstableCount := int(chain.lastMainChainUnit.NumberU64() - chain.stableUnitHeight)
	log.Debugf("Unstable unit count:%d", unstableCount)
	unstableUnits := make([]*modules.Unit, unstableCount)
	ustbHash := chain.lastMainChainUnit.Hash()
	log.DebugDynamic(func() string {
		str := "chainUnits has unit:"
		for hash, _ := range chain.chainUnits {
			str += hash.String() + ";"
		}
		return str
	})
	for i := 0; i < unstableCount; i++ {
		u, ok := chain.chainUnits[ustbHash]
		if !ok {
			log.Errorf("chainUnits don't have unit[%s]", ustbHash.String())
		}
		unstableUnits[unstableCount-i-1] = u
		//log.Debugf("Unstable unit:%#v, Hash:%s", u, ustbHash.String())
		ustbHash = u.ParentHash()[0]
	}
	return unstableUnits
}

//判断当前设置是保存Header还是Unit，将对应的对象保存到Tempdb数据库
func (chain *MemDag) saveUnit2TempDb(unit *modules.Unit) {
	log.Debugf("Save unit[%s] to tempdb", unit.Hash().String())
	if chain.saveHeaderOnly {
		chain.tempdbunitRep.SaveHeader(unit.Header())
	} else {
		//chain.tempdbunitRep.SaveUnit(unit, false)
		chain.tempUnitProduceRep.PushUnit(unit)
	}
}

//从ChainUnits集合中删除一个单元以及其所有子孙单元
func (chain *MemDag) removeUnitAndChildren(hash common.Hash) {
	log.Debugf("Remove unit[%s] and it's children from chain unit", hash.String())
	for h, unit := range chain.chainUnits {
		if h == hash {
			delete(chain.chainUnits, h)
			log.Debugf("Remove unit[%s] from chainUnits", hash.String())
		} else {
			if unit.ParentHash()[0] == hash {
				chain.removeUnitAndChildren(h)
			}
		}
	}
}

func (chain *MemDag) AddUnit(unit *modules.Unit, txpool txspool.ITxPool) error {
	defer func(start time.Time) {
		log.Debugf("MemDag[%s] AddUnit cost time: %v ,index: %d", dagconfig.DagConfig.GetGasToken().String(),
			time.Since(start), unit.NumberU64())
	}(time.Now())

	if unit == nil {
		return errors.ErrNullPoint
	}
	//token := unit.Number().AssetID
	if unit.NumberU64() <= chain.stableUnitHeight {
		log.Infof("This unit is too old! Ignore it,Stable unit height:%d", chain.stableUnitHeight)
		return nil
	}
	chain.lock.Lock()
	defer chain.lock.Unlock()
	return chain.addUnit(unit, txpool)
}
func (chain *MemDag) addUnit(unit *modules.Unit, txpool txspool.ITxPool) error {
	parentHash := unit.ParentHash()[0]
	uHash := unit.Hash()
	//threshold, _ := chain.ldbPropRep.GetChainThreshold()
	//token := unit.Number().AssetID
	if _, ok := chain.chainUnits[parentHash]; ok || parentHash == chain.stableUnitHash {
		//add unit to chain
		log.Debugf("chain[%p] Add unit[%s] to chainUnits", chain, uHash.String())
		chain.chainUnits[uHash] = unit
		//add at the end of main chain unit
		if parentHash == chain.lastMainChainUnit.Hash() {
			//Add a new unit to main chain
			chain.setLastMainchainUnit(unit)
			//update txpool's tx status to pending
			if len(unit.Txs) > 0 {
				log.Debugf("Update tx[%#x] status to pending in txpool", unit.Txs.GetTxIds())
				txpool.SetPendingTxs(unit.Hash(), unit.Txs)
			}
			//增加了单元后检查是否满足稳定单元的条件
			if !chain.checkStableCondition(txpool) {
				chain.saveUnit2TempDb(unit)
				//这个单元不是稳定单元，需要加入Tempdb
			} else {
				log.Debugf("unit[%s] checkStableCondition =true", unit.Hash().String())
			}
		} else { //Fork unit
			if unit.NumberU64() > chain.lastMainChainUnit.NumberU64() { //Need switch main chain
				//switch main chain, build db
				//如果分支上的确认数大于等于当前主链，则切换主链
				oldMainchainAddrCount := chain.getChainAddressCount(chain.lastMainChainUnit)
				forkChainAddrCount := chain.getChainAddressCount(unit)
				if forkChainAddrCount >= oldMainchainAddrCount {
					chain.switchMainChain(unit, txpool)
				} else {
					log.Infof("Unit[%s] is in fork chain, and address count=%d, less than main chain address count=%d", unit.Hash().String(), forkChainAddrCount, oldMainchainAddrCount)
				}
			}
		}

		//orphan unit can add below this unit?
		chain.processOrphan(uHash, txpool)
	} else {
		//add unit to orphan
		log.Infof("This unit[%s] is an orphan unit", uHash.String())
		chain.orphanUnits[uHash] = unit
	}
	return nil
}

//计算一个单元到稳定单元之间有多少个确认地址数
func (chain *MemDag) getChainAddressCount(lastUnit *modules.Unit) int {
	//token := lastUnit.Number().AssetID
	addrs := map[common.Address]bool{}
	unitHash := lastUnit.Hash()
	for unitHash != chain.stableUnitHash {
		unit := chain.chainUnits[unitHash]
		addrs[unit.Author()] = true
		unitHash = unit.ParentHash()[0]
	}
	return len(addrs)
}

func (chain *MemDag) switchMainChain(newUnit *modules.Unit, txpool txspool.ITxPool) {
	//token := newUnit.Number().AssetID
	oldLastMainchainUnit := chain.lastMainChainUnit
	log.Debugf("Switch main chain unit from %s to %s", oldLastMainchainUnit.Hash().String(), newUnit.Hash().String())

	//reverse txpool tx status
	unstableUnits := chain.getMainChainUnits()
	for _, unit := range unstableUnits {
		if unit.Hash() != oldLastMainchainUnit.Hash() {
			txs := unit.Transactions()
			if len(txs) > 0 {
				log.Debugf("Reset unit[%#x] 's txs status to not pending", unit.UnitHash)
				txpool.ResetPendingTxs(txs)
			}
		}
	}
	chain.setLastMainchainUnit(newUnit)
	//基于新主链，更新TxPool的状态
	newUnstableUnits := chain.getMainChainUnits()
	for _, unit := range newUnstableUnits {
		if len(unit.Txs) > 0 {
			log.Debugf("Update tx[%#x] status to pending in txpool", unit.Txs.GetTxIds())
			txpool.SetPendingTxs(unit.Hash(), unit.Txs)
		}
	}
	//基于新主链的单元和稳定单元，重新构建Tempdb
	chain.rebuildTempdb()
}

//枚举每一个孤儿单元，如果发现有单元的ParentHash是指定Hash，那么这说明这不再是一个孤儿单元，
//将其从孤儿单元列表中删除，并添加到ChainUnits中。
func (chain *MemDag) processOrphan(unitHash common.Hash, txpool txspool.ITxPool) {
	for hash, unit := range chain.orphanUnits {
		if unit.ParentHash()[0] == unitHash {
			log.Debugf("Orphan unit[%s] can add to chain now.", unit.Hash().String())
			delete(chain.orphanUnits, hash)
			chain.addUnit(unit, txpool) //这个方法里面又会处理剩下的孤儿单元，从而形成递归
			break
		}
	}
}
func (chain *MemDag) removeLowOrphanUnit(lessThan uint64) {
	for hash, unit := range chain.orphanUnits {
		if unit.NumberU64() <= lessThan {
			log.Debugf("Orphan unit[%s] height[%d] is too low, remove it.", unit.Hash().String(), unit.NumberU64())
			delete(chain.orphanUnits, hash)
		}
	}
}
func (chain *MemDag) getChainUnit(hash common.Hash) (*modules.Unit, error) {
	if unit, ok := chain.chainUnits[hash]; ok {
		return unit, nil
	}
	return nil, errors.ErrNotFound
}

//func (chain *MemDag) Exists(uHash common.Hash) bool {
//	_, ok := chain.chainUnits[uHash]
//	return ok
//}
func (chain *MemDag) GetLastMainChainUnit() *modules.Unit {
	chain.lock.RLock()
	defer chain.lock.RUnlock()
	return chain.lastMainChainUnit
}

//设置最新的主链单元，并更新PropDB
func (chain *MemDag) setLastMainchainUnit(unit *modules.Unit) {
	//token := unit.Number().AssetID
	//if chain.lastMainChainUnit == nil {
	//	chain.lastMainChainUnit = make(map[modules.AssetId]*modules.Unit)
	//}
	chain.lastMainChainUnit = unit
	//chain.ldbPropRep.SetNewestUnit(unit.Header())
}

//设置最新的稳定单元，并更新PropDB
//func (chain *MemDag) setStableUnit(unit *modules.Unit) {
//	//if chain.stableUnitHash == nil {
//	//	chain.stableUnitHash = make(map[modules.AssetId]common.Hash)
//	//}
//	//if chain.stableUnitHeight == nil {
//	//	chain.stableUnitHeight = make(map[modules.AssetId]uint64)
//	//}
//	//token := unit.Number().AssetID
//	hash := unit.Hash()
//	index := unit.NumberU64()
//	chain.stableUnitHash = hash
//	chain.stableUnitHeight = index
//	chain.ldbPropRep.SetNewestUnit(unit.UnitHeader)
//}

//查询所有不稳定单元（不包括孤儿单元）
func (chain *MemDag) GetChainUnits() map[common.Hash]*modules.Unit {
	chain.lock.RLock()
	defer chain.lock.RUnlock()
	return chain.chainUnits
}
