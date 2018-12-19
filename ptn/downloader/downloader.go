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

// Package downloader contains the manual full chain synchronisation.
package downloader

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	palletone "github.com/palletone/go-palletone"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	dagerrors "github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/statistics/metrics"
)

var (
	MaxHashFetch    = 512 // Amount of hashes to be fetched per retrieval request
	MaxBlockFetch   = 128 // Amount of blocks to be fetched per retrieval request
	MaxHeaderFetch  = 192 // Amount of block headers to be fetched per retrieval request
	MaxSkeletonSize = 128 // Number of header fetches to need for a skeleton assembly
	MaxBodyFetch    = 128 // Amount of block bodies to be fetched per retrieval request
	//MaxReceiptFetch = 256 // Amount of transaction receipts to allow fetching per request
	MaxStateFetch = 384 // Amount of node state values to allow fetching per request

	MaxForkAncestry  = 3 * configure.EpochDuration // Maximum chain reorganisation
	rttMinEstimate   = 2 * time.Second             // Minimum round-trip time to target for download requests
	rttMaxEstimate   = 20 * time.Second            // Maximum round-trip time to target for download requests
	rttMinConfidence = 0.1                         // Worse confidence factor in our estimated RTT value
	ttlScaling       = 3                           // Constant scaling factor for RTT -> TTL conversion
	ttlLimit         = time.Minute                 // Maximum TTL allowance to prevent reaching crazy timeouts

	qosTuningPeers   = 5    // Number of peers to tune based on (best peers)
	qosConfidenceCap = 10   // Number of peers above which not to modify RTT confidence
	qosTuningImpact  = 0.25 // Impact that a new tuning target has on the previous value

	maxQueuedHeaders  = 32 * 1024 // [ptn/62] Maximum number of headers to queue for import (DOS protection)
	maxHeadersProcess = 2048      // Number of header download results to import at once into the chain
	maxResultsProcess = 2048      // Number of content download results to import at once into the chain

	fsHeaderCheckFrequency = 100             // Verification frequency of the downloaded headers during fast sync
	fsHeaderSafetyNet      = 2048            // Number of headers to discard in case a chain violation is detected
	fsHeaderForceVerify    = 24              // Number of headers to verify before and after the pivot to accept it
	fsHeaderContCheck      = 3 * time.Second // Time interval to check for header continuations during state download
	fsMinFullBlocks        = 64              // Number of blocks to retrieve fully even in fast sync
)

var (
	errBusy                    = errors.New("busy")
	errUnknownPeer             = errors.New("peer is unknown or unhealthy")
	errBadPeer                 = errors.New("action from bad peer ignored")
	errStallingPeer            = errors.New("peer is stalling")
	errNoPeers                 = errors.New("no peers to keep download active")
	errTimeout                 = errors.New("timeout")
	errEmptyHeaderSet          = errors.New("empty header set by peer")
	errPeersUnavailable        = errors.New("no peers available or all tried for download")
	errInvalidAncestor         = errors.New("retrieved ancestor is invalid")
	errInvalidChain            = errors.New("retrieved hash chain is invalid")
	errInvalidBlock            = errors.New("retrieved block is invalid")
	errInvalidBody             = errors.New("retrieved block body is invalid")
	errInvalidReceipt          = errors.New("retrieved receipt is invalid")
	errCancelBlockFetch        = errors.New("block download canceled (requested)")
	errCancelHeaderFetch       = errors.New("block header download canceled (requested)")
	errCancelBodyFetch         = errors.New("block body download canceled (requested)")
	errCancelReceiptFetch      = errors.New("receipt download canceled (requested)")
	errCancelStateFetch        = errors.New("state data download canceled (requested)")
	errCancelHeaderProcessing  = errors.New("header processing canceled (requested)")
	errCancelContentProcessing = errors.New("content processing canceled (requested)")
	errNoSyncActive            = errors.New("no sync active")
	errTooOld                  = errors.New("peer doesn't speak recent enough protocol version (need version >= 62)")
)

type Downloader struct {
	mode SyncMode       // Synchronisation mode defining the strategy used (per sync cycle)
	mux  *event.TypeMux // Event multiplexer to announce sync operation events

	queue *queue   // Scheduler for selecting the hashes to download
	peers *peerSet // Set of active peers from which download can proceed
	//levelDb palletdb.Database

	rttEstimate   uint64 // Round trip time to target for download requests
	rttConfidence uint64 // Confidence in the estimated RTT (unit: millionths to allow atomic ops)

	// Statistics
	syncStatsChainOrigin uint64 // Origin block number where syncing started at
	syncStatsChainHeight uint64 // Highest block number known when syncing started
	syncStatsState       stateSyncStats
	syncStatsLock        sync.RWMutex // Lock protecting the sync stats fields

	lightdag LightDag
	dag      BlockDag
	txpool   txspool.ITxPool
	// Callbacks
	dropPeer peerDropFn // Drops a peer for misbehaving

	// Status
	synchroniseMock func(id string, hash common.Hash) error // Replacement for synchronise during testing
	synchronising   int32
	notified        int32
	committed       int32

	// Channels
	headerCh   chan dataPack // [ptn/62] Channel receiving inbound block headers
	bodyCh     chan dataPack // [ptn/62] Channel receiving inbound block bodies
	receiptCh  chan dataPack // [ptn/63] Channel receiving inbound receipts
	bodyWakeCh chan bool     // [ptn/62] Channel to signal the block body fetcher of new tasks
	//receiptWakeCh chan bool              // [ptn/63] Channel to signal the receipt fetcher of new tasks
	headerProcCh chan []*modules.Header // [ptn/62] Channel to feed the header processor new tasks

	// for stateFetcher
	stateSyncStart chan *stateSync
	trackStateReq  chan *stateReq
	stateCh        chan dataPack // [ptn/63] Channel receiving inbound node state data

	// Cancellation and termination
	cancelPeer string         // Identifier of the peer currently being used as the master (cancel on drop)
	cancelCh   chan struct{}  // Channel to cancel mid-flight syncs
	cancelLock sync.RWMutex   // Lock to protect the cancel channel and peer in delivers
	cancelWg   sync.WaitGroup // Make sure all fetcher goroutines have exited.

	quitCh   chan struct{} // Quit channel to signal termination
	quitLock sync.RWMutex  // Lock to prevent double closes

	// Testing hooks
	syncInitHook     func(uint64, uint64)    // Method to call upon initiating a new sync run
	bodyFetchHook    func([]*modules.Header) // Method to call upon starting a block body fetch
	receiptFetchHook func([]*modules.Header) // Method to call upon starting a receipt fetch
	chainInsertHook  func([]*fetchResult)    // Method to call upon inserting a chain of blocks (possibly in multiple invocations)
}

// LightDag encapsulates functions required to synchronise a light chain.
type LightDag interface {
	HasHeader(common.Hash, uint64) bool
	GetHeaderByHash(common.Hash) *modules.Header
	CurrentHeader() *modules.Header
	InsertHeaderDag([]*modules.Header, int) (int, error)
	GetAllLeafNodes() ([]*modules.Header, error)
	//Rollback([]common.Hash)
}

// BlockDag encapsulates functions required to sync a (full or fast) dag.
type BlockDag interface {
	LightDag
	GetUnitByHash(common.Hash) (*modules.Unit, error)
	CurrentUnit() *modules.Unit
	FastSyncCommitHead(common.Hash) error
	//SaveDag(unit modules.Unit, isGenesis bool) (int, error)
	//InsertDag(modules.Units) (int, error)
	InsertDag(units modules.Units, txpool txspool.ITxPool) (int, error)

	//TODO :
	//LightDag
	//HasBlock(common.Hash, uint64) bool
	//GetBlockByHash(common.Hash) *types.Block
	//CurrentBlock() *types.Block
	//CurrentFastBlock() *types.Block
	//FastSyncCommitHead(common.Hash) error
	//InsertChain(types.Blocks) (int, error)
	//InsertReceiptChain(types.Blocks, []types.Receipts) (int, error)
}

