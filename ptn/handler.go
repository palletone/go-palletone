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
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/consensus"
	"github.com/palletone/go-palletone/consensus/jury"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/palletcache"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/common/crypto"
	util2 "github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/comm"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/manger"
	"github.com/palletone/go-palletone/contracts/utils"
	dagerrors "github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptn/fetcher"
	"github.com/palletone/go-palletone/validator"
	"github.com/palletone/go-palletone/vm/common"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header

	// txChanSize is the size of channel listening to TxPreEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
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
	//protocolName string
	mainAssetId   modules.AssetId
	receivedCache palletcache.ICache
	fastSync      uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs     uint32 // Flag whether we're considered synchronized (enables transaction processing)

	lightSync uint32 //Flag whether light sync is enabled

	txpool   txPool
	maxPeers int

	downloader *downloader.Downloader
	fetcher    *fetcher.Fetcher
	peers      *peerSet

	//lightdownloader *downloader.Downloader
	//lightFetcher    *lps.LightFetcher
	//lightPeers      *peerSet

	SubProtocols []p2p.Protocol

	eventMux *event.TypeMux
	txCh     chan modules.TxPreEvent
	txSub    event.Subscription

	dag dag.IDag

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh      chan *peer
	txsyncCh       chan *txsync
	quitSync       chan struct{}
	dockerQuitSync chan struct{}
	noMorePeers    chan struct{}

	//consensus test for p2p
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
	groupSigCh  chan mp.GroupSigEvent
	groupSigSub event.Subscription

	// append by Albert·Gou
	vssDealCh  chan mp.VSSDealEvent
	vssDealSub event.Subscription

	// append by Albert·Gou
	vssResponseCh  chan mp.VSSResponseEvent
	vssResponseSub event.Subscription

	//contract exec
	contractProc consensus.ContractInf
	contractCh   chan jury.ContractEvent
	contractSub  event.Subscription

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup

	genesis *modules.Unit

	activeMediatorsUpdatedCh  chan modules.ActiveMediatorsUpdatedEvent
	activeMediatorsUpdatedSub event.Subscription

	toGroupSignCh  chan modules.ToGroupSignEvent
	toGroupSignSub event.Subscription
}

// NewProtocolManager returns a new PalletOne sub protocol manager. The PalletOne sub protocol manages peers capable
// with the PalletOne network.
func NewProtocolManager(mode downloader.SyncMode, networkId uint64, gasToken modules.AssetId, txpool txPool,
	dag dag.IDag, mux *event.TypeMux, producer producer, genesis *modules.Unit,
	contractProc consensus.ContractInf, engine core.ConsensusEngine) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkId: networkId,
		dag:       dag,
		//protocolName: protocolName,
		txpool:     txpool,
		eventMux:   mux,
		consEngine: engine,
		peers:      newPeerSet(),
		//lightPeers:   newPeerSet(),
		newPeerCh:      make(chan *peer),
		noMorePeers:    make(chan struct{}),
		txsyncCh:       make(chan *txsync),
		quitSync:       make(chan struct{}),
		dockerQuitSync: make(chan struct{}),
		genesis:        genesis,
		producer:       producer,
		contractProc:   contractProc,
		lightSync:      uint32(1),
		receivedCache:  freecache.NewCache(5 * 1024 * 1024),
	}
	symbol, _, _, _, _ := gasToken.ParseAssetId()
	protocolName := symbol
	//asset, err := modules.NewAsset(strings.ToUpper(gasToken), modules.AssetType_FungibleToken,
	// 8, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, modules.UniqueIdType_Null, modules.UniqueId{})
	//if err != nil {
	//	log.Error("ProtocolManager new asset err", err)
	//	return nil, err
	//}
	//manager.mainAssetId = asset.AssetId
	manager.mainAssetId = gasToken
	// Figure out whether to allow fast sync or not
	/*blockchain.CurrentBlock().NumberU64() > 0 */
	//TODO must modify.The second start would Blockchain not empty, fast sync disabled
	//if mode == downloader.FastSync && dag.CurrentUnit().UnitHeader.Index() > 0 {
	//	log.Info("dag not empty, fast sync disabled")
	//	mode = downloader.FullSync
	//}

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
		// Compatible; initialize the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    protocolName,
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
				return manager.NodeInfo(genesis.UnitHash)
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				if p := manager.peers.Peer(id.TerminalString()); p != nil {
					return p.Info(p.Caps()[0].Name)
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}

	// Construct the different synchronization mechanisms
	manager.downloader = downloader.New(mode, manager.eventMux, manager.removePeer, nil, dag, txpool)
	manager.fetcher = manager.newFetcher()

	//manager.lightdownloader = downloader.New(downloader.LightSync, manager.eventMux, nil, nil, dag, txpool)
	//manager.lightFetcher = manager.newLightFetcher()
	return manager, nil
}
func (pm *ProtocolManager) IsExistInCache(id []byte) bool {
	_, err := pm.receivedCache.Get(id)
	if err != nil { //Not exist, add it!
		pm.receivedCache.Set(id, nil, 60*5)
	}
	return err == nil
}

