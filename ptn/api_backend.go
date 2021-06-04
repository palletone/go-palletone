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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/bloombits"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/light/les"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/ptnjson/statistics"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/txspool"
	"github.com/shopspring/decimal"
	"strings"
)

const channelId = "palletone"

// PtnApiBackend implements ethapi.Backend for full nodes
type PtnApiBackend struct {
	ptn   *PalletOne
	mutex sync.Mutex
}

func (b *PtnApiBackend) Lock() {
	b.mutex.Lock()
}

func (b *PtnApiBackend) Unlock() {
	b.mutex.Unlock()
}

func (b *PtnApiBackend) EnableGasFee() bool {
	return b.ptn.config.EnableGasFee
}

func (b *PtnApiBackend) Dag() dag.IDag {
	return b.ptn.dag
}

func (b *PtnApiBackend) TxPool() txspool.ITxPool {
	return b.ptn.txPool
}

func (b *PtnApiBackend) SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error {
	return b.ptn.SignAndSendTransaction(addr, tx)
}

func (b *PtnApiBackend) GetKeyStore() *keystore.KeyStore {
	return b.ptn.GetKeyStore()
}

//func (b *PtnApiBackend) TransferPtn(from, to string, amount decimal.Decimal,
//	text *string) (*ptnapi.TxExecuteResult, error) {
//	return b.ptn.TransferPtn(from, to, amount, text)
//}

//func (b *PtnApiBackend) ChainConfig() *configure.ChainConfig {
//	return nil
//}

func (b *PtnApiBackend) SetHead(number uint64) {
	//b.ptn.protocolManager.downloader.Cancel()
	//b.ptn.dag.SetHead(number)
}

func (b *PtnApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*modules.Header, error) {
	// Packed block is only known by the miner
	return &modules.Header{}, nil
}

