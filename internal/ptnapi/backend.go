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

// Package ethapi implements the general PalletOne API functions.
package ptnapi

import (
	"context"
	"math/big"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/ptnjson"
	//"github.com/palletone/go-palletone/dag/coredata"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/ptn/downloader"
	"time"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General PalletOne API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	SuggestPrice(ctx context.Context) (*big.Int, error)
	ChainDb() ptndb.Database
	EventMux() *event.TypeMux
	AccountManager() *accounts.Manager

	// BlockChain API
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*modules.Header, error)
	//BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *modules.Header, error)
	//GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
	//GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
	GetTd(blockHash common.Hash) *big.Int
	//SubscribeChainEvent(ch chan<- coredata.ChainEvent) event.Subscription
	//SubscribeChainHeadEvent(ch chan<- coredata.ChainHeadEvent) event.Subscription
	//SubscribeChainSideEvent(ch chan<- coredata.ChainSideEvent) event.Subscription

	// TxPool API
	SendTx(ctx context.Context, signedTx *modules.Transaction) error
	GetPoolTransactions() (modules.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *modules.Transaction
	GetTxByTxid_back(txid string) (*ptnjson.GetTxIdResult, error)
	//GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Hash]*modules.Transaction, map[common.Hash]*modules.Transaction)
	SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription

	//ChainConfig() *configure.ChainConfig
	//CurrentBlock() *types.Block

	//test
	SendConsensus(ctx context.Context) error

	// wallet api
	WalletTokens(address string) (map[string]*modules.AccountToken, error)
	WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error)

	// Get Contract Api
	GetContract(id common.Address) (*modules.Contract, error)

	// Get Header
	GetHeader(hash common.Hash, index uint64) (*modules.Header, error)

	// Get Unit
	GetUnit(hash common.Hash) *modules.Unit

	// Get UnitNumber
	GetUnitNumber(hash common.Hash) uint64

	// GetCanonicalHash
	GetCanonicalHash(number uint64) (common.Hash, error)

	// Get state
	GetHeadHeaderHash() (common.Hash, error)

	GetHeadUnitHash() (common.Hash, error)

	GetHeadFastUnitHash() (common.Hash, error)

	GetTrieSyncProgress() (uint64, error)

	GetUtxoEntry(outpoint *modules.OutPoint) (*ptnjson.UtxoJson, error)

	GetAddrOutput(addr string) ([]modules.Output, error)
	//------- Get addr utxo start ------//
	GetAddrOutpoints(addr string) ([]modules.OutPoint, error)
	GetAddrUtxos(addr string) ([]ptnjson.UtxoJson, error)
	GetAllUtxos() ([]ptnjson.UtxoJson, error)

	/* ---------------------save token info ------------------------*/
	SaveTokenInfo(token_info *modules.TokenInfo) (string, error)

	GetAddrTransactions(addr string) (modules.Transactions, error)
	GetAllTokenInfo() (*modules.AllTokenInfo, error)
	GetTokenInfo(key []byte) (*modules.TokenInfo, error)
	//contract control
	ContractInstall(ccName string, ccPath string, ccVersion string) (TemplateId []byte, err error)
	ContractDeploy(templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, err error)
	ContractInvoke(deployId []byte, txid string, args [][]byte, timeout time.Duration) (rspPayload []byte, err error)
	ContractStop(deployId []byte, txid string, deleteImage bool) error
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "ptn",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "ptn",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "ptn",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "ptn",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		}, {
			Namespace: "ptn",
			Version:   "1.0",
			Service:   NewPublicDagAPI(apiBackend),
			Public:    true,
		},
	}
}
