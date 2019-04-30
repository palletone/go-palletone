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

package light

//type LesApiBackend struct {
//	eth *LightEthereum
//}

import (
	"context"
	"math/big"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
)

type LesApiBackend struct {
	ptn *LightPalletone
	//gpo *gasprice.Oracle
}

//func (b *LesApiBackend) ChainConfig() *params.ChainConfig {
//	return b.eth.chainConfig
//}

func (b *LesApiBackend) CurrentBlock() *modules.Unit {
	return &modules.Unit{}
}

func (b *LesApiBackend) SetHead(number uint64) {
	//b.eth.protocolManager.downloader.Cancel()
	//b.eth.blockchain.SetHead(number)
}

func (b *LesApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*modules.Header, error) {
	return nil, nil
	//if blockNr == rpc.LatestBlockNumber || blockNr == rpc.PendingBlockNumber {
	//	return b.eth.blockchain.CurrentHeader(), nil
	//}
	//
	//return b.eth.blockchain.GetHeaderByNumberOdr(ctx, uint64(blockNr))
}

func (b *LesApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*modules.Unit, error) {
	return nil, nil
	//header, err := b.HeaderByNumber(ctx, blockNr)
	//if header == nil || err != nil {
	//	return nil, err
	//}
	//return b.GetBlock(ctx, header.Hash())
}

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *modules.Header, error) {
	return nil, nil, nil
}

func (b *LesApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*modules.Unit, error) {
	return nil, nil
}

//func (b *LesApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
//	return light.GetBlockReceipts(ctx, b.eth.odr, blockHash, core.GetBlockNumber(b.eth.chainDb, blockHash))
//}
//
//func (b *LesApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
//	return light.GetBlockLogs(ctx, b.eth.odr, blockHash, core.GetBlockNumber(b.eth.chainDb, blockHash))
//}

func (b *LesApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return nil
}

//func (b *LesApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
//	state.SetBalance(msg.From(), math.MaxBig256)
//	context := core.NewEVMContext(msg, header, b.eth.blockchain, nil)
//	return vm.NewEVM(context, state, b.eth.chainConfig, vmCfg), state.Error, nil
//}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *modules.Transaction) error {
	return nil
	//return b.eth.txPool.Add(ctx, signedTx)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	b.ptn.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (modules.Transactions, error) {
	return b.ptn.txPool.GetTransactions()
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *modules.Transaction {
	return b.ptn.txPool.GetTransaction(txHash)
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.ptn.txPool.GetNonce(ctx, addr)
}

func (b *LesApiBackend) Stats() (pending int, queued int, reserve int) {
	return b.ptn.txPool.Stats(), 0, 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Hash]*modules.Transaction, map[common.Hash]*modules.Transaction) {
	return nil, nil
	//return b.ptn.txPool.Content()
}

func (b *LesApiBackend) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	return b.ptn.txPool.SubscribeTxPreEvent(ch)
}

func (b *LesApiBackend) SubscribeChainEvent(ch chan<- modules.ChainEvent) event.Subscription {
	return nil //return b.eth.dag.SubscribeChainEvent(ch)
}

func (b *LesApiBackend) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return nil //return b.eth.blockchain.SubscribeChainHeadEvent(ch)
}

func (b *LesApiBackend) SubscribeChainSideEvent(ch chan<- modules.ChainSideEvent) event.Subscription {
	return nil //return b.eth.blockchain.SubscribeChainSideEvent(ch)
}

func (b *LesApiBackend) SubscribeLogsEvent(ch chan<- []*modules.Log) event.Subscription {
	return nil //return b.eth.blockchain.SubscribeLogsEvent(ch)
}

func (b *LesApiBackend) SubscribeRemovedLogsEvent(ch chan<- modules.RemovedLogsEvent) event.Subscription {
	return nil //return b.eth.blockchain.SubscribeRemovedLogsEvent(ch)
}

func (b *LesApiBackend) Downloader() *downloader.Downloader {
	return nil //return b.eth.Downloader()
}

func (b *LesApiBackend) ProtocolVersion() int {
	return b.ptn.LesVersion() + 10000
}

func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	//return b.gpo.SuggestPrice(ctx)
	return nil, nil
}

func (b *LesApiBackend) ChainDb() ptndb.Database {
	return b.ptn.unitDb
}

func (b *LesApiBackend) EventMux() *event.TypeMux {
	return nil //return b.eth.eventMux
}

