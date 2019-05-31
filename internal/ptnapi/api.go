// Copyright 2018 PalletOne

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

package ptnapi

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	defaultGasPrice = 0.0001 * configure.PalletOne
)

const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the
	// RPC server is allowed to stay open without authenticating before it
	// is closed.
	rpcAuthTimeoutSeconds = 10
	// uint256Size is the number of bytes needed to represent an unsigned
	// 256-bit integer.
	uint256Size = 32
	// gbtNonceRange is two 32-bit big-endian hexadecimal integers which
	// represent the valid ranges of nonces returned by the getblocktemplate
	// RPC.
	gbtNonceRange = "00000000ffffffff"
	// gbtRegenerateSeconds is the number of seconds that must pass before
	// a new template is generated when the previous block hash has not
	// changed and there have been changes to the available transactions
	// in the memory pool.
	gbtRegenerateSeconds = 60
	// maxProtocolVersion is the max protocol version the server supports.
	maxProtocolVersion = 70002
)

type ContractInstallRsp struct {
	ReqId string `json:"reqId"`
	TplId string `json:"tplId"`
}

type ContractDeployRsp struct {
	ReqId      string `json:"reqId"`
	ContractId string `json:"ContractId"`
}

type JuryList struct {
	Addr []string `json:"account"`
}

// PublicPalletOneAPI provides an API to access PalletOne related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicPalletOneAPI struct {
	b Backend
}

// NewPublicPalletOneAPI creates a new PalletOne protocol API.
func NewPublicPalletOneAPI(b Backend) *PublicPalletOneAPI {
	return &PublicPalletOneAPI{b}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicPalletOneAPI) GasPrice(ctx context.Context) (*big.Int, error) {
	return s.b.SuggestPrice(ctx)
}

// ProtocolVersion returns the current PalletOne protocol version this node supports
func (s *PublicPalletOneAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(s.b.ProtocolVersion())
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicPalletOneAPI) Syncing() (interface{}, error) {
	progress := s.b.Downloader().Progress()

	// Return not syncing if the synchronisation already completed
	//	if progress.CurrentBlock >= progress.HighestBlock {
	//		return false, nil
	//	}
	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(progress.StartingBlock),
		//"currentBlock":  hexutil.Uint64(progress.CurrentBlock),
		"highestBlock": hexutil.Uint64(progress.HighestBlock),
		"pulledStates": hexutil.Uint64(progress.PulledStates),
		"knownStates":  hexutil.Uint64(progress.KnownStates),
	}, nil
}

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicTxPoolAPI struct {
	b Backend
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI(b Backend) *PublicTxPoolAPI {
	return &PublicTxPoolAPI{b}
}

// Content returns the transactions contained within the transaction pool.
func (s *PublicTxPoolAPI) Content() map[string]map[string]*RPCTransaction {
	content := map[string]map[string]*RPCTransaction{
		"pending": make(map[string]*RPCTransaction),
		"queued":  make(map[string]*RPCTransaction),
	}
	pending, queue := s.b.TxPoolContent()
	dump1 := make(map[string]*RPCTransaction)
	// Flatten the pending transactions
	for hash, tx := range pending {
		dump1[hash.String()] = newRPCPendingTransaction(tx)
	}
	content["pending"] = dump1
	// Flatten the queued transactions
	dump2 := make(map[string]*RPCTransaction)
	for hash, tx := range queue {
		dump2[hash.String()] = newRPCPendingTransaction(tx)
	}
	content["queued"] = dump2
	return content
}

// Status returns the number of pending and queued transaction in the pool.
func (s *PublicTxPoolAPI) Status() map[string]hexutil.Uint {
	pending, queue, orphans := s.b.Stats()
	return map[string]hexutil.Uint{
		"pending": hexutil.Uint(pending),
		"queued":  hexutil.Uint(queue),
		"orphans": hexutil.Uint(orphans),
	}
}
func (s *PublicTxPoolAPI) Queue() map[common.Hash]*modules.Transaction {
	_, queue := s.b.TxPoolContent()
	result := make(map[common.Hash]*modules.Transaction)
	for hash, tx := range queue {
		result[hash] = tx.Tx
	}
	return result
}

func (s *PublicTxPoolAPI) Pending() ([]*ptnjson.TxPoolPendingJson, error) {
	queue, err := s.b.Queued()
	pending := make([]*ptnjson.TxPoolPendingJson, 0)
	for _, tx := range queue {
		item := ptnjson.ConvertTxPoolTx2PendingJson(tx)
		pending = append(pending, item)
	}
	return pending, err
}

/*
// Inspect retrieves the content of the transaction pool and flattens it into an
// easily inspectable list.
func (s *PublicTxPoolAPI) Inspect() map[string]map[string]map[string]string {
	content := map[string]map[string]map[string]string{
		"pending": make(map[string]map[string]string),
		"queued":  make(map[string]map[string]string),
	}
	pending, queue := s.b.TxPoolContent()

	// Define a formatter to flatten a transaction into a string
	var format = func(tx *modules.Transaction) string {
		if to := tx.To(); to != nil {
			return fmt.Sprintf("%s: %v wei + %v gas × %v wei", tx.To().Hex(), tx.Value(), tx.Gas(), tx.GasPrice())
		}
		return fmt.Sprintf("contract creation: %v wei + %v gas × %v wei", tx.Value(), tx.Gas(), tx.GasPrice())
	}
	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}
*/
// PublicAccountAPI provides an API to access accounts managed by this node.
// It offers only methods that can retrieve accounts.
type PublicAccountAPI struct {
	am *accounts.Manager
}

// NewPublicAccountAPI creates a new PublicAccountAPI.
func NewPublicAccountAPI(am *accounts.Manager) *PublicAccountAPI {
	return &PublicAccountAPI{am: am}
}

// Accounts returns the collection of accounts this node manages
func (s *PublicAccountAPI) Accounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// PublicBlockChainAPI provides an API to access the PalletOne blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new PalletOne blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{b}
}

//// BlockNumber returns the block number of the chain head.
//func (s *PublicBlockChainAPI) BlockNumber() *big.Int {
//	header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
//	return header.Number
//}

// GetBalance returns the amount of wei for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicBlockChainAPI) GetBalance(ctx context.Context, address string) (map[string]decimal.Decimal, error) {
	utxos, err := s.b.GetAddrUtxos(address)
	if err != nil {
		return nil, err
	}
	result := make(map[string]decimal.Decimal)
	for _, utxo := range utxos {
		asset, _ := modules.StringToAsset(utxo.Asset)
		if bal, ok := result[utxo.Asset]; ok {
			result[utxo.Asset] = bal.Add(ptnjson.AssetAmt2JsonAmt(asset, utxo.Amount))
		} else {
			result[utxo.Asset] = ptnjson.AssetAmt2JsonAmt(asset, utxo.Amount)
		}
	}
	return result, nil
}

func (s *PublicBlockChainAPI) GetTokenTxHistory(ctx context.Context, assetStr string) ([]*ptnjson.TxHistoryJson, error) {
	asset := &modules.Asset{}
	err := asset.SetString(assetStr)
	if err != nil {
		return nil, errors.New("Invalid asset string")
	}
	result, err := s.b.GetAssetTxHistory(asset)

	return result, err
}
func (s *PublicBlockChainAPI) ListSysConfig(ctx context.Context) ([]*ptnjson.ConfigJson, error) {

	result, err := s.b.GetAllSysConfig()
	return result, err
}

//func (s *PublicBlockChainAPI) WalletTokens(ctx context.Context, address string) (string, error) {
//	result, err := s.b.WalletTokens(address)
//	if err != nil {
//		log.Error("WalletTokens:", "error", err.Error())
//	}
//	//fmt.Println("result len=", len(result))
//	b, err := json.Marshal(result)
//
//	if err != nil {
//		log.Error("WalletTokens 2222:", "error", err.Error())
//	}
//	return string(b), nil
//}
//
//func (s *PublicBlockChainAPI) WalletBalance(ctx context.Context, address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
//	return s.b.WalletBalance(address, assetid, uniqueid, chainid)
//}

/*
// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		response, err := s.rpcOutputBlock(block, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

// GetBlockByHash returns the requested block. When fullTx is true all transactions in the block are returned in full
// detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return s.rpcOutputBlock(block, true, fullTx)
	}
	return nil, err
}

// GetUncleByBlockNumberAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", blockNr, "hash", block.Hash(), "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}

// GetUncleByBlockHashAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", block.Number(), "hash", blockHash, "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}
*/

