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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"go.dedis.ch/kyber/v3/share/dkg/pedersen"
	"go.dedis.ch/kyber/v3/share/vss/pedersen"
)

func (mp *MediatorPlugin) newDKGAndInitVSSBuf() {
	mp.dkgLock.RLock()
	defer mp.dkgLock.RUnlock()
	mp.vssBufLock.RLock()
	defer mp.vssBufLock.RUnlock()

	dag := mp.dag
	lams := mp.GetLocalActiveMediators()
	initPubs := dag.GetActiveMediatorInitPubs()
	curThreshold := dag.ChainThreshold()

	lamc := len(lams)
	mp.activeDKGs = make(map[common.Address]*dkg.DistKeyGenerator, lamc)

	ams := dag.GetActiveMediators()
	aSize := len(ams)

	for _, localMed := range lams {
		initSec := mp.mediators[localMed].InitPrivKey
		dkgr, err := dkg.NewDistKeyGenerator(mp.suite, initSec, initPubs, curThreshold)
		if err != nil {
			log.Debugf(err.Error())
			continue
		}
		mp.activeDKGs[localMed] = dkgr

		mp.dealBuf[localMed] = make(chan *dkg.Deal, aSize-1)
		mp.respBuf[localMed] = make(map[common.Address]chan *dkg.Response, aSize)
		for _, vrfrMed := range ams {
			mp.respBuf[localMed][vrfrMed] = make(chan *dkg.Response, aSize-1)
		}
	}
}

func (mp *MediatorPlugin) startVSSProtocol() {
	log.Debugf("start completing the VSS protocol")

	// 开启处理其他 mediator 的 deals 的循环，准备完成vss协议
	mp.launchVSSDealLoops()

	interval := mp.dag.GetGlobalProp().ChainParameters.MediatorInterval / 2

	// 隔1个生产间隔，等待其他节点接收新unit，并处理好vss协议相关准备工作
	select {
	case <-mp.quit:
		return
	case <-time.After(time.Second * time.Duration(interval)):
		mp.broadcastVSSDeals()
	}

	// 再隔1个生产间隔，处理response
	select {
	case <-mp.quit:
		return
	case <-time.After(time.Second * time.Duration(interval)):
		mp.launchVSSRespLoops()
	}

	// 再隔1个生产间隔，验证vss协议是否完成，并开始群签名
	select {
	case <-mp.quit:
		return
	case <-time.After(time.Second * time.Duration(interval)):
		mp.completeVSSProtocol()
	}
}

func (mp *MediatorPlugin) completeVSSProtocol() {
	// 停止所有vss相关的循环
	mp.stopVSS <- struct{}{}

	// 删除vss相关缓存
	mp.vssBufLock.RLock()
	lamc := len(mp.mediators)
	mp.dealBuf = make(map[common.Address]chan *dkg.Deal, lamc)
	mp.respBuf = make(map[common.Address]map[common.Address]chan *dkg.Response, lamc)
	mp.vssBufLock.RUnlock()

	// 验证vss是否完成，并开启群签名
	mp.launchGroupSignLoops()
}

func (mp *MediatorPlugin) launchGroupSignLoops() {
	lams := mp.GetLocalActiveMediators()
	mp.dkgLock.RLock()
	defer mp.dkgLock.RUnlock()

	for _, localMed := range lams {
		dkgr, ok := mp.activeDKGs[localMed]
		if !ok || dkgr == nil {
			log.Debugf("the mediator(%v)'s dkg is not existed, or it is not active", localMed.String())
			continue
		}

		if dkgr.Certified() {
			log.Debugf("the mediator(%v)'s DKG verification passed", localMed.Str())

			go mp.signUnitsTBLS(localMed)
			go mp.recoverUnitsTBLS(localMed)
		}
	}
}

func (mp *MediatorPlugin) launchVSSDealLoops() {
	lams := mp.GetLocalActiveMediators()
	//ams := mp.dag.GetActiveMediators()

	for _, localMed := range lams {
		go mp.processDealLoop(localMed)

		//for _, vrfrMed := range ams {
		//	go mp.processResponseLoop(localMed, vrfrMed)
		//}
	}
}

func (mp *MediatorPlugin) launchVSSRespLoops() {
	lams := mp.GetLocalActiveMediators()
	ams := mp.dag.GetActiveMediators()

	for _, localMed := range lams {
		for _, vrfrMed := range ams {
			go mp.processResponseLoop(localMed, vrfrMed)
		}
	}
}

func (mp *MediatorPlugin) processDealLoop(localMed common.Address) {
	log.Debugf("the local active mediator(%v) run the loop to deal", localMed.Str())

	mp.vssBufLock.Lock()
	dealCh, ok := mp.dealBuf[localMed]
	mp.vssBufLock.Unlock()

	if !ok {
		log.Debugf("the mediator(%v)'s dealBuf is not initialized", localMed.Str())
		return
	}

	for {
		select {
		case <-mp.quit:
			return
		case <-mp.stopVSS:
			return
		case deal := <-dealCh:
			// 处理deal，并广播response
			mp.processVSSDeal(localMed, deal)
		}
	}
}

