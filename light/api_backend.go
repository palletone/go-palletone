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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/internal/ptnapi"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/ptnjson/statistics"
	"github.com/shopspring/decimal"
)

type LesApiBackend struct {
	ptn *LightPalletone
	//gpo *gasprice.Oracle
}

func (b *LesApiBackend) SignAndSendRequest(addr common.Address, tx *modules.Transaction) error {
	return nil
}

func (b *LesApiBackend) CurrentBlock() *modules.Unit {
	return &modules.Unit{}
}
func (b *LesApiBackend) QueryProofOfExistenceByReference(ref string) ([]*ptnjson.ProofOfExistenceJson, error) {
	return nil, nil
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

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB,
	*modules.Header, error) {
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

func (b *LesApiBackend) GetChainParameters() *core.ChainParameters {
	return nil
}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *modules.Transaction) error {
	return b.ptn.txPool.AddLocal(signedTx)
}
func (b *LesApiBackend) SendTxs(ctx context.Context, signedTxs []*modules.Transaction) []error {
	return b.ptn.txPool.AddLocals(signedTxs)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	//b.ptn.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (modules.Transactions, error) {
	//return b.ptn.txPool.GetTransactions()
	return nil, nil
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *modules.Transaction {
	//return b.ptn.txPool.GetTransaction(txHash)
	return nil
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	//return b.ptn.txPool.GetNonce(ctx, addr)
	return uint64(0), nil
}

func (b *LesApiBackend) Stats() (pending int, queued int, reserve int) {
	//return b.ptn.txPool.Stats(), 0, 0
	return 0, 0, 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Hash]*modules.TxPoolTransaction,
	map[common.Hash]*modules.TxPoolTransaction) {
	return nil, nil
	//return b.ptn.txPool.Content()
}
func (b *LesApiBackend) Queued() ([]*modules.TxPoolTransaction, error) {
	return nil, nil
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

func (b *LesApiBackend) GetUnstableUnits() []*ptnjson.UnitSummaryJson {
	units := b.ptn.dag.GetUnstableUnits()
	result := make([]*ptnjson.UnitSummaryJson, len(units))
	for i, unit := range units {
		result[i] = ptnjson.ConvertUnit2SummaryJson(unit)
	}
	return result
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

//test
func (b *LesApiBackend) SendConsensus(ctx context.Context) error {
	return nil
}

func (b *LesApiBackend) SaveCommon(key, val []byte) error {
	return nil
}

// dag's get common
func (b *LesApiBackend) GetCommon(key []byte) ([]byte, error) {
	return b.ptn.dag.GetCommon(key)
}
func (b *LesApiBackend) GetCommonByPrefix(prefix []byte) map[string][]byte {
	return b.ptn.dag.GetCommonByPrefix(prefix)
}

// Get Contract Api
func (b *LesApiBackend) GetContract(contractAddr common.Address) (*ptnjson.ContractJson, error) {
	return nil, nil
}

//get level db
func (b *LesApiBackend) GetUnitByHash(hash common.Hash) *modules.Unit {
	return nil
}
func (b *LesApiBackend) GetUnitByNumber(number *modules.ChainIndex) *modules.Unit {
	return nil
}
func (b *LesApiBackend) GetUnitsByIndex(start, end decimal.Decimal, asset string) []*modules.Unit {
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
		if err == nil {
			units = append(units, unit)
		}
	}
	return units
}
func (b *LesApiBackend) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	return b.ptn.dag.GetHeaderByHash(hash)
}
func (b *LesApiBackend) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	return b.ptn.dag.GetHeaderByNumber(number)
}
func (b *LesApiBackend) GetTxByReqId(hash common.Hash) (*ptnjson.TxWithUnitInfoJson, error) {
	return nil, nil
}

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
func (b *LesApiBackend) GetStxoEntry(outpoint *modules.OutPoint) (*ptnjson.StxoJson, error) {
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
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}

	utxos, _ := b.ptn.dag.GetAddrUtxos(address)
	result := make([]*ptnjson.UtxoJson, 0)
	for o, u := range utxos {
		o := o
		ujson := ptnjson.ConvertUtxo2Json(&o, u)
		result = append(result, ujson)
	}
	return result, nil

}
func (b *LesApiBackend) GetAddrUtxos2(addr string) ([]*ptnjson.UtxoJson, error) {
	return nil, nil
}