func (pm *ProtocolManager) newFetcher() *fetcher.Fetcher {
	validatorFn := func(unit *modules.Unit) error {
		//return dagerrors.ErrFutureBlock
		hash := unit.Hash()
		verr := validator.ValidateUnitBasic(unit)
		if verr != nil && !validator.IsOrphanError(verr) {
			return verr
		}
		log.Debugf("Importing propagated block insert DAG End ValidateUnitExceptGroupSig, unit: %s,validated: %v",
			hash.String(), verr)
		return dagerrors.ErrFutureBlock
	}
	heighter := func(assetId modules.AssetId) uint64 {
		log.Debug("Enter PalletOne Fetcher heighter")
		defer log.Debug("End PalletOne Fetcher heighter")
		unit := pm.dag.GetCurrentUnit(assetId)
		if unit != nil {
			return unit.NumberU64()
		}
		return uint64(0)
	}
	inserter := func(blocks modules.Units) (int, error) {
		// If fast sync is running, deny importing weird blocks
		if atomic.LoadUint32(&pm.fastSync) == 1 {
			log.Warn("Discarded bad propagated block", "number", blocks[0].Number().Index, "hash", blocks[0].Hash())
			return 0, errors.New("fasting sync")
		}
		log.Debug("Fetcher", "manager.dag.InsertDag index:", blocks[0].Number().Index, "hash", blocks[0].Hash())

		atomic.StoreUint32(&pm.acceptTxs, 1) // Mark initial sync done on any fetcher import

		// setPending txs in txpool
		for _, u := range blocks {
			hash := u.Hash()
			if pm.dag.IsHeaderExist(hash) {
				continue
			}
			pm.txpool.SetPendingTxs(hash, u.NumberU64(), u.Transactions())
		}

		account, err := pm.dag.InsertDag(blocks, pm.txpool, false)
		if err == nil {
			go func() {
				var (
					events = make([]interface{}, 0, 2)
				)
				events = append(events, modules.ChainHeadEvent{Unit: blocks[0]})
				events = append(events, modules.ChainEvent{Unit: blocks[0], Hash: blocks[0].UnitHash})
				pm.dag.PostChainEvents(events)
			}()
		}
		return account, err

	}
	return fetcher.New(pm.dag.IsHeaderExist, validatorFn, pm.BroadcastUnit, heighter, inserter, pm.removePeer)
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
	//pm.lightdownloader.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	peer.Peer.Disconnect(p2p.DiscUselessPeer)
}

