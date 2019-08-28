/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package shim

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/palletone/go-palletone/common"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

// Chaincode interface must be implemented by all chaincodes. The runs
// the transactions by calling these functions as specified.
type Chaincode interface {
	// Init is called during Instantiate transaction after the chaincode containerGetTxID
	// has been established for the first time, allowing the chaincode to
	// initialize its internal data
	Init(stub ChaincodeStubInterface) pb.Response

	// Invoke is called to update or query the ledger in a proposal transaction.
	// Updated state variables are not committed to the ledger until the
	// transaction is committed.
	Invoke(stub ChaincodeStubInterface) pb.Response
}

// ChaincodeStubInterface is used by deployable chaincode apps to access and
// modify their ledgers
type ChaincodeStubInterface interface {
	// GetArgs returns the arguments intended for the chaincode Init and Invoke
	// as an arrayte a of byrrays.
	GetArgs() [][]byte

	// GetStringArgs returns the arguments intended for the chaincode Init and
	// Invoke as a string array. Only use GetStringArgs if the client passes
	// arguments intended to be used as strings.
	GetStringArgs() []string

	// GetFunctionAndParameters returns the first argument as the function
	// name and the rest of the arguments as parameters in a string array.
	// Only use GetFunctionAndParameters if the client passes arguments intended
	// to be used as strings.
	GetFunctionAndParameters() (string, []string)

	// GetArgsSlice returns the arguments intended for the chaincode Init and
	// Invoke as a byte array
	GetArgsSlice() ([]byte, error)

	// GetTxID returns the tx_id of the transaction proposal, which is unique per
	// transaction and per client. See ChannelHeader in protos/common/common.proto
	// for further details.
	GetTxID() string

	// GetChannelID returns the channel the proposal is sent to for chaincode to process.
	// This would be the channel_id of the transaction proposal (see ChannelHeader
	// in protos/common/common.proto) except where the chaincode is calling another on
	// a different channel
	GetChannelID() string

	// InvokeChaincode locally calls the specified chaincode `Invoke` using the
	// same transaction context; that is, chaincode calling chaincode doesn't
	// create a new transaction message.
	// If the called chaincode is on the same channel, it simply adds the called
	// chaincode read set and write set to the calling transaction.
	// If the called chaincode is on a different channel,
	// only the Response is returned to the calling chaincode; any PutState calls
	// from the called chaincode will not have any effect on the ledger; that is,
	// the called chaincode on a different channel will not have its read set
	// and write set applied to the transaction. Only the calling chaincode's
	// read set and write set will be applied to the transaction. Effectively
	// the called chaincode on a different channel is a `Query`, which does not
	// participate in state validation checks in subsequent commit phase.
	// If `channel` is empty, the caller's channel is assumed.
	InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response

	// GetState returns the value of the specified `key` from the
	// ledger. Note that GetState doesn't read data from the writeset, which
	// has not been committed to the ledger. In other words, GetState doesn't
	// consider data modified by PutState that has not been committed.
	// If the key does not exist in the state database, (nil, nil) is returned.
	GetState(key string) ([]byte, error)
	GetGlobalState(key string) ([]byte, error)
	GetContractState(contractAddr common.Address, key string) ([]byte, error)
	GetStateByPrefix(prefix string) ([]*modules.KeyValue, error)

	// PutState puts the specified `key` and `value` into the transaction's
	// writeset as a data-write proposal. PutState doesn't effect the ledger
	// until the transaction is validated and successfully committed.
	// Simple keys must not be an empty string and must not start with null
	// character (0x00), in order to avoid range query collisions with
	// composite keys, which internally get prefixed with 0x00 as composite
	// key namespace.
	PutState(key string, value []byte) error
	PutGlobalState(key string, value []byte) error

	OutChainCall(outChainName string, method string, params []byte) ([]byte, error)

	//OutChainUtil(outChainName string, params []byte) ([]byte, error)
	//OutChainTokenOperation(outChainName string, params []byte) ([]byte, error)
	//OutChainContractOperation(outChainName string, params []byte) ([]byte, error)

	//retrun local jury's signature
	SendJury(msgType uint32, consultContent []byte, myAnswer []byte) ([]byte, error)
	//return all jury's Address and Address,format:[]JuryMsgAddr
	//type JuryMsgAddr struct {
	//	Address string
	//	Address  []byte
	//}
	RecvJury(msgType uint32, consultContent []byte, timeout uint32) ([]byte, error)

	// DelState records the specified `key` to be deleted in the writeset of
	// the transaction proposal. The `key` and its value will be deleted from
	// the ledger when the transaction is validated and successfully committed.
	DelState(key string) error
	DelGlobalState(key string) error

	// GetTxTimestamp returns the timestamp when the transaction was created. This
	// is taken from the transaction ChannelHeader, therefore it will indicate the
	// client's timestamp and will have the same value across all endorsers.
	GetTxTimestamp(rangeNumber uint32) (*timestamp.Timestamp, error)

	// SetEvent allows the chaincode to set an event on the response to the
	// proposal to be included as part of a transaction. The event will be
	// available within the transaction in the committed block regardless of the
	// validity of the transaction.
	SetEvent(name string, payload []byte) error

	//获取合约的一些配置参数
	//GetSystemConfig(filed string) (value string, err error)
	GetSystemConfig() (gp *modules.GlobalProperty, err error)
	//获取支付合约的 from 地址
	GetInvokeAddress() (invokeAddr common.Address, err error)
	//获取支付ptn数量
	GetInvokeTokens() (invokeTokens []*modules.InvokeTokens, err error)
	//获取所有的世界状态
	GetContractAllState() (states map[string]*modules.ContractStateValue, err error)
	//获取调用合约所支付的PTN手续费
	GetInvokeFees() (invokeFees *modules.AmountAsset, err error)
	//获取合约地址
	GetContractID() ([]byte, string)
	//获得某地址的Token余额
	//如果地址为空则表示当前合约
	//如果token为空则表示查询所有Token余额
	GetTokenBalance(address string, token *modules.Asset) ([]*modules.InvokeTokens, error)
	//将合约上锁定的某种Token支付出去
	PayOutToken(addr string, invokeTokens *modules.AmountAsset, lockTime uint32) error
	//获取invoke参数，包括invokeAddr,tokens,fee,funcName,params
	GetInvokeParameters() (invokeAddr common.Address, invokeTokens []*modules.InvokeTokens, invokeFees *modules.AmountAsset, funcName string, params []string, err error)
	//定义并发行一种全新的Token
	DefineToken(tokenType byte, define []byte, creator string) error
	//增发一种之前已经定义好的Token
	//如果是ERC20增发，则uniqueId为空，如果是ERC721增发，则必须指定唯一的uniqueId
	SupplyToken(assetId []byte, uniqueId []byte, amt uint64, creator string) error

	// 根据证书ID获得证书字节数据
	GetRequesterCert() (certBytes []byte, err error)
	// 验证证书是否合法, error返回的是不合法的原因
	IsRequesterCertValid() (bool, error)

	// GetStateByRange returns a range iterator over a set of keys in the
	// ledger. The iterator can be used to iterate over all keys
	// between the startKey (inclusive) and endKey (exclusive).
	// The keys are returned by the iterator in lexical order. Note
	// that startKey and endKey can be empty string, which implies unbounded range
	// query on start or end.
	// Call Close() on the returned StateQueryIteratorInterface object when done.
	// The query is re-executed during validation phase to ensure result set
	// has not changed since transaction endorsement (phantom reads detected).
	//TODO
	//GetStateByRange(startKey, endKey string) (StateQueryIteratorInterface, error)

	// GetStateByPartialCompositeKey queries the state in the ledger based on
	// a given partial composite key. This function returns an iterator
	// which can be used to iterate over all composite keys whose prefix matches
	// the given partial composite key. The `objectType` and attributes are
	// expected to have only valid utf8 strings and should not contain
	// U+0000 (nil byte) and U+10FFFF (biggest and unallocated code point).
	// See related functions SplitCompositeKey and CreateCompositeKey.
	// Call Close() on the returned StateQueryIteratorInterface object when done.
	// The query is re-executed during validation phase to ensure result set
	// has not changed since transaction endorsement (phantom reads detected).
	//TODO
	//GetStateByPartialCompositeKey(objectType string, keys []string) (StateQueryIteratorInterface, error)

	// CreateCompositeKey combines the given `attributes` to form a composite
	// key. The objectType and attributes are expected to have only valid utf8
	// strings and should not contain U+0000 (nil byte) and U+10FFFF
	// (biggest and unallocated code point).
	// The resulting composite key can be used as the key in PutState().
	//TODO
	//CreateCompositeKey(objectType string, attributes []string) (string, error)

	// SplitCompositeKey splits the specified key into attributes on which the
	// composite key was formed. Composite keys found during range queries
	// or partial composite key queries can therefore be split into their
	// composite parts.
	//TODO
	//SplitCompositeKey(compositeKey string) (string, []string, error)

	// GetQueryResult performs a "rich" query against a state database. It is
	// only supported for state databases that support rich query,
	// e.g.CouchDB. The query string is in the native syntax
	// of the underlying state database. An iterator is returned
	// which can be used to iterate (next) over the query result set.
	// The query is NOT re-executed during validation phase, phantom reads are
	// not detected. That is, other committed transactions may have added,
	// updated, or removed keys that impact the result set, and this would not
	// be detected at validation/commit time.  Applications susceptible to this
	// should therefore not use GetQueryResult as part of transactions that update
	// ledger, and should limit use to read-only chaincode operations.
	//TODO
	//GetQueryResult(query string) (StateQueryIteratorInterface, error)

	// GetHistoryForKey returns a history of key values across time.
	// For each historic key update, the historic value and associated
	// transaction id and timestamp are returned. The timestamp is the
	// timestamp provided by the client in the proposal header.
	// GetHistoryForKey requires peer configuration
	// core.ledger.history.enableHistoryDatabase to be true.
	// The query is NOT re-executed during validation phase, phantom reads are
	// not detected. That is, other committed transactions may have updated
	// the key concurrently, impacting the result set, and this would not be
	// detected at validation/commit time. Applications susceptible to this
	// should therefore not use GetHistoryForKey as part of transactions that
	// update ledger, and should limit use to read-only chaincode operations.
	//GetHistoryForKey(key string) (HistoryQueryIteratorInterface, error)

	// GetCreator returns `SignatureHeader.Creator` (e.g. an identity)
	// of the `SignedProposal`. This is the identity of the agent (or user)
	// submitting the transaction.
	//TODO
	//GetCreator() ([]byte, error)

	// GetTransient returns the `ChaincodeProposalPayload.Transient` field.
	// It is a map that contains data (e.g. cryptographic material)
	// that might be used to implement some form of application-level
	// confidentiality. The contents of this field, as prescribed by
	// `ChaincodeProposalPayload`, are supposed to always
	// be omitted from the transaction and excluded from the ledger.
	//TODO
	//GetTransient() (map[string][]byte, error)

	// GetBinding returns the transaction binding, which is used to enforce a
	// link between application data (like those stored in the transient field
	// above) to the proposal itself. This is useful to avoid possible replay
	// attacks.
	//TODO
	//GetBinding() ([]byte, error)

	// GetDecorations returns additional data (if applicable) about the proposal
	// that originated from the peer. This data is set by the decorators of the
	// peer, which append or mutate the chaincode input passed to the chaincode.
	//TODO
	//GetDecorations() map[string][]byte

	// GetSignedProposal returns the SignedProposal object, which contains all
	// data elements part of a transaction proposal.
	//TODO
	//GetSignedProposal() (*pb.SignedProposal, error)
}

