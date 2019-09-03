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
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"

	//"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/tokenengine"

	//"github.com/shopspring/decimal"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	NONE   = "NONE"
	ALL    = "ALL"
	SINGLE = "SINGLE"
)

type ContractInstallRsp struct {
	ReqId string `json:"reqId"`
	TplId string `json:"tplId"`
}

type ContractDeployRsp struct {
	ReqId      string `json:"reqId"`
	ContractId string `json:"ContractId"`
}

type ContractFeeRsp struct {
	TxSize         float64 `json:"tx_size(byte)"`
	TimeOut        uint32 `json:"time_out(s)"`
	ApproximateFee float64 `json:"approximate_fee(dao)"`
}

type JuryList struct {
	Addr []string `json:"account"`
}

type ContractFeeLevelRsp struct {
	ContractTxTimeoutUnitFee  uint64  `json:"contract_tx_timeout_unit_fee"`
	ContractTxSizeUnitFee     uint64  `json:"contract_tx_size_unit_fee"`
	ContractTxInstallFeeLevel float64 `json:"contract_tx_install_fee_level"`
	ContractTxDeployFeeLevel  float64 `json:"contract_tx_deploy_fee_level"`
	ContractTxInvokeFeeLevel  float64 `json:"contract_tx_invoke_fee_level"`
	ContractTxStopFeeLevel    float64 `json:"contract_tx_stop_fee_level"`
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

// 未确认的交易列表
func (s *PublicTxPoolAPI) Pending() ([]*ptnjson.TxPoolPendingJson, error) {
	queue, err := s.b.Queued()
	pending := make([]*ptnjson.TxPoolPendingJson, 0)
	for _, tx := range queue {
		item := ptnjson.ConvertTxPoolTx2PendingJson(tx)
		pending = append(pending, item)
	}
	return pending, err
}

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
		return newRPCTransaction(tx.Tx, tx.UnitHash, tx.UnitIndex, tx.Index)
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
//func (s *PublicTransactionPoolAPI) sign(addr common.Address, tx *modules.Transaction) (*modules.Transaction, error) {
//	// Look up the wallet containing the requested signer
//	account := accounts.Account{Address: addr}
//
//	wallet, err := s.b.AccountManager().Find(account)
//	if err != nil {
//		return nil, err
//	}
//	// Request the wallet to sign the transaction
//	var chainID *big.Int
//	return wallet.SignTx(account, tx, chainID)
//}

//func forking
func forking(ctx context.Context, b Backend) uint64 {
	b.SendConsensus(ctx)
	return 0
}

//func queryDb(ctx context.Context, b Backend, condition string) string {
//	b.SendConsensus(ctx)
//	return ""
//}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *modules.Transaction) (common.Hash, error) {
	if tx.IsNewContractInvokeRequest() {
		reqId, err := b.SendContractInvokeReqTx(tx)
		return reqId, err
	}

	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}
