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
	//"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"net"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header

	ethVersion = 63 // equivalent eth version for the downloader

	MaxHeaderFetch           = 192 // Amount of block headers to be fetched per retrieval request
	MaxBodyFetch             = 32  // Amount of block bodies to be fetched per retrieval request
	MaxReceiptFetch          = 128 // Amount of transaction receipts to allow fetching per request
	MaxCodeFetch             = 64  // Amount of contract codes to allow fetching per request
	MaxProofsFetch           = 64  // Amount of merkle proofs to be fetched per retrieval request
	MaxHelperTrieProofsFetch = 64  // Amount of merkle proofs to be fetched per retrieval request
	MaxTxSend                = 64  // Amount of transactions to be send per request
	MaxTxStatus              = 256 // Amount of transactions to queried per request

	disableClientRemovePeer = false
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type BlockChain interface {
	//Config() *params.ChainConfig
	HasHeader(hash common.Hash, number uint64) bool
	GetHeader(hash common.Hash, number uint64) *modules.Header
	GetHeaderByHash(hash common.Hash) *modules.Header
	CurrentHeader() *modules.Header
	GetTd(hash common.Hash, number uint64) *big.Int
	//State() (*state.StateDB, error)
	InsertHeaderChain(chain []*modules.Header, checkFreq int) (int, error)
	Rollback(chain []common.Hash)
	GetHeaderByNumber(number uint64) *modules.Header
	GetBlockHashesFromHash(hash common.Hash, max uint64) []common.Hash
	//Genesis() *types.Block
	SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription
}

type txPool interface {
	AddRemotes(txs []*modules.Transaction) []error
	//Status(hashes []common.Hash) []core.TxStatus
}

type ProtocolManager struct {
	lightSync bool
	txpool    txPool
	//txrelay     *LesTxRelay
	networkId uint64
	//chainConfig *params.ChainConfig
	dag     dag.IDag
	assetId modules.AssetId
	//chainDb     ethdb.Database
	odr        *LesOdr
	server     *LesServer
	serverPool *serverPool
	genesis    *modules.Unit
	//lesTopic   discv5.Topic
	reqDist   *requestDistributor
	retriever *retrieveManager

	downloader *downloader.Downloader
	fetcher    *lightFetcher
	peers      *peerSet
	maxPeers   int

	SubProtocols []p2p.Protocol

	eventMux *event.TypeMux

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg *sync.WaitGroup
}

// NewProtocolManager returns a new ethereum sub protocol manager. The Palletone sub protocol manages peers capable
// with the ethereum network.
func NewProtocolManager(lightSync bool, mode downloader.SyncMode, networkId uint64, gasToken modules.AssetId, txpool txPool,
	dag dag.IDag, mux *event.TypeMux, genesis *modules.Unit) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		lightSync: lightSync,
		eventMux:  mux,
		assetId:   gasToken,
		genesis:   genesis,
		//blockchain:  blockchain,
		//chainConfig: chainConfig,
		//chainDb:     chainDb,
		//odr:         odr,
		dag:       dag,
		networkId: networkId,
		txpool:    txpool,
		//txrelay:     txrelay,
		peers:     newPeerSet(),
		newPeerCh: make(chan *peer),
		//quitSync:    quitSync,
		wg:          new(sync.WaitGroup),
		noMorePeers: make(chan struct{}),
	}

	// Initiate a sub-protocol for every implemented version we can handle
	protocolVersions := ClientProtocolVersions
	manager.SubProtocols = make([]p2p.Protocol, 0, len(protocolVersions))
	for _, version := range protocolVersions {
		// Compatible, initialize the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    "lps",
			Version: version,
			Length:  ProtocolLengths[version],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				var entry *poolEntry
				peer := manager.newPeer(int(version), networkId, p, rw)
				if manager.serverPool != nil {
					addr := p.RemoteAddr().(*net.TCPAddr)
					entry = manager.serverPool.connect(peer, addr.IP, uint16(addr.Port))
				}
				peer.poolEntry = entry
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					err := manager.handle(peer)
					if entry != nil {
						manager.serverPool.disconnect(entry)
					}
					return err
				case <-manager.quitSync:
					if entry != nil {
						manager.serverPool.disconnect(entry)
					}
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}

	removePeer := manager.removePeer
	if disableClientRemovePeer {
		removePeer = func(id string) {}
	}

	if lightSync {
		manager.downloader = downloader.New(downloader.LightSync, manager.eventMux, removePeer, nil, dag, nil)
		manager.peers.notify((*downloaderPeerNotify)(manager))
		manager.fetcher = newLightFetcher(manager)
	}

	return manager, nil
}

