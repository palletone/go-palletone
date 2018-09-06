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
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

func GenInitPair(suite vss.Suite) (kyber.Scalar, kyber.Point) {
	sc := suite.Scalar().Pick(suite.RandomStream())

	return sc, suite.Point().Mul(sc, nil)
}

func (mp *MediatorPlugin) BroadcastVSSDeals() {
	lams := mp.GetLocalActiveMediators()
	nParticipants := mp.getDag().GetActiveMediatorCount()
	ams := mp.getDag().GetActiveMediators()

	for _, medF := range lams {
		resps := make([]*dkg.Response, 0, nParticipants)
		mp.resps[medF] = resps

		dkg, ok := mp.dkgs[medF]
		if !ok || dkg == nil {
			log.Error(fmt.Sprintf("The following mediator`s dkg was not successfully generated: %v",
				medF.String()))
			continue
		}

		deals, err := dkg.Deals()
		if err != nil {
			log.Error(err.Error())
		}

		for index, deal := range deals {
			event := VSSDealEvent{
				DstMed: ams[index],
				Deal:   deal,
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

func (mp *MediatorPlugin) processVSSDeal(deal *VSSDealEvent) {
	dstMed := deal.DstMed

	dkg, ok := mp.dkgs[dstMed]
	if !ok || dkg == nil {
		log.Error(fmt.Sprintf("The following mediator`s dkg is not existed: %v", dstMed.String()))
		return
	}

	resp, err := dkg.ProcessDeal(deal.Deal)
	if err != nil {
		log.Error(err.Error())
	}

	if resp.Response.Status != vss.StatusApproval {
		log.Error(fmt.Sprintf("DKG: own deal gave a complaint: %v", dstMed.String()))
	}

	// todo broadcasts the resulting response
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