func submitTxs(ctx context.Context, b Backend, txs []*modules.Transaction) []error {
	errs := b.SendTxs(ctx, txs)
	if errs != nil {
		return errs
	}
	return nil
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
		// amount := ptnjson.Ptn2Dao(ptnAmt)
		//		// Ensure amount is in the valid range for monetary amounts.
		// if amount <= 0 /*|| amount > ptnjson.MaxDao*/ {
		// 	return "", &ptnjson.RPCError{
		// 		Code:    ptnjson.ErrRPCType,
		// 		Message: "Invalid amount",
		// 	}
		// }
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
		pkScript := tokenengine.Instance.GenerateLockScript(addr)
		// Convert the amount to satoshi.
		dao := ptnjson.Ptn2Dao(ptnAmt)
		//if err != nil {
		//	context := "Failed to convert amount"
		//	return "", internalRPCError(err.Error(), context)
		//}
		assetId := dagconfig.DagConfig.GetGasToken()
		txOut := modules.NewTxOut(dao, pkScript, assetId.ToAsset())
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
		if v.Asset.Equal(tokenAsset) {
			allUtxo[k] = v
		}
	}

	for _, tx := range poolTxs {
		for msgindex, msg := range tx.Tx.TxMessages {
			if msg.App == modules.APP_PAYMENT {
				pay := msg.Payload.(*modules.PaymentPayload)

				data, _ := json.Marshal(pay)
				if len(pay.Outputs) == 0 {
					//一个交易是可能Output为0的，所有Input都交了手续费
					log.Debugf("Payment output length=0,pay:%s", string(data))
					continue
				}
				if pay.Outputs[0].Asset == nil {
					log.Errorf("Payment output asset=nil,pay:%s", string(data))
				}

				for outIndex, output := range pay.Outputs {
					if !output.Asset.Equal(tokenAsset) {
						continue
					}
					op := modules.OutPoint{}
					op.TxHash = tx.Tx.Hash()
					op.MessageIndex = uint32(msgindex)
					op.OutIndex = uint32(outIndex)
					addr, err = tokenengine.Instance.GetAddressFromScript(output.PkScript)
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

/*func (s *PublicTransactionPoolAPI) CmdCreateTransaction(ctx context.Context, from string, to string, amount, fee decimal.Decimal) (string, error) {

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
	daolimit, _ := decimal.NewFromString("0.0001")
	if !fee.GreaterThanOrEqual(daolimit) {
		return "", fmt.Errorf("fee cannot less than 100000 Dao ")
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
	// fmt.Println(result)
	return result, nil
}*/
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
/*func signTokenTx(tx *modules.Transaction, cmdInputs []ptnjson.RawTxInput, flags string,
	pubKeyFn tokenengine.AddressGetPubKey, hashFn tokenengine.AddressGetSign) error {
	var hashType uint32
	switch flags {
	case ALL:
		hashType = tokenengine.SigHashAll
	case NONE:
		hashType = tokenengine.SigHashNone
	case SINGLE:
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
}*/

/*
func (s *PrivateTransactionPoolAPI) unlockKS(addr common.Address, password string, duration *uint64) error {
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
}*/

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
//	err = signTokenTx(tx, rawInputs, ALL, getPubKeyFn, getSignFn)
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
func SignRawTransaction(cmd *ptnjson.SignRawTransactionCmd, pubKeyFn tokenengine.AddressGetPubKey, hashFn tokenengine.AddressGetSign, addr common.Address) (ptnjson.SignRawTransactionResult, error) {

	serializedTx, err := decodeHexStr(cmd.RawTx)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	if err := rlp.DecodeBytes(serializedTx, tx); err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	//log.Debugf("InputOne txid:{%+v}", tx.TxMessages[0].Payload.(*modules.PaymentPayload).Input[0])

	var hashType uint32
	switch strings.ToUpper(*cmd.Flags) {
	case ALL:
		hashType = tokenengine.SigHashAll
	case NONE:
		hashType = tokenengine.SigHashNone
	case SINGLE:
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
		//if err != nil {
		//	return ptnjson.SignRawTransactionResult{}, DeserializationError{err}
		//}
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
	signErrs, err = tokenengine.Instance.SignTxAllPaymentInput(tx, hashType, inputpoints, redeem, pubKeyFn, hashFn)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, DeserializationError{err}
	}
	for msgidx, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}
		for inputindex := range payload.Inputs {
			err = tokenengine.Instance.ScriptValidate(PkScript, nil, tx, msgidx, inputindex)
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
				pk_addr, err := tokenengine.Instance.GetAddressFromScript(payload.Outputs[k].PkScript)
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
	// All returned errors (not OOM, which panics) encountered during
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
		return para[2:]
	}
	return para
}
func MakeAddress(ks *keystore.KeyStore, account string) (accounts.Account, error) {
	// If the specified account is a valid address, return it
	addr, err := common.StringToAddress(account)
	if err == nil {
		return accounts.Account{Address: addr}, nil
	} else {
		return accounts.Account{}, fmt.Errorf("invalid account address: %v", account)
	}

}

/*func (s *PublicTransactionPoolAPI) helpSignTx(tx *modules.Transaction,
	password string) ([]common.SignatureError, error) {
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		ks := s.b.GetKeyStore()
		account, _ := MakeAddress(ks, addr.String())
		ks.Unlock(account, password)
		return ks.GetPublicKey(addr)
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		account, _ := MakeAddress(ks, addr.String())
		return ks.SignMessageWithPassphrase(account, password, msg)
	}
	utxos := s.getTxUtxoLockScript(tx)
	return tokenengine.SignTxAllPaymentInput(tx, tokenengine.SigHashAll, utxos, nil, getPubKeyFn, getSignFn)

}*/

/*func (s *PublicTransactionPoolAPI) getTxUtxoLockScript(tx *modules.Transaction) map[modules.OutPoint][]byte {
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
}*/

//转为压力测试准备数据用
func (s *PublicTransactionPoolAPI) BatchSign(ctx context.Context, txid string, fromAddress, toAddress string, amount int, count int, password string) ([]string, error) {
	txHash := common.HexToHash(txid)
	toAddr, _ := common.StringToAddress(toAddress)
	fromAddr, _ := common.StringToAddress(fromAddress)
	utxoScript := tokenengine.Instance.GenerateLockScript(fromAddr)
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
		lockScript := tokenengine.Instance.GenerateLockScript(toAddr)
		pay.AddTxOut(modules.NewTxOut(uint64(amount), lockScript, asset))
		tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay))
		utxoLookup := map[modules.OutPoint][]byte{}
		utxoLookup[*outPoint] = utxoScript
		errs, err := tokenengine.Instance.SignTxAllPaymentInput(tx, tokenengine.SigHashAll, utxoLookup, nil, func(addresses common.Address) ([]byte, error) {
			return pubKey, nil
		},
			func(addresses common.Address, msg []byte) ([]byte, error) {
				return ks.SignMessage(addresses, msg)
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
/*func (s *PublicTransactionPoolAPI) SignRawTransaction(ctx context.Context, params string, hashtype string, password string, duration *uint64) (ptnjson.SignRawTransactionResult, error) {

	//transaction inputs
	if params == "" {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is empty")
	}
	upper_type := strings.ToUpper(hashtype)
	if upper_type != ALL && upper_type != NONE && upper_type != SINGLE {
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
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		//account, _ := MakeAddress(ks, addr.String())
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		return ks.SignMessage(addr, msg)
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
		}
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
}*/

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
/*func (s *PublicTransactionPoolAPI) SendRawTransaction(ctx context.Context, encodedTx string) (common.Hash, error) {
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
}*/

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
	signature, err := wallet.SignMessage(account, data)
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

/*
// PendingTransactions returns the transactions that are in the transaction pool and have a from address that is one of
// the accounts this node manages.
func (s *PublicTransactionPoolAPI) PendingTransactions() ([]*RPCTransaction, error) {
	pending, err := s.b.GetPoolTransactions()

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

func (api *PrivateDebugAPI) GetAllTokenBalance() map[string]uint64 {
	result := make(map[string]uint64)
	utxos, _ := api.b.GetAllUtxos()
	for _, utxo := range utxos {
		amt, ok := result[utxo.Asset]
		if ok {
			result[utxo.Asset] = amt + utxo.Amount
		} else {
			result[utxo.Asset] = utxo.Amount
		}
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
