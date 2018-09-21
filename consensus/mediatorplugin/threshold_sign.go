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

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/share/dkg/pedersen"
	"github.com/dedis/kyber/share/vss/pedersen"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

func GenInitPair(suite vss.Suite) (kyber.Scalar, kyber.Point) {
	sc := suite.Scalar().Pick(suite.RandomStream())

	return sc, suite.Point().Mul(sc, nil)
}

func (mp *MediatorPlugin) StartVSSProtocol() {
	// todo 广播vss deal， 开启定时器， 并开启循环接收本地mediator完成vss协议的消息
	log.Info("Start completing the VSS protocol.")

	go mp.BroadcastVSSDeals()

	//timeout := time.NewTimer(3 * time.Second)
	//defer timeout.Stop()
	//select {
	//case <-mp.quit:
	//	return
	//case <-timeout.C:
	//	for med, dkg := range mp.dkgs {
	//		log.Debug(fmt.Sprintf("%v 's DKG certifing is %v", med.Str(), dkg.Certified()))
	//	}
	//}
}

func (mp *MediatorPlugin) BroadcastVSSDeals() {
	for localMed, dkg := range mp.dkgs {
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

func (mp *MediatorPlugin) getLocalActiveMediatorDKG(add common.Address) *dkg.DistKeyGenerator {
	if !mp.IsLocalActiveMediator(add) {
		log.Error(fmt.Sprintf("The following mediator is not local active mediator: %v", add.String()))
		return nil
	}

	dkg, ok := mp.dkgs[add]
	if !ok || dkg == nil {
		log.Error(fmt.Sprintf("The following mediator`s dkg is not existed: %v", add.String()))
		return nil
	}

	return dkg
}

func (mp *MediatorPlugin) processVSSDeal(dealEvent *VSSDealEvent) {
	dag := mp.getDag()
	localMed := dag.GetActiveMediatorAddr(dealEvent.DstIndex)

	dkgr := mp.getLocalActiveMediatorDKG(localMed)
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

func (mp *MediatorPlugin) processResponseLoop(localMed, vrfrMed common.Address) {
	dkgr := mp.getLocalActiveMediatorDKG(localMed)
	if dkgr == nil {
		return
	}

	aSize := mp.getDag().GetActiveMediatorCount()
	respCount := 0
	// localMed 对 vrfrMed 的 response 在 ProcessDeal 生成 response 时 自动处理了
	if vrfrMed != localMed {
		respCount++
	}

	respCh := mp.respBuf[localMed][vrfrMed]

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
				log.Info(fmt.Sprintf("%v's DKG verification passed!", localMed.Str()))
				certified = true
			}
		}

		return
	}

	for {
		select {
		case <-mp.quit:
			return
		case resp := <-respCh:
			processResp(resp)
			finished, certified := isFinishedAndCertified()
			if certified {
				// todo 发送验证成功消息, 开启unit分片签名

			}
			if finished {
				return
			}
		}
	}
}

func (mp *MediatorPlugin) addToResponseBuf(respEvent *VSSResponseEvent) {
	resp := respEvent.Resp
	lams := mp.GetLocalActiveMediators()
	for _, localMed := range lams {
		dag := mp.getDag()

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

func (mp *MediatorPlugin) ToUnitTBLSSign(peer string, unit *modules.Unit) error {
	op := &toBLSSigned{
		origin: peer,
		unit:   unit,
	}

	select {
	case <-mp.quit:
		return errTerminated
	default:
		go mp.unitBLSSign(op)
		return nil
	}
}

func (mp *MediatorPlugin) unitBLSSign(toBLSSigned *toBLSSigned) {
	//todo

}
