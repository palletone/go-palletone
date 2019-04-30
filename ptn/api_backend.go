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
	"encoding/hex"
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/bloombits"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
)

// PtnApiBackend implements ethapi.Backend for full nodes
type PtnApiBackend struct {
	ptn *PalletOne
	//gpo *gasprice.Oracle
}

func (b *PtnApiBackend) Dag() dag.IDag {
	return b.ptn.dag
}

//func (b *PtnApiBackend) SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error {
//	return b.ptn.SignAndSendTransaction(addr, tx)
//}

func (b *PtnApiBackend) GetKeyStore() *keystore.KeyStore {
	return b.ptn.GetKeyStore()
}

func (b *PtnApiBackend) TransferPtn(from, to string, amount decimal.Decimal,
	text *string) (*mp.TxExecuteResult, error) {
	return b.ptn.TransferPtn(from, to, amount, text)
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
		for _, tx := range batch {
			txs = append(txs, txspool.PooltxToTx(tx))
		}
	}
	return txs, nil
}

func (b *PtnApiBackend) GetPoolTransaction(hash common.Hash) *modules.Transaction {
	tx, _ := b.ptn.txPool.Get(hash)
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
	for _, msgcopy := range tx.TxMessages {
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

//func (b *PtnApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
//	return b.ptn.txPool.State().GetNonce(addr), nil
//}

func (b *PtnApiBackend) Stats() (int, int, int) {
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

//func (b *PtnApiBackend) WalletTokens(address string) (map[string]*modules.AccountToken, error) {
//	//comAddr, err := common.StringToAddress("P1NsG3kiKJc87M6Di6YriqHxqfPhdvxVj2B")
//	comAddr, err := common.StringToAddress(address)
//	if err != nil {
//		return nil, err
//	}
//	return b.ptn.dag.WalletTokens(comAddr)
//}
//
//func (b *PtnApiBackend) WalletBalance(address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
//	comAddr, err := common.StringToAddress(address)
//	if err != nil {
//		return 0, err
//	}
//	return b.ptn.dag.WalletBalance(comAddr, assetid, uniqueid, chainid)
//}

// GetContract
func (b *PtnApiBackend) GetContract(id string) (*modules.Contract, error) {
	return b.ptn.dag.GetContract(common.Hex2Bytes(id))
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
		txj := ptnjson.ConvertTx2HistoryJson(tx, b.ptn.dag.GetUtxoEntry)
		txjs = append(txjs, txj)
	}
	return txjs, nil
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
		txs_json = append(txs_json, ptnjson.ConvertTx2SummaryJson(tx, hash, header.Number.Index, header.Time, uint64(txIdx), b.ptn.dag.GetUtxoEntry))
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
	return ptnjson.ConvertTxWithUnitInfo2FullJson(tx, b.ptn.dag.GetUtxoEntry), nil
}
func (b *PtnApiBackend) GetTxByReqId(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error) {
	tx, err := b.ptn.dag.GetTxByReqId(hash)
	if err != nil {
		return nil, err
	}
	return ptnjson.ConvertTxWithUnitInfo2FullJson(tx, b.ptn.dag.GetUtxoEntry), nil
}
func (b *PtnApiBackend) GetTxSearchEntry(hash common.Hash) (*ptnjson.TxSerachEntryJson, error) {
	entry, err := b.ptn.dag.GetTxSearchEntry(hash)
	return ptnjson.ConvertTxEntry2Json(entry), err
}

// GetPoolTxByHash return a json of the tx in pool.
func (b *PtnApiBackend) GetTxPoolTxByHash(hash common.Hash) (*ptnjson.TxPoolTxJson, error) {
	tx, unit_hash := b.ptn.txPool.Get(hash)
	return ptnjson.ConvertTxPoolTx2Json(tx, unit_hash), nil
}

func (b *PtnApiBackend) GetPoolTxsByAddr(addr string) ([]*modules.TxPoolTransaction, error) {
	tx, err := b.ptn.txPool.GetPoolTxsByAddr(addr)
	return tx, err
}

// func (b *PtnApiBackend) GetTxsPoolTxByHash(hash common.Hash) (*ptnjson.TxPoolTxJson, error) {
// 	tx, unit_hash := b.ptn.txPool.Get(hash)
// 	return ptnjson.ConvertTxPoolTx2Json(tx, unit_hash), nil
// }

func (b *PtnApiBackend) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	return b.ptn.dag.GetHeaderByHash(hash)
}

func (b *PtnApiBackend) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	return b.ptn.dag.GetHeaderByNumber(number)
}

