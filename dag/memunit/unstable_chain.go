package memunit

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	common2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

type UnstableChain struct {
	token             modules.IDType16
	stableUnitHash    common.Hash
	stableUnitHeight  uint64
	orphanUnits       map[common.Hash]*modules.Unit
	chainUnits        map[common.Hash]*modules.Unit
	lastMainchainUnit *modules.Unit
	tempdbunitRep     common2.IUnitRepository
	tempUtxoRep       common2.IUtxoRepository
	tempStateRep      common2.IStateRepository
	ldbunitRep        common2.IUnitRepository
	ldbPropRep        common2.IPropRepository
	tempdb            *Tempdb
}

func NewUnstableChain(token modules.IDType16, db ptndb.Database, stableUnitRep common2.IUnitRepository) *UnstableChain {
	propRep := common2.NewPropRepository4Db(db)

	tempdb, _ := NewTempdb(db)
	trep := common2.NewUnitRepository4Db(tempdb)
	tutxoRep := common2.NewUtxoRepository4Db(tempdb)
	tstateRep := common2.NewStateRepository4Db(tempdb)

	stablehash, stbIndex, err := propRep.GetLastStableUnit(token)
	if err != nil {
		log.Errorf("Cannot retrieve last stable unit from db for token:%s", token.String())
		return nil
	}
	stableUnit, _ := stableUnitRep.GetUnit(stablehash)
	return &UnstableChain{
		token:             token,
		ldbunitRep:        stableUnitRep,
		ldbPropRep:        propRep,
		tempdbunitRep:     trep,
		tempUtxoRep:       tutxoRep,
		tempStateRep:      tstateRep,
		tempdb:            tempdb,
		orphanUnits:       make(map[common.Hash]*modules.Unit),
		chainUnits:        make(map[common.Hash]*modules.Unit),
		stableUnitHash:    stablehash,
		stableUnitHeight:  stbIndex.Index,
		lastMainchainUnit: stableUnit,
	}
}
func (chain *UnstableChain) Init(stablehash common.Hash, stableHeight uint64) {
	chain.stableUnitHash = stablehash
	chain.stableUnitHeight = stableHeight
	chain.tempdb.Clear()
	chain.lastMainchainUnit, _ = chain.ldbunitRep.GetUnit(stablehash)

	for k := range chain.orphanUnits {
		delete(chain.orphanUnits, k)
	}
	for k := range chain.chainUnits {
		delete(chain.chainUnits, k)
	}
}
func (chain *UnstableChain) GetUnstableRepositories() (common2.IUnitRepository, common2.IUtxoRepository, common2.IStateRepository) {
	return chain.tempdbunitRep, chain.tempUtxoRep, chain.tempStateRep
}
func (chain *UnstableChain) SetUnitGroupSign(uHash common.Hash, groupPubKey []byte, groupSign []byte, txpool txspool.ITxPool) error {

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
func (chain *UnstableChain) SetStableUnit(hash common.Hash, height uint64, txpool txspool.ITxPool) {
	//oldStableHash := chain.stableUnitHash
	log.Debugf("Set stable unit to %s,height:%d", hash.String(), height)
	stableCount := int(height - chain.stableUnitHeight)
	newStableUnits := make([]*modules.Unit, stableCount)
	stbHash := hash
	for i := 0; i < stableCount; i++ {
		u := chain.chainUnits[stbHash]
		newStableUnits[stableCount-i-1] = u
		stbHash = u.ParentHash()[0]
	}
	//Save stable unit and it's parent
	for _, unit := range newStableUnits {
		chain.setNextStableUnit(unit, txpool)
	}

	chain.stableUnitHash = hash
	chain.stableUnitHeight = height

	// Rebuild temp db
	chain.rebuildTempdb()
}

//设置当前稳定单元的子单元为稳定单元
func (chain *UnstableChain) setNextStableUnit(unit *modules.Unit, txpool txspool.ITxPool) {
	hash := unit.Hash()
	height := unit.NumberU64()
	//remove fork units
	for _, funit := range chain.chainUnits {
		if funit.NumberU64() <= height && funit.Hash() != hash {
			chain.removeUnitAndChildren(funit.Hash())
		}
	}
	//Save stable unit to ldb
	chain.ldbunitRep.SaveUnit(unit, txpool, false, true)
	//remove new stable unit
	delete(chain.chainUnits, hash)
	//Set stable unit
	chain.stableUnitHash = hash
	chain.stableUnitHeight = height
}

func (chain *UnstableChain) checkStableCondition(needAddrCount int, txpool txspool.ITxPool) bool {
	unstableCount := int(chain.lastMainchainUnit.NumberU64() - chain.stableUnitHeight)
	//每个单元被多少个地址确认过(包括自己)
	unstableCofirmAddrs := make(map[common.Hash]map[common.Address]bool)
	childrenCofirmAddrs := make(map[common.Address]bool)
	ustbHash := chain.lastMainchainUnit.Hash()
	childrenCofirmAddrs[chain.lastMainchainUnit.Author()] = true
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
		if len(hs) >= needAddrCount {
			log.Debugf("Unit[%s] has enough confirm address, make it to stable.", ustbHash.String())
			chain.SetStableUnit(ustbHash, u.NumberU64(), txpool)

			return true
		}
		log.Debugf("Unstable unit[%s] has confirm address count:%d", ustbHash.String(), len(hs))

		ustbHash = u.ParentHash()[0]
	}
	return false
}
func (chain *UnstableChain) rebuildTempdb() {
	log.Debugf("Clear tempdb and reubild data")
	chain.tempdb.Clear()
	unstableCount := int(chain.lastMainchainUnit.NumberU64() - chain.stableUnitHeight)
	log.Debugf("Unstable unit count:%d", unstableCount)
	unstableUnits := make([]*modules.Unit, unstableCount)
	ustbHash := chain.lastMainchainUnit.Hash()
	for i := 0; i < unstableCount; i++ {
		u := chain.chainUnits[ustbHash]
		unstableUnits[unstableCount-i-1] = u
		//log.Debugf("Unstable unit:%#v, Hash:%s", u, ustbHash.String())
		ustbHash = u.ParentHash()[0]
	}
	for _, unit := range unstableUnits {
		chain.tempdbunitRep.SaveUnit(unit, nil, false, true)
	}
}