// GetCode returns the code stored at the given address in the state for the given block number.
func (s *PublicBlockChainAPI) GetCode(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	/*
		state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
		if state == nil || err != nil {
			return nil, err
		}
		code := state.GetCode(address)
		return code, state.Error()
	*/
	return hexutil.Bytes{}, nil
}

// GetStorageAt returns the storage from the state at the given address, key and
// block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta block
// numbers are also allowed.
func (s *PublicBlockChainAPI) GetStorageAt(ctx context.Context, address common.Address, key string, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	/*
		state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
		if state == nil || err != nil {
			return nil, err
		}
		res := state.GetState(address, common.HexToHash(key))
		return res[:], state.Error()
	*/
	return hexutil.Bytes{}, nil
}

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      hexutil.Uint64  `json:"gas"`
	GasPrice hexutil.Big     `json:"gasPrice"`
	Value    hexutil.Big     `json:"value"`
	Data     hexutil.Bytes   `json:"data"`
}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (s *PublicBlockChainAPI) Call(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	return hexutil.Bytes{}, nil
}

// EstimateGas returns an estimate of the amount of gas needed to execute the
// given transaction against the current pending block.
func (s *PublicBlockChainAPI) EstimateGas(ctx context.Context, args CallArgs) (hexutil.Uint64, error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	//var (
	//lo  uint64 = configure.TxGas - 1
	//	hi  uint64
	//	cap uint64
	//)
	//if uint64(args.Gas) >= configure.TxGas {
	//	hi = uint64(args.Gas)
	//} else {
	// Retrieve the current pending block to act as the gas ceiling
	//		block, err := s.b.BlockByNumber(ctx, rpc.PendingBlockNumber)
	//		if err != nil {
	//			return 0, err
	//		}
	//		hi = block.GasLimit()
	//}
	//cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	//executable := func(gas uint64) bool {
	//	args.Gas = hexutil.Uint64(gas)
	//
	//	return true
	//}
	// Execute the binary search and hone in on an executable gas limit
	//for lo+1 < hi {
	//	mid := (hi + lo) / 2
	//	if !executable(mid) {
	//		lo = mid
	//	} else {
	//		hi = mid
	//	}
	//}
	// Reject the transaction as invalid if it still fails at the highest allowance
	//if hi == cap {
	//	if !executable(hi) {
	//		return 0, fmt.Errorf("gas required exceeds allowance or always failing transaction")
	//	}
	//}
	//return hexutil.Uint64(hi), nil
	return 0, nil
}

// Start forking command.
func (s *PublicBlockChainAPI) Forking(ctx context.Context, rate uint64) uint64 {
	return forking(ctx, s.b)
}

//Query leveldb

func (s *PublicBlockChainAPI) GetPrefix(condition string) string /*map[string][]byte*/ {
	log.Info("PublicBlockChainAPI", "GetPrefix condition:", condition)
	pre := s.b.GetPrefix(condition)
	prefix := map[string]string{}
	for key, value := range pre {
		prefix[key] = *(*string)(unsafe.Pointer(&value))
	}
	content, err := json.Marshal(prefix)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetPrefix Marshal err:", err, "prefix:", prefix)
		return "Marshal err"
	}
	return *(*string)(unsafe.Pointer(&content))
}

func (s *PublicBlockChainAPI) CcstartChaincodeContainer(ctx context.Context, deployId string, txid string) (string, error) {
	depId, _ := hex.DecodeString(deployId)
	log.Info("CcstartChaincodeContainer:" + deployId + ":" + txid + "_")
	//TODO deleteImage 为 true 时，目前是会删除基础镜像的
	deplo1, err := s.b.ContractStartChaincodeContainer(depId, txid)
	return string(deplo1), err
}

func (s *PublicBlockChainAPI) DecodeTx(ctx context.Context, hex string) (string, error) {
	return s.b.DecodeTx(hex)
}
func (s *PublicBlockChainAPI) EncodeTx(ctx context.Context, json string) (string, error) {
	return s.b.EncodeTx(json)
}

func (s *PublicBlockChainAPI) Election(ctx context.Context, sid string) (string, error) {
	log.Info("-----Election:", "id", sid)

	id, _ := strconv.Atoi(sid)
	rsp, err := s.b.ElectionVrf(uint32(id))
	log.Info("-----Election:" + hex.EncodeToString(rsp))
	return hex.EncodeToString(rsp), err
}

func (s *PublicBlockChainAPI) SetJuryAccount(ctx context.Context, addr, pwd string) string {
	jAddr, _ := common.StringToAddress(addr)
	log.Info("-----UpdateJuryAccount:", "addr", jAddr, "pwd", pwd)

	if s.b.UpdateJuryAccount(jAddr, pwd) {
		return "OK"
	} else {
		return "Fail"
	}
}

func (s *PublicBlockChainAPI) GetJuryAccount(ctx context.Context) *JuryList {
	log.Info("-----GetJuryAccount")
	jAccounts := s.b.GetJuryAccount()
	jlist := &JuryList{
		Addr: make([]string, len(jAccounts)),
	}
	for i := 0; i < len(jAccounts); i++ {
		jlist.Addr[i] = jAccounts[i].String()
		log.Info("-----GetJuryAccount", "addr", jlist.Addr[i])
	}

	return jlist
}

//SPV
func (s *PublicBlockChainAPI) GetProofTxInfoByHash(ctx context.Context, txhash string) ([][]byte, error) {
	return s.b.GetProofTxInfoByHash(txhash)
}

func (s *PublicBlockChainAPI) ProofTransactionByHash(ctx context.Context, txhash string) (string, error) {
	return s.b.ProofTransactionByHash(txhash)
}

func (s *PublicBlockChainAPI) ProofTransactionByRlptx(ctx context.Context, rlptx [][]byte) (string, error) {
	return s.b.ProofTransactionByRlptx(rlptx)
}

func (s *PublicBlockChainAPI) SyncUTXOByAddr(ctx context.Context, addr string) string {
	return s.b.SyncUTXOByAddr(addr)
}

func (s *PublicBlockChainAPI) StartCorsSync(ctx context.Context) (string, error) {
	return s.b.StartCorsSync()
}

