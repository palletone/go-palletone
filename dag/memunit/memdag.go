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
	"github.com/palletone/go-palletone/tokenengine"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	common2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/validator"
	"go.dedis.ch/kyber/v3/sign/bls"
)

type MemDag struct {
	token              modules.AssetId
	stableUnitHash     common.Hash
	stableUnitHeight   uint64
	lastMainChainUnit  *modules.Unit
	threshold          int
	height_hashs       sync.Map
	orphanUnits        sync.Map
	orphanUnitsParants sync.Map
	chainUnits         sync.Map
	tempdb             sync.Map

	ldbunitRep        common2.IUnitRepository
	ldbPropRep        common2.IPropRepository
	ldbUnitProduceRep common2.IUnitProduceRepository
	saveHeaderOnly    bool
	lock              sync.RWMutex
	cache             palletcache.ICache
	// append by albert·gou 用于通知群签名
	toGroupSignFeed  event.Feed
	toGroupSignScope event.SubscriptionScope
	db               ptndb.Database
	tokenEngine      tokenengine.ITokenEngine
	quit             chan struct{} // used for exit
}

func (pmg *MemDag) Close() {
	pmg.toGroupSignScope.Close()
}

func (pmg *MemDag) SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription {
	return pmg.toGroupSignScope.Track(pmg.toGroupSignFeed.Subscribe(ch))
}

func (pmg *MemDag) SetStableThreshold(count int) {
	pmg.lock.Lock()
	defer pmg.lock.Unlock()
	pmg.threshold = count
}

func NewMemDag(token modules.AssetId, threshold int, saveHeaderOnly bool, db ptndb.Database,
	stableUnitRep common2.IUnitRepository, propRep common2.IPropRepository,
	stableStateRep common2.IStateRepository, cache palletcache.ICache,
	tokenEngine tokenengine.ITokenEngine) *MemDag {
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
	memdag := &MemDag{
		token:              token,
		threshold:          threshold,
		ldbunitRep:         stableUnitRep,
		ldbPropRep:         propRep,
		tempdb:             sync.Map{},
		height_hashs:       sync.Map{},
		orphanUnits:        sync.Map{},
		orphanUnitsParants: sync.Map{},
		chainUnits:         sync.Map{},
		stableUnitHash:     stablehash,
		stableUnitHeight:   stbIndex.Index,
		lastMainChainUnit:  stableUnit,
		saveHeaderOnly:     saveHeaderOnly,
		cache:              cache,
		ldbUnitProduceRep:  ldbUnitProduceRep,
		db:                 db,
		tokenEngine:        tokenEngine,
	}
	temp, _ := NewChainTempDb(db, cache, tokenEngine, saveHeaderOnly)
	temp.Unit = stableUnit
	memdag.tempdb.Store(stablehash, temp)
	memdag.chainUnits.Store(stablehash, temp)

	go memdag.loopRebuildTmpDb()
	return memdag
}
func (chain *MemDag) loopRebuildTmpDb() {
	rebuild := time.NewTicker(10 * time.Minute)
	defer rebuild.Stop()
	for {
		select {
		case <-rebuild.C:
			if chain.lastMainChainUnit.Hash() == chain.stableUnitHash || len(chain.getChainUnits()) <= 1 {
				// temp db don't need rebuild.
				continue
			}
			chain.lock.Lock()
			chain.rebuildTempdb()
			chain.lock.Unlock()
		case <-chain.quit:
			return
		}
	}
}
func (chain *MemDag) GetUnstableRepositories() (common2.IUnitRepository, common2.IUtxoRepository,
	common2.IStateRepository, common2.IPropRepository, common2.IUnitProduceRepository) {
	last_main_hash := chain.lastMainChainUnit.Hash()
	temp_rep, err := chain.getChainUnit(last_main_hash)
	if err != nil { // 重启后memdag的chainUnits还清被清空，需要重新以memdag的db构建unstable repositoreis
		temp_inter, has := chain.tempdb.Load(last_main_hash)
		if !has {
			log.Errorf("the last_unit: %s , is not exist in memdag", last_main_hash.String())
			tempdb, _ := NewTempdb(chain.db)
			trep := common2.NewUnitRepository4Db(tempdb, chain.tokenEngine)
			tutxoRep := common2.NewUtxoRepository4Db(tempdb, chain.tokenEngine)
			tstateRep := common2.NewStateRepository4Db(tempdb)
			tpropRep := common2.NewPropRepository4Db(tempdb)
			tunitProduceRep := common2.NewUnitProduceRepository(trep, tpropRep, tstateRep)
			return trep, tutxoRep, tstateRep, tpropRep, tunitProduceRep
		}
		tempdb := temp_inter.(*ChainTempDb)
		return tempdb.UnitRep, tempdb.UtxoRep, tempdb.StateRep, tempdb.PropRep, tempdb.UnitProduceRep
	}
	return temp_rep.UnitRep, temp_rep.UtxoRep, temp_rep.StateRep, temp_rep.PropRep, temp_rep.UnitProduceRep
}

