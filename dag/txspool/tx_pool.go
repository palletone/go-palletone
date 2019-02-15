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
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/pkg/errors"
)

type dag interface {
	CurrentUnit() *modules.Unit
	GetUnitByHash(hash common.Hash) (*modules.Unit, error)
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
	SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription
}
type txPoolEntry struct {
	tx        *modules.Transaction
	fee       int64
	txWeight  int64
	sigOpCost int64

	enTime   int64
	enHeight int32

	feeDelta int64

	nCountWithDescendants int
	nSizeWithDescendants  int64
	nFeeWithDescendants   int

	nCountWithAncestors     int
	nSizeWithAncestors      int64
	nFeeWithAncestors       int64
	nSigOpCostWithAncestors int64
}

func newTxPoolEntry(tx *modules.Transaction) *txPoolEntry {
	return &txPoolEntry{
		tx: tx,
	}
}

// txPoolEntry :
func (txe *txPoolEntry) getSize() int64 {
	size := txe.txWeight //TODO:optimize algorithm
	return size
}

type _txPoolConfig struct {
	version int
}

//TxPool :
type _TxPool struct {
	config _txPoolConfig
	chain  dag

	mapTx     map[common.Hash]*txPoolEntry              //tx_hash  --> txPoolEntry
	mapLinks  map[common.Hash]*txLink                   //tx_hash  --> pair(tx_tancestors, tx_descendants)
	mapNextTx map[modules.OutPoint]*modules.Transaction //outpoint --> transaction from (txpool || chain)
	mapDelta  map[common.Hash]int64                     //tx_hash  --> fee amount

	mu sync.RWMutex
}

// NewTxPool :
func _NewTxPool(config _txPoolConfig, chain dag) *_TxPool {
	return &_TxPool{
		config:    config,
		chain:     chain,
		mapTx:     make(map[common.Hash]*txPoolEntry, 0),
		mapLinks:  make(map[common.Hash]*txLink, 0),
		mapNextTx: make(map[modules.OutPoint]*modules.Transaction, 0),
		mapDelta:  make(map[common.Hash]int64, 0),
	}
}

func (pool *_TxPool) calculateAncestorsInTxPool(txEntry *txPoolEntry, limitAncestorCount int64, limitAncestorSize int64, limitDescendantCount int64, limitDescendantSize int64) (*txHashSet, error) {
	parentHashes, full := getTxParentsTxHash(txEntry.tx, true, limitAncestorCount)
	if full {
		return nil, errors.New("too many unconfirmed parents")
	}

	currentSize := txEntry.getSize()
	setAncestors := newTxHashSet()
	for parentHashes.size() != 0 {
		for ph := range parentHashes.loop() {
			setAncestors.insert(ph)
			phParents := pool.mapLinks[ph].getParents()
			if phParents.size() != 0 {
				parentHashes.merge(phParents)
			}
			parentHashes.delete(ph)
		}
		//TODO:if descendant numbers over limit return err.
		if currentSize > limitAncestorSize {
			return nil, errors.New("exceeds ancestor size limit")
		}
	}
	return setAncestors, nil
}

func (pool *_TxPool) addTx(tx modules.Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	txHash := tx.Hash()

	//1. check existence
	if _, exist := pool.mapTx[txHash]; exist && nil != pool.mapTx[txHash] {
		//log.Debug("tx already exist")
		return nil
	}

	//2. update txpool
	entry := newTxPoolEntry(&tx)
	setAncestors, err := pool.calculateAncestorsInTxPool(entry, 50000, 500000000, 50000, 500000000)
	if err != nil {
		return errors.Wrap(err, "calculateAncestorsInTxPool error")
	}
	pool.mapTx[txHash] = entry

	txLink := newTxLink(setAncestors, nil)
	pool.mapLinks[txHash] = txLink

	return nil
}

type txLink struct {
	parentsSet  *txHashSet
	childrenSet *txHashSet
}

func (l *txLink) getParents() *txHashSet {
	return l.parentsSet
}

func (l *txLink) getChildern() *txHashSet {
	return l.childrenSet
}

func newTxLink(parents *txHashSet, children *txHashSet) *txLink {
	txLink := &txLink{
		parentsSet:  newTxHashSet(),
		childrenSet: newTxHashSet(),
	}

	if nil != parents {
		txLink.parentsSet.replaceBy(parents)
	}
	if nil != children {
		txLink.childrenSet.replaceBy(children)
	}
	return txLink
}

type set interface {
	exist(elem common.Hash) bool
	insert(elem common.Hash)
	delete(elem common.Hash)
	size() int64
	loop() map[common.Hash]bool
}

type txHashSet struct {
	s map[common.Hash]bool
}

func newTxHashSet() *txHashSet {
	return &txHashSet{
		s: make(map[common.Hash]bool, 0),
	}
}

// loop : return a map inside set
func (set *txHashSet) loop() map[common.Hash]bool {
	return set.s
}

func (set *txHashSet) exist(elem common.Hash) bool {
	_, status := set.s[elem]
	return status
}

func (set *txHashSet) insert(elem common.Hash) {
	set.s[elem] = true
}

func (set *txHashSet) size() int64 {
	return int64(len(set.s))
}

func (set *txHashSet) insertList(elems []common.Hash) {
	for _, elem := range elems {
		set.insert(elem)
	}
}

func (set *txHashSet) delete(elem common.Hash) {
	if set.exist(elem) {
		delete(set.s, elem)
	}
}

func (set *txHashSet) merge(rset *txHashSet) {
	for e := range rset.s {
		set.s[e] = true
	}
}

func (set *txHashSet) replaceBy(rset *txHashSet) {
	set.s = rset.s
}

/*
 * util
 */

// getAllParentsTxHash
func getTxParentsTxHash(tx *modules.Transaction, limit bool, limitNumber int64) (parentHashes *txHashSet, full bool) {
	parentHashes = newTxHashSet()
	full = false

	for _, msg := range tx.Messages() {
		if msg.App == modules.APP_PAYMENT {
			payment, ok := msg.Payload.(*modules.PaymentPayload)
			if ok {
				for _, input := range payment.Inputs {
					parentHashes.insert(input.PreviousOutPoint.TxHash)
					if limit == true && parentHashes.size()+1 > limitNumber {
						full = true
						return
					}
				}
			}
		}
	}
	return
}