func (b *PtnApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB,
	*modules.Header, error) {
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

func (b *PtnApiBackend) SendTx(tx *modules.Transaction) error {
	return b.ptn.contractPorcessor.AddLocalTx(tx)
}

func (b *PtnApiBackend) SendTxs(signedTxs []*modules.Transaction) []error {
	result := []error{}
	for _, tx := range signedTxs {
		err := b.ptn.txPool.AddLocal(tx)
		if err != nil {
			result = append(result, err)
		}
	}
	return result
}

func (b *PtnApiBackend) GetPoolTransactions() (modules.Transactions, error) {
	pending, err := b.ptn.txPool.Packed()
	if err != nil {
		return nil, err
	}
	var txs modules.Transactions
	for _, batch := range pending {
		for _, tx := range batch {
			txs = append(txs, tx.Tx)
		}
	}
	return txs, nil
}

func (b *PtnApiBackend) GetPoolTransaction(hash common.Hash) *modules.Transaction {
	tx, _ := b.ptn.txPool.GetTx(hash)
	return tx.Tx
}

func (b *PtnApiBackend) GetTxByTxid_back(txid string) (*ptnjson.GetTxIdResult, error) {
	hash := common.Hash{}
	if err := hash.SetHexString(txid); err != nil {
		return nil, err
	}
	tx, err := b.ptn.dag.GetTransaction(hash)
	if err != nil {
		return nil, err
	}
	//var hex_hash string
	//if unitHash != (common.Hash{}) {
	//	hex_hash = unitHash.String()
	//}
	var txresult []byte
	for _, msgcopy := range tx.TxMessages() {
		if msgcopy.App == modules.APP_DATA {
			if msg, ok := msgcopy.Payload.(*modules.DataPayload); ok {
				txresult = msg.MainData
			}
		}
	}
	txOutReply := &ptnjson.GetTxIdResult{
		Txid:     txid,
		Apptype:  "APP_DATA",
		Content:  txresult,
		Coinbase: true,
		UnitHash: tx.UnitHash.String(),
	}
	return txOutReply, nil
}

func (b *PtnApiBackend) GetChainParameters() *core.ChainParameters {
	return b.Dag().GetChainParameters()
}

func (b *PtnApiBackend) Status() (int, int, int) {
	return b.ptn.txPool.Status()
}
func (b *PtnApiBackend) TxPoolClear() {
	b.ptn.TxPool().Clear()
}
func (b *PtnApiBackend) TxPoolContent() (map[common.Hash]*txspool.TxPoolTransaction,
	map[common.Hash]*txspool.TxPoolTransaction) {
	return b.ptn.TxPool().Content()
}
func (b *PtnApiBackend) TxPoolOrphan() ([]*txspool.TxPoolTransaction, error) {
	return b.ptn.TxPool().Orphan()
}
func (b *PtnApiBackend) TxPoolPacked() (map[common.Hash][]*txspool.TxPoolTransaction, error){
	return b.ptn.TxPool().Packed()
}
func (b *PtnApiBackend) TxPoolUnpack() ([]*txspool.TxPoolTransaction, error){
	return b.ptn.TxPool().Unpack()
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
	//for i := 0; i < bloomFilterThreads; i++ {
	//	go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.ptn.bloomRequests)
	//}
}

//func (b *PtnApiBackend) WalletTokens(address string) (map[string]*modules.AccountToken, error) {
//	//comAddr, err := common.StringToAddress("P1NsG3kiKJc87M6Di6YriqHxqfPhdvxVj2B")
//	comAddr, err := common.StringToAddress(address)
//	if err != nil {
//		return nil, err
//	}
//	return b.ptn.dag.WalletTokens(comAddr)
//}
//

//in = contractId1:v1;contractId2:v2;contractId3:v3
func getContractSysVersion(contractAddr common.Address, in string) string { //contractId []byte
	cvs := strings.Split(in, ";")
	log.Debugf("cvs len[%d]:%v, adr:%s", len(cvs), cvs, contractAddr.String())
	for _, ls := range cvs {
		cv := strings.Split(ls, ":")
		if len(cv) > 1 {
			if contractAddr.String() == cv[0] {
				log.Debugf("getContractSysVersion ok, version:%s", cv[1])
				return cv[1]
			}
		}
	}
	return ""
}

// GetContract
func (b *PtnApiBackend) GetContract(addr common.Address) (*ptnjson.ContractJson, error) {
	contract, err := b.ptn.dag.GetContract(addr.Bytes())
	if err != nil {
		return nil, err
	}
	contractAddr := common.NewAddress(contract.ContractId, common.ContractHash)
	if contractAddr.IsSystemContractAddress() {
		contractVersion := getContractSysVersion(contractAddr, b.ptn.dag.GetChainParameters().ContractSystemVersion)
		if contractVersion == "" {
			contractVersion = "ptn01"
		}
		contract.Version = contractVersion
	}

	cjson := ptnjson.ConvertContract2Json(contract)
	cjson.Template = ptnjson.GetSysContractABI(addr.String())
	if nil == cjson.Template {
		tpl, err := b.ptn.dag.GetContractTpl(contract.TemplateId)
		if err != nil {
			return cjson, nil
		}
		cjson.Template = ptnjson.ConvertContractTemplate2Json(tpl)
	}
	return cjson, nil
}
func (b *PtnApiBackend) QueryDbByKey(key []byte) *ptnjson.DbRowJson {
	val, err := b.ptn.dag.QueryDbByKey(key)
	if err != nil {

		return nil
	}
	return ptnjson.NewDbRowJson(key, val)
}
func (b *PtnApiBackend) QueryDbByPrefix(prefix []byte) []*ptnjson.DbRowJson {
	vals, err := b.ptn.dag.QueryDbByPrefix(prefix)
	if err != nil {

		return nil
	}
	result := []*ptnjson.DbRowJson{}
	for _, val := range vals {
		j := ptnjson.NewDbRowJson(val.Key, val.Value)
		result = append(result, j)
	}
	return result
}

// Get Header
func (b *PtnApiBackend) GetHeader(hash common.Hash) (*modules.Header, error) {
	return b.ptn.dag.GetHeaderByHash(hash)
}

// Get Unit
func (b *PtnApiBackend) GetUnit(hash common.Hash) *modules.Unit {
	u, _ := b.ptn.dag.GetUnitByHash(hash)
	return u
}

// Get UnitNumber
func (b *PtnApiBackend) GetUnitNumber(hash common.Hash) uint64 {
	number, err := b.ptn.dag.GetUnitNumber(hash)
	if err != nil {
		log.Warnf("GetUnitNumber when b.ptn.dag.GetUnitNumber,%s", err.Error())
		return uint64(0)
	}
	return number.Index
}

//
func (b *PtnApiBackend) GetAssetTxHistory(asset *modules.Asset) ([]*ptnjson.TxHistoryJson, error) {
	txs, err := b.ptn.dag.GetAssetTxHistory(asset)
	if err != nil {
		return nil, err
	}
	txjs := []*ptnjson.TxHistoryJson{}
	for _, tx := range txs {
		txj := ptnjson.ConvertTx2HistoryJson(tx, b.ptn.dag.GetTxOutput)
		txjs = append(txjs, txj)
	}
	return txjs, nil
}

func (b *PtnApiBackend) GetAssetExistence(asset string) ([]*ptnjson.ProofOfExistenceJson, error) {
	poes, err := b.ptn.dag.GetAssetReference([]byte(asset))
	if err != nil {
		return nil, err
	}
	result := []*ptnjson.ProofOfExistenceJson{}
	for _, poe := range poes {
		j := ptnjson.ConvertProofOfExistence2Json(poe)
		result = append(result, j)
	}
	return result, nil
}

// Get state
//func (b *PtnApiBackend) GetHeadHeaderHash() (common.Hash, error) {
//	return b.ptn.dag.GetHeadHeaderHash()
//}
//
//func (b *PtnApiBackend) GetHeadUnitHash() (common.Hash, error) {
//	return b.ptn.dag.GetHeadUnitHash()
//}
//
//func (b *PtnApiBackend) GetHeadFastUnitHash() (common.Hash, error) {
//	return b.ptn.dag.GetHeadFastUnitHash()
//}

func (b *PtnApiBackend) MemdagInfos() (*modules.MemdagInfos, error) {
	return b.ptn.dag.MemdagInfos()
}
func (b *PtnApiBackend) GetTrieSyncProgress() (uint64, error) {
	return b.ptn.dag.GetTrieSyncProgress()
}
func (b *PtnApiBackend) GetUnstableUnits() []*ptnjson.UnitSummaryJson {
	units := b.ptn.dag.GetUnstableUnits()
	result := make([]*ptnjson.UnitSummaryJson, len(units))
	for i, unit := range units {
		result[i] = ptnjson.ConvertUnit2SummaryJson(unit)
	}
	return result
}
func (b *PtnApiBackend) GetUnitByHash(hash common.Hash) *modules.Unit {
	unit, err := b.ptn.dag.GetUnitByHash(hash)
	if err != nil {
		return nil
	}
	return unit
}
func (b *PtnApiBackend) GetUnitByNumber(number *modules.ChainIndex) *modules.Unit {
	unit, err := b.ptn.dag.GetUnitByNumber(number)
	if err != nil {
		return nil
	}
	return unit
}
func (b *PtnApiBackend) GetUnitsByIndex(start, end decimal.Decimal, asset string) []*modules.Unit {
	index1 := uint64(start.IntPart())
	index2 := uint64(end.IntPart())
	units := make([]*modules.Unit, 0)
	token, _, err := modules.String2AssetId(asset)
	if err != nil {
		log.Info("the asset str is not correct token string.")
		return nil
	}
	for i := index1; i <= index2; i++ {
		number := new(modules.ChainIndex)
		number.Index = i
		number.AssetID = token
		unit, err := b.ptn.dag.GetUnitByNumber(number)
		if unit == nil || err != nil {
			log.Info("PublicBlockChainAPI", "GetUnitByNumber GetUnitByNumber is nil number:", number.String(),
				"error", err)
		}
		//jsonUnit := ptnjson.ConvertUnit2Json(unit, s.b.Dag().GetUtxoEntry)
		units = append(units, unit)
	}
	return units
}

func (b *PtnApiBackend) GetUnitTxsInfo(hash common.Hash) ([]*ptnjson.TxSummaryJson, error) {
	header, err := b.ptn.dag.GetHeaderByHash(hash)
	if err != nil {
		return nil, err
	}
	txs, err := b.ptn.dag.GetUnitTransactions(hash)
	if err != nil {
		return nil, err
	}
	txs_json := make([]*ptnjson.TxSummaryJson, 0)

	for txIdx, tx := range txs {
		txs_json = append(txs_json, ptnjson.ConvertTx2SummaryJson(tx, hash, header.GetNumber().Index, header.Timestamp(),
			uint64(txIdx), b.ptn.dag.GetTxOutput))
	}
	return txs_json, nil
}

func (b *PtnApiBackend) GetUnitTxsHashHex(hash common.Hash) ([]string, error) {
	hashs, err := b.ptn.dag.GetUnitTxsHash(hash)
	if err != nil {
		return nil, err
	}
	hexs := make([]string, 0)
	for _, hash := range hashs {
		hexs = append(hexs, hash.String())
	}
	return hexs, nil
}

func (b *PtnApiBackend) GetTxByHash(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error) {
	tx, err := b.ptn.dag.GetTransaction(hash)
	if err != nil {
		return nil, err
	}
	return ptnjson.ConvertTxWithUnitInfo2FullJson(tx, b.ptn.dag.GetTxOutput), nil
}
func (b *PtnApiBackend) GetTxByReqId(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error) {
	tx, err := b.ptn.dag.GetTxByReqId(hash)
	if err != nil {
		return nil, err
	}
	return ptnjson.ConvertTxWithUnitInfo2FullJson(tx, b.ptn.dag.GetTxOutput), nil
}
func (b *PtnApiBackend) GetTxSearchEntry(hash common.Hash) (*ptnjson.TxSerachEntryJson, error) {
	entry, err := b.ptn.dag.GetTxSearchEntry(hash)
	if err != nil {
		return nil, err
	}
	return ptnjson.ConvertTxEntry2Json(entry), err
}
func (b *PtnApiBackend) GetTxPackInfo(txHash common.Hash) (*ptnjson.TxPackInfoJson, error) {
	entry, err := b.ptn.dag.GetTxPackInfo(txHash)
	if err != nil {
		return nil, err
	}
	return ptnjson.ConvertTxPackInfoJson(entry), err
}

// GetPoolTxByHash return a json of the tx in pool.
func (b *PtnApiBackend) GetTxPoolTxByHash(hash common.Hash) (*ptnjson.TxPoolTxJson, error) {
	tx, err := b.ptn.txPool.GetTx(hash)
	if err != nil {
		return nil, fmt.Errorf("the tx[%s] is not exist in txppol.", hash.String())
	}
	return ptnjson.ConvertTxPoolTx2Json(tx, tx.UnitHash), nil
}

func (b *PtnApiBackend) GetUnpackedTxsByAddr(addr string) ([]*txspool.TxPoolTransaction, error) {
	address, _ := common.StringToAddress(addr)
	tx, err := b.ptn.txPool.GetUnpackedTxsByAddr(address)
	return tx, err
}
func (b *PtnApiBackend) GetPoolAddrUtxos(addr common.Address, token *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, error) {
	return b.ptn.txPool.GetAddrUtxos(addr, token)
}

func (b *PtnApiBackend) QueryProofOfExistenceByReference(ref string) ([]*ptnjson.ProofOfExistenceJson, error) {
	poes, err := b.ptn.dag.QueryProofOfExistenceByReference([]byte(ref))
	if err != nil {
		return nil, err
	}
	result := []*ptnjson.ProofOfExistenceJson{}
	for _, poe := range poes {
		j := ptnjson.ConvertProofOfExistence2Json(poe)
		result = append(result, j)
	}
	return result, nil
}

func (b *PtnApiBackend) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	return b.ptn.dag.GetHeaderByHash(hash)
}

