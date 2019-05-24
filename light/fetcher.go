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

// Package les implements the Light Palletone Subprotocol.
package light

import (
	"errors"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	dagerrors "github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	blockDelayTimeout = time.Second * 10 // timeout for a peer to announce a head that has already been confirmed by others
	maxNodeCount      = 20               // maximum number of fetcherTreeNode entries remembered for each peer
	//retryQueue         = time.Millisecond * 100
	//softRequestTimeout = time.Millisecond * 500
	//hardRequestTimeout = time.Second * 10

	arriveTimeout = 500 * time.Millisecond // Time allowance before an announced block is explicitly requested
	gatherSlack   = 100 * time.Millisecond // Interval used to collate almost-expired announces with fetches
	fetchTimeout  = 5 * time.Second        // Maximum allotted time to return an explicitly requested block
	maxUncleDist  = 14                     /*7*/ // Maximum allowed backward distance from the chain head
	maxQueueDist  = 32                     // Maximum allowed distance from the chain head to queue
	hashLimit     = 256                    // Maximum number of unique blocks a peer may have announced
	blockLimit    = 64                     // Maximum number of unique blocks a peer may have delivered
)

var (
	errTerminated = errors.New("terminated")
)

// blockRetrievalFn is a callback type for retrieving a block from the local chain.
type headerRetrievalFn func(common.Hash) (*modules.Header, error)

// headerRequesterFn is a callback type for sending a header retrieval request.
type headerRequesterFn func(common.Hash) error

// headerVerifierFn is a callback type to verify a block's header for fast propagation.
type headerVerifierFn func(header *modules.Header) error

// blockBroadcasterFn is a callback type for broadcasting a block to connected peers.
type headerBroadcasterFn func(header *modules.Header, propagate bool)

// chainHeightFn is a callback type to retrieve the current chain height.
type lightChainHeightFn func(assetId modules.AssetId) uint64

// chainInsertFn is a callback type to insert a batch of blocks into the local chain.
type headerInsertFn func(headers []*modules.Header) (int, error)

// peerDropFn is a callback type for dropping a peer detected as malicious.
type peerDropFn func(id string)

// headerFilterTask represents a batch of headers needing fetcher filtering.
type headerFilterTask struct {
	peer    string            // The source peer of block headers
	headers []*modules.Header // Collection of headers to filter
	time    time.Time         // Arrival time of the headers
}

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
	headerFilter chan chan *headerFilterTask

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
	verifyHeader     headerVerifierFn    // Checks if a block's headers have a valid proof of work
	broadcastHeader  headerBroadcasterFn // Broadcasts a block to connected peers
	insertHeader     headerInsertFn      // Injects a batch of blocks into the chain
	dropPeer         peerDropFn          // Drops a peer for misbehaving
	getHeaderByHash  headerRetrievalFn   // Retrieves a block from the local chain
	lightChainHeight lightChainHeightFn  // Retrieves the current chain's height
}