// removePeer initiates disconnection from a peer by removing it from the peer set
func (pm *ProtocolManager) removePeer(id string) {
	pm.peers.Unregister(id)
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	if pm.lightSync {
		go pm.syncer()
	} else {
		go func() {
			for range pm.newPeerCh {
			}
		}()
	}
}

func (pm *ProtocolManager) Stop() {
	// Showing a log message. During download / process this could actually
	// take between 5 to 10 seconds and therefor feedback is required.
	log.Info("Stopping light Palletone protocol")

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	close(pm.quitSync) // quits syncer, fetcher

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for any process action
	pm.wg.Wait()

	log.Info("Light Palletone protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, nv uint64, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, nv, p, newMeteredMsgWriter(rw))
}

// handle is the callback invoked to manage the life cycle of a les peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}

	log.Debug("Light Palletone peer connected", "name", p.Name())

	// Execute the LES handshake
	//var (
	//	head   = pm.dag.CurrentHeader(pm.assetId)
	//	number = head.Number
	//	//td     = pm.blockchain.GetTd(hash, number)
	//)
	genesis, err := pm.dag.GetGenesisUnit()
	if err != nil {
		if err != nil {
			log.Error("Light PalletOne New", "get genesis err:", err)
			return err
		}
	}

	var (
		number   = &modules.ChainIndex{}
		headhash = common.Hash{}
	)
	if head := pm.dag.CurrentHeader(pm.assetId); head != nil {
		number = head.Number
		headhash = head.Hash()
	}
	if err := p.Handshake(number, genesis.Hash(), pm.server, headhash); err != nil {
		log.Debug("Light Palletone handshake failed", "err", err)
		return err
	}
	if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
		rw.Init(p.version)
	}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		log.Error("Light Palletone peer registration failed", "err", err)
		return err
	}
	defer func() {
		if pm.server != nil && pm.server.fcManager != nil && p.fcClient != nil {
			p.fcClient.Remove(pm.server.fcManager)
		}
		pm.removePeer(p.id)
	}()
	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	if pm.lightSync {
		p.lock.Lock()
		//head := p.headInfo
		p.lock.Unlock()
		if pm.fetcher != nil {
			//pm.fetcher.announce(p, head)
		}

		if p.poolEntry != nil {
			pm.serverPool.registered(p.poolEntry)
		}
	}

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		// new block announce loop
		for {
			select {
			case announce := <-p.announceChn:
				log.Debug("Light Palletone ProtocolManager->handle", "announce", announce)
				//data, err := rlp.EncodeToBytes(announce)
				//if err != nil {
				//	log.Error("rlp.EncodeToBytes", "err", err)
				//	return
				//}
				//log.Debug("Light Palletone ProtocolManager->handle", "announce bytes", data)
				//var req announceData
				//err = rlp.DecodeBytes(data, &req)
				//if err != nil {
				//	log.Error("rlp.DecodeBytes", "err", err)
				//	return
				//}
				p.SendAnnounce(announce)
			case <-stop:
				return
			}
		}
	}()

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			log.Debug("Light Palletone message handling failed", "err", err)
			return err
		}
	}
}

