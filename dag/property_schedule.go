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

package dag

import (
	"fmt"

	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

func (d *Dag) GetGlobalProp() *modules.GlobalProperty {
	return d.propdb.GetGlobalProp()
}

func (d *Dag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	return d.propdb.GetDynGlobalProp()
}

func (d *Dag) GetMediatorSchl() *modules.MediatorSchedule {
	return d.propdb.GetMediatorSchl()
}

// @author Albert·Gou
func (d *Dag) ValidateUnitExceptGroupSig(unit *modules.Unit, isGenesis bool) bool {
	unitState := d.validate.ValidateUnitExceptGroupSig(unit, isGenesis)
	if unitState != modules.UNIT_STATE_VALIDATED &&
		unitState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
		return false
	}
	return true
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorNodes() map[string]*discover.Node {
	return d.GetGlobalProp().GetActiveMediatorNodes()
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorInitPubs() []kyber.Point {
	return d.GetGlobalProp().GetActiveMediatorInitPubs()
}

// author Albert·Gou
func (d *Dag) GetCurThreshold() int {
	return d.GetGlobalProp().GetCurThreshold()
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorCount() int {
	return d.GetGlobalProp().GetActiveMediatorCount()
}

// author Albert·Gou
func (d *Dag) GetActiveMediators() []common.Address {
	return d.GetGlobalProp().GetActiveMediators()
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorAddr(index int) common.Address {
	return d.GetGlobalProp().GetActiveMediatorAddr(index)
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorNode(index int) *discover.Node {
	return d.GetGlobalProp().GetActiveMediatorNode(index)
}

// author Albert·Gou
func (d *Dag) GetActiveMediator(add common.Address) *core.Mediator {
	return d.GetGlobalProp().GetActiveMediator(add)
}

// author Albert·Gou
func (d *Dag) IsActiveMediator(add common.Address) bool {
	return d.GetGlobalProp().IsActiveMediator(add)
}
func (d *Dag) StoreGlobalProp(gp *modules.GlobalProperty) error {
	return d.propdb.StoreGlobalProp(gp)
}
func (d *Dag) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {
	return d.propdb.StoreDynGlobalProp(dgp)
}
func (d *Dag) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	return d.propdb.RetrieveGlobalProp()
}
func (d *Dag) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	return d.propdb.RetrieveDynGlobalProp()
}
func (d *Dag) StoreMediatorSchl(ms *modules.MediatorSchedule) error {
	return d.propdb.StoreMediatorSchl(ms)
}
func (d *Dag) RetrieveMediatorSchl() (*modules.MediatorSchedule, error) {
	return d.propdb.RetrieveMediatorSchl()
}

func (dag *Dag) InitPropertyDB(genesis *core.Genesis, genesisUnitHash common.Hash) error {
	//  全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	gp := modules.InitGlobalProp(genesis)
	if err := dag.StoreGlobalProp(gp); err != nil {
		log.Error(fmt.Sprintf("Failed to write global properties: %v", err))
		return err
	}

	//  动态全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	dgp := modules.InitDynGlobalProp(genesis, genesisUnitHash)
	if err := dag.StoreDynGlobalProp(dgp); err != nil {
		log.Error(fmt.Sprintf("Failed to write dynamic global properties: %v", err))
		return err
	}

	//  初始化mediator调度器，并存在数据库
	// @author Albert·Gou
	ms := modules.InitMediatorSchl(gp, dgp)
	if err := dag.StoreMediatorSchl(ms); err != nil {
		log.Error(fmt.Sprintf("Failed to write mediator schedule: %v", err))
		return err
	}
	return nil
}