func (pm *ProtocolManager) Start(srvr *p2p.Server, maxPeers int, syncCh chan bool) {
	pm.srvr = srvr
	pm.maxPeers = maxPeers

	// start sync handlers
	//定时与相邻个体进行全链的强制同步,syncer()首先启动fetcher成员，然后进入一个无限循环，
	//每次循环中都会向相邻peer列表中“最优”的那个peer作一次区块全链同步
	go pm.syncer(syncCh)

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
	// send unit group signature
	pm.groupSigCh = make(chan mp.GroupSigEvent)
	pm.groupSigSub = pm.producer.SubscribeGroupSigEvent(pm.groupSigCh)
	go pm.groupSigBroadcastLoop()

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

	//contract exec
	if pm.contractProc != nil {
		pm.contractCh = make(chan jury.ContractEvent)
		pm.contractSub = pm.contractProc.SubscribeContractEvent(pm.contractCh)
	}

	pm.activeMediatorsUpdatedCh = make(chan modules.ActiveMediatorsUpdatedEvent)
	pm.activeMediatorsUpdatedSub = pm.dag.SubscribeActiveMediatorsUpdatedEvent(pm.activeMediatorsUpdatedCh)
	go pm.activeMediatorsUpdatedEventRecvLoop()

	pm.toGroupSignCh = make(chan modules.ToGroupSignEvent)
	pm.toGroupSignSub = pm.dag.SubscribeToGroupSignEvent(pm.toGroupSignCh)
	go pm.toGroupSignEventRecvLoop()

	if pm.consEngine != nil {
		pm.ceCh = make(chan core.ConsensusEvent, txChanSize)
		pm.ceSub = pm.consEngine.SubscribeCeEvent(pm.ceCh)
		go pm.ceBroadcastLoop()
	}
	//  是否为linux系統
	if runtime.GOOS == "linux" {
		//创建 docker client
		client, err := util.NewDockerClient()
		if err != nil {
			log.Error("util.NewDockerClient", "error", err)
			return
		}
		//创建gptn-net网络
		_, err = client.NetworkInfo(core.DefaultUccNetworkMode)
		if err != nil {
			log.Debugf("client.NetworkInfo error: %s", err.Error())
			_, err := client.CreateNetwork(docker.CreateNetworkOptions{Name: core.DefaultUccNetworkMode, Driver: "bridge"})
			if err != nil {
				log.Debugf("client.CreateNetwork error: %s", err.Error())
				return
			}
		}
		//拉取gptn发布版本对应的goimg基础镜像，防止卡住
		go func() {
			goimg := contractcfg.Goimg + ":" + contractcfg.GptnVersion
			_, err = client.InspectImage(goimg)
			if err != nil {
				log.Debugf("Image %s does not exist locally, attempt pull", goimg)
				err = client.PullImage(docker.PullImageOptions{Repository: contractcfg.Goimg, Tag: contractcfg.GptnVersion}, docker.AuthConfiguration{})
				if err != nil {
					log.Debugf("Failed to pull %s: %s", goimg, err)
				}
			}
			//  获取本地列表，迁移服务器的重启容器，容器不存在且不过期
			dag, err := comm.GetCcDagHand()
			if err != nil {
				log.Debugf("get contract dag error %s", err.Error())
				return
			}
			ccs, err := manger.GetChaincodes(dag)
			if err != nil {
				log.Debugf("get chaincodes error %s", err.Error())
				return
			}
			//启动退出的容器，包括本地有的和本地没有的
			for _, c := range ccs {
				//conName := c.Name+c.Version+":"+contractcfg.GetConfig().ContractAddress
				rd, _ := crypto.GetRandomBytes(32)
				txid := util2.RlpHash(rd)
				//  启动gptn时启动Jury对应的没有过期的用户合约容器
				if !c.IsExpired {
					manger.RestartContainer(dag, "palletone", c.Id, txid.String())
				}
			}
		}()

		go pm.dockerLoop(client)

	}
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping PalletOne protocol")

	pm.newProducedUnitSub.Unsubscribe()
	pm.sigShareSub.Unsubscribe()
	pm.groupSigSub.Unsubscribe()
	pm.vssDealSub.Unsubscribe()
	pm.vssResponseSub.Unsubscribe()
	pm.activeMediatorsUpdatedSub.Unsubscribe()
	pm.toGroupSignSub.Unsubscribe()
	pm.contractSub.Unsubscribe()
	pm.txSub.Unsubscribe() // quits txBroadcastLoop
	//pm.ceSub.Unsubscribe()

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
	//pm.lightPeers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	//stop dockerLoop
	//pm.dockerQuitSync <- struct{}{}
	close(pm.dockerQuitSync)
	log.Info("PalletOne protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, p, newMeteredMsgWriter(rw))
}

// handle is the callback invoked to manage the life cycle of an ptn peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	log.Debug("Enter ProtocolManager handle", "peer id:", p.id)
	defer log.Debug("End ProtocolManager handle", "peer id:", p.id)

	//if len(p.Caps()) > 0 && (pm.SubProtocols[0].Name != p.Caps()[0].Name) {
	//	return pm.PartitionHandle(p)
	//}
	return pm.LocalHandle(p)

}

