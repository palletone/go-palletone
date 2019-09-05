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

// Package ethclient provides a client for the palletone RPC API.
package ptnclient

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
)

// Client defines typed wrappers for the Palletone RPC API.
type Client struct {
	c *rpc.Client
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *Client {
	return &Client{c}
}

func (ec *Client) Close() {
	ec.c.Close()
}

// Blockchain Access

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
//func (ec *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
//	return ec.getBlock(ctx, "ptn_getBlockByHash", hash, true)
//}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
//func (ec *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
//	return ec.getBlock(ctx, "ptn_getBlockByNumber", toBlockNumArg(number), true)
//}

// type rpcBlock struct {
// 	Hash         common.Hash      `json:"hash"`
// 	Transactions []rpcTransaction `json:"transactions"`
// 	UncleHashes  []common.Hash    `json:"uncles"`
// }

//func (ec *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*types.Block, error) {
//	return &types.Block{}, nil
//}

// HeaderByHash returns the block header with the given hash.
func (ec *Client) HeaderByHash(ctx context.Context, hash common.Hash) (*modules.Header, error) {
	var head *modules.Header
	err := ec.c.CallContext(ctx, &head, "ptn_getBlockByHash", hash, false)
	if err == nil && head == nil {
		err = palletone.ErrNotFound
	}
	return head, err
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (ec *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*modules.Header, error) {
	var head *modules.Header
	err := ec.c.CallContext(ctx, &head, "ptn_getBlockByNumber", toBlockNumArg(number), false)
	if err == nil && head == nil {
		err = palletone.ErrNotFound
	}
	return head, err
}

type rpcTransaction struct {
	tx *modules.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string
	BlockHash   common.Hash
	From        common.Address
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

// TransactionByHash returns the transaction with the given hash.
func (ec *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *modules.Transaction,
	isPending bool, err error) {
	var json *rpcTransaction
	err = ec.c.CallContext(ctx, &json, "ptn_getTransactionByHash", hash)
	//	if err != nil {
	//		return nil, false, err
	//	} else if json == nil {
	//		return nil, false, palletone.ErrNotFound
	//	} else if _, r, _ := json.tx.RawSignatureValues(); r == nil {
	//		return nil, false, fmt.Errorf("server returned transaction without signature")
	//	}
	//setSenderFromServer(json.tx, json.From, json.BlockHash)
	return json.tx, json.BlockNumber == nil, err
}

// TransactionSender returns the sender address of the given transaction. The transaction
// must be known to the remote node and included in the blockchain at the given block and
// index. The sender is the one derived by the protocol at the time of inclusion.
//
// There is a fast-path for transactions retrieved by TransactionByHash and
// TransactionInBlock. Getting their sender address can be done without an RPC interaction.
//func (ec *Client) TransactionSender(ctx context.Context, tx *modules.Transaction, block common.Hash, index uint) (common.Address, error) {
//	// Try to load the address from the cache.
//	sender, err := types.Sender(&senderFromServer{blockhash: block}, tx)
//	if err == nil {
//		return sender, nil
//	}
//	var meta struct {
//		Hash common.Hash
//		From common.Address
//	}
//	if err = ec.c.CallContext(ctx, &meta, "ptn_getTransactionByBlockHashAndIndex", block, hexutil.Uint64(index)); err != nil {
//		return common.Address{}, err
//	}
//	if meta.Hash == (common.Hash{}) || meta.Hash != tx.Hash() {
//		return common.Address{}, errors.New("wrong inclusion block/index")
//	}
//	return meta.From, nil
//}

// TransactionCount returns the total number of transactions in the given block.
func (ec *Client) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	var num hexutil.Uint
	err := ec.c.CallContext(ctx, &num, "ptn_getBlockTransactionCountByHash", blockHash)
	return uint(num), err
}

// TransactionInBlock returns a single transaction at index in the given block.
func (ec *Client) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*modules.Transaction, error) {
	var json *rpcTransaction
	err := ec.c.CallContext(ctx, &json, "ptn_getTransactionByBlockHashAndIndex", blockHash, hexutil.Uint64(index))
	//if err == nil {
	//	//		if json == nil {
	//	//			return nil, palletone.ErrNotFound
	//	//		} else if _, r, _ := json.tx.RawSignatureValues(); r == nil {
	//	//			return nil, fmt.Errorf("server returned transaction without signature")
	//	//		}
	//}
	//setSenderFromServer(json.tx, json.From, json.BlockHash)
	return json.tx, err
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
//func (ec *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
//	var r *types.Receipt
//	err := ec.c.CallContext(ctx, &r, "ptn_getTransactionReceipt", txHash)
//	if err == nil {
//		if r == nil {
//			return nil, palletone.ErrNotFound
//		}
//	}
//	return r, err
//}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

type rpcProgress struct {
	StartingBlock hexutil.Uint64
	CurrentBlock  hexutil.Uint64
	HighestBlock  hexutil.Uint64
	PulledStates  hexutil.Uint64
	KnownStates   hexutil.Uint64
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (ec *Client) SyncProgress(ctx context.Context) (*palletone.SyncProgress, error) {
	var raw json.RawMessage
	if err := ec.c.CallContext(ctx, &raw, "ptn_syncing"); err != nil {
		return nil, err
	}
	// Handle the possible response types
	var syncing bool
	if err := json.Unmarshal(raw, &syncing); err == nil {
		return nil, nil // Not syncing (always false)
	}
	var progress *rpcProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, err
	}
	return &palletone.SyncProgress{
		StartingBlock: uint64(progress.StartingBlock),
		CurrentBlock:  uint64(progress.CurrentBlock),
		HighestBlock:  uint64(progress.HighestBlock),
		PulledStates:  uint64(progress.PulledStates),
		KnownStates:   uint64(progress.KnownStates),
	}, nil
}

// SubscribeNewHead subscribes to notifications about the current blockchain head
// on the given channel.
func (ec *Client) SubscribeNewHead(ctx context.Context, ch chan<- *modules.Header) (palletone.Subscription, error) {
	return ec.c.EthSubscribe(ctx, ch, "newHeads")
}

// State Access

// NetworkID returns the network ID (also known as the chain ID) for this chain.
func (ec *Client) NetworkID(ctx context.Context) (*big.Int, error) {
	version := new(big.Int)
	var ver string
	if err := ec.c.CallContext(ctx, &ver, "net_version"); err != nil {
		return nil, err
	}
	if _, ok := version.SetString(ver, 10); !ok {
		return nil, fmt.Errorf("invalid net_version result %q", ver)
	}
	return version, nil
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ec *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "ptn_getBalance", account, toBlockNumArg(blockNumber))
	return (*big.Int)(&result), err
}

// StorageAt returns the value of key in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "ptn_getStorageAt", account, key, toBlockNumArg(blockNumber))
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "ptn_getCode", account, toBlockNumArg(blockNumber))
	return result, err
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (ec *Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "ptn_getTransactionCount", account, toBlockNumArg(blockNumber))
	return uint64(result), err
}

