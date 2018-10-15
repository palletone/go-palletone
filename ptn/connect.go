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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package ptn

import (
	"errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
)

func (pm *ProtocolManager) mediatorConnect() {
	if pm.isTest {
		return
	}
	if !pm.producer.LocalHaveActiveMediator() {
		log.Info("This node is not Mediator")
		return
	}
	log.Info("Mediator Connect")

	peers := pm.dag.GetActiveMediatorNodes()

	//not exsit and no self will connect
	for _, peer := range peers {
		log.Debug("ProtocolManager", "GetActiveMediatorNodes:", peer.ID.String(), "local peerId:", pm.srvr.NodeInfo().ID)
		if peer.ID.String() != pm.srvr.NodeInfo().ID && pm.peers.Peer(peer.ID.String()) == nil {
			pm.srvr.AddPeer(peer)
		}
	}
}

//1.is not mediator,so save only two mediator connects,and disconnect others connects.
//2.also mediator,move peersTransition sockets to peers and delete the old mediator
func (pm *ProtocolManager) TransitionConvert() {
	if !pm.producer.LocalHaveActiveMediator() {
		log.Info("This node is not Mediator")
		peers := pm.peers.GetPeers()
		for _, peer := range peers {
			if !peer.mediator {
				continue
			}
			url := "pnode://" + peer.Peer.ID().String() + "@" + peer.Peer.RemoteAddr().String()
			log.Debug("TransitionConvert", "url:", url)
			node, err := discover.ParseNode(url)
			if err != nil {
				log.Error("TransitionConvert", "invalid pnode: %v", err)
				continue
			}
			pm.srvr.RemovePeer(node)
		}
		return
	}
	newPeers := pm.dag.GetActiveMediatorNodes()
	oldPeers := pm.peers.GetPeers()

	for _, oldPeer := range oldPeers {
		if _, ok := newPeers[oldPeer.ID().String()]; ok {
			continue
		}
		url := "pnode://" + oldPeer.Peer.ID().String() + "@" + oldPeer.Peer.RemoteAddr().String()
		node, err := discover.ParseNode(url)
		if err != nil {
			log.Error("TransitionConvert", "invalid pnode: %v", err)
			continue
		}
		pm.srvr.RemovePeer(node)
	}

	oldPeers = pm.peers.GetPeers()
	for _, newPeer := range newPeers {
		if pm.isexist(newPeer.ID.String(), oldPeers) {
			continue
		}
		pm.srvr.AddPeer(newPeer)
	}

}

func (pm *ProtocolManager) isexist(pid string, peers []*peer) bool {
	for _, peer := range peers {
		if pid == peer.id {
			return true
		}
	}
	return false
}

func (pm *ProtocolManager) peerCheck(p *peer) error {
	//TODO must delete
	return nil
	if err := pm.mediatorCheck(p); err != nil {
		log.Debug("mediatorCheck")
		return err
	}
	if err := pm.noMediatorCheck(p); err != nil {
		log.Debug("noMediatorCheck")
		return err
	}

	return nil
}

func (pm *ProtocolManager) mediatorCheck(p *peer) error {
	log.Info("ProtocolManager mediatorCheck")
	if pm.isTest {
		return nil
	}
	if p.mediator {
		peers := pm.dag.GetActiveMediatorNodes()
		if _, ok := peers[p.ID().TerminalString()]; ok {
			//TODO check the number of mediator connctions and the number of nomediator connections
			//if pm.peers.mediatorCheck(p, pm.maxPeers, len(peers)) {
			//}
		} else {
			log.Info("PalletOne handshake failed lying selef is mediator")
			return errors.New("PalletOne handshake failed lying selef is mediator")
		}
	}
	return nil
}

func (pm *ProtocolManager) noMediatorCheck(p *peer) error {
	log.Info("ProtocolManager noMediatorCheck")
	if pm.isTest {
		return nil
	}
	if !p.mediator {
		peers := pm.dag.GetActiveMediatorNodes()
		if _, ok := peers[p.ID().TerminalString()]; !ok {
			//TODO check the number of mediator connctions and the number of nomediator connections
			if !pm.peers.noMediatorCheck(pm.maxPeers, len(peers)-1) {
				log.Info("The number of no ediator connections full")
				return errors.New("The number of no ediator connections full")
			}
		} else {
			log.Info("PalletOne handshake failed lying self is not mediator")
			return errors.New("PalletOne handshake failed lying self is not mediator")
		}
	}
	return nil
}