func (pm *ProtocolManager) LocalHandle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		log.Debug("ProtocolManager", "handler DiscTooManyPeers:", p2p.DiscTooManyPeers,
			"pm.peers.Len()", pm.peers.Len(), "peers", pm.peers.GetPeers())
		return p2p.DiscTooManyPeers
	}
	log.Debug("PalletOne peer connected", "name", p.id, "p Trusted:", p.Peer.Info().Network.Trusted)
	// @分区后需要用token获取
	//head := pm.dag.CurrentHeader(pm.mainAssetId)
	var (
		number = &modules.ChainIndex{}
		hash   = common.Hash{}
		stable = &modules.ChainIndex{}
	)
	if head := pm.dag.CurrentHeader(pm.mainAssetId); head != nil {
		number = head.Number
		hash = head.Hash()
		stable = pm.dag.GetStableChainIndex(pm.mainAssetId)
	}

	log.Debug("ProtocolManager LocalHandle pre Handshake", "index", number.Index, "stable", stable)
	// Execute the PalletOne handshake
	if err := p.Handshake(pm.networkId, number, pm.genesis.Hash(), hash, stable); err != nil {
		log.Debug("PalletOne handshake failed", "err", err)
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
	//pm.syncTransactions(p)

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			log.Debug("PalletOne message handling failed", "err", err, "p.id", p.id)
			return err
		}
	}
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

	defer msg.Discard()

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return pm.StatusMsg(msg, p)

		// Block header query, collect the requested headers and reply
	case msg.Code == GetBlockHeadersMsg:
		// Decode the complex header query
		return pm.GetBlockHeadersMsg(msg, p)

	case msg.Code == BlockHeadersMsg:
		// A batch of headers arrived to one of our previous requests
		return pm.BlockHeadersMsg(msg, p)

	case msg.Code == GetBlockBodiesMsg:
		// Decode the retrieval message
		return pm.GetBlockBodiesMsg(msg, p)

	case msg.Code == BlockBodiesMsg:
		// A batch of block bodies arrived to one of our previous requests
		return pm.BlockBodiesMsg(msg, p)

	case msg.Code == GetNodeDataMsg:
		// Decode the retrieval message
		return pm.GetNodeDataMsg(msg, p)

	case msg.Code == NodeDataMsg:
		// A batch of node state data arrived to one of our previous requests
		return pm.NodeDataMsg(msg, p)

	case msg.Code == NewBlockHashesMsg:
		return pm.NewBlockHashesMsg(msg, p)

	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block
		return pm.NewBlockMsg(msg, p)

	case msg.Code == TxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		return pm.TxMsg(msg, p)

		// append by Albert·Gou
	case msg.Code == SigShareMsg:
		return pm.SigShareMsg(msg, p)

		// 21*21 deal => 21*20 dealMsg
		// append by Albert·Gou
	case msg.Code == VSSDealMsg:
		return pm.VSSDealMsg(msg, p)

		// 21*21 deal => 21*21 resp
		// 21*21 resp => 21*21*20 respMsg
		// append by Albert·Gou
	case msg.Code == VSSResponseMsg:
		return pm.VSSResponseMsg(msg, p)

	case msg.Code == GroupSigMsg:
		return pm.GroupSigMsg(msg, p)

	case msg.Code == ContractMsg:
		return pm.ContractMsg(msg, p)

	case msg.Code == ElectionMsg:
		return pm.ElectionMsg(msg, p)

	case msg.Code == AdapterMsg:
		return pm.AdapterMsg(msg, p)

	case msg.Code == GetLeafNodesMsg:
		return pm.GetLeafNodesMsg(msg, p)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
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

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case event := <-pm.txCh:
			pm.BroadcastTx(event.Tx.Hash(), event.Tx)

			// Err() channel will be closed when unsubscribing.
		case <-pm.txSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) ContractReqLocalSend(event jury.ContractEvent) {
	log.Debug("ContractReqLocalSend", "event", event.Tx.Hash())
	pm.contractCh <- event
}