/*
// FilterLogs executes a filter query.
func (ec *Client) FilterLogs(ctx context.Context, q palletone.FilterQuery) ([]types.Log, error) {
	var result []types.Log
	err := ec.c.CallContext(ctx, &result, "ptn_getLogs", toFilterArg(q))
	return result, err
}

// SubscribeFilterLogs subscribes to the results of a streaming filter query.
func (ec *Client) SubscribeFilterLogs(ctx context.Context, q palletone.FilterQuery, ch chan<- types.Log) (palletone.Subscription, error) {
	return ec.c.EthSubscribe(ctx, ch, "logs", toFilterArg(q))
}
*/
// func toFilterArg(q palletone.FilterQuery) interface{} {
// 	arg := map[string]interface{}{
// 		"fromBlock": toBlockNumArg(q.FromBlock),
// 		"toBlock":   toBlockNumArg(q.ToBlock),
// 		"address":   q.Addresses,
// 		"topics":    q.Topics,
// 	}
// 	if q.FromBlock == nil {
// 		arg["fromBlock"] = "0x0"
// 	}
// 	return arg
// }

// Pending State

// PendingBalanceAt returns the wei balance of the given account in the pending state.
func (ec *Client) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "ptn_getBalance", account, "pending")
	return (*big.Int)(&result), err
}

