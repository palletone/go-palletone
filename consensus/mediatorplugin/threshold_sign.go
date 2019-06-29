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
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/share/dkg/pedersen"
	"go.dedis.ch/kyber/v3/share/vss/pedersen"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

func (mp *MediatorPlugin) startVSSProtocol() {
	dag := mp.dag
	if !mp.productionEnabled && !dag.IsSynced() {
		log.Debugf("we're not synced")
		return
	}

	log.Debugf("Start completing the VSS protocol.")
	go mp.BroadcastVSSDeals()
}

func (mp *MediatorPlugin) getLocalActiveDKG(add common.Address) (*dkg.DistKeyGenerator, error) {
	if !mp.IsLocalActiveMediator(add) {
		return nil, fmt.Errorf("the mediator(%v) is not local active mediator", add.String())
	}

	dkg, ok := mp.activeDKGs[add]
	if !ok || dkg == nil {
		return nil, fmt.Errorf("the mediator(%v)'s dkg is not existed", add.String())
	}

	return dkg, nil
}

func (mp *MediatorPlugin) BroadcastVSSDeals() {
	for localMed, dkg := range mp.activeDKGs {
		deals, err := dkg.Deals()
		if err != nil {
			log.Debugf(err.Error())
			continue
		}

		go mp.processResponseLoop(localMed, localMed)
		log.Debugf("the mediator(%v) broadcast vss deals", localMed.Str())

		for index, deal := range deals {
			event := VSSDealEvent{
				DstIndex: uint(index),
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
	if !mp.groupSigningEnabled {
		return nil
	}

	dag := mp.dag
	localMed := dag.GetActiveMediatorAddr(int(dealEvent.DstIndex))

	dkgr, err := mp.getLocalActiveDKG(localMed)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	deal := dealEvent.Deal

	resp, err := dkgr.ProcessDeal(deal)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	vrfrMed := dag.GetActiveMediatorAddr(int(deal.Index))
	log.Debugf("the mediator(%v) received the vss deal from the mediator(%v)",
		localMed.Str(), vrfrMed.Str())
	go mp.processResponseLoop(localMed, vrfrMed)

	if resp.Response.Status != vss.StatusApproval {
		err = fmt.Errorf("DKG: own deal gave a complaint: %v", localMed.String())
		log.Debugf(err.Error())
		return err
	}

	respEvent := VSSResponseEvent{
		Resp: resp,
	}
	go mp.vssResponseFeed.Send(respEvent)
	log.Debugf("the mediator(%v) broadcast the vss response to the mediator(%v)",
		localMed.Str(), vrfrMed.Str())

	return nil
}

func (mp *MediatorPlugin) AddToResponseBuf(respEvent *VSSResponseEvent) {
	if !mp.groupSigningEnabled {
		return
	}

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
		log.Debugf(err.Error())
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
			log.Debugf(err.Error())
			return false
		}

		if jstf != nil {
			log.Debugf("DKG: wrong Process Response: %v", localMed.String())
			return false
		}

		return true
	}

	isFinishedAndCertified := func() (finished, certified bool) {
		respCount++

		if respCount == aSize-1 {
			finished = true

			if dkgr.Certified() {
				log.Debugf("the mediator(%v)'s DKG verification passed!", localMed.Str())

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
					go mp.signUnitsTBLS(localMed)
					go mp.recoverUnitsTBLS(localMed)

					delete(mp.respBuf, localMed)
				}

				return
			}
		}
	}
}

func (mp *MediatorPlugin) signUnitsTBLS(localMed common.Address) {
	medUnitsBuf, ok := mp.toTBLSSignBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no units to sign TBLS yet", localMed.Str())
		return
	}

	rangeFn := func(key, value interface{}) bool {
		newUnitHash, ok := key.(common.Hash)
		if !ok {
			log.Debugf("key converted to Hash failed")
			return true
		}
		go mp.signUnitTBLS(localMed, newUnitHash)

		return true
	}

	medUnitsBuf.Range(rangeFn)
}

func (mp *MediatorPlugin) recoverUnitsTBLS(localMed common.Address) {
	mp.recoverBufLock.RLock()
	defer mp.recoverBufLock.RUnlock()

	sigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no signature shares to recover group sign yet", localMed.Str())
		return
	}

	for unitHash := range sigSharesBuf {
		go mp.recoverUnitTBLS(localMed, unitHash)
	}
}