func (chain *MemDag) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	if inter, has := chain.tempdb.Load(chain.lastMainChainUnit.Hash()); has {
		temp := inter.(*ChainTempDb)
		return temp.UnitRep.GetHeaderByHash(hash)
	}
	return nil, errors.New("not found")
}
func (chain *MemDag) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	if inter, has := chain.tempdb.Load(chain.lastMainChainUnit.Hash()); has {
		temp := inter.(*ChainTempDb)
		return temp.UnitRep.GetHeaderByNumber(number)
	}
	return nil, errors.New("not found")
}

//func (chain *MemDag) getHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
//	chain_units := chain.getChainUnits()
//	for _, unit := range chain_units {
//		if index := unit.UnitHeader.ChainIndex(); index.Equal(number) {
//			return unit.Header(), nil
//		}
//	}
//	return nil, fmt.Errorf("the header[%s] not exist.", number.String())
//}

func (chain *MemDag) SetUnitGroupSign(uHash common.Hash /*, groupPubKey []byte*/, groupSign []byte,
	txpool txspool.ITxPool) error {
	//1. Set this unit as stable
	unit_temp, err := chain.getChainUnit(uHash)
	if err != nil {
		log.Debugf("get Chain Unit error: %v", err.Error())
		return err
	}
	unit := unit_temp.Unit
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
	log.Debugf("Try to update unit[%s] header group sign, and send go groupSign event.", uHash.String())
	// 进行下一个unit的群签名
	go chain.toGroupSignFeed.Send(modules.ToGroupSignEvent{})

	return chain.ldbunitRep.SaveHeader(header)
}

//设置某个单元和高度为稳定单元。设置后会更新当前的稳定单元，并将所有稳定单元写入到StableDB中，并且将ChainUnit中的稳定单元删除。
//然后基于最新的稳定单元，重建Tempdb数据库
func (chain *MemDag) setStableUnit(hash common.Hash, height uint64, txpool txspool.ITxPool) {
	tt := time.Now()
	log.Debugf("Set stable unit to %s,height:%d", hash.String(), height)
	stable_height := chain.stableUnitHeight
	stableCount := int(height - stable_height)
	if stableCount <= 0 {
		log.Errorf("Current stable height is %d, impossible set stable height to %d", stable_height, height)
		return
	}
	newStableUnits := make([]*modules.Unit, stableCount)
	stbHash := hash
	chain_units := chain.getChainUnits()
	for i := 0; i < stableCount; i++ {
		if u, has := chain_units[stbHash]; has {
			newStableUnits[stableCount-i-1] = u
			stbHash = u.ParentHash()[0]
		}
	}
	//Save stable unit and it's parent
	max_height := height
	for _, unit := range newStableUnits {
		if unit.NumberU64() > max_height {
			max_height = unit.NumberU64()
		}
		chain.setNextStableUnit(unit, txpool)

	}
	// 更新tempdb ，将低于稳定单元的分叉链都删除
	go chain.delHeightUnitsAndTemp(max_height)
	log.DebugDynamic(func() string {
		return fmt.Sprintf("set next stable unit cost time: %s ,index: %d, hash: %s",
			time.Since(tt), height, hash.String())
	})
	//remove fork units, and remove lower than stable unit
	for _, funit := range chain_units {
		if funit.NumberU64() <= max_height && funit.Hash() != hash {
			chain.removeUnitAndChildren(funit.Hash(), txpool)
		}
	}
	//remove too low orphan unit
	go chain.removeLowOrphanUnit(max_height, txpool)
}

