/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package ptnapi

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"time"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/ptnjson/statistics"
	"github.com/shopspring/decimal"
)

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
// - startingBlock: block number this node started to synchronize from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicPalletOneAPI) Syncing() (interface{}, error) {
	progress := s.b.Downloader().Progress()

	// Return not syncing if the synchronization already completed
	if progress.CurrentBlock >= progress.HighestBlock {
		return false, nil
	}
	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(progress.StartingBlock),
		"currentBlock":  hexutil.Uint64(progress.CurrentBlock),
		"highestBlock":  hexutil.Uint64(progress.HighestBlock),
		//"pulledStates": hexutil.Uint64(progress.PulledStates),
		//"knownStates":  hexutil.Uint64(progress.KnownStates),
	}, nil
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
// // latest header should always be available
//	header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber)
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

func (s *PublicBlockChainAPI) GetTokenTxHistory(ctx context.Context,
	assetStr string) ([]*ptnjson.TxHistoryJson, error) {
	asset := &modules.Asset{}
	err := asset.SetString(assetStr)
	if err != nil {
		return nil, errors.New("Invalid asset string")
	}
	result, err := s.b.GetAssetTxHistory(asset)

	return result, err
}

func (s *PublicBlockChainAPI) GetAssetExistence(ctx context.Context,
	asset string) ([]*ptnjson.ProofOfExistenceJson, error) {
	result, err := s.b.GetAssetExistence(asset)
	return result, err

}

func (s *PublicBlockChainAPI) ListSysConfig() ([]*ptnjson.ConfigJson, error) {
	cp := s.b.Dag().GetChainParameters()

	return ptnjson.ConvertAllSysConfigToJson(cp), nil
}

func (s *PublicBlockChainAPI) GetChainParameters() (*core.ChainParameters, error) {
	return s.b.Dag().GetChainParameters(), nil
}

func (s *PublicBlockChainAPI) GetPledge(addStr string) (*modules.PledgeStatusJson, error) {
	// 参数检查
	_, err := common.StringToAddress(addStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", addStr)
	}

	// 构建参数
	cArgs := [][]byte{defaultMsg0, defaultMsg1, []byte(modules.QueryPledgeStatusByAddr), []byte(addStr)}
	txid := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().Unix())).Int31n(100000000))

	// 调用系统合约
	rsp, err := s.b.ContractQuery(syscontract.DepositContractAddress.Bytes(), txid[:], cArgs, 0)
	if err != nil {
		return nil, err
	}

	pledge := &modules.PledgeStatusJson{}
	err = json.Unmarshal(rsp, pledge)
	if err == nil {
		return pledge, nil
	}

	return nil, fmt.Errorf(string(rsp))
}

func (s *PublicBlockChainAPI) AddressBalanceStatistics(ctx context.Context, token string,
	topN int) (*statistics.TokenAddressBalanceJson, error) {
	result, err := s.b.GetAddressBalanceStatistics(token, topN)

	return result, err
}

//
//func (s *PublicBlockChainAPI) WalletBalance(ctx context.Context, address string, assetid []byte, uniqueid []byte,
// chainid uint64) (uint64, error) {
//	return s.b.WalletBalance(address, assetid, uniqueid, chainid)
//}

/*
// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber,
fullTx bool) (map[string]interface{}, error) {
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
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash,
fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return s.rpcOutputBlock(block, true, fullTx)
	}
	return nil, err
}

// GetUncleByBlockNumberAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber,
index hexutil.Uint) (map[string]interface{}, error) {
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
func (s *PublicBlockChainAPI) GetUncleByBlockHashAndIndex(ctx context.Context, blockHash common.Hash,
index hexutil.Uint) (map[string]interface{}, error) {
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
func (s *PublicBlockChainAPI) GetCode(ctx context.Context, address common.Address,
	blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
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
func (s *PublicBlockChainAPI) GetStorageAt(ctx context.Context, address common.Address, key string,
	blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
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
func (s *PublicBlockChainAPI) Call(ctx context.Context, args CallArgs,
	blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
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
		value := value
		prefix[key] = *(*string)(unsafe.Pointer(&value))
	}
	content, err := json.Marshal(prefix)
	if err != nil {
		log.Info("PublicBlockChainAPI", "GetPrefix Marshal err:", err, "prefix:", prefix)
		return "Marshal err"
	}
	return *(*string)(unsafe.Pointer(&content))
}

func (s *PublicBlockChainAPI) DecodeTx(ctx context.Context, hex string) (string, error) {
	return s.b.DecodeTx(hex)
}
func (s *PublicBlockChainAPI) DecodeJsonTx(ctx context.Context, hex string) (string, error) {
	return s.b.DecodeJsonTx(hex)
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
