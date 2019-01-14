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
 */

package mediatorplugin

import (
	"fmt"

	"github.com/dedis/kyber/share"
	"github.com/dedis/kyber/share/dkg/pedersen"
	"github.com/dedis/kyber/share/vss/pedersen"
	"github.com/dedis/kyber/sign/tbls"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

func (mp *MediatorPlugin) startVSSProtocol() {
	dag := mp.dag
	if !mp.productionEnabled && !dag.IsSynced() {
		log.Debug("we're not synced")
		return
	}

	log.Debug("Start completing the VSS protocol.")
	go mp.BroadcastVSSDeals()
}

func (mp *MediatorPlugin) getLocalActiveDKG(add common.Address) (*dkg.DistKeyGenerator, error) {
	if !mp.IsLocalActiveMediator(add) {
		return nil, fmt.Errorf("the following mediator is not local active mediator: %v", add.String())
	}

	dkg, ok := mp.activeDKGs[add]
	if !ok || dkg == nil {
		return nil, fmt.Errorf("the following mediator`s dkg is not existed: %v", add.String())
	}

	return dkg, nil
}

func (mp *MediatorPlugin) BroadcastVSSDeals() {
	for localMed, dkg := range mp.activeDKGs {
		deals, err := dkg.Deals()
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		go mp.processResponseLoop(localMed, localMed)
		log.Debugf("the mediator(%v) broadcast vss deals", localMed.Str())

		for index, deal := range deals {
			event := VSSDealEvent{
				DstIndex: index,
				Deal:     deal,
			}

			go mp.vssDealFeed.Send(event)
		}
	}
}

func (mp *MediatorPlugin) SubscribeVSSDealEvent(ch chan<- VSSDealEvent) event.Subscription {
	return mp.vssDealScope.Track(mp.vssDealFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) ProcessVSSDeal(dealEvent *VSSDealEvent) error {
	dag := mp.dag
	localMed := dag.GetActiveMediatorAddr(dealEvent.DstIndex)

	dkgr, err := mp.getLocalActiveDKG(localMed)
	if err != nil {
		log.Debug(err.Error())
		return err
	}

	deal := dealEvent.Deal

	resp, err := dkgr.ProcessDeal(deal)
	if err != nil {
		log.Debug(err.Error())
		return err
	}

	vrfrMed := dag.GetActiveMediatorAddr(int(deal.Index))
	log.Debugf("the mediator(%v) received the vss deal from the mediator(%v)", localMed.Str(), vrfrMed.Str())
	go mp.processResponseLoop(localMed, vrfrMed)

	if resp.Response.Status != vss.StatusApproval {
		err = fmt.Errorf("DKG: own deal gave a complaint: %v", localMed.String())
		log.Debug(err.Error())
		return err
	}

	respEvent := VSSResponseEvent{
		Resp: resp,
	}
	go mp.vssResponseFeed.Send(respEvent)
	log.Debugf("the mediator(%v) broadcast the vss response to the mediator(%v)", localMed.Str(), vrfrMed.Str())

	return nil
}

func (mp *MediatorPlugin) AddToResponseBuf(respEvent *VSSResponseEvent) {
	resp := respEvent.Resp
	lams := mp.GetLocalActiveMediators()
	for _, localMed := range lams {
		dag := mp.dag

		//ignore the message from myself
		srcIndex := resp.Response.Index
		srcMed := dag.GetActiveMediatorAddr(int(srcIndex))
		if srcMed == localMed {
			continue
		}

		vrfrMed := dag.GetActiveMediatorAddr(int(resp.Index))
		log.Debugf("the mediator(%v) received the vss response from the mediator(%v) to the mediator(%v)",
			localMed.Str(), srcMed.Str(), vrfrMed.Str())

		if _, ok := mp.respBuf[localMed][vrfrMed]; !ok {
			log.Debugf("the mediator(%v)'s respBuf corresponding the mediator(%v) is not initialized",
				localMed.Str(), vrfrMed.Str())
		}
		mp.respBuf[localMed][vrfrMed] <- resp
	}
}

func (mp *MediatorPlugin) SubscribeVSSResponseEvent(ch chan<- VSSResponseEvent) event.Subscription {
	return mp.vssResponseScope.Track(mp.vssResponseFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) processResponseLoop(localMed, vrfrMed common.Address) {
	dkgr, err := mp.getLocalActiveDKG(localMed)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	aSize := mp.dag.ActiveMediatorsCount()
	respCount := 0
	// localMed 对 vrfrMed 的 response 在 ProcessDeal 生成 response 时 自动处理了
	if vrfrMed != localMed {
		respCount++
	}

	processResp := func(resp *dkg.Response) bool {
		jstf, err := dkgr.ProcessResponse(resp)
		if err != nil {
			log.Debug(err.Error())
			return false
		}

		if jstf != nil {
			log.Debug(fmt.Sprintf("DKG: wrong Process Response: %v", localMed.String()))
			return false
		}

		return true
	}

	isFinishedAndCertified := func() (finished, certified bool) {
		respCount++

		if respCount == aSize-1 {
			finished = true

			if dkgr.Certified() {
				log.Debug(fmt.Sprintf("%v's DKG verification passed!", localMed.Str()))

				certified = true
			}
		}

		return
	}

	if _, ok := mp.respBuf[localMed][vrfrMed]; !ok {
		log.Debugf("the mediator(%v)'s respBuf corresponding the mediator(%v) is not initialized",
			localMed.Str(), vrfrMed.Str())
	}

	log.Debugf("the mediator(%v) run the loop to process response regarding the mediator(%v)",
		localMed.Str(), vrfrMed.Str())
	respCh := mp.respBuf[localMed][vrfrMed]

	for {
		select {
		case <-mp.quit:
			return
		case resp := <-respCh:
			processResp(resp)
			finished, certified := isFinishedAndCertified()
			if finished {
				delete(mp.respBuf[localMed], vrfrMed)

				if certified {
					go mp.signTBLSLoop(localMed)
					go mp.recoverUnitsTBLS(localMed)

					delete(mp.respBuf, localMed)
				}

				return
			}
		}
	}
}

func (mp *MediatorPlugin) recoverUnitsTBLS(localMed common.Address) {
	medSigShareBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Debug(fmt.Sprintf("the following mediator has no signature shares yet: %v", localMed.Str()))
		return
	}

	for newUnitHash := range medSigShareBuf {
		go mp.recoverUnitTBLS(localMed, newUnitHash)
	}
}

func (mp *MediatorPlugin) AddToTBLSSignBuf(newUnit *modules.Unit) {
	lams := mp.GetLocalActiveMediators()
	curThrshd := mp.dag.ChainThreshold()

	for _, localMed := range lams {
		log.Debugf("the mediator(%v) received a unit to be grouped sign: %v",
			localMed.Str(), newUnit.UnitHash.TerminalString())

		if _, ok := mp.toTBLSSignBuf[localMed]; !ok {
			mp.toTBLSSignBuf[localMed] = make(chan *modules.Unit, curThrshd)
		}

		mp.toTBLSSignBuf[localMed] <- newUnit
	}
}

func (mp *MediatorPlugin) SubscribeSigShareEvent(ch chan<- SigShareEvent) event.Subscription {
	return mp.sigShareScope.Track(mp.sigShareFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) signTBLSLoop(localMed common.Address) {
	dkgr, err := mp.getLocalActiveDKG(localMed)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debug(err.Error())
		return
	}

	dag := mp.dag
	if _, ok := mp.toTBLSSignBuf[localMed]; !ok {
		mp.toTBLSSignBuf[localMed] = make(chan *modules.Unit, dag.ChainThreshold())
	}

	newUnitBuf := mp.toTBLSSignBuf[localMed]
	log.Debugf("the mediator(%v) run the loop of TBLS sign", localMed.Str())

	signTBLS := func(newUnit *modules.Unit) (sigShare []byte, success bool) {
		// 1.如果单元没有群公钥， 则跳过群签名
		_, err = newUnit.GroupPubKey()
		if err != nil {
			log.Debug(err.Error())
			return
		}

		// 2. 验证本 unit
		if !dag.ValidateUnitExceptGroupSig(newUnit, false) {
			log.Debugf("the unit validate except group sig fail: %v", newUnit.UnitHash.TerminalString())
			return
		}

		// 3. 判断父 unit 是否不可逆
		parentHash := newUnit.ParentHash()[0]
		if !dag.IsIrreversibleUnit(parentHash) {
			log.Debugf("the unit's(%v) parent unit(%v) is not irreversible",
				newUnit.UnitHash.TerminalString(), parentHash.TerminalString())
			return
		}

		var err error
		hash := newUnit.Hash()

		sigShare, err = tbls.Sign(mp.suite, dks.PriShare(), hash[:])
		if err != nil {
			log.Debug(err.Error())
			return
		}

		success = true
		log.Debugf("the mediator(%v) group-signed the unit(%v)", localMed.Str(),
			newUnit.UnitHash.TerminalString())
		return
	}

	// todo 换届后，如果该mediator不是活跃的话，则到达一定时刻强制关闭循环
	for {
		select {
		case <-mp.quit:
			return
		case newUnit := <-newUnitBuf:
			sigShare, success := signTBLS(newUnit)
			if success {
				go mp.sigShareFeed.Send(SigShareEvent{UnitHash: newUnit.Hash(), SigShare: sigShare})
			}
		}
	}
}

// 收集签名分片
func (mp *MediatorPlugin) AddToTBLSRecoverBuf(newUnitHash common.Hash, sigShare []byte) error {
	log.Debugf("received the sign shares of the unit: %v", newUnitHash.TerminalString())

	dag := mp.dag
	newUnit, err := dag.GetUnitByHash(newUnitHash)
	if newUnit == nil || err != nil {
		err = fmt.Errorf("fail to get unit by hash in dag: %v", newUnitHash.TerminalString())
		log.Debug(err.Error())
		return err
	}

	//newUnitHash := newUnit.UnitHash
	localMed := newUnit.Author()

	medSigShareBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		err = fmt.Errorf("the following mediator's toTBLSRecoverBuf has not initialized yet: %v", localMed.Str())
		log.Debug(err.Error())
		return err
	}

	// 当buf不存在时，说明已经recover出群签名, 或者已经过了unit确认时间，忽略该签名分片
	sigShareSet, ok := medSigShareBuf[newUnitHash]
	if !ok {
		err = fmt.Errorf("the unit already has recovered the group signature: %v", newUnitHash.TerminalString())
		log.Debugf(err.Error())
		return err
	}

	sigShareSet.append(sigShare)

	// recover群签名
	go mp.recoverUnitTBLS(localMed, newUnitHash)
	return nil
}

func (mp *MediatorPlugin) SubscribeGroupSigEvent(ch chan<- GroupSigEvent) event.Subscription {
	return mp.groupSigScope.Track(mp.groupSigFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) recoverUnitTBLS(localMed common.Address, unitHash common.Hash) {
	sigShareSet, ok := mp.toTBLSRecoverBuf[localMed][unitHash]
	if !ok {
		log.Debugf(fmt.Sprintf("the following mediator has no sign shares yet: %v", localMed.Str()))
		return
	}

	sigShareSet.lock()
	defer sigShareSet.unlock()

	// 为了保证多协程安全， 加锁后，再判断一次
	if _, ok = mp.toTBLSRecoverBuf[localMed][unitHash]; !ok {
		return
	}

	dag := mp.dag
	unit, err := dag.GetUnitByHash(unitHash)
	if unit == nil || err != nil {
		err = fmt.Errorf("fail to get unit by hash in dag: %v", unitHash.TerminalString())
		log.Debug(err.Error())
		return
	}

	var mSize, threshold int
	// 判断是否是换届前的单元
	if unit.Timestamp() <= dag.GetDynGlobalProp().LastMaintenanceTime {
		mSize = dag.PrecedingMediatorsCount()
		threshold = dag.PrecedingThreshold()
	} else {
		mSize = dag.ActiveMediatorsCount()
		threshold = dag.ChainThreshold()
	}

	if sigShareSet.len() < threshold {
		log.Debugf("the count of sign shares of the unit(%v) does not reach the threshold(%v)",
			unitHash.TerminalString(), threshold)
		return
	}

	dkgr, err := mp.getLocalActiveDKG(localMed)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debug(err.Error())
		return
	}

	suite := mp.suite
	pubPoly := share.NewPubPoly(suite, suite.Point().Base(), dks.Commitments())
	groupSig, err := tbls.Recover(suite, pubPoly, unitHash[:], sigShareSet.popSigShares(), threshold, mSize)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	log.Debugf("Recovered the Unit(%v)'s the group signature: ",
		unitHash.TerminalString(), hexutil.Encode(groupSig))

	// recover后 删除buf
	delete(mp.toTBLSRecoverBuf[localMed], unitHash)
	mp.dag.SetUnitGroupSign(unitHash, groupSig, mp.ptn.TxPool())
	go mp.groupSigFeed.Send(GroupSigEvent{UnitHash: unitHash, GroupSig: groupSig})
}