//设置当前稳定单元的指定父单元为稳定单元
func (chain *MemDag) setNextStableUnit(unit *modules.Unit, txpool txspool.ITxPool) {
	hash := unit.Hash()
	height := unit.NumberU64()
	// memdag不依赖apply unit的存储，因此用协程提高setStable的效率
	// 虽然与memdag无关，但是下一个unit的 apply 处理依赖上一个unit apply的结果，所以不能用协程并发处理
	chain.saveUnitToDb(chain.ldbunitRep, chain.ldbUnitProduceRep, unit)
	if !chain.saveHeaderOnly && len(unit.Txs) > 1 {
		go txpool.SendStoredTxs(unit.Txs.GetTxIds())
	}
	log.Debugf("Remove unit[%s] from chainUnits", hash.String())
	//remove new stable unit
	chain.chainUnits.Delete(hash)
	//Set stable unit
	chain.stableUnitHash = hash
	chain.stableUnitHeight = height
}

func (chain *MemDag) checkUnitIrreversibleWithGroupSign(unit *modules.Unit) bool {
	if unit.GetGroupPubKeyByte() == nil || unit.GetGroupSign() == nil {
		return false
	}

	pubKey, err := unit.GetGroupPubKey()
	if err != nil {
		log.Debug(err.Error())
		return false
	}

	err = bls.Verify(core.Suite, pubKey, unit.UnitHash[:], unit.GetGroupSign())
	if err != nil {
		log.Debug(err.Error())
		return false
	}

	return true
}

// 判断当前主链上的单元是否有满足稳定单元的确认数，如果有，则更新稳定单元，并重建Temp数据库，返回True
// 如果没有，则不进行任何操作，返回False
func (chain *MemDag) checkStableCondition(unit *modules.Unit, txpool txspool.ITxPool) bool {
	// append by albert, 使用群签名判断是否稳定
	if chain.checkUnitIrreversibleWithGroupSign(unit) {
		log.Debugf("the unit(%s) have group sign(%s), make it to irreversible.",
			unit.UnitHash.TerminalString(), hexutil.Encode(unit.GetGroupSign()))
		chain.setStableUnit(unit.UnitHash, unit.NumberU64(), txpool)
		return true
	}

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
		if u == nil {
			continue
		}
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
			log.Debugf("Unit[%s] height:%d has enough confirm address count=%d, make it to stable.",
				ustbHash.String(), unit.NumberU64(), len(hs))
			chain.setStableUnit(ustbHash, u.NumberU64(), txpool)
			return true
		}
		ustbHash = u.ParentHash()[0]
	}
	return false
}

//清空主链的Tempdb，然后基于稳定单元到最新主链单元的路径，构建新的Tempdb
func (chain *MemDag) rebuildTempdb() {
	to_del := make([]common.Hash, 0)
	chain.tempdb.Range(func(k, v interface{}) bool {
		hash := k.(common.Hash)
		if unit_temp, err := chain.getChainUnit(hash); err == nil {
			if num := unit_temp.Unit.NumberU64(); num < chain.stableUnitHeight {
				to_del = append(to_del, hash)
			} else if num == chain.stableUnitHeight {
				if unit_temp.Unit.Hash() != chain.stableUnitHash {
					to_del = append(to_del, hash)
				}
			}
		}
		return true
	})
	for _, h := range to_del {
		inter, ok := chain.tempdb.Load(h)
		if ok {
			temp := inter.(*ChainTempDb)
			temp.Tempdb.Clear()
		}
		chain.tempdb.Delete(h)
	}
}

