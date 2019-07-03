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
	"sort"
	"strconv"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

type priorityHeap []*modules.TxPoolTransaction

func (h priorityHeap) Len() int { return len(h) }
func (h priorityHeap) Less(i, j int) bool {
	ap, _ := strconv.ParseFloat(h[i].Priority_lvl, 64)
	bp, _ := strconv.ParseFloat(h[j].Priority_lvl, 64)
	return ap > bp
}
func (h priorityHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *priorityHeap) Push(x interface{}) {
	item := x.(*modules.TxPoolTransaction)
	*h = append(*h, item)
	sort.Sort(*h)
}

func (h *priorityHeap) LastPop() interface{} {
	old := *h
	n := len(old)
	if n <= 0 {
		return nil
	}
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
func (h *priorityHeap) Pop() interface{} {
	old := *h
	n := len(old)
	if n <= 0 {
		return nil
	}
	x := old[0]
	*h = old[1:]
	return x
}

// txPricedList is a price-sorted heap to allow operating on transactions pool
// contents in a price-incrementing way.
type txPrioritiedList struct {
	all    *sync.Map     // Pointer to the map of all transactions
	items  *priorityHeap // Heap of priority of all the stored transactions
	stales int           // Number of stale priority points to (re-heap trigger)
	mu     sync.RWMutex
}

// newTxPricedList creates a new price-sorted transaction heap.
func newTxPrioritiedList(all *sync.Map) *txPrioritiedList {
	return &txPrioritiedList{
		all:   all,
		items: new(priorityHeap),
	}
}

// Put inserts a new transaction into the heap.
func (l *txPrioritiedList) Put(tx *modules.TxPoolTransaction) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items.Push(tx)
}
func (l *txPrioritiedList) Get() *modules.TxPoolTransaction {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for len(*l.items) > 0 {
		tx := l.items.Pop().(*modules.TxPoolTransaction)
		if _, ok := (*l.all).Load(tx.Tx.Hash()); !ok {
			continue
		}
		if tx.Pending || tx.Discarded {
			continue
		}
		return tx
	}
	return nil
}
func (l *txPrioritiedList) All() map[common.Hash]*modules.TxPoolTransaction {
	txs := make(map[common.Hash]*modules.TxPoolTransaction)
	(*l.all).Range(func(k, v interface{}) bool {
		var hash common.Hash
		hash.SetBytes((k.(common.Hash)).Bytes())
		tx := v.(*modules.TxPoolTransaction)
		txs[hash] = tx
		return true
	})
	return txs
}

// Removed notifies the prices transaction list that an old transaction dropped
// from the pool. The list will just keep a counter of stale objects and update
// the heap if a large enough ratio of transactions go stale.
func (l *txPrioritiedList) Removed() {
	// Bump the stale counter, but exit if still too low (< 20%)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stales++
	if l.stales <= len(*l.items)/5 {
		return
	}
	// Seems we've reached a critical number of stale transactions, reheap
	reheap := make(priorityHeap, 0)

	l.stales, l.items = 0, &reheap
	all := l.All()
	for _, tx := range all {
		if !tx.Pending && !tx.Discarded {
			*l.items = append(*l.items, tx)
		}
	}
	sort.Sort(*l.items)
}

func (l *txPrioritiedList) Cap(threshold float64) []*modules.TxPoolTransaction {
	save := make([]*modules.TxPoolTransaction, 0)
	drop := make([]*modules.TxPoolTransaction, 0)
	l.mu.Lock()
	defer l.mu.Unlock()
	for len(*l.items) > 0 {
		tx := l.items.Pop().(*modules.TxPoolTransaction)
		if _, has := (*l.all).Load(tx.Tx.Hash()); !has {
			l.stales--
			continue
		}
		priority, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
		if priority > threshold {
			save = append(save, tx)
			break
		}
		drop = append(drop, tx)
	}
	for _, tx := range save {
		l.items.Push(tx)
	}
	return drop
}

// Underpriced checks whether a transaction is cheaper than (or as cheap as) the
// lowest priced transaction currently being tracked.
func (l *txPrioritiedList) Underpriced(tx *modules.TxPoolTransaction) bool {
	all := l.All()
	if _, has := all[tx.Tx.Hash()]; has {
		return false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	// Discard stale price points if found at the heap start
	for len(*l.items) > 0 {
		head := []*modules.TxPoolTransaction(*l.items)[0]
		if _, ok := (*l.all).Load(head.Tx.Hash()); !ok {
			l.stales--
			l.items.Pop()
			continue
		}
		break
	}
	// Check if the transaction is underpriced or not
	if len(*l.items) == 0 {
		log.Error("Pricing query for empty pool") // This cannot happen, print to catch programming errors
		return false
	}
	cheapest := []*modules.TxPoolTransaction(*l.items)[0]
	cp, _ := strconv.ParseFloat(cheapest.Priority_lvl, 64)
	tp, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	return cp >= tp
}

// Discard finds a number of most underpriced transactions, removes them from the
// priced list and returns them for further removal from the entire pool.
func (l *txPrioritiedList) Discard(count int) modules.TxPoolTxs {
	drop := make(modules.TxPoolTxs, 0, count) // Remote underpriced transactions to drop
	all := l.All()
	l.mu.RLock()
	defer l.mu.RUnlock()
	for len(*l.items) > 0 && count > 0 {
		// Discard stale transactions if found during cleanup
		tx := l.items.Pop().(*modules.TxPoolTransaction)
		if _, ok := all[tx.Tx.Hash()]; !ok {
			l.stales--
			continue
		}
		drop = append(drop, tx)
		count--
	}
	return drop
}