func (b *LesApiBackend) AccountManager() *accounts.Manager {
	return b.ptn.accountManager
}

//func (b *LesApiBackend) BloomStatus() (uint64, uint64) {
//	if b.eth.bloomIndexer == nil {
//		return 0, 0
//	}
//	sections, _, _ := b.eth.bloomIndexer.Sections()
//	return light.BloomTrieFrequency, sections
//}

//func (b *LesApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
//	for i := 0; i < bloomFilterThreads; i++ {
//		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.eth.bloomRequests)
//	}
//}
// General PalletOne API

//func (b *LesApiBackend)ProtocolVersion() int{
//
//}
//SuggestPrice(ctx context.Context) (*big.Int, error)
//ChainDb() ptndb.Database
//EventMux() *event.TypeMux
//AccountManager() *accounts.Manager

// BlockChain API
//SetHead(number uint64)
//HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*modules.Header, error)
//BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
//StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *modules.Header, error)
//GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
//GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
//GetTd(blockHash common.Hash) *big.Int
//SubscribeChainEvent(ch chan<- coredata.ChainEvent) event.Subscription
//SubscribeChainHeadEvent(ch chan<- coredata.ChainHeadEvent) event.Subscription
//SubscribeChainSideEvent(ch chan<- coredata.ChainSideEvent) event.Subscription
func (b *LesApiBackend) GetUnstableUnits() []*ptnjson.UnitSummaryJson {
	return nil
}

// TxPool API
//SendTx(ctx context.Context, signedTx *modules.Transaction) error
//GetPoolTransactions() (modules.Transactions, error)
//GetPoolTransaction(txHash common.Hash) *modules.Transaction
func (b *LesApiBackend) GetTxByTxid_back(txid string) (*ptnjson.GetTxIdResult, error) {
	return nil, nil
}
func (b *LesApiBackend) GetTxPoolTxByHash(hash common.Hash) (*ptnjson.TxPoolTxJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetPoolTxsByAddr(addr string) ([]*modules.TxPoolTransaction, error) {
	return nil, nil
}

//GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
//Stats() (int, int, int)
//TxPoolContent() (map[common.Hash]*modules.Transaction, map[common.Hash]*modules.Transaction)
//SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription

//ChainConfig() *configure.ChainConfig
//CurrentBlock() *types.Block

//test
func (b *LesApiBackend) SendConsensus(ctx context.Context) error {
	return nil
}

// wallet api
//WalletTokens(address string) (map[string]*modules.AccountToken, error)
//WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error)

// dag's get common
func (b *LesApiBackend) GetCommon(key []byte) ([]byte, error) {
	return nil, nil
}
func (b *LesApiBackend) GetCommonByPrefix(prefix []byte) map[string][]byte {
	return nil
}

// Get Contract Api
func (b *LesApiBackend) GetContract(hex_id string) (*modules.Contract, error) {
	return nil, nil
}

//get level db
func (b *LesApiBackend) GetUnitByHash(hash common.Hash) *modules.Unit {
	return nil
}
func (b *LesApiBackend) GetUnitByNumber(number *modules.ChainIndex) *modules.Unit {
	return nil
}
func (b *LesApiBackend) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	return nil, nil
}
func (b *LesApiBackend) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	return nil, nil
}
func (b *LesApiBackend) GetTxByReqId(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error) {
	return nil, nil
}

// get state
//GetHeadUnitHash() (common.Hash, error)
//GetHeadHeaderHash() (common.Hash, error)
//GetHeadFastUnitHash() (common.Hash, error)
//GetCanonicalHash(number uint64) (common.Hash, error)

// get transaction interface
func (b *LesApiBackend) GetUnitTxsInfo(hash common.Hash) ([]*ptnjson.TxSummaryJson, error) {
	return nil, nil
}