// CommonIteratorInterface allows a chaincode to check whether any more result
// to be fetched from an iterator and close it when done.
//type CommonIteratorInterface interface {
// HasNext returns true if the range query iterator contains additional keys
// and values.
//HasNext() bool

// Close closes the iterator. This should be called when done
// reading from the iterator to free up resources.
//Close() error
//}

// StateQueryIteratorInterface allows a chaincode to iterate over a set of
// key/value pairs returned by range and execute query.
//type StateQueryIteratorInterface interface {
// Inherit HasNext() and Close()
//CommonIteratorInterface

// Next returns the next key and value in the range and execute query iterator.
//Next() (*queryresult.KV, error)
//}

// HistoryQueryIteratorInterface allows a chaincode to iterate over a set of
// key/value pairs returned by a history query.
//type HistoryQueryIteratorInterface interface {
//	// Inherit HasNext() and Close()
//	CommonIteratorInterface
//
//	// Next returns the next key and value in the history query iterator.
//	Next() (*queryresult.KeyModification, error)
//}

// MockQueryIteratorInterface allows a chaincode to iterate over a set of
// key/value pairs returned by range query.
// TODO: Once the execute query and history query are implemented in MockStub,
// we need to update this interface
//type MockQueryIteratorInterface interface {
//	StateQueryIteratorInterface
//}
