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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package txspool

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
	// rmTxChanSize is the size of channel listening to RemovedTransactionEvent.
	rmTxChanSize = 10
)

var (
	evictionInterval    = time.Minute     // Time interval to check for evictable transactions
	statsReportInterval = 8 * time.Second // Time interval to report transaction pool stats
)
var (
	// ErrInvalidSender is returned if the transaction contains an invalid signature.
	ErrInvalidSender = errors.New("invalid sender")

	// ErrNonceTooLow is returned if the nonce of a transaction is lower than the
	// one present in the local chain.
	ErrNonceTooLow = errors.New("nonce too low")

	// ErrTxFeeTooLow is returned if a transaction's tx_fee is below the value of TXFEE.
	ErrTxFeeTooLow = errors.New("txfee too low")

	// ErrUnderpriced is returned if a transaction's gas price is below the minimum
	// configured for the transaction pool.
	ErrUnderpriced = errors.New("transaction underpriced")

	// ErrReplaceUnderpriced is returned if a transaction is attempted to be replaced
	// with a different one without the required price bump.
	ErrReplaceUnderpriced = errors.New("replacement transaction underpriced")

	// ErrInsufficientFunds is returned if the total cost of executing a transaction
	// is higher than the balance of the user's account.
	ErrInsufficientFunds = errors.New("insufficient funds for gas * price + value")

	// ErrNegativeValue is a sanity error to ensure noone is able to specify a
	// transaction with a negative value.
	ErrNegativeValue = errors.New("negative value")

	// ErrOversizedData is returned if the input data of a transaction is greater
	// than some meaningful limit a user might use. This is not a consensus error
	// making the transaction invalid, rather a DOS protection.
	ErrOversizedData = errors.New("oversized data")
)

type dags interface {
	CurrentUnit() *modules.Unit
	GetUnitByHash(hash common.Hash) (*modules.Unit, error)
	//GetStoredUnitTxs(hashs chan<- common.Hash) error

	GetUtxoView(tx *modules.Transaction) (*UtxoViewpoint, error)
	SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription
	// getTxfee
	GetTxFee(pay *modules.Transaction) (*modules.InvokeFees, error)
}

// TxPoolConfig are the configuration parameters of the transaction pool.
type TxPoolConfig struct {
	NoLocals  bool          // Whether local transaction handling should be disabled
	Journal   string        // Journal of local transactions to survive node restarts
	Rejournal time.Duration // Time interval to regenerate the local transaction journal

	FeeLimit  uint64 // Minimum tx's fee  to enforce for acceptance into the pool
	PriceBump uint64 // Minimum price bump percentage to replace an already existing transaction (nonce)

	AccountSlots uint64 // Minimum number of executable transaction slots guaranteed per account
	GlobalSlots  uint64 // Maximum number of executable transaction slots for all accounts
	AccountQueue uint64 // Maximum number of non-executable transaction slots permitted per account
	GlobalQueue  uint64 // Maximum number of non-executable transaction slots for all accounts

	Lifetime   time.Duration // Maximum amount of time non-executable transaction are queued
	Removetime time.Duration // Maximum amount of time txpool transaction are removed
}

// DefaultTxPoolConfig contains the default configurations for the transaction
// pool.
var DefaultTxPoolConfig = TxPoolConfig{
	NoLocals:  false,
	Journal:   "transactions.rlp",
	Rejournal: time.Hour,

	FeeLimit:  1,
	PriceBump: 10,

	AccountSlots: 16,
	GlobalSlots:  4096,
	AccountQueue: 64,
	GlobalQueue:  1024,

	Lifetime:   3 * time.Hour,
	Removetime: 30 * time.Minute,
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *TxPoolConfig) sanitize() TxPoolConfig {
	conf := *config
	if conf.Rejournal < time.Second {
		log.Warn("Sanitizing invalid txpool journal time", "provided", conf.Rejournal, "updated", time.Second)
		conf.Rejournal = time.Second
	}
	if conf.PriceBump < 1 {
		log.Warn("Sanitizing invalid txpool price bump", "provided", conf.PriceBump, "updated", DefaultTxPoolConfig.PriceBump)
		conf.PriceBump = DefaultTxPoolConfig.PriceBump
	}
	return conf
}

type TxPool struct {
	config       TxPoolConfig
	logger       log.ILogger
	unit         dags
	txfee        *big.Int
	txFeed       event.Feed
	scope        event.SubscriptionScope
	chainHeadCh  chan modules.ChainHeadEvent
	chainHeadSub event.Subscription
	mu           sync.RWMutex

	locals  *utxoSet   // Set of local transaction to exempt from eviction rules
	journal *txJournal // Journal of local transaction to back up to disk

	beats map[modules.OutPoint]time.Time
	queue map[common.Hash]*modules.TxPoolTransaction

	pending         map[common.Hash][]*modules.TxPoolTransaction // All currently processable transactions
	all             map[common.Hash]*modules.TxPoolTransaction   // All transactions to allow lookups
	priority_priced *txPricedList                                // All transactions sorted by price and priority

	outpoints map[modules.OutPoint]*modules.TxPoolTransaction //

	wg sync.WaitGroup // for shutdown sync

	homestead bool
	quit      chan struct{} // used for exit
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

// NewTxPool creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func NewTxPool(config TxPoolConfig, unit dags, l log.ILogger) *TxPool { // chainconfig *params.ChainConfig,
	// Sanitize the input to ensure no vulnerable gas prices are set
	config = (&config).sanitize()

	// Create the transaction pool with its initial settings
	pool := &TxPool{
		config:      config,
		unit:        unit,
		logger:      l,
		queue:       make(map[common.Hash]*modules.TxPoolTransaction),
		beats:       make(map[modules.OutPoint]time.Time),
		pending:     make(map[common.Hash][]*modules.TxPoolTransaction),
		all:         make(map[common.Hash]*modules.TxPoolTransaction),
		chainHeadCh: make(chan modules.ChainHeadEvent, chainHeadChanSize),
		txfee:       new(big.Int).SetUint64(config.FeeLimit),
		outpoints:   make(map[modules.OutPoint]*modules.TxPoolTransaction),
	}
	pool.locals = newUtxoSet()
	pool.priority_priced = newTxPricedList(&pool.all)
	//pool.reset(nil, unit.CurrentUnit().Header())

	// If local transactions and journaling is enabled, load from disk
	if !config.NoLocals && config.Journal != "" {
		log.Info("Journal path:" + config.Journal)
		pool.journal = newTxJournal(config.Journal)

		if err := pool.journal.load(pool.AddLocal); err != nil {
			log.Warn("Failed to load transaction journal", "err", err)
		}
		if err := pool.journal.rotate(pool.local()); err != nil {
			log.Warn("Failed to rotate transaction journal", "err", err)
		}
	}
	// Subscribe events from blockchain
	pool.chainHeadSub = pool.unit.SubscribeChainHeadEvent(pool.chainHeadCh)

	// Start the event loop and return
	pool.wg.Add(1)
	go pool.loop()

	return pool
}