func (b *LesApiBackend) GetUnitTxsHashHex(hash common.Hash) ([]string, error) {
	return nil, nil
}
func (b *LesApiBackend) GetTxByHash(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetTxSearchEntry(hash common.Hash) (*ptnjson.TxSerachEntryJson, error) {
	return nil, nil
}

//TODO wangjiyou
func (b *LesApiBackend) GetPrefix(prefix string) map[string][]byte {
	return nil
}

func (b *LesApiBackend) GetUtxoEntry(outpoint *modules.OutPoint) (*ptnjson.UtxoJson, error) {
	return nil, nil
}
func (b *LesApiBackend) QueryDbByKey(key []byte) *ptnjson.DbRowJson {
	return nil
}
func (b *LesApiBackend) QueryDbByPrefix(prefix []byte) []*ptnjson.DbRowJson {
	return nil
}

//GetAddrOutput(addr string) ([]modules.Output, error)
//------- Get addr utxo start ------//
func (b *LesApiBackend) GetAddrOutpoints(addr string) ([]modules.OutPoint, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error) {
	return common.Address{}, nil
}
func (b *LesApiBackend) GetAddrUtxos(addr string) ([]*ptnjson.UtxoJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAddrRawUtxos(addr string) (map[modules.OutPoint]*modules.Utxo, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAllUtxos() ([]*ptnjson.UtxoJson, error) {
	return nil, nil
}

func (b *LesApiBackend) GetAddrTxHistory(addr string) ([]*ptnjson.TxHistoryJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAssetTxHistory(asset *modules.Asset) ([]*ptnjson.TxHistoryJson, error) {
	return nil, nil
}

//contract control
func (b *LesApiBackend) ContractInstall(ccName string, ccPath string, ccVersion string) (TemplateId []byte, err error) {
	return nil, nil
}
func (b *LesApiBackend) ContractDeploy(templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, err error) {
	return nil, nil
}

//ContractInvoke(txBytes []byte) (rspPayload []byte, err error)
func (b *LesApiBackend) ContractInvoke(deployId []byte, txid string, args [][]byte, timeout time.Duration) (rspPayload []byte, err error) {
	return nil, nil
}
func (b *LesApiBackend) ContractStop(deployId []byte, txid string, deleteImage bool) error {
	return nil
}
func (b *LesApiBackend) DecodeTx(hex string) (string, error) {
	return "", nil
}
func (b *LesApiBackend) EncodeTx(jsonStr string) (string, error) {
	return "", nil
}

func (b *LesApiBackend) ContractInstallReqTx(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string) (reqId common.Hash, tplId []byte, err error) {
	return
}
func (b *LesApiBackend) ContractDeployReqTx(from, to common.Address, daoAmount, daoFee uint64, templateId []byte, args [][]byte, timeout time.Duration) (reqId common.Hash, depId []byte, err error) {
	return
}
func (b *LesApiBackend) ContractInvokeReqTx(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int, contractAddress common.Address, args [][]byte, timeout uint32) (reqId common.Hash, err error) {
	return
}
func (b *LesApiBackend) ContractInvokeReqTokenTx(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, asset string, contractAddress common.Address, args [][]byte, timeout uint32) (reqId common.Hash, err error) {
	return
}
func (b *LesApiBackend) ContractStopReqTx(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address, deleteImage bool) (reqId common.Hash, err error) {
	return
}
func (b *LesApiBackend) ElectionVrf(id uint32) ([]byte, error) {
	return nil, nil
}
func (b *LesApiBackend) UpdateJuryAccount(addr common.Address, pwd string) bool {
	return false
}
func (b *LesApiBackend) GetJuryAccount() []common.Address {
	return nil
}

func (b *LesApiBackend) ContractQuery(contractId []byte, txid string, args [][]byte, timeout time.Duration) (rspPayload []byte, err error) {
	return nil, nil
}

func (b *LesApiBackend) Dag() dag.IDag {
	//return b.Dag()
	return nil
}

//SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error
func (b *LesApiBackend) TransferPtn(from, to string, amount decimal.Decimal, text *string) (*mp.TxExecuteResult, error) {
	return nil, nil
}
func (b *LesApiBackend) GetKeyStore() *keystore.KeyStore {
	return nil
}

// get tx hash by req id
func (b *LesApiBackend) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return common.Hash{}, nil
}

func (b *LesApiBackend) GetFileInfo(filehash string) ([]*modules.FileInfo, error) {
	return nil, nil
}

//SPV
func (b *LesApiBackend) ProofTransactionByHash(tx string) (string, error) {
	return b.ptn.ProtocolManager().ReqProofByTxHash(tx), nil

}

func (b *LesApiBackend) ProofTransactionByRlptx(rlptx string) (string, error) {
	return b.ptn.ProtocolManager().ReqProofByRlptx(rlptx), nil
}

func (b *LesApiBackend) ValidationPath(tx string) ([]byte, error) {
	return []byte("lll"), nil
}
