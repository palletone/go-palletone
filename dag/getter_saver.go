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
	"strconv"
	"time"

	"github.com/dedis/kyber"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func (d *Dag) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := d.stablePropRep.RetrieveGlobalProp()
	return gp
}

func (d *Dag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp, _ := d.stablePropRep.RetrieveDynGlobalProp()
	return dgp
}

func (d *Dag) GetMediatorSchl() *modules.MediatorSchedule {
	ms, _ := d.stablePropRep.RetrieveMediatorSchl()
	return ms
}

func (d *Dag) SaveGlobalProp(gp *modules.GlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.stablePropRep.StoreGlobalProp(gp)
	return
}

func (d *Dag) SaveDynGlobalProp(dgp *modules.DynamicGlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.stablePropRep.StoreDynGlobalProp(dgp)
	return
}

func (d *Dag) SaveMediatorSchl(ms *modules.MediatorSchedule, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.stablePropRep.StoreMediatorSchl(ms)
	return
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorNodes() map[string]*discover.Node {
	nodes := make(map[string]*discover.Node)

	meds := d.GetActiveMediators()
	for _, add := range meds {
		med := d.GetActiveMediator(add)
		node := med.Node
		nodes[node.ID.TerminalString()] = node
	}

	return nodes
}

// author Albert·Gou
func (d *Dag) GetActiveMediatorInitPubs() []kyber.Point {
	aSize := d.ActiveMediatorsCount()
	pubs := make([]kyber.Point, aSize, aSize)

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
		log.Debug(fmt.Sprintf("%v is not active mediator!", add.Str()))
		return nil
	}

	return d.GetMediator(add)
}

func (d *Dag) GetMediator(add common.Address) *core.Mediator {
	med, err := d.unstableStateRep.RetrieveMediator(add)
	if err != nil {
		log.Error("dag", "GetMediator RetrieveMediator err:", err, "address:", add)
		return nil
	}
	return med
}

func (d *Dag) SaveMediator(med *core.Mediator, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.unstableStateRep.StoreMediator(med)
	return
}

func (dag *Dag) GetSlotAtTime(when time.Time) uint32 {
	return dag.stablePropRep.GetSlotAtTime(when)
}

func (dag *Dag) GetSlotTime(slotNum uint32) time.Time {
	return dag.stablePropRep.GetSlotTime(dag.GetGlobalProp(), dag.GetDynGlobalProp(), slotNum)
}

func (dag *Dag) GetScheduledMediator(slotNum uint32) common.Address {
	return dag.stablePropRep.GetScheduledMediator(slotNum)
}

func (dag *Dag) HeadUnitTime() int64 {
	gasToken := dagconfig.DagConfig.GetGasToken()
	t, _ := dag.stablePropRep.GetNewestUnitTimestamp(gasToken)
	return t
}

func (dag *Dag) HeadUnitNum() uint64 {
	gasToken := dagconfig.DagConfig.GetGasToken()
	_, idx, _ := dag.stablePropRep.GetNewestUnit(gasToken)
	return idx.Index
}

func (dag *Dag) LastMaintenanceTime() int64 {
	return int64(dag.GetDynGlobalProp().LastMaintenanceTime)
}

func (dag *Dag) HeadUnitHash() common.Hash {
	gasToken := dagconfig.DagConfig.GetGasToken()
	hash, _, _ := dag.stablePropRep.GetNewestUnit(gasToken)
	return hash
}

func (dag *Dag) GetMediators() map[common.Address]bool {
	return dag.unstableStateRep.GetMediators()
}

func (dag *Dag) GetApprovedMediatorList() ([]*modules.MediatorRegisterInfo, error) {
	return dag.unstableStateRep.GetApprovedMediatorList()
}

func (dag *Dag) IsApprovedMediator(address common.Address) bool {
	return dag.unstableStateRep.IsApprovedMediator(address)
}

func (dag *Dag) IsMediator(address common.Address) bool {
	return dag.unstableStateRep.IsMediator(address)
}

func (dag *Dag) ActiveMediators() map[common.Address]bool {
	return dag.GetGlobalProp().ActiveMediators
}

func (dag *Dag) CurrentFeeSchedule() core.FeeSchedule {
	return dag.GetGlobalProp().ChainParameters.CurrentFees
}

func (dag *Dag) GetChainParameters() core.ChainParameters {
	return dag.GetGlobalProp().ChainParameters
}

func (dag *Dag) GetImmutableChainParameters() core.ImmutableChainParameters {
	return dag.GetGlobalProp().ImmutableParameters
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
	for add, _ := range pmds {
		med := d.GetMediator(add)
		node := med.Node
		nodes[node.ID.TerminalString()] = node
	}
	return nodes
}

func (d *Dag) GetAccountVotedMediators(addr common.Address) []common.Address {
	data, err := d.unstableStateRep.GetAccountState(addr, constants.VOTED_MEDIATORS)
	if err != nil {
		log.Debugf(err.Error())
		return nil
	}

	votedMediators := make([]common.Address, 0)
	err = rlp.DecodeBytes(data.Value, &votedMediators)
	if err != nil {
		log.Debugf(err.Error())
		return nil
	}

	return votedMediators
}

func (d *Dag) LookupAccount() map[common.Address]*modules.AccountInfo {
	return d.unstableStateRep.LookupAccount()
}

func (d *Dag) GetPtnBalance(addr common.Address) uint64 {
	return d.unstableStateRep.GetAccountBalance(addr)
}

func (d *Dag) GetMediatorInfo(address common.Address) *modules.MediatorInfo {
	mi, _ := d.unstableStateRep.RetrieveMediatorInfo(address)
	return mi
}

func (d *Dag) JuryCount() int {
	return 100 //todo test

	juryList, err := d.unstableStateRep.GetJuryCandidateList()
	if err != nil {
		return len(juryList)
	}
	return 0
}

func (d *Dag) GetActiveJuries() []common.Address {
	return nil

	//return d.unstableStateRep.GetJuryCandidateList()
}

func (d *Dag) IsActiveJury(addr common.Address) bool {
	return true //todo for test

	return d.unstableStateRep.IsJury(addr)
}

func (d *Dag) getActiveMediatorCount() int {
	activeMediatorCountStr, _, _ := d.stableStateRep.GetConfig("ActiveMediatorCount")
	activeMediatorCount, _ := strconv.ParseUint(string(activeMediatorCountStr), 10, 16)

	return int(activeMediatorCount)
}