// PendingStorageAt returns the value of key in the contract storage of the given account in the pending state.
func (ec *Client) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "ptn_getStorageAt", account, key, "pending")
	return result, err
}

// PendingCodeAt returns the contract code of the given account in the pending state.
func (ec *Client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "ptn_getCode", account, "pending")
	return result, err
}

// PendingNonceAt returns the account nonce of the given account in the pending state.
// This is the nonce that should be used for the next transaction.
func (ec *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "ptn_getTransactionCount", account, "pending")
	return uint64(result), err
}

// PendingTransactionCount returns the total number of transactions in the pending state.
func (ec *Client) PendingTransactionCount(ctx context.Context) (uint, error) {
	var num hexutil.Uint
	err := ec.c.CallContext(ctx, &num, "ptn_getBlockTransactionCountByNumber", "pending")
	return uint(num), err
}

// TODO: SubscribePendingTransactions (needs server side)

// Contract Calling

// CallContract executes a message call transaction, which is directly executed in the VM
// of the node, but never mined into the blockchain.
//
// blockNumber selects the block height at which the call runs. It can be nil, in which
// case the code is taken from the latest known block. Note that state from very old
// blocks might not be available.
func (ec *Client) CallContract(ctx context.Context, msg palletone.CallMsg, blockNumber *big.Int) ([]byte, error) {
	var hex hexutil.Bytes
	err := ec.c.CallContext(ctx, &hex, "ptn_call", toCallArg(msg), toBlockNumArg(blockNumber))
	if err != nil {
		return nil, err
	}
	return hex, nil
}

// PendingCallContract executes a message call transaction using the EVM.
// The state seen by the contract call is the pending state.
func (ec *Client) PendingCallContract(ctx context.Context, msg palletone.CallMsg) ([]byte, error) {
	var hex hexutil.Bytes
	err := ec.c.CallContext(ctx, &hex, "ptn_call", toCallArg(msg), "pending")
	if err != nil {
		return nil, err
	}
	return hex, nil
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (ec *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	var hex hexutil.Big
	if err := ec.c.CallContext(ctx, &hex, "ptn_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction based on
// the current pending state of the backend blockchain. There is no guarantee that this is
// the true gas limit requirement as other transactions may be added or removed by miners,
// but it should provide a basis for setting a reasonable default.
func (ec *Client) EstimateGas(ctx context.Context, msg palletone.CallMsg) (uint64, error) {
	var hex hexutil.Uint64
	err := ec.c.CallContext(ctx, &hex, "ptn_estimateGas", toCallArg(msg))
	if err != nil {
		return 0, err
	}
	return uint64(hex), nil
}

/*func (ec *Client) CmdCreateTransaction(ctx context.Context, from string, to string, amount uint64, fee uint64) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "ptn_cmdCreateTransaction", from, to, amount)
	return result, err
}*/

func (ec *Client) GetPtnTestCoin(ctx context.Context, from string, to string, amount string, password string, duration *uint64) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "wallet_getPtnTestCoin", from, to, amount, password, duration)
	return result, err
}

func (ec *Client) TransferToken(ctx context.Context, asset string, from string, to string, amount uint64, fee uint64, password string, extra string, duration *uint64) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "wallet_transferToken", asset, from, to, amount, fee, extra, password, duration)
	return result, err
}

