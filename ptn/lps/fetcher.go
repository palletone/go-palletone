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

// Package les implements the Light Ethereum Subprotocol.
package lps

import (
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	blockDelayTimeout  = time.Second * 10 // timeout for a peer to announce a head that has already been confirmed by others
	maxNodeCount       = 20               // maximum number of fetcherTreeNode entries remembered for each peer
	retryQueue         = time.Millisecond * 100
	softRequestTimeout = time.Millisecond * 500
	hardRequestTimeout = time.Second * 10
)

// headerRequesterFn is a callback type for sending a header retrieval request.
type headerRequesterFn func(common.Hash) error

// headerVerifierFn is a callback type to verify a block's header for fast propagation.
type headerVerifierFn func(header *modules.Header) error

// blockBroadcasterFn is a callback type for broadcasting a block to connected peers.
type headerBroadcasterFn func(block *modules.Unit, propagate bool)

// chainHeightFn is a callback type to retrieve the current chain height.
//type chainHeightFn func(assetId modules.IDType16) uint64

// chainInsertFn is a callback type to insert a batch of blocks into the local chain.
type headerInsertFn func(modules.Units) (int, error)

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
	origin string
	header *modules.Header
}

// announce is the hash notification of the availability of a new block in the
// network.
type announce struct {
	hash   common.Hash         // Hash of the block being announced
	number *modules.ChainIndex /*uint64*/ // Number of the block being announced (0 = unknown | old protocol)
	header *modules.Header     // Header of the block partially reassembled (new protocol)
	time   time.Time           // Timestamp of the announcement

	origin string // Identifier of the peer originating the notification

	fetchHeader headerRequesterFn // Fetcher function to retrieve the header of an announced block
	//fetchBodies bodyRequesterFn   // Fetcher function to retrieve the body of an announced block
}

// Fetcher is responsible for accumulating block announcements from various peers
// and scheduling them for retrieval.
type lightFetcher struct {
	// Various event channels
	inject chan *inject

	//blockFilter  chan chan []*modules.Unit
	headerFilter chan chan *headerFilterTask

	done chan common.Hash
	quit chan struct{}

	// Announce states
	announces  map[string]int              // Per peer announce counts to prevent memory exhaustion
	announced  map[common.Hash][]*announce // Announced blocks, scheduled for fetching
	fetching   map[common.Hash]*announce   // Announced blocks, currently fetching
	fetched    map[common.Hash][]*announce // Blocks with headers fetched, scheduled for body retrieval
	completing map[common.Hash]*announce   // Blocks with headers, currently body-completing

	// Block cache
	queue  *prque.Prque            // Queue containing the import operations (block number sorted)
	queues map[string]int          // Per peer block counts to prevent memory exhaustion
	queued map[common.Hash]*inject // Set of already queued blocks (to dedupe imports)

	// Callbacks
	verifyHeader    headerVerifierFn    // Checks if a block's headers have a valid proof of work
	broadcastHeader headerBroadcasterFn // Broadcasts a block to connected peers
	insertHeader    headerInsertFn      // Injects a batch of blocks into the chain
	dropPeer        peerDropFn          // Drops a peer for misbehaving
	//getBlock       blockRetrievalFn   // Retrieves a block from the local chain
	//chainHeight     chainHeightFn       // Retrieves the current chain's height
}
