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
)

func (d *Dag) GetGlobalProp() *modules.GlobalProperty {
	gp, _ := d.propRep.RetrieveGlobalProp()
	return gp
}

func (d *Dag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	dgp, _ := d.propRep.RetrieveDynGlobalProp()
	return dgp
}

func (d *Dag) GetMediatorSchl() *modules.MediatorSchedule {
	ms, _ := d.propRep.RetrieveMediatorSchl()
	return ms
}

func (d *Dag) SaveGlobalProp(gp *modules.GlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propRep.StoreGlobalProp(gp)
	return
}

func (d *Dag) SaveDynGlobalProp(dgp *modules.DynamicGlobalProperty, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propRep.StoreDynGlobalProp(dgp)
	return
}

func (d *Dag) SaveMediatorSchl(ms *modules.MediatorSchedule, onlyStore bool) {
	if !onlyStore {
		// todo 更新缓存
	}

	d.propRep.StoreMediatorSchl(ms)
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
	if !d.GetGlobalProp().IsActiveMediator(add) {
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
	return dag.propRep.GetSlotAtTime(dag.GetGlobalProp(), dag.GetDynGlobalProp(), when)
}

func (dag *Dag) GetSlotTime(slotNum uint32) time.Time {
	return dag.propRep.GetSlotTime(dag.GetGlobalProp(), dag.GetDynGlobalProp(), slotNum)
}

func (dag *Dag) GetScheduledMediator(slotNum uint32) common.Address {
	return dag.propRep.GetScheduledMediator(slotNum)
}

func (dag *Dag) HeadUnitTime() int64 {
	t, _ := dag.propRep.GetNewestUnitTimestamp(modules.PTNCOIN)
	return t
}

func (dag *Dag) HeadUnitNum() uint64 {
	_, idx, _ := dag.propRep.GetNewestUnit(modules.PTNCOIN)
	return idx.Index
}

func (dag *Dag) LastMaintenanceTime() int64 {
	return int64(dag.GetDynGlobalProp().LastMaintenanceTime)
}

func (dag *Dag) HeadUnitHash() common.Hash {
	hash, _, _ := dag.propRep.GetNewestUnit(modules.PTNCOIN)
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

func (d *Dag) GetVotedMediator(addr common.Address) map[common.Address]bool {
	accountInfo, err := d.unstableStateRep.RetrieveAccountInfo(addr)
	if err != nil {
		accountInfo = modules.NewAccountInfo()
	}

	return accountInfo.VotedMediators
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

func (d *Dag) IsActiveJury(add common.Address) bool {
	return d.GetGlobalProp().IsActiveJury(add)
}

func (d *Dag) GetActiveJuries() []common.Address {
	return d.GetGlobalProp().GetActiveJuries()
}
