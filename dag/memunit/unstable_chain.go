package memunit

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type UnstableChain struct{
	stableUnitHash common.Hash
	stableUnitHeight uint64
	orphanUnits []*modules.Unit
	chainUnits []*modules.Unit
	lastMainchainUnit *modules.Unit
	tempdb Tempdb
}
func(chain *UnstableChain) SetStableUnit(hash common.Hash,height uint64){
	chain.stableUnitHash=hash
	chain.stableUnitHeight=height
	//Todo remove fork units
	//Todo Rebuild db
}
func(chain *UnstableChain) AddUnit(unit *modules.Unit) error{
	return nil
}