// ExecutionResult groups all structured logs emitted by the EVM
// while replaying a transaction in debug mode as well as transaction
// execution status, the amount of gas used and the return value
type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode
type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   error              `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

/*
// formatLogs formats EVM returned structured logs for json output
func FormatLogs(logs []vm.StructLog) []StructLogRes {
	formatted := make([]StructLogRes, len(logs))
	for index, trace := range logs {
		formatted[index] = StructLogRes{
			Pc:      trace.Pc,
			Op:      trace.Op.String(),
			Gas:     trace.Gas,
			GasCost: trace.GasCost,
			Depth:   trace.Depth,
			Error:   trace.Err,
		}
		if trace.Stack != nil {
			stack := make([]string, len(trace.Stack))
			for i, stackValue := range trace.Stack {
				stack[i] = fmt.Sprintf("%x", math.PaddedBigBytes(stackValue, 32))
			}
			formatted[index].Stack = &stack
		}
		if trace.Memory != nil {
			memory := make([]string, 0, (len(trace.Memory)+31)/32)
			for i := 0; i+32 <= len(trace.Memory); i += 32 {
				memory = append(memory, fmt.Sprintf("%x", trace.Memory[i:i+32]))
			}
			formatted[index].Memory = &memory
		}
		if trace.Storage != nil {
			storage := make(map[string]string)
			for i, storageValue := range trace.Storage {
				storage[fmt.Sprintf("%x", i)] = fmt.Sprintf("%x", storageValue)
			}
			formatted[index].Storage = &storage
		}
	}
	return formatted
}

// rpcOutputBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func (s *PublicBlockChainAPI) rpcOutputBlock(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Header() // copies the header once
	fields := map[string]interface{}{
		"number":           (*hexutil.Big)(head.Number),
		"hash":             b.Hash(),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"sha3Uncles":       head.UncleHash,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            head.Coinbase,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"totalDifficulty":  (*hexutil.Big)(s.b.GetTd(b.Hash())),
		"extraData":        hexutil.Bytes(head.Extra),
		"size":             hexutil.Uint64(b.Size()),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        (*hexutil.Big)(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
	}

	if inclTx {
		formatTx := func(tx *modules.Transaction) (interface{}, error) {
			return tx.Hash(), nil
		}

		if fullTx {
			formatTx = func(tx *modules.Transaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}

		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range b.Transactions() {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}

	uncles := b.Uncles()
	uncleHashes := make([]common.Hash, len(uncles))
	for i, uncle := range uncles {
		uncleHashes[i] = uncle.Hash()
	}
	fields["uncles"] = uncleHashes

	return fields, nil
}
*/
// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	//UnitHash common.Hash `json:"unit_Hash"`
	//From      common.Address `json:"from"`
	UnitIndex uint64      `json:"unit_index"`
	Hash      common.Hash `json:"hash"`

	TransactionIndex hexutil.Uint `json:"transaction_index"`
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *modules.Transaction, blockHash common.Hash, unitIndex, index uint64) *RPCTransaction {
	result := &RPCTransaction{
		Hash: tx.Hash(),
	}
	if blockHash != (common.Hash{}) {
		//result.UnitHash = blockHash
		result.UnitIndex = unitIndex
		result.TransactionIndex = hexutil.Uint(index)
	}
	result.UnitIndex = unitIndex
	return result
}

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx *modules.TxPoolTransaction) *RPCTransaction {
	if tx.UnitHash != (common.Hash{}) {
		return newRPCTransaction(tx.Tx, tx.UnitHash, tx.UnitIndex, uint64(tx.Index))
	}
	return newRPCTransaction(tx.Tx, common.Hash{}, ^uint64(0), ^uint64(0))
}

/*
// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *types.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index)
}

// newRPCRawTransactionFromBlockIndex returns the bytes of a transaction given a block and a transaction index.
func newRPCRawTransactionFromBlockIndex(b *types.Block, index uint64) hexutil.Bytes {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	blob, _ := rlp.EncodeToBytes(txs[index])
	return blob
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCTransaction {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}
*/
// PublicTransactionPoolAPI exposes methods for the RPC interface
type PublicTransactionPoolAPI struct {
	b         Backend
	nonceLock *AddrLocker
}
type PublicReturnInfo struct {
	Item string      `json:"item"`
	Info interface{} `json:"info"`
	Hex  string      `json:"hex"`
}

func NewPublicReturnInfo(name string, info interface{}) *PublicReturnInfo {
	return &PublicReturnInfo{name, info, ""}
}
func NewPublicReturnInfoWithHex(name string, info interface{}, rlpData []byte) *PublicReturnInfo {
	return &PublicReturnInfo{name, info, hex.EncodeToString(rlpData)}
}

// NewPublicTransactionPoolAPI creates a new RPC service with methods specific for the transaction pool.
func NewPublicTransactionPoolAPI(b Backend, nonceLock *AddrLocker) *PublicTransactionPoolAPI {
	return &PublicTransactionPoolAPI{b, nonceLock}
}

func (s *PublicTransactionPoolAPI) GetAddrOutpoints(ctx context.Context, addr string) (string, error) {
	items, err := s.b.GetAddrOutpoints(addr)
	if err != nil {
		return "", err
	}
	info := NewPublicReturnInfo("address_outpoints", items)
	result_json, _ := json.Marshal(info)
	return string(result_json), nil
}
func (s *PublicTransactionPoolAPI) GetAddrUtxos(ctx context.Context, addr string) (string, error) {
	items, err := s.b.GetAddrUtxos(addr)

	if err != nil {
		return "", err
	}
	info := NewPublicReturnInfo("address_utxos", items)
	result_json, _ := json.Marshal(info)
	return string(result_json), nil
}
func (s *PublicTransactionPoolAPI) GetAllUtxos(ctx context.Context) (string, error) {
	items, err := s.b.GetAllUtxos()
	if err != nil {
		log.Error("Get all utxo failed.", "error", err, "result", items)
		return "", err
	}

	info := NewPublicReturnInfo("all_utxos", items)

	result_json, err := json.Marshal(info)
	if err != nil {
		log.Error("Get all utxo ,json marshal failed.", "error", err)
	}

	return string(result_json), nil
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block with the given block number.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	//	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
	//		n := hexutil.Uint(len(block.Transactions()))
	//		return &n
	//	}
	return nil
}

/*
// GetBlockTransactionCountByHash returns the number of transactions in the block with the given hash.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}


// GetTransactionByBlockNumberAndIndex returns the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockNumberAndIndex returns the bytes of the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockHashAndIndex returns the bytes of the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}
*/
// GetTransactionCount returns the number of transactions the given address has sent for the given block number
func (s *PublicTransactionPoolAPI) GetTransactionCount(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (*hexutil.Uint64, error) {
	/*
		state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
		if state == nil || err != nil {
			return nil, err
		}
		nonce := state.GetNonce(address)
		return (*hexutil.Uint64)(&nonce), state.Error()
	*/
	v := hexutil.Uint64(0)
	return &v, nil
}

func (s *PublicTransactionPoolAPI) GetTransactionsByTxid(ctx context.Context, txid string) (*ptnjson.GetTxIdResult, error) {
	tx, err := s.b.GetTxByTxid_back(txid)
	if err != nil {
		log.Error("Get transcation by hash ", "unit_hash", txid, "error", err.Error())
		return nil, err
	}
	return tx, nil
}

func (s *PublicTransactionPoolAPI) GetTxHashByReqId(ctx context.Context, hashHex string) (string, error) {
	hash := common.HexToHash(hashHex)
	item, err := s.b.GetTxHashByReqId(hash)

	info := NewPublicReturnInfo("tx_hash", item)
	result_json, _ := json.Marshal(info)
	return string(result_json), err
}

// GetTxPoolTxByHash returns the pool transaction for the given hash
func (s *PublicTransactionPoolAPI) GetTxPoolTxByHash(ctx context.Context, hex string) (string, error) {
	log.Debug("this is hash tx's hash hex to find tx.", "hex", hex)
	hash := common.HexToHash(hex)
	log.Debug("this is hash tx's hash  to find tx.", "hash", hash.String())
	item, err := s.b.GetTxPoolTxByHash(hash)
	if err != nil {
		return "pool_tx:null", err
	} else {
		info := NewPublicReturnInfo("txpool_tx", item)
		result_json, _ := json.Marshal(info)
		return string(result_json), nil
	}
}

/* old version
// GetRawTransactionByHash returns the bytes of the transaction for the given hash.
func (s *PublicTransactionPoolAPI) GetRawTransactionByHash(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	var tx *modules.Transaction

	// Retrieve a finalized transaction, or a pooled otherwise
	if tx, _, _, _ = coredata.GetTransaction(s.b.ChainDb(), hash); tx == nil {
		if tx = s.b.GetPoolTransaction(hash); tx == nil {
			// Transaction not found anywhere, abort
			return nil, nil
		}
	}
	// Serialize to RLP and return
	return rlp.EncodeToBytes(tx)
}

/*
// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (s *PublicTransactionPoolAPI) GetTransactionReceipt(ctx context.Context, hash common.Hash) (map[string]interface{}, error) {
	tx, blockHash, blockNumber, index := coredata.GetTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}
	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}
	//receipt := receipts[index]

	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		//signer = types.NewEIP155Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)

	fields := map[string]interface{}{
		"blockHash":        blockHash,
		"blockNumber":      hexutil.Uint64(blockNumber),
		"transactionHash":  hash,
		"transactionIndex": hexutil.Uint64(index),
		"from":             from,
		"to":               tx.To(),
		//"gasUsed":           hexutil.Uint64(receipt.GasUsed),
		//"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
		"contractAddress": nil,
		//"logs":              receipt.Logs,
		//"logsBloom":         receipt.Bloom,
	}

	// Assign receipt status or post state.
	//if len(receipt.PostState) > 0 {
	//fields["root"] = hexutil.Bytes(receipt.PostState)
	//} else {
	//	fields["status"] = hexutil.Uint(receipt.Status)
	//}
	//if receipt.Logs == nil {
	//	fields["logs"] = [][]*types.Log{}
	//}
	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	//if receipt.ContractAddress != (common.Address{}) {
	//	fields["contractAddress"] = receipt.ContractAddress
	//}
	return fields, nil
}
*/
// sign is a helper function that signs a transaction with the private key of the given address.
func (s *PublicTransactionPoolAPI) sign(addr common.Address, tx *modules.Transaction) (*modules.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Request the wallet to sign the transaction
	var chainID *big.Int
	return wallet.SignTx(account, tx, chainID)
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`
}