var reqList = []uint64{GetBlockHeadersMsg, GetBlockBodiesMsg, GetCodeMsg, GetReceiptsMsg, GetProofsV1Msg, SendTxMsg, SendTxV2Msg, GetTxStatusMsg, GetHeaderProofsMsg, GetProofsV2Msg, GetHelperTrieProofsMsg}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	log.Trace("Light Palletone message arrived", "code", msg.Code, "bytes", msg.Size)

	//costs := p.fcCosts[msg.Code]
	//reject := func(reqCnt, maxCnt uint64) bool {
	//	if p.fcClient == nil || reqCnt > maxCnt {
	//		return true
	//	}
	//	bufValue, _ := p.fcClient.AcceptRequest()
	//	cost := costs.baseCost + reqCnt*costs.reqCost
	//	if cost > pm.server.defParams.BufLimit {
	//		cost = pm.server.defParams.BufLimit
	//	}
	//	if cost > bufValue {
	//		recharge := time.Duration((cost - bufValue) * 1000000 / pm.server.defParams.MinRecharge)
	//		log.Error("Request came too early", "recharge", common.PrettyDuration(recharge))
	//		return true
	//	}
	//	return false
	//}

	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	//var deliverMsg *Msg

	// Handle the message depending on its contents
	switch msg.Code {
	case StatusMsg:
		log.Trace("Received status message")
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case AnnounceMsg:
		return pm.AnnounceMsg(msg, p)

	case GetBlockHeadersMsg:
		return pm.GetBlockHeadersMsg(msg, p)

	case BlockHeadersMsg:
		return pm.BlockHeadersMsg(msg, p)

	case GetBlockBodiesMsg:
		return nil //pm.GetBlockBodiesMsg(msg, p)

	case BlockBodiesMsg:
		return nil //pm.BlockBodiesMsg(msg, p)

	case GetCodeMsg:
		return nil //pm.GetCodeMsg(msg, p)

	case CodeMsg:
		return nil //pm.CodeMsg(msg, p)

	case GetProofsV1Msg:
		return nil //pm.GetProofsMsg(msg, p)

	case GetProofsV2Msg:
		log.Trace("Received les/2 proofs request")

	case ProofsV1Msg:
		return nil //pm.ProofsMsg(msg, p)

	case ProofsV2Msg:
		log.Trace("Received les/2 proofs response")

	case GetHeaderProofsMsg:
		return nil //pm.GetHeaderProofsMsg(msg, p)

	case GetHelperTrieProofsMsg:
		return nil //pm.GetHelperTrieProofsMsg(msg, p)

	case HeaderProofsMsg:
		return nil // pm.HeaderProofsMsg(msg, p)

	case HelperTrieProofsMsg:
		return nil //pm.HelperTrieProofsMsg(msg, p)

	case SendTxMsg:
		return nil //pm.SendTxMsg(msg, p)

	case SendTxV2Msg:

	case GetTxStatusMsg:
		return nil //pm.GetTxStatusMsg(msg, p)

	case TxStatusMsg:
		return nil //pm.TxStatusMsg(msg, p)

	default:
		log.Trace("Received unknown message", "code", msg.Code)
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}

	//if deliverMsg != nil {
	//	err := pm.retriever.deliver(p, deliverMsg)
	//	if err != nil {
	//		p.responseErrors++
	//		if p.responseErrors > maxResponseErrors {
	//			return err
	//		}
	//	}
	//}
	return nil
}

/*
// getAccount retrieves an account from the state based at root.
func (pm *ProtocolManager) getAccount(statedb *state.StateDB, root, hash common.Hash) (state.Account, error) {
	trie, err := trie.New(root, statedb.Database().TrieDB())
	if err != nil {
		return state.Account{}, err
	}
	blob, err := trie.TryGet(hash[:])
	if err != nil {
		return state.Account{}, err
	}
	var account state.Account
	if err = rlp.DecodeBytes(blob, &account); err != nil {
		return state.Account{}, err
	}
	return account, nil
}

// getHelperTrie returns the post-processed trie root for the given trie ID and section index
func (pm *ProtocolManager) getHelperTrie(id uint, idx uint64) (common.Hash, string) {
	switch id {
	case htCanonical:
		sectionHead := core.GetCanonicalHash(pm.chainDb, (idx+1)*light.CHTFrequencyClient-1)
		return light.GetChtV2Root(pm.chainDb, idx, sectionHead), light.ChtTablePrefix
	case htBloomBits:
		sectionHead := core.GetCanonicalHash(pm.chainDb, (idx+1)*light.BloomTrieFrequency-1)
		return light.GetBloomTrieRoot(pm.chainDb, idx, sectionHead), light.BloomTrieTablePrefix
	}
	return common.Hash{}, ""
}

// getHelperTrieAuxData returns requested auxiliary data for the given HelperTrie request
func (pm *ProtocolManager) getHelperTrieAuxData(req HelperTrieReq) []byte {
	switch {
	case req.Type == htCanonical && req.AuxReq == auxHeader && len(req.Key) == 8:
		blockNum := binary.BigEndian.Uint64(req.Key)
		hash := core.GetCanonicalHash(pm.chainDb, blockNum)
		return core.GetHeaderRLP(pm.chainDb, hash, blockNum)
	}
	return nil
}

func (pm *ProtocolManager) txStatus(hashes []common.Hash) []txStatus {
	stats := make([]txStatus, len(hashes))
	for i, stat := range pm.txpool.Status(hashes) {
		// Save the status we've got from the transaction pool
		stats[i].Status = stat

		// If the transaction is unknown to the pool, try looking it up locally
		if stat == core.TxStatusUnknown {
			if block, number, index := core.GetTxLookupEntry(pm.chainDb, hashes[i]); block != (common.Hash{}) {
				stats[i].Status = core.TxStatusIncluded
				stats[i].Lookup = &core.TxLookupEntry{BlockHash: block, BlockIndex: number, Index: index}
			}
		}
	}
	return stats
}
*/
// NodeInfo represents a short summary of the Palletone sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network uint64      `json:"network"` // Palletone network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Number  uint64      `json:"number"`  // Total difficulty of the host's blockchain
	Head    common.Hash `json:"head"`    // SHA3 hash of the host's best owned block
	//Genesis    common.Hash         `json:"genesis"`    // SHA3 hash of the host's genesis block
	//Config     *params.ChainConfig `json:"config"`     // Chain configuration for the fork rules
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (self *ProtocolManager) NodeInfo() *NodeInfo {
	//head := self.blockchain.CurrentHeader()
	//hash := head.Hash()

	return &NodeInfo{
		Network: self.networkId,
		Number:  uint64(0),
		//Genesis:    self.blockchain.Genesis().Hash(),
		//Config:     self.blockchain.Config(),
		//Head: hash,
	}
}

