// Copyright 2016 The go-ethereum Authors
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

package light

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
	"sync/atomic"
	"time"
)

const (
	forceSyncCycle      = 10 * time.Second // Time interval to force syncs, even if few peers are available
	minDesiredPeerCount = 5                // Amount of peers desired to start syncing
)

// syncer is responsible for periodically synchronizing with the network, both
// downloading hashes and blocks as well as handling the announcement handler.
func (pm *ProtocolManager) syncer(syncCh chan bool) {
	// Start and ensure cleanup of sync mechanisms
	pm.fetcher.Start()
	defer pm.fetcher.Stop()
	defer pm.downloader.Terminate()

	// Wait for different events to fire synchronization operations
	if pm.lightSync {
		forceSync := time.NewTicker(forceSyncCycle)
		defer forceSync.Stop()
		for {
			select {
			case <-pm.newPeerCh:
				// Make sure we have peers to select from, then sync
				if pm.peers.Len() < minDesiredPeerCount {
					break
				}
				go pm.syncall()

			case <-forceSync.C:
				// Force a sync even if not enough peers are present
				go pm.syncall()

			case <-pm.noMorePeers:
				return
			}
		}
	} else {
		for {
			select {
			case <-pm.newPeerCh:
				// Make sure we have peers to select from, then sync
			case <-syncCh:
				go pm.syncall()

			case <-pm.noMorePeers:
				return
			}
		}
	}
}

func (pm *ProtocolManager) syncall() {
	log.Debug("Enter Light PalletOne syncall")
	defer log.Debug("End Light PalletOne syncall")
	if atomic.LoadUint32(&pm.fastSync) == 0 {
		log.Debug("Light PalletOne syncall synchronizing")
		return
	}

	p := pm.peers.BestPeer(pm.assetId)
	if p == nil {
		log.Debug("Light PalletOne syncall peer is nil")
		return
	}
	headers, err := pm.downloader.FetchAllToken(p.id)
	if err != nil {
		log.Debug("Light PalletOne syncall FetchAllToken", "err", err)
	}
	//log.Debug("Light PalletOne syncall FetchAllToken", "len(headers)", len(headers), "headers", headers)
	for _, header := range headers {
		log.Debug("Light PalletOne syncall synchronize", "asset", header.Number.AssetID,
			"index", header.Number.Index)
		pm.synchronize(p, header.Number.AssetID)
	}
}

// synchronize tries to sync up our local block chain with a remote peer.
func (pm *ProtocolManager) synchronize(peer *peer, assetId modules.AssetId) {
	// Short circuit if no peers are available
	if peer == nil {
		return
	}

	if !pm.lightSync && pm.assetId == assetId {
		log.Debug("Light PalletOne synchronize pm.assetId == assetId")
		return
	}

	if pm.lightSync && pm.assetId != assetId {
		return
	}

	if pm.assetId != assetId {
		access := false
		if pcs, err := pm.dag.GetPartitionChains(); err == nil {
			for _, pc := range pcs {
				for _, token := range pc.CrossChainTokens {
					if token == assetId {
						access = true
					}
				}
			}
		} else {
			return
		}
		if !access {
			return
		}
	}

	headhash, number := peer.HeadAndNumber(assetId)
	if common.EmptyHash(headhash) || number == nil {
		log.Debug("Light PalletOne synchronize is nil", "assetId", assetId)
		return
	}

	lheader := pm.dag.CurrentHeader(assetId)
	if lheader != nil && lheader.Number.Index >= number.Index {
		log.Debug("Light PalletOne synchronize is not need sync", "local index",
			lheader.Number.Index, "peer index", number.Index)
		return
	}
	if lheader == nil {
		log.Debug("Light PalletOne synchronize local header is nil", "assetid", assetId)
	}

	if atomic.LoadUint32(&pm.fastSync) == 0 {
		log.Debug("Light PalletOne synchronizing")
		return
	}
	atomic.StoreUint32(&pm.fastSync, 0)
	defer atomic.StoreUint32(&pm.fastSync, 1)

	log.Debug("Enter Light PalletOne ProtocolManager synchronize", "assetid", assetId, "index", number.Index)
	defer log.Debug("End Light PalletOne ProtocolManager synchronize", "assetid", assetId,
		"index", number.Index)

	if err := pm.downloader.Synchronize(peer.id, headhash, number.Index,
		downloader.LightSync, number.AssetID); err != nil {
		log.Debug("Light PalletOne ProtocolManager synchronize", "Synchronize err:", err)
		return
	}

	//if atomic.LoadUint32(&pm.fastSync) == 0 {
	//	log.Debug("Fast sync complete, auto disabling")
	//	atomic.StoreUint32(&pm.fastSync, 1)
	//}

	header := pm.dag.CurrentHeader(assetId)
	if header != nil && header.Number.Index > 0 {
		go pm.BroadcastLightHeader(header)
	}

}
