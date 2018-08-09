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

package ptn

import (
	"context"
	"math/big"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/bloombits"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/accounts"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/ptn/downloader"
)

// EthApiBackend implements ethapi.Backend for full nodes
type EthApiBackend struct {
	eth *PalletOne
	//gpo *gasprice.Oracle
}

//func (b *EthApiBackend) ChainConfig() *configure.ChainConfig {
//	return nil
//}

func (b *EthApiBackend) SetHead(number uint64) {
	b.eth.protocolManager.downloader.Cancel()
}

func (b *EthApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*modules.Header, error) {
	// Pending block is only known by the miner
	return &modules.Header{}, nil
}

func (b *EthApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *modules.Header, error) {
	return &state.StateDB{}, &modules.Header{}, nil
}

func (b *EthApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return &big.Int{}
}

/*
func (b *EthApiBackend) SubscribeChainEvent(ch chan<- coredata.ChainEvent) event.Subscription {
	return nil
}

func (b *EthApiBackend) SubscribeChainHeadEvent(ch chan<- coredata.ChainHeadEvent) event.Subscription {
	return nil
}

func (b *EthApiBackend) SubscribeChainSideEvent(ch chan<- coredata.ChainSideEvent) event.Subscription {
	return nil
}
*/
func (b *EthApiBackend) SendConsensus(ctx context.Context) error {
	b.eth.Engine().Engine()
	return nil
}

func (b *EthApiBackend) SendTx(ctx context.Context, signedTx *modules.Transaction) error {
	return b.eth.txPool.AddLocal(signedTx)
}

func (b *EthApiBackend) GetPoolTransactions() (modules.Transactions, error) {
	pending, err := b.eth.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs modules.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *EthApiBackend) GetPoolTransaction(hash common.Hash) *modules.Transaction {
	return b.eth.txPool.Get(hash)
}

//func (b *EthApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
//	return b.eth.txPool.State().GetNonce(addr), nil
//}

func (b *EthApiBackend) Stats() (pending int, queued int) {
	return b.eth.txPool.Stats()
}

func (b *EthApiBackend) TxPoolContent() (map[common.Address]modules.Transactions, map[common.Address]modules.Transactions) {
	return b.eth.TxPool().Content()
}

func (b *EthApiBackend) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	return b.eth.TxPool().SubscribeTxPreEvent(ch)
}

func (b *EthApiBackend) Downloader() *downloader.Downloader {
	return b.eth.Downloader()
}

func (b *EthApiBackend) ProtocolVersion() int {
	return b.eth.EthVersion()
}

func (b *EthApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return &big.Int{}, nil
}

func (b *EthApiBackend) ChainDb() ptndb.Database {
	return nil
}

func (b *EthApiBackend) EventMux() *event.TypeMux {
	return b.eth.EventMux()
}

func (b *EthApiBackend) AccountManager() *accounts.Manager {
	return b.eth.AccountManager()
}

func (b *EthApiBackend) BloomStatus() (uint64, uint64) {
	return uint64(0), uint64(0)
}

func (b *EthApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.eth.bloomRequests)
	}
}

func (b *EthApiBackend) WalletTokens(address common.Address) (map[string]*modules.AccountToken, error) {
	return b.eth.dag.WalletTokens(address)
}

func (b *EthApiBackend) WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
	return b.eth.dag.WalletBalance(address, assetid, uniqueid, chainid)
}
