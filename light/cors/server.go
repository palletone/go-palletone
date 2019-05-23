package cors

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn"
	"sync/atomic"
	"time"
)

var (
	rttMinEstimate   = 2 * time.Second  // Minimum round-trip time to target for download requests
	rttMaxEstimate   = 20 * time.Second // Maximum round-trip time to target for download requests
	rttMinConfidence = 0.1              // Worse confidence factor in our estimated RTT value
	ttlScaling       = 3                // Constant scaling factor for RTT -> TTL conversion
	ttlLimit         = time.Minute      // Maximum TTL allowance to prevent reaching crazy timeouts
)

type CorsServer struct {
	config          *ptn.Config
	protocolManager *ProtocolManager
	privateKey      *ecdsa.PrivateKey
	corss           *p2p.Server
	quitSync        chan struct{}
}

func NewCoresServer(ptn *ptn.PalletOne, config *ptn.Config) (*CorsServer, error) {
	quitSync := make(chan struct{})
	gasToken := config.Dag.GetGasToken()
	genesis, err := ptn.Dag().GetGenesisUnit()
	if err != nil {
		log.Error("Light PalletOne New", "get genesis err:", err)
		return nil, err
	}
	//TODO version network gastoken genesis by

	pm, err := NewCorsProtocolManager(true, config.NetworkId, gasToken,
		ptn.Dag(), ptn.EventMux(), genesis, make(chan struct{}))

	if err != nil {
		log.Error("NewlesServer NewProtocolManager", "err", err)
		return nil, err
	}

	srv := &CorsServer{
		config:          config,
		protocolManager: pm,
		quitSync:        quitSync,
	}
	pm.server = srv
	log.Debug("NewCoresServer", "len(srv.protocolManager.SubProtocols)", srv.protocolManager.SubProtocols)
	return srv, nil
}

func (s *CorsServer) Protocols() []p2p.Protocol {
	return nil
}

func (s *CorsServer) CorsProtocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

func (s *CorsServer) StartCorsSync() (string, error) {
	return s.protocolManager.StartCorsSync()
}

// Start starts the LES server
func (s *CorsServer) Start(srvr *p2p.Server, corss *p2p.Server) {
	s.protocolManager.Start(s.config.LightPeers)
	s.privateKey = corss.PrivateKey
	s.corss = corss
	s.protocolManager.blockLoop()
}

// Stop stops the LES service
func (s *CorsServer) Stop() {
	go func() {
		<-s.protocolManager.noMorePeers
	}()
	s.protocolManager.Stop()
}

func (pm *ProtocolManager) blockLoop() {
	pm.wg.Add(1)
	headCh := make(chan modules.ChainEvent, 10)
	headSub := pm.dag.SubscribeChainEvent(headCh)
	go func() {
		var lastHead *modules.Header
		for {
			select {
			case ev := <-headCh:
				peers := pm.peers.AllPeers()
				if len(peers) > 0 && atomic.LoadUint32(&pm.corsSync) == 0 {
					header := ev.Unit.Header()
					hash := ev.Hash
					number := header.Number.Index
					if lastHead == nil || (header.Number.Index > lastHead.Number.Index) {
						lastHead = header
						log.Debug("Announcing block to peers", "number", number, "hash", hash)
						announce := announceData{Hash: hash, Number: *lastHead.Number, Header: *lastHead}

						for _, p := range peers {
							log.Debug("Light Palletone", "ProtocolManager->blockLoop p.ID", p.ID())
							p.announceChn <- announce
						}
					}
				}
			case <-pm.quitSync:
				headSub.Unsubscribe()
				pm.wg.Done()
				return
			}
		}
	}()
}

func (pm *ProtocolManager) AddCorsPeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	if pm.server.corss == nil {
		return false, nil
	}
	// Try to add the url as a static peer and return
	node, err := discover.ParseNode(url)
	if err != nil {
		return false, fmt.Errorf("invalid pnode: %v", err)
	}
	pm.server.corss.AddPeer(node)
	return true, nil
}

/*
type MainChain struct {
	GenesisHash common.Hash
	Status      byte //Active:1 ,Terminated:0,Suspended:2
	SyncModel   byte //Push:1 , Pull:2, Push+Pull:0
	GasToken    AssetId
	Peers       []string // IP:port format string
}
*/
func (pm *ProtocolManager) GetMainChain() (*modules.MainChain, error) {
	return pm.dag.GetMainChain()
	//mainchain := &modules.MainChain{}
	//mainchain.NetworkId = 1
	//mainchain.Version = 1
	//mainchain.GenesisHash.SetHexString("0x927c94780c89b450cf2d9bcb3febea8457bcb830f5867b9d85c74ce4df3d2ac4")
	//mainchain.GasToken = modules.PTNCOIN
	//return mainchain, nil
}

func (pm *ProtocolManager) GetPartitionChain() ([]*modules.PartitionChain, error) {
	return pm.dag.GetPartitionChains()
	//mainchains := []*modules.PartitionChain{}
	//mainchain := &modules.PartitionChain{}
	//mainchain.NetworkId = 1
	//mainchain.Version = 1
	//mainchain.GenesisHash.SetHexString("0x927c94780c89b450cf2d9bcb3febea8457bcb830f5867b9d85c74ce4df3d2ac4")
	//mainchain.GasToken = modules.PTNCOIN
	//mainchains = append(mainchains, mainchain)
	//return mainchains, nil
}

/*
type MainChain struct {
	GenesisHash common.Hash
	Status      byte //Active:1 ,Terminated:0,Suspended:2
	SyncModel   byte //Push:1 , Pull:2, Push+Pull:0
	GasToken    AssetId
	NetworkId   uint64
	Version     uint64
	Peers       []string // pnode://publickey@IP:port format string
}
*/
