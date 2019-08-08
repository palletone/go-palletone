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
	//"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag"
	dagerrors "github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
	"time"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header
	MaxHeaderFetch    = 192             // Amount of block headers to be fetched per retrieval request
	forceSyncCycle    = 10 * time.Second
	waitPushSync      = 200 * time.Millisecond
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var (
	errIncompatibleConfig = errors.New("incompatible configuration")
	errCancelHeaderFetch  = errors.New("header cors canceled (requested)")
	errBadPeer            = errors.New("action from bad peer ignored")
	errTimeout            = errors.New("timeout")
)

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type ProtocolManager struct {
	lightSync bool
	networkId uint64
	dag       dag.IDag
	assetId   modules.AssetId

	genesis *modules.Unit

	downloader *downloader.Downloader
	fetcher    *LightFetcher
	peers      *peerSet
	maxPeers   int

	SubProtocols []p2p.Protocol

	eventMux *event.TypeMux
	server   *CorsServer

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg *sync.WaitGroup

	mainchain *modules.MainChain
	mclock    sync.RWMutex
	corsSync  uint32

	rttEstimate   uint64
	rttConfidence uint64
	headerCh      chan dataPack
	needboradcast map[string]uint64
	bdlock        sync.RWMutex
}

// NewProtocolManager returns a new ethereum sub protocol manager. The Palletone sub protocol manages peers capable
// with the ethereum network.
func NewCorsProtocolManager(lightSync bool, networkId uint64, gasToken modules.AssetId,
	dag dag.IDag, mux *event.TypeMux, genesis *modules.Unit, quitSync chan struct{}) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	log.Debug("Enter NewCorsProtocolManager")
	manager := &ProtocolManager{
		lightSync:   lightSync,
		eventMux:    mux,
		assetId:     gasToken,
		genesis:     genesis,
		dag:         dag,
		networkId:   networkId,
		peers:       newPeerSet(),
		newPeerCh:   make(chan *peer),
		wg:          new(sync.WaitGroup),
		noMorePeers: make(chan struct{}),
		quitSync:    quitSync,
		corsSync:    0,

		rttEstimate:   uint64(rttMaxEstimate),
		rttConfidence: uint64(1000000),
		headerCh:      make(chan dataPack, 1),
		needboradcast: make(map[string]uint64),
	}

	// Initiate a sub-protocol for every implemented version we can handle
	protocolVersions := ClientProtocolVersions
	manager.SubProtocols = make([]p2p.Protocol, 0, len(protocolVersions))
	for _, ver := range protocolVersions {
		version := ver
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    "cors",
			Version: version,
			Length:  ProtocolLengths[version],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := manager.newPeer(int(version), NetworkId, p, rw)
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
					return p.Info(manager.assetId)
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}

	if manager.lightSync {
		manager.fetcher = manager.newLightFetcher()
		manager.downloader = downloader.New(downloader.LightSync, manager.eventMux, manager.removePeer, nil, dag, nil)
	}
	log.Debug("End NewCorsProtocolManager", "len(manager.SubProtocols)", len(manager.SubProtocols))
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
		return dagerrors.ErrFutureBlock
	}
	headerBroadcaster := func(p *peer, header *modules.Header, propagate bool) {
		log.Debug("Cors ProtocolManager headerBroadcaster", "hash:", header.Hash().String())
		pm.BroadcastCorsHeader(p, header)
	}
	inserter := func(headers []*modules.Header) (int, error) {
		// If fast sync is running, deny importing weird blocks
		log.Debug("Cors ProtocolManager InsertLightHeader", "manager.dag.InsertDag index:",
			headers[0].Number.Index, "hash", headers[0].Hash())
		return pm.dag.InsertLightHeader(headers)
	}
	return NewLightFetcher(pm.dag.GetHeaderByHash, pm.dag.GetLightChainHeight, headerVerifierFn,
		headerBroadcaster, inserter, pm.removePeer)
}

func (pm *ProtocolManager) BroadcastCorsHeader(p *peer, header *modules.Header) {
	pm.bdlock.RLock()
	v, ok := pm.needboradcast[p.id]
	pm.bdlock.RUnlock()
	if ok && header.Number.Index >= v {
		log.Debug("Cors ProtocolManager BroadcastCorsHeader", "assetid:", header.Number.AssetID.String(),
			"index:", header.Index(), "hash", header.Hash())
		pm.server.SendEvents(header)
	}
}