func (b *PtnApiBackend) GetPrefix(prefix string) map[string][]byte {
	return b.ptn.dag.GetCommonByPrefix([]byte(prefix))
} //getprefix

func (b *PtnApiBackend) GetUtxoEntry(outpoint *modules.OutPoint) (*ptnjson.UtxoJson, error) {

	//This function query from txpool first, not exist, then query from leveldb.
	utxo, err := b.ptn.txPool.GetUtxoEntry(outpoint)
	if err != nil {
		log.Errorf("Utxo not found in txpool and leveldb, key:%s", outpoint.String())
		return nil, err
	}
	ujson := ptnjson.ConvertUtxo2Json(outpoint, utxo)
	return ujson, nil
}

//func (b *PtnApiBackend) GetAddrOutput(addr string) ([]modules.Output, error) {
//	return b.ptn.dag.GetAddrOutput(addr)
//}

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

func (b *PtnApiBackend) GetAddrUtxos(addr string) ([]*ptnjson.UtxoJson, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}

	utxos, _ := b.ptn.dag.GetAddrUtxos(address)
	result := []*ptnjson.UtxoJson{}
	for o, u := range utxos {
		ujson := ptnjson.ConvertUtxo2Json(&o, u)
		result = append(result, ujson)
	}
	return result, nil
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
		ujson := ptnjson.ConvertUtxo2Json(&o, u)
		result = append(result, ujson)
	}
	return result, nil

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
	txjs := []*ptnjson.TxHistoryJson{}
	for _, tx := range txs {
		txj := ptnjson.ConvertTx2HistoryJson(tx, b.ptn.dag.GetUtxoEntry)
		txjs = append(txjs, txj)
	}
	return txjs, nil
}

//contract control
func (b *PtnApiBackend) ContractInstall(ccName string, ccPath string, ccVersion string) (TemplateId []byte, err error) {
	//tempid := []byte{0x1, 0x2, 0x3}
	log.Debugf("======>ContractInstall:name[%s]path[%s]version[%s]", ccName, ccPath, ccVersion)

	//payload, err := cc.Install("palletone", ccName, ccPath, ccVersion)
	payload, err := b.ptn.contract.Install("palletone", ccName, ccPath, ccVersion)

	return payload.TemplateId, err
}

func (b *PtnApiBackend) ContractDeploy(templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, err error) {
	//depid := []byte{0x4, 0x5, 0x6}
	log.Debugf("======>ContractDeploy:tmId[%s]txid[%s]", hex.EncodeToString(templateId), txid)

	//depid, _, err := cc.Deploy("palletone", templateId, txid, args, timeout)
	depid, _, err := b.ptn.contract.Deploy("palletone", templateId, txid, args, timeout)
	return depid, err
}

