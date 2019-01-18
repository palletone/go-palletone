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
	stableUnitHash    common.Hash
	stableUnitHeight  uint64
	orphanUnits       map[common.Hash]*modules.Unit
	chainUnits        map[common.Hash]*modules.Unit
	lastMainchainUnit *modules.Unit
	tempdbunitRep     common2.IUnitRepository
	ldbunitRep        common2.IUnitRepository
	tempdb            *Tempdb
}

func NewUnstableChain(db ptndb.Database, tempdb *Tempdb, stablehash common.Hash, stableHeight uint64) *UnstableChain {
	ldbRep := common2.NewUnitRepository4Db(db)
	stableUnit, _ := ldbRep.GetUnit(stablehash)
	trep := common2.NewUnitRepository4Db(tempdb)
	return &UnstableChain{
		ldbunitRep:        ldbRep,
		tempdbunitRep:     trep,
		tempdb:            tempdb,
		orphanUnits:       make(map[common.Hash]*modules.Unit),
		chainUnits:        make(map[common.Hash]*modules.Unit),
		stableUnitHash:    stablehash,
		stableUnitHeight:  stableHeight,
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
	for _, unit := range newStableUnits {
		chain.ldbunitRep.SaveUnit(unit, txpool, false, true)
	}

	chain.stableUnitHash = hash
	chain.stableUnitHeight = height
	//remove fork units
	for _, unit := range chain.chainUnits {
		if unit.NumberU64() <= height {
			chain.removeUnitAndChildren(unit.Hash())
		}
	}
	// Rebuild temp db
	chain.rebuildTempdb()
}
func (chain *UnstableChain) checkStableCondition(needAddrCount int, txpool txspool.ITxPool) bool {
	unstableCount := int(chain.lastMainchainUnit.NumberU64() - chain.stableUnitHeight)
	unstableCofirmAddrs := make(map[common.Hash]map[common.Address]bool)
	ustbHash := chain.lastMainchainUnit.Hash()
	for i := 0; i < unstableCount; i++ {
		u := chain.chainUnits[ustbHash]
		hs := unstableCofirmAddrs[ustbHash]
		hs[u.Author()] = true
		if len(hs) >= needAddrCount {
			log.Debugf("Unit[%s] has enough confirm address, make it to stable.", ustbHash.String())
			chain.SetStableUnit(ustbHash, u.NumberU64(), txpool)

			return true
		}
		ustbHash = u.ParentHash()[0]
	}
	return false
}
func (chain *UnstableChain) rebuildTempdb() {
	log.Debugf("Clear tempdb and reubild data")
	chain.tempdb.Clear()
	unstableCount := int(chain.lastMainchainUnit.NumberU64() - chain.stableUnitHeight)
	unstableUnits := make([]*modules.Unit, unstableCount)
	ustbHash := chain.lastMainchainUnit.Hash()
	for i := 0; i < unstableCount; i++ {
		u := chain.chainUnits[ustbHash]
		unstableUnits[unstableCount-i-1] = u
		ustbHash = u.ParentHash()[0]
	}
	for _, unit := range unstableUnits {
		chain.tempdbunitRep.SaveUnit(unit, nil, false, true)
	}
}

func (chain *UnstableChain) removeUnitAndChildren(hash common.Hash) {
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

	if _, ok := chain.chainUnits[parentHash]; ok {
		//add unit to chain
		chain.chainUnits[uHash] = unit
		//Switch main chain?
		if parentHash == chain.lastMainchainUnit.Hash() {
			//Add a new unit to main chain
			chain.lastMainchainUnit = unit
			if !chain.checkStableCondition(2, txpool) {
				chain.tempdbunitRep.SaveUnit(unit, nil, false, true)
			}
		} else { //Fork unit
			if unit.NumberU64() > chain.lastMainchainUnit.NumberU64() { //Need switch main chain
				//switch main chain, build db
				chain.lastMainchainUnit = unit
				if !chain.checkStableCondition(2, txpool) {
					chain.rebuildTempdb()
				}
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
func (chain *UnstableChain) GetUnit(hash common.Hash) (*modules.Unit, error) {
	if unit, ok := chain.chainUnits[hash]; ok {
		return unit, nil
	}
	return nil, errors.ErrNotFound
}

func (chain *UnstableChain) Exists(uHash common.Hash) bool {
	_, ok := chain.chainUnits[uHash]
	return ok
}
func (chain *UnstableChain) GetLastMainchainUnit() *modules.Unit {
	return chain.lastMainchainUnit
}
func (chain *UnstableChain) GetChainUnits() map[common.Hash]*modules.Unit {
	return chain.chainUnits
}