//获得从稳定单元到最新单元的主链上的单元列表，从久到新排列
func (chain *MemDag) getMainChainUnits() []*modules.Unit {
	unstableCount := int(chain.lastMainChainUnit.NumberU64() - chain.stableUnitHeight)
	log.Debugf("Unstable unit count:%d", unstableCount)
	unstableUnits := make([]*modules.Unit, unstableCount)
	ustbHash := chain.lastMainChainUnit.Hash()
	chain_units := chain.getChainUnits()
	log.DebugDynamic(func() string {
		str := "chainUnits has unit:"
		for hash := range chain_units {
			str += hash.String() + ";"
		}
		return str
	})
	for i := 0; i < unstableCount; i++ {
		u, ok := chain_units[ustbHash]
		if !ok {
			log.Errorf("chainUnits don't have unit[%s], last_main[%s]",
				ustbHash.String(), chain.lastMainChainUnit.Hash().String())
			continue
		}
		unstableUnits[unstableCount-i-1] = u
		ustbHash = u.ParentHash()[0]
	}
	return unstableUnits
}

func (chain *MemDag) getForkUnits(unit *modules.Unit) []*modules.Unit {
	chain_units := chain.getChainUnits()
	unstableCount := int(unit.NumberU64() - chain.stableUnitHeight)
	if unstableCount <= 1 {
		return append(make([]*modules.Unit, 0), unit)
	}
	hash := unit.ParentHash()[0]
	fork_len := unstableCount - 1
	unstableUnits := make([]*modules.Unit, fork_len)
	for i := 0; i < fork_len; i++ {
		u, ok := chain_units[hash]
		if !ok {
			log.Errorf("getforks chainUnits don't have unit[%s], last_main[%s]",
				hash.String(), chain.lastMainChainUnit.Hash().String())
		}
		unstableUnits[fork_len-i-1] = u
		hash = u.ParentHash()[0]
	}
	return append(unstableUnits, unit)
}

