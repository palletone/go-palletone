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
	log.Debug("Start completing the VSS protocol.")

	go mp.BroadcastVSSDeals()
}

func (mp *MediatorPlugin) getLocalActiveDKG(add common.Address) (*dkg.DistKeyGenerator, error) {
	if !mp.IsLocalActiveMediator(add) {
		return nil, fmt.Errorf("The following mediator is not local active mediator: %v", add.String())
	}

	dkg, ok := mp.activeDKGs[add]
	if !ok || dkg == nil {
		return nil, fmt.Errorf("The following mediator`s dkg is not existed: %v", add.String())
	}

	return dkg, nil
}

func (mp *MediatorPlugin) BroadcastVSSDeals() {
	for localMed, dkg := range mp.activeDKGs {
		deals, err := dkg.Deals()
		if err != nil {
			log.Debug(err.Error())
		}

		for index, deal := range deals {
			event := VSSDealEvent{
				DstIndex: index,
				Deal:     deal,
			}

			go mp.vssDealFeed.Send(event)
		}

		go mp.processResponseLoop(localMed, localMed)
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
		mp.respBuf[localMed][vrfrMed] <- resp
	}
}

func (mp *MediatorPlugin) SubscribeVSSResponseEvent(ch chan<- VSSResponseEvent) event.Subscription {
	return mp.vssResponseScope.Track(mp.vssResponseFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) processResponseLoop(localMed, vrfrMed common.Address) {
	dkgr, err := mp.getLocalActiveDKG(localMed)
	if err != nil {
		return
	}

	aSize := mp.dag.GetActiveMediatorCount()
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
		log.Debug(fmt.Sprintf("the following mediator also has no signature shares: %v", localMed.Str()))
		return
	}

	for newUnitHash := range medSigShareBuf {
		go mp.recoverUnitTBLS(localMed, newUnitHash)
	}
}

func (mp *MediatorPlugin) ToUnitTBLSSign(newUnit *modules.Unit) error {
	select {
	case <-mp.quit:
		return errTerminated
	default:
		go mp.addToTBLSSignBuf(newUnit)
		return nil
	}
}

func (mp *MediatorPlugin) addToTBLSSignBuf(newUnit *modules.Unit) {
	lams := mp.GetLocalActiveMediators()
	curThrshd := mp.dag.ChainThreshold()

	for _, localMed := range lams {
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
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debug(err.Error())
		return
	}

	dag := mp.dag
	newUnitBuf := mp.toTBLSSignBuf[localMed]

	signTBLS := func(newUnit *modules.Unit) (sigShare []byte, success bool) {
		// 1. 验证本 unit
		if !dag.ValidateUnitExceptGroupSig(newUnit, false) {
			return
		}

		// 2. 判断父 unit 是否不可逆
		if !dag.IsIrreversibleUnit(newUnit.ParentHash()[0]) {
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
		return
	}

	for {
		select {
		case <-mp.quit:
			return
		case newUnit := <-newUnitBuf:
			sigShare, success := signTBLS(newUnit)
			if success {
				go mp.sigShareFeed.Send(SigShareEvent{UnitHash: newUnit.Hash(), SigShare: sigShare})
				//go mp.addToTBLSRecoverBuf(newUnit, sigShare)
			}
		}
	}
}

func (mp *MediatorPlugin) ToTBLSRecover(sigShare *SigShareEvent) error {
	select {
	case <-mp.quit:
		return errTerminated
	default:
		//localMed, _ := mp.dag.GetUnitByHash(sigShare.UnitHash)
		//go mp.addToTBLSRecoverBuf(localMed, sigShare.SigShare)
		go mp.addToTBLSRecoverBuf(sigShare.UnitHash, sigShare.SigShare)
		return nil
	}
}

// 收集签名分片
//func (mp *MediatorPlugin) addToTBLSRecoverBuf(newUnit *modules.Unit, sigShare []byte) {
func (mp *MediatorPlugin) addToTBLSRecoverBuf(newUnitHash common.Hash, sigShare []byte) {
	dag := mp.dag
	newUnit, err := dag.GetUnitByHash(newUnitHash)
	if newUnit == nil || err != nil {
		log.Debug(fmt.Sprintf("fail to get unit by hash in dag: %v", newUnitHash))
		return
	}

	//newUnitHash := newUnit.UnitHash
	localMed := newUnit.Author()

	medSigShareBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Debug(fmt.Sprintf("the following mediator's toTBLSRecoverBuf has not initialized yet: %v", localMed.Str()))
		return
	}

	// 当buf不存在时，说明已经recover出群签名，忽略该签名分片
	sigShareSet, ok := medSigShareBuf[newUnitHash]
	if !ok {
		log.Debugf("the unit already has recovered the group signature: %v", newUnitHash)
		return
	}

	sigShareSet.append(sigShare)

	// recover群签名
	go mp.recoverUnitTBLS(localMed, newUnitHash)
}

func (mp *MediatorPlugin) SubscribeGroupSigEvent(ch chan<- GroupSigEvent) event.Subscription {
	return mp.groupSigScope.Track(mp.groupSigFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) recoverUnitTBLS(localMed common.Address, unitHash common.Hash) {
	sigShareSet, ok := mp.toTBLSRecoverBuf[localMed][unitHash]
	if !ok {
		log.Debug(fmt.Sprintf("the following mediator also has no signature shares: %v", localMed.Str()))
		return
	}

	sigShareSet.lock()
	defer sigShareSet.unlock()

	// 为了保证多协程安全， 加锁后，再判断一次
	if _, ok = mp.toTBLSRecoverBuf[localMed][unitHash]; !ok {
		return
	}

	dag := mp.dag
	aSize := dag.GetActiveMediatorCount()
	curThreshold := dag.ChainThreshold()

	if sigShareSet.len() < curThreshold {
		return
	}

	dkgr, err := mp.getLocalActiveDKG(localMed)
	if err != nil {
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debug(err.Error())
		return
	}

	suite := mp.suite
	pubPoly := share.NewPubPoly(suite, suite.Point().Base(), dks.Commitments())
	groupSig, err := tbls.Recover(suite, pubPoly, unitHash[:], sigShareSet.popSigShares(), curThreshold, aSize)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	log.Debug("Recovered the Unit that hash: " + unitHash.TerminalString() +
		" the group signature: " + hexutil.Encode(groupSig))

	// recover后 删除buf
	delete(mp.toTBLSRecoverBuf[localMed], unitHash)
	go mp.groupSigFeed.Send(GroupSigEvent{UnitHash: unitHash, GroupSig: groupSig})
}
