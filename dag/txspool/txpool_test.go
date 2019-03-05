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
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
)

var testTxPoolConfig TxPoolConfig

func init() {
	testTxPoolConfig = DefaultTxPoolConfig
	testTxPoolConfig.Journal = "test_transactions.rlp"
}

type UnitDag4Test struct {
	Db            *palletdb.MemDatabase
	utxodb        storage.IUtxoDb
	mux           sync.RWMutex
	GenesisUnit   *modules.Unit
	gasLimit      uint64
	chainHeadFeed *event.Feed
	outpoints     map[string]map[modules.OutPoint]*modules.Utxo
}

// NewTxPool4Test return TxPool structure for testing.
func NewTxPool4Test() *TxPool {
	//l := log.NewTestLog()
	testDag := NewUnitDag4Test()
	return NewTxPool(DefaultTxPoolConfig, testDag)
}

func NewUnitDag4Test() *UnitDag4Test {
	db, _ := palletdb.NewMemDatabase()
	utxodb := storage.NewUtxoDb(db)
	//idagdb := storage.NewDagDb(db)

	propdb := storage.NewPropertyDb(db)
	hash := common.HexToHash("0x0e7e7e3bd7c1e9ce440089712d61de38f925eb039f152ae03c6688ed714af729")
	idx := &modules.ChainIndex{AssetID: modules.PTNCOIN, Index: 0}
	h := modules.NewHeader([]common.Hash{hash}, uint64(1), []byte("hello"))
	h.Number = idx
	propdb.SetNewestUnit(h)
	//idagdb.PutHeadUnitHash()
	mutex := new(sync.RWMutex)

	ud := &UnitDag4Test{db, utxodb, *mutex, nil, 10000, new(event.Feed), nil}
	ud.outpoints = make(map[string]map[modules.OutPoint]*modules.Utxo)
	return ud
}
func (ud *UnitDag4Test) CurrentUnit() *modules.Unit {
	return modules.NewUnit(&modules.Header{
		Extra: []byte("test pool"),
	}, nil)
}

func (ud *UnitDag4Test) GetUnitByHash(hash common.Hash) (*modules.Unit, error) {
	return ud.CurrentUnit(), nil
}

func (ud *UnitDag4Test) StateAt(common.Hash) (*palletdb.MemDatabase, error) {
	return ud.Db, nil
}
func (ud *UnitDag4Test) GetHeaderByHash(common.Hash) (*modules.Header, error) {
	return nil, nil
}
func (ud *UnitDag4Test) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	if ud.outpoints == nil {
		return nil, fmt.Errorf("outpoints is nil ")
	}
	for _, utxos := range ud.outpoints {
		if utxos != nil {
			if u, has := utxos[*outpoint]; has {
				return u, nil
			}
		}
	}
	return nil, fmt.Errorf("not found!")
}

func (ud *UnitDag4Test) GetUtxoView(tx *modules.Transaction) (*UtxoViewpoint, error) {
	neededSet := make(map[modules.OutPoint]struct{})
	preout := modules.OutPoint{TxHash: tx.Hash()}
	for i, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				msgIdx := uint32(i)
				preout.MessageIndex = msgIdx
				for j := range msg.Outputs {
					txoutIdx := uint32(j)
					preout.OutIndex = txoutIdx
					neededSet[preout] = struct{}{}
				}
			}
		}

	}
	view := NewUtxoViewpoint()
	ud.addUtxoview(view, tx)
	ud.mux.RLock()
	err := view.FetchUtxos(ud.utxodb, neededSet)
	ud.mux.RUnlock()
	return view, err
}

func (ud *UnitDag4Test) addUtxoview(view *UtxoViewpoint, tx *modules.Transaction) {
	ud.mux.Lock()
	view.AddTxOuts(tx)
	ud.mux.Unlock()
}
func (ud *UnitDag4Test) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return ud.chainHeadFeed.Subscribe(ch)
}
func (ud *UnitDag4Test) GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error) {
	return &modules.AmountAsset{}, nil
}

func (ud *UnitDag4Test) GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error) {

	return nil, nil
}
func (ud *UnitDag4Test) GetTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	return nil, nil
}