// removePeer initiates disconnection from a peer by removing it from the peer set
func (pm *ProtocolManager) removePeer(id string) {
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	pm.downloader.UnregisterPeer(id)
	pm.peers.Unregister(id)
	peer.Peer.Disconnect(p2p.DiscUselessPeer)
}

func (pm *ProtocolManager) mainchainpeers() int {
	pm.mclock.RLock()
	defer pm.mclock.RUnlock()
	return len(pm.mainchain.Peers)
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	if pm.lightSync {
		go func() {
			pm.fetcher.Start()
			defer pm.fetcher.Stop()
			defer pm.downloader.Terminate()
			forceSync := time.NewTicker(forceSyncCycle)
			defer forceSync.Stop()
			for {
				select {
				case <-pm.newPeerCh:
					//go pm.StartCorsSync()

				case <-forceSync.C:
					// Force a sync even if not enough peers are present
					if pm.maxPeers > 0 {
						log.Debug("Cors PalletOne ProtocolManager StartCorsSync", "maxpeers", pm.maxPeers)
						go pm.StartCorsSync()
					}
				case <-pm.noMorePeers:
					return
				}
			}
		}()
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
	log.Info("Stopping cors Palletone protocol")

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

	log.Info("Cors Palletone protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, nv uint64, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, nv, p, rw)
}

// handle is the callback invoked to manage the life cycle of a les peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}

	log.Debug("Enter Cors Palletone peer connected", "id", p.ID())
	defer log.Debug("End Cors Palletone peer connected", "id", p.ID())

	// Execute the Cors handshake
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

	pcs, err := pm.dag.GetPartitionChains()
	if err != nil {
		log.Debug("Cors PalletOne ProtocolManager handle GetPartitionChains", "err", err)
	}

	if err := p.Handshake(number, genesis.Hash(), headhash, pm.assetId, pcs); err != nil {
		log.Debug("Cors PalletOne handshake failed", "err", err)
		return err
	}
	//if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
	//	rw.Init(p.version)
	//}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		log.Error("Cors Palletone peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	if pm.lightSync {
		if err := pm.downloader.RegisterLightPeer(p.id, p.version, p); err != nil {
			return err
		}
	}

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		// new block announce loop
		for {
			select {
			case announce := <-p.announceChn:
				log.Debug("Cors Palletone ProtocolManager->handle", "announce", announce)
				p.SendSingleHeader([]*modules.Header{&announce.Header})

			case <-stop:
				return
			}
		}
	}()

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			log.Debug("Cors PalletOne message handling failed", "err", err)
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
	log.Trace("Cors Palletone message arrived", "code", msg.Code, "bytes", msg.Size)

	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	// Handle the message depending on its contents
	switch msg.Code {
	case StatusMsg:
		log.Trace("Received status message")
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case CorsHeaderMsg:
		return pm.CorsHeaderMsg(msg, p)

	case CorsHeadersMsg:
		return pm.CorsHeadersMsg(msg, p)

	case GetCurrentHeaderMsg:
		return pm.GetCurrentHeaderMsg(msg, p)

	case CurrentHeaderMsg:
		return pm.CurrentHeaderMsg(msg, p)

	case GetBlockHeadersMsg:
		return pm.GetBlockHeadersMsg(msg, p)

	case BlockHeadersMsg:
		return pm.BlockHeadersMsg(msg, p)

	default:
		log.Trace("Received unknown message", "code", msg.Code)
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}

}

// NodeInfo represents a short summary of the Palletone sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network uint64      `json:"network"` // Palletone network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Index   uint64      `json:"number"`  // Total difficulty of the host's blockchain
	Head    common.Hash `json:"head"`    // SHA3 hash of the host's best owned block
	Genesis common.Hash `json:"genesis"` // SHA3 hash of the host's genesis block
	//Config     *params.ChainConfig `json:"config"`     // Chain configuration for the fork rules
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo(genesisHash common.Hash) *NodeInfo {
	header := pm.dag.CurrentHeader(pm.assetId)

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
		Network: pm.networkId,
		Index:   index,
		Genesis: genesisHash,
		Head:    hash,
	}
}