//判断当前设置是保存Header还是Unit，将对应的对象保存到Tempdb数据库
func (chain *MemDag) saveUnitToDb(unitRep common2.IUnitRepository, produceRep common2.IUnitProduceRepository,
	unit *modules.Unit) {
	log.Debugf("Save unit[%s] to db", unit.Hash().String())
	if chain.saveHeaderOnly {
		unitRep.SaveNewestHeader(unit.Header())
	} else {
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
				go txpool.ResetPendingTxs(txs)
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
func (chain *MemDag) AddStableUnit(unit *modules.Unit) {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	hash := unit.Hash()
	log.Debugf("add stable unit to dag, hash[%s], index:%d", hash.String(), unit.NumberU64())
	chain.saveUnitToDb(chain.ldbunitRep, chain.ldbUnitProduceRep, unit)
	//Set stable unit
	chain.stableUnitHash = hash
	chain.stableUnitHeight = unit.NumberU64()
}

func (chain *MemDag) AddUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenerate bool) (common2.IUnitRepository,
	common2.IUtxoRepository, common2.IStateRepository, common2.IPropRepository,
	common2.IUnitProduceRepository, error) {
	start := time.Now()
	if unit == nil {
		return nil, nil, nil, nil, nil, errors.ErrNullPoint
	}
	chain.lock.Lock()
	defer chain.lock.Unlock()
	if unit.NumberU64() <= chain.stableUnitHeight {
		log.Debugf("This unit is too old! Ignore it,stable unit height:%d, stable hash:%s",
			chain.stableUnitHeight, chain.stableUnitHash.String())
		go txpool.ResetPendingTxs(unit.Transactions())
		return nil, nil, nil, nil, nil, nil
	}
	chain_units := chain.getChainUnits()
	if _, has := chain_units[unit.Hash()]; has { // 不重复添加
		log.Debugf("MemDag[%s] received a repeated unit, hash[%s] ", chain.token.String(), unit.Hash().String())
		return nil, nil, nil, nil, nil, nil
	}
	a, b, c, d, e, err := chain.addUnit(unit, txpool, isGenerate)
	log.DebugDynamic(func() string {
		return fmt.Sprintf("MemDag[%s]: index: %d, hash: %s,AddUnit cost time: %v ,", chain.token.String(),
			unit.NumberU64(), unit.Hash().String(), time.Since(start))
	})

	if err == nil {
		// 进行下一个unit的群签名
		log.Debugf("send toGroupSign event")
		go chain.toGroupSignFeed.Send(modules.ToGroupSignEvent{})
	}

	return a, b, c, d, e, err
}

//添加单元到memdag
//1.判断该单元的父单元的位置。
//1.1.如果父单元是lastMainChainUnit,则保存该单元到主链。
//1.2.如果父单元在主链上，并且父单元不是lastMainChainUnit，此时要重新fork一个侧链，将该单元及所有祖先单元全部保存到该fork链。
//1.3.如果父单元不在主链上，并且该单元不是孤儿单元，则保存该单元到相应fork链。
//1.4.如果父单元不在memdag(孤儿单元),则将该单元保存orphan Memdag里。
//2.保存到主链上的单元，判断主链若有满足稳定条件的单元，则将该单元及祖先单元全部置为稳定单元。
//3.保存到侧链上的单元，若满足切换主链的条件，则要切换主链（switchMainChain）。
//4.添加完一个非孤儿单元后，判断是否有孤儿单元是该单元亲子单元，如有则将对应的孤儿单元连到链上。
func (chain *MemDag) addUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenerate bool) (common2.IUnitRepository,
	common2.IUtxoRepository, common2.IStateRepository, common2.IPropRepository,
	common2.IUnitProduceRepository, error) {
	parentHash := unit.ParentHash()[0]
	uHash := unit.Hash()
	height := unit.NumberU64()
	if inter, ok := chain.chainUnits.Load(parentHash); ok || parentHash == chain.stableUnitHash {
		//add unit to chain
		log.Debugf("chain[%s] Add unit[%s] to chainUnits", chain.token.String(), uHash.String())
		//add at the end of main chain unit
		if parentHash == chain.lastMainChainUnit.Hash() {
			//Add a new unit to main chain
			var temp_db *ChainTempDb
			inter_temp, has := chain.tempdb.Load(parentHash)
			if !has { // 分叉链
				p_temp := inter.(*ChainTempDb)
				temp_db, _ = NewChainTempDb(p_temp.Tempdb, chain.cache, chain.tokenEngine, chain.saveHeaderOnly)
			} else {
				temp_db = inter_temp.(*ChainTempDb)
			}
			if !isGenerate {
				var validateCode validator.ValidationCode
				if chain.saveHeaderOnly {
					validateCode = temp_db.Validator.ValidateHeader(unit.UnitHeader)
				} else {
					validateCode = temp_db.Validator.ValidateUnitExceptGroupSig(unit)
				}
				if validateCode != validator.TxValidationCode_VALID {
					vali_err := validator.NewValidateError(validateCode)
					log.Debugf("validate main chain unit error, %s, unit hash:%s",
						vali_err.Error(), uHash.String())
					// reset unit's txs
					go txpool.ResetPendingTxs(unit.Transactions())
					return nil, nil, nil, nil, nil, vali_err
				}
			}
			tempdb, _ := temp_db.AddUnit(unit, chain.saveHeaderOnly)
			// go tempdb.AddUnit(unit, chain.saveHeaderOnly)
			chain.tempdb.Store(uHash, tempdb)
			chain.chainUnits.Store(uHash, tempdb)
			if has {
				chain.tempdb.Delete(parentHash)
				chain.setLastMainchainUnit(unit)

				start := time.Now()
				if chain.checkStableCondition(unit, txpool) { //增加了单元后检查是否满足稳定单元的条件
					// 进行下一个unit的群签名
					log.Debugf("send toGroupSign event")
					go chain.toGroupSignFeed.Send(modules.ToGroupSignEvent{})
					log.Debugf("unit[%s] checkStableCondition =true", uHash.String())
					log.DebugDynamic(func() string {
						return fmt.Sprintf("check stable cost time: %s ,index: %d, hash: %s",
							time.Since(start), height, uHash.String())
					})
				}
			}
			//update txpool's tx status to pending
			// todo 如果该单元高度远低于全网的稳定单元的高度，忽略setPendingTxs
			if len(unit.Txs) > 0 {
				go txpool.SetPendingTxs(unit.Hash(), height, unit.Transactions())
			}

		} else { //Fork unit
			start1 := time.Now()
			var main_temp *ChainTempDb
			inter_main, has := chain.tempdb.Load(parentHash)
			if !has { // 分叉
				main_temp, _ = NewChainTempDb(chain.db, chain.cache, chain.tokenEngine, chain.saveHeaderOnly)
				forks := chain.getForkUnits(unit)
				for i := 0; i < len(forks)-1; i++ {
					main_temp, _ = main_temp.AddUnit(forks[i], chain.saveHeaderOnly)
				}
			} else {
				main_temp = inter_main.(*ChainTempDb)
			}
			if !isGenerate {
				var validateCode validator.ValidationCode
				if chain.saveHeaderOnly {
					validateCode = main_temp.Validator.ValidateHeader(unit.UnitHeader)
				} else {
					validateCode = main_temp.Validator.ValidateUnitExceptGroupSig(unit)
				}
				if validateCode != validator.TxValidationCode_VALID {
					vali_err := validator.NewValidateError(validateCode)
					log.Debugf("validate fork unit error, %s, unit hash:%s", vali_err.Error(), uHash.String())
					// reset unit's txs
					go txpool.ResetPendingTxs(unit.Transactions())
					return nil, nil, nil, nil, nil, vali_err
				}
			}
			temp, _ := main_temp.AddUnit(unit, chain.saveHeaderOnly)
			chain.tempdb.Delete(parentHash) // 删除parent的tempdb
			chain.tempdb.Store(uHash, temp)
			chain.chainUnits.Store(uHash, temp)

			log.DebugDynamic(func() string {
				return fmt.Sprintf("save fork unit cost time: %s ,index: %d, hash: %s",
					time.Since(start1), height, uHash.String())
			})
			// 满足切换主链条件， 则切换主链，更新主链单元。
			if height > chain.lastMainChainUnit.NumberU64() {
				log.Infof("switch main chain starting, fork index:%d, chain index:%d ,"+
					"fork hash:%s, main hash:%s", height, chain.lastMainChainUnit.NumberU64(),
					uHash.String(), chain.lastMainChainUnit.Hash().String())
				chain.switchMainChain(unit, txpool)
				log.DebugDynamic(func() string {
					main_chains := chain.getMainChainUnits()
					hashs := make([]common.Hash, 0)
					for _, u := range main_chains {
						hashs = append(hashs, u.UnitHash)
					}
					return fmt.Sprintf("switch chain end , main_chains:[%#x]", hashs)
				})
			}
		}
		//orphan unit can add below this unit?
		if inter, has := chain.orphanUnitsParants.Load(uHash); has {
			chain.orphanUnitsParants.Delete(uHash)
			next_hash := inter.(common.Hash)
			chain.processOrphan(next_hash, txpool, isGenerate)
		}
	} else {
		//add unit to orphan
		log.Debugf("This unit[%s] is an orphan unit", uHash.String())
		chain.orphanUnits.Store(uHash, unit)
		chain.orphanUnitsParants.Store(unit.ParentHash()[0], uHash)
	}
	chain.addUnitHeight(unit)
	inter_tmp, _ := chain.tempdb.Load(chain.lastMainChainUnit.Hash())
	tmp := inter_tmp.(*ChainTempDb)
	return tmp.UnitRep, tmp.UtxoRep, tmp.StateRep, tmp.PropRep, tmp.UnitProduceRep, nil
}