// Tests that if the transaction count belonging to multiple accounts go above
// some hard threshold, if they are under the minimum guaranteed slot count then
// the transactions are still kept.
func TestTransactionAddingTxs(t *testing.T) {
	t0 := time.Now()
	fmt.Println("TestTransactionAddingTxs start.... ", t0)
	t.Parallel()

	// Create the pool to test the limit enforcement with
	db, _ := palletdb.NewMemDatabase()
	//l := log.NewTestLog()
	utxodb := storage.NewUtxoDb(db)
	mutex := new(sync.RWMutex)
	unitchain := &UnitDag4Test{db, utxodb, *mutex, nil, 10000, new(event.Feed), nil}
	config := testTxPoolConfig
	config.GlobalSlots = 4096
	var pending_cache, queue_cache, all, origin int
	pool := NewTxPool(config, unitchain)

	defer pool.Stop()

	txs := modules.Transactions{}

	msgs := make([]*modules.Message, 0)
	msgs1 := make([]*modules.Message, 0)
	addr := new(common.Address)
	addr.SetString("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	script := tokenengine.GenerateP2PKHLockScript(addr.Bytes())
	// step. compute total income

	totalIncome := uint64(100000000)
	// step2. create payload
	createT := big.Int{}

	outpoint := modules.OutPoint{
		// TxHash: ,
		MessageIndex: 2,
		OutIndex:     3,
	}
	input := modules.Input{
		PreviousOutPoint: &outpoint,
		SignatureScript:  []byte("xxxxxxxxxx"),
		Extra:            createT.SetInt64(time.Now().Unix()).Bytes(),
	}
	time.Sleep(1 * time.Second)
	input1 := modules.Input{
		PreviousOutPoint: &outpoint,
		SignatureScript:  []byte("cccccccccc"),
		Extra:            createT.SetInt64(time.Now().Unix()).Bytes(),
	}
	time.Sleep(1 * time.Second)
	input2 := modules.Input{
		PreviousOutPoint: &outpoint,
		SignatureScript:  []byte("vvvvvvvvvv"),
		Extra:            createT.SetInt64(time.Now().Unix()).Bytes(),
	}
	output := modules.Output{
		Value: totalIncome,
		Asset: &modules.Asset{
			AssetId: modules.BTCCOIN,
		},
		PkScript: script,
	}

	payload0 := &modules.PaymentPayload{
		Inputs:  []*modules.Input{&input},
		Outputs: []*modules.Output{&output},
	}
	payload1 := &modules.PaymentPayload{
		Inputs:  []*modules.Input{&input1},
		Outputs: []*modules.Output{&output},
	}
	payload2 := &modules.PaymentPayload{
		Inputs:  []*modules.Input{&input2},
		Outputs: []*modules.Output{&output},
	}
	for i := 0; i < 16; i++ {
		if i == 0 {
			msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, payload0))
		}
		if i == 1 {
			msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, payload1))
		}
		if i == 2 {
			msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, payload2))
		}
		msgs = append(msgs, modules.NewMessage(modules.APP_DATA, &modules.DataPayload{MainData: []byte(fmt.Sprintf("text%d%v", i, time.Now()))}))
	}

	for j := 0; j < 16; j++ {
		txs = append(txs, transaction(append(msgs1, msgs[j])))
	}
	fmt.Println("range txs start.... ", time.Now().Unix()-t0.Unix())
	// Import the batch and verify that limits have been enforced
	//	pool.AddRemotes(txs)
	for i, tx := range txs {
		if txs[i].Size() > 0 {
			continue
		} else {
			log.Debug("bad tx:", tx.Hash().String(), tx.Size())
		}
	}

	origin = len(txs)
	txpool_txs := make([]*modules.TxPoolTransaction, 0)
	pool_tx := new(modules.TxPoolTransaction)

	for i, tx := range txs {
		p_tx := TxtoTxpoolTx(pool, tx)
		p_tx.GetTxFee()
		p_tx.TxFee = &modules.AmountAsset{Amount: 20, Asset: tx.Asset()}
		txpool_txs = append(txpool_txs, p_tx)
		if i == len(txs)-1 {
			pool_tx = p_tx
		}
	}

	t1 := time.Now()
	fmt.Println("addlocals start.... ", t1)
	pool.AddLocals(txpool_txs)
	pendingTxs, _ := pool.pending()
	pending := 0
	p_txs := make([]*modules.TxPoolTransaction, 0)
	for _, txs := range pendingTxs {
		for _, tx := range txs {
			pending++
			p_txs = append(p_txs, tx)
		}
	}
	log.Debugf("pending:%d", pending)
	fmt.Println("addlocals over.... ", time.Now().Unix()-t0.Unix())
	for hash, list := range pendingTxs {
		if len(list) != 16 {
			t.Errorf("addr %x: total pending transactions mismatch: have %d, want %d", hash.String(), len(list), 16)
		} else {
			log.Debug("account matched.", "pending addr:", addr.String(), "amont:", len(list))
		}
	}
	fmt.Println("defer start.... ", time.Now().Unix()-t0.Unix())
	//  test GetSortedTxs{}
	unit_hash := common.HexToHash("0x0e7e7e3bd7c1e9ce440089712d61de38f925eb039f152ae03c6688ed714af729")
	defer func(p *TxPool) {
		if txs, total := p.GetSortedTxs(unit_hash); total.Float64() > dagconfig.DefaultConfig.UnitTxSize {
			all = len(txs)
			msg := fmt.Sprintf("total %v:total sizeof transactions is unexpected", total.Float64())
			t.Error(msg)
		} else {
			log.Debugf(" total size is :%v ,the cout:%d ", total, len(txs))
			for i, tx := range txs {
				if i < len(txs)-1 {
					if txs[i].Priority_lvl < txs[i+1].Priority_lvl {
						t.Error("sorted failed.", i, tx.Priority_lvl)
					}
				}
			}
			all = len(txs)
			for _, tx := range p.all {
				if tx.Pending {
					pending_cache++
				} else {
					queue_cache++
				}
			}
		}

		//  add tx : failed , and discared the tx.
		err := p.addTx(pool_tx, !pool.config.NoLocals)
		if err == nil {
			log.Error("test added tx failed.")
			return
		}
		err1 := p.resetPendingTx(pool_tx.Tx)
		if err1 != nil {
			log.Debug("resetPendingTx failed ", "error", err1)
		}
		err2 := p.addTx(pool_tx, !pool.config.NoLocals)
		if err2 != nil {
			log.Debug("addtx again info success", "error", err2)
		} else {
			log.Error("test added tx failed.")
		}
		log.Debugf("data:%d,%d,%d,%d,%d", origin, all, len(pool.all), pending_cache, queue_cache)
		fmt.Println("defer over.... spending time：", time.Now().Unix()-t0.Unix())
	}(pool)
}
func transaction(msg []*modules.Message) *modules.Transaction {
	return pricedTransaction(msg)
}
func pricedTransaction(msgs []*modules.Message) *modules.Transaction {
	tx := modules.NewTransaction(msgs)
	//tx.SetHash(rlp.RlpHash(tx))
	return tx
}

