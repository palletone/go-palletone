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
	return ap < bp
}
func (h priorityHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *priorityHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*modules.TxPoolTransaction)
	item.Index = n
	*h = append(*h, item)
	sort.Sort(h)
}

// -1 标识该数据已经出了优先级队列了 ,弹出优先级最高的
func (h *priorityHeap) Pop() interface{} {
	old := *h
	n := len(old)
	if n <= 0 {
		return nil
	}
	x := old[n-1]
	*h = old[0 : n-1]
	x.Index = -1
	return x
}
func (h *priorityHeap) LastPop() interface{} {
	old := *h
	n := len(old)
	if n <= 0 {
		return nil
	}
	x := old[0]
	*h = old[1:]
	x.Index = -1
	return x
}

func (h *priorityHeap) Update(item *modules.TxPoolTransaction, priority float64) {
	lvl := strconv.FormatFloat(priority, 'E', -1, 64)
	item.Priority_lvl = lvl
	for i, tx := range *h {
		if tx.Tx.Hash() == item.Tx.Hash() {
			(*h)[i] = item
		}
	}
	sort.Sort(h)
}

// txPricedList is a price-sorted heap to allow operating on transactions pool
// contents in a price-incrementing way.
type txPricedList struct {
	all    *map[common.Hash]*modules.TxPoolTransaction // Pointer to the map of all transactions
	items  *priorityHeap                               // Heap of prices of all the stored transactions
	stales int                                         // Number of stale price points to (re-heap trigger)
	sync.RWMutex
}

// newTxPricedList creates a new price-sorted transaction heap.
func newTxPricedList(all *map[common.Hash]*modules.TxPoolTransaction) *txPricedList {
	return &txPricedList{
		all:   all,
		items: new(priorityHeap),
	}
}

// Put inserts a new transaction into the heap.
func (l *txPricedList) Put(tx *modules.TxPoolTransaction) *priorityHeap {
	l.Lock()
	defer l.Unlock()
	l.items.Push(tx)
	//(*l.all)[tx.Tx.Hash()] = tx
	return l.items
}
func (l *txPricedList) Get() *modules.TxPoolTransaction {
	l.RLock()
	defer l.RUnlock()
	if l != nil {
		if l.items.Len() > 0 {
			tx, ok := l.items.Pop().(*modules.TxPoolTransaction)
			if ok {
				if tx.Tx != nil {
					return tx
				} else {
					log.Info("get items tx failed.")
				}
			}
		}
	}
	return nil
}

// Removed notifies the prices transaction list that an old transaction dropped
// from the pool. The list will just keep a counter of stale objects and update
// the heap if a large enough ratio of transactions go stale.
func (l *txPricedList) Removed(hash common.Hash) {
	l.Lock()
	defer l.Unlock()
	// Seems we've reached a critical number of stale transactions, reheap
	reheap := make(priorityHeap, 0)
	var exist bool
	for i, tx := range *l.items {
		if hash == tx.Tx.Hash() {
			exist = true
			if length := l.items.Len(); length > 1 {
				reheap = append(reheap, (*l.items)[0:i]...)
				reheap = append(reheap, (*l.items)[i+1:]...)
				//(*l.items)[i], (*l.items)[length-1] = (*l.items)[length-1], (*l.items)[i]
				//(*l.items) = (*l.items)[0 : length-1]
				l.stales--
			}
		}
	}
	if exist {
		l.items = &reheap
	}
}

// Underpriced checks whether a transaction is cheaper than (or as cheap as) the
// lowest priced transaction currently being tracked.
func (l *txPricedList) Underpriced(tx *modules.TxPoolTransaction) bool {
	l.RLock()
	defer l.RUnlock()

	// Discard stale price points if found at the heap start
	for len(*l.items) > 0 {
		head := []*modules.TxPoolTransaction(*l.items)[0]
		if _, ok := (*l.all)[head.Tx.Hash()]; !ok {
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
func (l *txPricedList) Discard(count int) modules.TxPoolTxs {
	drop := make(modules.TxPoolTxs, 0, count) // Remote underpriced transactions to drop
	l.Lock()
	defer l.Unlock()

	for len(*l.items) > 0 && count > 0 {
		// Discard stale transactions if found during cleanup
		tx := l.items.LastPop().(*modules.TxPoolTransaction)
		if _, ok := (*l.all)[tx.Tx.Hash()]; !ok {
			l.stales--
			continue
		}
		drop = append(drop, tx)
		if len(drop) >= count {
			break
		}
	}
	return drop
}
