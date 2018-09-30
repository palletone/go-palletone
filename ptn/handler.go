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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/common/rlp"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptn/fetcher"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header

	// txChanSize is the size of channel listening to TxPreEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
)

var (
	daoChallengeTimeout = 15 * time.Second // Time allowance for a node to reply to the DAO handshake challenge
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

//var tempGetBlockBodiesMsgSum int = 0

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type ProtocolManager struct {
	networkId uint64
	srvr      *p2p.Server

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txpool   txPool
	maxPeers int

	downloader *downloader.Downloader
	fetcher    *fetcher.Fetcher
	peers      *peerSet

	SubProtocols []p2p.Protocol

	eventMux *event.TypeMux
	txCh     chan modules.TxPreEvent
	txSub    event.Subscription

	dag dag.IDag

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	//consensus to p2p
	consEngine core.ConsensusEngine
	ceCh       chan core.ConsensusEvent
	ceSub      event.Subscription

	// append by Albert·Gou
	producer           producer
	newProducedUnitCh  chan mp.NewProducedUnitEvent
	newProducedUnitSub event.Subscription

	// append by Albert·Gou
	sigShareCh  chan mp.SigShareEvent
	sigShareSub event.Subscription

	// append by Albert·Gou
	vssDealCh  chan mp.VSSDealEvent
	vssDealSub event.Subscription

	vssResponseCh  chan mp.VSSResponseEvent
	vssResponseSub event.Subscription

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup

	genesis *modules.Unit

	peersTransition  *peerSet
	transCycleConnCh chan int
}

// NewProtocolManager returns a new PalletOne sub protocol manager. The PalletOne sub protocol manages peers capable
// with the PalletOne network.
func NewProtocolManager(mode downloader.SyncMode, networkId uint64, txpool txPool,
	engine core.ConsensusEngine, dag dag.IDag, mux *event.TypeMux, producer producer,
	genesis *modules.Unit) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkId:        networkId,
		dag:              dag,
		txpool:           txpool,
		eventMux:         mux,
		consEngine:       engine,
		peers:            newPeerSet(),
		newPeerCh:        make(chan *peer),
		noMorePeers:      make(chan struct{}),
		txsyncCh:         make(chan *txsync),
		quitSync:         make(chan struct{}),
		transCycleConnCh: make(chan int, 1),
		genesis:          genesis,
		producer:         producer,
		peersTransition:  newPeerSet(),
	}

	// Figure out whether to allow fast sync or not
	/*blockchain.CurrentBlock().NumberU64() > 0 */
	//TODO must modify.The second start would Blockchain not empty, fast sync disabled
	if mode == downloader.FastSync && dag.CurrentUnit().UnitHeader.Index() > 0 {
		log.Info("dag not empty, fast sync disabled")
		mode = downloader.FullSync
	}
	if mode == downloader.FastSync {
		manager.fastSync = uint32(1)
	}

	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p.Protocol, 0, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		// Skip protocol version if incompatible with the mode of operation
		if mode == downloader.FastSync && version < ptn1 {
			continue
		}
		// Compatible; initialise the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  ProtocolLengths[i],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := manager.newPeer(int(version), p, rw)
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					return manager.handle(peer)
				case <-manager.quitSync:
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				if p := manager.peers.Peer(id.TerminalString()); p != nil {
					return p.Info()
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}

	// Construct the different synchronisation mechanisms
	manager.downloader = downloader.New(mode, manager.eventMux, manager.removePeer, nil, dag)

	validator := func(header *modules.Header) error {
		//TODO must recover
		return nil //dag.VerifyHeader(header, false)
	}
	heighter := func(assetId modules.IDType16) uint64 {
		unit := dag.GetCurrentUnit(assetId)
		if unit != nil {
			return unit.NumberU64()
		}
		return uint64(0)
	}
	inserter := func(blocks modules.Units) (int, error) {
		// If fast sync is running, deny importing weird blocks
		if atomic.LoadUint32(&manager.fastSync) == 1 {
			log.Warn("Discarded bad propagated block", "number", blocks[0].Number(), "hash", blocks[0].Hash())
			return 0, nil
		}
		atomic.StoreUint32(&manager.acceptTxs, 1) // Mark initial sync done on any fetcher import
		return manager.dag.InsertDag(blocks)
	}
	manager.fetcher = fetcher.New(dag.GetUnitByHash, validator, manager.BroadcastUnit, heighter, inserter, manager.removePeer)
	return manager, nil
}

