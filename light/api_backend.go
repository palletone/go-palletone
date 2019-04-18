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

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
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

//func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
//	header, err := b.HeaderByNumber(ctx, blockNr)
//	if header == nil || err != nil {
//		return nil, nil, err
//	}
//	return light.NewState(ctx, header, b.eth.odr), header, nil
//}

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

//func (b *LesApiBackend) GetTd(blockHash common.Hash) *big.Int {
//	return b.eth.blockchain.GetTdByHash(blockHash)
//}

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

func (b *LesApiBackend) Stats() (pending int, queued int) {
	return b.ptn.txPool.Stats(), 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Address]modules.Transactions, map[common.Address]modules.Transactions) {
	return b.ptn.txPool.Content()
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

//func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
//	return b.gpo.SuggestPrice(ctx)
//}

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
