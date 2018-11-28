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
	"time"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/share"
	"github.com/dedis/kyber/share/dkg/pedersen"
	"github.com/dedis/kyber/share/vss/pedersen"
	"github.com/dedis/kyber/sign/bls"
	"github.com/dedis/kyber/sign/tbls"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

func (mp *MediatorPlugin) StartVSSProtocol() {
	go log.Info("Start completing the VSS protocol.")

	go mp.BroadcastVSSDeals()

	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	select {
	case <-mp.quit:
		return
	case <-timeout.C:
		go mp.endVSSProtocol(true)
	}
}

func (mp *MediatorPlugin) endVSSProtocol(timeout bool) (completed bool) {
	// todo 判断本地所有的dkg是否都完成了vss协议, 并发送协议完成的消息

	completed = mp.areCertified()
	if !completed && timeout {
		mp.setTimeout()
	}

	return
}

func (mp *MediatorPlugin) setTimeout() {
	for _, dkg := range mp.activeDKGs {
		dkg.SetTimeout()
	}
}

func (mp *MediatorPlugin) areCertified() (certified bool) {
	mp.ctfLock.Lock()

	for _, flag := range mp.certifiedFlag {
		certified = flag
		if !certified {
			break
		}
	}

	mp.ctfLock.Unlock()
	return
}

func (mp *MediatorPlugin) getLocalActiveDKG(add common.Address) *dkg.DistKeyGenerator {
	if !mp.IsLocalActiveMediator(add) {
		log.Error(fmt.Sprintf("The following mediator is not local active mediator: %v", add.String()))
		return nil
	}

	dkg, ok := mp.activeDKGs[add]
	if !ok || dkg == nil {
		log.Error(fmt.Sprintf("The following mediator`s dkg is not existed: %v", add.String()))
		return nil
	}

	return dkg
}