// BroadcastUnit will either propagate a unit to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastUnit(unit *modules.Unit, propagate bool) {
	hash := unit.Hash()
	// If propagation is requested, send to a subset of the peer
	data, err := rlp.EncodeToBytes(unit)
	if err != nil {
		log.Errorf("BroadcastUnit rlp encode err:%s", err.Error())
	}

	peers := pm.peers.PeersWithoutUnit(hash)
	for _, peer := range peers {
		peer.SendNewRawUnit(unit, data)
	}
	log.Trace("BroadcastUnit Propagated block", "index:", unit.Header().Number.Index,
		"hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(unit.ReceivedAt)))
}

func (pm *ProtocolManager) ElectionBroadcast(event jury.ElectionEvent, local bool) {
	//log.Debug("ElectionBroadcast", "event num", event.Event.(jury.ElectionRequestEvent),
	// "data", event.Event.(jury.ElectionRequestEvent).Data)
	peers := pm.peers.GetPeers()

	for _, peer := range peers {
		err := peer.SendElectionEvent(event)
		if err != nil {
			log.Debug("ElectionBroadcast", "err", err)
		}
	}
	if local {
		go pm.contractProc.ProcessElectionEvent(&event)
	}
}

func (pm *ProtocolManager) AdapterBroadcast(event jury.AdapterEvent) {
	peers := pm.peers.GetPeers()
	for _, peer := range peers {
		peer.SendAdapterEvent(event)
	}
}

func (pm *ProtocolManager) ContractBroadcast(event jury.ContractEvent, local bool) {
	reqId := event.Tx.RequestHash()
	peers := pm.peers.GetPeers()
	log.Debugf("[%s]ContractBroadcast, event type[%d], peers num[%d]", reqId.String()[0:8], event.CType, len(peers))
	for _, peer := range peers {
		if err := peer.SendContractTransaction(event); err != nil {
			log.Error("ContractBroadcast", "SendContractTransaction err:", err.Error())
		}
	}
	if local {
		go func() {
			err := pm.contractProc.ProcessContractEvent(&event)
			if err != nil {
				log.Errorf("[%s]ContractBroadcast, error:%s", reqId.String()[0:8], err.Error())
			}
		}()
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
func (pm *ProtocolManager) NodeInfo(genesisHash common.Hash) *NodeInfo {
	unit := pm.dag.GetCurrentUnit(pm.mainAssetId)
	var (
		index = uint64(0)
		hash  = common.Hash{}
	)
	if unit != nil {
		index = unit.Number().Index
		hash = unit.UnitHash
	}

	return &NodeInfo{
		Network: pm.networkId,
		Index:   index,
		Genesis: genesisHash,
		Head:    hash,
	}
}

//test for p2p broadcast
func (pm *ProtocolManager) BroadcastCe(ce []byte) {
	peers := pm.peers.GetPeers()
	for _, peer := range peers {
		peer.SendConsensus(ce)
	}
}
func (pm *ProtocolManager) ceBroadcastLoop() {
	for {
		select {
		case event := <-pm.ceCh:
			pm.BroadcastCe(event.Ce)

			// Err() channel will be closed when unsubscribing.
		case <-pm.ceSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) dockerLoop(client *docker.Client) {
	log.Debugf("starting docker loop")
	dag, err := comm.GetCcDagHand()
	if err != nil {
		log.Infof("db.GetCcDagHand err: %s", err.Error())
		return
	}
	for {
		select {
		case <-pm.dockerQuitSync:
			log.Infof("quit from docker loop")
			return
		case <-time.After(time.Duration(30) * time.Second):
			log.Debugf("each 30 second to get all containers")
			//  程序启动过程中的,获取所有的容器
			//  获取所有容器
			cons, err := utils.GetAllContainers(client)
			if err != nil {
				log.Errorf("utils.GetAllContainers error %s", err.Error())
			}
			//  重启退出且不过期容器
			manger.RestartContainers(client, dag, cons)
			//  删除过期容器
			manger.RemoveExpiredConatiners(client, dag, dag.GetChainParameters().RmExpConFromSysParam, cons)
		}
	}
}
