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
 * @author PalletOne core developer Jiyou Wang <dev@pallet.one>
 * @date 2018
 */
package cors

import (
	"errors"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	maxUncleDist = 14 /*7*/ // Maximum allowed backward distance from the chain head
	maxQueueDist = 32 // Maximum allowed distance from the chain head to queue
	// Maximum number of unique blocks a peer may have delivered
	blockLimit = 65536 //headersiz:1024 bytes   fetchersize:1024*1024*64bytes    blockLimit:1024*64

)

var (
	errTerminated = errors.New("terminated")
)

// blockRetrievalFn is a callback type for retrieving a block from the local chain.
type headerRetrievalFn func(common.Hash) (*modules.Header, error)

// headerRequesterFn is a callback type for sending a header retrieval request.
//type headerRequesterFn func(common.Hash) error

// headerVerifierFn is a callback type to verify a block's header for fast propagation.
type headerVerifierFn func(header *modules.Header) error

// blockBroadcasterFn is a callback type for broadcasting a block to connected peers.
type headerBroadcasterFn func(p *peer, header *modules.Header, propagate bool)

// chainHeightFn is a callback type to retrieve the current chain height.
type lightChainHeightFn func(assetId modules.AssetId) uint64

// chainInsertFn is a callback type to insert a batch of blocks into the local chain.
type headerInsertFn func(headers []*modules.Header) (int, error)

// peerDropFn is a callback type for dropping a peer detected as malicious.
type peerDropFn func(id string)

// headerFilterTask represents a batch of headers needing fetcher filtering.
//type headerFilterTask struct {
//	peer    string            // The source peer of block headers
//	headers []*modules.Header // Collection of headers to filter
//	time    time.Time         // Arrival time of the headers
//}

// inject represents a schedules import operation.
type inject struct {
	origin *peer
	header *modules.Header
}

// Fetcher is responsible for accumulating block announcements from various peers
// and scheduling them for retrieval.
type LightFetcher struct {
	// Various event channels
	//notify chan *announce
	inject chan *inject

	//blockFilter  chan chan []*modules.Unit
	//headerFilter chan chan *headerFilterTask

	done chan common.Hash
	quit chan struct{}

	// Announce states
	//announces  map[string]int              // Per peer announce counts to prevent memory exhaustion
	//announced  map[common.Hash][]*announce // Announced blocks, scheduled for fetching
	//fetching   map[common.Hash]*announce   // Announced blocks, currently fetching
	//fetched    map[common.Hash][]*announce // Blocks with headers fetched, scheduled for body retrieval
	//completing map[common.Hash]*announce   // Blocks with headers, currently body-completing

	// Block cache
	queue  *prque.Prque            // Queue containing the import operations (block number sorted)
	queues map[string]int          // Per peer block counts to prevent memory exhaustion
	queued map[common.Hash]*inject // Set of already queued blocks (to dedupe imports)

	// Callbacks
	broadcastHeader  headerBroadcasterFn // Broadcasts a block to connected peers
	insertHeader     headerInsertFn      // Injects a batch of blocks into the chain
	dropPeer         peerDropFn          // Drops a peer for misbehaving
	lightChainHeight lightChainHeightFn  // Retrieves the current chain's height
}

// New creates a block fetcher to retrieve blocks based on hash announcements.
func NewLightFetcher(getHeaderByHash headerRetrievalFn, lightChainHeight lightChainHeightFn,
	verifyHeader headerVerifierFn, broadcastHeader headerBroadcasterFn, insertHeader headerInsertFn,
	dropPeer peerDropFn) *LightFetcher {
	return &LightFetcher{
		//notify:           make(chan *announce),
		inject: make(chan *inject),
		//headerFilter: make(chan chan *headerFilterTask),
		done: make(chan common.Hash),
		quit: make(chan struct{}),

		queue:  prque.New(),
		queues: make(map[string]int),
		queued: make(map[common.Hash]*inject),
		//verifyHeader:     verifyHeader,
		broadcastHeader:  broadcastHeader,
		insertHeader:     insertHeader,
		lightChainHeight: lightChainHeight,
		//getHeaderByHash:  getHeaderByHash,
		dropPeer: dropPeer,
	}
}

