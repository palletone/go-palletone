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
	"fmt"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	common2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

type MemDag struct {
	token              modules.AssetId
	stableUnitHash     common.Hash
	stableUnitHeight   uint64
	lastMainChainUnit  *modules.Unit
	threshold          int
	orphanUnits        sync.Map
	orphanUnitsParants sync.Map
	chainUnits         sync.Map
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

	// append by albert·gou 用于通知群签名
	toGroupSignFeed  event.Feed
	toGroupSignScope event.SubscriptionScope
}

func (pmg *MemDag) Close() {
	pmg.toGroupSignScope.Close()
}

func (pmg *MemDag) SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription {
	return pmg.toGroupSignScope.Track(pmg.toGroupSignFeed.Subscribe(ch))
}

//
//type PartitionMemDag struct {
//	*MemDag
//	threshold int
//}

func (pmg *MemDag) SetStableThreshold(count int) {
	pmg.lock.Lock()
	defer pmg.lock.Unlock()
	pmg.threshold = count
}

//func NewPartitionMemDag(token modules.AssetId, threshold int, saveHeaderOnly bool, db ptndb.Database,
//	stableUnitRep common2.IUnitRepository, propRep common2.IPropRepository,
//	stableStateRep common2.IStateRepository) *PartitionMemDag {
//	return &PartitionMemDag{
//		MemDag:    NewMemDag(token, saveHeaderOnly, db, stableUnitRep, propRep, stableStateRep),
//		threshold: threshold,
//	}
//}

func NewMemDag(token modules.AssetId, threshold int, saveHeaderOnly bool, db ptndb.Database,
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
	log.Debugf("Init MemDag[%s], get last stable unit[%s] to set lastMainChainUnit", token.String(), stablehash.String())

	return &MemDag{
		token:              token,
		threshold:          threshold,
		ldbunitRep:         stableUnitRep,
		ldbPropRep:         propRep,
		tempdbunitRep:      trep,
		tempUtxoRep:        tutxoRep,
		tempStateRep:       tstateRep,
		tempPropRep:        tpropRep,
		tempdb:             tempdb,
		orphanUnits:        sync.Map{},
		orphanUnitsParants: sync.Map{},
		chainUnits:         sync.Map{},
		stableUnitHash:     stablehash,
		stableUnitHeight:   stbIndex.Index,
		lastMainChainUnit:  stableUnit,
		saveHeaderOnly:     saveHeaderOnly,

		ldbUnitProduceRep:  ldbUnitProduceRep,
		tempUnitProduceRep: tempUnitProduceRep,
	}
}

func (chain *MemDag) GetUnstableRepositories() (common2.IUnitRepository, common2.IUtxoRepository, common2.IStateRepository, common2.IPropRepository, common2.IUnitProduceRepository) {
	return chain.tempdbunitRep, chain.tempUtxoRep, chain.tempStateRep, chain.tempPropRep, chain.tempUnitProduceRep
}

func (chain *MemDag) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	//chain_units := chain.getChainUnits()
	//unit, has := chain_units[hash]
	//if has {
	//	return unit.Header(), nil
	//}
	return chain.tempdbunitRep.GetHeaderByHash(hash)
}
func (chain *MemDag) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	//unit, err := chain.getHeaderByNumber(number)
	//if err == nil {
	//	return unit, nil
	//}
	return chain.tempdbunitRep.GetHeaderByNumber(number)
}

func (chain *MemDag) getHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	chain_units := chain.getChainUnits()
	for _, unit := range chain_units {
		if index := unit.UnitHeader.ChainIndex(); index.Equal(number) {
			return unit.Header(), nil
		}
	}
	return nil, fmt.Errorf("the header[%s] not exist.", number.String())
}

