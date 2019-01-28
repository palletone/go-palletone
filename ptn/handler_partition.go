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
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/lps"
)

func (pm *ProtocolManager) newLightFetcher() *lps.LightFetcher {
	headerVerifierFn := func(header *modules.Header) error {
		return nil
	}
	headerBroadcaster := func(header *modules.Header, propagate bool) {

	}
	peerDrop := func(id string) {

	}
	return lps.New(pm.dag.GetLightHeaderByHash, pm.dag.GetLightChainHeight, headerVerifierFn,
		headerBroadcaster, pm.dag.InsertLightHeader, peerDrop)
}

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
		if err := pm.partionHandleMsg(p); err != nil {
			log.Debug("PalletOne message handling failed", "err", err)
			return err
		}
	}
}

func (pm *ProtocolManager) removeLightPeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.lightPeers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing PalletOne peer", "peer", id)

	if err := pm.lightPeers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) partionHandleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}

	defer msg.Discard()

	// Handle the message depending on its contents
	switch {
	case msg.Code == NewBlockHeaderMsg:
		return pm.NewBlockHeaderMsg(msg, p)
	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (pm *ProtocolManager) NewBlockHeaderMsg(msg p2p.Msg, p *peer) error {
	var header *modules.Header
	if err := msg.Decode(header); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	log.Debug("ProtocolManager NewBlockHeaderMsg", "p.id", p.id, "header:", *header)
	defer log.Debug("ProtocolManager NewBlockHeaderMsg", "p.id", p.id, "header:", *header)
	//pm.lightFetcher.Enqueue(p.id, header)
	//TODO if local peer index < request peer index,should sync with the same protocal peers
	return nil
}

//subprotocal equal ptn or not equal ptn
func (pm *ProtocolManager) BroadcastLightHeader(header *modules.Header, subProtocolName string) {
	log.Info("ProtocolManager", "BroadcastLightHeader index:", header.Index(), "protocal name:", subProtocolName)
	if subProtocolName == ProtocolName && header.Number.AssetID != modules.PTNCOIN {
		//TODO broadcast other token header in self(ptn) peers
		log.Info("===broadcast other token header in self(ptn) peers===")
		return
	}
	if subProtocolName != ProtocolName {
		//broacast self token(not ptn token) to ptn peers
		hash := header.Hash()
		peers := pm.lightPeers.GetPeers()
		for _, peer := range peers {
			peer.SendLightHeader(header)
		}
		log.Trace("BroadcastLightHeader Propagated header", "protocalname", pm.SubProtocols[0].Name, "index:", header.Number.Index, "hash", hash, "recipients", len(peers))
	}
}
