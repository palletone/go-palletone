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

package cors

import (
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"math/rand"
	"sync/atomic"
	"time"
)

// dataPack is a data message returned by a peer for some query.
type dataPack interface {
	PeerId() string
	Items() int
	Stats() string
}

// headerPack is a batch of block headers returned by a peer.
type headerPack struct {
	peerId  string
	headers []*modules.Header
}

func (p *headerPack) PeerId() string { return p.peerId }
func (p *headerPack) Items() int     { return len(p.headers) }
func (p *headerPack) Stats() string  { return fmt.Sprintf("%d", len(p.headers)) }

func (pm *ProtocolManager) StartCorsSync() (string, error) {
	mainchain, err := pm.dag.GetMainChain()
	if mainchain == nil || err != nil {
		log.Debug("Cors ProtocolManager StartCorsSync", "GetMainChain err", err)
		return err.Error(), err
	}
	pm.mainchain = mainchain
	pm.mclock.RLock()
	for _, peer := range mainchain.Peers {
		node, err := discover.ParseNode(peer)
		if err != nil {
			return fmt.Sprintf("Cors ProtocolManager StartCorsSync invalid pnode: %v", err), err
		}
		log.Debug("Cors ProtocolManager StartCorsSync", "peer:", peer)
		pm.server.corss.AddPeer(node)
	}
	pm.mclock.RUnlock()

	go func() {
		for {

			if pm.peers.Len() >= pm.mainchainpeers()/2+1 {
				pm.Sync()
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}()

	return "OK", nil
}

func (pm *ProtocolManager) Sync() {
	log.Debug("Enter Cors ProtocolManager Sync")
	defer log.Debug("End Cors ProtocolManager Sync")
	if atomic.LoadUint32(&pm.corsSync) == 0 {
		atomic.StoreUint32(&pm.corsSync, 1)
		index, _ := pm.pushSync()
		log.Info("Cors Sync OK", "index", index)
		atomic.StoreUint32(&pm.corsSync, 0)
	}
}

func (pm *ProtocolManager) pushSync() (uint64, error) {
	var (
		bytes   common.StorageSize
		headers []*modules.Header
		index   uint64
		flag    int
	)

	pheader, err := pm.fetchHeader()
	if err != nil {
		log.Debug("Cors ProtocolManager", "pushSync fetchHeader err", err)
		return 0, err
	}

	flag = 0
	if pheader.Number.Index <= fsMinFullBlocks {
		index = 0
	} else {
		index = pheader.Number.Index - fsMinFullBlocks
	}

	log.Debug("Cors ProtocolManager", "pheader.index", pheader.Number.Index, "push index", index, "pushSync fetchHeader header", pheader)

	number := &modules.ChainIndex{pm.assetId, index}
	for {
		bytes = 0
		headers = []*modules.Header{}

		for bytes < softResponseLimit && len(headers) < MaxHeaderFetch {
			bytes += estHeaderRlpSize
			number.Index = index
			header, err := pm.dag.GetHeaderByNumber(number)
			if err != nil {
				if len(headers) == MaxHeaderFetch {
					index--
					break
				} else {
					flag = 1
				}
				break
			}
			headers = append(headers, header)
			index++
		}

		rand.Seed(time.Now().UnixNano())
		peers := pm.peers.AllPeers()
		x := rand.Intn(len(peers))
		p := peers[x]
		log.Info("Cors ProtocolManager", "pushSync SendHeaders len(headers)", len(headers), "index", index)
		if len(headers) == 0 {
			header := modules.Header{}
			number := modules.ChainIndex{pm.assetId, 0}
			header.Number = &number
			headers = append(headers, &header)
		}
		p.SendHeaders(headers)
		if flag == 1 {
			break
		} else {
			time.Sleep(waitPushSync)
		}
	}
	return index, nil
}

// requestTTL returns the current timeout allowance for a single download request
// to finish under.
func (pm *ProtocolManager) requestTTL() time.Duration {
	var (
		rtt  = time.Duration(atomic.LoadUint64(&pm.rttEstimate))
		conf = float64(atomic.LoadUint64(&pm.rttConfidence)) / 1000000.0
	)
	ttl := time.Duration(ttlScaling) * time.Duration(float64(rtt)/conf)
	if ttl > ttlLimit {
		ttl = ttlLimit
	}
	return ttl
}

func (pm *ProtocolManager) fetchHeader() (*modules.Header, error) {
	// Request the advertised remote head block and wait for the response
	rand.Seed(time.Now().UnixNano())
	peers := pm.peers.AllPeers()
	log.Debug("Cors ProtocolManager fetchHeader", "len(peers)", len(peers))
	x := rand.Intn(len(peers))
	p := peers[x]
	log.Debug("Retrieving remote all token", "peer", p.ID())
	var number modules.ChainIndex
	number.AssetID = pm.assetId
	go p.RequestCurrentHeader(number)

	ttl := pm.requestTTL()
	timeout := time.After(ttl)
	for {
		select {
		case <-pm.quitSync:
			return nil, errCancelHeaderFetch

		case packet := <-pm.headerCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Debug("Received headers from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			headers := packet.(*headerPack).headers
			if len(headers) != 1 {
				log.Debug("Multiple headers for single request", "headers", len(headers), "peer", p.id)
				return nil, errBadPeer
			}
			log.Debug("Remote leaf nodes", "counts", len(headers), "peer", packet.PeerId())
			return headers[0], nil

		case <-timeout:
			log.Debug("Waiting for head header timed out", "elapsed", ttl, "peer", p.id)
			return nil, errTimeout
		}
	}
	return nil, nil
}