//func (ec *Client) walletCreateTransaction(ctx context.Context, from string, to string, amount uint64, fee uint64) (string, error) {
//	var result string
//	err := ec.c.CallContext(ctx, &result, "wallet_createRawTransaction", from, to, amount, fee)
//	return result, err
//}
func (ec *Client) CreateRawTransaction(ctx context.Context, params string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "ptn_createRawTransaction", params)
	return result, err
}
func (ec *Client) SignRawTransaction(ctx context.Context, params string, password string, duration *uint64) (*ptnjson.SignRawTransactionResult, error) {
	var result *ptnjson.SignRawTransactionResult
	err := ec.c.CallContext(ctx, &result, "wallet_signRawTransaction", params, password, duration)
	return result, err
}

// SendTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendTransaction(ctx context.Context, tx *modules.Transaction) error {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}
	return ec.c.CallContext(ctx, nil, "ptn_sendRawTransaction", common.ToHex(data))
}
func (ec *Client) WalletSendTransaction(ctx context.Context, params string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "wallet_sendRawTransaction", params)
	return result, err
}
func toCallArg(msg palletone.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}

// Forking tool's client for the palletone RPC API
func (ec *Client) ForkingAt(ctx context.Context, account common.Address, rate uint64) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "ptn_forking", account, rate)
	return uint64(result), err
}

//func (ec *Client) GetUnitByHashAt(ctx context.Context, condition string) (string, error) {
//	var result string
//	log.Println("GetUnitByHashAt condition:", condition)
//	err := ec.c.CallContext(ctx, &result, "dag_getUnitByHash", condition)
//	return result, err
//}
//
////GetUnitByNumber
//func (ec *Client) GetUnitByNumberAt(ctx context.Context, condition string) (string, error) {
//	var result string
//	log.Println("GetUnitByNumberAt condition:", condition)
//	err := ec.c.CallContext(ctx, &result, "dag_getUnitByNumber", condition)
//	return result, err
//}
//
////GetPrefix
//func (ec *Client) GetPrefix(ctx context.Context, condition string) (string, error) {
//	var result string
//	log.Println("GetPrefix condition:", condition)
//	err := ec.c.CallContext(ctx, &result, "ptn_getPrefix", condition)
//	return result, err
//}

//func (ec *Client) CcinstallAt(ctx context.Context, ccname string, ccpath string, ccversion string) (uint64, error) {
//	var result hexutil.Uint64
//	log.Printf("==============================CcInstallAt:" + ccname + ":" + ccpath + ":" + ccversion)
//	err := ec.c.CallContext(ctx, &result, "contract_ccinstall", ccname, ccpath, ccversion)
//	return uint64(result), err
//}
//
//func (ec *Client) CcdeployAt(ctx context.Context, templateId string, txid string) (uint64, error) {
//	var result hexutil.Uint64
//	log.Printf("==============================CcdeployAt:" + templateId + ":" + txid + ":")
//	err := ec.c.CallContext(ctx, &result, "contract_ccdeploy", templateId, txid)
//	return uint64(result), err
//}
//
//func (ec *Client) CcinvokeAt(ctx context.Context, deployId string, txid string) (uint64, error) {
//	var result hexutil.Uint64
//	log.Printf("==============================CcinvokeAt:" + deployId + ":" + txid + ":")
//	err := ec.c.CallContext(ctx, &result, "contract_ccinvoke", deployId, txid)
//	return uint64(result), err
//}
//
//func (ec *Client) CcstopAt(ctx context.Context, deployId string, txid string) (uint64, error) {
//	var result hexutil.Uint64
//	log.Printf("==============================CcstopAt:" + deployId + ":" + txid + ":")
//	err := ec.c.CallContext(ctx, &result, "contract_ccstop", deployId, txid)
//	return uint64(result), err
//}

/**
rpc wallet 操作接口
The apis for wallet by rpc
*/
func (ec *Client) WalletTokens(ctx context.Context, addr string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "ptn_walletTokens", addr)
	return result, err
}