// New creates a new downloader to fetch hashes and blocks from remote peers.
func New(mode SyncMode, mux *event.TypeMux, dropPeer peerDropFn, lightdag LightDag, dag BlockDag, txpool txspool.ITxPool) *Downloader {

	if lightdag == nil {
		lightdag = dag
	}

	dl := &Downloader{
		mode: mode,
		//levelDb:        levelDb,
		mux:           mux,
		queue:         newQueue(),
		peers:         newPeerSet(),
		rttEstimate:   uint64(rttMaxEstimate),
		rttConfidence: uint64(1000000),
		lightdag:      lightdag,
		dag:           dag,
		txpool:        txpool,
		dropPeer:      dropPeer,
		headerCh:      make(chan dataPack, 1),
		bodyCh:        make(chan dataPack, 1),
		receiptCh:     make(chan dataPack, 1),
		bodyWakeCh:    make(chan bool, 1),
		//receiptWakeCh:  make(chan bool, 1),
		headerProcCh:   make(chan []*modules.Header, 1),
		quitCh:         make(chan struct{}),
		stateCh:        make(chan dataPack),
		stateSyncStart: make(chan *stateSync),
		syncStatsState: stateSyncStats{
			processed: 0, //TODO must recover core.GetTrieSyncProgress(stateDb),
		},
		trackStateReq: make(chan *stateReq),
	}
	go dl.qosTuner()
	go dl.stateFetcher()
	return dl
}

// Progress retrieves the synchronisation boundaries, specifically the origin
// block where synchronisation started at (may have failed/suspended); the block
// or header sync is currently at; and the latest known block which the sync targets.
//
// In addition, during the state download phase of fast synchronisation the number
// of processed and the total number of known states are also returned. Otherwise
// these are zero.
func (d *Downloader) Progress() palletone.SyncProgress {
	// Lock the current stats and return the progress
	d.syncStatsLock.RLock()
	defer d.syncStatsLock.RUnlock()

	current := uint64(0)
	switch d.mode {
	case FullSync:
		unit := d.dag.CurrentUnit()
		if unit != nil {
			current = unit.Number().Index
		}
	case FastSync:
		unit := d.dag.CurrentUnit()
		if unit != nil {
			current = unit.Number().Index
		}
		//case LightSync:
		//current = d.lightdag.CurrentHeader().Number.Uint64()
	}

	return palletone.SyncProgress{
		StartingBlock: d.syncStatsChainOrigin,
		CurrentBlock:  current,
		HighestBlock:  d.syncStatsChainHeight,
		PulledStates:  d.syncStatsState.processed,
		KnownStates:   d.syncStatsState.processed + d.syncStatsState.pending,
	}
}

// Synchronising returns whether the downloader is currently retrieving blocks.
func (d *Downloader) Synchronising() bool {
	return atomic.LoadInt32(&d.synchronising) > 0
}

// RegisterPeer injects a new download peer into the set of block source to be
// used for fetching hashes and blocks from.
func (d *Downloader) RegisterPeer(id string, version int, peer Peer) error {
	logger := log.New("peer", id)
	logger.Trace("Registering sync peer")
	if err := d.peers.Register(newPeerConnection(id, version, peer, *logger)); err != nil {
		logger.Error("Failed to register sync peer", "err", err)
		return err
	}
	d.qosReduceConfidence()

	return nil
}

// RegisterLightPeer injects a light client peer, wrapping it so it appears as a regular peer.
func (d *Downloader) RegisterLightPeer(id string, version int, peer LightPeer) error {
	return d.RegisterPeer(id, version, &lightPeerWrapper{peer})
}

// UnregisterPeer remove a peer from the known list, preventing any action from
// the specified peer. An effort is also made to return any pending fetches into
// the queue.
func (d *Downloader) UnregisterPeer(id string) error {
	// Unregister the peer from the active peer set and revoke any fetch tasks
	logger := log.New("peer", id)
	logger.Trace("Unregistering sync peer")
	if err := d.peers.Unregister(id); err != nil {
		logger.Error("Failed to unregister sync peer", "err", err)
		return err
	}
	d.queue.Revoke(id)

	// If this peer was the master peer, abort sync immediately
	d.cancelLock.RLock()
	master := id == d.cancelPeer
	d.cancelLock.RUnlock()

	if master {
		d.cancel()
	}
	return nil
}

// Synchronise tries to sync up our local block chain with a remote peer, both
// adding various sanity checks as well as wrapping it with various log entries.
func (d *Downloader) Synchronise(id string, head common.Hash, index uint64, mode SyncMode, assetId modules.IDType16) error {
	//return nil
	err := d.synchronise(id, head, index, mode, assetId)
	switch err {
	case nil:
	case errBusy:

	case errTimeout, errBadPeer, errStallingPeer,
		errEmptyHeaderSet, errPeersUnavailable, errTooOld,
		errInvalidAncestor, errInvalidChain:
		log.Warn("Synchronisation failed, dropping peer", "peer", id, "err", err)
		if d.dropPeer == nil {
			// The dropPeer method is nil when `--copydb` is used for a local copy.
			// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
			log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", id)
		} else {
			d.dropPeer(id)
		}

	default:
		log.Warn("Synchronisation failed, retrying", "err", err)
	}
	return err
}

// synchronise will select the peer and use it for synchronising. If an empty string is given
// it will use the best peer possible and synchronize if its TD is higher than our own. If any of the
// checks fail an error will be returned. This method is synchronous
func (d *Downloader) synchronise(id string, hash common.Hash, index uint64, mode SyncMode, assetId modules.IDType16) error {
	log.Info("Enter Downloader synchronise", "peer id:", id)
	defer log.Info("End Downloader synchronise", "peer id:", id)
	// Mock out the synchronisation if testing
	if d.synchroniseMock != nil {
		return d.synchroniseMock(id, hash)
	}
	log.Debug("Downloader->synchronise", "d.synchronising:", d.synchronising)
	// Make sure only one goroutine is ever allowed past this point at once
	if !atomic.CompareAndSwapInt32(&d.synchronising, 0, 1) {
		log.Debug("Downloader->synchronise is Busy")
		return errBusy
	}
	defer atomic.StoreInt32(&d.synchronising, 0)

	// Post a user notification of the sync (only once per session)
	if atomic.CompareAndSwapInt32(&d.notified, 0, 1) {
		log.Info("Downloader synchronisation started")
	}
	// Reset the queue, peer set and wake channels to clean any internal leftover state
	d.queue.Reset()
	d.peers.Reset()

	for _, ch := range []chan bool{d.bodyWakeCh} {
		select {
		case <-ch:
		default:
		}
	}

	for _, ch := range []chan dataPack{d.headerCh, d.bodyCh} {
		for empty := false; !empty; {
			select {
			case <-ch:
			default:
				empty = true
			}
		}
	}
	for empty := false; !empty; {
		select {
		case <-d.headerProcCh:
		default:
			empty = true
		}
	}
	// Create cancel channel for aborting mid-flight and mark the master peer
	d.cancelLock.Lock()
	d.cancelCh = make(chan struct{})
	d.cancelPeer = id
	d.cancelLock.Unlock()

	defer d.Cancel() // No matter what, we can't leave the cancel channel open

	// Set the requested sync mode, unless it's forbidden
	d.mode = mode

	// Retrieve the origin peer and initiate the downloading process
	p := d.peers.Peer(id)
	if p == nil {
		return errUnknownPeer
	}
	return d.syncWithPeer(p, hash, index, assetId)
}

