package cors

import (
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/ptn"
)

type CorsServer struct {
	config          *ptn.Config
	protocolManager *ProtocolManager
	//lesTopics       []discv5.Topic
	privateKey *ecdsa.PrivateKey
	quitSync   chan struct{}

	//chtIndexer, bloomTrieIndexer *core.ChainIndexer
}

func NewCoresServer(ptn *ptn.PalletOne, config *ptn.Config) (*CorsServer, error) {
	quitSync := make(chan struct{})
	gasToken := config.Dag.GetGasToken()
	genesis, err := ptn.Dag().GetGenesisUnit()
	if err != nil {
		log.Error("Light PalletOne New", "get genesis err:", err)
		return nil, err
	}

	pm, err := NewProtocolManager(false, newPeerSet(), config.NetworkId, gasToken,
		ptn.Dag(), ptn.EventMux(), genesis)
	if err != nil {
		log.Error("NewlesServer NewProtocolManager", "err", err)
		return nil, err
	}

	srv := &CorsServer{
		config:          config,
		protocolManager: pm,
		quitSync:        quitSync,
	}

	return srv, nil
}

func (s *CorsServer) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start starts the LES server
func (s *CorsServer) Start(srvr *p2p.Server) {
	s.protocolManager.Start(s.config.LightPeers)
	s.privateKey = srvr.PrivateKey
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

}