// setDefaults is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(ctx context.Context, b Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 90000
	}
	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	}
	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		//		nonce, err := b.GetPoolNonce(ctx, args.From)
		//		if err != nil {
		//			return err
		//		}
		//		args.Nonce = (*hexutil.Uint64)(&nonce)
	}
	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return errors.New(`Both "data" and "input" are set and not equal. Please use "input" to pass transaction call data.`)
	}
	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}
		if len(input) == 0 {
			return errors.New(`contract creation without any data provided`)
		}
	}
	return nil
}

func (args *SendTxArgs) toTransaction() *modules.Transaction {
	var input []byte
	if args.Data != nil {
		input = *args.Data
	} else if args.Input != nil {
		input = *args.Input
	}
	input = input
	if args.To == nil {
		//return modules.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	}
	//return modules.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	return &modules.Transaction{}
}

//func forking
func forking(ctx context.Context, b Backend) uint64 {
	b.SendConsensus(ctx)
	return 0
}

func queryDb(ctx context.Context, b Backend, condition string) string {
	b.SendConsensus(ctx)
	return ""
}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *modules.Transaction) (common.Hash, error) {
	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	/*
		if tx.To() == nil {
			signer := types.MakeSigner(b.ChainConfig(), b.CurrentBlock().Number())
			from, err := types.Sender(signer, tx)
			if err != nil {
				return common.Hash{}, err
			}
			addr := crypto.CreateAddress(from, tx.Nonce())
			log.Info("Submitted contract creation", "fullhash", tx.Hash().Hex(), "contract", addr.Hex())
		} else {
			log.Info("Submitted transaction", "fullhash", tx.Hash().Hex(), "recipient", tx.To())
		}*/
	return tx.Hash(), nil
}

const (
	MaxTxInSequenceNum uint32 = 0xffffffff
)

//create raw transction
func CreateRawTransaction( /*s *rpcServer*/ c *ptnjson.CreateRawTransactionCmd) (string, error) {

	// Validate the locktime, if given.
	if c.LockTime != nil &&
		(*c.LockTime < 0 || *c.LockTime > int64(MaxTxInSequenceNum)) {
		return "", &ptnjson.RPCError{
			Code:    ptnjson.ErrRPCInvalidParameter,
			Message: "Locktime out of range",
		}
	}
	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	//先构造PaymentPayload结构，再组装成Transaction结构
	pload := new(modules.PaymentPayload)
	for _, input := range c.Inputs {
		txHash := common.HexToHash(input.Txid)

		prevOut := modules.NewOutPoint(txHash, input.MessageIndex, input.Vout)
		txInput := modules.NewTxIn(prevOut, []byte{})
		pload.AddTxIn(txInput)
	}
	// Add all transaction outputs to the transaction after performing
	//	// some validity checks.
	//	//only support mainnet
	//	var params *chaincfg.Params
	for _, addramt := range c.Amounts {
		encodedAddr := addramt.Address
		ptnAmt := addramt.Amount
		amount := ptnjson.Ptn2Dao(ptnAmt)
		//		// Ensure amount is in the valid range for monetary amounts.
		if amount <= 0 /*|| amount > ptnjson.MaxDao*/ {
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCType,
				Message: "Invalid amount",
			}
		}
		addr, err := common.StringToAddress(encodedAddr)
		if err != nil {
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		switch addr.GetType() {
		case common.PublicKeyHash:
		case common.ScriptHash:
		case common.ContractHash:
			//case *ptnjson.AddressPubKeyHash:
			//case *ptnjson.AddressScriptHash:
		default:
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		// Create a new script which pays to the provided address.
		pkScript := tokenengine.GenerateLockScript(addr)
		// Convert the amount to satoshi.
		dao := ptnjson.Ptn2Dao(ptnAmt)
		if err != nil {
			context := "Failed to convert amount"
			return "", internalRPCError(err.Error(), context)
		}
		assetId := dagconfig.DagConfig.GetGasToken()
		txOut := modules.NewTxOut(uint64(dao), pkScript, assetId.ToAsset())
		pload.AddTxOut(txOut)
	}
	//	// Set the Locktime, if given.
	if c.LockTime != nil {
		pload.LockTime = uint32(*c.LockTime)
	}
	//	// Return the serialized and hex-encoded transaction.  Note that this
	//	// is intentionally not directly returning because the first return
	//	// value is a string and it would result in returning an empty string to
	//	// the client instead of nothing (nil) in the case of an error.
	mtx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	mtx.TxMessages = append(mtx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pload))
	//mtx.TxHash = mtx.Hash()
	mtxbt, err := rlp.EncodeToBytes(mtx)
	if err != nil {
		return "", err
	}
	//log.Debugf("payload input outpoint:%s", pload.Input[0].PreviousOutPoint.TxHash.String())
	mtxHex := hex.EncodeToString(mtxbt)
	return mtxHex, nil
}

type GetUtxoEntry func(outpoint *modules.OutPoint) (*ptnjson.UtxoJson, error)

func SelectUtxoFromDagAndPool(dbUtxo map[modules.OutPoint]*modules.Utxo, poolTxs []*modules.TxPoolTransaction,
	from string, asset string) (map[modules.OutPoint]*modules.Utxo, error) {
	tokenAsset, err := modules.StringToAsset(asset)
	if err != nil {
		return nil, err
	}
	var addr common.Address
	// store tx input utxo outpoint
	inputsOutpoint := []modules.OutPoint{}
	allUtxo := make(map[modules.OutPoint]*modules.Utxo)
	for k, v := range dbUtxo {
		if v.Asset.IsSimilar(tokenAsset) {
			allUtxo[k] = v
		}
	}

	for _, tx := range poolTxs {
		for msgindex, msg := range tx.Tx.TxMessages {
			if msg.App == modules.APP_PAYMENT {
				pay := msg.Payload.(*modules.PaymentPayload)
				if pay.Outputs[0].Asset.IsSimilar(tokenAsset) == false {
					continue
				}
				for outIndex, output := range pay.Outputs {
					op := modules.OutPoint{}
					op.TxHash = tx.Tx.Hash()
					op.MessageIndex = uint32(msgindex)
					op.OutIndex = uint32(outIndex)
					addr, err = tokenengine.GetAddressFromScript(output.PkScript)
					if err != nil {
						return nil, err
					}
					if addr.String() == from {
						allUtxo[op] = modules.NewUtxo(output, pay.LockTime, time.Now().Unix())
					}

				}
				for _, input := range pay.Inputs {

					if input.PreviousOutPoint != nil {
						inputsOutpoint = append(inputsOutpoint, *input.PreviousOutPoint)
					}
				}
			}
		}
	}
	for _, used := range inputsOutpoint {
		delete(allUtxo, used)
	}
	//vaildutxos := core.Utxos{}
	//for k, v := range allUtxo {
	//	json := modules.NewUtxoWithOutPoint( v,k)
	//	vaildutxos = append(vaildutxos, json)
	//}
	return allUtxo, nil
}
func (s *PublicTransactionPoolAPI) CmdCreateTransaction(ctx context.Context, from string, to string, amount, fee decimal.Decimal) (string, error) {

	//realNet := &chaincfg.MainNetParams
	var LockTime int64
	LockTime = 0

	amounts := []ptnjson.AddressAmt{}
	if from == "" {
		return "", fmt.Errorf("sender address is empty")
	}
	if to == "" {
		return "", fmt.Errorf("receiver address is empty")
	}
	_, ferr := common.StringToAddress(from)
	if ferr != nil {
		return "", fmt.Errorf("sender address is invalid")
	}
	_, terr := common.StringToAddress(to)
	if terr != nil {
		return "", fmt.Errorf("receiver address is invalid")
	}

	amounts = append(amounts, ptnjson.AddressAmt{to, amount})
	if len(amounts) == 0 || !amount.IsPositive() {
		return "", fmt.Errorf("amounts is invalid")
	}

	utxos, err := s.b.GetAddrRawUtxos(from)
	if err != nil {
		return "", err
	}
	poolTxs, err := s.b.GetPoolTxsByAddr(from)
	if err != nil {
		return "", err
	}
	ptn := dagconfig.DagConfig.GasToken
	resultUtxo, err := SelectUtxoFromDagAndPool(utxos, poolTxs, from, ptn)
	if err != nil {
		return "", fmt.Errorf("Select utxo err")
	}

	if !fee.IsPositive() {
		return "", fmt.Errorf("fee is invalid")
	}
	daoAmount := ptnjson.Ptn2Dao(amount.Add(fee))
	allutxos, _ := convertUtxoMap2Utxos(resultUtxo)
	taken_utxo, change, err := core.Select_utxo_Greedy(allutxos, daoAmount)
	if err != nil {
		return "", fmt.Errorf("Select utxo err")
	}

	var inputs []ptnjson.TransactionInput
	var input ptnjson.TransactionInput
	for _, u := range taken_utxo {
		utxo := u.(*modules.UtxoWithOutPoint)
		input.Txid = utxo.TxHash.String()
		input.MessageIndex = utxo.MessageIndex
		input.Vout = utxo.OutIndex
		inputs = append(inputs, input)
	}

	if change > 0 {
		amounts = append(amounts, ptnjson.AddressAmt{from, ptnjson.Dao2Ptn(change)})
	}

	arg := ptnjson.NewCreateRawTransactionCmd(inputs, amounts, &LockTime)
	result, _ := CreateRawTransaction(arg)
	fmt.Println(result)
	return result, nil
}
func convertUtxoMap2Utxos(maps map[modules.OutPoint]*modules.Utxo) (core.Utxos, *modules.Asset) {
	utxos := core.Utxos{}
	asset := &modules.Asset{}
	for k, v := range maps {
		utxos = append(utxos, modules.NewUtxoWithOutPoint(v, k))
		asset = v.Asset
	}
	return utxos, asset
}

