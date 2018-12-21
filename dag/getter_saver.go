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
	"time"

	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

func (d *Dag) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := d.propdb.RetrieveGlobalProp()
	return gp
}

func (d *Dag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp, _ := d.propdb.RetrieveDynGlobalProp()
	return dgp
}

func (d *Dag) GetMediatorSchl() *modules.MediatorSchedule {
	ms, _ := d.propdb.RetrieveMediatorSchl()
	return ms
}

func (d *Dag) SaveGlobalProp(gp *modules.GlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propdb.StoreGlobalProp(gp)
	return
}

func (d *Dag) SaveDynGlobalProp(dgp *modules.DynamicGlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propdb.StoreDynGlobalProp(dgp)
	return
}

func (d *Dag) SaveMediatorSchl(ms *modules.MediatorSchedule, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propdb.StoreMediatorSchl(ms)
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
	aSize := d.GetActiveMediatorCount()
	pubs := make([]kyber.Point, aSize, aSize)

	meds := d.GetActiveMediators()
	for i, add := range meds {
		med := d.GetActiveMediator(add)

		pubs[i] = med.InitPartPub
	}

	return pubs
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
	ma := d.GetActiveMediatorAddr(index)
	med := d.GetActiveMediator(ma)

	return med.Node
}

// author Albert·Gou
func (d *Dag) GetActiveMediator(add common.Address) *core.Mediator {
	if !d.GetGlobalProp().IsActiveMediator(add) {
		log.Debug(fmt.Sprintf("%v is not active mediator!", add.Str()))
		return nil
	}

	return d.GetMediator(add)
}

func (d *Dag) GetMediator(add common.Address) *core.Mediator {
	med, err := d.statedb.RetrieveMediator(add)
	if err != nil {
		log.Debug("dag", "GetMediator RetrieveMediator err:", err, "address:", add)
		return nil
	}
	return med
}

func (d *Dag) SaveMediator(med *core.Mediator, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.statedb.StoreMediator(med)
	return
}

func (dag *Dag) GetSlotAtTime(when time.Time) uint32 {
	return modules.GetSlotAtTime(dag.GetGlobalProp(), dag.GetDynGlobalProp(), when)
}

func (dag *Dag) GetSlotTime(slotNum uint32) time.Time {
	return modules.GetSlotTime(dag.GetGlobalProp(), dag.GetDynGlobalProp(), slotNum)
}

func (dag *Dag) GetScheduledMediator(slotNum uint32) common.Address {
	return dag.GetMediatorSchl().GetScheduledMediator(dag.GetDynGlobalProp(), slotNum)
}

func (dag *Dag) HeadUnitTime() int64 {
	return dag.GetDynGlobalProp().HeadUnitTime
}

func (dag *Dag) HeadUnitNum() uint64 {
	return dag.GetDynGlobalProp().HeadUnitNum
}

func (dag *Dag) HeadUnitHash() common.Hash {
	return dag.GetDynGlobalProp().HeadUnitHash
}

func (dag *Dag) GetMediators() map[common.Address]bool {
	return dag.statedb.GetMediators()
}

func (dag *Dag) GetAllMediatorInCandidateList() ([]*modules.MediatorInfo, error) {
	return dag.statedb.GetMediatorCandidateList()
}

func (dag *Dag) IsInMediatorCandidateList(address common.Address) bool {
	return dag.statedb.IsInMediatorCandidateList(address)
}

func (dag *Dag) IsMediator(address common.Address) bool {
	return dag.statedb.IsMediator(address)
}

func (dag *Dag) ActiveMediators() map[common.Address]bool {
	return dag.GetGlobalProp().ActiveMediators
}

func (dag *Dag) CurrentFeeSchedule() core.FeeSchedule {
	return dag.GetGlobalProp().ChainParameters.CurrentFees
}

func (dag *Dag) GetUnit(hash common.Hash) (*modules.Unit, error) {
	unit, err := dag.Memdag.GetUnit(hash)

	if unit == nil || err != nil {
		unit, err = dag.dagdb.GetUnit(hash)
	}

	if unit == nil || err != nil {
		log.Debug("get unit by hash is failed.", "hash", hash)
	}

	return unit, err
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

func (d *Dag) GetVotedMediator(addr common.Address) map[common.Address]bool {
	accountInfo, err := d.statedb.RetrieveAccountInfo(addr)
	if err != nil {
		accountInfo = modules.NewAccountInfo()
	}

	return accountInfo.VotedMediators
}

func (d *Dag) LookupAccount() map[common.Address]*modules.AccountInfo {
	return d.statedb.LookupAccount()
}

func (d *Dag) GetMediatorInfo(address common.Address) *storage.MediatorInfo {
	mi, _ := d.statedb.RetrieveMediatorInfo(address)
	return mi
}

func (d *Dag) IsActiveJury(add common.Address) bool {
	return d.GetGlobalProp().IsActiveJury(add)
}

func (d *Dag) GetActiveJuries() []common.Address {
	return d.GetGlobalProp().GetActiveJuries()
}
