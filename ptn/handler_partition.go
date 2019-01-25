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
 * @author PalletOne core developer <dev@pallet.one>
 * @date 2019
 */
package ptn

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
)

func (pm *ProtocolManager) PartitionHandle(p *peer) error {
	if pm.lightPeers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		log.Info("ProtocolManager", "handler DiscTooManyPeers:", p2p.DiscTooManyPeers)
		return p2p.DiscTooManyPeers
	}
	log.Debug("PalletOne peer connected", "name", p.Name())

	head := pm.dag.CurrentHeader()
	// Execute the PalletOne handshake
	if err := p.Handshake(pm.networkId, head.Number, pm.genesis.Hash() /*mediator,*/, head.Hash()); err != nil {
		log.Debug("PalletOne handshake failed", "err", err)
		return err
	}

	if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
		rw.Init(p.version)
	}

	// Register the peer locally
	if err := pm.lightPeers.Register(p); err != nil {
		log.Error("PalletOne peer registration failed", "err", err)
		return err
	}
	defer pm.removeLightPeer(p.id)

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			log.Debug("PalletOne message handling failed", "err", err)
			return err
		}
	}
}

func (pm *ProtocolManager) removeLightPeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing PalletOne peer", "peer", id)

	// Unregister the peer from the downloader and PalletOne peer set
	pm.downloader.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}