func (mp *MediatorPlugin) processVSSDeal(localMed common.Address, deal *dkg.Deal) {
	mp.dkgLock.RLock()
	defer mp.dkgLock.RUnlock()

	dkgr, ok := mp.activeDKGs[localMed]
	if !ok || dkgr == nil {
		log.Debugf("the mediator(%v)'s dkg is not existed, or it is not active", localMed.String())
		return
	}

	vrfrMed := mp.dag.GetActiveMediatorAddr(int(deal.Index))
	log.Debugf("the mediator(%v) received the vss deal from the mediator(%v)",
		localMed.Str(), vrfrMed.Str())

	resp, err := dkgr.ProcessDeal(deal)
	if err != nil {
		log.Debugf("dkg: cannot process own deal: " + err.Error())
		return
	}

	if resp.Response.Status != vss.StatusApproval {
		err = fmt.Errorf("DKG: own deal gave a complaint: %v", localMed.String())
		log.Debugf(err.Error())
		return
	}

	respEvent := VSSResponseEvent{
		Resp: resp,
	}
	mp.vssResponseFeed.Send(respEvent)
	log.Debugf("the mediator(%v) broadcast the vss response to the mediator(%v)",
		localMed.Str(), vrfrMed.Str())
}

func (mp *MediatorPlugin) broadcastVSSDeals() {
	mp.dkgLock.RLock()
	defer mp.dkgLock.RUnlock()

	// 将deal广播给其他节点
	for localMed, dkg := range mp.activeDKGs {
		deals, err := dkg.Deals()
		if err != nil {
			log.Debugf(err.Error())
			continue
		}
		log.Debugf("the mediator(%v) broadcast vss deals", localMed.Str())

		for index, deal := range deals {
			event := VSSDealEvent{
				DstIndex: uint32(index),
				Deal:     deal,
			}
			mp.vssDealFeed.Send(event)
		}
	}
}

func (mp *MediatorPlugin) SubscribeVSSDealEvent(ch chan<- VSSDealEvent) event.Subscription {
	return mp.vssDealScope.Track(mp.vssDealFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) AddToDealBuf(dealEvent *VSSDealEvent) {
	if !mp.groupSigningEnabled {
		return
	}

	dag := mp.dag
	localMed := dag.GetActiveMediatorAddr(int(dealEvent.DstIndex))

	deal := dealEvent.Deal
	vrfrMed := dag.GetActiveMediatorAddr(int(deal.Index))
	log.Debugf("the mediator(%v) received the vss deal from the mediator(%v)",
		localMed.Str(), vrfrMed.Str())

	mp.vssBufLock.RLock()
	dealCh, ok := mp.dealBuf[localMed]
	if !ok {
		log.Debugf("the mediator(%v)'s dealBuf is not initialized", localMed.Str())
	} else {
		dealCh <- deal
	}
	mp.vssBufLock.RUnlock()
}

func (mp *MediatorPlugin) SubscribeVSSResponseEvent(ch chan<- VSSResponseEvent) event.Subscription {
	return mp.vssResponseScope.Track(mp.vssResponseFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) AddToResponseBuf(respEvent *VSSResponseEvent) {
	if !mp.groupSigningEnabled {
		return
	}

	resp := respEvent.Resp
	lams := mp.GetLocalActiveMediators()
	for _, localMed := range lams {
		dag := mp.dag

		// ignore the message from myself
		srcIndex := resp.Response.Index
		srcMed := dag.GetActiveMediatorAddr(int(srcIndex))
		if srcMed == localMed {
			continue
		}

		vrfrMed := dag.GetActiveMediatorAddr(int(resp.Index))
		log.Debugf("the mediator(%v) received the vss response from the mediator(%v) to the mediator(%v)",
			localMed.Str(), srcMed.Str(), vrfrMed.Str())

		mp.vssBufLock.RLock()
		respCh, ok := mp.respBuf[localMed][vrfrMed]
		if !ok {
			log.Debugf("the mediator(%v)'s respBuf corresponding the mediator(%v) is not initialized",
				localMed.Str(), vrfrMed.Str())
		} else {
			respCh <- resp
		}
		mp.vssBufLock.RUnlock()
	}
}

func (mp *MediatorPlugin) processResponseLoop(localMed, vrfrMed common.Address) {
	log.Debugf("the mediator(%v) run the loop to process response regarding the mediator(%v)",
		localMed.Str(), vrfrMed.Str())

	mp.vssBufLock.Lock()
	respCh, ok := mp.respBuf[localMed][vrfrMed]
	mp.vssBufLock.Unlock()

	if !ok {
		log.Debugf("the mediator(%v)'s respBuf corresponding the mediator(%v) is not initialized",
			localMed.Str(), vrfrMed.Str())
		return
	}

	for {
		select {
		case <-mp.quit:
			return
		case <-mp.stopVSS:
			return
		case resp := <-respCh:
			mp.processVSSResp(localMed, resp)
		}
	}
}

func (mp *MediatorPlugin) processVSSResp(localMed common.Address, resp *dkg.Response) {
	mp.dkgLock.RLock()
	defer mp.dkgLock.RUnlock()

	dkgr, ok := mp.activeDKGs[localMed]
	if !ok || dkgr == nil {
		log.Debugf("the mediator(%v)'s dkg is not existed, or it is not active", localMed.String())
		return
	}

	jstf, err := dkgr.ProcessResponse(resp)
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	if jstf != nil {
		log.Debugf("DKG: wrong Process Response: %v", localMed.String())
		return
	}
}