// loop is the transaction pool's main event loop, waiting for and reacting to
// outside blockchain events as well as for various reporting and transaction
// eviction events.
func (pool *TxPool) loop() {
	defer pool.wg.Done()

	// Start the stats reporting and transaction eviction tickers
	var prevPending, prevQueued, prevStales int

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()

	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()

	journal := time.NewTicker(pool.config.Rejournal)
	defer journal.Stop()
	// delete txspool's confirmed tx
	deleteTxTimer := time.NewTicker(10 * time.Minute)
	defer deleteTxTimer.Stop()

	// Track the previous head headers for transaction reorgs
	head := pool.unit.CurrentUnit()
	// Keep waiting for and reacting to the various events
	for {
		select {
		// Handle ChainHeadEvent
		case ev := <-pool.chainHeadCh:
			if ev.Unit != nil {
				pool.mu.Lock()

				pool.reset(head.Header(), ev.Unit.Header())
				head = ev.Unit

				pool.mu.Unlock()
			}
		// Be unsubscribed due to system stopped
		//would recover
		case <-pool.chainHeadSub.Err():
			return

		// Handle stats reporting ticks
		case <-report.C:
			pool.mu.RLock()
			pending, queued := pool.stats()
			stales := pool.priority_priced.stales
			pool.mu.RUnlock()

			if pending != prevPending || queued != prevQueued || stales != prevStales {
				log.Debug("Transaction pool status report", "executable", pending, "queued", queued, "stales", stales)
				prevPending, prevQueued, prevStales = pending, queued, stales
			}

		// Handle inactive account transaction eviction
		case <-evict.C:

		// Handle local transaction journal rotation ----- once a honr -----
		case <-journal.C:
			if pool.journal != nil {
				pool.mu.Lock()
				if err := pool.journal.rotate(pool.local()); err != nil {
					log.Warn("Failed to rotate local tx journal", "err", err)
				}
				pool.mu.Unlock()
			}
			// delete tx
		case <-deleteTxTimer.C:
			go pool.DeleteTx()

			// quit
		case <-pool.quit:
			log.Info("txspool are quit now", "time", time.Now().String())
			return
		}

	}
}

// reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
func (pool *TxPool) reset(oldHead, newHead *modules.Header) {

	// If we're reorging an old state, reinject all dropped transactions
	var reinject modules.Transactions

	if oldHead != nil && modules.HeaderEqual(oldHead, newHead) {
		// If the reorg is too deep, avoid doing it (will happen during fast sync)
		oldNum := oldHead.Index()
		newNum := newHead.Index()

		if depth := uint64(math.Abs(float64(oldNum) - float64(newNum))); depth > 64 {
			log.Debug("Skipping deep transaction reorg", "depth", depth)
		} else {
			// Reorg seems shallow enough to pull in all transactions into memory
			var discarded, included modules.Transactions

			var (
				rem, _ = pool.unit.GetUnitByHash(oldHead.Hash())
				add, _ = pool.unit.GetUnitByHash(newHead.Hash())
			)
			for rem.NumberU64() > add.NumberU64() {
				discarded = append(discarded, rem.Transactions()...)
				if rem, _ = pool.unit.GetUnitByHash(rem.ParentHash()[0]); rem == nil {
					log.Error("Unrooted old unit seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
			}
			for add.NumberU64() > rem.NumberU64() {
				included = append(included, add.Transactions()...)
				if add, _ = pool.unit.GetUnitByHash(add.ParentHash()[0]); add == nil {
					log.Error("Unrooted new unit seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			for rem.Hash() != add.Hash() {
				discarded = append(discarded, rem.Transactions()...)
				if rem, _ = pool.unit.GetUnitByHash(rem.ParentHash()[0]); rem == nil {
					log.Error("Unrooted old unit seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
				included = append(included, add.Transactions()...)
				if add, _ = pool.unit.GetUnitByHash(add.ParentHash()[0]); add == nil {
					log.Error("Unrooted new unit seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			reinject = modules.TxDifference(discarded, included)
		}
	}
	// Initialize the internal state to the current head
	if newHead == nil {
		newHead = pool.unit.CurrentUnit().Header() // Special case during testing
	}

	// statedb, err := pool.chain.StateAt(newHead.Root)
	// if err != nil {
	// 	log.Error("Failed to reset txpool state", "err", err)
	// 	return
	// }

	//pool.currentState = statedb
	//pool.pendingState = state.ManageState(statedb)

	// Inject any transactions discarded due to reorgs
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	pooltxs := make([]*modules.TxPoolTransaction, 0)
	for _, tx := range reinject {
		pooltxs = append(pooltxs, TxtoTxpoolTx(pool, tx))
	}

	pool.addTxsLocked(pooltxs, false)

	// validate the pool of pending transactions, this will remove
	// any transactions that have been included in the block or
	// have been invalidated because of another transaction (e.g.
	// higher gas price)
	pool.demoteUnexecutables()

	// Check the queue and move transactions over to the pending if possible
	// or remove those that have become invalid
	pool.promoteExecutables(nil)

}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) Stats() (int, int) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.stats()
}

// stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) stats() (int, int) {
	count := 0
	for _, txs := range pool.pending {
		count += len(txs)
	}
	return count, len(pool.queue)
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (pool *TxPool) Content() (map[common.Hash]*modules.Transaction, map[common.Hash]*modules.Transaction) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Hash]*modules.Transaction)
	queue := make(map[common.Hash]*modules.Transaction)
	for _, txs := range pool.pending {
		for _, tx := range txs {
			pending[tx.Tx.Hash()] = tx.Tx
		}
	}
	for hash, tx := range pool.queue {
		queue[hash] = tx.Tx
	}
	return pending, queue
}

// Pending retrieves all currently processable transactions, groupped by origin
// account and sorted by priority level. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *TxPool) Pending() (map[common.Hash][]*modules.TxPoolTransaction, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Hash][]*modules.TxPoolTransaction)
	for unit_hash, txs := range pool.pending {
		this := make([]*modules.TxPoolTransaction, 0)
		this = txs[:]
		pending[unit_hash] = this
	}
	return pending, nil
}

// AllHashs returns a slice of hashes for all of the transactions in the txpool.
func (pool *TxPool) AllHashs() []*common.Hash {
	pool.mu.RLock()
	hashs := make([]*common.Hash, len(pool.all))
	i := 0
	for hash := range pool.all {
		hashcopy := hash
		hashs[i] = &hashcopy
		i++
	}
	pool.mu.RUnlock()
	return hashs
}

func (pool *TxPool) AllTxpoolTxs() map[common.Hash]*modules.TxPoolTransaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	txs := pool.all
	return txs
}

//
func (pool *TxPool) AllTxs() []*modules.Transaction {
	pool.mu.RLock()
	txs := make([]*modules.Transaction, len(pool.all))
	i := 0
	for _, txcopy := range pool.all {
		tx := PooltxToTx(txcopy)
		txs[i] = tx
		i++
	}
	pool.mu.RUnlock()
	return txs
}
func (pool *TxPool) Count() int {
	pool.mu.RLock()
	count := len(pool.all)
	pool.mu.RUnlock()
	return count
}

// local retrieves all currently known local transactions, groupped by origin
// account and sorted by price. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *TxPool) local() map[common.Hash]*modules.TxPoolTransaction {
	txs := make(map[common.Hash]*modules.TxPoolTransaction)
	for _, list := range pool.pending {
		for _, tx := range list {
			if tx != nil {
				txs[tx.Tx.Hash()] = tx
			}
		}
	}
	return txs
}

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (price and size).
func (pool *TxPool) validateTx(tx *modules.TxPoolTransaction, local bool) error {
	// Don't accept the transaction if it already in the pool .
	hash := tx.Tx.Hash()
	if pool.isTransactionInPool(&hash) {
		return errors.New(fmt.Sprintf("already have transaction %v", tx.Tx.Hash()))
	}
	// 交易的校验， 包括inputs校验
	// dagcommon.ValidateTx(db , tx, nil )
	// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
	if tx.Tx.Size() > 32*1024 {
		return ErrOversizedData
	}
	// 交易费太低的交易，不能通过验证。
	if pool.txfee.Cmp(tx.GetTxFee()) > 0 {
		return ErrTxFeeTooLow
	}

	if len(tx.From) > 0 {
		for _, from := range tx.From {
			local = local || pool.locals.contains(*from) // tx maybe local even if the transaction arrived from the network
			if !local && pool.txfee.Cmp(tx.GetTxFee()) > 0 {
				return ErrTxFeeTooLow
			}
		}
	}

	// Make sure the transaction is signed properly

	// Verify crypto signatures for each input and reject the transaction if any don't verify.
	// 调用检测签名的接口 ： ValidateTransactionScripts

	return nil
}

// This function MUST be called with the txpool lock held (for reads).
func (pool *TxPool) isTransactionInPool(hash *common.Hash) bool {
	if _, exist := pool.all[*hash]; exist {
		return true
	}
	return false
}

// IsTransactionInPool returns whether or not the passed transaction already exists in the main pool.
func (pool *TxPool) IsTransactionInPool(hash *common.Hash) bool {
	pool.mu.RLock()
	inpool := pool.isTransactionInPool(hash)
	pool.mu.RUnlock()
	return inpool
}
func TxtoTxpoolTx(txpool ITxPool, tx *modules.Transaction) *modules.TxPoolTransaction {
	txpool_tx := new(modules.TxPoolTransaction)
	txpool_tx.Tx = tx

	for _, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				for _, script := range msg.Inputs {
					txpool_tx.From = append(txpool_tx.From, script.PreviousOutPoint)
				}
			}
		}
	}

	txpool_tx.CreationDate = time.Now()
	txpool_tx.Nonce = txpool.GetNonce(tx.Hash()) + 1
	txpool_tx.Priority_lvl = txpool_tx.GetPriorityLvl()
	txpool_tx.TxFee, _ = txpool.GetTxFee(tx)

	return txpool_tx
}

func PooltxToTx(pooltx *modules.TxPoolTransaction) *modules.Transaction {
	return pooltx.Tx
}
func PoolTxstoTxs(pool_txs []*modules.TxPoolTransaction) []modules.Transaction {
	txs := make([]modules.Transaction, 0)
	for _, p_tx := range pool_txs {
		txs = append(txs, *p_tx.Tx)
	}
	return txs
}

func (pool *TxPool) GetNonce(hash common.Hash) uint64 {
	if tx, has := pool.all[hash]; has {
		return tx.Nonce
	}
	return 0
}

