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
	"github.com/palletone/go-palletone/common/ptndb"
	comm2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/tokenengine"
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

func NewChainTempDb(db ptndb.Database,
	cache palletcache.ICache, tokenEngine tokenengine.ITokenEngine, saveHeaderOnly bool) (*ChainTempDb, error) {
	tempdb, _ := NewTempdb(db)
	trep := comm2.NewUnitRepository4Db(tempdb, tokenEngine)
	tutxoRep := comm2.NewUtxoRepository4Db(tempdb, tokenEngine)
	tstateRep := comm2.NewStateRepository4Db(tempdb)
	tpropRep := comm2.NewPropRepository4Db(tempdb)
	tunitProduceRep := comm2.NewUnitProduceRepository(trep, tpropRep, tstateRep)
	val := validator.NewValidate(trep, tutxoRep, tstateRep, tpropRep, cache)
	if saveHeaderOnly { //轻节点，只有Header数据，无法做高级验证
		val = validator.NewValidate(trep, nil, nil, nil, cache)
	}
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
