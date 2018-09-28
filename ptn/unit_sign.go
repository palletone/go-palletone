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

package ptn

import (
	"fmt"

	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/modules"
)

// @author Albert·Gou
type producer interface {
	// SubscribeNewUnitEvent should return an event subscription of
	// NewUnitEvent and send events to the given channel.
	SubscribeNewUnitEvent(ch chan<- mp.NewUnitEvent) event.Subscription
	// UnitBLSSign is to TBLS sign the unit
	ToUnitTBLSSign(unit *modules.Unit) error

	SubscribeSigShareEvent(ch chan<- mp.SigShareEvent) event.Subscription
	ToTBLSRecover(sigShare *mp.SigShareEvent) error

	SubscribeVSSDealEvent(ch chan<- mp.VSSDealEvent) event.Subscription
	ToProcessDeal(deal *mp.VSSDealEvent) error

	SubscribeVSSResponseEvent(ch chan<- mp.VSSResponseEvent) event.Subscription
	ToProcessResponse(resp *mp.VSSResponseEvent) error

	LocalHaveActiveMediator() bool
	StartVSSProtocol()
}

// @author Albert·Gou
func (self *ProtocolManager) newUnitBroadcastLoop() {
	for {
		select {
		case event := <-self.newUnitCh:
			self.BroadcastNewUnit(event.Unit)

			// appended by wangjiyou
			self.BroadcastUnit(event.Unit, true)
			self.BroadcastUnit(event.Unit, false)

			// Err() channel will be closed when unsubscribing.
		case <-self.newUnitSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (self *ProtocolManager) sigShareTransmitLoop() {
	for {
		select {
		case event := <-self.sigShareCh:
			med := self.dag.GetUnit(event.Hash).UnitAuthor()
			node := self.dag.GetActiveMediator(*med).Node
			self.TransmitSigShare(node, &event)

			// Err() channel will be closed when unsubscribing.
		case <-self.sigShareSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) TransmitSigShare(node *discover.Node, sigShare *mp.SigShareEvent) {
	peer, self := pm.getTransitionPeer(node)
	if self {
		//size, reader, err := rlp.EncodeToReader(sigShare)
		//if err != nil {
		//	log.Error(err.Error())
		//}
		//
		//var s mp.SigShareEvent
		//stream := rlp.NewStream(reader, uint64(size))
		//if err := stream.Decode(&s); err != nil {
		//	log.Error(err.Error())
		//}
		//pm.producer.ToTBLSRecover(&s)

		pm.producer.ToTBLSRecover(sigShare)
		return
	}

	if peer == nil {
		return
	}

	err := peer.SendSigShare(sigShare)
	if err != nil {
		log.Error(err.Error())
	}
}

// @author Albert·Gou
func (self *ProtocolManager) vssDealTransmitLoop() {
	for {
		select {
		case event := <-self.vssDealCh:
			node := self.dag.GetActiveMediatorNode(event.DstIndex)
			self.TransmitVSSDeal(node, &event)

			// Err() channel will be closed when unsubscribing.
		case <-self.vssDealSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (self *ProtocolManager) vssResponseBroadcastLoop() {
	for {
		select {
		case event := <-self.vssResponseCh:
			self.BroadcastVssResp(&event)

			// Err() channel will be closed when unsubscribing.
		case <-self.vssResponseSub.Err():
			return
		}
	}
}

// @author Albert·Gou
//func (pm *ProtocolManager) BroadcastVssResp(dstId string, resp *mp.VSSResponseEvent) {
func (pm *ProtocolManager) BroadcastVssResp(resp *mp.VSSResponseEvent) {
	// comment by Albert·Gou
	//dstId := node.ID.TerminalString()
	//peer := pm.peers.Peer(dstId)
	//if peer == nil {
	//	log.Error(fmt.Sprintf("peer not exist: %v", node.String()))
	//}

	// comment by Albert·Gou
	//if pm.peers.PeersWithoutVssResp(dstId) {
	//	return
	//}
	//pm.peers.MarkVssResp(dstId)

	peers := pm.GetTransitionPeers() //pm.GetActiveMediatorPeers()
	for _, peer := range peers {
		if peer == nil {
			//size, reader, err := rlp.EncodeToReader(resp)
			//if err != nil {
			//	log.Error(err.Error())
			//}
			//
			//var r mp.VSSResponseEvent
			//s := rlp.NewStream(reader, uint64(size))
			//if err := s.Decode(&r); err != nil {
			//	log.Error(err.Error())
			//}
			//pm.producer.ToProcessResponse(&r)

			pm.producer.ToProcessResponse(resp)
			continue
		}

		// comment by Albert·Gou
		//dstId := peer.id
		//if pm.peers.PeersWithoutVssResp(dstId) {
		//	return
		//}
		//pm.peers.MarkVssResp(dstId)

		// comment by Albert·Gou
		//msg := &vssRespMsg{
		//	NodeId: dstId,
		//	Resp:   resp,
		//}
		//
		//err := peer.SendVSSResponse(msg)

		err := peer.SendVSSResponse(resp)
		if err != nil {
			log.Info(err.Error())
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) TransmitVSSDeal(node *discover.Node, deal *mp.VSSDealEvent) {
	peer, self := pm.getTransitionPeer(node)
	if self {
		//size, reader, err := rlp.EncodeToReader(deal)
		//if err != nil {
		//	log.Error(err.Error())
		//}
		//
		//var d mp.VSSDealEvent
		//s := rlp.NewStream(reader, uint64(size))
		//if err := s.Decode(&d); err != nil {
		//	log.Error(err.Error())
		//}
		//pm.producer.ToProcessDeal(&d)

		pm.producer.ToProcessDeal(deal)
		return
	}

	if peer == nil {
		return
	}

	// comment by Albert·Gou
	// // append by wangjiyou
	//if pm.peers.PeersWithoutVss(dstId) {
	//	return
	//}
	//pm.peers.MarkVss(dstId)

	//msg := &vssMsg{
	//	NodeId: dstId,
	//	Deal:   deal,
	//}
	//err := peer.SendVSSDeal(msg)

	err := peer.SendVSSDeal(deal)
	if err != nil {
		log.Error(err.Error())
	}
}

// @author Albert·Gou
// BroadcastNewUnit will propagate a new produced unit to all of active mediator's peers
func (pm *ProtocolManager) BroadcastNewUnit(unit *modules.Unit) {
	peers := pm.GetActiveMediatorPeers()
	for _, peer := range peers {
		if peer == nil {
			pm.producer.ToUnitTBLSSign(unit)
			continue
		}

		err := peer.SendNewProducedUnit(unit)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// getActiveMediatorPeer, retrieve specified active mediator peer.
// If it is the node itself, p is nil and self is true
// @author Albert·Gou
func (pm *ProtocolManager) getActiveMediatorPeer(node *discover.Node) (p *peer, self bool) {
	id := node.ID
	if pm.srvr.Self().ID == id {
		self = true
	}

	p = pm.peers.Peer(id.TerminalString())
	if p == nil && !self {
		log.Debug(fmt.Sprintf("Active Mediator Peer not exist: %v", node.String()))
	}

	return
}

// AtiveMeatorPeers retrieves a list of peers that active mediator.
// If the value is nil, it is the node itself
// @author Albert·Gou
func (pm *ProtocolManager) GetActiveMediatorPeers() map[string]*peer {
	nodes := pm.dag.GetActiveMediatorNodes()
	list := make(map[string]*peer, len(nodes))

	for id, node := range nodes {
		peer, self := pm.getActiveMediatorPeer(node)
		if peer != nil || self {
			list[id] = peer
		}
	}

	return list
}