func (b *PtnApiBackend) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	return b.ptn.dag.GetHeaderByNumber(number)
}

func (b *PtnApiBackend) GetPrefix(prefix string) map[string][]byte {
	return b.ptn.dag.GetCommonByPrefix([]byte(prefix), false)
} //getprefix
func (b *PtnApiBackend) utxoQuery(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	preUtxo, err := b.ptn.txPool.GetUtxoFromAll(outpoint)
	if err == nil {
		return preUtxo, nil
	}
	return b.ptn.dag.GetUtxoEntry(outpoint)
}
func (b *PtnApiBackend) GetUtxoEntry(outpoint *modules.OutPoint) (*ptnjson.UtxoJson, error) {
	//This function query from txpool first, not exist, then query from leveldb.
	utxo, err := b.utxoQuery(outpoint)
	if err != nil {
		log.Errorf("Utxo not found in txpool and leveldb, key:%s", outpoint.String())
		return nil, err
	}
	ujson := ptnjson.ConvertUtxo2Json(outpoint, utxo)
	return ujson, nil
}

func (b *PtnApiBackend) GetStxoEntry(outpoint *modules.OutPoint) (*ptnjson.StxoJson, error) {
	stxo, err := b.ptn.dag.GetStxoEntry(outpoint)
	if err != nil {
		return nil, err
	}
	j := ptnjson.ConvertStxo2Json(outpoint, stxo)
	return j, nil
}