func (mp *MediatorPlugin) AddToTBLSSignBufs(newUnit *modules.Unit) {
	if !mp.groupSigningEnabled {
		return
	}

	var ms []common.Address
	if newUnit.Timestamp() <= mp.dag.LastMaintenanceTime() {
		ms = mp.GetLocalPrecedingMediators()
	} else {
		ms = mp.GetLocalActiveMediators()
	}

	for _, localMed := range ms {
		log.Debugf("the mediator(%v) received a unit(%v) to be group-signed",
			localMed.Str(), newUnit.UnitHash.TerminalString())
		go mp.addToTBLSSignBuf(localMed, newUnit)
	}
}

func (mp *MediatorPlugin) addToTBLSSignBuf(localMed common.Address, newUnit *modules.Unit) {
	if _, ok := mp.toTBLSSignBuf[localMed]; !ok {
		mp.toTBLSSignBuf[localMed] = new(sync.Map)
	}

	unitHash := newUnit.UnitHash
	mp.toTBLSSignBuf[localMed].LoadOrStore(unitHash, newUnit)
	go mp.signUnitTBLS(localMed, unitHash)

	// 当 unit 过了确认时间后，及时删除待群签名的 unit，防止内存溢出
	expiration := mp.dag.UnitIrreversibleTime()
	deleteBuf := time.NewTimer(expiration)

	select {
	case <-mp.quit:
		return
	case <-deleteBuf.C:
		if _, ok := mp.toTBLSSignBuf[localMed].Load(unitHash); ok {
			log.Debugf("the unit(%v) has expired confirmation time, no longer need the mediator(%v)"+
				" to sign-group", unitHash.TerminalString(), localMed.Str())
			mp.toTBLSSignBuf[localMed].Delete(unitHash)
		}
	}
}

func (mp *MediatorPlugin) SubscribeSigShareEvent(ch chan<- SigShareEvent) event.Subscription {
	return mp.sigShareScope.Track(mp.sigShareFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) signUnitTBLS(localMed common.Address, unitHash common.Hash) {
	medUnitsBuf, ok := mp.toTBLSSignBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v) has no units to sign TBLS yet", localMed.Str())
		return
	}

	dag := mp.dag
	var (
		dkgr    *dkg.DistKeyGenerator
		newUnit *modules.Unit
	)
	// 1. 获取群签名所需数据
	{
		value, ok := medUnitsBuf.Load(unitHash)
		if !ok {
			log.Debugf("the mediator(%v) has no unit(%v) to sign TBLS",
				localMed.Str(), unitHash.TerminalString())
			return
		}

		newUnit, ok = value.(*modules.Unit)
		if !ok {
			log.Debugf("value converted to Unit pointer failed")
			return
		}

		// 判断是否是换届前的单元
		if newUnit.Timestamp() <= dag.LastMaintenanceTime() {
			dkgr, ok = mp.precedingDKGs[localMed]
		} else {
			dkgr, ok = mp.activeDKGs[localMed]
		}

		if !ok {
			log.Debugf("the mediator(%v)'s dkg is not existed", localMed.Str())
			return
		}
	}

	// 2. 判断群签名的相关条件
	{
		// 1.如果单元没有群公钥， 则跳过群签名
		_, err := newUnit.GroupPubKey()
		if err != nil {
			log.Debugf(err.Error())
			return
		}

		// 2. 验证本 unit
		// 已经被验证了，不需要再验证了
		//if dag.ValidateUnitExceptGroupSig(newUnit) != nil {
		//	log.Debugf("the unit validate except group sig fail: %v", newUnit.UnitHash.TerminalString())
		//	return
		//}

		// 3. 判断父 unit 是否不可逆
		parentHash := newUnit.ParentHash()[0]
		if !dag.IsIrreversibleUnit(parentHash) {
			log.Debugf("the unit's(%v) parent unit(%v) is not irreversible",
				newUnit.UnitHash.TerminalString(), parentHash.TerminalString())
			return
		}
	}

	// 3. 群签名
	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	sigShare, err := tbls.Sign(mp.suite, dks.PriShare(), unitHash[:])
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	// 4. 群签名成功后的处理
	log.Debugf("the mediator(%v) signed-group the unit(%v)", localMed.Str(),
		newUnit.UnitHash.TerminalString())
	mp.toTBLSSignBuf[localMed].Delete(unitHash)
	go mp.sigShareFeed.Send(SigShareEvent{UnitHash: newUnit.Hash(), SigShare: sigShare})
}

