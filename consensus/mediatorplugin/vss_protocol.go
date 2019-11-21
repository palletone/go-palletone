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
	log.Debugf("initialize all mediator's dkgs, dealBufs, and responseBufs")
	// 初始化dkg，并初始化与完成vss相关的buf
	mp.dkgLock.Lock()
	log.Debugf("dkgLock.Lock()")
	defer log.Debugf("dkgLock.Unlock()")
	defer mp.dkgLock.Unlock()
	mp.vssBufLock.Lock()
	log.Debugf("vssBufLock.Lock()")
	defer log.Debugf("vssBufLock.Unlock()")
	defer mp.vssBufLock.Unlock()

	dag := mp.dag
	lams := mp.GetLocalActiveMediators()
	initPubs := dag.GetActiveMediatorInitPubs()
	curThreshold := dag.ChainThreshold()

	lamc := len(lams)
	mp.activeDKGs = make(map[common.Address]*dkg.DistKeyGenerator, lamc)

	ams := dag.GetActiveMediators()
	aSize := len(ams)

	for _, localMed := range lams {
		// 初始化本地所有mediator的dkg
		initSec := mp.mediators[localMed].InitPrivKey
		dkgr, err := dkg.NewDistKeyGenerator(mp.suite, initSec, initPubs, curThreshold)
		if err != nil {
			log.Debugf(err.Error())
			continue
		}
		mp.activeDKGs[localMed] = dkgr

		// 初始化所有与完成vss相关的buf
		mp.dealBuf[localMed] = make(chan *dkg.Deal, aSize-1)
		mp.respBuf[localMed] = make(map[common.Address]chan *dkg.Response, aSize)
		//mp.respBuf[localMed] = make(map[common.Address]chan *dkg.Response, aSize-1)
		for _, vrfrMed := range ams {
			//if vrfrMed == localMed {
			//	continue
			//}
			mp.respBuf[localMed][vrfrMed] = make(chan *dkg.Response, aSize)
			//mp.respBuf[localMed][vrfrMed] = make(chan *dkg.Response, aSize-1)
		}
	}
}

func (mp *MediatorPlugin) startVSSProtocol() {
	log.Debugf("start completing the VSS protocol")

	// 开启处理其他 mediator 的 deals 的循环，准备完成vss协议
	mp.launchVSSDealLoops()

	interval := mp.dag.GetGlobalProp().ChainParameters.MediatorInterval
	sleepTime := time.Second * time.Duration(interval)

	// 隔1个生产间隔，等待其他节点接收新unit，并做好vss协议相关准备工作
	select {
	case <-mp.quit:
		return
	case <-time.After(sleepTime):
		// 广播 vss deal 给其他节点，并处理来自其他节点的deal
		go mp.broadcastVSSDeals()
	}

	// 再隔1个生产间隔，才处理response，防止对应的 deal 还没收到的情况
	select {
	case <-mp.quit:
		return
	case <-time.After(sleepTime):
		go mp.launchVSSRespLoops()
	}

	// 再隔半个生产间隔，验证vss协议是否完成，并开始群签名
	select {
	case <-mp.quit:
		return
	case <-time.After(time.Second * time.Duration((interval+1)/2)):
		go mp.completeVSSProtocol()
	}
}

func (mp *MediatorPlugin) completeVSSProtocol() {
	log.Debugf("to complete vss protocol")
	// 停止所有vss相关的循环
	go func() {
		mp.stopVSS <- struct{}{}
	}()

	// 删除vss相关缓存
	mp.vssBufLock.Lock()
	log.Debugf("vssBufLock.Lock()")
	lamc := len(mp.mediators)
	mp.dealBuf = make(map[common.Address]chan *dkg.Deal, lamc)
	mp.respBuf = make(map[common.Address]map[common.Address]chan *dkg.Response, lamc)
	log.Debugf("vssBufLock.Unlock()")
	mp.vssBufLock.Unlock()

	// 验证vss是否完成，并开启群签名
	go mp.launchGroupSignLoops()
}

func (mp *MediatorPlugin) launchGroupSignLoops() {
	lams := mp.GetLocalActiveMediators()
	mp.dkgLock.Lock()
	log.Debugf("dkgLock.Lock()")
	defer log.Debugf("dkgLock.Unlock()")
	defer mp.dkgLock.Unlock()

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
		} else {
			log.Debugf("the mediator(%v)'s DKG verification failed", localMed.Str())
		}
	}
}

func (mp *MediatorPlugin) launchVSSDealLoops() {
	lams := mp.GetLocalActiveMediators()

	for _, localMed := range lams {
		go mp.processDealLoop(localMed)
	}
}

//func (mp *MediatorPlugin) launchDealAndRespLoops() {
func (mp *MediatorPlugin) launchVSSRespLoops() {
	lams := mp.GetLocalActiveMediators()
	ams := mp.dag.GetActiveMediators()

	for _, localMed := range lams {
		//go mp.processDealLoop(localMed)

		for _, vrfrMed := range ams {
			go mp.processResponseLoop(localMed, vrfrMed)
		}
	}
}

