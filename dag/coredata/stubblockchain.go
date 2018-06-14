// Copyright 2014 The go-palletone Authors
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

// Package core implements the Ethereum consensus protocol.
package coredata

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/consensus"
	//"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/contracts/types"
	"github.com/palletone/go-palletone/p2p/pandb"
	"github.com/palletone/go-palletone/vm"
)

type BlockChain struct{}
type CacheConfig struct {
	Disabled      bool          // Whether to disable trie write caching (archive node)
	TrieNodeLimit int           // Memory limit (MB) at which to flush the current in-memory trie to disk
	TrieTimeLimit time.Duration // Time limit after which to flush the current in-memory trie to disk
}

func (bc *BlockChain) InsertChain(chain types.Blocks) (int, error) {
	return 0, nil
}
func (bc *BlockChain) CurrentBlock() *types.Block {
	return &types.Block{}
}
func (bc *BlockChain) HasBlock(hash common.Hash, number uint64) bool {
	return true
}
func (bc *BlockChain) HasBlockAndState(hash common.Hash, number uint64) bool {
	return true
}
func (bc *BlockChain) Export(w io.Writer) error {
	return nil
}

func (bc *BlockChain) ExportN(w io.Writer, first uint64, last uint64) error {
	return nil
}
func NewBlockChain(db pandb.Database, cacheConfig *CacheConfig, chainConfig *configure.ChainConfig, engine consensus.Engine /*, vmConfig vm.Config*/) (*BlockChain, error) {
	return &BlockChain{}, nil
}

type ChainIndexerChain interface{}
type ChainIndexerBackend interface{}

type ChainIndexer struct{}

func NewChainIndexer(chainDb, indexDb pandb.Database, backend ChainIndexerBackend, section, confirm uint64, throttling time.Duration, kind string) *ChainIndexer {
	return &ChainIndexer{}
}

func (c *ChainIndexer) Sections() (uint64, uint64, common.Hash) {
	return uint64(0), uint64(0), common.Hash{}
}

func (c *ChainIndexer) SectionHead(section uint64) common.Hash {
	return common.Hash{}
}

func (c *ChainIndexer) AddKnownSectionHead(section uint64, shead common.Hash) {

}

func (c *ChainIndexer) Start(chain ChainIndexerChain) {

}

func (c *ChainIndexer) Close() error {
	return nil
}

func (c *ChainIndexer) AddChildIndexer(indexer *ChainIndexer) {

}

var (
	// ErrKnownBlock is returned when a block to import is already known locally.
	ErrKnownBlock = errors.New("block already known")

	// ErrGasLimitReached is returned by the gas pool if the amount of gas required
	// by a transaction is higher than what's left in the block.
	ErrGasLimitReached = errors.New("gas limit reached")

	// ErrBlacklistedHash is returned if a block to import is on the blacklist.
	ErrBlacklistedHash = errors.New("blacklisted hash")

	// ErrNonceTooHigh is returned if the nonce of a transaction is higher than the
	// next one expected based on the local chain.
	ErrNonceTooHigh = errors.New("nonce too high")
)

// TxPreEvent is posted when a transaction enters the transaction pool.
type TxPreEvent struct{ Tx *types.Transaction }

// PendingLogsEvent is posted pre mining and notifies of pending logs.
type PendingLogsEvent struct {
	Logs []*types.Log
}

// PendingStateEvent is posted pre mining and notifies of pending state changes.
type PendingStateEvent struct{}

// NewMinedBlockEvent is posted when a block has been imported.
type NewMinedBlockEvent struct{ Block *types.Block }

// RemovedTransactionEvent is posted when a reorg happens
type RemovedTransactionEvent struct{ Txs types.Transactions }

// RemovedLogsEvent is posted when a reorg happens
type RemovedLogsEvent struct{ Logs []*types.Log }

type ChainEvent struct {
	Block *types.Block
	Hash  common.Hash
	Logs  []*types.Log
}

type ChainSideEvent struct {
	Block *types.Block
}

type ChainHeadEvent struct{ Block *types.Block }

/////////////genesis//////////////////////

type Genesis struct{}

/*
type Genesis struct {
	Config     *configure.ChainConfig `json:"config"`
	Nonce      uint64              `json:"nonce"`
	Timestamp  uint64              `json:"timestamp"`
	ExtraData  []byte              `json:"extraData"`
	GasLimit   uint64              `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int            `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash         `json:"mixHash"`
	Coinbase   common.Address      `json:"coinbase"`
	Alloc      GenesisAlloc        `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}
*/
// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db pandb.Database) (*types.Block, error) {
	return &types.Block{}, nil
}