func (mp *MediatorPlugin) BroadcastVSSDeals() {
	for localMed, dkg := range mp.activeDKGs {
		deals, err := dkg.Deals()
		if err != nil {
			log.Error(err.Error())
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

func (mp *MediatorPlugin) ToProcessDeal(deal *VSSDealEvent) error {
	select {
	case <-mp.quit:
		return errTerminated
	default:
		go mp.processVSSDeal(deal)
		return nil
	}
}

func (mp *MediatorPlugin) processVSSDeal(dealEvent *VSSDealEvent) {
	dag := mp.dag
	localMed := dag.GetActiveMediatorAddr(dealEvent.DstIndex)

	dkgr := mp.getLocalActiveDKG(localMed)
	if dkgr == nil {
		return
	}

	deal := dealEvent.Deal

	resp, err := dkgr.ProcessDeal(deal)
	if err != nil {
		log.Error(err.Error())
		return
	}

	vrfrMed := dag.GetActiveMediatorAddr(int(deal.Index))
	go mp.processResponseLoop(localMed, vrfrMed)

	if resp.Response.Status != vss.StatusApproval {
		log.Error(fmt.Sprintf("DKG: own deal gave a complaint: %v", localMed.String()))
		return
	}

	respEvent := VSSResponseEvent{
		Resp: resp,
	}
	go mp.vssResponseFeed.Send(respEvent)
}

func (mp *MediatorPlugin) ToProcessResponse(resp *VSSResponseEvent) error {
	select {
	case <-mp.quit:
		return errTerminated
	default:
		go mp.addToResponseBuf(resp)
		return nil
	}
}

func (mp *MediatorPlugin) addToResponseBuf(respEvent *VSSResponseEvent) {
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
	dkgr := mp.getLocalActiveDKG(localMed)
	if dkgr == nil {
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
			log.Error(err.Error())
			return false
		}

		if jstf != nil {
			log.Error(fmt.Sprintf("DKG: wrong Process Response: %v", localMed.String()))
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

				mp.ctfLock.Lock()
				mp.certifiedFlag[localMed] = true
				mp.ctfLock.Unlock()
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

					go mp.endVSSProtocol(false)
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
		log.Error("the following mediator is not local: %v", localMed.Str())
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
	for _, localMed := range lams {
		mp.toTBLSSignBuf[localMed] <- newUnit
	}
}

func (mp *MediatorPlugin) SubscribeSigShareEvent(ch chan<- SigShareEvent) event.Subscription {
	return mp.sigShareScope.Track(mp.sigShareFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) signTBLSLoop(localMed common.Address) {
	dkgr := mp.getLocalActiveDKG(localMed)
	if dkgr == nil {
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Error(err.Error())
		return
	}

	dag := mp.dag
	newUnitBuf := mp.toTBLSSignBuf[localMed]

	signTBLS := func(newUnit *modules.Unit) (sigShare []byte, success bool) {
		if !dag.ValidateUnitExceptGroupSig(newUnit, false) {
			return
		}

		// todo 判断父unit是否被确认

		var err error
		hash := newUnit.Hash()

		sigShare, err = tbls.Sign(mp.suite, dks.PriShare(), hash[:])
		if err != nil {
			log.Error(err.Error())
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
				// todo 后面改为由p2p广播
				//go mp.sigShareFeed.Send(SigShareEvent{UnitHash: newUnit.Hash(), SigShare: sigShare})
				go mp.addToTBLSRecoverBuf(newUnit, sigShare)
			}
		}
	}
}

func (mp *MediatorPlugin) ToTBLSRecover(sigShare *SigShareEvent) error {
	select {
	case <-mp.quit:
		return errTerminated
	default:
		localMed, _ := mp.dag.GetUnitByHash(sigShare.UnitHash)
		go mp.addToTBLSRecoverBuf(localMed, sigShare.SigShare)
		//go mp.addToTBLSRecoverBuf(sigShare.UnitHash, sigShare.SigShare)
		return nil
	}
}

// 收集签名分片
func (mp *MediatorPlugin) addToTBLSRecoverBuf(newUnit *modules.Unit, sigShare []byte) {
	//func (mp *MediatorPlugin) addToTBLSRecoverBuf(newUnitHash common.Hash, sigShare []byte) {
	//	dag := mp.dag
	//	newUnit := dag.GetUnit(newUnitHash)
	if newUnit == nil {
		log.Error("newUnit is nil!")
	}

	newUnitHash := newUnit.UnitHash
	localMed := newUnit.UnitAuthor()

	medSigShareBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Error("the following mediator is not local: %v", localMed.Str())
		return
	}

	// 当buf不存在时，说明已经recover出群签名，忽略该签名分片
	sigShareSet, ok := medSigShareBuf[newUnitHash]
	if !ok {
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

	dkgr := mp.getLocalActiveDKG(localMed)
	if dkgr == nil {
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Error(err.Error())
		return
	}

	suite := mp.suite
	pubPoly := share.NewPubPoly(suite, suite.Point().Base(), dks.Commitments())
	groupSig, err := tbls.Recover(suite, pubPoly, unitHash[:], sigShareSet.popSigShares(), curThreshold, aSize)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Debug("Recovered the Unit that hash: " + unitHash.Hex() +
		" the group signature: " + hexutil.Encode(groupSig))

	// recover后 删除buf
	delete(mp.toTBLSRecoverBuf[localMed], unitHash)
	go mp.groupSigFeed.Send(GroupSigEvent{UnitHash: unitHash, GroupSig: groupSig})

	// todo 后面改为由dag模块调用
	go mp.VerifyUnitGroupSig(dks.Public(), unitHash, groupSig)
}

func (mp *MediatorPlugin) VerifyUnitGroupSig(groupPublicKey kyber.Point, unitHash common.Hash, groupSig []byte) error {
	//func (mp *MediatorPlugin) VerifyUnitGroupSig(groupPublicKey kyber.Point, newUnit *modules.Unit) error {
	err := bls.Verify(mp.suite, groupPublicKey, unitHash[:], groupSig)
	if err == nil {
		//log.Debug("the group signature: " + hexutil.Encode(groupSig) +
		//	" of the Unit that hash: " + unitHash.Hex() + " is verified through!")
	} else {
		log.Error("the group signature: " + hexutil.Encode(groupSig) + " of the Unit that hash: " +
			unitHash.Hex() + " is verified that an error has occurred: " + err.Error())
	}

	return err
}