// 缓存该高度的所有单元hash
func (chain *MemDag) addUnitHeight(unit *modules.Unit) {
	height := unit.NumberU64()
	hs := make([]common.Hash, 0)
	inter, has := chain.height_hashs.Load(height)
	if has {
		all := inter.([]common.Hash)
		hs = append(hs, all...)
	}
	hs = append(hs, unit.Hash())
	chain.height_hashs.Store(height, hs)
}

// 单元稳定后，清空该高度的所有缓存
func (chain *MemDag) delHeightUnitsAndTemp(height uint64) {
	to_del_h := make([]uint64, 0)
	to_del_hash := make([]common.Hash, 0)
	chain.height_hashs.Range(func(k, v interface{}) bool {
		h := k.(uint64)
		if h <= height {
			to_del_h = append(to_del_h, h)
			hashs := v.([]common.Hash)
			to_del_hash = append(to_del_hash, hashs...)
		}
		return true
	})
	for _, h := range to_del_h {
		chain.height_hashs.Delete(h)
	}
	for _, hash := range to_del_hash {
		if hash != chain.stableUnitHash {
			chain.tempdb.Delete(hash)
		}
	}
}

//计算一个单元到稳定单元之间有多少个确认地址数
//func (chain *MemDag) getChainAddressCount(lastUnit *modules.Unit) int {
//	addrs := map[common.Address]bool{}
//	unitHash := lastUnit.Hash()
//	units := chain.getChainUnits()
//	for unitHash != chain.stableUnitHash {
//		unit := units[unitHash]
//		addrs[unit.Author()] = true
//		unitHash = unit.ParentHash()[0]
//	}
//	return len(addrs)
//}