func (b *PtnApiBackend) ContractInvoke(deployId []byte, txid string, args [][]byte, timeout time.Duration) ([]byte, error) {
	log.Debugf("======>ContractInvoke:deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)

	unit, err := b.ptn.contract.Invoke("palletone", deployId, txid, args, timeout)
	//todo print rwset
	if err != nil {
		return nil, err
	}
	return unit.Payload, err
	// todo tmp
	//b.ptn.contractPorcessor.ContractTxReqBroadcast(deployId, txid, args, timeout)
	//return nil, nil
}

func (b *PtnApiBackend) ContractQuery(contractId []byte, txid string, args [][]byte, timeout time.Duration) (rspPayload []byte, err error) {
	//contractAddr := common.HexToAddress(hex.EncodeToString(contractId))
	rsp, err := b.ptn.contract.Invoke("palletone", contractId, txid, args, timeout)
	if err != nil {
		return nil, err
	}
	log.Debugf("=====>ContractQuery:contractId[%s]txid[%s]", hex.EncodeToString(contractId), txid)
	//fmt.Printf("contract query rsp = %#v\n", string(rsp.Payload))
	return rsp.Payload, nil
}

func (b *PtnApiBackend) ContractStop(deployId []byte, txid string, deleteImage bool) error {
	log.Debugf("======>ContractStop:deployId[%s]txid[%s]", hex.EncodeToString(deployId), txid)

	//err := cc.Stop("palletone", deployId, txid, deleteImage)
	_, err := b.ptn.contract.Stop("palletone", deployId, txid, deleteImage)
	return err
}

//
func (b *PtnApiBackend) ContractInstallReqTx(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string) (reqId common.Hash, tplId []byte, err error) {
	return b.ptn.contractPorcessor.ContractInstallReq(from, to, daoAmount, daoFee, tplName, path, version, true, nil)
}
func (b *PtnApiBackend) ContractDeployReqTx(from, to common.Address, daoAmount, daoFee uint64, templateId []byte, args [][]byte, timeout time.Duration) (reqId common.Hash, depId []byte, err error) {
	return b.ptn.contractPorcessor.ContractDeployReq(from, to, daoAmount, daoFee, templateId, args, timeout)
}
func (b *PtnApiBackend) ContractInvokeReqTx(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int, contractAddress common.Address, args [][]byte, timeout uint32) (reqId common.Hash, err error) {
	return b.ptn.contractPorcessor.ContractInvokeReq(from, to, daoAmount, daoFee, certID, contractAddress, args, timeout)
}
func (b *PtnApiBackend) ContractInvokeReqTokenTx(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, assetToken string, contractAddress common.Address, args [][]byte, timeout uint32) (reqId common.Hash, err error) {
	return b.ptn.contractPorcessor.ContractInvokeReqToken(from, to, toToken, daoAmount, daoFee, daoAmountToken, assetToken, contractAddress, args, timeout)
}
func (b *PtnApiBackend) ContractStopReqTx(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address, deleteImage bool) (reqId common.Hash, err error) {
	return b.ptn.contractPorcessor.ContractStopReq(from, to, daoAmount, daoFee, contractId, deleteImage)
}
func (b *PtnApiBackend) ElectionVrf(id uint32) ([]byte, error) {
	return b.ptn.contractPorcessor.ElectionVrfReq(id)
}
func (b *PtnApiBackend) UpdateJuryAccount(addr common.Address, pwd string) bool {
	return b.ptn.contractPorcessor.UpdateJuryAccount(addr, pwd)
}

func (b *PtnApiBackend) GetJuryAccount() []common.Address {
	return b.ptn.contractPorcessor.GetJuryAccount()
}

func (b *PtnApiBackend) GetCommon(key []byte) ([]byte, error) {
	return b.ptn.dag.GetCommon(key)
}

func (b *PtnApiBackend) GetCommonByPrefix(prefix []byte) map[string][]byte {
	return b.ptn.dag.GetCommonByPrefix(prefix)
}
func (b *PtnApiBackend) DecodeTx(hexStr string) (string, error) {
	tx := &modules.Transaction{}
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	err = rlp.DecodeBytes(bytes, tx)
	if err != nil {
		return "", err
	}
	txjson := ptnjson.ConvertTx2FullJson(tx, nil)
	json, err := json.Marshal(txjson)
	return string(json), err
}
func (b *PtnApiBackend) EncodeTx(jsonStr string) (string, error) {
	txjson := &ptnjson.TxJson{}
	json.Unmarshal([]byte(jsonStr), txjson)
	tx := ptnjson.ConvertJson2Tx(txjson)
	bytes, err := rlp.EncodeToBytes(tx)

	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), err
}

func (b *PtnApiBackend) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return b.ptn.dag.GetTxHashByReqId(reqid)
}

func (b *PtnApiBackend) GetFileInfo(filehash string) ([]*modules.FileInfo, error) {
	return b.ptn.dag.GetFileInfo([]byte(filehash))
}

//SPV
func (s *PtnApiBackend) ProofTransactionByHash(tx string) (string, error) {
	return "", nil
}

func (s *PtnApiBackend) ProofTransactionByRlptx(rlptx string) (string, error) {
	return "", nil
}

func (b *PtnApiBackend) ValidationPath(tx string) ([]byte, error) {
	return nil, nil
}