//sign raw tx
func signTokenTx(tx *modules.Transaction, cmdInputs []ptnjson.RawTxInput, flags string,
	pubKeyFn tokenengine.AddressGetPubKey, hashFn tokenengine.AddressGetSign) error {
	var hashType uint32
	switch flags {
	case "ALL":
		hashType = tokenengine.SigHashAll
	case "NONE":
		hashType = tokenengine.SigHashNone
	case "SINGLE":
		hashType = tokenengine.SigHashSingle
	case "ALL|ANYONECANPAY":
		hashType = tokenengine.SigHashAll | tokenengine.SigHashAnyOneCanPay
	case "NONE|ANYONECANPAY":
		hashType = tokenengine.SigHashNone | tokenengine.SigHashAnyOneCanPay
	case "SINGLE|ANYONECANPAY":
		hashType = tokenengine.SigHashSingle | tokenengine.SigHashAnyOneCanPay
	default:
		return errors.New("Invalid sighash parameter")
	}

	inputPoints := make(map[modules.OutPoint][]byte)
	var redeem []byte
	for _, rti := range cmdInputs {
		inputHash := common.HexToHash(rti.Txid)

		script, err := decodeHexStr(rti.ScriptPubKey)
		if err != nil {
			return err
		}
		// redeemScript for multisig tx
		if rti.RedeemScript != "" {
			redeemScript, err := decodeHexStr(rti.RedeemScript)
			if err != nil {
				return errors.New("Invalid redeemScript")
			}
			redeem = redeemScript
		}
		inputPoints[modules.OutPoint{
			TxHash:       inputHash,
			OutIndex:     rti.Vout,
			MessageIndex: rti.MessageIndex,
		}] = script
	}

	//
	var signErrors []common.SignatureError
	signErrors, err := tokenengine.SignTxAllPaymentInput(tx, hashType, inputPoints, redeem, pubKeyFn, hashFn)
	if err != nil {
		return err
	}
	fmt.Println(len(signErrors))

	return nil
}

func (s *PublicTransactionPoolAPI) unlockKS(addr common.Address, password string, duration *uint64) error {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	ks := s.b.GetKeyStore()
	err := ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		errors.New("get addr by outpoint is err")
		return err
	}
	return nil
}

//func (s *PublicTransactionPoolAPI) TransferToken(ctx context.Context, asset string, from string, to string,
//	amount decimal.Decimal, fee decimal.Decimal, Extra string, password string, duration *uint64) (common.Hash, error) {
//	//
//	tokenAsset, err := modules.StringToAsset(asset)
//	if err != nil {
//		fmt.Println(err.Error())
//		return common.Hash{}, err
//	}
//	if !fee.IsPositive() {
//		return common.Hash{}, fmt.Errorf("fee is ZERO ")
//	}
//	//
//	fromAddr, err := common.StringToAddress(from)
//	if err != nil {
//		fmt.Println(err.Error())
//		return common.Hash{}, err
//	}
//	toAddr, err := common.StringToAddress(to)
//	if err != nil {
//		fmt.Println(err.Error())
//		return common.Hash{}, err
//	}
//	//all utxos
//	dbUtxos, err := s.b.GetAddrRawUtxos(from)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	poolTxs, err := s.b.GetPoolTxsByAddr(from)
//	if err != nil {
//		return common.Hash{}, err
//	}
//
//	utxosToken, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, asset)
//	if err != nil {
//		return common.Hash{}, fmt.Errorf("Select utxo err")
//	}
//	utxosGasToken, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, dagconfig.DagConfig.GasToken)
//	if err != nil {
//		return common.Hash{}, fmt.Errorf("Select utxo err")
//	}
//	//
//	////ptn utxos and token utxos
//	//utxosPTN := core.Utxos{}
//	//utxosToken := core.Utxos{}
//	//ptn := dagconfig.DagConfig.GasToken
//	//dagOutpoint_token := []modules.OutPoint{}
//	//dagOutpoint_ptn := []modules.OutPoint{}
//	//for _, json := range utxoJsons {
//	//	//utxos = append(utxos, &json)
//	//	if json.Asset == asset {
//	//		utxosToken = append(utxosToken, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
//	//		dagOutpoint_token = append(dagOutpoint_token, modules.OutPoint{TxHash: common.HexToHash(json.TxHash), MessageIndex: json.MessageIndex, OutIndex: json.OutIndex})
//	//	}
//	//	if json.Asset == ptn {
//	//		utxosPTN = append(utxosPTN, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
//	//		dagOutpoint_ptn = append(dagOutpoint_ptn, modules.OutPoint{TxHash: common.HexToHash(json.TxHash), MessageIndex: json.MessageIndex, OutIndex: json.OutIndex})
//	//	}
//	//}
//	//
//	//
//	//
//	//}
//	//else{
//	//ptn utxos and token utxos
//	/*for _, json := range utxoJsons {
//		if json.Asset == ptn {
//			utxosPTN = append(utxosPTN, &ptnjson.UtxoJson{TxHash: json.TxHash,
//				MessageIndex:   json.MessageIndex,
//				OutIndex:       json.OutIndex,
//				Amount:         json.Amount,
//				Asset:          json.Asset,
//				PkScriptHex:    json.PkScriptHex,
//				PkScriptString: json.PkScriptString,
//				LockTime:       json.LockTime})
//		} else {
//			if json.Asset == asset {
//				utxosToken = append(utxosToken, &ptnjson.UtxoJson{TxHash: json.TxHash,
//					MessageIndex:   json.MessageIndex,
//					OutIndex:       json.OutIndex,
//					Amount:         json.Amount,
//					Asset:          json.Asset,
//					PkScriptHex:    json.PkScriptHex,
//					PkScriptString: json.PkScriptString,
//					LockTime:       json.LockTime})
//			}
//		}
//	}*/
//	// }
//	//1.
//	tokenAmount := ptnjson.JsonAmt2AssetAmt(tokenAsset, amount)
//	feeAmount := ptnjson.Ptn2Dao(fee)
//	tx, usedUtxos, err := createTokenTx(fromAddr, toAddr, tokenAmount, feeAmount, utxosGasToken, utxosToken, tokenAsset)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	if Extra != "" {
//		textPayload := new(modules.DataPayload)
//		textPayload.MainData = []byte(asset)
//		textPayload.ExtraData = []byte(Extra)
//		tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_DATA, textPayload))
//	}
//
//	//lockscript
//	getPubKeyFn := func(addr common.Address) ([]byte, error) {
//		//TODO use keystore
//		ks := s.b.GetKeyStore()
//		return ks.GetPublicKey(addr)
//	}
//	//sign tx
//	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
//		ks := s.b.GetKeyStore()
//		return ks.SignHash(addr, hash)
//	}
//	//raw inputs
//	var rawInputs []ptnjson.RawTxInput
//	PkScript := tokenengine.GenerateLockScript(fromAddr)
//	PkScriptHex := trimx(hexutil.Encode(PkScript))
//	for _, msg := range tx.TxMessages {
//		payload, ok := msg.Payload.(*modules.PaymentPayload)
//		if ok == false {
//			continue
//		}
//		for _, txin := range payload.Inputs {
//			/*inpoint := modules.OutPoint{
//				TxHash:       txin.PreviousOutPoint.TxHash,
//				OutIndex:     txin.PreviousOutPoint.OutIndex,
//				MessageIndex: txin.PreviousOutPoint.MessageIndex,
//			}
//			uvu, eerr := s.b.GetUtxoEntry(&inpoint)
//			if eerr != nil {
//				return common.Hash{}, err
//			}*/
//			TxHash := trimx(txin.PreviousOutPoint.TxHash.String())
//			OutIndex := txin.PreviousOutPoint.OutIndex
//			MessageIndex := txin.PreviousOutPoint.MessageIndex
//			input := ptnjson.RawTxInput{TxHash, OutIndex, MessageIndex, PkScriptHex, ""}
//			rawInputs = append(rawInputs, input)
//			/*addr, err := tokenengine.GetAddressFromScript(hexutil.MustDecode(PkScriptHex))
//			if err != nil {
//				return common.Hash{}, err
//				//fmt.Println("get addr by outpoint is err")
//			}*/
//			/*TxHash := trimx(uvu.TxHash)
//			PkScriptHex := trimx(uvu.PkScriptHex)
//			input := ptnjson.RawTxInput{TxHash, uvu.OutIndex, uvu.MessageIndex, PkScriptHex, ""}
//			rawInputs = append(rawInputs, input)*/
//		}
//	}
//	//2.
//	err = s.unlockKS(fromAddr, password, duration)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	//3.
//	err = signTokenTx(tx, rawInputs, "ALL", getPubKeyFn, getSignFn)
//	if err != nil {
//		return common.Hash{}, err
//	}
//	//4.
//	return submitTransaction(ctx, s.b, tx)
//}