func (pm *ProtocolManager) removeTransitionPeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peersTransition.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing PalletOne peer", "peer", id)

	// Unregister the peer from the PalletOne peer set
	if err := pm.peersTransition.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing PalletOne peer", "peer", id)

	// Unregister the peer from the downloader and PalletOne peer set
	pm.downloader.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) Start(srvr *p2p.Server, maxPeers int) {
	pm.srvr = srvr
	pm.maxPeers = maxPeers

	pm.ceCh = make(chan core.ConsensusEvent, txChanSize)
	pm.ceSub = pm.consEngine.SubscribeCeEvent(pm.ceCh)
	go pm.ceBroadcastLoop()
	// start sync handlers
	//定时与相邻个体进行全链的强制同步,syncer()首先启动fetcher成员，然后进入一个无限循环，
	//每次循环中都会向相邻peer列表中“最优”的那个peer作一次区块全链同步
	go pm.syncer()

	//txsyncLoop负责把pending的交易发送给新建立的连接。
	//txsyncLoop负责每个新连接的初始事务同步。
	//当新的对等体出现时，我们转发所有当前待处理的事务。
	//为了最小化出口带宽使用，我们一次将一个小包中的事务发送给一个对等体。
	go pm.txsyncLoop()

	// broadcast transactions
	// 广播交易的通道。 txCh会作为txpool的TxPreEvent订阅通道。
	// txpool有了这种消息会通知给这个txCh。 广播交易的goroutine会把这个消息广播出去。
	pm.txCh = make(chan modules.TxPreEvent, txChanSize)
	// 订阅的回执
	pm.txSub = pm.txpool.SubscribeTxPreEvent(pm.txCh)
	// 启动广播的goroutine
	go pm.txBroadcastLoop()

	// append by Albert·Gou
	// broadcast new unit produced by mediator
	pm.newProducedUnitCh = make(chan mp.NewProducedUnitEvent)
	pm.newProducedUnitSub = pm.producer.SubscribeNewProducedUnitEvent(pm.newProducedUnitCh)
	go pm.newProducedUnitBroadcastLoop()

	// append by Albert·Gou
	// send signature share
	pm.sigShareCh = make(chan mp.SigShareEvent)
	pm.sigShareSub = pm.producer.SubscribeSigShareEvent(pm.sigShareCh)
	go pm.sigShareTransmitLoop()

	// append by Albert·Gou
	// send  VSS deal
	pm.vssDealCh = make(chan mp.VSSDealEvent)
	pm.vssDealSub = pm.producer.SubscribeVSSDealEvent(pm.vssDealCh)
	go pm.vssDealTransmitLoop()

	// append by Albert·Gou
	// broadcast  VSS Response
	pm.vssResponseCh = make(chan mp.VSSResponseEvent)
	pm.vssResponseSub = pm.producer.SubscribeVSSResponseEvent(pm.vssResponseCh)
	go pm.vssResponseBroadcastLoop()

	//TODO xiaozhi
	go pm.mediatorConnect()
}

// @author Albert·Gou
//func (pm *ProtocolManager) BroadcastVssResp(dstId string, resp *mp.VSSResponseEvent) {
func (pm *ProtocolManager) BroadcastVssResp(resp *mp.VSSResponseEvent) {
	// comment by Albert·Gou
	//dstId := node.ID.TerminalString()
	//peer := pm.peers.Peer(dstId)
	//if peer == nil {
	//	log.Error(fmt.Sprintf("peer not exist: %v", node.String()))
	//}

	// comment by Albert·Gou
	//if pm.peers.PeersWithoutVssResp(dstId) {
	//	return
	//}
	//pm.peers.MarkVssResp(dstId)

	peers := pm.GetTransitionPeers() //pm.GetActiveMediatorPeers()
	for _, peer := range peers {
		if peer == nil {
			//size, reader, err := rlp.EncodeToReader(resp)
			//if err != nil {
			//	log.Error(err.Error())
			//}
			//
			//var r mp.VSSResponseEvent
			//s := rlp.NewStream(reader, uint64(size))
			//if err := s.Decode(&r); err != nil {
			//	log.Error(err.Error())
			//}
			//pm.producer.ToProcessResponse(&r)

			pm.producer.ToProcessResponse(resp)
			continue
		}

		// comment by Albert·Gou
		//dstId := peer.id
		//if pm.peers.PeersWithoutVssResp(dstId) {
		//	return
		//}
		//pm.peers.MarkVssResp(dstId)

		// comment by Albert·Gou
		//msg := &vssRespMsg{
		//	NodeId: dstId,
		//	Resp:   resp,
		//}
		//
		//err := peer.SendVSSResponse(msg)

		err := peer.SendVSSResponse(resp)
		if err != nil {
			log.Info(err.Error())
		}
	}
}

