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
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/accounts"
	//"github.com/palletone/go-palletone/dag/coredata"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptnjson"
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

	// dag's get common
	GetCommon(key []byte) ([]byte, error)
	GetCommonByPrefix(prefix []byte) map[string][]byte
	// Get Contract Api
	GetContract(hex_id string) (*modules.Contract, error)

	//get level db
	GetUnitByHash(hash common.Hash) *modules.Unit
	GetUnitByNumber(number modules.ChainIndex) *modules.Unit
	GetHeaderByHash(hash common.Hash) *modules.Header
	GetHeaderByNumber(number modules.ChainIndex) *modules.Header

	// get transaction interface
	GetUnitTxsInfo(hash common.Hash) ([]*ptnjson.TransactionJson, error)
	GetUnitTxsHashHex(hash common.Hash) ([]string, error)
	GetTxByHash(hash common.Hash) (*ptnjson.TransactionJson, error)

	//TODO wangjiyou
	GetPrefix(prefix string) map[string][]byte //getprefix

	GetUtxoEntry(outpoint *modules.OutPoint) (*ptnjson.UtxoJson, error)
	QueryDbByKey(key []byte) *ptnjson.DbRowJson
	QueryDbByPrefix(prefix []byte) []*ptnjson.DbRowJson
	GetAddrOutput(addr string) ([]modules.Output, error)
	//------- Get addr utxo start ------//
	GetAddrOutpoints(addr string) ([]modules.OutPoint, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetAddrUtxos(addr string) ([]ptnjson.UtxoJson, error)
	GetAllUtxos() ([]ptnjson.UtxoJson, error)

	/* ---------------------save token info ------------------------*/
	SaveTokenInfo(token_info *modules.TokenInfo) (*ptnjson.TokenInfoJson, error)

	GetAddrTransactions(addr string) (modules.Transactions, error)
	GetAllTokenInfo() (*modules.AllTokenInfo, error)
	GetTokenInfo(key string) (*ptnjson.TokenInfoJson, error)
	//contract control
	ContractInstall(ccName string, ccPath string, ccVersion string) (TemplateId []byte, err error)
	ContractDeploy(templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, err error)
	ContractInvoke(deployId []byte, txid string, txBytes []byte, args [][]byte, timeout time.Duration) (rspPayload []byte, err error)
	ContractStop(deployId []byte, txid string, deleteImage bool) error
	DecodeTx(hex string) (string, error)
	EncodeTx(jsonStr string) (string, error)
	ContractTxReqBroadcast(deployId []byte, txid string, txBytes []byte, args [][]byte, timeout time.Duration) (rspPayload []byte, err error)

	ContractTxCreat(deployId []byte, txBytes []byte, args [][]byte, timeout time.Duration) (rspPayload []byte, err error)
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "ptn",
			Version:   "1.0",
			Service:   NewPublicPalletOneAPI(apiBackend),
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
			Namespace: "dag",
			Version:   "2.0",
			Service:   NewPublicDagAPI(apiBackend),
			Public:    true,
		},
	}
}