// add validates a transaction and inserts it into the non-executable queue for
// later pending promotion and execution. If the transaction is a replacement for
// an already pending or queued one, it overwrites the previous and returns this
// so outer code doesn't uselessly call promote.
//
// If a newly added transaction is marked as local, its sending account will be
// whitelisted, preventing any associated transaction from being dropped out of
// the pool due to pricing constraints.
func (pool *TxPool) add(tx *modules.TxPoolTransaction, local bool) (bool, error) {
	// If the transaction is already known, discard it
	hash := tx.Tx.Hash()

	if pool.all[hash] != nil {
		pool.logger.Trace("Discarding already known transaction", "hash", hash, "old_hash", pool.all[hash].Tx.Hash())
		return false, fmt.Errorf("known transaction: %x", hash)
	}
	// If the transaction fails basic validation, discard it
	if err := pool.validateTx(tx, local); err != nil {
		pool.logger.Trace("Discarding invalid transaction", "hash", hash, "err", err)
		return false, err
	}

	if err := pool.checkPoolDoubleSpend(tx); err != nil {
		return false, err
	}

	utxoview, err := pool.FetchInputUtxos(tx.Tx)
	if err != nil {
		pool.logger.Errorf("fetchInputUtxos by txid[%s] failed:%s", tx.Tx.Hash().String(), err)
		return false, err
	}

	pool.logger.Debug("fetch utxoview info:", "utxoinfo", utxoview)
	// Check the transaction if it exists in the main chain and is not already fully spent.
	preout := modules.OutPoint{TxHash: hash}
	for i, msgcopy := range tx.Tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				for j := range msg.Outputs {
					preout.MessageIndex = uint32(i)
					preout.OutIndex = uint32(j)
					// get utxo entry , if the utxo entry is spent, then return  error.
					utxo := utxoview.LookupUtxo(preout)
					if utxo != nil && !utxo.IsSpent() {
						return false, errors.New("transaction already exists.")
					}
					utxoview.RemoveUtxo(preout)
				}
			}
		}
	}
	log.Debug("add output utxoview info: ", "utxoinfo", utxoview.entries[preout])

	// If the transaction pool is full, discard underpriced transactions
	if uint64(len(pool.all)) >= pool.config.GlobalSlots+pool.config.GlobalQueue {
		// If the new transaction is underpriced, don't accept it
		if pool.priority_priced.Underpriced(tx, pool.locals) {
			log.Trace("Discarding underpriced transaction", "hash", hash, "price", tx.GetTxFee().Int64())
			return false, ErrUnderpriced
		}
		// New transaction is better than our worse ones, make room for it
		drop := pool.priority_priced.Discard(len(pool.all)-int(pool.config.GlobalSlots+pool.config.GlobalQueue-1), pool.locals)
		for _, tx := range drop {
			log.Trace("Discarding freshly underpriced transaction", "hash", tx.Tx.Hash(), "price", tx.GetTxFee().Int64())
			pool.removeTransaction(tx, true)
		}
	}
	// If the transaction is replacing an already pending one, do directly
	txHash := tx.Tx.Hash()
	for _, lists := range pool.pending {
		for _, list := range lists {
			if list != nil {
				// New transaction is better, replace old one
				if txHash.String() == list.Tx.Hash().String() {
					if list.Priority_lvl < tx.Priority_lvl {
						//delete(pool.all, txHash)
						tx.RemStatus = true
						pool.priority_priced.Removed(txHash)
					}
					return true, nil
				}
			}
		}
	}

	// Add the transaction to the pool  and mark the referenced outpoints as spent by the pool.
	pool.all[hash] = tx
	for _, msgcopy := range tx.Tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				for _, txin := range msg.Inputs {
					if pool.outpoints == nil {
						pool.outpoints = make(map[modules.OutPoint]*modules.TxPoolTransaction)
					}
					pool.outpoints[*txin.PreviousOutPoint] = tx
				}
			}
		}
	}
	pool.priority_priced.Put(tx)
	pool.journalTx(tx)

	// We've directly injected a replacement transaction, notify subsystems
	go pool.txFeed.Send(modules.TxPreEvent{tx.Tx})

	// New transaction isn't replacing a pending one, push into queue
	replace, err := pool.enqueueTx(hash, tx)
	if err != nil {
		return false, err
	}
	// Mark local addresses and journal local transactions
	if local {
		for _, from := range tx.From {
			pool.locals.add(*from)
		}
	}
	// pool.journalTx(tx)

	pool.logger.Trace("Pooled new future transaction", "hash", hash, "repalce", replace, "err", err)
	return replace, nil
}

// enqueueTx inserts a new transaction into the non-executable transaction queue.
//
// Note, this method assumes the pool lock is held!
func (pool *TxPool) enqueueTx(hash common.Hash, tx *modules.TxPoolTransaction) (bool, error) {
	// Try to insert the transaction into the future queue

	old, ok := pool.queue[hash]
	if ok {
		// An older transaction was better, discard this
		if old.GetPriorityLvl() > tx.GetPriorityLvl() {
			return false, ErrReplaceUnderpriced
		}
		delete(pool.all, hash)
	}

	pool.all[hash] = tx
	return true, nil
}

// journalTx adds the specified transaction to the local disk journal if it is
// deemed to have been sent from a local account.
func (pool *TxPool) journalTx(tx *modules.TxPoolTransaction) {
	// Only journal if it's enabled and the transaction is local
	if len(tx.From) > 0 {
		for _, from := range tx.From {
			if pool.journal == nil || !pool.locals.contains(*from) {
				log.Trace("Pool journal is nil.", "journal", pool.journal.path)
				return
			}
		}
	}

	if err := pool.journal.insert(tx); err != nil {
		log.Warn("Failed to journal local transaction", "err", err)
	}
}