// @author Albert·Gou
func (self *ProtocolManager) vssResponseBroadcastLoop() {
	for {
		select {
		case event := <-self.vssResponseCh:
			self.BroadcastVssResp(&event)

			// Err() channel will be closed when unsubscribing.
		case <-self.vssResponseSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) TransmitVSSDeal(node *discover.Node, deal *mp.VSSDealEvent) {
	peer, self := pm.getTransitionPeer(node)
	if self {
		//size, reader, err := rlp.EncodeToReader(deal)
		//if err != nil {
		//	log.Error(err.Error())
		//}
		//
		//var d mp.VSSDealEvent
		//s := rlp.NewStream(reader, uint64(size))
		//if err := s.Decode(&d); err != nil {
		//	log.Error(err.Error())
		//}
		//pm.producer.ToProcessDeal(&d)

		pm.producer.ToProcessDeal(deal)
		return
	}

	if peer == nil {
		return
	}

	// comment by Albert·Gou
	// // append by wangjiyou
	//if pm.peers.PeersWithoutVss(dstId) {
	//	return
	//}
	//pm.peers.MarkVss(dstId)

	//msg := &vssMsg{
	//	NodeId: dstId,
	//	Deal:   deal,
	//}
	//err := peer.SendVSSDeal(msg)

	err := peer.SendVSSDeal(deal)
	if err != nil {
		log.Error(err.Error())
	}
}

// @author Albert·Gou
func (self *ProtocolManager) vssDealTransmitLoop() {
	for {
		select {
		case event := <-self.vssDealCh:
			node := self.dag.GetActiveMediatorNode(event.DstIndex)
			self.TransmitVSSDeal(node, &event)

			// Err() channel will be closed when unsubscribing.
		case <-self.vssDealSub.Err():
			return
		}
	}
}

// @author Albert·Gou
type producer interface {
	// SubscribeNewProducedUnitEvent should return an event subscription of
	// NewProducedUnitEvent and send events to the given channel.
	SubscribeNewProducedUnitEvent(ch chan<- mp.NewProducedUnitEvent) event.Subscription
	// UnitBLSSign is to TBLS sign the unit
	ToUnitTBLSSign(unit *modules.Unit) error

	SubscribeSigShareEvent(ch chan<- mp.SigShareEvent) event.Subscription
	ToTBLSRecover(sigShare *mp.SigShareEvent) error

	SubscribeVSSDealEvent(ch chan<- mp.VSSDealEvent) event.Subscription
	ToProcessDeal(deal *mp.VSSDealEvent) error

	SubscribeVSSResponseEvent(ch chan<- mp.VSSResponseEvent) event.Subscription
	ToProcessResponse(resp *mp.VSSResponseEvent) error

	LocalHaveActiveMediator() bool
	StartVSSProtocol()
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping PalletOne protocol")

	// append by Albert·Gou
	pm.newProducedUnitSub.Unsubscribe()
	pm.sigShareSub.Unsubscribe()
	pm.vssDealSub.Unsubscribe()
	pm.vssResponseSub.Unsubscribe()

	pm.txSub.Unsubscribe() // quits txBroadcastLoop

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	//pm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Quit fetcher, txsyncLoop.
	close(pm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	log.Info("PalletOne protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, p, newMeteredMsgWriter(rw))
}

// handle is the callback invoked to manage the life cycle of an ptn peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	//TODO must modify make sure  have enough connections for mediators
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		log.Info("ProtocolManager", "handler DiscTooManyPeers:", p2p.DiscTooManyPeers)
		return p2p.DiscTooManyPeers
	}
	log.Debug("PalletOne peer connected", "name", p.Name())

	//TODO Devin
	//var unitRep common2.IUnitRepository
	//unitRep = common2.NewUnitRepository4Db(pm.dag.Db)

	// Execute the PalletOne handshake
	head := pm.dag.CurrentHeader()
	mediator := pm.producer.LocalHaveActiveMediator()

	if err := p.Handshake(pm.networkId, head.Number, pm.genesis.Hash(), mediator); err != nil {
		log.Debug("PalletOne handshake failed", "err", err)
		return err
	}

	if err := pm.noMediatorCheck(p); err != nil {
		return err
	}

	if err := pm.mediatorCheck(p); err != nil {
		return err
	}

	if err := pm.transitionRun(p); err != nil {
		return err
	}

	if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
		rw.Init(p.version)
	}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		log.Error("PalletOne peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	if err := pm.downloader.RegisterPeer(p.id, p.version, p); err != nil {
		return err
	}
	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	pm.syncTransactions(p)
	log.Debug("PalletOne pm handle syncTransactions")

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			log.Debug("PalletOne message handling failed", "err", err)
			return err
		}
	}
}