//create raw transction
/*
func (s *PublicTransactionPoolAPI) CreateRawTransaction(ctx context.Context, params string) (string, error) {
	var rawTransactionGenParams ptnjson.RawTransactionGenParams
	err := json.Unmarshal([]byte(params), &rawTransactionGenParams)
	if err != nil {
		return "", err
	}
	//transaction inputs
	var inputs []ptnjson.TransactionInput
	for _, inputOne := range rawTransactionGenParams.Inputs {
		input := ptnjson.TransactionInput{inputOne.Txid, inputOne.Vout, inputOne.MessageIndex}
		inputs = append(inputs, input)
	}
	if len(inputs) == 0 {
		return "", nil
	}
	//realNet := &chaincfg.MainNetParams
	amounts := []ptnjson.AddressAmt{}
	for _, outOne := range rawTransactionGenParams.Outputs {
		if len(outOne.Address) == 0 || outOne.Amount.LessThanOrEqual(decimal.New(0, 0)) {
			continue
		}
		amounts = append(amounts, ptnjson.AddressAmt{outOne.Address, outOne.Amount})
	}
	if len(amounts) == 0 {
		return "", nil
	}

	arg := ptnjson.NewCreateRawTransactionCmd(inputs, amounts, &rawTransactionGenParams.Locktime)
	result, _ := CreateRawTransaction(arg)
	fmt.Println(result)
	return result, nil
}*/

//sign rawtranscation
func SignRawTransaction(icmd interface{}, pubKeyFn tokenengine.AddressGetPubKey, hashFn tokenengine.AddressGetSign, addr common.Address) (ptnjson.SignRawTransactionResult, error) {
	cmd := icmd.(*ptnjson.SignRawTransactionCmd)
	serializedTx, err := decodeHexStr(cmd.RawTx)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	if err := rlp.DecodeBytes(serializedTx, &tx); err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	//log.Debugf("InputOne txid:{%+v}", tx.TxMessages[0].Payload.(*modules.PaymentPayload).Input[0])

	var hashType uint32
	switch *cmd.Flags {
	case "ALL":
		hashType = tokenengine.SigHashAll
	case "NONE":
		hashType = tokenengine.SigHashNone
	case "SINGLE":
		hashType = tokenengine.SigHashSingle
	case "ALL|ANYONECANPAY":
		hashType = tokenengine.SigHashAll | tokenengine.SigHashAnyOneCanPay
	case "NONE|ANYONECANPAY":
		hashType = tokenengine.SigHashNone | tokenengine.SigHashAnyOneCanPay
	case "SINGLE|ANYONECANPAY":
		hashType = tokenengine.SigHashSingle | tokenengine.SigHashAnyOneCanPay
	default:
		//e := errors.New("Invalid sighash parameter")
		return ptnjson.SignRawTransactionResult{}, err
	}

	inputpoints := make(map[modules.OutPoint][]byte)
	//scripts := make(map[string][]byte)
	//var params *chaincfg.Params
	var cmdInputs []ptnjson.RawTxInput
	if cmd.Inputs != nil {
		cmdInputs = *cmd.Inputs
	}
	var redeem []byte
	var PkScript []byte
	for _, rti := range cmdInputs {
		inputHash := common.HexToHash(rti.Txid)
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, DeserializationError{err}
		}
		script, err := decodeHexStr(trimx(rti.ScriptPubKey))
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, err
		}
		// redeemScript is only actually used iff the user provided
		// private keys. In which case, it is used to get the scripts
		// for signing. If the user did not provide keys then we always
		// get scripts from the wallet.
		//		// Empty strings are ok for this one and hex.DecodeString will
		//		// DTRT.
		if rti.RedeemScript != "" {
			redeemScript, err := decodeHexStr(rti.RedeemScript)
			if err != nil {
				return ptnjson.SignRawTransactionResult{}, err
			}
			//lockScript := tokenengine.GenerateP2SHLockScript(crypto.Hash160(redeemScript))
			//addressMulti,err:=tokenengine.GetAddressFromScript(lockScript)
			//if err != nil {
			//	return nil, DeserializationError{err}
			//}
			//mutiAddr = addressMulti
			//scripts[addressMulti.Str()] = redeemScript
			redeem = redeemScript
		}
		inputpoints[modules.OutPoint{
			TxHash:       inputHash,
			OutIndex:     rti.Vout,
			MessageIndex: rti.MessageIndex,
		}] = script
		PkScript = script
	}

	//var keys map[common.Address]*ecdsa.PrivateKey
	//if cmd.PrivKeys != nil {
	//	keys = make(map[common.Address]*ecdsa.PrivateKey)
	//
	//	if cmd.PrivKeys != nil {
	//		for _, key := range *cmd.PrivKeys {
	//			privKey, _ := crypto.FromWIF(key)
	//			//privKeyBytes, _ := hex.DecodeString(key)
	//			//privKey, _ := crypto.ToECDSA(privKeyBytes)
	//			addr := crypto.PubkeyToAddress(&privKey.PublicKey)
	//			keys[addr] = privKey
	//		}
	//	}
	//}

	var signErrs []common.SignatureError
	signErrs, err = tokenengine.SignTxAllPaymentInput(tx, hashType, inputpoints, redeem, pubKeyFn, hashFn)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, DeserializationError{err}
	}
	for msgidx, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
			continue
		}
		for inputindex, _ := range payload.Inputs {
			err = tokenengine.ScriptValidate(PkScript, nil, tx, msgidx, inputindex)
			if err != nil {
				return ptnjson.SignRawTransactionResult{}, DeserializationError{err}
			}
		}

		for k := range payload.Outputs {
			switch hashType & 0x1f {
			case tokenengine.SigHashNone:
				payload.Outputs[k].PkScript = payload.Outputs[k].PkScript[0:0] // Empty slice.
			case tokenengine.SigHashSingle:
				// Resize output array to up to and including requested index.
				payload.Outputs = payload.Outputs[:1+1]
				pk_addr, err := tokenengine.GetAddressFromScript(payload.Outputs[k].PkScript)
				if err != nil {
					return ptnjson.SignRawTransactionResult{}, errors.New("Get addr FromScript is err when signtx")
				}
				// All but current output get zeroed out.
				if pk_addr != addr {
					payload.Outputs[k].PkScript = nil
					payload.Outputs[k].Value = 0
					payload.Outputs[k].Asset = &modules.Asset{}
				}
			}
		}
	}
	// All returned errors (not OOM, which panics) encounted during
	// bytes.Buffer writes are unexpected.
	mtxbt, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	signedHex := hex.EncodeToString(mtxbt)
	signErrors := make([]ptnjson.SignRawTransactionError, 0, len(signErrs))
	return ptnjson.SignRawTransactionResult{
		Hex:      signedHex,
		Txid:     tx.Hash().String(),
		Complete: len(signErrors) == 0,
		Errors:   signErrors,
	}, nil
}