// syncWithPeer starts a block synchronization based on the hash chain from the
// specified peer and head hash.
func (d *Downloader) syncWithPeer(p *peerConnection, hash common.Hash, index uint64, assetId modules.IDType16) (err error) {
	d.mux.Post(StartEvent{})
	defer func() {
		// reset on error
		if err != nil {
			d.mux.Post(FailedEvent{err})
		} else {
			d.mux.Post(DoneEvent{})
		}
	}()

	if p.version < 1 {
		return errTooOld
	}

	log.Info("Synchronising with the network", "peer", p.id, "ptn", p.version, "head", hash, "index", index, "mode", d.mode)
	defer func(start time.Time) {
		log.Debug("Synchronisation terminated", "elapsed", time.Since(start), "peer", p.id)
	}(time.Now())

	// Look up the sync boundaries: the common ancestor and the target block
	latest, err := d.fetchHeight(p, assetId)
	if err != nil {
		log.Info("fetchHeight", "err:", err)
		return err
	}

	height := latest.Number.Index
	localIndex := d.dag.CurrentUnit().Number().Index

	log.Info("Downloader", "syncWithPeer local index", localIndex, "latest peer index", height)
	if localIndex >= height {
		return nil
	}

	origin, err := d.findAncestor(p, latest, assetId)
	if err != nil {
		return err
	}
	log.Info("=====findAncestor=====", "origin:", origin)

	// Ensure our origin point is below any fast sync pivot point
	pivot := uint64(0)
	//fmt.Println("Downloader->syncWithPeer pre", "height:", height, "origin:", origin, "pivot:", pivot)
	if d.mode == FastSync {
		if height <= uint64(fsMinFullBlocks) {
			origin = 0
		} else {
			pivot = height - uint64(fsMinFullBlocks)
			if pivot <= origin {
				origin = pivot - 1
			}
		}
	}
	//fmt.Println("Downloader->syncWithPeer last", "height:", height, "origin:", origin, "pivot:", pivot)
	log.Debug("Downloader->syncWithPeer last", "origin:", origin, "pivot:", pivot)
	d.committed = 1
	if d.mode == FastSync && pivot != 0 {
		d.committed = 0
	}
	// Initiate the sync using a concurrent header and content retrieval algorithm
	d.queue.Prepare(origin+1, d.mode)
	if d.syncInitHook != nil {
		d.syncInitHook(origin, height)
	}

	fetchers := []func() error{
		func() error { return d.fetchHeaders(p, origin+1, pivot, assetId) },
		func() error { return d.fetchBodies(origin+1, assetId) },
		func() error { return d.processHeaders(origin+1, pivot, index, assetId) },
	}
	if d.mode == FastSync {
		fetchers = append(fetchers, func() error { return d.processFastSyncContent(latest, assetId) })
	} else if d.mode == FullSync {
		fetchers = append(fetchers, d.processFullSyncContent)
	}
	return d.spawnSync(fetchers)
}

// spawnSync runs d.process and all given fetcher functions to completion in
// separate goroutines, returning the first error that appears.
//1 循环执行fetchers，fetchers是上面传过来的一组函数，
//fetch header、fetch body、 fetch receipt、process header等
//2 然后等待读取errc channel内容，等待sync完成。
//注意这里errc是一个缓冲channel，个数为fetchers的长度，就是会等待fetchers中的每个函数执行完成返回。
//所以这里实现了pending的效果，就是一直要等到sync完成才会结束sync
//3 如果fast sync的话，fetchers的最后一个函数是processFastSyncContent()；
//full sync模式下最后一个函数是processFullSyncContent()
func (d *Downloader) spawnSync(fetchers []func() error) error {
	errc := make(chan error, len(fetchers))
	d.cancelWg.Add(len(fetchers))
	for _, fn := range fetchers {
		fn := fn
		go func() { defer d.cancelWg.Done(); errc <- fn() }()
	}
	// Wait for the first error, then terminate the others.
	var err error
	for i := 0; i < len(fetchers); i++ {
		if i == len(fetchers)-1 {
			// Close the queue when all fetchers have exited.
			// This will cause the block processor to end when
			// it has processed the queue.
			d.queue.Close()
		}
		if err = <-errc; err != nil {
			break
		}
	}
	d.queue.Close()
	d.Cancel()
	return err
}

// cancel aborts all of the operations and resets the queue. However, cancel does
// not wait for the running download goroutines to finish. This method should be
// used when cancelling the downloads from inside the downloader.
func (d *Downloader) cancel() {
	// Close the current cancel channel
	d.cancelLock.Lock()
	if d.cancelCh != nil {
		select {
		case <-d.cancelCh:
			// Channel was already closed
		default:
			close(d.cancelCh)
		}
	}
	d.cancelLock.Unlock()
}

// Cancel aborts all of the operations and waits for all download goroutines to
// finish before returning.
func (d *Downloader) Cancel() {
	d.cancel()
	d.cancelWg.Wait()
}

// Terminate interrupts the downloader, canceling all pending operations.
// The downloader cannot be reused after calling Terminate.
func (d *Downloader) Terminate() {
	// Close the termination channel (make sure double close is allowed)
	d.quitLock.Lock()
	select {
	case <-d.quitCh:
	default:
		close(d.quitCh)
	}
	d.quitLock.Unlock()

	// Cancel any pending download requests
	d.Cancel()
}

// fetchHeight retrieves the head header of the remote peer to aid in estimating
// the total time a pending synchronisation would take.
func (d *Downloader) fetchHeight(p *peerConnection, assetId modules.IDType16) (*modules.Header, error) {
	log.Debug("Retrieving remote chain height", "peer", p.id)

	// Request the advertised remote head block and wait for the response
	//headerHash, number := p.peer.Head(assetId)
	//if common.EmptyHash(headerHash) && number.Index <= 0 {
	//	log.Debug("Downloader", "fetchHeight header hash:", headerHash, "number:", number.Index)
	//	return nil, errPeersUnavailable
	//}
	headerHash, _ := p.peer.Head(assetId)
	go p.peer.RequestHeadersByHash(headerHash, 1, 0, false)

	ttl := d.requestTTL()
	timeout := time.After(ttl)
	for {
		select {
		case <-d.cancelCh:
			return nil, errCancelBlockFetch

		case packet := <-d.headerCh:
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
			head := headers[0]
			log.Debug("Remote head header identified", "number", head.Number.Index, "hash", head.Hash(), "peer", packet.PeerId())
			return head, nil

		case <-timeout:
			log.Debug("Waiting for head header timed out", "elapsed", ttl, "peer", p.id)
			return nil, errTimeout

		case <-d.bodyCh:
		case <-d.receiptCh:
			// Out of bounds delivery, ignore
		}
	}
}

// findAncestor tries to locate the common ancestor link of the local chain and
// a remote peers blockchain. In the general case when our node was in sync and
// on the correct chain, checking the top N links should already get us a match.
// In the rare scenario when we ended up on a long reorganisation (i.e. none of
// the head links match), we do a binary search to find the common ancestor.

