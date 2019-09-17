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
	_, _, _, rep, _ := d.Memdag.GetUnstableRepositories()
	//gp, _ := d.unstablePropRep.RetrieveGlobalProp()
	gp, _ := rep.RetrieveGlobalProp()
	return gp
}

func (d *Dag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	_, _, _, rep, _ := d.Memdag.GetUnstableRepositories()
	//dgp, _ := d.unstablePropRep.RetrieveDynGlobalProp()
	dgp, _ := rep.RetrieveDynGlobalProp()
	return dgp
}

func (d *Dag) GetMediatorSchl() *modules.MediatorSchedule {
	_, _, _, rep, _ := d.Memdag.GetUnstableRepositories()
	//ms, _ := d.unstablePropRep.RetrieveMediatorSchl()
	ms, _ := rep.RetrieveMediatorSchl()
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
func (d *Dag) GetActiveMediatorNode(index int) *discover.Node {
	ma := d.GetActiveMediatorAddr(index)
	med := d.GetActiveMediator(ma)

	return med.Node
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
	_, _, state, _, _ := d.Memdag.GetUnstableRepositories()
	//med, err := d.unstableStateRep.RetrieveMediator(add)

	med, err := state.RetrieveMediator(add)
	if err != nil {
		log.Debugf("Retrieve Mediator error: %v", err.Error())
		return nil
	}

	return med
}

func (dag *Dag) GetSlotAtTime(when time.Time) uint32 {
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	return rep.GetSlotAtTime(when)
	//return dag.unstablePropRep.GetSlotAtTime(when)
}

func (dag *Dag) GetNewestUnitTimestamp(token modules.AssetId) (int64, error) {
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	return rep.GetNewestUnitTimestamp(token)
}

func (dag *Dag) GetSlotTime(slotNum uint32) time.Time {
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	return rep.GetSlotTime(slotNum)
	//return dag.unstablePropRep.GetSlotTime(slotNum)
}

func (dag *Dag) GetScheduledMediator(slotNum uint32) common.Address {
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	return rep.GetScheduledMediator(slotNum)
	//return dag.unstablePropRep.GetScheduledMediator(slotNum)
}

func (dag *Dag) HeadUnitTime() int64 {
	gasToken := dagconfig.DagConfig.GetGasToken()
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	t, _ := rep.GetNewestUnitTimestamp(gasToken)
	return t
}

func (dag *Dag) HeadUnitNum() uint64 {
	gasToken := dagconfig.DagConfig.GetGasToken()
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	_, idx, _ := rep.GetNewestUnit(gasToken)
	return idx.Index
}

func (dag *Dag) LastMaintenanceTime() int64 {
	return int64(dag.GetDynGlobalProp().LastMaintenanceTime)
}

func (dag *Dag) HeadUnitHash() common.Hash {
	gasToken := dagconfig.DagConfig.GetGasToken()
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	hash, _, _ := rep.GetNewestUnit(gasToken)
	return hash
}

func (dag *Dag) GetMediators() map[common.Address]bool {
	_, _, state, _, _ := dag.Memdag.GetUnstableRepositories()

	return state.GetMediators()
	//return dag.unstableStateRep.GetMediators()
}

func (dag *Dag) GetMediatorCount() int {
	_, _, state, _, _ := dag.Memdag.GetUnstableRepositories()

	return len(state.GetMediators())
	//return len(dag.unstableStateRep.GetMediators())
}

func (dag *Dag) LookupMediatorInfo() []*modules.MediatorInfo {
	_, _, state, _, _ := dag.Memdag.GetUnstableRepositories()

	return state.LookupMediatorInfo()
	//return dag.unstableStateRep.LookupMediatorInfo()
}

func (dag *Dag) IsMediator(address common.Address) bool {
	_, _, state, _, _ := dag.Memdag.GetUnstableRepositories()

	return state.IsMediator(address)
	//return dag.unstableStateRep.IsMediator(address)
}

func (dag *Dag) GetChainParameters() *core.ChainParameters {
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	return rep.GetChainParameters()
	//return dag.unstablePropRep.GetChainParameters()
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
	_, _, state, _, _ := d.Memdag.GetUnstableRepositories()
	return state.GetAccountVotedMediators(addr)
	//return d.unstableStateRep.GetAccountVotedMediators(addr)
}

func (d *Dag) GetPtnBalance(addr common.Address) uint64 {
	_, _, state, _, _ := d.Memdag.GetUnstableRepositories()

	return state.GetAccountBalance(addr)
	//return d.unstableStateRep.GetAccountBalance(addr)
}

func (d *Dag) GetMediatorInfo(address common.Address) *modules.MediatorInfo {
	_, _, state, _, _ := d.Memdag.GetUnstableRepositories()

	mi, _ := state.RetrieveMediatorInfo(address)
	//mi, _ := d.unstableStateRep.RetrieveMediatorInfo(address)
	return mi
}

func (d *Dag) JuryCount() uint {
	//todo test
	//return 20

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
	//return true //todo

	return d.unstableStateRep.IsJury(addr)
}

func (d *Dag) GetContractDevelopers() ([]common.Address, error) {
	return d.unstableStateRep.GetContractDeveloperList()
}

func (d *Dag) IsContractDeveloper(addr common.Address) bool {
	//return true //todo
	return d.unstableStateRep.IsContractDeveloper(addr)
}

func (d *Dag) GetUnitHash(number *modules.ChainIndex) (common.Hash, error) {
	return d.unstableUnitRep.GetHashByNumber(number)
}

// return all mediators voted results
func (d *Dag) MediatorVotedResults() (map[string]uint64, error) {
	_, _, state, _, _ := d.Memdag.GetUnstableRepositories()
	return state.GetMediatorVotedResults()
	//return d.unstableStateRep.GetMediatorVotedResults()
}

func (d *Dag) GetVotingForMediator(addStr string) (map[string]uint64, error) {
	_, _, state, _, _ := d.Memdag.GetUnstableRepositories()
	return state.GetVotingForMediator(addStr)
}
