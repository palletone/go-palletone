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
	"encoding/json"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"go.dedis.ch/kyber/v3/sign/bls"
)

func (d *Dag) SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription {
	return d.Memdag.SubscribeToGroupSignEvent(ch)
}

func (d *Dag) IsActiveMediator(add common.Address) bool {
	return d.GetGlobalProp().IsActiveMediator(add)
}

func (d *Dag) IsPrecedingMediator(add common.Address) bool {
	return d.GetGlobalProp().IsPrecedingMediator(add)
}

func (dag *Dag) InitPropertyDB(genesis *core.Genesis, unit *modules.Unit) error {
	//  全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	gp := modules.InitGlobalProp(genesis)
	if err := dag.stablePropRep.StoreGlobalProp(gp); err != nil {
		return err
	}

	//  动态全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	dgp := modules.InitDynGlobalProp(unit)
	if err := dag.stablePropRep.StoreDynGlobalProp(dgp); err != nil {
		return err
	}
	//dag.stablePropRep.SetNewestUnit(unit.Header())

	//  初始化mediator调度器，并存在数据库
	// @author Albert·Gou
	ms := modules.InitMediatorSchl(gp, dgp)
	if err := dag.stablePropRep.StoreMediatorSchl(ms); err != nil {
		return err
	}
	dag.stablePropRep.UpdateMediatorSchedule()

	return nil
}

func (dag *Dag) InitStateDB(genesis *core.Genesis, unit *modules.Unit) error {
	version := &modules.StateVersion{
		Height:  unit.Number(),
		TxIndex: ^uint32(0),
	}
	ws := &modules.ContractWriteSet{
		IsDelete: false,
		//Key:      modules.MediatorList,
		//Value: imcB,
	}

	// Create initial mediators
	list := make(map[string]bool, len(genesis.InitialMediatorCandidates))

	for _, imc := range genesis.InitialMediatorCandidates {
		// 存储 mediator info
		addr, err := imc.Validate()
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		mi := modules.NewMediatorInfo()
		mi.MediatorInfoBase = imc.MediatorInfoBase
		//*mi.MediatorApplyInfo = *imc.MediatorApplyInfo

		err = dag.stableStateRep.StoreMediatorInfo(addr, mi)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		// 将保证金设为0
		md := modules.NewMediatorDeposit()
		md.Status = modules.Agree
		md.Role = modules.Mediator
		md.ApplyEnterTime = time.Unix(unit.Timestamp(), 0).UTC().Format(modules.Layout2)

		byte, err := json.Marshal(md)
		if err != nil {
			return err
		}

		ws.Value = byte
		ws.Key = storage.MediatorDepositKey(imc.AddStr)
		err = dag.stableStateRep.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws, version)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		list[mi.AddStr] = true
	}

	// 存储 initMediatorCandidates/JuryCandidates
	imcB, err := json.Marshal(list)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}
	ws.Value = imcB

	//Mediator
	ws.Key = modules.MediatorList
	err = dag.stableStateRep.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws, version)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	//Jury
	ws.Key = modules.JuryList
	err = dag.stableStateRep.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws, version)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	return nil
}

func (dag *Dag) IsSynced() bool {
	//gp := dag.GetGlobalProp()
	//dgp := dag.GetDynGlobalProp()

	//nowFine := time.Now()
	//now := time.Unix(nowFine.Add(500*time.Millisecond).Unix(), 0)
	now := time.Now()
	// 防止误判，获取之后的第2个生产槽时间
	//nextSlotTime := dag.unstablePropRep.GetSlotTime(gp, dgp, 1)
	//nextSlotTime := dag.unstablePropRep.GetSlotTime(2)
	_, _, _, rep, _ := dag.Memdag.GetUnstableRepositories()
	nextSlotTime := rep.GetSlotTime(2)

	return nextSlotTime.After(now)
}

// author Albert·Gou
func (d *Dag) ChainThreshold() int {
	return d.GetGlobalProp().ChainThreshold()
}

func (d *Dag) PrecedingThreshold() int {
	return d.GetGlobalProp().PrecedingThreshold()
}

func (d *Dag) UnitIrreversibleTime() time.Duration {
	gp := d.GetGlobalProp()
	cp := gp.ChainParameters
	it := uint(gp.ChainThreshold()+int(cp.MaintenanceSkipSlots)) * uint(cp.MediatorInterval)
	return time.Duration(it) * time.Second
}

func (d *Dag) IsIrreversibleUnit(hash common.Hash) bool {
	header, err := d.unstableUnitRep.GetHeaderByHash(hash)
	if err != nil {
		//return false // 存在于memdag，不稳定
		header, err = d.stableUnitRep.GetHeaderByHash(hash)
		if err != nil {
			log.Debugf("UnitRep GetHeaderByHash error:%s", err.Error())
			return false // 不存在该unit
		}
	}

	if header.NumberU64() > d.GetIrreversibleUnitNum(header.GetAssetId()) {
		return false
	}

	return true
}

func (d *Dag) GetIrreversibleUnitNum(id modules.AssetId) uint64 {
	_, idx, err := d.stablePropRep.GetNewestUnit(id)
	if err != nil {
		log.Debugf("stableUnitRep GetNewestUnit error:%s", err.Error())
		return 0
	}

	return idx.Index
}

func (d *Dag) VerifyUnitGroupSign(unitHash common.Hash, groupSign []byte) error {
	header, err := d.GetHeaderByHash(unitHash)
	if err != nil {
		log.Debug(err.Error())
		return err
	}

	pubKey, err := header.GetGroupPubKey()
	if err != nil {
		log.Debug(err.Error())
		return err
	}

	err = bls.Verify(core.Suite, pubKey, unitHash[:], groupSign)
	if err != nil {
		log.Debug("the group signature: " + hexutil.Encode(groupSign) + " of the Unit that hash: " +
			unitHash.Hex() + " is verified that an error has occurred: " + err.Error())
		return err
	}
	return nil
}

// 判断该mediator是下一个产块mediator
func (dag *Dag) IsConsecutiveMediator(nextMediator common.Address) bool {
	dgp := dag.GetDynGlobalProp()

	if !dgp.IsShuffledSchedule && nextMediator.Equal(dgp.LastMediator) {
		return true
	}

	return false
}

// 计算最近64个生产slots的mediator参与度，不包括当前unit
// Calculate the percent of unit production slots that were missed in the
// past 64 units, not including the current unit.
func (dag *Dag) MediatorParticipationRate() uint32 {
	popCount := func(x uint64) uint8 {
		m := []uint64{
			0x5555555555555555,
			0x3333333333333333,
			0x0F0F0F0F0F0F0F0F,
			0x00FF00FF00FF00FF,
			0x0000FFFF0000FFFF,
			0x00000000FFFFFFFF,
		}

		var i, w uint8
		for i, w = 0, 1; i < 6; i, w = i+1, w+w {
			x = (x & m[i]) + ((x >> w) & m[i])
		}

		return uint8(x)
	}

	recentSlotsFilled := dag.GetDynGlobalProp().RecentSlotsFilled
	participationRate := core.PalletOne100Percent * int(popCount(recentSlotsFilled)) / 64

	return uint32(participationRate)
}