type downloaderPeerNotify ProtocolManager

type peerConnection struct {
	manager *ProtocolManager
	peer    *peer
}

func (pc *peerConnection) Head() (common.Hash, *big.Int) {
	return common.Hash{}, nil
	// return pc.peer.HeadAndTd()
}

func (pc *peerConnection) RequestHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	return nil
	//reqID := genReqID()
	//rq := &distReq{
	//	getCost: func(dp distPeer) uint64 {
	//		peer := dp.(*peer)
	//		return peer.GetRequestCost(GetBlockHeadersMsg, amount)
	//	},
	//	canSend: func(dp distPeer) bool {
	//		return dp.(*peer) == pc.peer
	//	},
	//	request: func(dp distPeer) func() {
	//		peer := dp.(*peer)
	//		cost := peer.GetRequestCost(GetBlockHeadersMsg, amount)
	//		peer.fcServer.QueueRequest(reqID, cost)
	//		return func() { peer.RequestHeadersByHash(reqID, cost, origin, amount, skip, reverse) }
	//	},
	//}
	//_, ok := <-pc.manager.reqDist.queue(rq)
	//if !ok {
	//	return ErrNoPeers
	//}
	//return nil
}

func (pc *peerConnection) RequestHeadersByNumber(origin uint64, amount int, skip int, reverse bool) error {
	//reqID := genReqID()
	//rq := &distReq{
	//	getCost: func(dp distPeer) uint64 {
	//		peer := dp.(*peer)
	//		return peer.GetRequestCost(GetBlockHeadersMsg, amount)
	//	},
	//	canSend: func(dp distPeer) bool {
	//		return dp.(*peer) == pc.peer
	//	},
	//	request: func(dp distPeer) func() {
	//		peer := dp.(*peer)
	//		cost := peer.GetRequestCost(GetBlockHeadersMsg, amount)
	//		peer.fcServer.QueueRequest(reqID, cost)
	//		return func() { peer.RequestHeadersByNumber(reqID, cost, origin, amount, skip, reverse) }
	//	},
	//}
	//_, ok := <-pc.manager.reqDist.queue(rq)
	//if !ok {
	//	return ErrNoPeers
	//}
	return nil
}

func (d *downloaderPeerNotify) registerPeer(p *peer) {
	//pm := (*ProtocolManager)(d)
	//pc := &peerConnection{
	//	manager: pm,
	//	peer:    p,
	//}
	//pm.downloader.RegisterLightPeer(p.id, ethVersion, pc)
}

func (d *downloaderPeerNotify) unregisterPeer(p *peer) {
	pm := (*ProtocolManager)(d)
	pm.downloader.UnregisterPeer(p.id)
}