func (chain *UnstableChain) removeUnitAndChildren(hash common.Hash) {
	log.Debugf("Remove unit[%s] and it's children from chain unit", hash.String())
	for h, unit := range chain.chainUnits {
		if h == hash {
			delete(chain.chainUnits, h)
		} else {
			if unit.ParentHash()[0] == hash {
				chain.removeUnitAndChildren(h)
			}
		}
	}
}

func (chain *UnstableChain) AddUnit(unit *modules.Unit, txpool txspool.ITxPool) error {
	parentHash := unit.ParentHash()[0]
	uHash := unit.Hash()
	log.Debugf("Try to add unit[%s] to unstable chain", uHash.String())

	if _, ok := chain.chainUnits[parentHash]; ok || parentHash == chain.stableUnitHash {
		//add unit to chain
		chain.chainUnits[uHash] = unit
		//Switch main chain?
		if parentHash == chain.lastMainchainUnit.Hash() {
			log.Debug("This is a new main chain unit")
			//Add a new unit to main chain
			chain.setLastMainchainUnit(unit)
			if !chain.checkStableCondition(2, txpool) {
				chain.tempdbunitRep.SaveUnit(unit, nil, false, true)
			}
		} else { //Fork unit
			log.Debug("This is a fork unit")
			if unit.NumberU64() > chain.lastMainchainUnit.NumberU64() { //Need switch main chain
				//switch main chain, build db
				chain.setLastMainchainUnit(unit)
				if !chain.checkStableCondition(2, txpool) {
					chain.rebuildTempdb()
				}
			} else {
				log.Infof("This unit is too old! Ignore it")
			}
		}
		//orphan unit can add below this unit?
		chain.processOrphan(uHash, txpool)
	} else {
		//add unit to orphan
		log.Infof("This unit[%s] is a orphan unit", uHash.String())
		chain.orphanUnits[uHash] = unit
	}
	return nil
}
func (chain *UnstableChain) processOrphan(unitHash common.Hash, txpool txspool.ITxPool) {
	for hash, unit := range chain.orphanUnits {
		if unit.ParentHash()[0] == unitHash {
			log.Debugf("Orphan unit[%s] can add to chain now.", unit.Hash().String())
			delete(chain.orphanUnits, hash)
			chain.AddUnit(unit, txpool)
			break
		}
	}
}
func (chain *UnstableChain) getChainUnit(hash common.Hash) (*modules.Unit, error) {
	if unit, ok := chain.chainUnits[hash]; ok {
		return unit, nil
	}
	return nil, errors.ErrNotFound
}

//func (chain *UnstableChain) Exists(uHash common.Hash) bool {
//	_, ok := chain.chainUnits[uHash]
//	return ok
//}
func (chain *UnstableChain) GetLastMainchainUnit() *modules.Unit {
	return chain.lastMainchainUnit
}
func (chain *UnstableChain) setLastMainchainUnit(unit *modules.Unit) {
	chain.lastMainchainUnit = unit
	chain.ldbPropRep.SetNewestUnit(unit.Header())
}
func (chain *UnstableChain) GetChainUnits() map[common.Hash]*modules.Unit {
	return chain.chainUnits
}