func (b *LesApiBackend) GetAddrRawUtxos(addr string) (map[modules.OutPoint]*modules.Utxo, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	return b.ptn.dag.GetAddrUtxos(address)
}
func (b *LesApiBackend) GetAllUtxos() ([]*ptnjson.UtxoJson, error) {
	return nil, nil
	//utxos, err := b.ptn.dag.GetAllUtxos()
	//if err != nil {
	//	return nil, err
	//}
	//result := []*ptnjson.UtxoJson{}
	//for o, u := range utxos {
	//	ujson := ptnjson.ConvertUtxo2Json(&o, u)
	//	result = append(result, ujson)
	//}
	//return result, nil
}
func (b *LesApiBackend) GetAddrTokenFlow(addr, token string) ([]*ptnjson.TokenFlowJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAddrTxHistory(addr string) ([]*ptnjson.TxHistoryJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAssetTxHistory(asset *modules.Asset) ([]*ptnjson.TxHistoryJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAssetExistence(asset string) ([]*ptnjson.ProofOfExistenceJson, error) {
	return nil, nil
}

//contract control
func (b *LesApiBackend) ContractInstall(ccName string, ccPath string, ccVersion string, ccDescription, ccAbi,
	ccLanguage string) (TemplateId []byte, err error) {
	return nil, nil
}
func (b *LesApiBackend) ContractDeploy(templateId []byte, txid string, args [][]byte,
	timeout time.Duration) (deployId []byte, err error) {
	return nil, nil
}

//ContractInvoke(txBytes []byte) (rspPayload []byte, err error)
func (b *LesApiBackend) ContractInvoke(deployId []byte, txid string, args [][]byte,
	timeout time.Duration) (rspPayload []byte, err error) {
	return nil, nil
}
func (b *LesApiBackend) ContractStop(deployId []byte, txid string, deleteImage bool) error {
	return nil
}
func (b *LesApiBackend) DecodeTx(hex string) (string, error) {
	return "", nil
}
func (b *LesApiBackend) DecodeJsonTx(hex string) (string, error) {
	return "", nil
}
func (b *LesApiBackend) EncodeTx(jsonStr string) (string, error) {
	return "", nil
}

func (b *LesApiBackend) ContractInstallReqTx(from, to common.Address, daoAmount, daoFee uint64,
	tplName, path, version string, description, abi, language string, addrs []common.Address) (reqId common.Hash,
	tplId []byte, err error) {
	return
}
func (b *LesApiBackend) ContractDeployReqTx(from, to common.Address, daoAmount, daoFee uint64,
	templateId []byte, args [][]byte, extData []byte, timeout time.Duration) (reqId common.Hash,
	depId common.Address, err error) {
	return
}
func (b *LesApiBackend) ContractInvokeReqTx(from, to common.Address, daoAmount, daoFee uint64,
	certID *big.Int, contractAddress common.Address, args [][]byte, timeout uint32) (reqId common.Hash, err error) {
	return
}
func (b *LesApiBackend) SendContractInvokeReqTx(requestTx *modules.Transaction) (common.Hash, error) {
	return common.Hash{}, nil
}

func (b *LesApiBackend) ContractInvokeReqTokenTx(from, to, toToken common.Address, daoAmount,
	daoFee, daoAmountToken uint64, asset string, contractAddress common.Address, args [][]byte,
	timeout uint32) (reqId common.Hash, err error) {
	return
}

func (b *LesApiBackend) ContractStopReqTx(from, to common.Address, daoAmount, daoFee uint64,
	contractId common.Address, deleteImage bool) (reqId common.Hash, err error) {
	return
}

func (b *LesApiBackend) ContractInstallReqTxFee(from, to common.Address, daoAmount, daoFee uint64, tplName,
	path, version string, description, abi, language string, addrs []common.Address) (fee float64, size float64, tm uint32,
	err error) {
	return
}
func (b *LesApiBackend) ContractDeployReqTxFee(from, to common.Address, daoAmount, daoFee uint64, templateId []byte,
	args [][]byte, extData []byte, timeout time.Duration) (fee float64, size float64, tm uint32, err error) {
	return
}
func (b *LesApiBackend) ContractInvokeReqTxFee(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	contractAddress common.Address, args [][]byte, timeout uint32) (fee float64, size float64, tm uint32, err error) {
	return
}
func (b *LesApiBackend) ContractStopReqTxFee(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address,
	deleteImage bool) (fee float64, size float64, tm uint32, err error) {
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

func (b *LesApiBackend) ContractQuery(contractId []byte, txid string, args [][]byte,
	timeout time.Duration) (rspPayload []byte, err error) {
	return nil, nil
}

func (b *LesApiBackend) Dag() dag.IDag {
	//return b.Dag()
	return nil
}

func (b *LesApiBackend) TxPool() txspool.ITxPool {
	return b.ptn.txPool
}

func (b *LesApiBackend) SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error {
	//return b.ptn.SignAndSendTransaction(addr, tx)
	return nil
}

//SignAndSendTransaction(addr common.Address, tx *modules.Transaction) error
func (b *LesApiBackend) TransferPtn(from, to string, amount decimal.Decimal,
	text *string) (*ptnapi.TxExecuteResult, error) {
	//return b.ptn.TransferPtn(from, to, amount, text)
	return nil, nil
}
func (b *LesApiBackend) GetKeyStore() *keystore.KeyStore {
	return b.ptn.GetKeyStore()
}

// get tx hash by req id
func (b *LesApiBackend) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return common.Hash{}, nil
}

func (b *LesApiBackend) GetFileInfo(filehash string) ([]*modules.FileInfo, error) {
	return nil, nil
}

//SPV
func (b *LesApiBackend) GetProofTxInfoByHash(txhash string) ([][]byte, error) {
	return nil, nil
}

func (b *LesApiBackend) ProofTransactionByHash(tx string) (string, error) {
	return b.ptn.ProtocolManager().ReqProofByTxHash(tx), nil
}

func (b *LesApiBackend) ProofTransactionByRlptx(rlptx [][]byte) (string, error) {
	return b.ptn.ProtocolManager().ReqProofByRlptx(rlptx), nil
}

func (b *LesApiBackend) SyncUTXOByAddr(addr string) string {
	return b.ptn.ProtocolManager().SyncUTXOByAddr(addr)
}

func (b *LesApiBackend) StartCorsSync() (string, error) {
	return "light node have not cors server", errors.New("light node have not cors server")
}

func (b *LesApiBackend) GetAllContractTpl() ([]*ptnjson.ContractTemplateJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetAllContracts() ([]*ptnjson.ContractJson, error) {
	return nil, nil
}
func (b *LesApiBackend) GetContractsByTpl(tplId []byte) ([]*ptnjson.ContractJson, error) {
	return nil, nil
}

func (b *LesApiBackend) GetContractState(contractid []byte, key string) ([]byte, *modules.StateVersion, error) {
	return nil, nil, nil
}
func (b *LesApiBackend) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue,
	error) {
	return nil, nil
}
func (b *LesApiBackend) GetAddressBalanceStatistics(token string, topN int) (*statistics.TokenAddressBalanceJson,
	error) {
	return nil, nil
}

func (b *LesApiBackend) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return nil, nil
}
