// Copyright 2014 The go-palletone Authors
// This file is part of the go-palletone library.
//
// The go-palletone library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-palletone library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-palletone library. If not, see <http://www.gnu.org/licenses/>.

// Package eth implements the Ethereum protocol.
package pan

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	//"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/consensus"
	"github.com/palletone/go-palletone/core/accounts"
	//"github.com/palletone/go-palletone/consensus/clique"
	//
	"github.com/palletone/go-palletone/common/bloombits"
	"github.com/palletone/go-palletone/contracts/types"
	"github.com/palletone/go-palletone/dag/coredata"
	//"github.com/palletone/go-palletone/vm"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/pandb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/internal/ethapi"
	"github.com/palletone/go-palletone/pan/downloader"
	"github.com/palletone/go-palletone/pan/filters"
	"github.com/palletone/go-palletone/pan/gasprice"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *coredata.ChainIndexer)
}

// Ethereum implements the Ethereum full node service.
type Ethereum struct {
	config *Config
	//chainConfig *configure.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the Ethereum
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *coredata.TxPool
	blockchain      *coredata.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb pandb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         core.ConsensusEngine //consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *coredata.ChainIndexer         // Bloom indexer operating during block imports

	ApiBackend *EthApiBackend

	//miner     *miner.Miner//wangjiyou
	gasPrice  *big.Int
	etherbase common.Address

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)
}

func (s *Ethereum) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Ethereum object (including the
// initialisation of the common Ethereum object)
func New(ctx *node.ServiceContext, config *Config) (*Ethereum, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run eth.Ethereum in light sync mode, use les.LightEthereum")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	/*chainConfig, genesisHash,*/ _, _, genesisErr := coredata.SetupGenesisBlock(chainDb, config.Genesis)

	if _, ok := genesisErr.(*configure.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}

	eth := &Ethereum{
		config:  config,
		chainDb: chainDb,
		//chainConfig:    chainConfig,//wangjiyou
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx /*&config.Ethash, chainConfig,*/, chainDb), //wangjiyou  Ethash pow
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		etherbase:      config.Etherbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, configure.BloomBitsBlocks),
	}

	log.Info("Initialising Ethereum protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	eth.txPool = coredata.NewTxPool(config.TxPool) //wangjiyou

	if eth.protocolManager, err = NewProtocolManager(config.SyncMode, config.NetworkId, eth.eventMux, eth.txPool, eth.engine, chainDb); err != nil {
		log.Error("NewProtocolManager err:", err)
		return nil, err
	}

	eth.ApiBackend = &EthApiBackend{eth, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	eth.ApiBackend.gpo = gasprice.NewOracle(eth.ApiBackend, gpoParams)
	return eth, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(configure.VersionMajor<<16 | configure.VersionMinor<<8 | configure.VersionPatch),
			"gptn",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > configure.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", configure.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (pandb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*pandb.LDBDatabase); ok {
		db.Meter("eth/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Ethereum service
func CreateConsensusEngine(ctx *node.ServiceContext /*config *ethash.Config, chainConfig *configure.ChainConfig,*/, db pandb.Database) core.ConsensusEngine {
	engine := consensus.New()
	return engine
}

// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Ethereum) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.ApiBackend)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			//Service:   NewPublicMinerAPI(s),
			Public: true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			//Service:   NewPrivateMinerAPI(s),
			Public: false,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			//Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Ethereum) ResetWithGenesisBlock(gb *types.Block) {
	//s.blockchain.ResetWithGenesisBlock(gb)//wangjiyou
}

func (s *Ethereum) Etherbase() (eb common.Address, err error) {
	s.lock.RLock()
	etherbase := s.etherbase
	s.lock.RUnlock()

	if etherbase != (common.Address{}) {
		return etherbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			etherbase := accounts[0].Address

			s.lock.Lock()
			s.etherbase = etherbase
			s.lock.Unlock()

			log.Info("Etherbase automatically configured", "address", etherbase)
			return etherbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("etherbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Ethereum) SetEtherbase(etherbase common.Address) {
	self.lock.Lock()
	self.etherbase = etherbase
	self.lock.Unlock()
}

func (s *Ethereum) StopMining()    { /*s.miner.Stop()*/ }
func (s *Ethereum) IsMining() bool { /*return s.miner.Mining()*/ return true }

//func (s *Ethereum) Miner() *miner.Miner { return s.miner }//wangjiyou

func (s *Ethereum) AccountManager() *accounts.Manager { return s.accountManager }

//func (s *Ethereum) BlockChain() *coredata.BlockChain   { return s.blockchain }
func (s *Ethereum) TxPool() *coredata.TxPool           { return s.txPool }
func (s *Ethereum) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Ethereum) Engine() core.ConsensusEngine       { return s.engine }
func (s *Ethereum) ChainDb() pandb.Database            { return s.chainDb }
func (s *Ethereum) IsListening() bool                  { return true } // Always listening
func (s *Ethereum) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Ethereum) NetVersion() uint64                 { return s.networkId }
func (s *Ethereum) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Ethereum) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
	/*
		if s.lesServer == nil {
			return s.protocolManager.SubProtocols
		}
		return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
	*/
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *Ethereum) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	/*
		if s.config.LightServ > 0 {
			if s.config.LightPeers >= srvr.MaxPeers {
				return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
			}
			maxPeers -= s.config.LightPeers
		}*/
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	/*go func() {
		time.Sleep(time.Duration(15) * time.Second)
		s.Engine().Engine()
	}()*/
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *Ethereum) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	//s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	//s.miner.Stop()
	s.engine.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
