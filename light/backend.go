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
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/internal/ptnapi"
	"github.com/palletone/go-palletone/light/cors"
)

type LightPalletone struct {
	config *ptn.Config

	//odr         *LesOdr
	//relay       *LesTxRelay
	//chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers *peerSet
	//txPool *les.TxPool
	txPool *txspool.TxPool
	//blockchain      *light.LightChain
	protocolManager *ProtocolManager

	corsProtocolManager *cors.ProtocolManager

	serverPool *serverPool
	// DB interfaces
	dag dag.IDag
	// DB interfaces
	unitDb ptndb.Database // Block chain database

	ApiBackend *LesApiBackend

	eventMux *event.TypeMux
	//engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *ptnapi.PublicNetAPI

	wg sync.WaitGroup

	txCh  chan modules.TxPreEvent
	txSub event.Subscription
}

func New(ctx *node.ServiceContext, config *ptn.Config, protocolname string) (*LightPalletone, error) {
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
	gasToken := config.Dag.GetGasToken()

	quitSync := make(chan struct{})

	lptn := &LightPalletone{
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
	}

	//lptn.relay = NewLesTxRelay(peers, leth.reqDist)
	//lptn.serverPool = newServerPool(chainDb, quitSync, &leth.wg)
	//lptn.retriever = newRetrieveManager(peers, leth.reqDist, leth.serverPool)
	//lptn.odr = NewLesOdr(chainDb, leth.chtIndexer, leth.bloomTrieIndexer, leth.bloomIndexer, leth.retriever)

	lptn.txPool = txspool.NewTxPool(config.TxPool, lptn.dag)

	if lptn.protocolManager, err = NewProtocolManager(true, lptn.peers, config.NetworkId, gasToken, nil,
		dag, lptn.eventMux, genesis, quitSync, configure.LPSProtocol); err != nil {
		return nil, err
	}

	if lptn.corsProtocolManager, err = cors.NewCorsProtocolManager(true, config.NetworkId, gasToken,
		dag, lptn.eventMux, genesis, make(chan struct{})); err != nil {
		return nil, err
	}

	lptn.ApiBackend = &LesApiBackend{lptn}
	return lptn, nil
}

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
func (s *LightPalletone) APIs() []rpc.API {
	//return []rpc.API{}

	return append(ptnapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "ptn",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "ptn",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)

}

//func (s *LightPalletone) ResetWithGenesisBlock(gb *types.Block) {
//	s.blockchain.ResetWithGenesisBlock(gb)
//}

func (s *LightPalletone) ProtocolManager() *ProtocolManager { return s.protocolManager }
func (s *LightPalletone) TxPool() *txspool.TxPool           { return s.txPool }

//func (s *LightPalletone) Engine() consensus.Engine           { return s.engine }
func (s *LightPalletone) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightPalletone) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightPalletone) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightPalletone) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

func (s *LightPalletone) CorsProtocols() []p2p.Protocol {
	return s.corsProtocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *LightPalletone) Start(srvr *p2p.Server, corss *p2p.Server) error {
	//s.startBloomHandlers()
	log.Debug("Light client mode is an experimental feature")
	s.netRPCService = ptnapi.NewPublicNetAPI(srvr, s.networkId)
	//s.protocolManager.Start(s.config.LightPeers, corss)
	s.protocolManager.Start(s.config.LightPeers, corss, nil)

	s.txCh = make(chan modules.TxPreEvent, txChanSize)
	s.txSub = s.txPool.SubscribeTxPreEvent(s.txCh)
	go s.txBroadcastLoop()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *LightPalletone) Stop() error {

	s.protocolManager.Stop()
	s.txPool.Stop()
	s.txSub.Unsubscribe() // quits txBroadcastLoop

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.unitDb.Close()
	close(s.shutdownChan)
	return nil
}

