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
	"log"
	"math/big"
	"time"
        "github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/bloombits"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/accounts"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/ptn/downloader"

	//cc "github.com/palletone/go-palletone/contracts/manger"
	"encoding/hex"
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
	return b.ptn.txPool.AddLocal(txspool.TxtoTxpoolTx(b.ptn.txPool, signedTx))
}

func (b *PtnApiBackend) GetPoolTransactions() (modules.Transactions, error) {
	pending, err := b.ptn.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs modules.Transactions
	for _, batch := range pending {
		txs = append(txs, txspool.PooltxToTx(batch))
	}
	return txs, nil
}

func (b *PtnApiBackend) GetPoolTransaction(hash common.Hash) *modules.Transaction {
	return txspool.PooltxToTx(b.ptn.txPool.Get(hash))
}

func (b *PtnApiBackend) GetTxByTxid_back(txid string) (*ptnjson.GetTxIdResult, error) {
	hash := common.Hash{}
	if err := hash.SetHexString(txid); err != nil {
		return nil, err
	}
	tx, err := b.ptn.dag.GetTransactionByHash(hash)
	if err != nil {
		return nil, err
	}
        var txresult []byte
	for _, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_TEXT {
			if msg, ok := msgcopy.Payload.(*modules.TextPayload); ok {
				txresult = msg.Text
			}
		}
	}
	txOutReply := &ptnjson.GetTxIdResult{
        Txid:  txid,
        Apptype :"APP_TEXT",
        Content : txresult,
        Coinbase: true ,
	}
	return txOutReply, nil
}

//func (b *PtnApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
//	return b.ptn.txPool.State().GetNonce(addr), nil
//}

func (b *PtnApiBackend) Stats() (pending int, queued int) {
	return b.ptn.txPool.Stats()
}

func (b *PtnApiBackend) TxPoolContent() (map[common.Hash]*modules.Transaction, map[common.Hash]*modules.Transaction) {
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

func (b *PtnApiBackend) WalletTokens(address string) (map[string]*modules.AccountToken, error) {
	//comAddr, err := common.StringToAddress("P1NsG3kiKJc87M6Di6YriqHxqfPhdvxVj2B")
	comAddr, err := common.StringToAddress(address)
	if err != nil {
		return nil, err
	}
	return b.ptn.dag.WalletTokens(comAddr)
}

func (b *PtnApiBackend) WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
	comAddr, err := common.StringToAddress(address)
	if err != nil {
		return 0, err
	}
	return b.ptn.dag.WalletBalance(comAddr, assetid, uniqueid, chainid)
}

// GetContract
func (b *PtnApiBackend) GetContract(id common.Hash) (*modules.Contract, error) {
	return b.ptn.dag.GetContract(id)
}

// Get Header
func (b *PtnApiBackend) GetHeader(hash common.Hash, index uint64) (*modules.Header, error) {
	return b.ptn.dag.GetHeader(hash, index)
}

// Get Unit
func (b *PtnApiBackend) GetUnit(hash common.Hash) *modules.Unit {
	return b.ptn.dag.GetUnit(hash)
}

// Get UnitNumber
func (b *PtnApiBackend) GetUnitNumber(hash common.Hash) uint64 {
	number, err := b.ptn.dag.GetUnitNumber(hash)
	if err != nil {
		log.Println("GetUnitNumber when b.ptn.dag.GetUnitNumber", err.Error())
		return uint64(0)
	}
	return number.Index
}

// GetCanonicalHash
func (b *PtnApiBackend) GetCanonicalHash(number uint64) (common.Hash, error) {
	return b.ptn.dag.GetCanonicalHash(number)
}

// Get state
func (b *PtnApiBackend) GetHeadHeaderHash() (common.Hash, error) {
	return b.ptn.dag.GetHeadHeaderHash()
}

func (b *PtnApiBackend) GetHeadUnitHash() (common.Hash, error) {
	return b.ptn.dag.GetHeadUnitHash()
}

func (b *PtnApiBackend) GetHeadFastUnitHash() (common.Hash, error) {
	return b.ptn.dag.GetHeadFastUnitHash()
}

func (b *PtnApiBackend) GetTrieSyncProgress() (uint64, error) {
	return b.ptn.dag.GetTrieSyncProgress()
}

func (b *PtnApiBackend) GetUtxoEntry(key []byte) (*modules.Utxo, error) {
	return b.ptn.dag.GetUtxoEntry(key)
}

func (b *PtnApiBackend) GetAddrOutput(addr string) ([]modules.Output, error) {
	return b.ptn.dag.GetAddrOutput(addr)
}

func (b *PtnApiBackend) GetAddrTransactions(addr string) (modules.Transactions, error) {
	return b.ptn.dag.GetAddrTransactions(addr)
}

//contract control
func (b *PtnApiBackend) ContractInstall(ccName string, ccPath string, ccVersion string) (TemplateId []byte, err error) {
	//tempid := []byte{0x1, 0x2, 0x3}
	log.Printf("======>ContractInstall:name[%s]path[%s]version[%s]", ccName, ccPath, ccVersion)

	//payload, err := cc.Install("palletone", ccName, ccPath, ccVersion)
	payload, err := b.ptn.contract.Install("palletone", ccName, ccPath, ccVersion)

	return payload.TemplateId, err
}

func (b *PtnApiBackend) ContractDeploy(templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, err error) {
	//depid := []byte{0x4, 0x5, 0x6}
	log.Printf("======>ContractDeploy:tmId[%s]txid[%s]", hex.EncodeToString(templateId), txid)

	//depid, _, err := cc.Deploy("palletone", templateId, txid, args, timeout)
	depid, _, err := b.ptn.contract.Deploy("palletone", templateId, txid, args, timeout)
	return depid, err
}

func (b *PtnApiBackend) ContractInvoke(deployId []byte, txid string, args [][]byte, timeout time.Duration) ([]byte, error) {
	log.Printf("======>ContractInvoke:deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)

	//_, err := cc.Invoke("palletone", deployId, txid, args, timeout)
	unit, err := b.ptn.contract.Invoke("palletone", deployId, txid, args, timeout)
	//todo print rwset
	if err != nil {
		return nil, err
	}

	return unit.Payload, err
}

func (b *PtnApiBackend) ContractStop(deployId []byte, txid string, deleteImage bool) error {
	log.Printf("======>ContractStop:deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)

	//err := cc.Stop("palletone", deployId, txid, deleteImage)
	err := b.ptn.contract.Stop("palletone", deployId, txid, deleteImage)
	return err
}