/*
	1.CheckMediator notice consensus when full mediator connected.
	2.Holdup
	3.Proceing vss
	4.Vss success.mediator notice ptn.Clean ps.mediators
	5.Continue when Transition to complete
*/
func (pm *ProtocolManager) handleTransitionMsg(p *peer) error {
	//	if p.mediator && !pm.peersTransition.mediators.Has(p.ID().TerminalString()) {
	//		log.Debug("PalletOne handshake failed Lying is mediator")
	//		return errors.New("PalletOne handshake failed Lying is mediator")
	//	}

	if err := pm.peersTransition.Register(p); err != nil {
		log.Error("PalletOne peer registration failed", "err", err)
		return err
	}
	defer pm.removeTransitionPeer(p.id)

	if pm.peersTransition.MediatorsSize()-1 == len(pm.peersTransition.peers) {
		//notice consensus all mediator connected
		//if The main network is launched for the first time.
		//(Judge node have private key).Only notice consensus go to the next vote.
		//Vote finish we will transition

		//if temporary mediators should to notice consensus and p.transitionCh <- transitionCancel
		pm.transCycleConnCh <- transitionCancel
		//go pm.producer.StartVSSProtocol()
	}
	for {

		log.Debug("PalletOne handleTransitionMsg transitions ReadMsg")
		// Read the next message from the remote peer, and ensure it's fully consumed
		msg, err := p.rw.ReadMsg()
		if err != nil {
			return err
		}
		if msg.Size > ProtocolMaxMsgSize {
			return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
		}
		//TODO judge msg.Code must vss code when peer In the vss processing stage.
		//Otherwise, immediatly return errResp.On the basis of ps.mediators

		defer msg.Discard()

		select {
		case event := <-p.transitionCh:
			if event == transitionCancel {
				//TODO restart transtion connect
				log.Debug("PalletOne handleTransitionMsg transitions each other connected ok")
				return nil
			}
		default:
		}

		// Handle the message depending on its contents
		switch {
		case msg.Code == TransitionReq:
		case msg.Code == TransitionResp:
			//21*21 resp
			// append by Albert·Gou
		case msg.Code == VSSDealMsg:
			// comment by Albert·Gou
			//var vssmsg vssMsg
			//if err := msg.Decode(&vssmsg); err != nil {

			var deal mp.VSSDealEvent
			if err := msg.Decode(&deal); err != nil {
				log.Info("===VSSDealMsg===", "err:", err)
				return errResp(ErrDecode, "%v: %v", msg, err)
			}
			pm.producer.ToProcessDeal(&deal)

			// comment by Albert·Gou
			////TODO vssmark
			//if !pm.peers.PeersWithoutVss(vssmsg.NodeId) {
			//	pm.producer.ToProcessDeal(vssmsg.Deal)
			//	pm.peers.MarkVss(vssmsg.NodeId)
			//	pm.BroadcastVss(vssmsg.NodeId, vssmsg.Deal)
			//}

			// append by Albert·Gou
		case msg.Code == VSSResponseMsg:
			var resp mp.VSSResponseEvent
			if err := msg.Decode(&resp); err != nil {
				log.Info("===VSSResponseMsg===", "err:", err)
				return errResp(ErrDecode, "%v: %v", msg, err)
			}
			pm.producer.ToProcessResponse(&resp)
		default:
			return errResp(ErrInvalidMsgCode, "%v", msg.Code)
		}
	}

	/*
		//2.holdup
		step := <-p.transitionCh
		switch {
		case step == transitionStep1:
			log.Info("PalletOne transition all mediators each other connected", "mediators num:", pm.peersTransition.MediatorsSize())
		}

		select {
		case event := <-p.transitionCh:
			if event == transitionCancel {
				//TODO restart transtion connect
			}
		}
	*/
}