func trimx(para string) string {
	if strings.HasPrefix(para, "0x") || strings.HasPrefix(para, "0X") {
		return fmt.Sprintf("%s", para[2:])
	}
	return para
}
func MakeAddress(ks *keystore.KeyStore, account string) (accounts.Account, error) {
	// If the specified account is a valid address, return it
	addr, err := common.StringToAddress(account)
	if err == nil {
		return accounts.Account{Address: addr}, nil
	} else {
		return accounts.Account{}, fmt.Errorf("invalid account address: %s", account)
	}

}

func (s *PublicTransactionPoolAPI) helpSignTx(tx *modules.Transaction, password string) ([]common.SignatureError, error) {
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		ks := s.b.GetKeyStore()
		account, _ := MakeAddress(ks, addr.String())
		ks.Unlock(account, password)
		return ks.GetPublicKey(addr)
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		account, _ := MakeAddress(ks, addr.String())
		return ks.SignHashWithPassphrase(account, password, hash)
	}
	utxos := s.getTxUtxoLockScript(tx)
	return tokenengine.SignTxAllPaymentInput(tx, tokenengine.SigHashAll, utxos, nil, getPubKeyFn, getSignFn)

}
func (s *PublicTransactionPoolAPI) getTxUtxoLockScript(tx *modules.Transaction) map[modules.OutPoint][]byte {
	result := map[modules.OutPoint][]byte{}

	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay := msg.Payload.(*modules.PaymentPayload)
			for _, input := range pay.Inputs {
				utxo, _ := s.b.GetUtxoEntry(input.PreviousOutPoint)
				lockScript, _ := hexutil.Decode(utxo.PkScriptHex)
				result[*input.PreviousOutPoint] = lockScript
			}
		}
	}
	return result
}

//转为压力测试准备数据用
func (s *PublicTransactionPoolAPI) BatchSign(ctx context.Context, txid string, fromAddress, toAddress string, amount int, count int, password string) ([]string, error) {
	txHash := common.HexToHash(txid)
	toAddr, _ := common.StringToAddress(toAddress)
	fromAddr, _ := common.StringToAddress(fromAddress)
	utxoScript := tokenengine.GenerateLockScript(fromAddr)
	ks := s.b.GetKeyStore()
	ks.Unlock(accounts.Account{Address: fromAddr}, password)
	pubKey, _ := ks.GetPublicKey(fromAddr)
	result := []string{}
	asset := dagconfig.DagConfig.GetGasToken().ToAsset()
	for i := 0; i < count; i++ {
		tx := &modules.Transaction{}
		pay := &modules.PaymentPayload{}
		outPoint := modules.NewOutPoint(txHash, 0, uint32(i))
		pay.AddTxIn(modules.NewTxIn(outPoint, []byte{}))
		lockScript := tokenengine.GenerateLockScript(toAddr)
		pay.AddTxOut(modules.NewTxOut(uint64(amount*100000000), lockScript, asset))
		tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay))
		utxoLookup := map[modules.OutPoint][]byte{}
		utxoLookup[*outPoint] = utxoScript
		errs, err := tokenengine.SignTxAllPaymentInput(tx, tokenengine.SigHashAll, utxoLookup, nil, func(addresses common.Address) ([]byte, error) {
			return pubKey, nil
		},
			func(addresses common.Address, hash []byte) ([]byte, error) {
				return ks.SignHash(addresses, hash)
			})
		if len(errs) > 0 || err != nil {
			return nil, err
		}
		encodeTx, _ := rlp.EncodeToBytes(tx)
		result = append(result, hex.EncodeToString(encodeTx))
	}
	return result, nil
}

//sign rawtranscation
//create raw transction
func (s *PublicTransactionPoolAPI) SignRawTransaction(ctx context.Context, params string, hashtype string, password string, duration *uint64) (ptnjson.SignRawTransactionResult, error) {

	//transaction inputs
	if params == "" {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is empty")
	}
	upper_type := strings.ToUpper(hashtype)
	if upper_type != "ALL" && upper_type != "NONE" && upper_type != "SINGLE" {
		return ptnjson.SignRawTransactionResult{}, errors.New("Hashtype is error,error type:" + hashtype)
	}
	serializedTx, err := decodeHexStr(params)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is invalid")
	}

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	if err := rlp.DecodeBytes(serializedTx, &tx); err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params decode is invalid")
	}

	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()

		return ks.GetPublicKey(addr)
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		//return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		//account, _ := MakeAddress(ks, addr.String())
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		return ks.SignHash(addr, hash)
		//return crypto.Sign(hash, privKey)
	}
	var srawinputs []ptnjson.RawTxInput

	var addr common.Address
	var keys []string
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
			continue
		}
		for _, txin := range payload.Inputs {
			inpoint := modules.OutPoint{
				TxHash:       txin.PreviousOutPoint.TxHash,
				OutIndex:     txin.PreviousOutPoint.OutIndex,
				MessageIndex: txin.PreviousOutPoint.MessageIndex,
			}
			uvu, eerr := s.b.GetUtxoEntry(&inpoint)
			if eerr != nil {
				log.Error(eerr.Error())
				return ptnjson.SignRawTransactionResult{}, err
			}
			TxHash := trimx(uvu.TxHash)
			PkScriptHex := trimx(uvu.PkScriptHex)
			input := ptnjson.RawTxInput{TxHash, uvu.OutIndex, uvu.MessageIndex, PkScriptHex, ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.GetAddressFromScript(hexutil.MustDecode(uvu.PkScriptHex))
			if err != nil {
				log.Error(err.Error())
				return ptnjson.SignRawTransactionResult{}, errors.New("get addr FromScript is err")
			}
		}
		/*for _, txout := range payload.Outputs {
			err = tokenengine.ScriptValidate(txout.PkScript, tx, 0, 0)
			if err != nil {
			}
		}*/
	}
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return ptnjson.SignRawTransactionResult{}, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	ks := s.b.GetKeyStore()
	err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		newErr := errors.New("get addr by outpoint get err:" + err.Error())
		log.Error(newErr.Error())
		return ptnjson.SignRawTransactionResult{}, newErr
	}

	newsign := ptnjson.NewSignRawTransactionCmd(params, &srawinputs, &keys, ptnjson.String(hashtype))
	result, err := SignRawTransaction(newsign, getPubKeyFn, getSignFn, addr)
	if !result.Complete {
		log.Error("Not complete!!!")
		for _, e := range result.Errors {
			log.Error("SignError:" + e.Error)
		}
	}
	return result, err
}

// SendTransaction creates a transaction for the given argument, sign it and submit it to the
// transaction pool.
func (s *PublicTransactionPoolAPI) SendTransaction(ctx context.Context, args SendTxArgs) (common.Hash, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int

	signed, err := wallet.SignTx(account, tx, chainID)
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, signed)
}

type Authentifier struct {
	Address string `json:"address"`
	R       []byte `json:"r"`
	S       []byte `json:"s"`
	V       []byte `json:"v"`
}

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicTransactionPoolAPI) SendRawTransaction(ctx context.Context, encodedTx string) (common.Hash, error) {
	//transaction inputs
	if encodedTx == "" {
		return common.Hash{}, errors.New("Params is Empty")
	}
	tx := new(modules.Transaction)
	serializedTx, err := decodeHexStr(encodedTx)
	if err != nil {
		return common.Hash{}, errors.New("encodedTx is invalid")
	}

	if err := rlp.DecodeBytes(serializedTx, tx); err != nil {
		return common.Hash{}, errors.New("encodedTx decode is invalid")
	}
	if 0 == len(tx.TxMessages) {
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
			continue
		}

		for _, txout := range payload.Outputs {
			outAmount += txout.Value
		}
	}
	return submitTransaction(ctx, s.b, tx)
}