func (p *LightPalletone) GetKeyStore() *keystore.KeyStore {
	return p.accountManager.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

func (self *LightPalletone) txBroadcastLoop() {
	for {
		select {
		case event := <-self.txCh:
			log.Debug("=====ProtocolManager=====", "txBroadcastLoop event.Tx", event.Tx)
			self.protocolManager.BroadcastTx(event.Tx.Hash(), event.Tx)

			// Err() channel will be closed when unsubscribing.
		case <-self.txSub.Err():
			return
		}
	}
}

//
//func (p *LightPalletone) SignGenericTransaction(from common.Address, tx *modules.Transaction) (*modules.Transaction, error) {
//	inputpoints := make(map[modules.OutPoint][]byte)
//
//	for i := 0; i < len(tx.TxMessages); i++ {
//		// 1. 获取PaymentPayload
//		msg := tx.TxMessages[i]
//		if msg.App != modules.APP_PAYMENT {
//			continue
//		}
//
//		//
//		payload, ok := msg.Payload.(*modules.PaymentPayload)
//		if !ok {
//			log.Debug("PaymentPayload conversion error, does not match TxMessage'APP type!")
//		}
//
//		// 2. 查询每个 Input 的 PkScript
//		for _, txin := range payload.Inputs {
//			inpoint := txin.PreviousOutPoint
//			utxo, err := p.dag.GetUtxoEntry(inpoint)
//			if err != nil {
//				return nil, err
//			}
//
//			inputpoints[*inpoint] = utxo.PkScript
//		}
//	}
//
//	// 3. 使用tokenengine 和 KeyStore 给 tx 签名
//	ks := p.GetKeyStore()
//	_, err := tokenengine.SignTxAllPaymentInput(tx, tokenengine.SigHashAll, inputpoints, nil,
//		ks.GetPublicKey, ks.SignHash)
//	if err != nil {
//		return nil, err
//	}
//
//	return tx, nil
//}

// @author Albert·Gou
//func (p *LightPalletone) SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error {
//	// 3. 签名 tx
//	tx, err := p.SignGenericTransaction(addr, tx)
//	if err != nil {
//		return err
//	}
//
//	// 4. 将 tx 放入 pool
//	txPool := p.TxPool()
//	err = txPool.AddLocal(txspool.TxtoTxpoolTx(txPool, tx))
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//// @author Albert·Gou
//func (p *LightPalletone) TransferPtn(from, to string, amount decimal.Decimal, text *string) (*mp.TxExecuteResult, error) {
//	// 参数检查
//	if from == to {
//		return nil, fmt.Errorf("please don't transfer ptn to yourself: %v", from)
//	}
//
//	if amount.Cmp(decimal.New(0, 0)) != 1 {
//		return nil, fmt.Errorf("the amount of the transfer must be greater than 0")
//	}
//
//	fromAdd, err := common.StringToAddress(from)
//	if err != nil {
//		return nil, fmt.Errorf("invalid account address: %v", from)
//	}
//
//	toAdd, err := common.StringToAddress(to)
//	if err != nil {
//		return nil, fmt.Errorf("invalid account address: %v", to)
//	}
//
//	// 判断本节点是否同步完成，数据是否最新
//	//if !p.dag.IsSynced() {
//	//	return nil, fmt.Errorf("the data of this node is not synced, and can't transfer now")
//	//}
//
//	// 1. 创建交易
//	tx, fee, err := p.dag.GenTransferPtnTx(fromAdd, toAdd, ptnjson.Ptn2Dao(amount), text, p.txPool)
//	if err != nil {
//		return nil, err
//	}
//
//	// 2. 签名和发送交易
//	err = p.SignAndSendTransaction(fromAdd, tx)
//	if err != nil {
//		return nil, err
//	}
//
//	// 3. 返回执行结果
//	textStr := ""
//	if text != nil {
//		textStr = *text
//	}
//
//	res := &mp.TxExecuteResult{}
//	res.TxContent = fmt.Sprintf("Account(%s) transfer %vPTN to account(%s) with message: '%s'",
//		from, amount, to, textStr)
//	res.TxHash = tx.Hash()
//	res.TxSize = tx.Size().TerminalString()
//	res.TxFee = fmt.Sprintf("%vdao", fee)
//	res.Warning = mp.DefaultResult
//
//	return res, nil
//}
//
