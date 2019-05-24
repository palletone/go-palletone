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

	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag"
	dagerrors "github.com/palletone/go-palletone/dag/errors"
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
	txChanSize              = 4096
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
	SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription
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
	fetcher    *LightFetcher
	peers      *peerSet
	maxPeers   int

	SubProtocols     []p2p.Protocol
	CorsSubProtocols []p2p.Protocol
	eventMux         *event.TypeMux

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg *sync.WaitGroup

	//SPV
	validation   *Validation
	utxosync     *UtxosSync
	protocolname string

	//cors
	corss *p2p.Server
}

// NewProtocolManager returns a new ethereum sub protocol manager. The Palletone sub protocol manages peers capable
// with the ethereum network.
func NewProtocolManager(lightSync bool, peers *peerSet, networkId uint64, gasToken modules.AssetId, txpool txPool,
	dag dag.IDag, mux *event.TypeMux, genesis *modules.Unit, quitSync chan struct{}, protocolname string) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		lightSync:    lightSync,
		eventMux:     mux,
		assetId:      gasToken,
		genesis:      genesis,
		quitSync:     quitSync,
		dag:          dag,
		networkId:    networkId,
		txpool:       txpool,
		protocolname: protocolname,
		peers:        peers,
		newPeerCh:    make(chan *peer),
		wg:           new(sync.WaitGroup),
		noMorePeers:  make(chan struct{}),
		validation:   NewValidation(dag),
		utxosync:     NewUtxosSync(dag),
	}

	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p.Protocol, 0, len(ClientProtocolVersions))
	for _, version := range ClientProtocolVersions {
		// Compatible, initialize the sub-protocol
		//version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    protocolname,
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
				return manager.NodeInfo(genesis.Hash())
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info(manager.assetId)
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

	if manager.lightSync {
		manager.downloader = downloader.New(downloader.LightSync, manager.eventMux, removePeer, nil, dag, nil)
		manager.peers.notify((*downloaderPeerNotify)(manager))
		manager.fetcher = manager.newLightFetcher()
	}

	return manager, nil
}

func (pm *ProtocolManager) newLightFetcher() *LightFetcher {
	headerVerifierFn := func(header *modules.Header) error {
		//hash := header.Hash()
		//log.Debugf("Importing propagated block insert DAG Enter ValidateUnitExceptGroupSig, unit: %s", hash.String())
		//defer log.Debugf("Importing propagated block insert DAG End ValidateUnitExceptGroupSig, unit: %s", hash.String())
		//verr := pm.dag.ValidateUnitExceptGroupSig(unit)
		//if verr != nil && !validator.IsOrphanError(verr) {
		//	return dagerrors.ErrFutureBlock
		//}
		//TODO must modify
		return dagerrors.ErrFutureBlock
	}
	headerBroadcaster := func(header *modules.Header, propagate bool) {
		log.Info("ProtocolManager headerBroadcaster", "hash:", header.Hash().String())
		pm.BroadcastLightHeader(header)
	}
	inserter := func(headers []*modules.Header) (int, error) {
		// If fast sync is running, deny importing weird blocks
		//TODO must add lock
		//if pm.lightSync {
		//	log.Warn("Discarded lighting sync propagated block", "number", headers[0].Number.Index, "hash", headers[0].Hash())
		//	return 0, errors.New("fasting sync")
		//}
		log.Debug("light Fetcher", "manager.dag.InsertDag index:", headers[0].Number.Index, "hash", headers[0].Hash())
		return pm.dag.InsertLightHeader(headers)
	}
	return NewLightFetcher(pm.dag.GetHeaderByHash, pm.dag.GetLightChainHeight, headerVerifierFn,
		headerBroadcaster, inserter, pm.removePeer)
}

func (pm *ProtocolManager) BroadcastLightHeader(header *modules.Header) {
	log.Info("ProtocolManager", "BroadcastLightHeader index:", header.Index(), "sub protocal name:", header.Number.AssetID.String())
	peers := pm.peers.PeersWithoutHeader(header.Number.AssetID, header.Hash())
	announce := announceData{Hash: header.Hash(), Number: *header.Number, Header: *header}
	for _, p := range peers {
		if p == nil {
			continue
		}
		log.Debug("Light Palletone", "BroadcastLightHeader announceType", p.announceType)
		switch p.announceType {
		case announceTypeNone:
			select {
			case p.announceChn <- announce:
			default:
				pm.removePeer(p.id)
			}
		case announceTypeSimple:

		case announceTypeSigned:

		}
	}
	log.Trace("BroadcastLightHeader Propagated header", "protocalname", pm.SubProtocols[0].Name, "index:", header.Number.Index, "hash", header.Hash(), "recipients", len(peers))
	return
}

// removePeer initiates disconnection from a peer by removing it from the peer set
func (pm *ProtocolManager) removePeer(id string) {
	pm.peers.Unregister(id)
}