// promoteTx adds a transaction to the pending (processable) list of transactions.
//
// Note, this method assumes the pool lock is held!
func (pool *TxPool) promoteTx(hash common.Hash, tx *modules.TxPoolTransaction) {
	// Try to insert the transaction into the pending queue
	tx_hash := tx.Tx.Hash()
	old := new(modules.TxPoolTransaction)
	for _, lists := range pool.pending {
		for _, this := range lists {
			if this.Tx.Hash().String() == tx_hash.String() {
				old := this
				if old.Pending || old.Confirmed {
					// An older transaction was better, discard this
					//delete(pool.all, hash)
					old.RemStatus = true
					pool.all[hash] = old
					pool.priority_priced.Removed(hash)
					return
				}
			}
		}
	}

	// Otherwise discard any previous transaction and mark this
	if old.Tx != nil {
		//delete(pool.all, old.Tx.Hash())
		pool.priority_priced.Removed(old.Tx.Hash())
	}
	// Failsafe to work around direct pending inserts (tests)
	if pool.all[tx_hash] == nil {
		tx.Pending = true
		pool.all[tx_hash] = tx
		list := pool.pending[hash]
		if list == nil {
			list = make([]*modules.TxPoolTransaction, 0)
		}
		list = append(list, tx)
		pool.pending[hash] = list

	} else {
		tx.Pending = true
		list := pool.pending[hash]
		if list == nil {
			list = make([]*modules.TxPoolTransaction, 0)
			list = append(list, tx)
		} else {
			var exist bool
			for i, this := range list {
				if this.Tx.Hash().String() == tx.Tx.Hash().String() {
					list[i] = tx
					exist = true
					break
				}
			}
			if !exist {
				list = append(list, tx)
			}
		}
		pool.pending[hash] = list
		pool.all[hash] = old
	}
	// Set the potentially new pending nonce and notify any subsystems of the new tx
	if len(tx.From) > 0 {
		for _, from := range tx.From {
			pool.beats[*from] = time.Now()
		}
	}

	go pool.txFeed.Send(modules.TxPreEvent{tx.Tx})
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one in the mean time, ensuring it goes around the local
// pricing constraints.
func (pool *TxPool) AddLocal(tx *modules.TxPoolTransaction) error {
	//tx.SetPriorityLvl(tx.GetPriorityLvl())
	return pool.addTx(tx, !pool.config.NoLocals)
}

// AddRemote enqueues a single transaction into the pool if it is valid. If the
// sender is not among the locally tracked ones, full pricing constraints will
// apply.
func (pool *TxPool) AddRemote(tx *modules.Transaction) error {
	pool_tx := TxtoTxpoolTx(pool, tx)
	return pool.addTx(pool_tx, false)
}

// AddLocals enqueues a batch of transactions into the pool if they are valid,
// marking the senders as a local ones in the mean time, ensuring they go around
// the local pricing constraints.
func (pool *TxPool) AddLocals(txs []*modules.TxPoolTransaction) []error {
	return pool.addTxs(txs, !pool.config.NoLocals)
}

// AddRemotes enqueues a batch of transactions into the pool if they are valid.
// If the senders are not among the locally tracked ones, full pricing constraints
// will apply.
func (pool *TxPool) AddRemotes(txs []*modules.Transaction) []error {
	pool_txs := make([]*modules.TxPoolTransaction, 0)
	for _, tx := range txs {
		pool_txs = append(pool_txs, TxtoTxpoolTx(pool, tx))
	}
	return pool.addTxs(pool_txs, false)
}

type Tag uint64

func (mp *TxPool) ProcessTransaction(tx *modules.Transaction, allowOrphan bool, rateLimit bool, tag Tag) ([]*TxDesc, error) {
	//log.Trace("Processing transaction %v", tx.Hash())

	// Protect concurrent access.
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Potentially accept the transaction to the memory pool.
	missingParents, txD, err := mp.maybeAcceptTransaction(tx, true, rateLimit, false)
	if err != nil {
		log.Info("txpool", "accept transaction err:", err)
		return nil, err
	}
	missingParents = missingParents
	txD = txD

	// if len(missingParents) == 0 {
	// 	// Accept any orphan transactions that depend on this
	// 	// transaction (they may no longer be orphans if all inputs
	// 	// are now available) and repeat for those accepted
	// 	// transactions until there are no more.
	// 	newTxs := mp.processOrphans(tx)
	// 	acceptedTxs := make([]*TxDesc, len(newTxs)+1)

	// 	// Add the parent transaction first so remote nodes
	// 	// do not add orphans.
	// 	acceptedTxs[0] = txD
	// 	copy(acceptedTxs[1:], newTxs)

	// 	return acceptedTxs, nil
	// }

	// The transaction is an orphan (has inputs missing).  Reject
	// it if the flag to allow orphans is not set.
	/*if !allowOrphan {
		// Only use the first missing parent transaction in
		// the error message.
		//
		// NOTE: RejectDuplicate is really not an accurate
		// reject code here, but it matches the reference
		// implementation and there isn't a better choice due
		// to the limited number of reject codes.  Missing
		// inputs is assumed to mean they are already spent
		// which is not really always the case.
		str := fmt.Sprintf("orphan transaction %v references "+
			"outputs of unknown or fully-spent "+
			"transaction %v", tx.Hash(), missingParents[0])
		return nil, txRuleError(wire.RejectDuplicate, str)
	}*/

	// Potentially add the orphan transaction to the orphan pool.
	//err = mp.maybeAddOrphan(tx, tag)
	return nil, nil
}

func IsCoinBase(tx *modules.Transaction) bool {
	if len(tx.TxMessages) != 1 {
		return false
	}
	msg, ok := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	if !ok {
		return false
	}
	prevOut := msg.Inputs[0].PreviousOutPoint
	if prevOut.TxHash != (common.Hash{}) {
		return false
	}
	return true
}

// maybeAcceptTransaction is the internal function which implements the public
// MaybeAcceptTransaction.  See the comment for MaybeAcceptTransaction for
// more details.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) maybeAcceptTransaction(tx *modules.Transaction, isNew, rateLimit, rejectDupOrphans bool) ([]*common.Hash, *TxDesc, error) {
	txHash := tx.Hash()
	// Don't accept the transaction if it already exists in the pool.  This
	// applies to orphan transactions as well when the reject duplicate
	// orphans flag is set.  This check is intended to be a quick check to
	// weed out duplicates.
	if mp.isTransactionInPool(&txHash) {
		str := fmt.Sprintf("already have transaction %v", txHash)
		//str = str
		log.Info("txpool", "", str)
		return nil, nil, nil //txRuleError(RejectDuplicate, str)
	}

	// Perform preliminary sanity checks on the transaction.  This makes
	// use of blockchain which contains the invariant rules for what
	// transactions are allowed into blocks.
	err := CheckTransactionSanity(tx)
	if err != nil {
		log.Info("txpool", "Check Transaction Sanity err:", err)
		return nil, nil, err
	}

	// A standalone transaction must not be a coinbase transaction.
	if IsCoinBase(tx) {
		str := fmt.Sprintf("transaction %v is an individual coinbase",
			txHash)
		//str = str
		log.Info("txpool", "", str)
		return nil, nil, nil //txRuleError(RejectInvalid, str)
	}

	// The transaction may not use any of the same outputs as other
	// transactions already in the pool as that would ultimately result in a
	// double spend.  This check is intended to be quick and therefore only
	// detects double spends within the transaction pool itself.  The
	// transaction could still be double spending coins from the main chain
	// at this point.  There is a more in-depth check that happens later
	// after fetching the referenced transaction inputs from the main chain
	// which examines the actual spend data and prevents double spends.
	txpooltx := TxtoTxpoolTx(mp, tx)
	err = mp.checkPoolDoubleSpend(txpooltx)
	if err != nil {
		log.Info("txpool", "check PoolD oubleSpend err:", err)
		return nil, nil, err
	}

	// Fetch all of the unspent transaction outputs referenced by the inputs
	// to this transaction.  This function also attempts to fetch the
	// transaction itself to be used for detecting a duplicate transaction
	// without needing to do a separate lookup.
	/*utxoView, err := mp.fetchInputUtxos(tx)
	if err != nil {
		return nil, nil, err
	}

	// Don't allow the transaction if it exists in the main chain and is not
	// not already fully spent.
	prevOut := wire.OutPoint{Hash: *txHash}
	for txOutIdx := range tx.MsgTx().TxOut {
		prevOut.Index = uint32(txOutIdx)
		entry := utxoView.LookupEntry(prevOut)
		if entry != nil && !entry.IsSpent() {
			return nil, nil, txRuleError(wire.RejectDuplicate,
				"transaction already exists")
		}
		utxoView.RemoveEntry(prevOut)
	}*/

	// NOTE: if you modify this code to accept non-standard transactions,
	// you should add code here to check that the transaction does a
	// reasonable number of ECDSA signature verifications.

	/*
		// Verify crypto signatures for each input and reject the transaction if
		// any don't verify.
		err = blockchain.ValidateTransactionScripts(tx, utxoView,
			txscript.StandardVerifyFlags, mp.cfg.SigCache,
			mp.cfg.HashCache)
		if err != nil {
			if cerr, ok := err.(blockchain.RuleError); ok {
				return nil, nil, chainRuleError(cerr)
			}
			return nil, nil, err
		}*/

	// Add to transaction pool.
	//txD := mp.addTransaction(utxoView, tx, bestHeight, txFee)

	//log.Debugf("Accepted transaction %v (pool size: %v)", txHash,
	//	len(mp.pool))

	return nil, nil, nil
}