//switchover: vote mediator all connected next to switchover official peerset
func (pm *ProtocolManager) switchover(p *peer) error {
	return nil
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {

	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	//TODO judge msg.Code must vss code when peer In the vss processing stage.
	//Otherwise, immediatly return errResp.On the basis of ps.mediators

	defer msg.Discard()

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case msg.Code == GetBlockHeadersMsg:
		// Decode the complex header query
		var query getBlockHeadersData
		if err := msg.Decode(&query); err != nil {
			log.Info("GetBlockHeadersMsg Decode", "err:", err, "msg:", msg)
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		hashMode := query.Origin.Hash != (common.Hash{})

		// Gather headers until the fetch or network limits is reached
		var (
			bytes   common.StorageSize
			headers []*modules.Header
			unknown bool
		)

		for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
			// Retrieve the next header satisfying the query
			var origin *modules.Header
			if hashMode {
				//origin = pm.blockchain.GetHeaderByHash(query.Origin.Hash)
				origin = pm.dag.GetHeaderByHash(query.Origin.Hash)
			} else {
				//index *modules.ChainIndex
				origin = pm.dag.GetHeaderByNumber(query.Origin.Number)
			}
			if origin == nil {
				break
			}

			number := origin.Number.Index
			headers = append(headers, origin)
			bytes += estHeaderRlpSize

			// Advance to the next header of the query
			switch {
			case hashMode && query.Reverse:
				// Hash based traversal towards the genesis block
				for i := 0; i < int(query.Skip)+1; i++ {
					if header, err := pm.dag.GetHeader(query.Origin.Hash, number); err == nil && header != nil {
						if number != 0 {
							query.Origin.Hash = header.ParentsHash[0]
						}
						number--
					} else {
						log.Info("========GetBlockHeadersMsg========", "number", number, "err:", err)
						unknown = true
						break
					}
				}
			case hashMode && !query.Reverse:
				// Hash based traversal towards the leaf block
				var (
					current = origin.Number.Index
					next    = current + query.Skip + 1
					index   = origin.Number
				)
				if next <= current {
					infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
					log.Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
					unknown = true
				} else {
					index.Index = next
					if header := pm.dag.GetHeaderByNumber(index); header != nil {
						if pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)[query.Skip] == query.Origin.Hash {
							query.Origin.Hash = header.Hash()
						} else {
							unknown = true
						}
					} else {
						unknown = true
					}
				}
			case query.Reverse:
				// Number based traversal towards the genesis block
				if query.Origin.Number.Index >= query.Skip+1 {
					query.Origin.Number.Index -= query.Skip + 1
				} else {
					log.Info("========GetBlockHeadersMsg========", "query.Reverse", "unknown is true")
					unknown = true
				}

			case !query.Reverse:
				// Number based traversal towards the leaf block
				query.Origin.Number.Index += query.Skip + 1
			}
		}
		log.Debug("========GetBlockHeadersMsg========", "query.Amount", query.Amount, "send number:", len(headers))
		return p.SendUnitHeaders(headers)

	case msg.Code == BlockHeadersMsg:
		// A batch of headers arrived to one of our previous requests
		var headers []*modules.Header
		if err := msg.Decode(&headers); err != nil {
			log.Info("msg.Decode", "err:", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// If no headers were received, but we're expending a DAO fork check, maybe it's that
		if len(headers) == 0 {
			log.Debug("===handler->msg.Code == BlockHeadersMsg len(headers)is 0===")
			return nil
		}
		// Filter out any explicitly requested headers, deliver the rest to the downloader
		filter := len(headers) == 1
		if filter {
			// Irrelevant of the fork checks, send the header to the fetcher just in case
			headers = pm.fetcher.FilterHeaders(p.id, headers, time.Now())
		}
		if len(headers) > 0 || !filter {
			log.Debug("===BlockHeadersMsg ===", "len(headers):", len(headers))
			err := pm.downloader.DeliverHeaders(p.id, headers)
			if err != nil {
				log.Debug("Failed to deliver headers", "err", err.Error())
			}
		}

	case msg.Code == GetBlockBodiesMsg:
		// Decode the retrieval message
		//log.Debug("===GetBlockBodiesMsg===")
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather blocks until the fetch or network limits is reached
		var (
			hash  common.Hash
			bytes int
			//bodies []rlp.RawValue
			bodies blockBody
		)

		for bytes < softResponseLimit && len(bodies.Transactions) < downloader.MaxBlockFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			//TODO must recover
			// Retrieve the requested block body, stopping if enough was found
			//GetTransactionsByUnitHash(hash)
			txs, err := pm.dag.GetTransactionByHash(hash)
			if err != nil {
				log.Debug("===GetBlockBodiesMsg===", "GetTransactionByHash err:", err)
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			log.Debug("===GetBlockBodiesMsg===", "GetTransactionByHash txs:", txs)
			data, err := rlp.EncodeToBytes(txs)
			if err != nil {
				log.Debug("Get body rlp when rlp encode", "unit hash", hash.String(), "error", err.Error())
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			bytes += len(data)

			//for _, tx := range txs {
			bodies.Transactions = append(bodies.Transactions, txs)
			//}
		}
		//log.Debug("===GetBlockBodiesMsg===", "tempGetBlockBodiesMsgSum:", tempGetBlockBodiesMsgSum, "sum:", sum)
		log.Debug("===GetBlockBodiesMsg===", "len(bodies):", len(bodies.Transactions), "bytes:", bytes)
		return p.SendBlockBodies([]*blockBody{&bodies})
		//return p.SendBlockBodiesRLP(bodies)

	case msg.Code == BlockBodiesMsg:
		//log.Debug("===BlockBodiesMsg===")
		// A batch of block bodies arrived to one of our previous requests
		var request blockBodiesData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver them all to the downloader for queuing
		transactions := make([][]*modules.Transaction, len(request))
		sum := 0
		for i, body := range request {
			transactions[i] = body.Transactions
			sum++
		}

		log.Debug("===BlockBodiesMsg===", "len(transactions:)", len(transactions), "transactions[0]:", transactions[0])
		// Filter out any explicitly requested bodies, deliver the rest to the downloader
		filter := len(transactions) > 0
		if filter {
			log.Debug("===BlockBodiesMsg->FilterBodies===")
			transactions = pm.fetcher.FilterBodies(p.id, transactions, time.Now())
		}
		if len(transactions) > 0 || !filter {
			log.Debug("===BlockBodiesMsg->DeliverBodies===")
			err := pm.downloader.DeliverBodies(p.id, transactions)
			if err != nil {
				log.Debug("Failed to deliver bodies", "err", err.Error())
			}
		}
	case msg.Code == GetNodeDataMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash  common.Hash
			bytes int
			data  [][]byte
		)
		for bytes < softResponseLimit && len(data) < downloader.MaxStateFetch {
			// Retrieve the hash of the next state entry
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested state entry, stopping if enough was found
			/*if entry, err := pm.blockchain.TrieNode(hash); err == nil {
				data = append(data, entry)
				bytes += len(entry)
			}*/
		}
		return p.SendNodeData(data)

	case msg.Code == NodeDataMsg:
		// A batch of node state data arrived to one of our previous requests
		var data [][]byte
		if err := msg.Decode(&data); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverNodeData(p.id, data); err != nil {
			log.Debug("Failed to deliver node state data", "err", err.Error())
		}

	case msg.Code == NewBlockHashesMsg:
		log.Debug("===NewBlockHashesMsg===")
		var announces newBlockHashesData
		if err := msg.Decode(&announces); err != nil {
			log.Debug("===NewBlockHashesMsg===", "Decode err:", err)
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		// Mark the hashes as present at the remote node
		for _, block := range announces {
			p.MarkUnit(block.Hash)
		}
		// Schedule all the unknown hashes for retrieval
		unknown := make(newBlockHashesData, 0, len(announces))
		for _, block := range announces {
			if !pm.dag.HasUnit(block.Hash) {
				unknown = append(unknown, block)
			}
		}
		log.Debug("===NewBlockHashesMsg===", "len(unknown):", len(unknown))
		for _, block := range unknown {
			pm.fetcher.Notify(p.id, block.Hash, block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
		}
	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block

		var unit modules.Unit
		if err := msg.Decode(&unit); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		unit.ReceivedAt = msg.ReceivedAt
		unit.ReceivedFrom = p
		log.Info("===NewBlockMsg===", "index:", unit.Number().Index)

		// Mark the peer as owning the block and schedule it for import
		p.MarkUnit(unit.UnitHash)
		pm.fetcher.Enqueue(p.id, &unit)

		hash, number := p.Head(unit.Number().AssetID)

		if common.EmptyHash(hash) || (!common.EmptyHash(hash) && unit.UnitHeader.ChainIndex().Index > number.Index) {
			trueHead := unit.Hash()
			log.Info("=================handler p.SetHead===============")
			p.SetHead(trueHead, unit.UnitHeader.ChainIndex())
			// Schedule a sync if above ours. Note, this will not fire a sync for a gap of
			// a singe block (as the true TD is below the propagated block), however this
			// scenario should easily be covered by the fetcher.
			//如果在我们上面安排一个同步。注意，这将不会为单个块的间隙触发同步(因为真正的TD位于传播的块之下)，
			//但是这个场景应该很容易被fetcher所覆盖。
			currentUnit := pm.dag.CurrentUnit()
			if currentUnit != nil && unit.UnitHeader.ChainIndex().Index > currentUnit.UnitHeader.ChainIndex().Index {
				go pm.synchronise(p, unit.Number().AssetID)
			}
		}

	case msg.Code == TxMsg:
		log.Info("===============ProtocolManager TxMsg====================")
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
			log.Debug("ProtocolManager handlmsg TxMsg pm.acceptTxs==0")
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs []*modules.Transaction
		if err := msg.Decode(&txs); err != nil {
			log.Info("ProtocolManager handlmsg TxMsg", "Decode err:", err, "msg:", msg)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		//TODO VerifyTX
		log.Info("===============ProtocolManager", "TxMsg txs:", txs)
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}
			p.MarkTransaction(tx.Hash())
		}
		log.Info("===============ProtocolManager TxMsg AddRemotes====================")
		pm.txpool.AddRemotes(txs)
	case msg.Code == ConsensusMsg:
		var consensusmsg string
		if err := msg.Decode(&consensusmsg); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		log.Info("ConsensusMsg recv:", consensusmsg)
		if consensusmsg == "A" {
			p.SendConsensus("Hello I received A")
		}

	// append by Albert·Gou
	case msg.Code == NewProducedUnitMsg:
		// Retrieve and decode the propagated new produced unit
		var unit modules.Unit
		if err := msg.Decode(&unit); err != nil {
			log.Info("===NewProducedUnitMsg===", "err:", err)
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		pm.producer.ToUnitTBLSSign(&unit)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}

	return nil
}

// BroadcastTx will propagate a transaction to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTx(hash common.Hash, tx *modules.Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it
	peers := pm.peers.PeersWithoutTx(hash)
	//FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for _, peer := range peers {
		peer.SendTransactions(modules.Transactions{tx})
	}
	log.Trace("Broadcast transaction", "hash", hash, "recipients", len(peers))
}

func (self *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case event := <-self.txCh:
			log.Debug("=====ProtocolManager=====", "txBroadcastLoop event.Tx", event.Tx)
			self.BroadcastTx(event.Tx.Hash(), event.Tx)

		// Err() channel will be closed when unsubscribing.
		case <-self.txSub.Err():
			return
		}
	}
}