// MustCommit writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func (g *Genesis) MustCommit(db pandb.Database) *types.Block {
	return &types.Block{}
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db pandb.Database, addr common.Address, balance *big.Int) *types.Block {
	return &types.Block{}
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	return &Genesis{}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	return &Genesis{}
}

// DefaultRinkebyGenesisBlock returns the Rinkeby network genesis block.
func DefaultRinkebyGenesisBlock() *Genesis {
	return &Genesis{}
}

// DeveloperGenesisBlock returns the 'gpan --dev' genesis block. Note, this must
// be seeded with the
func DeveloperGenesisBlock(period uint64, faucet common.Address) *Genesis {
	return &Genesis{}
}
func SetupGenesisBlock(db pandb.Database, genesis *Genesis) (*configure.ChainConfig, common.Hash, error) {
	return &configure.ChainConfig{}, common.Hash{}, nil
}

////////headerchain///////////
type HeaderChain struct{}

func NewHeaderChain(chainDb pandb.Database, config *configure.ChainConfig, engine consensus.Engine, procInterrupt func() bool) (*HeaderChain, error) {
	return &HeaderChain{}, nil
}
func (hc *HeaderChain) CurrentHeader() *types.Header {
	return &types.Header{}
}
func (hc *HeaderChain) GetTd(hash common.Hash, number uint64) *big.Int {
	return &big.Int{}
}

//////////EVM///////////////////
type ChainContext interface {
	Engine() consensus.Engine
	GetHeader(common.Hash, uint64) *types.Header
}

// NewEVMContext creates a new context for use in the EVM.
func NewEVMContext(msg Message, header *types.Header, chain ChainContext, author *common.Address) vm.Context {
	return vm.Context{}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.Header, chain ChainContext) func(n uint64) common.Hash {
	return func(n uint64) common.Hash { return common.Hash{} }
}

// CanTransfer checks wether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	return true
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {

}

////////////GasPool///////////////
type GasPool uint64

// AddGas makes gas available for execution.
func (gp *GasPool) AddGas(amount uint64) *GasPool {
	if uint64(*gp) > math.MaxUint64-amount {
		panic("gas pool pushed above uint64")
	}
	*(*uint64)(gp) += amount
	return gp
}

// SubGas deducts the given amount from the pool if enough gas is
// available and returns an error otherwise.
func (gp *GasPool) SubGas(amount uint64) error {
	if uint64(*gp) < amount {
		return ErrGasLimitReached
	}
	*(*uint64)(gp) -= amount
	return nil
}

// Gas returns the amount of gas remaining in the pool.
func (gp *GasPool) Gas() uint64 {
	return uint64(*gp)
}

func (gp *GasPool) String() string {
	return fmt.Sprintf("%d", *gp)
}

///////////////////state_processor////////////////////

type StateProcessor struct {
	config *configure.ChainConfig // Chain configuration options
	bc     *BlockChain            // Canonical block chain
	engine consensus.Engine       // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *configure.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{}
}

func (p *StateProcessor) Process(block *types.Block /*statedb *state.StateDB,*/, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		allLogs  []*types.Log
	)
	return receipts, allLogs, *usedGas, nil
}

func ApplyTransaction(config *configure.ChainConfig, bc *BlockChain, author *common.Address, gp *GasPool /*statedb *state.StateDB,*/, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	return nil, 0, nil
}

////////////////state_transition////////////////////
var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

type StateTransition struct {
	gp         *GasPool
	msg        Message
	gas        uint64
	gasPrice   *big.Int
	initialGas uint64
	value      *big.Int
	data       []byte
	state      vm.StateDB
	evm        *vm.EVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address

	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation, homestead bool) (uint64, error) {
	return uint64(0), nil
}
func NewStateTransition(evm *vm.EVM, msg Message, gp *GasPool) *StateTransition {
	return &StateTransition{}
}

func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) ([]byte, uint64, bool, error) {
	return []byte{}, uint64(0), true, nil
}

func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error) { return }

///////////types///////////////////

type Validator interface {
	ValidateBody(block *types.Block) error
	ValidateState(block, parent *types.Block /*state *state.StateDB,*/, receipts types.Receipts, usedGas uint64) error
}

type Processor interface {
	Process(block *types.Block /*statedb *state.StateDB,*/, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error)
}