// addTx enqueues a single transaction into the pool if it is valid.
func (pool *TxPool) addTx(tx *modules.TxPoolTransaction, local bool) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Try to inject the transaction and update any state
	replace, err := pool.add(tx, local)
	if err != nil {
		return err
	}
	// If we added a new transaction, run promotion checks and return
	if !replace {
		if len(tx.From) > 0 {
			for _, from := range tx.From { // already validated
				pool.promoteExecutables([]modules.OutPoint{*from})
			}
		}

	}
	return nil
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (pool *TxPool) addTxs(txs []*modules.TxPoolTransaction, local bool) []error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.addTxsLocked(txs, local)
}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (pool *TxPool) addTxsLocked(txs []*modules.TxPoolTransaction, local bool) []error {
	// Add the batch of transaction, tracking the accepted ones
	dirty := make(map[modules.OutPoint]struct{})
	errs := make([]error, len(txs))
	for i, tx := range txs {
		var replace bool
		if replace, errs[i] = pool.add(tx, local); errs[i] == nil {
			if !replace {
				if len(tx.From) > 0 {
					for _, from := range tx.From { // already validated
						dirty[*from] = struct{}{}
					}
				}

			}
		}
	}
	// Only reprocess the internal state if something was actually added
	if len(dirty) > 0 {
		addrs := make([]modules.OutPoint, 0, len(dirty))
		for addr := range dirty {
			addrs = append(addrs, addr)
		}
		pool.promoteExecutables(addrs)
	}
	return errs
}

type TxStatus uint

const (
	TxStatusUnknown TxStatus = iota
	TxStatusQueued
	TxStatusPending
	TxStatusIncluded
)

// Status returns the status (unknown/pending/queued) of a batch of transactions
// identified by their hashes.
func (pool *TxPool) Status(hashes []common.Hash) []TxStatus {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	status := make([]TxStatus, len(hashes))
	for i, hash := range hashes {
		if tx := pool.all[hash]; tx != nil {
			//from := modules.RSVtoAddress(tx) // already validated
			if pool.queue[hash] != nil { //&& pool.pending[tx.TxHash].txs.items[tx.Nonce()] != nil
				status[i] = TxStatusQueued
			} else {
				status[i] = TxStatusPending
			}
		}
	}
	return status
}

// Get returns a transaction if it is contained in the pool
// and nil otherwise.
func (pool *TxPool) Get(hash common.Hash) (*modules.TxPoolTransaction, common.Hash) {
	// pool.mu.RLock()
	// defer pool.mu.RUnlock()

	tx := pool.all[hash]
	log.Debug("get tx info by hash in txpool... ", "info", tx)
	var u_hash common.Hash
	pending, err := pool.Pending()
	if err == nil {
		for unit_hash, txs := range pending {
			for _, p_tx := range txs {
				if p_tx.Tx.Hash() == hash {
					log.Debug("get tx info by hash in txpool... tx in unit hash:", "unit_hash", unit_hash, "p_tx", p_tx)
					return p_tx, unit_hash
				}
			}
		}
	}
	return tx, u_hash
}

// DeleteTx
func (pool *TxPool) DeleteTx() error {
	pool.mu.Lock()
	for hash, tx := range pool.all {
		if !tx.Confirmed {
			if tx.CreationDate.Add(DefaultTxPoolConfig.Lifetime).Before(time.Now()) {
				continue
			} else {
				// delete
				log.Debug("delete the non confirmed tx(overtime).", "tx_hash", tx.Tx.Hash())
				pool.DeleteTxByHash(hash)
			}
		}
		if tx.CreationDate.Add(DefaultTxPoolConfig.Removetime).After(time.Now()) {
			// delete
			log.Debug("delete the confirmed tx.", "tx_hash", tx.Tx.Hash())
			pool.DeleteTxByHash(hash)
		}
	}
	pool.mu.Unlock()
	return nil
}

func (pool *TxPool) DeleteTxByHash(hash common.Hash) error {
	tx, ok := pool.all[hash]
	if !ok {
		return errors.New(fmt.Sprintf("the tx(%s) isn't exist.", hash.String()))
	}
	log.Debug("delete the tx.", "time", time.Now().Second()-tx.CreationDate.Second(), "hash", hash.String())
	pool.priority_priced.Removed(hash)
	delete(pool.all, hash)
	// Remove the transaction from the pending lists and reset the account nonce
	for unit_hash, list := range pool.pending {
		for i, tx := range list {
			if tx.Tx.Hash().String() == hash.String() {
				newList := make([]*modules.TxPoolTransaction, 0)
				if i > 0 {
					newList = append(newList, list[:i]...)
				}
				if len(list) > i+1 {
					newList = append(newList, list[i+1:]...)
				}
				pool.pending[unit_hash] = newList
				if len(tx.From) > 0 {
					for _, from := range tx.From {
						delete(pool.beats, *from)
					}
				}
				// delete outpoints 's
				for _, msg := range tx.Tx.Messages() {
					if msg.App == modules.APP_PAYMENT {
						payment, ok := msg.Payload.(*modules.PaymentPayload)
						if ok {
							for _, input := range payment.Inputs {
								delete(pool.outpoints, *input.PreviousOutPoint)
							}
						}
					}
				}
				break
			}
		}
	}
	return nil
}