func (pm *ProtocolManager) Start(maxPeers int, corss *p2p.Server) {
	pm.maxPeers = maxPeers

	if pm.lightSync {
		go pm.syncer()
		pm.validation.Start()

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
	pm.validation.Stop()

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

	genesis, err := pm.dag.GetGenesisUnit()
	if err != nil {
		log.Error("Light PalletOne New", "get genesis err:", err)
		return err
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
		//p.lock.Lock()
		//head := p.headInfo
		//p.lock.Unlock()
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
				data, err := json.Marshal(announce.Header)
				if err != nil {
					log.Error("Light Palletone ProtocolManager->handle", "Marshal err", err, "announce", announce)
				} else {
					p.lightlock.Lock()
					//p.headInfo = &announce
					p.lightpeermsg[announce.Number.AssetID] = &announce
					p.lightlock.Unlock()
					if !p.fullnode {
						p.SendRawAnnounce(data)
					}
				}
			case <-stop:
				return
			}
		}
	}()

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			log.Debug("Light PalletOne message handling failed", "err", err)
			return err
		}
	}
}

var reqList = []uint64{GetBlockHeadersMsg, GetBlockBodiesMsg, GetCodeMsg, GetUTXOsMsg, GetProofsMsg, SendTxMsg, SendTxV2Msg, GetTxStatusMsg, GetHeaderProofsMsg, GetProofsV2Msg, GetHelperTrieProofsMsg}

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

	case GetProofsMsg:
		return pm.GetProofsMsg(msg, p)

	case GetProofsV2Msg:
		log.Trace("Received les/2 proofs request")

	case GetUTXOsMsg:
		log.Debug("Received les GetUTXOsMsg")
		return pm.GetUTXOsMsg(msg, p)

	case UTXOsMsg:
		log.Debug("Received les UTXOsMsg")
		return pm.UTXOsMsg(msg, p)

	case ProofsMsg:
		return pm.ProofsMsg(msg, p)

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
		return pm.SendTxMsg(msg, p)

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

func (pm *ProtocolManager) BroadcastTx(hash common.Hash, tx *modules.Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it
	peers := pm.peers.AllPeers(pm.assetId)
	//FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for _, peer := range peers {
		peer.SendTxs(0, 0, modules.Transactions{tx})
	}
	log.Trace("Broadcast transaction", "hash", hash, "recipients", len(peers))
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
	Index   uint64      `json:"Index"`   // Total difficulty of the host's blockchain
	Head    common.Hash `json:"head"`    // SHA3 hash of the host's best owned block
	Genesis common.Hash `json:"genesis"` // SHA3 hash of the host's genesis block
	//Config     *params.ChainConfig `json:"config"`     // Chain configuration for the fork rules
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (self *ProtocolManager) NodeInfo(genesisHash common.Hash) *NodeInfo {
	header := self.dag.CurrentHeader(self.assetId)

	var (
		index = uint64(0)
		hash  = common.Hash{}
	)
	if header != nil {
		index = header.Number.Index
		hash = header.Hash()
	} else {
		log.Debug("Light PalletOne NodeInfo header is nil")
	}

	return &NodeInfo{
		Network: self.networkId,
		Index:   index,
		Genesis: genesisHash,
		Head:    hash,
	}
}

type downloaderPeerNotify ProtocolManager

type peerConnection struct {
	manager *ProtocolManager
	peer    *peer
}

//Head(modules.AssetId) (common.Hash, *modules.ChainIndex)
//RequestHeadersByHash(common.Hash, int, int, bool) error
//RequestHeadersByNumber(*modules.ChainIndex, int, int, bool) error
//RequestDagHeadersByHash(common.Hash, int, int, bool) error
//RequestLeafNodes() error

func (pc *peerConnection) Head(assetId modules.AssetId) (common.Hash, *modules.ChainIndex) {
	//return common.Hash{}, nil
	return pc.peer.HeadAndNumber(assetId)
}

func (pc *peerConnection) RequestHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	log.Debug("peerConnection batch of headers by hash", "count", amount, "fromhash", origin, "skip", skip, "reverse", reverse)
	return p2p.Send(pc.peer.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Hash: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

func (pc *peerConnection) RequestHeadersByNumber(origin *modules.ChainIndex, amount int, skip int, reverse bool) error {
	log.Debug("peerConnection batch of headers by number", "count", amount, "from origin", origin, "skip", skip, "reverse", reverse)
	return p2p.Send(pc.peer.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Number: *origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}
func (p *peerConnection) RequestDagHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	//log.Debug("Fetching batch of headers", "count", amount, "fromhash", origin, "skip", skip, "reverse", reverse)
	return nil
}

func (p *peerConnection) RequestLeafNodes() error {
	//GetLeafNodes
	log.Debug("Fetching leaf nodes")
	return nil
	//return p2p.Send(p.rw, GetLeafNodesMsg, "")
}

func (d *downloaderPeerNotify) registerPeer(p *peer) {
	pm := (*ProtocolManager)(d)
	pc := &peerConnection{
		manager: pm,
		peer:    p,
	}
	pm.downloader.RegisterLightPeer(p.id, p.version, pc)
}

func (d *downloaderPeerNotify) unregisterPeer(p *peer) {
	pm := (*ProtocolManager)(d)
	pm.downloader.UnregisterPeer(p.id)
}