func (b *PtnApiBackend) GetAddrOutpoints(addr string) ([]modules.OutPoint, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}

	return b.ptn.dag.GetAddrOutpoints(address)
}
func (b *PtnApiBackend) GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error) {
	address, err := b.ptn.dag.GetAddrByOutPoint(outPoint)
	return address, err
}

func (b *PtnApiBackend) GetDagAddrUtxos(addr string) ([]*ptnjson.UtxoJson, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}

	utxos, _ := b.ptn.dag.GetAddrUtxos(address)
	result := covertUtxos2Json(utxos)
	return result, nil
}
func (b *PtnApiBackend) GetAddrUtxoTxs(addr string) ([]*ptnjson.TxWithUnitInfoJson, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}

	txs, _ := b.ptn.dag.GetAddrUtxoTxs(address)
	result := make([]*ptnjson.TxWithUnitInfoJson, 0, len(txs))
	for _, tx := range txs {

		txjson := ptnjson.ConvertTxWithUnitInfo2FullJson(tx, b.ptn.dag.GetUtxoEntry)
		result = append(result, txjson)
	}
	return result, nil
}

func covertUtxos2Json(utxos map[modules.OutPoint]*modules.Utxo) []*ptnjson.UtxoJson {
	result := []*ptnjson.UtxoJson{}
	for o, u := range utxos {
		o := o
		ujson := ptnjson.ConvertUtxo2Json(&o, u)
		result = append(result, ujson)
	}
	return result
}
func (b *PtnApiBackend) GetAddrUtxos2(addr string) ([]*ptnjson.UtxoJson, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}

	allUtxos, _ := b.ptn.dag.GetAddrUtxos(address)
	stbUtxos, _ := b.ptn.dag.GetAddrStableUtxos(address)
	//根据稳定UTXO和不稳定UTXO的对比，更新UTXO的FlagStatus
	result := make(map[modules.OutPoint]*ptnjson.UtxoJson)
	for op, utxo := range allUtxos {
		outpoint := op
		_, ok := stbUtxos[outpoint]
		ujson := ptnjson.ConvertUtxo2Json(&outpoint, utxo)
		if !ok { //不稳定的UTXO
			ujson.FlagStatus = "Unstable"
		} else {
			ujson.FlagStatus = "Normal"
		}
		result[outpoint] = ujson
	}
	//在Stable里面有，但是在All里面没有，说明已经被花费
	for op, utxo := range stbUtxos {
		outpoint := op
		_, ok := allUtxos[outpoint]
		if !ok { //已经被花费的UTXO
			ujson := ptnjson.ConvertUtxo2Json(&outpoint, utxo)
			ujson.FlagStatus = "Spending"
			result[outpoint] = ujson
		}
	}
	return covertUtxosJsonMap(result), nil
}
func covertUtxosJsonMap(utxos map[modules.OutPoint]*ptnjson.UtxoJson) []*ptnjson.UtxoJson {
	result := []*ptnjson.UtxoJson{}
	for _, u := range utxos {

		result = append(result, u)
	}
	return result
}

