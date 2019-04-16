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
package light

import (
	"fmt"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/core/accounts"
	//"github.com/ethereum/go-ethereum/consensus"
	//"github.com/palletone/go-palletone/core"
	//"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/palletone/go-palletone/ptn"
	"github.com/palletone/go-palletone/ptn/downloader"
	//	"github.com/palletone/go-palletone/ptn/filters"
	//"github.com/ethereum/go-ethereum/eth/gasprice"
	//"github.com/ethereum/go-ethereum/light"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/node"
	//"github.com/palletone/go-palletone/core/types"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/internal/ptnapi"
	"github.com/palletone/go-palletone/light/les"
)

type LightEthereum struct {
	config *ptn.Config

	//odr         *LesOdr
	//relay       *LesTxRelay
	//chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers  *peerSet
	txPool *les.TxPool
	//blockchain      *light.LightChain
	protocolManager *ProtocolManager
	//serverPool      *serverPool
	//reqDist         *requestDistributor
	//retriever       *retrieveManager
	// DB interfaces
	dag dag.IDag
	// DB interfaces
	unitDb ptndb.Database // Block chain database

	//bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	//bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux *event.TypeMux
	//engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *ptnapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *ptn.Config) (*LightEthereum, error) {
	chainDb, err := ptn.CreateDB(ctx, config /*, "lightchaindata"*/)
	if err != nil {
		return nil, err
	}
	dag, err := dag.NewDag(chainDb)
	if err != nil {
		log.Error("PalletOne New", "NewDag err:", err)
		return nil, err
	}
	genesis, err := dag.GetGenesisUnit()
	if err != nil {
		log.Error("PalletOne New", "get genesis err:", err)
		return nil, err
	}
	peers := newPeerSet()
	//quitSync := make(chan struct{})

	lptn := &LightEthereum{
		config: config,
		//chainConfig:      chainConfig,
		unitDb:   chainDb,
		eventMux: ctx.EventMux,
		peers:    peers,
		//reqDist:          newRequestDistributor(peers, quitSync),
		accountManager: ctx.AccountManager,
		//engine:           eth.CreateConsensusEngine(ctx, &config.Ethash, chainConfig, chainDb),
		shutdownChan: make(chan bool),
		networkId:    config.NetworkId,
		dag:          dag,
		//bloomRequests:    make(chan chan *bloombits.Retrieval),
		//bloomIndexer:     eth.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		//chtIndexer:       light.NewChtIndexer(chainDb, true),
		//bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	//leth.txPool = NewTxPool(leth.chainConfig, leth.blockchain, leth.relay)
	//NewProtocolManager(config.SyncMode, config.NetworkId, gasToken, ptn.txPool,
	//		ptn.dag, ptn.eventMux, ptn.mediatorPlugin, genesis, ptn.contractPorcessor, ptn.engine)
	gasToken := modules.AssetId{}
	//txPool := &TxPool{}
	if lptn.protocolManager, err = NewProtocolManager(config.SyncMode, config.NetworkId, gasToken, nil,
		dag, lptn.eventMux, nil, genesis, nil, nil); err != nil {
		return nil, err
	}
	//leth.ApiBackend = &LesApiBackend{leth, nil}
	lptn.ApiBackend = &LesApiBackend{lptn}
	return lptn, nil
}

//func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
//	var name string
//	switch protocolVersion {
//	case lpv1:
//		name = "LES"
//	case lpv2:
//		name = "LES2"
//	default:
//		panic(nil)
//	}
//	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
//}

type LightDummyAPI struct{}

// Etherbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Etherbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightEthereum) APIs() []rpc.API {
	return []rpc.API{}
	/*
		return append(ptnapi.GetAPIs(s.ApiBackend), []rpc.API{
			{
				Namespace: "eth",
				Version:   "1.0",
				Service:   &LightDummyAPI{},
				Public:    true,
			}, {
				Namespace: "eth",
				Version:   "1.0",
				Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
				Public:    true,
			}, {
				Namespace: "eth",
				Version:   "1.0",
				Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
				Public:    true,
			}, {
				Namespace: "net",
				Version:   "1.0",
				Service:   s.netRPCService,
				Public:    true,
			},
		}...)
	*/
}

//func (s *LightEthereum) ResetWithGenesisBlock(gb *types.Block) {
//	s.blockchain.ResetWithGenesisBlock(gb)
//}

//func (s *LightEthereum) BlockChain() *light.LightChain      { return s.blockchain }
//func (s *LightEthereum) TxPool() *light.TxPool              { return s.txPool }
//func (s *LightEthereum) Engine() consensus.Engine           { return s.engine }
func (s *LightEthereum) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightEthereum) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightEthereum) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightEthereum) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *LightEthereum) Start(srvr *p2p.Server) error {
	//s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = ptnapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	//protocolVersion := AdvertiseProtocolVersions[0]
	//s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *LightEthereum) Stop() error {
	//s.odr.Stop()
	//if s.bloomIndexer != nil {
	//	s.bloomIndexer.Close()
	//}
	//if s.chtIndexer != nil {
	//	s.chtIndexer.Close()
	//}
	//if s.bloomTrieIndexer != nil {
	//	s.bloomTrieIndexer.Close()
	//}
	//s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.unitDb.Close()
	close(s.shutdownChan)

	return nil
}