func TestUtxoViewPoint(t *testing.T) {
	view := NewUtxoViewpoint()
	outpoint := new(modules.OutPoint)
	utxo := new(modules.Utxo)
	outpoint.MessageIndex = 1
	outpoint.OutIndex = 2
	view.entries[*outpoint] = utxo
	utxo.Amount = 9999
	utxo.Spend()
	fmt.Println("enteris modified", outpoint, view.entries[*outpoint])
	if view.entries[*outpoint].Amount != 9999 {
		t.Error("failed", view.entries)
	}
	delete(view.entries, *outpoint)
}

func TestGetProscerTx(t *testing.T) {
	us := make([]*user, 0)
	var list []int
	us = append(us, &user{append(list, 1, 2), 2, append(list, 3, 4)}, &user{append(list, 3), 1, append(list, 5)}, &user{append(list, 4), 0, append(list, 7)}, &user{append(list, 7), 4, append(list, 8)}, &user{append(list, 8), 5, append(list, 9)}, &user{append(list, 0), 6, append(list, 1, 2)})

	l := getProscerTx(&user{append(list, 3), 1, append(list, 5)}, us)
	// 去重
	for i := 0; i < len(l)-1; i++ {
		for j := i + 1; j < len(l); j++ {
			if l[i] == l[j] {
				fmt.Println("重复", j, l)
				item := l[:i]
				item = append(item, l[j:]...)
				l = make([]int, 0)
				l = item[:]
				fmt.Println("重复", j, l)
			}
		}
	}
	if len(l) < 1 {
		fmt.Println("failed.", l)
	} else {
		for _, n := range l {
			fmt.Println("number:", n)
		}
	}
}

type user struct {
	inputs  []int
	u       int
	outputs []int
}

func getProscerTx(this *user, us []*user) []int {
	list := make([]int, 0)
	if len(us) > 0 {
		for _, num := range this.inputs {
			for _, u := range us {
				for _, out := range u.outputs {
					if out == num {
						list = append(list, u.u)
						fmt.Println("原始的num:", u.u)
						for _, next := range us {
							if next.u == u.u {
								if l := getProscerTx(next, us); len(l) > 0 {
									list = append(list, l...)
									fmt.Println("递归的num", l)
								}
							}
						}
					}
				}
			}
		}
	}

	return list
}
