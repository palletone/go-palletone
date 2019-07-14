package memunit

import (
	"github.com/palletone/go-palletone/common/ptndb"
	comm2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/validator"
)

type ChainTempDb struct {
	Tempdb         *Tempdb
	UnitRep        comm2.IUnitRepository
	UtxoRep        comm2.IUtxoRepository
	StateRep       comm2.IStateRepository
	PropRep        comm2.IPropRepository
	UnitProduceRep comm2.IUnitProduceRepository
	Validator      validator.Validator
	Unit           *modules.Unit
}

func NewChainTempDb(db ptndb.Database, cache palletcache.ICache) (*ChainTempDb, error) {
	tempdb, _ := NewTempdb(db)
	trep := comm2.NewUnitRepository4Db(tempdb)
	tutxoRep := comm2.NewUtxoRepository4Db(tempdb)
	tstateRep := comm2.NewStateRepository4Db(tempdb)
	tpropRep := comm2.NewPropRepository4Db(tempdb)
	tunitProduceRep := comm2.NewUnitProduceRepository(trep, tpropRep, tstateRep)
	val := validator.NewValidate(trep, tutxoRep, tstateRep, tpropRep, cache)

	return &ChainTempDb{
		Tempdb:         tempdb,
		UnitRep:        trep,
		UtxoRep:        tutxoRep,
		StateRep:       tstateRep,
		PropRep:        tpropRep,
		UnitProduceRep: tunitProduceRep,
		Validator:      val,
	}, nil
}

func (chain_temp *ChainTempDb) AddUnit(unit *modules.Unit, saveHeaderOnly bool) (*ChainTempDb, error) {
	if saveHeaderOnly {
		err := chain_temp.UnitRep.SaveNewestHeader(unit.Header())
		if err != nil {
			return chain_temp, err
		}
	} else {
		err := chain_temp.UnitProduceRep.PushUnit(unit)
		if err != nil {
			return chain_temp, err
		}
	}
	return &ChainTempDb{
		Tempdb:         chain_temp.Tempdb,
		UnitRep:        chain_temp.UnitRep,
		UtxoRep:        chain_temp.UtxoRep,
		StateRep:       chain_temp.StateRep,
		PropRep:        chain_temp.PropRep,
		UnitProduceRep: chain_temp.UnitProduceRep,
		Validator:      chain_temp.Validator,
		Unit:           unit,
	}, nil
}
