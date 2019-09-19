// Copyright 2014 The go-ethereum Authors
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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
	"time"
)

// Constants to match up protocol versions and messages
const (
	ptn1 = 1
)

// Official short name of the protocol used during capability negotiation.
var ProtocolName = "ptn"

// Supported versions of the ptn protocol (first is primary).
var ProtocolVersions = []uint{ptn1}

// Number of implemented message corresponding to different protocol versions.
var ProtocolLengths = []uint64{100, 8} //{17, 8}

const ProtocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message

// ptn protocol message codes
const (
	// Protocol messages belonging to ptn/1
	StatusMsg          = 0x00
	NewBlockHashesMsg  = 0x01
	TxMsg              = 0x02
	GetBlockHeadersMsg = 0x03
	BlockHeadersMsg    = 0x04
	GetBlockBodiesMsg  = 0x05
	BlockBodiesMsg     = 0x06
	NewBlockMsg        = 0x07
	NewBlockHeaderMsg  = 0x08
	VSSDealMsg         = 0x09
	VSSResponseMsg     = 0x0a
	SigShareMsg        = 0x0b
	GroupSigMsg        = 0x0c
	GetLeafNodesMsg    = 0x0d
	LeafNodesMsg       = 0x0e
	ContractMsg        = 0x0f
	ElectionMsg        = 0x10
	AdapterMsg         = 0x11

	GetNodeDataMsg = 0x20
	NodeDataMsg    = 0x21
	GetReceiptsMsg = 0x22
	ReceiptsMsg    = 0x23
)

type errCode int

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIdMismatch
	ErrGenesisBlockMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSuspendedPeer
)

func (e errCode) String() string {
	return errorToString[int(e)]
}

// XXX change once legacy code is out
var errorToString = map[int]string{
	ErrMsgTooLarge:             "Message too long",
	ErrDecode:                  "Invalid message",
	ErrInvalidMsgCode:          "Invalid message code",
	ErrProtocolVersionMismatch: "Protocol version mismatch",
	ErrNetworkIdMismatch:       "NetworkId mismatch",
	ErrGenesisBlockMismatch:    "Genesis block mismatch",
	ErrNoStatusMsg:             "No status message",
	ErrExtraStatusMsg:          "Extra status message",
	ErrSuspendedPeer:           "Suspended peer",
}

type sTxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx *modules.Transaction

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the block height when the entry was added to the the source
	// pool.
	Height int32

	// Fee is the total fee the transaction associated with the entry pays.
	Fee int64

	// FeePerKB is the fee the transaction pays in Satoshi per 1000 bytes.
	FeePerKB int64
}

// TxDesc is a descriptor containing a transaction in the mempool along with
// additional metadata.
type TxDesc struct {
	sTxDesc

	// StartingPriority is the priority of the transaction when it was added
	// to the pool.
	StartingPriority float64
}

type txPool interface {
	// AddRemotes should add the given transactions to the pool.
	Stop()
	AddLocal(tx *modules.Transaction) error
	AddLocals(txs []*modules.Transaction) []error
	AddSequenTx(tx *modules.Transaction) error
	AddSequenTxs(txs []*modules.Transaction) error
	AllHashs() []*common.Hash
	AllTxpoolTxs() map[common.Hash]*modules.TxPoolTransaction
	Content() (map[common.Hash]*modules.TxPoolTransaction, map[common.Hash]*modules.TxPoolTransaction)
	Get(hash common.Hash) (*modules.TxPoolTransaction, common.Hash)
	GetPoolTxsByAddr(addr string) ([]*modules.TxPoolTransaction, error)
	Stats() (int, int, int)
	GetSortedTxs(hash common.Hash, index uint64) ([]*modules.TxPoolTransaction, common.StorageSize)
	SendStoredTxs(hashs []common.Hash) error
	DiscardTxs(hashs []common.Hash) error
	//DiscardTx(hash common.Hash) error
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	AddRemote(tx *modules.Transaction) error
	AddRemotes([]*modules.Transaction) []error
	ProcessTransaction(tx *modules.Transaction, allowOrphan bool, rateLimit bool,
		tag txspool.Tag) ([]*txspool.TxDesc, error)
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Hash][]*modules.TxPoolTransaction, error)
	Queued() ([]*modules.TxPoolTransaction, error)
	SetPendingTxs(unit_hash common.Hash, num uint64, txs []*modules.Transaction) error
	ResetPendingTxs(txs []*modules.Transaction) error
	// SubscribeTxPreEvent should return an event subscription of
	// TxPreEvent and send events to the given channel.
	SubscribeTxPreEvent(chan<- modules.TxPreEvent) event.Subscription
	GetTxFee(tx *modules.Transaction) (*modules.AmountAsset, error)
	OutPointIsSpend(outPoint *modules.OutPoint) (bool, error)
	ValidateOrphanTx(tx *modules.Transaction) (bool, error)
}

// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	Index           *modules.ChainIndex
	GenesisUnit     common.Hash
	CurrentHeader   common.Hash
	//StableIndex     *modules.ChainIndex
}

// newBlockHashesData is the network packet for the block announcements.
type newBlockHashesData []hashOrNumber

// getBlockHeadersData represents a block header query.
type getBlockHeadersData struct {
	Origin  hashOrNumber // Block from which to retrieve headers
	Amount  uint64       // Maximum number of headers to retrieve
	Skip    uint64       // Blocks to skip between consecutive headers
	Reverse bool         // Query direction (false = rising towards latest, true = falling towards genesis)
}

// hashOrNumber is a combined field for specifying an origin block.
type hashOrNumber struct {
	Hash   common.Hash // Block hash from which to retrieve headers (excludes Number)
	Number modules.ChainIndex
}

/*
// EncodeRLP is a specialized encoder for hashOrNumber to encode only one of the
// two contained union fields.
func (hn *hashOrNumber) EncodeRLP(w io.Writer) error {
	if hn.Hash == (common.Hash{}) {
		return rlp.Encode(w, hn.Number)
	}
	//if hn.Number.Index != 0 {
	//	return fmt.Errorf("both origin hash (%x) and number (%d) provided", hn.Hash, hn.Number)
	//}
	return rlp.Encode(w, hn.Hash)
}

// DecodeRLP is a specialized decoder for hashOrNumber to decode the contents
// into either a block hash or a block number.
func (hn *hashOrNumber) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	origin, err := s.Raw()
	log.Debug("hashOrNumber", "DecodeRLP size:", size, "origin:", string(origin))

	if err == nil {
		switch {
		case size == 32:
			err = rlp.DecodeBytes(origin, &hn.Hash)
		//case size <= 8:
		default:
			err = rlp.DecodeBytes(origin, &hn.Number)
			//default:
			//	err = fmt.Errorf("invalid input size %d for origin", size)
		}
	}
	return err
}
*/
// blockBody represents the data content of a single block.
type blockBody struct {
	Transactions []*modules.Transaction // Transactions contained within a block
}

// blockBodiesData is the network packet for block content distribution.
type blockBodiesData []blockBody