func (d *Downloader) findAncestor(p *peerConnection, latest *modules.Header, assetId modules.IDType16) (uint64, error) {
	height := latest.Index()
	// Figure out the valid ancestor range to prevent rewrite attacks
	floor, ceil := int64(-1), d.lightdag.CurrentHeader().Number.Index

	//if d.mode == FullSync {
	//	ceil = d.dag.CurrentUnit().NumberU64()
	//} else if d.mode == FastSync {
	//	ceil = d.dag.CurrentUnit().NumberU64()
	//}
	if ceil >= MaxForkAncestry {
		floor = int64(ceil - MaxForkAncestry)
	}
	p.log.Debug("Looking for common ancestor", "local", ceil, "remote", height)
	// Request the topmost blocks to short circuit binary ancestor lookup
	head := ceil
	if head > height {
		head = height
	}
	from := int64(head) - int64(MaxHeaderFetch)
	if from < 0 {
		from = 0
	}
	// Span out with 15 block gaps into the future to catch bad head reports
	limit := 2 * MaxHeaderFetch / 16
	count := 1 + int((int64(ceil)-from)/16)

	if count > limit {
		count = limit
	}
	log.Debug("Downloader", "findAncestor RequestHeadersByNumber false from:", from, "count:", count)
	index := modules.ChainIndex{
		AssetID: assetId,
		IsMain:  true,
		Index:   uint64(from),
	}

	go p.peer.RequestHeadersByNumber(index, count, 15, false)
	//TODO xiaozhi
	// Wait for the remote response to the head fetch
	number, hash := uint64(0), common.Hash{}

	ttl := d.requestTTL()
	timeout := time.After(ttl)

	for finished := false; !finished; {
		select {
		case <-d.cancelCh:
			return 0, errCancelHeaderFetch

		case packet := <-d.headerCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Debug("Received headers from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			headers := packet.(*headerPack).headers
			if len(headers) == 0 {
				p.log.Warn("Empty head header set")
				return 0, errEmptyHeaderSet
			}
			// Make sure the peer's reply conforms to the request
			for i := 0; i < len(headers); i++ {
				if number := headers[i].Number.Index; number != uint64(from+int64(i)*16) {
					p.log.Warn("Head headers broke chain ordering", "index", i, "requested", from+int64(i)*16, "received", number)
					return 0, errInvalidChain
				}
			}

			// Check if a common ancestor was found
			finished = true
			for i := len(headers) - 1; i >= 0; i-- {

				// Skip any headers that underflow/overflow our requested set
				if headers[i].Number.Index < uint64(from) || headers[i].Number.Index > ceil {
					continue
				}

				// Otherwise check if we already know the header or not
				if (d.mode == FullSync && d.dag.HasHeader(headers[i].Hash(), headers[i].Number.Index)) || (d.mode != FullSync && d.lightdag.HasHeader(headers[i].Hash(), headers[i].Number.Index)) {
					number, hash = headers[i].Number.Index, headers[i].Hash()
					// If every header is known, even future ones, the peer straight out lied about its head
					if number > height && i == limit-1 {
						p.log.Warn("Lied about chain head", "reported", height, "found", number)
						return 0, errStallingPeer
					}
					break
				}
			}

		case <-timeout:
			p.log.Debug("Waiting for head header timed out", "elapsed", ttl)
			return 0, errTimeout

		case <-d.bodyCh:
		case <-d.receiptCh:
			// Out of bounds delivery, ignore
		}
	}
	// If the head fetch already found an ancestor, return
	if !common.EmptyHash(hash) {
		if int64(number) <= floor {
			p.log.Warn("Ancestor below allowance", "number", number, "hash", hash, "allowance", floor)
			return 0, errInvalidAncestor
		}
		log.Debug("Found common ancestor", "number", number, "hash", hash)
		return number, nil
	}
	// Ancestor not found, we need to binary search over our chain
	start, end := uint64(0), head
	if floor > 0 {
		start = uint64(floor)
	}
	for start+1 < end {
		// Split our chain interval in two, and request the hash to cross check
		check := (start + end) / 2

		ttl := d.requestTTL()
		timeout := time.After(ttl)

		index.Index = check
		go p.peer.RequestHeadersByNumber(index, 1, 0, false)

		// Wait until a reply arrives to this request
		for arrived := false; !arrived; {
			select {
			case <-d.cancelCh:
				return 0, errCancelHeaderFetch

			case packer := <-d.headerCh:
				// Discard anything not from the origin peer
				if packer.PeerId() != p.id {
					log.Debug("Received headers from incorrect peer", "peer", packer.PeerId())
					break
				}
				// Make sure the peer actually gave something valid
				headers := packer.(*headerPack).headers
				if len(headers) != 1 {
					p.log.Debug("Multiple headers for single request", "headers", len(headers))
					return 0, errBadPeer
				}
				arrived = true

				// Modify the search interval based on the response
				if (d.mode == FullSync && !d.dag.HasHeader(headers[0].Hash(), headers[0].Number.Index)) || (d.mode != FullSync && !d.dag.HasHeader(headers[0].Hash(), headers[0].Number.Index)) {
					end = check
					break
				}
				header := d.dag.GetHeaderByHash(headers[0].Hash()) // Independent of sync mode, header surely exists
				if header.Number.Index != check {
					p.log.Debug("Received non requested header", "number", header.Number.Index, "hash", header.Hash(), "request", check)
					return 0, errBadPeer
				}
				start = check

			case <-timeout:
				p.log.Debug("Waiting for search header timed out", "elapsed", ttl)
				return 0, errTimeout

			case <-d.bodyCh:
			case <-d.receiptCh:
				// Out of bounds delivery, ignore
			}
		}
	}
	// Ensure valid ancestry and return
	if int64(start) <= floor {
		p.log.Warn("Ancestor below allowance", "start", start, "hash", hash, "allowance", floor)
		return 0, errInvalidAncestor
	}
	p.log.Debug("Found common ancestor", "number", start, "hash", hash)
	return start, nil
}

// fetchHeaders keeps retrieving headers concurrently from the number
// requested, until no more are returned, potentially throttling on the way. To
// facilitate concurrency but still protect against malicious nodes sending bad
// headers, we construct a header chain skeleton using the "origin" peer we are
// syncing with, and fill in the missing headers using anyone else. Headers from
// other peers are only accepted if they map cleanly to the skeleton. If no one
// can fill in the skeleton - not even the origin peer - it's assumed invalid and
// the origin is dropped.
//fetchHeaders不断的重复这样的操作，发送header请求，等待所有的返回。直到完成所有的header请求。
// 为了提高并发性，同时仍然能够防止恶意节点发送错误的header，
//我们使用我们正在同步的“origin”peer构造一个头文件链骨架，并使用其他人填充缺失的header。
//其他peer的header只有在干净地映射到骨架上时才被接受。
//如果没有人能够填充骨架 - 甚至origin peer也不能填充 - 它被认为是无效的，并且origin peer也被丢弃。
func (d *Downloader) fetchHeaders(p *peerConnection, from uint64, pivot uint64, assetId modules.IDType16) error {
	log.Debug("Directing header downloads", "origin", from)
	defer log.Debug("Header download terminated")

	// Create a timeout timer, and the associated header fetcher
	// 默认skeleton为true，表示先获取骨架（间隔的headers），然后再从其他节点填充骨架间的headers
	// （1 / K）^（N / K）  N = 2048，K = 100
	skeleton := true            // Skeleton assembly phase or finishing up
	request := time.Now()       // time of the last skeleton fetch request
	timeout := time.NewTimer(0) // timer to dump a non-responsive active peer
	<-timeout.C                 // timeout channel should be initially empty
	defer timeout.Stop()

	var ttl time.Duration
	getHeaders := func(from uint64, assetId modules.IDType16) {
		request = time.Now()

		ttl = d.requestTTL()
		timeout.Reset(ttl)

		index := modules.ChainIndex{
			AssetID: assetId,
			IsMain:  true,
		}

		if skeleton {
			index.Index = from + uint64(MaxHeaderFetch) - 1
			p.log.Trace("Fetching skeleton headers", "count", MaxSkeletonSize, "from", from, "index:", index.Index)
			go p.peer.RequestHeadersByNumber(index, MaxSkeletonSize, MaxHeaderFetch-1, false)
		} else {
			index.Index = from
			p.log.Trace("Fetching full headers", "count", MaxHeaderFetch, "from", from, "index:", index.Index)
			go p.peer.RequestHeadersByNumber(index, MaxHeaderFetch, 0, false)
		}
	}
	// Start pulling the header chain skeleton until all is done
	getHeaders(from, assetId)

	for {
		select {
		case <-d.cancelCh:
			return errCancelHeaderFetch

		case packet := <-d.headerCh:
			// Make sure the active peer is giving us the skeleton headers
			if packet.PeerId() != p.id {
				log.Debug("Received skeleton from incorrect peer", "peer", packet.PeerId())
				break
			}
			headerReqTimer.UpdateSince(request)
			timeout.Stop()

			// If the skeleton's finished, pull any remaining head headers directly from the origin
			if packet.Items() == 0 && skeleton {
				skeleton = false
				getHeaders(from, assetId)
				continue
			}
			// If no more headers are inbound, notify the content fetchers and return
			if packet.Items() == 0 {
				// Don't abort header fetches while the pivot is downloading
				if atomic.LoadInt32(&d.committed) == 0 && pivot <= from {
					p.log.Debug("No headers, waiting for pivot commit")
					select {
					case <-time.After(fsHeaderContCheck):
						getHeaders(from, assetId)
						continue
					case <-d.cancelCh:
						return errCancelHeaderFetch
					}
				}
				// Pivot done (or not in fast sync) and no more headers, terminate the process
				log.Debug("No more headers available")
				select {
				case d.headerProcCh <- nil:
					return nil
				case <-d.cancelCh:
					return errCancelHeaderFetch
				}
			}
			headers := packet.(*headerPack).headers

			// If we received a skeleton batch, resolve internals concurrently
			if skeleton {
				filled, proced, err := d.fillHeaderSkeleton(from, headers, assetId)
				if err != nil {
					p.log.Debug("Skeleton chain invalid", "err", err)
					return errInvalidChain
				}
				headers = filled[proced:]
				from += uint64(proced)
			}
			// Insert all the new headers and fetch the next batch
			if len(headers) > 0 {
				p.log.Trace("Scheduling new headers", "count", len(headers), "from", from)
				select {
				case d.headerProcCh <- headers:
				case <-d.cancelCh:
					return errCancelHeaderFetch
				}
				from += uint64(len(headers))
			}
			getHeaders(from, assetId)

		case <-timeout.C:
			if d.dropPeer == nil {
				// The dropPeer method is nil when `--copydb` is used for a local copy.
				// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
				p.log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", p.id)
				break
			}
			// Header retrieval timed out, consider the peer bad and drop
			p.log.Debug("Header request timed out", "elapsed", ttl)
			headerTimeoutMeter.Mark(1)
			d.dropPeer(p.id)

			// Finish the sync gracefully instead of dumping the gathered data though
			for _, ch := range []chan bool{d.bodyWakeCh} {
				select {
				case ch <- false:
				case <-d.cancelCh:
				}
			}
			select {
			case d.headerProcCh <- nil:
			case <-d.cancelCh:
			}
			return errBadPeer
		}
	}
}

// fillHeaderSkeleton concurrently retrieves headers from all our available peers
// and maps them to the provided skeleton header chain.
//
// Any partial results from the beginning of the skeleton is (if possible) forwarded
// immediately to the header processor to keep the rest of the pipeline full even
// in the case of header stalls.
//
// The method returns the entire filled skeleton and also the number of headers
// already forwarded for processing.
func (d *Downloader) fillHeaderSkeleton(from uint64, skeleton []*modules.Header, assetId modules.IDType16) ([]*modules.Header, int, error) {
	log.Debug("Filling up skeleton", "from", from, "len(skeleton):", len(skeleton))
	d.queue.ScheduleSkeleton(from, skeleton)

	var (
		deliver = func(packet dataPack) (int, error) {
			pack := packet.(*headerPack)
			return d.queue.DeliverHeaders(pack.peerId, pack.headers, d.headerProcCh)
		}
		expire   = func() map[string]int { return d.queue.ExpireHeaders(d.requestTTL()) }
		throttle = func() bool { return false }
		reserve  = func(p *peerConnection, count int) (*fetchRequest, bool, error) {
			return d.queue.ReserveHeaders(p, count), false, nil
		}
		fetch = func(p *peerConnection, req *fetchRequest) error {
			return p.FetchHeaders(req.From, MaxHeaderFetch, assetId)
		}
		capacity = func(p *peerConnection) int { return p.HeaderCapacity(d.requestRTT()) }
		setIdle  = func(p *peerConnection, accepted int) { p.SetHeadersIdle(accepted) }
	)
	err := d.fetchParts(errCancelHeaderFetch, d.headerCh, deliver, d.queue.headerContCh, expire,
		d.queue.PendingHeaders, d.queue.InFlightHeaders, throttle, reserve,
		nil, fetch, d.queue.CancelHeaders, capacity, d.peers.HeaderIdlePeers, setIdle, "headers")

	log.Debug("Skeleton fill terminated", "err", err)

	filled, proced := d.queue.RetrieveHeaders()
	return filled, proced, err
}

// fetchBodies iteratively downloads the scheduled block bodies, taking any
// available peers, reserving a chunk of blocks for each, waiting for delivery
// and also periodically checking for timeouts.
func (d *Downloader) fetchBodies(from uint64, assetId modules.IDType16) error {
	log.Debug("Downloading block bodies", "origin", from)
	var (
		deliver = func(packet dataPack) (int, error) {
			pack := packet.(*bodyPack)
			return d.queue.DeliverBodies(pack.peerId, pack.transactions /*, pack.uncles*/)
		}
		expire   = func() map[string]int { return d.queue.ExpireBodies(d.requestTTL()) }
		fetch    = func(p *peerConnection, req *fetchRequest) error { return p.FetchBodies(req) }
		capacity = func(p *peerConnection) int { return p.BlockCapacity(d.requestRTT()) }
		setIdle  = func(p *peerConnection, accepted int) { p.SetBodiesIdle(accepted) }
	)

	err := d.fetchParts(errCancelBodyFetch, d.bodyCh, deliver, d.bodyWakeCh, expire,
		d.queue.PendingBlocks, d.queue.InFlightBlocks, d.queue.ShouldThrottleBlocks, d.queue.ReserveBodies,
		d.bodyFetchHook, fetch, d.queue.CancelBodies, capacity, d.peers.BodyIdlePeers, setIdle, "bodies")

	log.Debug("Block body download terminated", "err", err)
	return err
}

// fetchParts iteratively downloads scheduled block parts, taking any available
// peers, reserving a chunk of fetch requests for each, waiting for delivery and
// also periodically checking for timeouts.
//
// As the scheduling/timeout logic mostly is the same for all downloaded data
// types, this method is used by each for data gathering and is instrumented with
// various callbacks to handle the slight differences between processing them.
//
// The instrumentation parameters:
//  - errCancel:   error type to return if the fetch operation is cancelled (mostly makes logging nicer)
//  - deliveryCh:  channel from which to retrieve downloaded data packets (merged from all concurrent peers)
//  - deliver:     processing callback to deliver data packets into type specific download queues (usually within `queue`)
//  - wakeCh:      notification channel for waking the fetcher when new tasks are available (or sync completed)
//  - expire:      task callback method to abort requests that took too long and return the faulty peers (traffic shaping)
//  - pending:     task callback for the number of requests still needing download (detect completion/non-completability)
//  - inFlight:    task callback for the number of in-progress requests (wait for all active downloads to finish)
//  - throttle:    task callback to check if the processing queue is full and activate throttling (bound memory use)
//  - reserve:     task callback to reserve new download tasks to a particular peer (also signals partial completions)
//  - fetchHook:   tester callback to notify of new tasks being initiated (allows testing the scheduling logic)
//  - fetch:       network callback to actually send a particular download request to a physical remote peer
//  - cancel:      task callback to abort an in-flight download request and allow rescheduling it (in case of lost peer)
//  - capacity:    network callback to retrieve the estimated type-specific bandwidth capacity of a peer (traffic shaping)
//  - idle:        network callback to retrieve the currently (type specific) idle peers that can be assigned tasks
//  - setIdle:     network callback to set a peer back to idle and update its estimated capacity (traffic shaping)
//  - kind:        textual label of the type being downloaded to display in log mesages
func (d *Downloader) fetchParts(errCancel error, deliveryCh chan dataPack, deliver func(dataPack) (int, error), wakeCh chan bool,
	expire func() map[string]int, pending func() int, inFlight func() bool, throttle func() bool, reserve func(*peerConnection, int) (*fetchRequest, bool, error),
	fetchHook func([]*modules.Header), fetch func(*peerConnection, *fetchRequest) error, cancel func(*fetchRequest), capacity func(*peerConnection) int,
	idle func() ([]*peerConnection, int), setIdle func(*peerConnection, int), kind string) error {

	// Create a ticker to detect expired retrieval tasks
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	update := make(chan struct{}, 1)
	// Prepare the queue and fetch block parts until the block header fetcher's done
	finished := false
	for {
		select {
		case <-d.cancelCh:
			return errCancel

		case packet := <-deliveryCh:
			// If the peer was previously banned and failed to deliver its pack
			// in a reasonable time frame, ignore its message.
			if peer := d.peers.Peer(packet.PeerId()); peer != nil {
				// Deliver the received chunk of data and check chain validity
				accepted, err := deliver(packet)
				if err == errInvalidChain {
					return err
				}
				// Unless a peer delivered something completely else than requested (usually
				// caused by a timed out request which came through in the end), set it to
				// idle. If the delivery's stale, the peer should have already been idled.
				if err != errStaleDelivery {
					setIdle(peer, accepted)
				}
				// Issue a log to the user to see what's going on
				switch {
				case err == nil && packet.Items() == 0:
					peer.log.Trace("Requested data not delivered", "type", kind)
				case err == nil:
					peer.log.Trace("Delivered new batch of data", "type", kind, "count", packet.Stats())
				default:
					peer.log.Trace("Failed to deliver retrieved data", "type", kind, "err", err)
				}
			}

			// Blocks assembled, try to update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case cont := <-wakeCh:
			// The header fetcher sent a continuation flag, check if it's done
			if !cont {
				finished = true
			}
			// Headers arrive, try to update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case <-ticker.C:
			// Sanity check update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case <-update:
			// Short circuit if we lost all our peers
			if d.peers.Len() == 0 {
				return errNoPeers
			}
			// Check for fetch request timeouts and demote the responsible peers
			for pid, fails := range expire() {
				if peer := d.peers.Peer(pid); peer != nil {
					// If a lot of retrieval elements expired, we might have overestimated the remote peer or perhaps
					// ourselves. Only reset to minimal throughput but don't drop just yet. If even the minimal times
					// out that sync wise we need to get rid of the peer.
					// The reason the minimum threshold is 2 is because the downloader tries to estimate the bandwidth
					// and latency of a peer separately, which requires pushing the measures capacity a bit and seeing
					// how response times reacts, to it always requests one more than the minimum (i.e. min 2).
					if fails > 2 {
						peer.log.Trace("Data delivery timed out", "type", kind)
						setIdle(peer, 0)
					} else {
						peer.log.Debug("Stalling delivery, dropping", "type", kind)
						if d.dropPeer == nil {
							// The dropPeer method is nil when `--copydb` is used for a local copy.
							// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
							peer.log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", pid)
						} else {
							d.dropPeer(pid)
						}
					}
				}
			}
			// If there's nothing more to fetch, wait or terminate
			if pending() == 0 {
				if !inFlight() && finished {
					log.Debug("Data fetching completed", "type", kind)
					return nil
				}
				break
			}
			// Send a download request to all idle peers, until throttled
			progressed, throttled, running := false, false, inFlight()
			idles, total := idle()

			for _, peer := range idles {
				// Short circuit if throttling activated
				if throttle() {
					throttled = true
					break
				}
				// Short circuit if there is no more available task.
				if pending() == 0 {
					break
				}
				// Reserve a chunk of fetches for a peer. A nil can mean either that
				// no more headers are available, or that the peer is known not to
				// have them.
				request, progress, err := reserve(peer, capacity(peer))
				if err != nil {
					log.Debug("Downloader fetchParts", "err:", err)
					return err
				}
				if progress {
					progressed = true
				}
				if request == nil {
					continue
				}
				if request.From > 0 {
					peer.log.Trace("Requesting new batch of data", "type", "peer id", peer.id, kind, "from", request.From)
				} else {
					peer.log.Trace("Requesting new batch of data", "type", "peer id", peer.id, kind, "count", len(request.Headers), "from", request.Headers[0].Number.Index)
				}
				// Fetch the chunk and make sure any errors return the hashes to the queue
				if fetchHook != nil {
					fetchHook(request.Headers)
				}
				if err := fetch(peer, request); err != nil {
					// Although we could try and make an attempt to fix this, this error really
					// means that we've double allocated a fetch task to a peer. If that is the
					// case, the internal state of the downloader and the queue is very wrong so
					// better hard crash and note the error instead of silently accumulating into
					// a much bigger issue.
					panic(fmt.Sprintf("%v: %s fetch assignment failed", peer, kind))
				}
				running = true
			}
			// Make sure that we have peers available for fetching. If all peers have been tried
			// and all failed throw an error
			if !progressed && !throttled && !running && len(idles) == total && pending() > 0 {
				return errPeersUnavailable
			}
		}
	}
}

// processHeaders takes batches of retrieved headers from an input channel and
// keeps processing and scheduling them into the header chain and downloader's
// queue until the stream ends or a failure occurs.
//从输入通道获取一批又一批的检索头，并将它们处理和调度到头链和下加载程序的队列中，直到流结束或发生故障。
func (d *Downloader) processHeaders(origin uint64, pivot uint64, index uint64, assetId modules.IDType16) error {
	log.Debug("===Enter processHeaders===", "d.mode:", d.mode)
	defer log.Debug("===End processHeaders===")
	// Keep a count of uncertain headers to roll back
	rollback := []*modules.Header{}
	defer func() {
		log.Debug("===processHeaders===", "len(rollback):", len(rollback))
		if len(rollback) > 0 {
			//TODO must recover
			/*
				// Flatten the headers and roll them back
				hashes := make([]common.Hash, len(rollback))
				for i, header := range rollback {
					hashes[i] = header.Hash()
				}
				lastHeader, lastFastBlock, lastBlock := d.lightdag.CurrentHeader().Number, common.Big0, common.Big0
				if d.mode != LightSync {
					lastFastBlock = d.dag.CurrentFastBlock().Number()
					lastBlock = d.dag.CurrentBlock().Number()
				}
				d.lightdag.Rollback(hashes)
				curFastBlock, curBlock := common.Big0, common.Big0
				if d.mode != LightSync {
					curFastBlock = d.dag.CurrentFastBlock().Number()
					curBlock = d.dag.CurrentBlock().Number()
				}
				log.Warn("Rolled back headers", "count", len(hashes),
					"header", fmt.Sprintf("%d->%d", lastHeader, d.lightdag.CurrentHeader().Number),
					"fast", fmt.Sprintf("%d->%d", lastFastBlock, curFastBlock),
					"block", fmt.Sprintf("%d->%d", lastBlock, curBlock))
			*/
		}
	}()

	// Wait for batches of headers to process
	gotHeaders := false

	for {
		select {
		case <-d.cancelCh:
			return errCancelHeaderProcessing

		case headers := <-d.headerProcCh:
			// Terminate header processing if we synced up
			if len(headers) == 0 {
				// Notify everyone that headers are fully processed
				for _, ch := range []chan bool{d.bodyWakeCh} {
					select {
					case ch <- false:
					case <-d.cancelCh:
					}
				}
				// If no headers were retrieved at all, the peer violated its TD promise that it had a
				// better chain compared to ours. The only exception is if its promised blocks were
				// already imported by other means (e.g. fecher):
				//
				// R <remote peer>, L <local node>: Both at block 10
				// R: Mine block 11, and propagate it to L
				// L: Queue block 11 for import
				// L: Notice that R's head and TD increased compared to ours, start sync
				// L: Import of block 11 finishes
				// L: Sync begins, and finds common ancestor at 11
				// L: Request new headers up from 11 (R's TD was higher, it must have something)
				// R: Nothing to give
				if d.mode != LightSync {
					head := d.dag.CurrentUnit()

					if !gotHeaders && index > d.dag.GetHeaderByHash(head.Hash()).Index() {
						return errStallingPeer
					}
				}
				// If fast or light syncing, ensure promised headers are indeed delivered. This is
				// needed to detect scenarios where an attacker feeds a bad pivot and then bails out
				// of delivering the post-pivot blocks that would flag the invalid content.
				//
				// This check cannot be executed "as is" for full imports, since blocks may still be
				// queued for processing when the header download completes. However, as long as the
				// peer gave us something useful, we're already happy/progressed (above check).
				//TODO whether or not recover
				//if d.mode == FastSync || d.mode == LightSync {
				//	head := d.lightdag.CurrentHeader()
				//	//if td.Cmp(d.lightdag.GetTd(head.Hash(), head.Number.Uint64())) > 0 {
				//	if index > d.lightdag.GetHeaderByHash(head.Hash()).Index() {
				//		log.Info("Downloader", "processHeaders index:", index, "dag index:", d.lightdag.GetHeaderByHash(head.Hash()).Index(),
				//			"dag head hash:", head.Hash())
				//		return errStallingPeer
				//	}
				//}
				// Disable any rollback and return
				rollback = nil
				return nil
			}
			// Otherwise split the chunk of headers into batches and process them
			gotHeaders = true

			for len(headers) > 0 {
				// Terminate if something failed in between processing chunks
				select {
				case <-d.cancelCh:
					return errCancelHeaderProcessing
				default:
				}
				// Select the next chunk of headers to import
				limit := maxHeadersProcess
				if limit > len(headers) {
					limit = len(headers)
				}
				chunk := headers[:limit]

				// In case of header only syncing, validate the chunk immediately
				if d.mode == FastSync || d.mode == LightSync {
					// Collect the yet unknown headers to mark them as uncertain
					unknown := make([]*modules.Header, 0, len(headers))
					for _, header := range chunk {
						if !d.lightdag.HasHeader(header.Hash(), header.Number.Index) {
							unknown = append(unknown, header)
						}
					}
					// If we're importing pure headers, verify based on their recentness
					//TODO Whether or not recover

					//frequency := fsHeaderCheckFrequency
					//if chunk[len(chunk)-1].Number.Index+uint64(fsHeaderForceVerify) > pivot {
					//	frequency = 1
					//}
					//if n, err := d.lightdag.InsertHeaderDag(chunk, frequency); err != nil {
					//	// If some headers were inserted, add them too to the rollback list
					//	if n > 0 {
					//		rollback = append(rollback, chunk[:n]...)
					//	}
					//	log.Debug("Invalid header encountered", "number", chunk[n].Number, "hash", chunk[n].Hash(), "err", err)
					//	return errInvalidChain
					//}
					// All verifications passed, store newly found uncertain headers
					rollback = append(rollback, unknown...)
					if len(rollback) > fsHeaderSafetyNet {
						rollback = append(rollback[:0], rollback[len(rollback)-fsHeaderSafetyNet:]...)
					}
				}
				// Unless we're doing light chains, schedule the headers for associated content retrieval
				if d.mode == FullSync || d.mode == FastSync {
					// If we've reached the allowed number of pending headers, stall a bit
					for d.queue.PendingBlocks() >= maxQueuedHeaders /* || d.queue.PendingReceipts() >= maxQueuedHeaders*/ {
						select {
						case <-d.cancelCh:
							return errCancelHeaderProcessing
						case <-time.After(time.Second):
						}
					}
					// Otherwise insert the headers for content retrieval
					inserts := d.queue.Schedule(chunk, origin)
					if len(inserts) != len(chunk) {
						log.Debug("Stale headers")
						return errBadPeer
					}
				}
				headers = headers[limit:]
				origin += uint64(limit)
			}

			// Update the highest block number we know if a higher one is found.
			d.syncStatsLock.Lock()
			if d.syncStatsChainHeight < origin {
				d.syncStatsChainHeight = origin - 1
			}
			d.syncStatsLock.Unlock()

			// Signal the content downloaders of the availablility of new tasks
			for _, ch := range []chan bool{d.bodyWakeCh} {
				select {
				case ch <- true:
				default:
				}
			}
		}
	}
	return nil
}

// processFullSyncContent takes fetch results from the queue and imports them into the chain.
func (d *Downloader) processFullSyncContent() error {
	for {
		results := d.queue.Results(true)
		if len(results) == 0 {
			return nil
		}
		if d.chainInsertHook != nil {
			d.chainInsertHook(results)
		}
		if err := d.importBlockResults(results); err != nil {
			return err
		}
	}
}

func (d *Downloader) importBlockResults(results []*fetchResult) error {
	// Check for any early termination requests
	log.Debug("Enter Downloader->importBlockResults", "len(results):", len(results))
	defer log.Debug("End Downloader->importBlockResults")
	if len(results) == 0 {
		return nil
	}
	select {
	case <-d.quitCh:
		return errCancelContentProcessing
	default:
	}
	// Retrieve the a batch of results to import
	first, last := results[0].Header, results[len(results)-1].Header
	log.Debug("Inserting downloaded chain", "items", len(results),
		"index", first.Number.Index, "index", last.Number.Index)

	blocks := make([]*modules.Unit, len(results))
	for i, result := range results {
		blocks[i] = modules.NewUnitWithHeader(result.Header).WithBody(result.Transactions)
	}
	for _, u := range blocks {
		log.Debug("======importBlockResults=======", "index:", u.UnitHeader.Number.Index, "unit:", *u)
		units := []*modules.Unit{}
		units = append(units, u)
		if index, err := d.dag.InsertDag(units, d.txpool); err != nil && err.Error() != dagerrors.ErrUnitExist.Error() {
			log.Debug("Downloaded item processing failed", "number", results[index].Header.Number.Index, "hash", results[index].Header.Hash(), "err", err)
			return errInvalidChain
		}
	}
	//if index, err := d.dag.InsertDag(blocks); err != nil {
	//	log.Debug("Downloaded item processing failed", "number", results[index].Header.Number, "hash", results[index].Header.Hash(), "err", err)
	//	return errInvalidChain
	//}
	return nil
}

// processFastSyncContent takes fetch results from the queue and writes them to the database.
// It also controls the synchronisation of state nodes of the pivot block.
func (d *Downloader) processFastSyncContent(latest *modules.Header, assetId modules.IDType16) error {
	// Start syncing state of the reported head block. This should get us most of
	// the state of the pivot block.
	log.Debug("", "===Enter processFastSyncContent===latest.Number.Index:", latest.Number.Index)
	defer log.Debug("End processFastSyncContent")
	//TODO wangjiyou

	// Figure out the ideal pivot block. Note, that this goalpost may move if the
	// sync takes long enough for the chain head to move significantly.
	pivot := uint64(0)
	if height := latest.Number.Index; height > uint64(fsMinFullBlocks) {
		//fmt.Println("========================111111111111===============================height:", height)
		pivot = height - uint64(fsMinFullBlocks)
	}
	// To cater for moving pivot points, track the pivot block and subsequently
	// accumulated download results separately.
	var (
		oldPivot *fetchResult   // Locked in pivot block, might change eventually
		oldTail  []*fetchResult // Downloaded content after the pivot
	)
	for {
		// Wait for the next batch of downloaded data to be available, and if the pivot
		// block became stale, move the goalpost
		results := d.queue.Results(oldPivot == nil) // Block if we're not monitoring pivot staleness
		if len(results) == 0 {
			// If pivot sync is done, stop
			if oldPivot == nil {
				//fmt.Println("===processFastSyncContent===oldPivot == nil")
				return nil //stateSync.Cancel()
			}
			// If sync failed, stop
			select {
			case <-d.cancelCh:
				return errCancelBlockFetch //TODO must modify the real cancel reason   //stateSync.Cancel()
			default:
			}
		}
		if d.chainInsertHook != nil {
			d.chainInsertHook(results)
		}
		if oldPivot != nil {
			results = append(append([]*fetchResult{oldPivot}, oldTail...), results...)
		}

		// Split around the pivot block and process the two sides via fast/full sync
		if atomic.LoadInt32(&d.committed) == 0 {
			latest = results[len(results)-1].Header
			if height := latest.Number.Index; height > pivot+2*uint64(fsMinFullBlocks) {
				log.Warn("Pivot became stale, moving", "old", pivot, "new", height-uint64(fsMinFullBlocks))
				//fmt.Println("===========================2222222222============================")
				pivot = height - uint64(fsMinFullBlocks)
			}
		}
		//fmt.Println("splitAroundPivot pre", "pivot", pivot, "len(results):", len(results))
		P, beforeP, afterP := splitAroundPivot(pivot, results)
		//fmt.Println("splitAroundPivot last", "P", P, "len(beforeP):", len(beforeP), "len(afterP):", len(afterP))
		if err := d.commitFastSyncData(beforeP); err != nil {
			return err
		}

		if P != nil {
			// If new pivot block found, cancel old state retrieval and restart
			if oldPivot != P {
				oldPivot = P
			}

			// Wait for completion, occasionally checking for pivot staleness
			//exec := true
			select {
			case <-time.After(time.Millisecond):
				//fmt.Println("commitPivotBlock P index:", P.Header.Number.Index)
				if err := d.commitPivotBlock(P); err != nil {
					return err
				}
				oldPivot = nil
			case <-time.After(time.Second):
				//fmt.Println("time.After(time.Second)", "len(oldTail):", len(oldTail), "len(afterP):", len(afterP))
				oldTail = afterP
				continue
			}
			//fmt.Println("no time out")
		}
		//fmt.Println("importBlockResults", "len(afterP):", len(afterP))
		// Fast sync done, pivot commit done, full import
		if err := d.importBlockResults(afterP); err != nil {
			return err
		}
	}
}

func splitAroundPivot(pivot uint64, results []*fetchResult) (p *fetchResult, before, after []*fetchResult) {
	for _, result := range results {
		num := result.Header.Number.Index
		switch {
		case num < pivot:
			before = append(before, result)
		case num == pivot:
			p = result
		default:
			after = append(after, result)
		}
	}
	return p, before, after
}
func (d *Downloader) commitFastSyncData(results []*fetchResult /*, stateSync *stateSync*/) error {
	// Check for any early termination requests
	log.Debug("Enter commitFastSyncData", "len(results):", len(results))
	defer log.Debug("End commitFastSyncData")
	if len(results) == 0 {
		return nil
	}
	select {
	case <-d.quitCh:
		return errCancelContentProcessing
	default:
	}
	// Retrieve the a batch of results to import
	first, last := results[0].Header, results[len(results)-1].Header
	log.Debug("Inserting fast-sync blocks", "items", len(results),
		"firstnum", first.Number.Index, "lastnumn", last.Number.Index,
	)

	blocks := make(modules.Units, len(results))
	for i, result := range results {
		blocks[i] = modules.NewUnitWithHeader(result.Header).WithBody(result.Transactions)
	}
	for _, u := range blocks {
		//log.Debug("======commitFastSyncData=======", "index:", u.UnitHeader.Number.Index, "unit:", *u)
		units := []*modules.Unit{}
		units = append(units, u)
		if index, err := d.dag.InsertDag(units, d.txpool); err != nil && err.Error() != dagerrors.ErrUnitExist.Error() {
			log.Debug("Downloaded item processing failed", "number", results[index].Header.Number.Index, "hash", results[index].Header.Hash(), "err", err)
			return errInvalidChain
		}
	}

	//if index, err := d.dag.InsertDag(blocks); err != nil {
	//	log.Debug("Downloaded item processing failed", "number", results[index].Header.Number.Index, "hash", results[index].Header.Hash(), "err", err)
	//	return errInvalidChain
	//}
	return nil
}

func (d *Downloader) commitPivotBlock(result *fetchResult) error {
	log.Debug("Enter commitPivotBlock")
	defer log.Debug("End commitPivotBlock")
	block := modules.NewUnitWithHeader(result.Header).WithBody(result.Transactions)
	log.Debug("Committing fast sync pivot as new head", "index:", block.UnitHeader.Number.Index, "unit", *block)

	units := []*modules.Unit{}
	units = append(units, block)
	if _, err := d.dag.InsertDag(units, d.txpool); err != nil && err.Error() != dagerrors.ErrUnitExist.Error() {
		log.Debug("Downloaded item processing failed", "index:", block.UnitHeader.Number.Index, "err:", err)
		return errInvalidChain
	}
	atomic.StoreInt32(&d.committed, 1)
	return nil
}

// DeliverHeaders injects a new batch of block headers received from a remote
// node into the download schedule.
func (d *Downloader) DeliverHeaders(id string, headers []*modules.Header) (err error) {
	return d.deliver(id, d.headerCh, &headerPack{id, headers}, headerInMeter, headerDropMeter)
}

// DeliverBodies injects a new batch of block bodies received from a remote node.
func (d *Downloader) DeliverBodies(id string, transactions [][]*modules.Transaction /*, uncles [][]*modules.Header*/) (err error) {
	return d.deliver(id, d.bodyCh, &bodyPack{id, transactions /*, uncles*/}, bodyInMeter, bodyDropMeter)
}

// DeliverNodeData injects a new batch of node state data received from a remote node.
func (d *Downloader) DeliverNodeData(id string, data [][]byte) (err error) {
	return d.deliver(id, d.stateCh, &statePack{id, data}, stateInMeter, stateDropMeter)
}

// deliver injects a new batch of data received from a remote node.
func (d *Downloader) deliver(id string, destCh chan dataPack, packet dataPack, inMeter, dropMeter metrics.Meter) (err error) {
	// Update the delivery metrics for both good and failed deliveries
	inMeter.Mark(int64(packet.Items()))
	defer func() {
		if err != nil {
			dropMeter.Mark(int64(packet.Items()))
		}
	}()
	// Deliver or abort if the sync is canceled while queuing
	d.cancelLock.RLock()
	cancel := d.cancelCh
	d.cancelLock.RUnlock()
	if cancel == nil {
		return errNoSyncActive
	}
	select {
	case destCh <- packet:
		return nil
	case <-cancel:
		return errNoSyncActive
	}
}

// qosTuner is the quality of service tuning loop that occasionally gathers the
// peer latency statistics and updates the estimated request round trip time.
func (d *Downloader) qosTuner() {
	for {
		// Retrieve the current median RTT and integrate into the previoust target RTT
		rtt := time.Duration((1-qosTuningImpact)*float64(atomic.LoadUint64(&d.rttEstimate)) + qosTuningImpact*float64(d.peers.medianRTT()))
		atomic.StoreUint64(&d.rttEstimate, uint64(rtt))

		// A new RTT cycle passed, increase our confidence in the estimated RTT
		conf := atomic.LoadUint64(&d.rttConfidence)
		conf = conf + (1000000-conf)/2
		atomic.StoreUint64(&d.rttConfidence, conf)

		// Log the new QoS values and sleep until the next RTT
		log.Debug("Recalculated downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", d.requestTTL())
		select {
		case <-d.quitCh:
			return
		case <-time.After(rtt):
		}
	}
}

// qosReduceConfidence is meant to be called when a new peer joins the downloader's
// peer set, needing to reduce the confidence we have in out QoS estimates.
func (d *Downloader) qosReduceConfidence() {
	// If we have a single peer, confidence is always 1
	peers := uint64(d.peers.Len())
	if peers == 0 {
		// Ensure peer connectivity races don't catch us off guard
		return
	}
	if peers == 1 {
		atomic.StoreUint64(&d.rttConfidence, 1000000)
		return
	}
	// If we have a ton of peers, don't drop confidence)
	if peers >= uint64(qosConfidenceCap) {
		return
	}
	// Otherwise drop the confidence factor
	conf := atomic.LoadUint64(&d.rttConfidence) * (peers - 1) / peers
	if float64(conf)/1000000 < rttMinConfidence {
		conf = uint64(rttMinConfidence * 1000000)
	}
	atomic.StoreUint64(&d.rttConfidence, conf)

	rtt := time.Duration(atomic.LoadUint64(&d.rttEstimate))
	log.Debug("Relaxed downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", d.requestTTL())
}

// requestRTT returns the current target round trip time for a download request
// to complete in.
//
// Note, the returned RTT is .9 of the actually estimated RTT. The reason is that
// the downloader tries to adapt queries to the RTT, so multiple RTT values can
// be adapted to, but smaller ones are preferred (stabler download stream).
func (d *Downloader) requestRTT() time.Duration {
	return time.Duration(atomic.LoadUint64(&d.rttEstimate)) * 9 / 10
}

// requestTTL returns the current timeout allowance for a single download request
// to finish under.
func (d *Downloader) requestTTL() time.Duration {
	var (
		rtt  = time.Duration(atomic.LoadUint64(&d.rttEstimate))
		conf = float64(atomic.LoadUint64(&d.rttConfidence)) / 1000000.0
	)
	ttl := time.Duration(ttlScaling) * time.Duration(float64(rtt)/conf)
	if ttl > ttlLimit {
		ttl = ttlLimit
	}
	return ttl
}

func (d *Downloader) getMaxNodes(headers []*modules.Header, assetId modules.IDType16) (*modules.Header, error) {
	size := len(headers)
	if size == 0 {
		return nil, nil
	}
	if size == 1 {
		return headers[0], nil
	}

	maxHeader := modules.Header{}
	for _, header := range headers {
		if assetId == header.Number.AssetID && header.Number.Index > maxHeader.Number.Index {
			maxHeader = *header
		}
	}
	return &maxHeader, nil
}

/*TODO must save
//fmt.Println("findAncestor===")
//fmt.Println("local=", ceil)
//fmt.Println("remote=", height)
//floor, ceil := uint64(0), uint64(0)
//TODO xiaozhi
//headers, err := d.lightdag.GetAllLeafNodes()
//if err != nil {
//	log.Info("===findAncestor===", "GetAllLeafNodes err:", err)
//	return floor, nil
//}
//header, err := d.getMaxNodes(headers, assetId)
//
//if err != nil {
//	log.Info("===findAncestor===", "getMaxNodes err:", err)
//	return floor, err
//}
////TODO xiaozhi

//if header != nil {
//	ceil = header.Number.Index
//	p.log.Debug("Looking for common ancestor", "local assetid", header.Number.AssetID.String(), "local index", ceil, "remote", latest.Number.Index)
//} else {
//	ceil = 0
//	p.log.Debug("Looking for common ancestor", "local index", ceil, "remote", latest.Number.Index)
//}
*/