/*
	log.Info("handle", "p.Peer.ID()", p.Peer.ID())
	log.Info("handle", "p.Peer.LocalAddr().String()", p.Peer.LocalAddr().String())
	log.Info("handle", "p.Peer.RemoteAddr().String()", p.Peer.RemoteAddr().String())
	log.Info("handle", "p.Peer.String()", p.Peer.String())
	url := "pnode://" + p.Peer.ID().String() + "@" + p.Peer.RemoteAddr().String()
	log.Debug("handle", "url:", url)

//important:if not new mediator Deprecated
//1.new mediators each other connected.Add new connections to peersTransition.
// if new mediator existing in peers,copy peers to peersTransition
//2.modify send vss request whith peersTransition sockets
//3.
//4.pm.cancelOldMediatorConnect()
func (pm *ProtocolManager) TransitionConnect() {
	//TODO Are new mediator.No will return TransitionConvert.
	//if !pm.producer.LocalHaveActiveMediator() {
	//	log.Info("This node is not new Mediator")
	//	return
	//}

	log.Info("Mediator transition")
	//TODO modify get new mediator
	//oldPeers := pm.dag.GetActiveMediatorNodes()
	newPeers := pm.dag.GetActiveMediatorNodes()
	//pm.peersTransition
	for _, peer := range newPeers {
		if peer.ID.String() != pm.srvr.NodeInfo().ID && pm.peers.Peer(peer.ID.String()) == nil {
			pm.srvr.AddPeer(peer)
		} else {

		}
	}
}

//1.add flag.This node whether or not mediator
//2.check connections.
//	2.1 mediator node: The no mediator connections is maxpeers sub mediators
//	2.2 no mediator node:unlimited
//3.all the mediators node is connectin.Notice the mediator plugin
func (pm *ProtocolManager) transitionConnect() {
	if !pm.producer.LocalHaveActiveMediator() {
		log.Info("This node is not Mediator")
		return
	}
	log.Info("Mediator transition")

	pm.peersTransition.MediatorsClean()

	//add interval
	forceSync := time.NewTicker(forceSyncCycle)
	defer forceSync.Stop()
	for {
		select {
		case <-forceSync.C:
			if err := pm.startTransitionConnect(); err != nil {
				return
			}
		case <-pm.transCycleConnCh:
			pm.peersTransition.MediatorsClean()
			return
		default:
		}
	}
}

//Start MediatorNetwork
func (pm *ProtocolManager) startTransitionConnect() error {
	//TODO must modify the GetTransitionNodes
	peers := pm.dag.GetActiveMediatorNodes()
	if pm.maxPeers < len(peers)+3 {
		log.Error("PalletOne start", "maxpeers", pm.maxPeers, "mediator size", len(peers)+3) //3:nomediator
		return errors.New("maxpeers < mediator size")
	}

	if pm.peersTransition.mediators.Size() != len(peers) {
		nodes := []string{}
		for _, peer := range peers {
			nodeId := peer.ID.TerminalString()
			nodes = append(nodes, nodeId)
			pm.peersTransition.MediatorsReset(nodes)
		}
	}

	//not exsit and no self will connect
	for _, peer := range peers {
		if peer.ID.String() != pm.srvr.NodeInfo().ID && pm.peersTransition.Peer(peer.ID.String()) == nil {
			log.Debug("========transition AddPeer==========", "peer.ID.String():", peer.ID.String())
			pm.srvr.AddPeer(peer)
		}
	}

	log.Debug("PalletOne", "startMediatorNetwork mediators:", len(peers))
	return nil
}

//TODO notice handle to return and remove peer.21 channel
func (pm *ProtocolManager) cancelTransitionConnect() {
	peers := pm.peersTransition.GetPeers()
	for _, peer := range peers {
		peer.transitionCh <- transitionCancel
	}
}

func (pm *ProtocolManager) transitionRun(p *peer) error {
	if p.mediator && pm.producer.LocalHaveActiveMediator() {
		if pm.peersTransition.mediators.Has(p.ID().TerminalString()) {
			if err := pm.handleTransitionMsg(p); err != nil {
				return err
			}
		}
	}
	return nil
}
*/
