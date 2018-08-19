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

// PtnApiBackend implements ethapi.Backend for full nodes
type PtnApiBackend struct {
	ptn *PalletOne
	//gpo *gasprice.Oracle
}

//func (b *PtnApiBackend) ChainConfig() *configure.ChainConfig {
//	return nil
//}

func (b *PtnApiBackend) SetHead(number uint64) {
	//b.ptn.protocolManager.downloader.Cancel()
	//b.ptn.dag.SetHead(number)
}

func (b *PtnApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*modules.Header, error) {
	// Pending block is only known by the miner
	return &modules.Header{}, nil
}

func (b *PtnApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *modules.Header, error) {
	return &state.StateDB{}, &modules.Header{}, nil
}

func (b *PtnApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return &big.Int{}
}

/*
func (b *PtnApiBackend) SubscribeChainEvent(ch chan<- coredata.ChainEvent) event.Subscription {
	return nil
}

func (b *PtnApiBackend) SubscribeChainHeadEvent(ch chan<- coredata.ChainHeadEvent) event.Subscription {
	return nil
}

func (b *PtnApiBackend) SubscribeChainSideEvent(ch chan<- coredata.ChainSideEvent) event.Subscription {
	return nil
}
*/
func (b *PtnApiBackend) SendConsensus(ctx context.Context) error {
	b.ptn.Engine().Engine()
	return nil
}

func (b *PtnApiBackend) SendTx(ctx context.Context, signedTx *modules.Transaction) error {
	return b.ptn.txPool.AddLocal(signedTx)
}

func (b *PtnApiBackend) GetPoolTransactions() (modules.Transactions, error) {
	pending, err := b.ptn.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs modules.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *PtnApiBackend) GetPoolTransaction(hash common.Hash) *modules.Transaction {
	return b.ptn.txPool.Get(hash)
}

//func (b *PtnApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
//	return b.ptn.txPool.State().GetNonce(addr), nil
//}

func (b *PtnApiBackend) Stats() (pending int, queued int) {
	return b.ptn.txPool.Stats()
}

func (b *PtnApiBackend) TxPoolContent() (map[common.Address]modules.Transactions, map[common.Address]modules.Transactions) {
	return b.ptn.TxPool().Content()
}

func (b *PtnApiBackend) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	return b.ptn.TxPool().SubscribeTxPreEvent(ch)
}

func (b *PtnApiBackend) Downloader() *downloader.Downloader {
	return b.ptn.Downloader()
}

func (b *PtnApiBackend) ProtocolVersion() int {
	return b.ptn.EthVersion()
}

func (b *PtnApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return &big.Int{}, nil
}

func (b *PtnApiBackend) ChainDb() ptndb.Database {
	return nil
}

func (b *PtnApiBackend) EventMux() *event.TypeMux {
	return b.ptn.EventMux()
}

func (b *PtnApiBackend) AccountManager() *accounts.Manager {
	return b.ptn.AccountManager()
}

func (b *PtnApiBackend) BloomStatus() (uint64, uint64) {
	return uint64(0), uint64(0)
}

func (b *PtnApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.ptn.bloomRequests)
	}
}

func (b *PtnApiBackend) WalletTokens(address common.Address) (map[string]*modules.AccountToken, error) {
	return b.ptn.dag.WalletTokens(address)
}

func (b *PtnApiBackend) WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
	return b.ptn.dag.WalletBalance(address, assetid, uniqueid, chainid)
}
