// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ptn

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
)

const (
	forceSyncCycle      = 5 * time.Second // Time interval to force syncs, even if few peers are available
	minDesiredPeerCount = 5               //5                // Amount of peers desired to start syncing

	// This is the target size for the packs of transactions sent by txsyncLoop.
	// A pack can get larger than this if a single transactions exceeds this size.
	txsyncPackSize = 100 * 1024
)

type txsync struct {
	p   *peer
	txs []*modules.Transaction
}

// syncTransactions starts sending all currently pending transactions to the given peer.
//func (pm *ProtocolManager) syncTransactions(p *peer) {
//	var txs modules.Transactions
//	pending, _ := pm.txpool.Pending()
//	for _, this := range pending {
//		for _, batch := range this {
//			txs = append(txs, txspool.PooltxToTx(batch))
//		}
//	}
//	if len(txs) == 0 {
//		return
//	}
//	select {
//	case pm.txsyncCh <- &txsync{p, txs}:
//	case <-pm.quitSync:
//	}
//}

// txsyncLoop takes care of the initial transaction sync for each new
// connection. When a new peer appears, we relay all currently pending
// transactions. In order to minimize egress bandwidth usage, we send
// the transactions in small packs to one peer at a time.
func (pm *ProtocolManager) txsyncLoop() {
	var (
		pending = make(map[discover.NodeID]*txsync)
		sending = false               // whether a send is active
		pack    = new(txsync)         // the pack that is being sent
		done    = make(chan error, 1) // result of the send
	)

	// send starts a sending a pack of transactions from the sync.
	send := func(s *txsync) {
		// Fill pack with transactions up to the target size.
		size := common.StorageSize(0)
		pack.p = s.p
		pack.txs = pack.txs[:0]
		for i := 0; i < len(s.txs) && size < txsyncPackSize; i++ {
			pack.txs = append(pack.txs, s.txs[i])
			size += s.txs[i].Size()
		}
		// Remove the transactions that will be sent.
		s.txs = s.txs[:copy(s.txs, s.txs[len(pack.txs):])]
		if len(s.txs) == 0 {
			delete(pending, s.p.ID())
		}
		// Send the pack in the background.
		log.Trace("Sending batch of transactions", "count", len(pack.txs), "bytes", size)
		sending = true
		go func() { done <- pack.p.SendTransactions(pack.txs) }()
	}

	// pick chooses the next pending sync.
	pick := func() *txsync {
		if len(pending) == 0 {
			return nil
		}
		n := rand.Intn(len(pending)) + 1
		for _, s := range pending {
			if n--; n == 0 {
				return s
			}
		}
		return nil
	}

	for {
		select {
		case s := <-pm.txsyncCh:
			pending[s.p.ID()] = s
			if !sending {
				send(s)
			}
		case err := <-done:
			sending = false
			// Stop tracking peers that cause send failures.
			if err != nil {
				log.Debug("Transaction send failed", "err", err.Error())
				delete(pending, pack.p.ID())
			}
			// Schedule the next send.
			if s := pick(); s != nil {
				send(s)
			}
		case <-pm.quitSync:
			return
		}
	}
}

// syncer is responsible for periodically synchronizing with the network, both
// downloading hashes and blocks as well as handling the announcement handler.
func (pm *ProtocolManager) syncer(syncCh chan bool) {
	// Start and ensure cleanup of sync mechanisms
	pm.fetcher.Start()
	defer pm.fetcher.Stop()
	defer pm.downloader.Terminate()

	// Wait for different events to fire synchronization operations
	forceSync := time.NewTicker(forceSyncCycle)
	defer forceSync.Stop()

	for {
		select {
		case <-pm.newPeerCh:
			// Make sure we have peers to select from, then sync
			if pm.peers.Len() < minDesiredPeerCount {
				break
			}
			go pm.syncall(syncCh)

		case <-forceSync.C:
			// Force a sync even if not enough peers are present
			//log.Debug("start force Sync")
			go pm.syncall(syncCh)

		case <-pm.noMorePeers:
			return
		}
	}
}

func (pm *ProtocolManager) syncall(syncCh chan bool) {
	log.Debug("ProtocolManager syncall", "assetId", pm.mainAssetId)
	peer := pm.peers.BestPeer(pm.mainAssetId)
	pm.synchronize(peer, pm.mainAssetId, syncCh)
}

// synchronize tries to sync up our local block chain with a remote peer.
func (pm *ProtocolManager) synchronize(peer *peer, assetId modules.AssetId, syncCh chan bool) {
	// Short circuit if no peers are available
	if peer == nil {
		log.Debug("ProtocolManager synchronize peer is nil")
		return
	}
	log.Debug("Enter ProtocolManager synchronize", "peer id:", peer.id)
	defer log.Debug("End ProtocolManager synchronize", "peer id:", peer.id)

	// Make sure the peer's TD is higher than our own
	//TODO compare local assetId & chainIndex with remote peer assetId & chainIndex
	currentUnit := pm.dag.GetCurrentUnit(assetId)
	if currentUnit == nil {
		log.Error("synchronize currentUnit is nil have not genesis")
		return
	}
	index := currentUnit.Number().Index
	pHead, number := peer.Head(assetId)
	pindex := number.Index
	var err error = nil
	if currentUnit.Number().Index > 0 {
		_, err = pm.dag.GetUnitByHash(currentUnit.ParentHash()[0])
	}

	//if index >= pindex && pindex > 0 && err == nil {
	if index >= pindex && err == nil {
		if atomic.LoadUint32(&pm.fastSync) == 1 {
			log.Debug("Fast sync complete, auto disabling")
			atomic.StoreUint32(&pm.fastSync, 0)
		}
		atomic.StoreUint32(&pm.acceptTxs, 1)
		log.Debug("Do not need synchronize", "local peer.index:", pindex, "local index:", number.Index,
			"header hash:", pHead)
		//TODO notice light protocol to sync corsheader
		if syncCh != nil {
			syncCh <- true
		}
		return
	}
	log.Debug("ProtocolManager", "synchronize local unit index:", index, "local peer index:", pindex,
		"header hash:", pHead)
	// Otherwise try to sync with the downloader
	mode := downloader.FullSync

	if atomic.LoadUint32(&pm.fastSync) == 1 {
		// Fast sync was explicitly requested, and explicitly granted
		mode = downloader.FastSync
	}
	log.Debug("ProtocolManager", "synchronize local unit index:", index, "peer index:", pindex,
		"header hash:", pHead)
	pm.downloader.SetFastStableIndex(pm.peers.StableIndex(assetId))
	// Run the sync cycle, and disable fast sync if we've went past the pivot block
	if err := pm.downloader.Synchronize(peer.id, pHead, pindex, mode, assetId); err != nil {
		log.Debug("ptn sync downloader.", "Synchronize err:", err)
		return
	}

	if atomic.LoadUint32(&pm.fastSync) == 1 {
		log.Info("Fast sync complete, auto disabling")
		atomic.StoreUint32(&pm.fastSync, 0)
	}
	atomic.StoreUint32(&pm.acceptTxs, 1) // Mark initial sync done
	log.Info("ptn sync complete")

	cunit := pm.dag.GetCurrentUnit(assetId)
	if cunit != nil && cunit.UnitHeader.Number.Index > 0 {
		go pm.BroadcastUnit(cunit, false)
		//TODO notice light protocol to sync corsheader
		if syncCh != nil {
			syncCh <- true
		}
	}
}