// BroadcastUnit will either propagate a unit to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastUnit(unit *modules.Unit, propagate bool) {
	hash := unit.Hash()
	peers := pm.peers.PeersWithoutUnit(hash)

	// If propagation is requested, send to a subset of the peer
	if propagate {
		//TODO must recover
		/*
			for _, parentHash := range unit.ParentHash() {
				if parent := pm.dag.GetUnit(parentHash); parent == nil {
					log.Error("Propagating dangling block", "index", unit.Number().Index, "hash", hash)
					return
				}
			}
		*/
		// Send the block to a subset of our peers
		transfer := peers[:int(math.Sqrt(float64(len(peers))))]
		for _, peer := range transfer {
			peer.SendNewUnit(unit)
		}
		log.Trace("BroadcastUnit Propagated block", "hash", hash, "recipients", len(transfer), "duration", common.PrettyDuration(time.Since(unit.ReceivedAt)))
		return
	}

	// Otherwise if the block is indeed in out own chain, announce it
	if pm.dag.HasUnit(hash) {
		for _, peer := range peers {
			peer.SendNewUnitHashes([]common.Hash{hash}, []modules.ChainIndex{unit.Number()})
		}
		log.Trace("BroadcastUnit Announced block", "hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(unit.ReceivedAt)))
	} else {
		log.Debug("===BroadcastUnit===", "pm.dag.HasUnit(hash) is false hash:", hash.String())
	}
}

