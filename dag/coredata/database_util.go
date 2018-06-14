// Copyright 2015 The go-palletone Authors
// This file is part of the go-palletone library.
//
// The go-palletone library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-palletone library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-palletone library. If not, see <http://www.gnu.org/licenses/>.

package coredata

import (
	//"bytes"
	"encoding/binary"
	//"encoding/json"
	"errors"
	//"fmt"
	"math/big"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/types"
	"github.com/palletone/go-palletone/p2p/ethdb"
	//"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/statistics/metrics"
)

// DatabaseReader wraps the Get method of a backing data store.
type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
}

// DatabaseDeleter wraps the Delete method of a backing data store.
type DatabaseDeleter interface {
	Delete(key []byte) error
}

var (
	headHeaderKey = []byte("LastHeader")
	headBlockKey  = []byte("LastBlock")
	headFastKey   = []byte("LastFast")
	trieSyncKey   = []byte("TrieSync")

	// Data item prefixes (use single byte to avoid mixing data types, avoid `i`).
	headerPrefix        = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
	tdSuffix            = []byte("t") // headerPrefix + num (uint64 big endian) + hash + tdSuffix -> td
	numSuffix           = []byte("n") // headerPrefix + num (uint64 big endian) + numSuffix -> hash
	blockHashPrefix     = []byte("H") // blockHashPrefix + hash -> num (uint64 big endian)
	bodyPrefix          = []byte("b") // bodyPrefix + num (uint64 big endian) + hash -> block body
	blockReceiptsPrefix = []byte("r") // blockReceiptsPrefix + num (uint64 big endian) + hash -> block receipts
	lookupPrefix        = []byte("l") // lookupPrefix + hash -> transaction/receipt lookup metadata
	bloomBitsPrefix     = []byte("B") // bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash -> bloom bits

	preimagePrefix = "secure-key-"              // preimagePrefix + hash -> preimage
	configPrefix   = []byte("ethereum-config-") // config prefix for the db

	// Chain index prefixes (use `i` + single byte to avoid mixing data types).
	BloomBitsIndexPrefix = []byte("iB") // BloomBitsIndexPrefix is the data table of a chain indexer to track its progress

	// used by old db, now only used for conversion
	oldReceiptsPrefix = []byte("receipts-")
	oldTxMetaSuffix   = []byte{0x01}

	ErrChainConfigNotFound = errors.New("ChainConfig not found") // general config not found error

	preimageCounter    = metrics.NewRegisteredCounter("db/preimage/total", nil)
	preimageHitCounter = metrics.NewRegisteredCounter("db/preimage/hits", nil)
)

// TxLookupEntry is a positional metadata to help looking up the data content of
// a transaction or receipt given only its hash.
type TxLookupEntry struct {
	BlockHash  common.Hash
	BlockIndex uint64
	Index      uint64
}

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// GetCanonicalHash retrieves a hash assigned to a canonical block number.
func GetCanonicalHash(db DatabaseReader, number uint64) common.Hash {
	return common.Hash{}
}

// missingNumber is returned by GetBlockNumber if no header with the
// given block hash has been stored in the database
const missingNumber = uint64(0xffffffffffffffff)

// GetBlockNumber returns the block number assigned to a block hash
// if the corresponding header is present in the database
func GetBlockNumber(db DatabaseReader, hash common.Hash) uint64 {
	return uint64(0)
}

// GetHeadHeaderHash retrieves the hash of the current canonical head block's
// header. The difference between this and GetHeadBlockHash is that whereas the
// last block hash is only updated upon a full block import, the last header
// hash is updated already at header import, allowing head tracking for the
// light synchronization mechanism.
func GetHeadHeaderHash(db DatabaseReader) common.Hash {
	return common.Hash{}
}

// GetHeadBlockHash retrieves the hash of the current canonical head block.
func GetHeadBlockHash(db DatabaseReader) common.Hash {
	return common.Hash{}
}

