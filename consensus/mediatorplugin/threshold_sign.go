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

func (mp *MediatorPlugin) BroadcastVSSDeals() {
	for _, dkg := range mp.dkgs {
		deals, err := dkg.Deals()
		if err != nil {
			log.Error(err.Error())
		}

		for index, deal := range deals {
			event := VSSDealEvent{
				DstIndex: index,
				Deal:     deal,
			}

			mp.vssDealFeed.Send(event)
		}
	}
}

func (mp *MediatorPlugin) SubscribeVSSDealEvent(ch chan<- VSSDealEvent) event.Subscription {
	return mp.vssDealScope.Track(mp.vssDealFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) ToProcessDeal(deal *VSSDealEvent) error {
	select {
	case <-mp.quit:
		return errTerminated
	case mp.toProcessDealCh <- deal:
		return nil
	}
}

func (mp *MediatorPlugin) processDealLoop() {
	for {
		select {
		case <-mp.quit:
			return
		case deal := <-mp.toProcessDealCh:
			go mp.processVSSDeal(deal)
		}
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
	dstMed := mp.getDag().GetActiveMediatorAddr(dealEvent.DstIndex)

	dkgr := mp.getLocalActiveMediatorDKG(dstMed)
	if dkgr == nil {
		return
	}

	deal := dealEvent.Deal
	mp.vrfrReady[dstMed][deal.Index] = true
	go mp.notifyProcessResp(&dkgVerifier{dstMed, deal.Index})

	resp, err := dkgr.ProcessDeal(deal)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if resp.Response.Status != vss.StatusApproval {
		log.Error(fmt.Sprintf("DKG: own deal gave a complaint: %v", dstMed.String()))
		return
	}

	go mp.BroadcastVSSResponse(dstMed, resp)
}

func (mp *MediatorPlugin) notifyProcessResp(dvp *dkgVerifier) {
	mp.vrfrReadyCh <- dvp
}

// BroadcastVSSResponse, broadcast response to every other participant
func (mp *MediatorPlugin) BroadcastVSSResponse(srcMed common.Address, resp *dkg.Response) {
	event := VSSResponseEvent{
		Resp: resp,
	}

	mp.vssResponseFeed.Send(event)
}

func (mp *MediatorPlugin) ToProcessResponse(resp *VSSResponseEvent) error {
	select {
	case <-mp.quit:
		return errTerminated
	case mp.toProcessResponseCh <- resp:
		return nil
	}
}

func (mp *MediatorPlugin) processResponseLoop() {
	for {
		select {
		case <-mp.quit:
			return
		case resp := <-mp.toProcessResponseCh:
			go mp.processVSSResponse(resp)
		case dvp := <-mp.vrfrReadyCh:
			go mp.processResponseBuf(dvp)
		}
	}
}

func (mp *MediatorPlugin) initRespBuf(dstMed common.Address) {
	aSize := mp.getDag().GetActiveMediatorCount()
	for i := 0; i < aSize; i++ {
		mp.respBuf[dstMed][uint32(i)] = make(chan *dkg.Response, aSize-1)
	}
}

func (mp *MediatorPlugin) processResponseBuf(dvp *dkgVerifier) {
	dstMed := dvp.medLocal
	srcIndex := dvp.srcIndex
	dkg := mp.getLocalActiveMediatorDKG(dstMed)
	if dkg == nil {
		return
	}

	if !mp.vrfrReady[dstMed][srcIndex] {
		return
	}

	for resp := range mp.respBuf[dstMed][srcIndex] {

		jstf, err := dkg.ProcessResponse(resp)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if jstf != nil {
			log.Error(fmt.Sprintf("DKG: wrong Process Response: %v", dstMed.String()))
			continue
		}
	}
}

func (mp *MediatorPlugin) processVSSResponse(resp *VSSResponseEvent) {
	lams := mp.GetLocalActiveMediators()
	for _, dstMed := range lams {
		//ignore the message from myself
		srcIndex := resp.Resp.Response.Index
		srcMed := mp.getDag().GetActiveMediatorAddr(int(srcIndex))
		if srcMed == dstMed {
			continue
		}

		dkg := mp.getLocalActiveMediatorDKG(dstMed)
		if dkg == nil {
			continue
		}

		mp.respBuf[dstMed][srcIndex] <- resp.Resp
		go mp.notifyProcessResp(&dkgVerifier{dstMed, srcIndex})
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
	case mp.toBLSSigned <- op:
		return nil
	}
}

func (mp *MediatorPlugin) unitBLSSignLoop() {
	for {
		select {
		// Mediator Plugin terminating, abort operation
		case <-mp.quit:
			return
		case op := <-mp.toBLSSigned:
			//			PushUnit(mp.ptn.Dag(), op.unit)
			go mp.unitBLSSign(op)
		}
	}
}

func (mp *MediatorPlugin) unitBLSSign(toBLSSigned *toBLSSigned) {
	//todo

}
