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
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"go.dedis.ch/kyber/v3"
)

func (d *Dag) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := d.unstablePropRep.RetrieveGlobalProp()
	return gp
}

func (d *Dag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp, _ := d.unstablePropRep.RetrieveDynGlobalProp()
	return dgp
}

func (d *Dag) GetMediatorSchl() *modules.MediatorSchedule {
	ms, _ := d.unstablePropRep.RetrieveMediatorSchl()
	return ms
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorNodes() map[string]*discover.Node {
	nodes := make(map[string]*discover.Node)
	meds := d.GetActiveMediators()
	for _, add := range meds {
		med := d.GetMediator(add)
		if med == nil {
			continue
		}
		node := med.Node
		nodes[node.ID.TerminalString()] = node
	}
	return nodes
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorInitPubs() []kyber.Point {
	aSize := d.ActiveMediatorsCount()
	pubs := make([]kyber.Point, aSize)

	meds := d.GetActiveMediators()
	for i, add := range meds {
		med := d.GetActiveMediator(add)
		pubs[i] = med.InitPubKey
	}
	return pubs
}

// author Albert·Gou
func (d *Dag) ActiveMediatorsCount() int {
	return d.GetGlobalProp().ActiveMediatorsCount()
}

func (d *Dag) PrecedingMediatorsCount() int {
	return d.GetGlobalProp().PrecedingMediatorsCount()
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
func (d *Dag) GetActiveMediator(add common.Address) *core.Mediator {
	if !d.IsActiveMediator(add) {
		log.Debugf("%v is not active mediator!", add.Str())
		return nil
	}
	return d.GetMediator(add)
}

func (d *Dag) GetMediator(add common.Address) *core.Mediator {
	med, err := d.unstableStateRep.RetrieveMediator(add)
	if err != nil {
		log.Debugf("Retrieve Mediator error: %v", err.Error())
		return nil
	}
	return med
}

func (dag *Dag) GetSlotAtTime(when time.Time) uint32 {
	return dag.unstablePropRep.GetSlotAtTime(when)
}

func (dag *Dag) GetNewestUnitTimestamp(token modules.AssetId) (int64, error) {
	return dag.unstablePropRep.GetNewestUnitTimestamp(token)
}

func (dag *Dag) GetSlotTime(slotNum uint32) time.Time {
	return dag.unstablePropRep.GetSlotTime(slotNum)
}

func (dag *Dag) GetScheduledMediator(slotNum uint32) common.Address {
	return dag.unstablePropRep.GetScheduledMediator(slotNum)
}

func (dag *Dag) HeadUnitTime() int64 {
	gasToken := dagconfig.DagConfig.GetGasToken()
	t, _ := dag.unstablePropRep.GetNewestUnitTimestamp(gasToken)
	return t
}

func (dag *Dag) HeadUnitNum() uint64 {
	gasToken := dagconfig.DagConfig.GetGasToken()
	_, idx, _ := dag.unstablePropRep.GetNewestUnit(gasToken)
	return idx.Index
}

func (dag *Dag) LastMaintenanceTime() int64 {
	return int64(dag.GetDynGlobalProp().LastMaintenanceTime)
}

func (dag *Dag) HeadUnitHash() common.Hash {
	gasToken := dagconfig.DagConfig.GetGasToken()
	hash, _, _ := dag.unstablePropRep.GetNewestUnit(gasToken)
	return hash
}

func (dag *Dag) GetMediators() map[common.Address]bool {
	return dag.unstableStateRep.GetMediators()
}

func (dag *Dag) GetMediatorCount() int {
	return len(dag.unstableStateRep.GetMediators())
}

func (dag *Dag) LookupMediatorInfo() []*modules.MediatorInfo {
	return dag.unstableStateRep.LookupMediatorInfo()
}

func (dag *Dag) IsMediator(address common.Address) bool {
	return dag.unstableStateRep.IsMediator(address)
}

func (dag *Dag) GetChainParameters() *core.ChainParameters {
	return dag.unstablePropRep.GetChainParameters()
}

func (dag *Dag) GetImmutableChainParameters() *core.ImmutableChainParameters {
	return &dag.GetGlobalProp().ImmutableParameters
}

func (dag *Dag) GetUnitByHash(hash common.Hash) (*modules.Unit, error) {
	unit, err := dag.unstableUnitRep.GetUnit(hash)
	if err != nil {
		log.Debug("get unit by hash is failed.", "hash", hash)
		return nil, err
	}
	return unit, nil
}

func (d *Dag) GetPrecedingMediatorNodes() map[string]*discover.Node {
	nodes := make(map[string]*discover.Node)
	pmds := d.GetGlobalProp().PrecedingMediators
	for add := range pmds {
		med := d.GetMediator(add)
		node := med.Node
		nodes[node.ID.TerminalString()] = node
	}
	return nodes
}

func (d *Dag) GetAccountVotedMediators(addr common.Address) map[string]bool {
	return d.unstableStateRep.GetAccountVotedMediators(addr)
}

func (d *Dag) GetPtnBalance(addr common.Address) uint64 {
	return d.unstableStateRep.GetAccountBalance(addr)
}

func (d *Dag) GetMediatorInfo(address common.Address) *modules.MediatorInfo {
	mi, _ := d.unstableStateRep.RetrieveMediatorInfo(address)
	return mi
}

func (d *Dag) JuryCount() uint {
	juryList, err := d.unstableStateRep.GetJuryCandidateList()
	if err != nil {
		return 0
	}
	return uint(len(juryList))
}

func (d *Dag) GetActiveJuries() []common.Address {
	return nil //todo

	//return d.unstableStateRep.GetJuryCandidateList()
}

func (d *Dag) IsActiveJury(addr common.Address) bool {
	return d.unstableStateRep.IsJury(addr)
}

func (d *Dag) GetContractDevelopers() ([]common.Address, error) {
	return d.unstableStateRep.GetContractDeveloperList()
}

func (d *Dag) IsContractDeveloper(addr common.Address) bool {
	return d.unstableStateRep.IsContractDeveloper(addr)
}

func (d *Dag) GetUnitHash(number *modules.ChainIndex) (common.Hash, error) {
	return d.unstableUnitRep.GetHashByNumber(number)
}

// return all mediators voted results
func (d *Dag) MediatorVotedResults() (map[string]uint64, error) {
	return d.unstableStateRep.GetMediatorVotedResults()
}

func (d *Dag) GetVotingForMediator(addStr string) (map[string]uint64, error) {
	return d.unstableStateRep.GetVotingForMediator(addStr)
}

func (d *Dag) GetJurorReward(jurorAdd common.Address) common.Address {
	return d.unstableStateRep.GetJurorReward(jurorAdd)
}

func (d *Dag) UnstableHeadUnitProperty(asset modules.AssetId) (*modules.UnitProperty, error) {
	return d.unstablePropRep.GetHeadUnitProperty(asset)
}

func (d *Dag) StableHeadUnitProperty(asset modules.AssetId) (*modules.UnitProperty, error) {
	return d.stablePropRep.GetHeadUnitProperty(asset)
}