func (ec *Client) WalletBalance(ctx context.Context, address string, assetid []byte, uniqueid []byte, chainid uint64) (uint64, error) {
	var result uint64
	err := ec.c.CallContext(ctx, &result, "ptn_walletBalance", address, assetid, uniqueid, chainid)
	return result, err
}
func (ec *Client) GetTransactionsByTxid(ctx context.Context, txid string) (*modules.Transaction, error) {
	var result *modules.Transaction
	err := ec.c.CallContext(ctx, &result, "ptn_getTransactionsByTxid", txid)
	return result, err
}

// GetContract
func (ec *Client) GetContract(ctx context.Context, id common.Hash) (*modules.Contract, error) {
	result := new(modules.Contract)
	err := ec.c.CallContext(ctx, &result, "ptn_getContract", id)
	return result, err
}

// Get Header
func (ec *Client) GetHeader(ctx context.Context, hash common.Hash, index uint64) (*modules.Header, error) {
	result := new(modules.Header)
	err := ec.c.CallContext(ctx, &result, "ptn_getHeader", hash, index)
	return result, err
}

// Get Unit
func (ec *Client) GetUnit(ctx context.Context, hash common.Hash) (*modules.Unit, error) {
	result := new(modules.Unit)
	err := ec.c.CallContext(ctx, &result, "ptn_getUnit", hash)
	return result, err
}

// Get UnitNumber
func (ec *Client) GetUnitNumber(ctx context.Context, hash common.Hash) (uint64, error) {
	var result uint64
	err := ec.c.CallContext(ctx, &result, "dag_getUnitNumber", hash)
	return result, err
}

// GetCanonicalHash
//func (ec *Client) GetCanonicalHash(ctx context.Context, number uint64) (common.Hash, error) {
//	var result common.Hash
//	err := ec.c.CallContext(ctx, &result, "dag_getCanonicalHash", number)
//	return result, err
//}

// Get state
func (ec *Client) GetHeadHeaderHash(ctx context.Context) (common.Hash, error) {
	var result common.Hash
	err := ec.c.CallContext(ctx, &result, "dag_getHeadHeaderHash", nil)
	return result, err
}

func (ec *Client) GetHeadUnitHash(ctx context.Context) (common.Hash, error) {
	var result common.Hash
	err := ec.c.CallContext(ctx, &result, "dag_getHeadUnitHash", nil)
	return result, err
}

func (ec *Client) GetHeadFastUnitHash(ctx context.Context) (common.Hash, error) {
	var result common.Hash
	err := ec.c.CallContext(ctx, &result, "dag_getHeadFastUnitHash", nil)
	return result, err
}

func (ec *Client) GetTrieSyncProgress(ctx context.Context) (uint64, error) {
	var result uint64
	err := ec.c.CallContext(ctx, &result, "ptn_getTrieSyncProgress", nil)
	return result, err
}

func (ec *Client) GetUtxoEntry(ctx context.Context, key []byte) (*ptnjson.UtxoJson, error) {
	result := new(ptnjson.UtxoJson)
	err := ec.c.CallContext(ctx, &result, "dag_getUtxoEntry", key)
	return result, err
}

func (ec *Client) GetAddrOutput(ctx context.Context, addr string) ([]modules.Output, error) {
	result := make([]modules.Output, 0)
	err := ec.c.CallContext(ctx, &result, "dag_getAddrOutput", addr)
	return result, err
}
func (ec *Client) GetAddrOutpoints(ctx context.Context, addr string) ([]modules.OutPoint, error) {
	result := make([]modules.OutPoint, 0)
	err := ec.c.CallContext(ctx, &result, "dag_getAddrOutpoints", addr)
	return result, err
}
func (ec *Client) GetAddrUtxos(ctx context.Context, addr string) ([]ptnjson.UtxoJson, error) {
	result := make([]ptnjson.UtxoJson, 0)
	err := ec.c.CallContext(ctx, &result, "dag_getAddrUtxos", addr)
	return result, err
}
func (ec *Client) GetAllUtxos(ctx context.Context) ([]ptnjson.UtxoJson, error) {
	result := make([]ptnjson.UtxoJson, 0)
	err := ec.c.CallContext(ctx, &result, "dag_getAllUtxos", nil)
	return result, err
}