// Sign calculates an ECDSA signature for:
// keccack256("\x19Ethereum Signed Message:\n" + len(message) + message).
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The account associated with addr must be unlocked.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
func (s *PublicTransactionPoolAPI) Sign(addr common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Sign the requested hash with the wallet
	signature, err := wallet.SignHash(account, signHash(data))
	if err == nil {
		signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}
	return signature, err
}

// SignTransactionResult represents a RLP encoded signed transaction.
type SignTransactionResult struct {
	Raw hexutil.Bytes        `json:"raw"`
	Tx  *modules.Transaction `json:"tx"`
}

// SignTransaction will sign the given transaction with the from account.
// The node needs to have the private key of the account corresponding with
// the given from address and it needs to be unlocked.
func (s *PublicTransactionPoolAPI) SignTransaction(ctx context.Context, args SendTxArgs) (*SignTransactionResult, error) {
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if args.Nonce == nil {
		return nil, fmt.Errorf("nonce not specified")
	}
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	tx, err := s.sign(args.From, args.toTransaction())
	if err != nil {
		return nil, err
	}
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{data, tx}, nil
}

/*
// PendingTransactions returns the transactions that are in the transaction pool and have a from address that is one of
// the accounts this node manages.
func (s *PublicTransactionPoolAPI) PendingTransactions() ([]*RPCTransaction, error) {
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return nil, err
	}

	transactions := make([]*RPCTransaction, 0, len(pending))
	for _, tx := range pending {
		var signer types.Signer = types.HomesteadSigner{}
		if tx.Protected() {
			//signer = types.NewEIP155Signer(tx.ChainId())
		}
		from, _ := types.Sender(signer, tx)
		if _, err := s.b.AccountManager().Find(accounts.Account{Address: from}); err == nil {
			transactions = append(transactions, newRPCPendingTransaction(tx))
		}
	}
	return transactions, nil
}


// Resend accepts an existing transaction and a new gas price and limit. It will remove
// the given transaction from the pool and reinsert it with the new gas price and limit.
func (s *PublicTransactionPoolAPI) Resend(ctx context.Context, sendArgs SendTxArgs, gasPrice *hexutil.Big, gasLimit *hexutil.Uint64) (common.Hash, error) {
	if sendArgs.Nonce == nil {
		return common.Hash{}, fmt.Errorf("missing transaction nonce in transaction spec")
	}
	if err := sendArgs.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	matchTx := sendArgs.toTransaction()
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return common.Hash{}, err
	}

	for _, p := range pending {
		var signer types.Signer = types.HomesteadSigner{}
		if p.Protected() {
			//signer = types.NewEIP155Signer(p.ChainId())
		}
		wantSigHash := signer.Hash(matchTx)

		if pFrom, err := types.Sender(signer, p); err == nil && pFrom == sendArgs.From && signer.Hash(p) == wantSigHash {
			// Match. Re-sign and send the transaction.
			if gasPrice != nil && (*big.Int)(gasPrice).Sign() != 0 {
				sendArgs.GasPrice = gasPrice
			}
			if gasLimit != nil && *gasLimit != 0 {
				sendArgs.Gas = gasLimit
			}
			signedTx, err := s.sign(sendArgs.From, sendArgs.toTransaction())
			if err != nil {
				return common.Hash{}, err
			}
			if err = s.b.SendTx(ctx, signedTx); err != nil {
				return common.Hash{}, err
			}
			return signedTx.Hash(), nil
		}
	}

	return common.Hash{}, fmt.Errorf("Transaction %#x not found", matchTx.Hash())
}
*/
// PublicDebugAPI is the collection of PalletOne APIs exposed over the public
// debugging endpoint.
type PublicDebugAPI struct {
	b Backend
}

// NewPublicDebugAPI creates a new API definition for the public debug methods
// of the PalletOne service.
func NewPublicDebugAPI(b Backend) *PublicDebugAPI {
	return &PublicDebugAPI{b: b}
}

func (api *PublicDebugAPI) GetProtocolVersion() int {
	return api.b.ProtocolVersion()
}

//// GetBlockRlp retrieves the RLP encoded for of a single block.
//func (api *PublicDebugAPI) GetBlockRlp(ctx context.Context, number uint64) (string, error) {
//	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
//	if block == nil {
//		return "", fmt.Errorf("block #%d not found", number)
//	}
//	encoded, err := rlp.EncodeToBytes(block)
//	if err != nil {
//		return "", err
//	}
//	return fmt.Sprintf("%x", encoded), nil
//}

// PrintBlock retrieves a block and returns its pretty printed form.
//func (api *PublicDebugAPI) PrintBlock(ctx context.Context, number uint64) (string, error) {
//	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
//	if block == nil {
//		return "", fmt.Errorf("block #%d not found", number)
//	}
//	return spew.Sdump(block), nil
//}

// PrivateDebugAPI is the collection of PalletOne APIs exposed over the private
// debugging endpoint.
type PrivateDebugAPI struct {
	b Backend
}

// NewPrivateDebugAPI creates a new API definition for the private debug methods
// of the PalletOne service.
func NewPrivateDebugAPI(b Backend) *PrivateDebugAPI {
	return &PrivateDebugAPI{b: b}
}

// ChaindbProperty returns leveldb properties of the chain database.
func (api *PrivateDebugAPI) ChaindbProperty(property string) (string, error) {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return "", fmt.Errorf("chaindbProperty does not work for memory databases")
	}
	if property == "" {
		property = "leveldb.stats"
	} else if !strings.HasPrefix(property, "leveldb.") {
		property = "leveldb." + property
	}
	return ldb.LDB().GetProperty(property)
}

func (api *PrivateDebugAPI) ChaindbCompact() error {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return fmt.Errorf("chaindbCompact does not work for memory databases")
	}
	for b := byte(0); b < 255; b++ {
		log.Info("Compacting chain database", "range", fmt.Sprintf("0x%0.2X-0x%0.2X", b, b+1))
		err := ldb.LDB().CompactRange(util.Range{Start: []byte{b}, Limit: []byte{b + 1}})
		if err != nil {
			log.Error("Database compaction failed", "err", err)
			return err
		}
	}
	return nil
}

// SetHead rewinds the head of the blockchain to a previous block.
func (api *PrivateDebugAPI) SetHead(number hexutil.Uint64) {
	api.b.SetHead(uint64(number))
}
func (api *PrivateDebugAPI) QueryDbByKey(keyString string, keyHex string) *ptnjson.DbRowJson {
	if keyString != "" {
		return api.b.QueryDbByKey([]byte(keyString))
	}
	if keyHex != "" {
		key, _ := hex.DecodeString(keyHex)
		return api.b.QueryDbByKey(key)
	}
	return nil
}
func (api *PrivateDebugAPI) QueryDbByPrefix(keyString string, keyHex string) []*ptnjson.DbRowJson {
	var result []*ptnjson.DbRowJson
	if keyString != "" {
		result = api.b.QueryDbByPrefix([]byte(keyString))
	}
	if keyHex != "" {
		key, _ := hex.DecodeString(keyHex)
		result = api.b.QueryDbByPrefix(key)
	}
	if len(result) > 10 && (keyString == "" || keyHex == "") {
		//Data too long, only return top 10 rows
		log.Debug("QueryDbByPrefix Return result too long, truncate it, only return 10 rows. If you want to see full data, please input both 2 args")
		result = result[0:10]
	}
	return result
}
func (api *PrivateDebugAPI) SaveCommon(keyHex string, valueHex string) error {
	if keyHex == "" || valueHex == "" {
		return fmt.Errorf("saveCommon does not supported empty strings.")
	}
	key, err0 := hexutil.Decode(keyHex)
	if err0 != nil {
		return err0
	}
	value, err1 := hexutil.Decode(valueHex)
	if err1 != nil {
		return err1
	}
	//log.Info("saveCommon info", "key", string(key), "key_b", key)
	return api.b.SaveCommon(key[:], value[:])
}

// PublicNetAPI offers network related RPC methods
type PublicNetAPI struct {
	net            *p2p.Server
	networkVersion uint64
}

// NewPublicNetAPI creates a new net API instance.
func NewPublicNetAPI(net *p2p.Server, networkVersion uint64) *PublicNetAPI {
	return &PublicNetAPI{net, networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
func (s *PublicNetAPI) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
func (s *PublicNetAPI) PeerCount() hexutil.Uint {
	return hexutil.Uint(s.net.PeerCount())
}

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}