//发现一条更长的确认数更多的链，则放弃原有主链，切换成新主链
//1.将旧主链上包含的交易在交易池中重置(resetPending)。
//2.更改新主链上的交易状态，(setPending)。
//3.设置新主链的最新单元（setLastMainCHainUnit）。
func (chain *MemDag) switchMainChain(newUnit *modules.Unit, txpool txspool.ITxPool) {
	chain_units := chain.getChainUnits()
	forks_units := chain.getForkUnits(newUnit)
	if forks_units == nil {
		return
	}

	for _, m_u := range forks_units {
		hash := m_u.Hash()
		delete(chain_units, hash)
	}
	// 不在主链上的区块，将已打包交易回滚。
	for _, un_unit := range chain_units {
		txs := un_unit.Transactions()
		if len(txs) > 1 {
			// 用协程，resettPending和txpool的读写锁，会导致这里会很耗时。
			go txpool.ResetPendingTxs(txs)
		}
	}
	//基于新主链，更新TxPool的状态
	for _, unit := range forks_units {
		if len(unit.Txs) > 1 {
			log.Debugf("Update tx[%#x] status to pending in txpool", unit.Txs.GetTxIds())
			go txpool.SetPendingTxs(unit.Hash(), unit.NumberU64(), unit.Transactions())
		}
	}
	//设置最新主链单元
	chain.setLastMainchainUnit(newUnit)
}

//将其从孤儿单元列表中删除，并添加到ChainUnits中。
func (chain *MemDag) processOrphan(unitHash common.Hash, txpool txspool.ITxPool, isProduce bool) {
	unit, has := chain.getOrphanUnits()[unitHash]
	if has {
		log.Debugf("Orphan unit[%s] can add to chain now.", unit.Hash().String())
		chain.orphanUnits.Delete(unitHash)
		chain.addUnit(unit, txpool, isProduce)
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
			log.Debugf("Orphan unit[%s] height[%d] is too low, remove it.",
				unit.Hash().String(), unit.NumberU64())
			if txs := unit.Transactions(); len(txs) > 1 {
				go txpool.ResetPendingTxs(txs)
			}
			chain.orphanUnits.Delete(hash)
		}
	}
}
func (chain *MemDag) getChainUnits() map[common.Hash]*modules.Unit {
	units := make(map[common.Hash]*modules.Unit)
	chain.chainUnits.Range(func(k, v interface{}) bool {
		hash := k.(common.Hash)
		ct := v.(*ChainTempDb)
		units[hash] = ct.Unit
		return true
	})
	return units
}
func (chain *MemDag) getChainUnit(hash common.Hash) (*ChainTempDb, error) {
	inter, ok := chain.chainUnits.Load(hash)
	if ok {
		return inter.(*ChainTempDb), nil
	}
	return nil, errors.ErrNotFound
}
func (chain *MemDag) GetLastStableUnitInfo() (common.Hash, uint64) {
	chain.lock.RLock()
	defer chain.lock.RUnlock()
	return chain.stableUnitHash, chain.stableUnitHeight
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