// removeTx removes a single transaction from the queue, moving all subsequent
// transactions back to the future queue.
func (pool *TxPool) removeTx(hash common.Hash) {
	// Fetch the transaction we wish to delete
	tx, ok := pool.all[hash]
	if !ok {
		return
	}
	if tx.Tx.Hash() != hash {
		tx.Confirmed = true
		pool.all[tx.Tx.Hash()] = tx
		//delete(pool.all, tx.Tx.Hash())
	}
	// Remove it from the list of known transactions
	//delete(pool.all, hash)

	pool.priority_priced.Removed(hash)
	tx.Confirmed = true
	pool.all[hash] = tx
	// Remove the transaction from the pending lists and reset the account nonce
	for unit_hash, list := range pool.pending {
		for i, tx := range list {
			if tx.Tx.Hash().String() == hash.String() {
				newList := make([]*modules.TxPoolTransaction, 0)
				if i > 0 {
					newList = append(newList, list[:i]...)
				}
				if len(list) > i+1 {
					newList = append(newList, list[i+1:]...)
				}
				pool.pending[unit_hash] = newList
				if len(tx.From) > 0 {
					for _, from := range tx.From {
						delete(pool.beats, *from)
					}
				}
				break
			}
		}
	}

	// if removed, invalids := pending.Remove(tx); removed {
	// 	// If no more pending transactions are left, remove the list
	// 	if pending.Empty() {
	// 		delete(pool.pending, hash)
	// 		from := modules.MsgstoAddress(tx.TxMessages)
	// 		delete(pool.beats, from)
	// 	}
	// 	// Postpone any invalidated transactions
	// 	for _, tx := range invalids {
	// 		pool.enqueueTx(tx.TxHash, tx)
	// 	}
	// 	return
	// }

}
func (pool *TxPool) RemoveTxs(hashs []common.Hash) {
	for _, hash := range hashs {
		pool.removeTx(hash)
	}
}

func (pool *TxPool) removeTransaction(tx *modules.TxPoolTransaction, removeRedeemers bool) {
	hash := tx.Tx.Hash()
	if removeRedeemers {
		// Remove any transactions whitch rely on this one.
		for i, msgcopy := range tx.Tx.TxMessages {
			if msgcopy.App == modules.APP_PAYMENT {
				if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
					for j := uint32(0); j < uint32(len(msg.Outputs)); j++ {
						preout := modules.OutPoint{TxHash: hash, MessageIndex: uint32(i), OutIndex: j}
						if pooltxRedeemer, exist := pool.outpoints[preout]; exist {
							pool.removeTransaction(pooltxRedeemer, true)
						}
					}
				}
			}
		}
	}
	// Remove the transaction from the pending lists and reset the account nonce
	for unit_hash, list := range pool.pending {
		for i, tx := range list {
			if tx.Tx.Hash().String() == hash.String() {
				newList := make([]*modules.TxPoolTransaction, 0)
				if i > 0 {
					newList = append(newList, list[:i]...)
				}
				if len(list) > i+1 {
					newList = append(newList, list[i+1:]...)
				}
				pool.pending[unit_hash] = newList
				if len(tx.From) > 0 {
					for _, from := range tx.From {
						delete(pool.beats, *from)
					}
				}
				break
			}
		}
	}

	// Remove the transaction if needed.
	if pooltx, exist := pool.all[hash]; exist {
		// mark the referenced outpoints as unspent by the pool.
		for _, msgcopy := range pooltx.Tx.TxMessages {
			if msgcopy.App == modules.APP_PAYMENT {
				if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
					for _, input := range msg.Inputs {
						delete(pool.outpoints, *input.PreviousOutPoint)
					}
				}
			}
		}
		//delete(pool.all, hash)
		tx.Confirmed = true
		pool.all[hash] = tx
		pool.priority_priced.Removed(hash)
	}
}
func (pool *TxPool) RemoveTransaction(hash common.Hash, removeRedeemers bool) {
	pool.mu.Lock()
	if tx, exist := pool.all[hash]; exist {
		pool.removeTransaction(tx, removeRedeemers)
	} else {
		pool.removeTx(hash)
	}

	pool.mu.Unlock()
}

// RemoveDoubleSpends removes all transactions whitch spend outpoints spent by the passed
// transaction from the memory pool. Removing those transactions then leads to removing all
// transaction whitch rely on them, recursively. This is necessary when a blocks is connected
// to the main chain because the block may contain transactions whitch were previously unknow to
// the memory pool.
func (pool *TxPool) RemoveDoubleSpends(tx *modules.Transaction) {
	pool.mu.Lock()
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			inputs := msg.Payload.(*modules.PaymentPayload)
			for _, input := range inputs.Inputs {
				if tx, ok := pool.outpoints[*input.PreviousOutPoint]; ok {
					pool.removeTransaction(tx, true)
				}
			}
		}
	}
	pool.mu.Unlock()
}

func (pool *TxPool) checkPoolDoubleSpend(tx *modules.TxPoolTransaction) error {
	for _, msg := range tx.Tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			inputs, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				continue
			}
			if inputs != nil {
				for _, input := range inputs.Inputs {
					if input == nil {
						break
					}
					//if tx, ok := pool.outpoints[*input.PreviousOutPoint]; ok {
					//	str := fmt.Sprintf("output %v already spent by "+
					//		"transaction %x in the memory pool",
					//		input.PreviousOutPoint, tx.Tx.Hash())
					//	return errors.New(str)
					//}

					if _, err := pool.OutPointIsSpend(input.PreviousOutPoint); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (pool *TxPool) OutPointIsSpend(outPoint *modules.OutPoint) (bool, error) {
	if tx, ok := pool.outpoints[*outPoint]; ok {
		str := fmt.Sprintf("output %v already spent by "+
			"transaction %x in the memory pool",
			outPoint, tx.Tx.Hash())
		return true, errors.New(str)
	}

	return false, nil
}

// CheckSpend checks whether the passed outpoint is already spent by a transaction in the txpool
func (pool *TxPool) CheckSpend(output modules.OutPoint) *modules.Transaction {
	pool.mu.RLock()
	tx := pool.outpoints[output]
	pool.mu.RUnlock()
	return tx.Tx
}
func (pool *TxPool) FetchInputUtxos(tx *modules.Transaction) (*UtxoViewpoint, error) {
	utxoView, err := pool.unit.GetUtxoView(tx)
	if err != nil {
		fmt.Println("getUtxoView is error,", err)
		return nil, err
	}
	//TODO   spent input utxo, and add output utxo.
	for _, utxo := range utxoView.entries {
		utxo.Spend()
	}
	// Attempt to populate any missing inputs from the transaction pool.
	for i, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				for _, txIn := range msg.Inputs {
					preout := txIn.PreviousOutPoint
					utxo := utxoView.LookupUtxo(*preout)
					if utxo != nil && utxo.IsSpent() {
						continue
					}
					// attempt to populate any missing inputs form the tx pool.
					if pooltx, exist := pool.all[preout.TxHash]; exist {
						utxoView.AddTxOut(pooltx.Tx, uint32(i), preout.OutIndex)
					}
				}
			}
		}
	}
	return utxoView, nil
}