// New creates a block fetcher to retrieve blocks based on hash announcements.
func NewLightFetcher(getHeaderByHash headerRetrievalFn, lightChainHeight lightChainHeightFn, verifyHeader headerVerifierFn,
	broadcastHeader headerBroadcasterFn, insertHeader headerInsertFn, dropPeer peerDropFn) *LightFetcher {
	return &LightFetcher{
		//notify:           make(chan *announce),
		inject:       make(chan *inject),
		headerFilter: make(chan chan *headerFilterTask),
		done:         make(chan common.Hash),
		quit:         make(chan struct{}),
		//announces:        make(map[string]int),
		//announced:        make(map[common.Hash][]*announce),
		//fetching:         make(map[common.Hash]*announce),
		//fetched:          make(map[common.Hash][]*announce),
		//completing:       make(map[common.Hash]*announce),
		queue:            prque.New(),
		queues:           make(map[string]int),
		queued:           make(map[common.Hash]*inject),
		verifyHeader:     verifyHeader,
		broadcastHeader:  broadcastHeader,
		insertHeader:     insertHeader,
		lightChainHeight: lightChainHeight,
		getHeaderByHash:  getHeaderByHash,
		dropPeer:         dropPeer,
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

			//TODO must recover the check
			//// If too high up the chain or phase, continue later
			//height = f.lightChainHeight(op.header.Number.AssetID)
			//number := op.header.Index()
			//if number > height+1 {
			//	f.queue.Push(op, -float32(op.header.Index()))
			//	break
			//}
			//// Otherwise if fresh and still unknown, try and import
			//hash := op.header.Hash()
			//header, _ := f.getHeaderByHash(hash)
			//if number+maxUncleDist < height || header != nil {
			//	f.forgetBlock(hash)
			//	continue
			//}
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

			/*
				case notification := <-f.notify:
					// A block was announced, make sure the peer isn't DOSing us
					count := f.announces[notification.origin] + 1
					if count > hashLimit {
						log.Debug("Peer exceeded outstanding announces", "peer", notification.origin, "limit", hashLimit)
						break
					}
					// If we have a valid block number, check that it's potentially useful
					if notification.number.Index > 0 {
						if dist := int64(notification.number.Index) - int64(f.chainHeight(notification.number.AssetID)); dist < -maxUncleDist || dist > maxQueueDist {
							log.Debug("Peer discarded announcement", "peer", notification.origin, "number", notification.number, "hash", notification.hash, "distance", dist)
							break
						}
					}

					// All is well, schedule the announce if block's not yet downloading
					if _, ok := f.fetching[notification.hash]; ok {
						break
					}
					if _, ok := f.completing[notification.hash]; ok {
						break
					}
					f.announces[notification.origin] = count
					f.announced[notification.hash] = append(f.announced[notification.hash], notification)

					if len(f.announced) == 1 {
						f.rescheduleFetch(fetchTimer)
					}

				case <-fetchTimer.C:
					// At least one block's timer ran out, check for needing retrieval
					request := make(map[string][]common.Hash)

					for hash, announces := range f.announced {
						if time.Since(announces[0].time) > arriveTimeout-gatherSlack {
							// Pick a random peer to retrieve from, reset all others
							announce := announces[rand.Intn(len(announces))]
							f.forgetHash(hash)

							// If the block still didn't arrive, queue for fetching
							block, _ := f.getBlock(hash)
							if block == nil {
								request[announce.origin] = append(request[announce.origin], hash)
								f.fetching[hash] = announce
							}
						}
					}
					log.Debug("===fetcher <-fetchTimer.C===", "len(request):", len(request))
					// Send out all block header requests
					for peer, hashes := range request {
						log.Trace("Fetching scheduled headers", "peer", peer, "list", hashes)
						// Create a closure of the fetch and schedule in on a new thread
						fetchHeader, hashes := f.fetching[hashes[0]].fetchHeader, hashes
						go func(fetchHeader headerRequesterFn, hashes []common.Hash) {
							for _, hash := range hashes {
								fetchHeader(hash) // Suboptimal, but protocol doesn't allow batch header retrievals
							}
						}(fetchHeader, hashes)
					}
					// Schedule the next fetch if blocks are still pending
					f.rescheduleFetch(fetchTimer)

				case <-completeTimer.C:
					// At least one header's timer ran out, retrieve everything
					request := make(map[string][]common.Hash)

					for hash, announces := range f.fetched {
						// Pick a random peer to retrieve from, reset all others
						announce := announces[rand.Intn(len(announces))]
						f.forgetHash(hash)

						// If the block still didn't arrive, queue for completion
						block, _ := f.getBlock(hash)
						if block == nil {
							request[announce.origin] = append(request[announce.origin], hash)
							f.completing[hash] = announce
						}
					}
					// Send out all block body requests
					for peer, hashes := range request {
						log.Trace("Fetching scheduled bodies", "peer", peer, "list", hashes)

						// Create a closure of the fetch and schedule in on a new thread
						go f.completing[hashes[0]].fetchBodies(hashes)
					}
					// Schedule the next fetch if blocks are still pending
					f.rescheduleComplete(completeTimer)

				case filter := <-f.headerFilter:
					// Headers arrived from a remote peer. Extract those that were explicitly
					// requested by the fetcher, and return everything else so it's delivered
					// to other parts of the system.
					var task *headerFilterTask
					select {
					case task = <-filter:
					case <-f.quit:
						return
					}

					// Split the batch of headers into unknown ones (to return to the caller),
					// known incomplete ones (requiring body retrievals) and completed blocks.
					unknown, incomplete, complete := []*modules.Header{}, []*announce{}, []*modules.Unit{}
					for _, header := range task.headers {
						hash := header.Hash()

						// Filter fetcher-requested headers from other synchronisation algorithms
						if announce := f.fetching[hash]; announce != nil && announce.origin == task.peer && f.fetched[hash] == nil && f.completing[hash] == nil && f.queued[hash] == nil {
							// If the delivered header does not match the promised number, drop the announcer
							if header.Number.Index != announce.number.Index &&
								header.Number.AssetID == announce.number.AssetID {
								log.Trace("Invalid block number fetched", "peer", announce.origin, "hash", header.Hash(), "announced", announce.number, "provided", header.Number)
								f.dropPeer(announce.origin)
								f.forgetHash(hash)
								continue
							}
							// Only keep if not imported by other means
							blk, _ := f.getBlock(hash)
							if blk == nil {
								announce.header = header
								announce.time = task.time

								// If the block is empty (header only), short circuit into the final import queue
								//TODO modify
								if header.TxRoot == core.DeriveSha(modules.Transactions{}) {
									log.Trace("Block empty, skipping body retrieval", "peer", announce.origin, "number", header.Number, "hash", header.Hash())

									block := modules.NewUnitWithHeader(header)
									block.ReceivedAt = task.time

									complete = append(complete, block)
									f.completing[hash] = announce
									continue
								}
								// Otherwise add to the list of blocks needing completion
								incomplete = append(incomplete, announce)
							} else {
								log.Trace("Block already imported, discarding header", "peer", announce.origin, "number", header.Number, "hash", header.Hash())
								f.forgetHash(hash)
							}
						} else {
							// Fetcher doesn't know about it, add to the return list
							unknown = append(unknown, header)
						}
					}
					select {
					case filter <- &headerFilterTask{headers: unknown, time: task.time}:
					case <-f.quit:
						return
					}
					// Schedule the retrieved headers for body completion
					for _, announce := range incomplete {
						hash := announce.header.Hash()
						if _, ok := f.completing[hash]; ok {
							continue
						}
						f.fetched[hash] = append(f.fetched[hash], announce)
						if len(f.fetched) == 1 {
							f.rescheduleComplete(completeTimer)
						}
					}
					// Schedule the header-only blocks for import
					for _, block := range complete {
						if announce := f.completing[block.Hash()]; announce != nil {
							f.enqueue(announce.origin, block)
						}
					}
			*/
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

	// Run the import on a new thread
	log.Debug("Importing propagated block insert DAG", "peer", p.id, "number", header.Index(), "hash", hash)
	go func() {
		defer func() { f.done <- hash }()

		// If the parent's unknown, abort insertion
		//TODO must recover
		//parentsHash := header.ParentsHash
		//for _, parentHash := range parentsHash {
		//	had, _ := f.getHeaderByHash(parentHash)
		//	if had == nil {
		//		log.Debug("Unknown parent of propagated block", "peer", peer, "number", header.Number.Index, "hash", hash, "parent", parentHash)
		//		return
		//	}
		//}

		// Quickly validate the header and propagate the block if it passes
		switch err := f.verifyHeader(header); err {
		case nil:
			// All ok, quickly propagate to our peers
			go f.broadcastHeader(header, true)

		case dagerrors.ErrFutureBlock:
			// Weird future block, don't fail, but neither propagate

		default:
			// Something went very wrong, drop the peer
			log.Debug("Propagated block verification failed", "peer", p.id, "number", header.Index(), "hash", hash, "err", err)
			f.dropPeer(p.id)
			return
		}

		// Run the actual import and log any issues
		if _, err := f.insertHeader([]*modules.Header{header}); err != nil {
			log.Debug("Propagated block import failed", "peer", p.id, "number", header.Index(), "hash", hash, "err", err)
			return
		}
		p.lightlock.Lock()
		p.lightpeermsg[header.Number.AssetID] = &announceData{Hash: header.Hash(), Number: *header.Number}
		p.lightlock.Unlock()
		// If import succeeded, broadcast the block
		go f.broadcastHeader(header, false)

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
		log.Debug("Discarded propagated block, exceeded allowance", "peer", p.id, "number", header.Index(), "hash", hash, "limit", blockLimit)
		return
	}
	// Discard any past or too distant blocks
	heightChain := int64(f.lightChainHeight(header.Number.AssetID))
	if dist := int64(header.Number.Index) - heightChain; dist < -maxUncleDist || dist > maxQueueDist {
		log.Debug("Discarded propagated block, too far away", "peer", p.id, "number", header.Index(), "heightChain", heightChain, "distance", dist)
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
		log.Debug("Queued propagated block", "peer", p.id, "number", header.Index(), "hash", hash, "queued", f.queue.Size())
	}
}

// Enqueue tries to fill gaps the the fetcher's future import queue.
func (f *LightFetcher) Enqueue(p *peer, header *modules.Header) error {
	log.Debug("Enter LightFetcher Enqueue", "peer id", p.id, "header index:", header.Index())
	defer log.Debug("End LightFetcher Enqueue")
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
