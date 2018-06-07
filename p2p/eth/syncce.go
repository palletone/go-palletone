package eth

import (
	//"math/rand"
	//"sync/atomic"
	//"time"

	"github.com/palletone/go-palletone/common/log"
	//"github.com/palletone/go-palletone/common"
	//"github.com/palletone/go-palletone/common/event"
	//"github.com/palletone/go-palletone/contracts/types"
	//"github.com/palletone/go-palletone/common/log"
	//"github.com/palletone/go-palletone/p2p/discover"
)

func (self *ProtocolManager) ceBroadcastLoop() {
	log.Info("=========ceBroadcastLoop")
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
	log.Info("=========BroadcastCe:", ce)
	//PeerCount
	//pm.server.PeerCount()
	counts := pm.peers.GetPeers()
	//counts := pm.peers.Len()
	log.Info("=========BroadcastCe have peers:", counts)
	var index int = 0
	peers := pm.peers.GetPeers()

	for _, peer := range peers {
		log.Info("======index:", index)
		peer.SendConsensus(ce)
		index++
	}

}