// promoteExecutables moves transactions that have become processable from the
// future queue to the set of pending transactions. During this process, all
// invalidated transactions (low nonce, low balance) are deleted.
func (pool *TxPool) promoteExecutables(accounts []modules.OutPoint) {

	// If the pending limit is overflown, start equalizing allowances
	pending := 0
	for _, list := range pool.pending {
		pending += len(list)
	}
	if uint64(pending) > pool.config.GlobalSlots {
		// Assemble a spam order to penalize large transactors first
		spammers := prque.New()
		for _, list := range pool.pending {
			// Only evict transactions from high rollers
			for i, tx := range list {
				spammers.Push(tx.Tx.Hash(), float32(i))
			}
		}
		// Gradually drop transactions from offenders
		offenders := []common.Hash{}
		for uint64(pending) > pool.config.GlobalSlots && !spammers.Empty() {
			// Retrieve the next offender if not local address
			offender, _ := spammers.Pop()
			offenders = append(offenders, offender.(common.Hash))

			// Equalize balances until all the same or below threshold
			if len(offenders) > 1 {
				// Calculate the equalization threshold for all current offenders

				// Iteratively reduce all offenders until below limit or threshold reached
				for uint64(pending) > pool.config.GlobalSlots {
					for i := 0; i < len(offenders)-1; i++ {
						for _, list := range pool.pending {
							for _, tx := range list {
								hash := tx.Tx.Hash()
								if offenders[i].String() == hash.String() {
									// Drop the transaction from the global pools too
									delete(pool.all, hash)
									pool.priority_priced.Removed(hash)
									log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
									pending--
									break
								}
							}
						}
					}
				}
			}
		}
		// If still above threshold, reduce to limit or min allowance
		if uint64(pending) > pool.config.GlobalSlots && len(offenders) > 0 {
			for uint64(pending) > pool.config.GlobalSlots {
				for _, addr := range offenders {
					for _, list := range pool.pending {
						for _, tx := range list {
							hash := tx.Tx.Hash()
							if addr.String() == hash.String() {
								delete(pool.all, hash)
								pool.priority_priced.Removed(hash)
								log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
								pending--
								break
							}
						}
					}
				}
			}
		}
	}
}

// demoteUnexecutables removes invalid and processed transactions from the pools
// executable/pending queue and any subsequent transactions that become unexecutable
// are moved back into the future queue.
func (pool *TxPool) demoteUnexecutables() {
	// Iterate over all accounts and demote any non-executable transactions
	for hash, tx := range pool.queue {
		// Delete the entire queue entry if it became empty.
		if tx == nil {
			delete(pool.queue, hash)
			if len(tx.From) > 0 {
				for _, from := range tx.From {
					delete(pool.beats, *from)
				}
			}

		}
	}
}

// Stop terminates the transaction pool.
func (pool *TxPool) Stop() {
	// Unsubscribe all subscriptions registered from txpool
	fmt.Println("stop start.", time.Now())
	pool.scope.Close()
	fmt.Println("scope closed.", time.Now())
	// Unsubscribe subscriptions registered from blockchain
	pool.chainHeadSub.Unsubscribe()
	pool.wg.Wait()
	fmt.Println("journal close...")
	if pool.journal != nil {
		pool.journal.close()
	}
	log.Info("Transaction pool stopped")
}

// addressByHeartbeat is an account address tagged with its last activity timestamp.
type addressByHeartbeat struct {
	address   common.Address
	heartbeat time.Time
}

type addresssByHeartbeat []addressByHeartbeat

func (a addresssByHeartbeat) Len() int           { return len(a) }
func (a addresssByHeartbeat) Less(i, j int) bool { return a[i].heartbeat.Before(a[j].heartbeat) }
func (a addresssByHeartbeat) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

/******     utxoSet  *****/

// utxoSet is simply a set of addresses to check for existence
type utxoSet struct {
	utxos map[modules.OutPoint]struct{}
}

// newutxoSet creates a new address set with an associated signer for sender
// derivations.
func newUtxoSet() *utxoSet {
	return &utxoSet{
		utxos: make(map[modules.OutPoint]struct{}),
	}
}

// contains checks if a given address is contained within the set.
func (as *utxoSet) contains(addr modules.OutPoint) bool {
	if addr.IsEmpty() {
		return false
	}
	_, exist := as.utxos[addr]
	return exist
}

// containsTx checks if the sender of a given tx is within the set. If the sender
// cannot be derived, this method returns false.
func (as *utxoSet) containsTx(tx *modules.TxPoolTransaction) bool {
	// if addr, err := modules.Sender(as.signer, tx); err == nil {
	// 	return as.contains(addr)
	// }
	if len(tx.From) == 0 {
		return false
	}
	for _, from := range tx.From {
		if !as.contains(*from) {
			return false
		}
	}
	return true
}

// add inserts a new address into the set to track.
func (as *utxoSet) add(addr modules.OutPoint) {
	as.utxos[addr] = struct{}{}
}

func (pool *TxPool) SendStoredTxs(hashs []common.Hash) error {
	pool.RemoveTxs(hashs)
	return nil
}

/******  end utxoSet  *****/
//  这个接口后期需要调整， 需要先将all 进行排序， 然后按序从前到后一次取出足够多tx。
func (pool *TxPool) GetSortedTxs(hash common.Hash) ([]*modules.TxPoolTransaction, common.StorageSize) {
	var total common.StorageSize
	list := make([]*modules.TxPoolTransaction, 0)
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	for {
		tx := pool.priority_priced.Get()
		if tx == nil {
			log.Debug("Txspool get  priority_pricedtx failed.", "error", "tx is null")
			break
			//continue
		} else {
			log.Debug("Txspool get  priority_pricedtx success.", "tx_info", tx)
			if !tx.Pending && !tx.Confirmed {
				// dagconfig.DefaultConfig.UnitTxSize = 1024 * 16
				if total += tx.Tx.Size(); total <= common.StorageSize(dagconfig.DefaultConfig.UnitTxSize) {
					list = append(list, tx)
					// add  pending
					pool.promoteTx(hash, tx)
				} else {
					total = total - tx.Tx.Size()
					break
				}
			}
		}
	}
	return list, total
}

// SubscribeTxPreEvent registers a subscription of TxPreEvent and
// starts sending event to the given channel.
func (pool *TxPool) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}

func (pool *TxPool) GetTxFee(tx *modules.Transaction) (*modules.InvokeFees, error) {
	return pool.unit.GetTxFee(tx)
}