func (mp *MediatorPlugin) processDealLoop(localMed common.Address) {
	log.Debugf("the local active mediator(%v) run the loop to process deal", localMed.Str())

	mp.vssBufLock.RLock()
	dealCh, ok := mp.dealBuf[localMed]
	mp.vssBufLock.RUnlock()

	if !ok {
		log.Debugf("the mediator(%v)'s dealBuf has not initialized yet", localMed.Str())
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
			go mp.processVSSDeal(localMed, deal)
		}
	}
}

func (mp *MediatorPlugin) processVSSDeal(localMed common.Address, deal *dkg.Deal) {
	mp.dkgLock.Lock()
	log.Debugf("dkgLock.Lock()")
	defer log.Debugf("dkgLock.Unlock()")
	defer mp.dkgLock.Unlock()

	dkgr, ok := mp.activeDKGs[localMed]
	if !ok || dkgr == nil {
		log.Debugf("the mediator(%v)'s dkg is not existed, or it is not active", localMed.String())
		return
	}

	vrfrMed := mp.dag.GetActiveMediatorAddr(int(deal.Index))
	log.Debugf("the mediator(%v) process the vss deal from the mediator(%v)",
		localMed.Str(), vrfrMed.Str())

	resp, err := dkgr.ProcessDeal(deal)
	if err != nil {
		log.Debugf("dkg cannot process this deal: " + err.Error())
		return
	}

	if resp.Response.Status != vss.StatusApproval {
		err = fmt.Errorf("dag gave this deal a complaint: %v", localMed.String())
		log.Debugf(err.Error())
		return
	}

	cp := mp.dag.GetGlobalProp().ChainParameters
	deadline := time.Now().Unix() + int64(cp.MediatorInterval*cp.MaintenanceSkipSlots)

	respEvent := VSSResponseEvent{
		Resp:     resp,
		Deadline: uint64(deadline),
	}

	go mp.vssResponseFeed.Send(respEvent)
	log.Debugf("the mediator(%v) broadcast the vss response to the mediator(%v)",
		localMed.Str(), vrfrMed.Str())
}

func (mp *MediatorPlugin) broadcastVSSDeals() {
	mp.dkgLock.Lock()
	log.Debugf("dkgLock.Lock()")
	defer log.Debugf("dkgLock.Unlock()")
	defer mp.dkgLock.Unlock()

	cp := mp.dag.GetGlobalProp().ChainParameters
	deadline := time.Now().Unix() + int64(cp.MediatorInterval*cp.MaintenanceSkipSlots)

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
				Deadline: uint64(deadline),
			}
			go mp.vssDealFeed.Send(event)
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

	// 判断是否本地mediator的deal
	mp.vssBufLock.Lock()
	log.Debugf("vssBufLock.Lock()")
	dealCh, ok := mp.dealBuf[localMed]
	if ok {
		dealCh <- deal
		vrfrMed := dag.GetActiveMediatorAddr(int(deal.Index))
		log.Debugf("the mediator(%v) received the vss deal from the mediator(%v)",
			localMed.Str(), vrfrMed.Str())
	}
	log.Debugf("vssBufLock.Unlock()")
	mp.vssBufLock.Unlock()
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

		srcIndex := resp.Response.Index
		srcMed := dag.GetActiveMediatorAddr(int(srcIndex))
		if srcMed == localMed {
			//log.Debugf("ignore the vss response message from myself(mediator: %v)", srcMed.Str())
			continue
		}

		vrfrMed := dag.GetActiveMediatorAddr(int(resp.Index))

		mp.vssBufLock.Lock()
		log.Debugf("vssBufLock.Lock()")
		respCh, ok := mp.respBuf[localMed][vrfrMed]
		if ok {
			respCh <- resp
			log.Debugf("the mediator(%v) received the vss response from the mediator(%v) to the mediator(%v)",
				localMed.Str(), srcMed.Str(), vrfrMed.Str())
		}
		log.Debugf("vssBufLock.Unlock()")
		mp.vssBufLock.Unlock()
	}
}

func (mp *MediatorPlugin) processResponseLoop(localMed, vrfrMed common.Address) {
	log.Debugf("the mediator(%v) run the loop to process response regarding the mediator(%v)",
		localMed.Str(), vrfrMed.Str())

	mp.vssBufLock.RLock()
	respCh, ok := mp.respBuf[localMed][vrfrMed]
	mp.vssBufLock.RUnlock()

	if !ok {
		log.Debugf("the mediator(%v)'s respBuf corresponding the mediator(%v) has not initialized yet",
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
			go mp.processVSSResp(localMed, resp)
		}
	}
}

func (mp *MediatorPlugin) processVSSResp(localMed common.Address, resp *dkg.Response) {
	mp.dkgLock.Lock()
	log.Debugf("dkgLock.Lock()")
	defer log.Debugf("dkgLock.Unlock()")
	defer mp.dkgLock.Unlock()

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
