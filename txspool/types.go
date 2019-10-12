/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package txspool

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"math/big"
	"strconv"
	"sync"
	"time"
)

type TxPoolTransaction struct {
	Tx *modules.Transaction

	From         []*modules.OutPoint
	CreationDate time.Time           `json:"creation_date"`
	Priority_lvl string              `json:"priority_lvl"` // 打包的优先级
	UnitHash     common.Hash
	UnitIndex    uint64
	Pending      bool
	Confirmed    bool
	IsOrphan     bool
	Discarded    bool // will remove
	TxFee        []*modules.Addition `json:"tx_fee"`
	Index        uint64              `json:"index"` // index 是该Unit位置。
	Extra        []byte
	Tag          uint64
	Expiration   time.Time
	//该Tx依赖于哪些TxId作为先决条件
	DependOnTxs []common.Hash
}

func (tx *TxPoolTransaction) Less(otherTx interface{}) bool {
	ap, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	bp, _ := strconv.ParseFloat(otherTx.(*TxPoolTransaction).Priority_lvl, 64)
	return ap < bp
}

func (tx *TxPoolTransaction) GetPriorityLvl() string {
	if tx.Priority_lvl != "" && tx.Priority_lvl > "0" {
		return tx.Priority_lvl
	}
	var priority_lvl float64
	if txfee := tx.GetTxFee(); txfee.Int64() > 0 {
		if tx.CreationDate.Unix() <= 0 {
			tx.CreationDate = time.Now()
		}
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee.Int64())/
			tx.Tx.Size().Float64()* (1 + float64(time.Now().Second()-tx.CreationDate.Second())/(24*3600))), 64)
	}
	tx.Priority_lvl = strconv.FormatFloat(priority_lvl, 'f', -1, 64)
	return tx.Priority_lvl
}
func (tx *TxPoolTransaction) GetPriorityfloat64() float64 {
	level, _ := strconv.ParseFloat(tx.Priority_lvl, 64)
	if level > 0 {
		return level
	}
	var priority_lvl float64
	if txfee := tx.GetTxFee(); txfee.Int64() > 0 {
		if tx.CreationDate.Unix() <= 0 {
			tx.CreationDate = time.Now()
		}
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee.Int64())/
			tx.Tx.Size().Float64()* (1 + float64(time.Now().Second()-tx.CreationDate.Second())/(24*3600))), 64)
	}
	return priority_lvl
}
func (tx *TxPoolTransaction) SetPriorityLvl(priority float64) {
	tx.Priority_lvl = strconv.FormatFloat(priority, 'f', -1, 64)
}
func (tx *TxPoolTransaction) GetTxFee() *big.Int {
	var fee uint64
	if tx.TxFee != nil {
		for _, ad := range tx.TxFee {
			fee += ad.Amount
		}
	} else {
		fee = 20 // 20dao
	}
	return big.NewInt(int64(fee))
}

//type Transactions []*Transaction
type TxPoolTxs []*TxPoolTransaction

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice TxPoolTxs

func (s TxByPrice) Len() int      { return len(s) }
func (s TxByPrice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*TxPoolTransaction))
}
func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TxByPriority []*TxPoolTransaction

func (s TxByPriority) Len() int           { return len(s) }
func (s TxByPriority) Less(i, j int) bool { return s[i].Priority_lvl > s[j].Priority_lvl }
func (s TxByPriority) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPriority) Push(x interface{}) {
	*s = append(*s, x.(*TxPoolTransaction))
}

func (s *TxByPriority) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TxByCreationDate []*TxPoolTransaction

func (tc TxByCreationDate) Len() int           { return len(tc) }
func (tc TxByCreationDate) Less(i, j int) bool { return tc[i].Priority_lvl > tc[j].Priority_lvl }
func (tc TxByCreationDate) Swap(i, j int)      { tc[i], tc[j] = tc[j], tc[i] }

type SequeueTxPoolTxs struct {
	seqtxs []*TxPoolTransaction
	mu     sync.RWMutex
}

// add
func (seqTxs *SequeueTxPoolTxs) Len() int {
	seqTxs.mu.RLock()
	defer seqTxs.mu.RUnlock()
	return len((*seqTxs).seqtxs)
}
func (seqTxs *SequeueTxPoolTxs) Add(newPoolTx *TxPoolTransaction) {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	(*seqTxs).seqtxs = append((*seqTxs).seqtxs, newPoolTx)
}

// add priority
func (seqTxs *SequeueTxPoolTxs) AddPriority(newPoolTx *TxPoolTransaction) {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	if seqTxs.Len() == 0 {
		(*seqTxs).seqtxs = append((*seqTxs).seqtxs, newPoolTx)
	} else {
		added := false
		for i, item := range (*seqTxs).seqtxs {
			if newPoolTx.GetPriorityfloat64() > item.GetPriorityfloat64() {
				(*seqTxs).seqtxs = append((*seqTxs).seqtxs[:i], append([]*TxPoolTransaction{newPoolTx}, (*seqTxs).seqtxs[i:]...)...)
				added = true
				break
			}
		}
		if !added {
			(*seqTxs).seqtxs = append((*seqTxs).seqtxs, newPoolTx)
		}
	}
}

// get
func (seqTxs *SequeueTxPoolTxs) Get() *TxPoolTransaction {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	if seqTxs.Len() <= 0 {
		return nil
	}
	if seqTxs.Len() == 1 {
		first := (*seqTxs).seqtxs[0]
		(*seqTxs).seqtxs = make([]*TxPoolTransaction, 0)
		return first
	}
	first, rest := (*seqTxs).seqtxs[0], (*seqTxs).seqtxs[1:]
	(*seqTxs).seqtxs = rest
	return first
}

// get all
func (seqTxs *SequeueTxPoolTxs) All() []*TxPoolTransaction {
	seqTxs.mu.Lock()
	defer seqTxs.mu.Unlock()
	items := (*seqTxs).seqtxs[:]
	(*seqTxs).seqtxs = make([]*TxPoolTransaction, 0)
	return items
}
