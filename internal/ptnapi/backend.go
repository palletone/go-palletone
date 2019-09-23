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
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/ptnjson/statistics"
	"github.com/shopspring/decimal"
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
	GetUnstableUnits() []*ptnjson.UnitSummaryJson
	// TxPool API
	SendTx(ctx context.Context, signedTx *modules.Transaction) error
	SendTxs(ctx context.Context, signedTxs []*modules.Transaction) []error
	GetPoolTransactions() (modules.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *modules.Transaction
	GetTxByTxid_back(txid string) (*ptnjson.GetTxIdResult, error)
	GetTxPoolTxByHash(hash common.Hash) (*ptnjson.TxPoolTxJson, error)
	GetPoolTxsByAddr(addr string) ([]*modules.TxPoolTransaction, error)

	//GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (int, int, int)
	TxPoolContent() (map[common.Hash]*modules.TxPoolTransaction, map[common.Hash]*modules.TxPoolTransaction)
	Queued() ([]*modules.TxPoolTransaction, error)
	SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription

	//ChainConfig() *configure.ChainConfig
	//CurrentBlock() *types.Block

	//test
	SendConsensus(ctx context.Context) error

	// wallet api
	//WalletTokens(address string) (map[string]*modules.AccountToken, error)
	//WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error)
	QueryProofOfExistenceByReference(ref string) ([]*ptnjson.ProofOfExistenceJson, error)
	// dag's get common
	GetCommon(key []byte) ([]byte, error)
	GetCommonByPrefix(prefix []byte) map[string][]byte
	SaveCommon(key, val []byte) error
	// Get Contract Api
	GetContract(contractAddr common.Address) (*ptnjson.ContractJson, error)

	//get level db
	GetUnitByHash(hash common.Hash) *modules.Unit
	GetUnitByNumber(number *modules.ChainIndex) *modules.Unit
	GetUnitsByIndex(start, end decimal.Decimal, asset string) []*modules.Unit
	GetHeaderByHash(hash common.Hash) (*modules.Header, error)
	GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error)
	// get state
	//GetHeadUnitHash() (common.Hash, error)
	//GetHeadHeaderHash() (common.Hash, error)
	//GetHeadFastUnitHash() (common.Hash, error)
	//GetCanonicalHash(number uint64) (common.Hash, error)

	// get transaction interface
	GetUnitTxsInfo(hash common.Hash) ([]*ptnjson.TxSummaryJson, error)
	GetUnitTxsHashHex(hash common.Hash) ([]string, error)
	GetTxByHash(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error)
	GetTxByReqId(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error)
	GetTxSearchEntry(hash common.Hash) (*ptnjson.TxSerachEntryJson, error)

	//TODO wangjiyou
	GetPrefix(prefix string) map[string][]byte //getprefix

	GetUtxoEntry(outpoint *modules.OutPoint) (*ptnjson.UtxoJson, error)
	GetStxoEntry(outpoint *modules.OutPoint) (*ptnjson.StxoJson, error)
	QueryDbByKey(key []byte) *ptnjson.DbRowJson
	QueryDbByPrefix(prefix []byte) []*ptnjson.DbRowJson
	//GetAddrOutput(addr string) ([]modules.Output, error)
	//------- Get addr utxo start ------//
	GetAddrOutpoints(addr string) ([]modules.OutPoint, error)
	GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error)
	GetAddrUtxos(addr string) ([]*ptnjson.UtxoJson, error)
	GetAddrUtxos2(addr string) ([]*ptnjson.UtxoJson, error)
	GetAddrRawUtxos(addr string) (map[modules.OutPoint]*modules.Utxo, error)
	GetAllUtxos() ([]*ptnjson.UtxoJson, error)
	GetAddressBalanceStatistics(token string, topN int) (*statistics.TokenAddressBalanceJson, error)
	GetAddrTxHistory(addr string) ([]*ptnjson.TxHistoryJson, error)
	GetAddrTokenFlow(addr, token string) ([]*ptnjson.TokenFlowJson, error)
	GetAssetTxHistory(asset *modules.Asset) ([]*ptnjson.TxHistoryJson, error)
	GetAssetExistence(asset string) ([]*ptnjson.ProofOfExistenceJson, error)
	//contract control
	ContractInstall(ccName string, ccPath string, ccVersion string, ccDescription, ccAbi,
		ccLanguage string) (TemplateId []byte, err error)
	ContractDeploy(templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, err error)
	ContractInvoke(deployId []byte, txid string, args [][]byte, timeout time.Duration) (rspPayload []byte, err error)
	ContractStop(deployId []byte, txid string, deleteImage bool) error

	DecodeTx(hex string) (string, error)
	DecodeJsonTx(hex string) (string, error)
	EncodeTx(jsonStr string) (string, error)

	ContractInstallReqTx(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string,
		description, abi, language string, addrs []common.Address) (reqId common.Hash, tplId []byte, err error)
	ContractDeployReqTx(from, to common.Address, daoAmount, daoFee uint64, templateId []byte, args [][]byte,
		extData []byte, timeout time.Duration) (reqId common.Hash, contractAddr common.Address, err error)
	ContractInvokeReqTx(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
		contractAddress common.Address, args [][]byte, timeout uint32) (reqId common.Hash, err error)
	SendContractInvokeReqTx(requestTx *modules.Transaction) (reqId common.Hash, err error)
	ContractInvokeReqTokenTx(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64,
		asset string, contractAddress common.Address, args [][]byte, timeout uint32) (reqId common.Hash, err error)
	ContractStopReqTx(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address,
		deleteImage bool) (reqId common.Hash, err error)

	ContractInstallReqTxFee(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string,
		description, abi, language string, addrs []common.Address) (fee float64, size float64, tm uint32, err error)

	ContractDeployReqTxFee(from, to common.Address, daoAmount, daoFee uint64, templateId []byte,
		args [][]byte, extData []byte, timeout time.Duration) (fee float64, size float64, tm uint32, err error)

	ContractInvokeReqTxFee(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
		contractAddress common.Address, args [][]byte, timeout uint32) (fee float64, size float64, tm uint32, err error)

	ContractStopReqTxFee(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address,
		deleteImage bool) (fee float64, size float64, tm uint32, err error)

	ElectionVrf(id uint32) ([]byte, error)
	UpdateJuryAccount(addr common.Address, pwd string) bool
	GetJuryAccount() []common.Address

	ContractQuery(contractId []byte, txid string, args [][]byte, timeout time.Duration) (rspPayload []byte, err error)

	TxPool() txspool.ITxPool
	Dag() dag.IDag
	SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error
	SignAndSendRequest(addr common.Address, tx *modules.Transaction) error

	TransferPtn(from, to string, amount decimal.Decimal, text *string) (*TxExecuteResult, error)
	GetKeyStore() *keystore.KeyStore

	// get tx hash by req id
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)

	GetFileInfo(filehash string) ([]*modules.FileInfo, error)

	GetAllContractTpl() ([]*ptnjson.ContractTemplateJson, error)
	GetAllContracts() ([]*ptnjson.ContractJson, error)
	GetContractsByTpl(tplId []byte) ([]*ptnjson.ContractJson, error)
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	//get contract key
	GetContractState(contractid []byte, key string) ([]byte, *modules.StateVersion, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)

	//SPV
	GetProofTxInfoByHash(txhash string) ([][]byte, error)
	ProofTransactionByHash(txhash string) (string, error)
	ProofTransactionByRlptx(rlptx [][]byte) (string, error)
	SyncUTXOByAddr(addr string) string
	StartCorsSync() (string, error)
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
			Version:   "1.0",
			Service:   NewPublicDagAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "dag",
			Version:   "1.0",
			Service:   NewPrivateDagAPI(apiBackend),
			Public:    false,
		}, {
			Namespace: "wallet",
			Version:   "1.0",
			Service:   NewPublicWalletAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "wallet",
			Version:   "1.0",
			Service:   NewPrivateWalletAPI(apiBackend),
			Public:    false,
		}, {
			Namespace: "contract",
			Version:   "1.0",
			Service:   NewPublicContractAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "contract",
			Version:   "1.0",
			Service:   NewPrivateContractAPI(apiBackend),
			Public:    false,
		}, {
			Namespace: "mediator",
			Version:   "1.0",
			Service:   NewPrivateMediatorAPI(apiBackend),
			Public:    false,
		},
		{
			Namespace: "mediator",
			Version:   "1.0",
			Service:   NewPublicMediatorAPI(apiBackend),
			Public:    true,
		},
	}
}