func (chain *MemDag) SetUnitGroupSign(uHash common.Hash /*, groupPubKey []byte*/, groupSign []byte,
	txpool txspool.ITxPool) error {
	//1. Set this unit as stable
	unit, err := chain.getChainUnit(uHash)
	if err != nil {
		log.Debugf("get Chain Unit error: %v", err.Error())
		return err
	}

	if !(unit.NumberU64() > chain.stableUnitHeight) {
		return nil
	}

	chain.lock.Lock()
	defer chain.lock.Unlock()
	chain.setStableUnit(uHash, unit.NumberU64(), txpool)
	//2. Update unit.groupSign
	header := unit.Header()
	//header.GroupPubKey = groupPubKey
	header.GroupSign = groupSign
	log.Debugf("Try to update unit[%s] header group sign", uHash.String())

	// 下一个unit的群签名
	if err == nil {
		log.Debugf("sent toGroupSign event")
		go chain.toGroupSignFeed.Send(modules.ToGroupSignEvent{})
	}

	return chain.ldbunitRep.SaveHeader(header)
}

//设置某个单元和高度为稳定单元。设置后会更新当前的稳定单元，并将所有稳定单元写入到StableDB中，并且将ChainUnit中的稳定单元删除。
//然后基于最新的稳定单元，重建Tempdb数据库
func (chain *MemDag) setStableUnit(hash common.Hash, height uint64, txpool txspool.ITxPool) {
	log.Debugf("Set stable unit to %s,height:%d", hash.String(), height)
	stable_height := chain.stableUnitHeight
	stableCount := int(height - stable_height)
	if stableCount < 0 {
		log.Errorf("Current stable height is %d, impossible set stable height to %d", stable_height, height)
		return
	}
	newStableUnits := make([]*modules.Unit, stableCount)
	stbHash := hash
	for i := 0; i < stableCount; i++ {
		if u, has := chain.getChainUnits()[stbHash]; has {
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
	chain_units := chain.getChainUnits()
	for _, funit := range chain_units {
		if funit.NumberU64() <= height && funit.Hash() != hash {
			chain.removeUnitAndChildren(funit.Hash(), txpool)
		}
	}
	chain.saveUnitToDb(chain.ldbunitRep, chain.ldbUnitProduceRep, unit)

	if !chain.saveHeaderOnly && len(unit.Txs) > 0 {
		//log.Debugf("Set tx[%x] status to confirm in txpool", unit.Txs.GetTxIds())
		go txpool.SendStoredTxs(unit.Txs.GetTxIds())
	}
	log.Debugf("Remove unit[%s] from chainUnits", hash.String())
	//remove new stable unit
	chain.chainUnits.Delete(hash)
	//Set stable unit
	chain.stableUnitHash = hash
	chain.stableUnitHeight = height
	//remove too low orphan unit
	chain.removeLowOrphanUnit(unit.NumberU64(), txpool)
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
	units := chain.getChainUnits()
	// todo Albert·gou 待重做 优化逻辑
	for i := 0; i < unstableCount; i++ {
		u := units[ustbHash]
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

		if len(hs) >= chain.threshold {
			log.Debugf("Unit[%s] has enough confirm address count=%d, make it to stable.", ustbHash.String(), len(hs))
			chain.setStableUnit(ustbHash, u.NumberU64(), txpool)
			return true
		}
		//log.Debugf("Unstable unit[%s] has confirm address count: %d / %d", ustbHash.TerminalString(), len(hs), needAddrCount)

		ustbHash = u.ParentHash()[0]
	}
	return false
}

//清空Tempdb，然后基于稳定单元到最新主链单元的路径，构建新的Tempdb
func (chain *MemDag) rebuildTempdb() {
	log.Debugf("MemDag[%s] clear tempdb and rebuild data", chain.token.String())
	// 删除stable unit ,保留从stable unit 到 last unit 之间的数据。
	chain.tempdb.Clear()
	//if last_unit != nil {
	//	to_save_hash := chain.lastMainChainUnit.Hash()
	//	unstablecount := chain.lastMainChainUnit.NumberU64() - last_unit.NumberU64()
	//	// 保存last unit 到last main unit 之间的区块。
	//	for i := 0; i < int(unstablecount); i++ {
	//		u, has := chain.getChainUnits()[to_save_hash]
	//		if has {
	//			to_save_hash = u.ParentHash()[0]
	//			chain.saveUnitToDb(chain.tempdbunitRep, chain.tempUnitProduceRep, u)
	//		}
	//	}
	//	return
	//}
	unstableUnits := chain.getMainChainUnits()
	for _, unit := range unstableUnits {
		chain.saveUnitToDb(chain.tempdbunitRep, chain.tempUnitProduceRep, unit)
	}
}

//获得从稳定单元到最新单元的主链上的单元列表，从久到新排列
// todo 按assetid 返回
func (chain *MemDag) getMainChainUnits() []*modules.Unit {
	unstableCount := int(chain.lastMainChainUnit.NumberU64() - chain.stableUnitHeight)
	log.Debugf("Unstable unit count:%d", unstableCount)
	unstableUnits := make([]*modules.Unit, unstableCount)
	ustbHash := chain.lastMainChainUnit.Hash()
	chain_units := chain.getChainUnits()
	log.DebugDynamic(func() string {
		str := "chainUnits has unit:"
		for hash, _ := range chain_units {
			str += hash.String() + ";"
		}
		return str
	})
	for i := 0; i < unstableCount; i++ {
		u, ok := chain_units[ustbHash]
		if !ok {
			log.Errorf("chainUnits don't have unit[%s]", ustbHash.String())
		}
		unstableUnits[unstableCount-i-1] = u
		ustbHash = u.ParentHash()[0]
	}
	return unstableUnits
}

//判断当前设置是保存Header还是Unit，将对应的对象保存到Tempdb数据库
func (chain *MemDag) saveUnitToDb(unitRep common2.IUnitRepository, produceRep common2.IUnitProduceRepository, unit *modules.Unit) {
	log.Debugf("Save unit[%s] to db", unit.Hash().String())
	if chain.saveHeaderOnly {
		unitRep.SaveNewestHeader(unit.Header())
	} else {
		//chain.tempdbunitRep.SaveUnit(unit, false)
		produceRep.PushUnit(unit)
	}
}

//从ChainUnits集合中删除一个单元以及其所有子孙单元
func (chain *MemDag) removeUnitAndChildren(hash common.Hash, txpool txspool.ITxPool) {
	log.Debugf("Remove unit[%s] and it's children from chain unit", hash.String())
	chain_units := chain.getChainUnits()
	for h, unit := range chain_units {
		if h == hash {
			if txs := unit.Transactions(); len(txs) > 1 {
				txpool.ResetPendingTxs(txs)
			}
			chain.chainUnits.Delete(h)
			log.Debugf("Remove unit[%s] from chainUnits", hash.String())
		} else {
			if unit.ParentHash()[0] == hash {
				chain.removeUnitAndChildren(h, txpool)
			}
		}
	}
}

func (chain *MemDag) AddUnit(unit *modules.Unit, txpool txspool.ITxPool) error {
	start := time.Now()
	if unit == nil {
		return errors.ErrNullPoint
	}
	if unit.NumberU64() <= chain.stableUnitHeight {
		log.Infof("This unit is too old! Ignore it,stable unit height:%d, stable hash:%s", chain.stableUnitHeight, chain.stableUnitHash.String())
		return nil
	}
	chain_units := chain.getChainUnits()
	if _, has := chain_units[unit.Hash()]; has { // 不重复添加
		return nil
	}

	chain.lock.Lock()
	defer chain.lock.Unlock()
	err := chain.addUnit(unit, txpool)
	log.Debugf("MemDag[%s] AddUnit cost time: %v ,index: %d", chain.token.String(),
		time.Since(start), unit.NumberU64())

	if err == nil {
		// 下一个unit的群签名
		log.Debugf("sent toGroupSign event")
		go chain.toGroupSignFeed.Send(modules.ToGroupSignEvent{})
	}

	return err
}

func (chain *MemDag) addUnit(unit *modules.Unit, txpool txspool.ITxPool) error {
	parentHash := unit.ParentHash()[0]
	uHash := unit.Hash()

	if _, ok := chain.getChainUnits()[parentHash]; ok || parentHash == chain.stableUnitHash {
		//add unit to chain
		log.Debugf("chain[%p] Add unit[%s] to chainUnits", chain, uHash.String())
		//add at the end of main chain unit
		if parentHash == chain.lastMainChainUnit.Hash() {
			//Add a new unit to main chain
			// 判断当前高度是不是已经有区块了
			if _, err := chain.getHeaderByNumber(unit.Number()); err != nil {
				chain.setLastMainchainUnit(unit)
			} else {
				log.Infof("the chain is forked, save the equal units,fork:[%s]", uHash.String())
			}
			chain.chainUnits.Store(uHash, unit)
			//update txpool's tx status to pending
			if len(unit.Txs) > 0 {
				go txpool.SetPendingTxs(unit.Hash(), unit.NumberU64(), unit.Txs)
			}
			//增加了单元后检查是否满足稳定单元的条件
			// todo Albert·gou 待重做 优化逻辑
			if !chain.checkStableCondition(txpool) {
				//没有产生稳定单元，加入Tempdb
				chain.saveUnitToDb(chain.tempdbunitRep, chain.tempUnitProduceRep, unit)
				//这个单元不是稳定单元，需要加入Tempdb
			} else {
				// 下一个unit的群签名
				log.Debugf("sent toGroupSign event")
				go chain.toGroupSignFeed.Send(modules.ToGroupSignEvent{})

				log.Debugf("unit[%s] checkStableCondition =true", unit.Hash().String())
			}
		} else { //Fork unit
			chain.chainUnits.Store(uHash, unit)
			if unit.NumberU64() > chain.lastMainChainUnit.NumberU64() { //Need switch main chain
				//switch main chain, build db
				// 如果按节点产块的确认数来判断，可能会last_main_unit和new_unit的确认数相等，导致不切换主链。
				// todo 优化不被攻击
				chain.switchMainChain(unit, txpool)
			}
		}
		//orphan unit can add below this unit?
		if inter, has := chain.orphanUnitsParants.Load(uHash); has {
			chain.orphanUnitsParants.Delete(uHash)
			next_hash := inter.(common.Hash)
			chain.processOrphan(next_hash, txpool)
		}
	} else {
		//add unit to orphan
		log.Infof("This unit[%s] is an orphan unit", uHash.String())
		chain.orphanUnits.Store(uHash, unit)
		chain.orphanUnitsParants.Store(unit.ParentHash()[0], uHash)
	}
	return nil
}

//计算一个单元到稳定单元之间有多少个确认地址数
func (chain *MemDag) getChainAddressCount(lastUnit *modules.Unit) int {
	//token := lastUnit.Number().AssetID
	addrs := map[common.Address]bool{}
	unitHash := lastUnit.Hash()
	units := chain.getChainUnits()
	for unitHash != chain.stableUnitHash {
		unit := units[unitHash]
		addrs[unit.Author()] = true
		unitHash = unit.ParentHash()[0]
	}
	return len(addrs)
}

func (chain *MemDag) switchMainChain(newUnit *modules.Unit, txpool txspool.ITxPool) {
	oldLastMainchainUnit := chain.lastMainChainUnit
	old_last_unit_hash := oldLastMainchainUnit.Hash()
	log.Debugf("Switch main chain unit from %s to %s", old_last_unit_hash.String(), newUnit.Hash().String())
	chain.setLastMainchainUnit(newUnit)
	//reverse txpool tx status
	chain_units := chain.getChainUnits()
	main_chain_units := chain.getMainChainUnits()
	for _, m_u := range main_chain_units {
		hash := m_u.Hash()
		if _, has := chain_units[hash]; has {
			delete(chain_units, hash)
		}
	}
	// 不在主链上的区块，将已打包交易回滚。
	for _, un_unit := range chain_units {
		txs := un_unit.Transactions()
		if len(txs) > 1 {
			log.Debugf("Reset unit[%#x] 's txs status to not pending", un_unit.UnitHash)
			txpool.ResetPendingTxs(txs)
		}
	}
	//基于新主链，更新TxPool的状态
	var last_index int
	var to_set bool
	for i, unit := range main_chain_units { // 只需要更新上个最新单元到最新单元之间的数据
		if unit.Hash() == old_last_unit_hash {
			last_index = i
			to_set = true
		}
		if to_set && i > last_index {
			if len(unit.Txs) > 1 {
				log.Debugf("Update tx[%#x] status to pending in txpool", unit.Txs.GetTxIds())
				go txpool.SetPendingTxs(unit.Hash(), unit.NumberU64(), unit.Txs)
			}
		}
	}
	//基于新主链的单元和稳定单元，重新构建Tempdb
	chain.rebuildTempdb()
}

//枚举每一个孤儿单元，如果发现有单元的ParentHash是指定Hash，那么这说明这不再是一个孤儿单元，
//将其从孤儿单元列表中删除，并添加到ChainUnits中。
func (chain *MemDag) processOrphan(unitHash common.Hash, txpool txspool.ITxPool) {
	//for hash, unit := range chain.getOrphanUnits() {
	//	if unit.ParentHash()[0] == unitHash {
	//		log.Debugf("Orphan unit[%s] can add to chain now.", unit.Hash().String())
	//		chain.orphanUnitsParants.Delete(unitHash)
	//		chain.orphanUnits.Delete(hash)
	//		chain.addUnit(unit, txpool) //这个方法里面又会处理剩下的孤儿单元，从而形成递归
	//	}
	//}
	// 不用每次都循环
	unit, has := chain.getOrphanUnits()[unitHash]
	if has {
		log.Debugf("Orphan unit[%s] can add to chain now.", unit.Hash().String())
		chain.orphanUnits.Delete(unitHash)
		chain.addUnit(unit, txpool)
	}
}
func (chain *MemDag) getOrphanUnits() map[common.Hash]*modules.Unit {
	units := make(map[common.Hash]*modules.Unit)
	chain.orphanUnits.Range(func(k, v interface{}) bool {
		hash := k.(common.Hash)
		u := v.(*modules.Unit)
		u_hash := u.Hash()
		if hash != u_hash {
			chain.orphanUnits.Delete(hash)
			chain.orphanUnits.Store(u_hash, u)
		}
		units[u_hash] = u
		return true
	})
	return units
}
func (chain *MemDag) removeLowOrphanUnit(lessThan uint64, txpool txspool.ITxPool) {
	for hash, unit := range chain.getOrphanUnits() {
		if unit.NumberU64() <= lessThan {
			log.Debugf("Orphan unit[%s] height[%d] is too low, remove it.", unit.Hash().String(), unit.NumberU64())
			if txs := unit.Transactions(); len(txs) > 1 {
				txpool.ResetPendingTxs(txs)
			}
			chain.orphanUnits.Delete(hash)
		}
	}
}
func (chain *MemDag) getChainUnits() map[common.Hash]*modules.Unit {
	units := make(map[common.Hash]*modules.Unit)
	chain.chainUnits.Range(func(k, v interface{}) bool {
		hash := k.(common.Hash)
		u := v.(*modules.Unit)
		u_hash := u.Hash()
		if hash != u_hash {
			chain.chainUnits.Delete(hash)
			chain.chainUnits.Store(u_hash, u)
		}
		units[u_hash] = u
		return true
	})
	return units
}
func (chain *MemDag) getChainUnit(hash common.Hash) (*modules.Unit, error) {
	units := chain.getChainUnits()
	if units != nil {
		if unit, ok := units[hash]; ok {
			return unit, nil
		}
	}
	return nil, errors.ErrNotFound
}

func (chain *MemDag) GetLastMainChainUnit() *modules.Unit {
	chain.lock.RLock()
	defer chain.lock.RUnlock()
	return chain.lastMainChainUnit
}

//设置最新的主链单元，并更新PropDB
func (chain *MemDag) setLastMainchainUnit(unit *modules.Unit) {
	chain.lastMainChainUnit = unit
}

//查询所有不稳定单元（不包括孤儿单元）
func (chain *MemDag) GetChainUnits() map[common.Hash]*modules.Unit {
	return chain.getChainUnits()
}