// @author Albert·Gou
func (self *ProtocolManager) newProducedUnitBroadcastLoop() {
	for {
		select {
		case event := <-self.newProducedUnitCh:
			self.BroadcastNewProducedUnit(event.Unit)

			// appended by wangjiyou
			self.BroadcastUnit(event.Unit, true)
			self.BroadcastUnit(event.Unit, false)

		// Err() channel will be closed when unsubscribing.
		case <-self.newProducedUnitSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (self *ProtocolManager) sigShareTransmitLoop() {
	for {
		select {
		case event := <-self.sigShareCh:
			med := self.dag.GetUnit(event.Hash).UnitAuthor()
			node := self.dag.GetActiveMediator(*med).Node
			self.TransmitSigShare(node, &event)

			// Err() channel will be closed when unsubscribing.
		case <-self.sigShareSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) TransmitSigShare(node *discover.Node, sigShare *mp.SigShareEvent) {
	peer, self := pm.getTransitionPeer(node)
	if self {
		//size, reader, err := rlp.EncodeToReader(sigShare)
		//if err != nil {
		//	log.Error(err.Error())
		//}
		//
		//var s mp.SigShareEvent
		//stream := rlp.NewStream(reader, uint64(size))
		//if err := stream.Decode(&s); err != nil {
		//	log.Error(err.Error())
		//}
		//pm.producer.ToTBLSRecover(&s)

		pm.producer.ToTBLSRecover(sigShare)
		return
	}

	if peer == nil {
		return
	}

	err := peer.SendSigShare(sigShare)
	if err != nil {
		log.Error(err.Error())
	}
}

func (self *ProtocolManager) ceBroadcastLoop() {
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
	peers := pm.peers.GetPeers()
	for _, peer := range peers {
		peer.SendConsensus(ce)
	}
}

// NodeInfo represents a short summary of the PalletOne sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network uint64 `json:"network"` // PalletOne network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Index   uint64
	Genesis common.Hash `json:"genesis"` // SHA3 hash of the host's genesis block
	Head    common.Hash `json:"head"`    // SHA3 hash of the host's best owned block
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (self *ProtocolManager) NodeInfo() *NodeInfo {
	unit := self.dag.CurrentUnit()
	index := uint64(0)
	if unit != nil {
		index = unit.Number().Index
	}
	return &NodeInfo{
		Network: self.networkId,
		Index:   index,
	}
}

func TestMakeTransaction(nonce uint64) *modules.Transaction {
	pay := modules.PaymentPayload{
		Input:  []*modules.Input{},
		Output: []*modules.Output{},
	}
	holder := common.Address{}
	holder.SetString("P1MEh8GcaAwS3TYTomL1hwcbuhnQDStTmgc")
	msg0 := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay,
	}
	tx := &modules.Transaction{
		TxMessages: []*modules.Message{msg0},
	}
	txHash, err := rlp.EncodeToBytes(tx.TxMessages)
	if err != nil {
		msg := fmt.Sprintf("Get genesis transactions hash error: %s", err)
		log.Error(msg)
		return nil
	}
	tx.TxHash.SetBytes(txHash)

	return tx
}

// @author Albert·Gou
// BroadcastNewProducedUnit will propagate a new produced unit to all of active mediator's peers
func (pm *ProtocolManager) BroadcastNewProducedUnit(unit *modules.Unit) {
	peers := pm.GetActiveMediatorPeers()
	for _, peer := range peers {
		if peer == nil {
			pm.producer.ToUnitTBLSSign(unit)
			continue
		}

		err := peer.SendNewProducedUnit(unit)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// getActiveMediatorPeer, retrieve specified active mediator peer.
// If it is the node itself, p is nil and self is true
// @author Albert·Gou
func (pm *ProtocolManager) getActiveMediatorPeer(node *discover.Node) (p *peer, self bool) {
	id := node.ID
	if pm.srvr.Self().ID == id {
		self = true
	}

	p = pm.peers.Peer(id.TerminalString())
	if p == nil && !self {
		log.Debug(fmt.Sprintf("Active Mediator Peer not exist: %v", node.String()))
	}

	return
}

// AtiveMeatorPeers retrieves a list of peers that active mediator.
// If the value is nil, it is the node itself
// @author Albert·Gou
func (pm *ProtocolManager) GetActiveMediatorPeers() map[string]*peer {
	nodes := pm.dag.GetActiveMediatorNodes()
	list := make(map[string]*peer, len(nodes))

	for id, node := range nodes {
		peer, self := pm.getActiveMediatorPeer(node)
		if peer != nil || self {
			list[id] = peer
		}
	}

	return list
}

func (pm *ProtocolManager) getTransitionPeer(node *discover.Node) (p *peer, self bool) {
	id := node.ID
	if pm.srvr.Self().ID == id {
		self = true
	}

	p = pm.peersTransition.Peer(id.TerminalString())
	if p == nil && !self {
		log.Debug(fmt.Sprintf("Active Mediator Peer not exist: %v", node.String()))
	}

	return
}
func (pm *ProtocolManager) GetTransitionPeers() map[string]*peer {
	nodes := pm.dag.GetActiveMediatorNodes()
	list := make(map[string]*peer, len(nodes))

	for id, node := range nodes {
		peer, self := pm.getTransitionPeer(node)
		if peer != nil || self {
			list[id] = peer
		}
	}

	return list
}

/*

func test() {
	body := blockBody{}
	tx := TestMakeTransaction(uint64(tempGetBlockBodiesMsgSum))
	body.Transactions = append(body.Transactions, tx)
	data, err := rlp.EncodeToBytes(body)
	if err != nil {
		log.Debug("===GetBlockBodiesMsg===", "rlp.EncodeToBytes err:", err)
		continue
	}
	bodies = append(bodies, data)
	bytes += len(data)
	sum++
	tempGetBlockBodiesMsgSum++
}

	case msg.Code == GetBlockHeadersMsg:
		// Decode the complex header query
		var query getBlockHeadersData
		if err := msg.Decode(&query); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		hashMode := query.Origin.Hash != (common.Hash{})
		first := true
		maxNonCanonical := uint64(100)

		// Gather headers until the fetch or network limits is reached
		var (
			bytes   common.StorageSize
			headers []*modules.Header
			unknown bool
		)
		for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
			// Retrieve the next header satisfying the query
			var origin *modules.Header
			if hashMode {
				if first {
					first = false
					origin = pm.dag.GetHeaderByHash(query.Origin.Hash)
					if origin != nil {
						query.Origin.Number = origin.Number
					}
				} else {
					var err error
					origin, err = pm.dag.GetHeader(query.Origin.Hash, query.Origin.Number.Index)
					if err != nil {
						break
					}
				}
			} else {
				origin = pm.dag.GetHeaderByNumber(query.Origin.Number)
			}
			if origin == nil {
				break
			}
			headers = append(headers, origin)
			//TODO must modify
			bytes += estHeaderRlpSize

			// Advance to the next header of the query
			switch {
			case hashMode && query.Reverse:
				// Hash based traversal towards the genesis block
				ancestor := query.Skip + 1
				if ancestor == 0 {
					unknown = true
				} else {
					query.Origin.Hash, query.Origin.Number = pm.dag.GetAncestor(query.Origin.Hash, query.Origin.Number, ancestor, &maxNonCanonical)
					unknown = (query.Origin.Hash == common.Hash{})
				}
			case hashMode && !query.Reverse:
				// Hash based traversal towards the leaf block
				var (
					current = origin.Number.Uint64()
					next    = current + query.Skip + 1
				)
				if next <= current {
					infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
					p.Log().Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
					unknown = true
				} else {
					if header := pm.blockchain.GetHeaderByNumber(next); header != nil {
						nextHash := header.Hash()
						expOldHash, _ := pm.blockchain.GetAncestor(nextHash, next, query.Skip+1, &maxNonCanonical)
						if expOldHash == query.Origin.Hash {
							query.Origin.Hash, query.Origin.Number = nextHash, next
						} else {
							unknown = true
						}
					} else {
						unknown = true
					}
				}
			case query.Reverse:
				// Number based traversal towards the genesis block
				if query.Origin.Number >= query.Skip+1 {
					query.Origin.Number -= query.Skip + 1
				} else {
					unknown = true
				}

			case !query.Reverse:
				// Number based traversal towards the leaf block
				query.Origin.Number += query.Skip + 1
			}
		}
		return p.SendBlockHeaders(headers)
*/