func (b *PtnApiBackend) GetAddrRawUtxos(addr string) (map[modules.OutPoint]*modules.Utxo, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	return b.ptn.dag.GetAddrUtxos(address)
}

func (b *PtnApiBackend) GetAllUtxos() ([]*ptnjson.UtxoJson, error) {
	utxos, err := b.ptn.dag.GetAllUtxos()
	if err != nil {
		return nil, err
	}
	result := []*ptnjson.UtxoJson{}
	for o, u := range utxos {
		o := o
		ujson := ptnjson.ConvertUtxo2Json(&o, u)
		result = append(result, ujson)
	}
	return result, nil
}
func (b *PtnApiBackend) GetAddrTokenFlow(addr, token string) ([]*ptnjson.TokenFlowJson, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	asset, err := modules.StringToAsset(token)
	if err != nil {
		return nil, err
	}
	txs, err := b.ptn.dag.GetAddrTransactions(address)
	if err != nil {
		return nil, err
	}
	//按时间从旧到新排序
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Timestamp < txs[j].Timestamp
	})
	txjs := []*ptnjson.TokenFlowJson{}
	balance := uint64(0)
	for _, tx := range txs {
		txj, newbalance := ptnjson.ConvertTx2TokenFlowJson(address, asset, balance, tx, b.ptn.dag.GetTxOutput)
		txjs = append(txjs, txj...)
		//for _, t := range txj {
		//	txjs = append(txjs, t)
		//}
		balance = newbalance
	}
	return txjs, nil
}
func (b *PtnApiBackend) GetAddrTxHistory(addr string) ([]*ptnjson.TxHistoryJson, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	txs, err := b.ptn.dag.GetAddrTransactions(address)
	if err != nil {
		return nil, err
	}
	//按时间从旧到新排序
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Timestamp < txs[j].Timestamp
	})
	txjs := []*ptnjson.TxHistoryJson{}
	for _, tx := range txs {
		txj := ptnjson.ConvertTx2HistoryJson(tx, b.ptn.dag.GetTxOutput)
		txjs = append(txjs, txj)
	}
	return txjs, nil
}
func (b *PtnApiBackend) GetContractInvokeHistory(addr string) ([]*ptnjson.ContractInvokeHistoryJson, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	txs, err := b.ptn.dag.GetAddrTransactions(address)
	if err != nil {
		return nil, err
	}
	//按时间从旧到新排序
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Timestamp < txs[j].Timestamp
	})
	txjs := []*ptnjson.ContractInvokeHistoryJson{}
	for _, tx := range txs {
		txj := ptnjson.ConvertTx2ContractInvokeHistoryJson(tx, b.ptn.dag.GetTxOutput)
		txjs = append(txjs, txj)
	}
	return txjs, nil
}

func (b *PtnApiBackend)ContractEventBroadcast(event jury.ContractEvent, local bool) {
	b.ptn.ContractBroadcast(event, local)
}

func (b *PtnApiBackend) ContractInstall(ccName string, ccPath string, ccVersion string, ccDescription, ccAbi,
ccLanguage string) ([]byte, error) {
	//channelId := "palletone"
	payload, err := b.ptn.contract.Install(channelId, ccName, ccPath, ccVersion, ccDescription, ccAbi, ccLanguage)
	if err != nil {
		return nil, err
	}
	return payload.TemplateId, err
}

func (b *PtnApiBackend) ContractDeploy(templateId []byte, txid string, args [][]byte,
	timeout time.Duration) (deployId []byte, err error) {
	log.Debugf("======>ContractDeploy:tmId[%s]txid[%s]", hex.EncodeToString(templateId), txid)
	//channelId := "palletone"
	_, payload, err := b.ptn.contract.Deploy(rwset.RwM, channelId, templateId, txid, args, timeout)
	if err != nil {
		return nil, err
	}
	return payload.ContractId, err
}