func (ec *Client) GetAddrTransactions(ctx context.Context, addr string) (map[string]modules.Transactions, error) {
	result := make(map[string]modules.Transactions)
	err := ec.c.CallContext(ctx, &result, "dag_getAddrTransactions", addr)
	return result, err
}

//
//func (ec *Client) GetAllTokenInfo(ctx context.Context) (*modules.AllTokenInfo, error) {
//	result := new(modules.AllTokenInfo)
//	err := ec.c.CallContext(ctx, &result, "dag_getAllTokenInfo", nil)
//	return result, err
//}
//
//func (ec *Client) GetTokenInfo(ctx context.Context, key string) (*modules.TokenInfo, error) {
//	result := new(modules.TokenInfo)
//
//	err := ec.c.CallContext(ctx, &result, "dag_getTokenInfo", key)
//	return result, err
//}
//
//func (ec *Client) SaveTokenInfo(ctx context.Context, name, token, creator string) (*modules.TokenInfo, error) {
//	result := new(modules.TokenInfo)
//
//	err := ec.c.CallContext(ctx, &result, "dag_saveTokenInfo", name, token, creator)
//	return result, err
//}

func (ec *Client) GetCommon(ctx context.Context, key string) ([]byte, error) {
	result := make([]byte, 0)
	err := ec.c.CallContext(ctx, &result, "dag_getCommon", key)
	return result, err
}

func (ec *Client) GetCommonByPrefix(ctx context.Context, prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := ec.c.CallContext(ctx, &result, "dag_getCommonByPrefix", prefix)
	return result, err
}
func (ec *Client) DecodeTx(ctx context.Context, hex string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "ptn_decodeTx", hex)
	return result, err
}
func (ec *Client) DecodeJsonTx(ctx context.Context, hex string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "ptn_decodeJsonTx", hex)
	return result, err
}
func (ec *Client) EncodeTx(ctx context.Context, jsonStr string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "ptn_encodeTx", jsonStr)
	return result, err
}
func (ec *Client) GetUnitTransactions(ctx context.Context, hashHex string) ([]*ptnjson.TxSummaryJson, error) {
	result := make([]*ptnjson.TxSummaryJson, 0)
	err := ec.c.CallContext(ctx, &result, "dag_getUnitTxsInfo", hashHex)
	return result, err
}

func (ec *Client) GetUnitTxsHash(ctx context.Context, hashHex string) ([]string, error) {
	result := make([]string, 0)
	err := ec.c.CallContext(ctx, &result, "dag_getUnitTxsHashHex", hashHex)
	return result, err
}

// GetTransactionByHash
func (ec *Client) GetTransactionByHash(ctx context.Context, hashHex string) (*ptnjson.TxSummaryJson, error) {
	result := new(ptnjson.TxSummaryJson)
	err := ec.c.CallContext(ctx, &result, "dag_getTxByHash", hashHex)
	return result, err
}

// GetTxSearchEntry.
func (ec *Client) GetTxSearchEntry(ctx context.Context, hashHex string) (*ptnjson.TxSerachEntryJson, error) {
	result := new(ptnjson.TxSerachEntryJson)
	err := ec.c.CallContext(ctx, &result, "dag_getTxSearchEntry", hashHex)
	return result, err
}

// GetPoolTxByHash
func (ec *Client) GetTxPoolTxByHash(ctx context.Context, hex string) (*ptnjson.TxPoolTxJson, error) {
	result := new(ptnjson.TxPoolTxJson)
	err := ec.c.CallContext(ctx, &result, "dag_getTxPoolTxByHash", hex)
	return result, err
}

// GetTxHashByReqId
func (ec *Client) GetTxHashByReqId(ctx context.Context, hex string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "dag_getTxHashByReqId", hex)
	return result, err
}