// GetHeadFastBlockHash retrieves the hash of the current canonical head block during
// fast synchronization. The difference between this and GetHeadBlockHash is that
// whereas the last block hash is only updated upon a full block import, the last
// fast hash is updated when importing pre-processed blocks.
func GetHeadFastBlockHash(db DatabaseReader) common.Hash {
	return common.Hash{}
}

// GetTrieSyncProgress retrieves the number of tries nodes fast synced to allow
// reportinc correct numbers across restarts.
func GetTrieSyncProgress(db DatabaseReader) uint64 {
	return uint64(0)
}

// GetHeaderRLP retrieves a block header in its raw RLP database encoding, or nil
// if the header's not found.
func GetHeaderRLP(db DatabaseReader, hash common.Hash, number uint64) rlp.RawValue {
	return rlp.RawValue{}
}

// GetHeader retrieves the block header corresponding to the hash, nil if none
// found.
func GetHeader(db DatabaseReader, hash common.Hash, number uint64) *types.Header {
	return &types.Header{}
}

// GetBodyRLP retrieves the block body (transactions and uncles) in RLP encoding.
func GetBodyRLP(db DatabaseReader, hash common.Hash, number uint64) rlp.RawValue {
	return rlp.RawValue{}
}

// GetBody retrieves the block body (transactons, uncles) corresponding to the
// hash, nil if none found.
func GetBody(db DatabaseReader, hash common.Hash, number uint64) *types.Body {
	return &types.Body{}
}

// GetTd retrieves a block's total difficulty corresponding to the hash, nil if
// none found.
func GetTd(db DatabaseReader, hash common.Hash, number uint64) *big.Int {
	return &big.Int{}
}

// GetBlock retrieves an entire block corresponding to the hash, assembling it
// back from the stored header and body. If either the header or body could not
// be retrieved nil is returned.
//
// Note, due to concurrent download of header and block body the header and thus
// canonical hash can be stored in the database but the body data not (yet).
func GetBlock(db DatabaseReader, hash common.Hash, number uint64) *types.Block {
	return &types.Block{}
}

// GetBlockReceipts retrieves the receipts generated by the transactions included
// in a block given by its hash.
func GetBlockReceipts(db DatabaseReader, hash common.Hash, number uint64) types.Receipts {
	return types.Receipts{}
}

// GetTxLookupEntry retrieves the positional metadata associated with a transaction
// hash to allow retrieving the transaction or receipt by hash.
func GetTxLookupEntry(db DatabaseReader, hash common.Hash) (common.Hash, uint64, uint64) {
	return common.Hash{}, uint64(0), uint64(0)
}

// GetTransaction retrieves a specific transaction from the database, along with
// its added positional metadata.
func GetTransaction(db DatabaseReader, hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64) {
	return &types.Transaction{}, common.Hash{}, uint64(0), uint64(0)
}

// GetReceipt retrieves a specific transaction receipt from the database, along with
// its added positional metadata.
func GetReceipt(db DatabaseReader, hash common.Hash) (*types.Receipt, common.Hash, uint64, uint64) {
	return &types.Receipt{}, common.Hash{}, uint64(0), uint64(0)
}

// GetBloomBits retrieves the compressed bloom bit vector belonging to the given
// section and bit index from the.
func GetBloomBits(db DatabaseReader, bit uint, section uint64, head common.Hash) ([]byte, error) {
	return []byte{}, nil
}

// WriteCanonicalHash stores the canonical hash for the given block number.
func WriteCanonicalHash(db ethdb.Putter, hash common.Hash, number uint64) error {
	return nil
}

// WriteHeadHeaderHash stores the head header's hash.
func WriteHeadHeaderHash(db ethdb.Putter, hash common.Hash) error {
	return nil
}

// WriteHeadBlockHash stores the head block's hash.
func WriteHeadBlockHash(db ethdb.Putter, hash common.Hash) error {
	return nil
}

// WriteHeadFastBlockHash stores the fast head block's hash.
func WriteHeadFastBlockHash(db ethdb.Putter, hash common.Hash) error {
	return nil
}

// WriteTrieSyncProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func WriteTrieSyncProgress(db ethdb.Putter, count uint64) error {
	return nil
}

// WriteHeader serializes a block header into the database.
func WriteHeader(db ethdb.Putter, header *types.Header) error {
	return nil
}

// WriteBody serializes the body of a block into the database.
func WriteBody(db ethdb.Putter, hash common.Hash, number uint64, body *types.Body) error {
	return nil
}

// WriteTd serializes the total difficulty of a block into the database.
func WriteTd(db ethdb.Putter, hash common.Hash, number uint64, td *big.Int) error {
	return nil
}

// WriteBlock serializes a block into the database, header and body separately.
func WriteBlock(db ethdb.Putter, block *types.Block) error {
	return nil
}

// WriteBlockReceipts stores all the transaction receipts belonging to a block
// as a single receipt slice. This is used during chain reorganisations for
// rescheduling dropped transactions.
func WriteBlockReceipts(db ethdb.Putter, hash common.Hash, number uint64, receipts types.Receipts) error {
	return nil
}

// WriteTxLookupEntries stores a positional metadata for every transaction from
// a block, enabling hash based transaction and receipt lookups.
func WriteTxLookupEntries(db ethdb.Putter, block *types.Block) error {
	return nil
}

// WriteBloomBits writes the compressed bloom bits vector belonging to the given
// section and bit index.
func WriteBloomBits(db ethdb.Putter, bit uint, section uint64, head common.Hash, bits []byte) {

}

// DeleteCanonicalHash removes the number to hash canonical mapping.
func DeleteCanonicalHash(db DatabaseDeleter, number uint64) {
}

// DeleteHeader removes all block header data associated with a hash.
func DeleteHeader(db DatabaseDeleter, hash common.Hash, number uint64) {
}

// DeleteBody removes all block body data associated with a hash.
func DeleteBody(db DatabaseDeleter, hash common.Hash, number uint64) {
}

// DeleteTd removes all block total difficulty data associated with a hash.
func DeleteTd(db DatabaseDeleter, hash common.Hash, number uint64) {
}

// DeleteBlock removes all block data associated with a hash.
func DeleteBlock(db DatabaseDeleter, hash common.Hash, number uint64) {
}

// DeleteBlockReceipts removes all receipt data associated with a block hash.
func DeleteBlockReceipts(db DatabaseDeleter, hash common.Hash, number uint64) {
}

// DeleteTxLookupEntry removes all transaction data associated with a hash.
func DeleteTxLookupEntry(db DatabaseDeleter, hash common.Hash) {
}

// PreimageTable returns a Database instance with the key prefix for preimage entries.
func PreimageTable(db ethdb.Database) ethdb.Database {
	return nil
}

// WritePreimages writes the provided set of preimages to the database. `number` is the
// current block number, and is used for debug messages only.
func WritePreimages(db ethdb.Database, number uint64, preimages map[common.Hash][]byte) error {
	return nil
}

// GetBlockChainVersion reads the version number from db.
func GetBlockChainVersion(db DatabaseReader) int {
	return 0
}

// WriteBlockChainVersion writes vsn as the version number to db.
func WriteBlockChainVersion(db ethdb.Putter, vsn int) {
}

// WriteChainConfig writes the chain config settings to the database.
func WriteChainConfig(db ethdb.Putter, hash common.Hash, cfg *configure.ChainConfig) error {
	return nil
}

// GetChainConfig will fetch the network settings based on the given hash.
func GetChainConfig(db DatabaseReader, hash common.Hash) (*configure.ChainConfig, error) {
	return &configure.ChainConfig{}, nil
}

// FindCommonAncestor returns the last common ancestor of two block headers
func FindCommonAncestor(db DatabaseReader, a, b *types.Header) *types.Header {
	return nil
}
func WriteBodyRLP(db ethdb.Putter, hash common.Hash, number uint64, rlp rlp.RawValue) error {
	return nil
}