func (b *PtnApiBackend) ContractInvoke(deployId []byte, txid string, args [][]byte,
	timeout time.Duration) ([]byte, error) {
	log.Debugf("======>ContractInvoke:deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)
	//channelId := "palletone"
	ctx := &contracts.ContractProcessContext{RwM: rwset.RwM, Dag: b.Dag()}
	unit, err := b.ptn.contract.Invoke(ctx, channelId, deployId, txid, args, timeout)
	if err != nil {
		return nil, err
	}
	return unit.Payload, err
}

func (b *PtnApiBackend) ContractStop(deployId []byte, txid string, deleteImage bool) error {
	log.Debugf("======>ContractStop:deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)
	_, err := b.ptn.contract.Stop(rwset.RwM, "palletone", deployId, txid, deleteImage)
	return err
}

func (b *PtnApiBackend) SignAndSendRequest(addr common.Address, tx *modules.Transaction) error {
	_, err := b.ptn.contractPorcessor.SignAndExecuteAndSendRequest(addr, tx)
	return err
}

func (b *PtnApiBackend) SendContractInvokeReqTx(requestTx *modules.Transaction) (common.Hash, error) {
	//Devin：连续合约调用，这里验证不过，先注释
	//if !b.ptn.contractPorcessor.CheckTxValid(requestTx) {
	//	err := fmt.Sprintf("ProcessContractEvent, event Tx is invalid, txId:%s", requestTx.Hash().String())
	//	return common.Hash{}, errors.New(err)
	//}
	var ele *modules.ElectionNode
	if !requestTx.IsSystemContract() {
		ele, _ = b.Dag().GetContractJury(requestTx.GetContractId())
	}
	go b.ptn.ContractBroadcast(jury.ContractEvent{Ele: ele, CType: jury.CONTRACT_EVENT_EXEC, Tx: requestTx}, true)
	err := b.Dag().SaveLocalTx(requestTx)
	if err != nil {
		log.Errorf("Try to save request[%s] error:%s", requestTx.Hash().String(), err.Error())
	}
	return requestTx.RequestHash(), nil
}

func (b *PtnApiBackend) ContractInstallReqTxFee(from, to common.Address, daoAmount, daoFee uint64, tplName,
path, version string, description, abi, language string, addrs []common.Address) (fee float64, size float64, tm uint32,
	err error) {
	return b.ptn.contractPorcessor.ContractInstallReqFee(from, to, daoAmount, daoFee, tplName, path,
		version, description, abi, language, true, addrs)
}
func (b *PtnApiBackend) ContractDeployReqTxFee(from, to common.Address, daoAmount, daoFee uint64, templateId []byte,
	args [][]byte, extData []byte, timeout time.Duration) (fee float64, size float64, tm uint32, err error) {
	return b.ptn.contractPorcessor.ContractDeployReqFee(from, to, daoAmount, daoFee, templateId, args, extData, timeout)
}
func (b *PtnApiBackend) ContractInvokeReqTxFee(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	contractAddress common.Address, args [][]byte, timeout uint32) (fee float64, size float64, tm uint32, err error) {
	return b.ptn.contractPorcessor.ContractInvokeReqFee(from, to, daoAmount, daoFee, certID, contractAddress, args, timeout)
}
func (b *PtnApiBackend) ContractStopReqTxFee(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address,
	deleteImage bool) (fee float64, size float64, tm uint32, err error) {
	return b.ptn.contractPorcessor.ContractStopReqFee(from, to, daoAmount, daoFee, contractId, deleteImage)
}
func (b *PtnApiBackend) ContractQuery(id []byte, args [][]byte, timeout time.Duration) (rspPayload []byte, err error) {
	return b.ptn.contractPorcessor.ContractQuery(id, args, timeout)
}

func (b *PtnApiBackend) ElectionVrf(id uint32) ([]byte, error) {
	return b.ptn.contractPorcessor.ElectionVrfReq(id)
}
func (b *PtnApiBackend) UpdateJuryAccount(addr common.Address, pwd string) bool {
	return b.ptn.contractPorcessor.UpdateJuryAccount(addr, pwd)
}

func (b *PtnApiBackend) GetJuryAccount() []common.Address {
	return b.ptn.contractPorcessor.GetLocalJuryAddrs()
}
func (b *PtnApiBackend) SaveCommon(key, val []byte) error {
	return b.ptn.dag.SaveCommon(key, val)
}
func (b *PtnApiBackend) GetCommon(key []byte, stableDb bool) ([]byte, error) {
	return b.ptn.dag.GetCommon(key, stableDb)
}

func (b *PtnApiBackend) GetCommonByPrefix(prefix []byte, stableDb bool) map[string][]byte {
	return b.ptn.dag.GetCommonByPrefix(prefix, stableDb)
}
func (b *PtnApiBackend) GetAllData() ([][]byte, [][]byte) {
	return b.ptn.dag.GetAllData()
}
func (b *PtnApiBackend) DecodeTx(hexStr string) (string, error) {
	tx := &modules.Transaction{}
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	err = rlp.DecodeBytes(data, tx)
	if err != nil {
		return "", err
	}
	txjson := ptnjson.ConvertTx2FullJson(tx, b.Dag().GetUtxoEntry)
	jsondata, err := json.Marshal(txjson)
	return string(jsondata), err
}

func (b *PtnApiBackend) DecodeJsonTx(hexStr string) (string, error) {
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", errors.New("Decode Signedtx is invalid")
	}
	var btxjson []byte
	if err := rlp.DecodeBytes(decoded, &btxjson); err != nil {
		return "", errors.New("RLP Decode To Byte is invalid")
	}
	tx := modules.NewTransaction(make([]*modules.Message, 0))
	err = json.Unmarshal(btxjson, tx)
	if err != nil {
		return "", errors.New("Json Unmarshal To Tx is invalid")
	}
	txjson := ptnjson.ConvertTx2FullJson(tx, b.Dag().GetUtxoEntry)
	json, err := json.Marshal(txjson)
	return string(json), err
}

func (b *PtnApiBackend) EncodeTx(jsonStr string) (string, error) {
	txjson := &ptnjson.TxJson{}
	err := json.Unmarshal([]byte(jsonStr), txjson)
	if err != nil {
		return "", err
	}
	tx := ptnjson.ConvertJson2Tx(txjson)
	data, err := rlp.EncodeToBytes(tx)

	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), err
}

