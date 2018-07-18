// This file is part of go-palletone
// go-palletone is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-palletone is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-palletone. If not, see <http://www.gnu.org/licenses/>.
//
// @author PalletOne DevTeam dev@pallet.one
// @date 2018

package ptn

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

func (self *ProtocolManager) ceBroadcastLoop() {
	for {
		select {
		case event := <-self.ceCh:
			self.BroadcastCe(event.Ce)

		// Err() channel will be closed when unsubscribing.
		case <-self.ceSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) BroadcastCe(ce string) {
	peers := pm.peers.GetPeers()
	for _, peer := range peers {
		peer.SendConsensus(ce)
	}
}

// create unit broadcast loop
func (self *ProtocolManager) unitedBroadcastLoop() {

	//	// automatically stops if unsubscribe
	//	for obj := range self.minedBlockSub.Chan() {
	//		switch ev := obj.Data.(type) {
	//		case core.NewMinedBlockEvent:
	//			self.BroadcastBlock(ev.Block, true)  // First propagate block to peers
	//			self.BroadcastBlock(ev.Block, false) // Only then announce to the rest
	//		}
	//	}
}

// BroadcastUnit will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastUnit(unit *modules.Unit, propagate bool) {
	hash := unit.UnitHash
	peers := pm.peers.PeersWithoutUnit(hash)

	//If this node is not mediator then Send the block to a subset of our peers
	//transfer := peers[:int(math.Sqrt(float64(len(peers))))]
	for _, peer := range peers {
		peer.SendNewUnit(unit)
	}
	//"duration", common.PrettyDuration(time.Since(block.ReceivedAt))
	log.Trace("Propagated unit", "hash", hash, "recipients", len(peers))
	return
}