// 收集签名分片
func (mp *MediatorPlugin) AddToTBLSRecoverBuf(newUnitHash common.Hash, sigShare []byte) error {
	if !mp.groupSigningEnabled {
		return nil
	}

	log.Debugf("received the sign shares of the unit(%v)", newUnitHash.TerminalString())

	dag := mp.dag
	newUnit, err := dag.GetUnitByHash(newUnitHash)
	if newUnit == nil || err != nil {
		err = fmt.Errorf("fail to get unit by hash in dag: %v", newUnitHash.TerminalString())
		log.Debugf(err.Error())
		return err
	}

	localMed := newUnit.Author()
	mp.recoverBufLock.RLock()
	defer mp.recoverBufLock.RUnlock()

	medSigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		err = fmt.Errorf("the mediator(%v)'s toTBLSRecoverBuf has not initialized yet", localMed.Str())
		log.Debugf(err.Error())
		return err
	}

	// 当buf不存在时，说明已经recover出群签名, 或者已经过了unit确认时间，忽略该签名分片
	sigShareSet, ok := medSigSharesBuf[newUnitHash]
	if !ok {
		err = fmt.Errorf("the unit(%v) has already recovered the group signature",
			newUnitHash.TerminalString())
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
	mp.recoverBufLock.Lock()
	defer mp.recoverBufLock.Unlock()

	// 1. 获取所有的签名分片
	sigSharesBuf, ok := mp.toTBLSRecoverBuf[localMed]
	if !ok {
		log.Debugf("the mediator((%v) has no signature shares to recover group sign yet", localMed.Str())
		return
	}

	sigShareSet, ok := sigSharesBuf[unitHash]
	if !ok {
		log.Debugf("the mediator(%v) has no sign shares corresponding unit(%v) yet",
			localMed.Str(), unitHash.TerminalString())
		return
	}

	sigShareSet.lock()
	defer sigShareSet.unlock()

	// 为了保证多协程安全， 加锁后，再判断一次
	if _, ok = mp.toTBLSRecoverBuf[localMed][unitHash]; !ok {
		return
	}

	// 2. 获取阈值、mediator数量、DKG
	var (
		mSize, threshold int
		dkgr             *dkg.DistKeyGenerator
	)
	{
		dag := mp.dag
		unit, err := dag.GetUnitByHash(unitHash)
		if unit == nil || err != nil {
			err = fmt.Errorf("fail to get unit by hash in dag: %v", unitHash.TerminalString())
			log.Debugf(err.Error())
			return
		}

		// 判断是否是换届前的单元
		if unit.Timestamp() <= dag.LastMaintenanceTime() {
			mSize = dag.PrecedingMediatorsCount()
			threshold = dag.PrecedingThreshold()
			dkgr, ok = mp.precedingDKGs[localMed]
		} else {
			mSize = dag.ActiveMediatorsCount()
			threshold = dag.ChainThreshold()
			dkgr, ok = mp.activeDKGs[localMed]
		}

		if !ok {
			log.Debugf("the mediator(%v)'s dkg is not existed", localMed.Str())
			return
		}
	}

	// 3. 判断是否达到群签名的各种条件
	if sigShareSet.len() < threshold {
		log.Debugf("the count of sign shares of the unit(%v) does not reach the threshold(%v)",
			unitHash.TerminalString(), threshold)
		return
	}

	dks, err := dkgr.DistKeyShare()
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	// 4. recover群签名
	suite := mp.suite
	pubPoly := share.NewPubPoly(suite, suite.Point().Base(), dks.Commitments())
	groupSig, err := tbls.Recover(suite, pubPoly, unitHash[:], sigShareSet.popSigShares(), threshold, mSize)
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	log.Debugf("Recovered the Unit(%v)'s the Group-sign: %v",
		unitHash.TerminalString(), hexutil.Encode(groupSig))

	// 5. recover后的相关处理
	// recover后 删除buf
	delete(mp.toTBLSRecoverBuf[localMed], unitHash)
	go mp.dag.SetUnitGroupSign(unitHash, groupSig, mp.ptn.TxPool())
	go mp.groupSigFeed.Send(GroupSigEvent{UnitHash: unitHash, GroupSig: groupSig})
}