func (b *PtnApiBackend) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return b.ptn.dag.GetTxHashByReqId(reqid)
}

func (b *PtnApiBackend) GetFileInfo(filehash string) ([]*modules.ProofOfExistencesInfo, error) {
	return b.ptn.dag.GetFileInfo([]byte(filehash))
}

func (b *PtnApiBackend) GetProofOfExistencesByMaindata(maindata string) ([]*modules.ProofOfExistencesInfo, error) {
	return b.ptn.dag.GetFileInfo([]byte(maindata))
}

//SPV
//`json:"unit_hash"`
type proofTxInfo struct {
	Headerhash []byte       `json:"header_hash"`
	Triekey    []byte       `json:"trie_key"`
	Triepath   les.NodeList `json:"trie_path"`
}

func (s *PtnApiBackend) GetProofTxInfoByHash(strtxhash string) ([][]byte, error) {
	txhash := common.Hash{}
	txhash.SetHexString(strtxhash)
	tx, err := s.Dag().GetTransaction(txhash)
	if err != nil {
		return [][]byte{[]byte("Have not this transaction")}, err
	}
	unit, err := s.Dag().GetUnitByHash(tx.UnitHash)
	if err != nil {
		return [][]byte{[]byte("Have not exsit Unit")}, err
	}
	index := 0
	for _, tx := range unit.Txs {
		if tx.Hash() == txhash {
			break
		}
		index++
	}

	info := proofTxInfo{}
	info.Headerhash = unit.UnitHeader.Hash().Bytes()
	keybuf := new(bytes.Buffer)
	rlp.Encode(keybuf, uint(index))
	info.Triekey = keybuf.Bytes()

	tri, trieRootHash := core.GetTrieInfo(unit.Txs)

	if err := tri.Prove(info.Triekey, 0, &info.Triepath); err != nil {
		log.Debug("Light PalletOne", "GetProofTxInfoByHash err", err, "key", info.Triekey,
			"proof", info.Triepath)
		return [][]byte{[]byte(fmt.Sprintf("Get Trie err %v", err))}, err
	}

	if trieRootHash.String() != unit.UnitHeader.TxRoot().String() {
		log.Debug("Light PalletOne", "GetProofTxInfoByHash hash is not equal.trieRootHash.String()",
			trieRootHash.String(), "unit.UnitHeader.TxRoot.String()", unit.UnitHeader.TxRoot().String())
		return [][]byte{[]byte("trie root hash is not equal")}, errors.New("hash not equal")
	}

	data := [][]byte{}
	data = append(data, info.Headerhash)
	data = append(data, info.Triekey)

	path, err := rlp.EncodeToBytes(info.Triepath)
	if err != nil {
		return nil, err
	}
	data = append(data, path)

	return data, nil
}

func (s *PtnApiBackend) ProofTransactionByHash(tx string) (string, error) {
	return "", nil
}

func (s *PtnApiBackend) ProofTransactionByRlptx(rlptx [][]byte) (string, error) {
	return "", nil
}

func (b *PtnApiBackend) SyncUTXOByAddr(addr string) string {
	return "Error"
}

func (b *PtnApiBackend) StartCorsSync() (string, error) {
	if b.ptn.corsServer != nil {
		return b.ptn.corsServer.StartCorsSync()
	}
	return "cors server is nil", errors.New("cors server is nil")
}

func (b *PtnApiBackend) GetAllContractTpl() ([]*ptnjson.ContractTemplateJson, error) {
	tpls, err := b.ptn.dag.GetAllContractTpl()
	if err != nil {
		return nil, err
	}
	jsons := []*ptnjson.ContractTemplateJson{}
	for _, tpl := range tpls {
		jsons = append(jsons, ptnjson.ConvertContractTemplate2Json(tpl))
	}
	return jsons, nil
}
func (b *PtnApiBackend) GetAllContracts() ([]*ptnjson.ContractJson, error) {
	contracts, err := b.ptn.dag.GetAllContracts()
	if err != nil {
		return nil, err
	}
	jsons := []*ptnjson.ContractJson{}
	for _, c := range contracts {
		contractAddr := common.NewAddress(c.ContractId, common.ContractHash)
		if contractAddr.IsSystemContractAddress() {
			contractVersion := getContractSysVersion(contractAddr, b.ptn.dag.GetChainParameters().ContractSystemVersion)
			if contractVersion == "" {
				contractVersion = "ptn01"
			}
			c.Version = contractVersion
		}

		jsons = append(jsons, ptnjson.ConvertContract2Json(c))
	}
	return jsons, nil
}
func (b *PtnApiBackend) GetContractsByTpl(tplId []byte) ([]*ptnjson.ContractJson, error) {
	contracts, err := b.ptn.dag.GetContractsByTpl(tplId)
	if err != nil {
		return nil, err
	}
	jsons := []*ptnjson.ContractJson{}
	for _, c := range contracts {
		jsons = append(jsons, ptnjson.ConvertContract2Json(c))
	}
	return jsons, nil
}