// Start boots up the announcement based synchroniser, accepting and processing
// hash notifications and block fetches until termination requested.
func (f *LightFetcher) Start() {
	go f.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (f *LightFetcher) Stop() {
	close(f.quit)
}

// Loop is the main fetcher loop, checking and processing various notification
// events.
func (f *LightFetcher) loop() {
	// Iterate the block fetching until a quit is requested
	//fetchTimer := time.NewTimer(0)
	//completeTimer := time.NewTimer(0)
	for {
		//TODO Clean up any expired block fetches
		// Import any queued blocks that could potentially fit
		//var height uint64
		for !f.queue.Empty() {
			op := f.queue.PopItem().(*inject)
			f.insert(op.origin, op.header)
		}
		// Wait for an outside event to occur
		select {
		case <-f.quit:
			// Fetcher terminating, abort all operations
			return

		case op := <-f.inject:
			// A direct block insertion was requested, try and fill any pending gaps
			f.enqueue(op.origin, op.header)

		case hash := <-f.done:
			// A pending import finished, remove all traces of the notification
			f.forgetBlock(hash)
		}
	}
}

// forgetBlock removes all traces of a queued block from the fetcher's internal
// state.
func (f *LightFetcher) forgetBlock(hash common.Hash) {
	if insert := f.queued[hash]; insert != nil {
		f.queues[insert.origin.id]--
		if f.queues[insert.origin.id] == 0 {
			delete(f.queues, insert.origin.id)
		}
		delete(f.queued, hash)
	}
}

// insert spawns a new goroutine to run a block insertion into the chain. If the
// block's number is at the same height as the current import phase, it updates
// the phase states accordingly.
func (f *LightFetcher) insert(p *peer, header *modules.Header) {
	hash := header.Hash()
	go func() {
		defer func() { f.done <- hash }()
		// Run the actual import and log any issues
		if _, err := f.insertHeader([]*modules.Header{header}); err != nil {
			log.Debug("Propagated block import failed", "peer", p.id, "number", header.Number,
				"hash", hash, "err", err)
			return
		}
		// If import succeeded, broadcast the block
		go func() {
			time.Sleep(time.Duration(5) * time.Millisecond)
			f.broadcastHeader(p, header, false)
		}()
	}()
}

// enqueue schedules a new future import operation, if the block to be imported
// has not yet been seen.
func (f *LightFetcher) enqueue(p *peer, header *modules.Header) {
	log.Debug("Enter LightFetcher enqueue", "peer id", p.id, "header index:", header.Index())
	defer log.Debug("End LightFetcher enqueue")
	hash := header.Hash()
	// Ensure the peer isn't DOSing us
	count := f.queues[p.id] + 1
	if count > blockLimit {
		log.Debug("Discarded propagated block, exceeded allowance", "peer", p.id, "number", header.Index(),
			"hash", hash, "limit", blockLimit)
		return
	}
	log.Debug("Cors Fetcher propagated block, current allowance", "count", count, "limit", blockLimit)
	// Discard any past or too distant blocks
	heightChain := int64(f.lightChainHeight(header.Number.AssetID))
	if dist := int64(header.Number.Index) - heightChain; dist < -maxUncleDist || dist > maxQueueDist {
		log.Debug("Discarded propagated block, too far away", "peer", p.id, "number", header.Index(),
			"heightChain", heightChain, "distance", dist)
		return
	}
	// Schedule the block for future importing
	if _, ok := f.queued[hash]; !ok {
		op := &inject{
			origin: p,
			header: header,
		}
		f.queues[p.id] = count
		f.queued[hash] = op
		f.queue.Push(op, -float32(header.Index()))
		log.Debug("Queued propagated block", "peer", p.id, "number", header.Index(), "hash", hash,
			"queued", f.queue.Size())
	}
}

// Enqueue tries to fill gaps the the fetcher's future import queue.
func (f *LightFetcher) Enqueue(p *peer, header *modules.Header) error {
	//log.Debug("Enter CorsFetcher Enqueue", "peer id", p.id, "header index:", header.Index())
	//defer log.Debug("End CorsFetcher Enqueue")
	op := &inject{
		origin: p,
		header: header,
	}
	select {
	case f.inject <- op:
		return nil
	case <-f.quit:
		return errTerminated
	}
}