func (b *PtnApiBackend) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return b.ptn.dag.GetContractTpl(tplId)
}

func (b *PtnApiBackend) GetContractState(contractid []byte, key string) ([]byte, *modules.StateVersion, error) {
	return b.ptn.dag.GetContractState(contractid, key)
}

func (b *PtnApiBackend) GetContractStatesByPrefix(contractid []byte,
	prefix string) (map[string]*modules.ContractStateValue, error) {
	return b.ptn.dag.GetContractStatesByPrefix(contractid, prefix)
}
func (b *PtnApiBackend) GetContractStateJsonByPrefix(contractid []byte, prefix string) ([]ptnjson.ContractStateJson, error) {
	mapData, err := b.ptn.dag.GetContractStatesByPrefix(contractid, prefix)
	if err != nil {
		return nil, err
	}
	result := []ptnjson.ContractStateJson{}
	for k, v := range mapData {
		jdata := ptnjson.ContractStateJson{
			Key:     k,
			Value:   string(v.Value),
			Version: v.Version,
		}
		result = append(result, jdata)
	}
	return result, nil
}

func (b *PtnApiBackend) GetAddressBalanceStatistics(token string, topN int) (*statistics.TokenAddressBalanceJson,
	error) {
	utxos, err := b.ptn.dag.GetAllUtxos()
	if err != nil {
		return nil, err
	}
	asset, err := modules.StringToAsset(token)
	if err != nil {
		return nil, err
	}
	//token过滤
	pickedUtxos := []*modules.Utxo{}
	for _, utxo := range utxos {
		if utxo.Asset.IsSameAssetId(asset) {
			pickedUtxos = append(pickedUtxos, utxo)
		}
	}
	//统计各地址余额
	addrBalanceMap := make(map[common.Address]uint64)
	totalSupply := uint64(0)
	//过滤掉黑名单地址
	blacklistAddrs, _, _ := b.ptn.dag.GetBlacklistAddress()
	blacklistAddrMap := make(map[common.Address]bool)
	for _, addr := range blacklistAddrs {
		blacklistAddrMap[addr] = true
	}
	for _, utxo := range pickedUtxos {
		addr, err := tokenengine.Instance.GetAddressFromScript(utxo.PkScript)
		if err != nil {
			continue
		}
		if _, ok := blacklistAddrMap[addr]; ok {
			log.Debugf("Address[%s] is in black list don't statistic it", addr)
			continue
		}
		amount, ok := addrBalanceMap[addr]
		if ok {
			addrBalanceMap[addr] = amount + utxo.Amount
		} else {
			addrBalanceMap[addr] = utxo.Amount
		}
		totalSupply += utxo.Amount
	}

	//Map转换为[]addressBalance
	addressBalanceList := addressBalanceList{}
	for addr, balance := range addrBalanceMap {
		addressBalanceList = append(addressBalanceList, addressBalance{Address: addr, Balance: balance})
	}
	sort.Sort(addressBalanceList)
	//TopN并转换为Json对象
	result := &statistics.TokenAddressBalanceJson{}
	result.Token = asset.String()
	dec := asset.GetDecimal()
	result.TotalSupply = ptnjson.FormatAssetAmountByDecimal(totalSupply, dec)
	result.TotalAddressCount = len(addrBalanceMap)
	if topN == 0 {
		topN = len(addressBalanceList)
	} else if len(addressBalanceList) < topN {
		topN = len(addressBalanceList)
	}
	list := []statistics.AddressBalanceJson{}
	for i := 0; i < topN; i++ {
		ab := addressBalanceList[i]
		list = append(list, statistics.AddressBalanceJson{Address: ab.Address.String(),
			Balance: ptnjson.FormatAssetAmountByDecimal(ab.Balance, dec)})
	}
	result.AddressBalance = list
	return result, nil
}

type addressBalance struct {
	Address common.Address
	Balance uint64
}
type addressBalanceList []addressBalance

func (a addressBalanceList) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a addressBalanceList) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a addressBalanceList) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].Balance < a[i].Balance
}

func (b *PtnApiBackend) GetContractsWithJuryAddr(addr common.Hash) []*modules.Contract {
	return b.Dag().GetContractsWithJuryAddr(addr)
}
func (b *PtnApiBackend) GetAddressCount() int {
	return b.Dag().GetAddressCount()
}
