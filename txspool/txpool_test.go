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
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/parameter"
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
	return NewTxPool(testTxPoolConfig, freecache.NewCache(1*1024*1024), testDag, tokenengine.Instance)
}

func NewUnitDag4Test() *UnitDag4Test {
	db, _ := palletdb.NewMemDatabase()
	utxodb := storage.NewUtxoDb(db, tokenengine.Instance)

	propdb := storage.NewPropertyDb(db)
	hash := common.HexToHash("0x0e7e7e3bd7c1e9ce440089712d61de38f925eb039f152ae03c6688ed714af729")
	idx := &modules.ChainIndex{AssetID: modules.PTNCOIN, Index: 0}
	h := modules.NewHeader([]common.Hash{hash}, uint64(1), []byte("hello"))
	h.Number = idx
	propdb.SetNewestUnit(h)
	mutex := new(sync.RWMutex)

	ud := &UnitDag4Test{db, utxodb, *mutex, nil, 10000, new(event.Feed), make(map[string]map[modules.OutPoint]*modules.Utxo)}
	return ud
}

func (ud *UnitDag4Test) CurrentUnit(token modules.AssetId) *modules.Unit {
	return modules.NewUnit(&modules.Header{
		Number: &modules.ChainIndex{AssetID: token},
		Extra:  []byte("test pool"),
	}, nil)
}

func (ud *UnitDag4Test) GetUnitByHash(hash common.Hash) (*modules.Unit, error) {
	return ud.CurrentUnit(modules.PTNCOIN), nil
}

func (q *UnitDag4Test) GetMediators() map[common.Address]bool {
	return nil
}

func (q *UnitDag4Test) GetSlotAtTime(when time.Time) uint32 {
	return 0
}

func (q *UnitDag4Test) GetMediator(add common.Address) *core.Mediator {
	return nil
}

func (q *UnitDag4Test) GetScheduledMediator(slotNum uint32) common.Address {
	return common.Address{}
}

func (q *UnitDag4Test) GetNewestUnitTimestamp(token modules.AssetId) (int64, error) {
	return 0, nil
}

func (q *UnitDag4Test) GetChainParameters() *core.ChainParameters {
	return nil
}

func (ud *UnitDag4Test) StateAt(common.Hash) (*palletdb.MemDatabase, error) {
	return ud.Db, nil
}
func (ud *UnitDag4Test) GetHeaderByHash(common.Hash) (*modules.Header, error) {
	return nil, nil
}
func (ud *UnitDag4Test) IsTransactionExist(hash common.Hash) (bool, error) {
	return false, nil
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
func (ud *UnitDag4Test) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	return nil, nil
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

func (ud *UnitDag4Test) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return nil, nil
}
func (ud *UnitDag4Test) GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error){
	return []common.Address{},nil,nil
}

func (ud *UnitDag4Test) GetMinFee() (*modules.AmountAsset, error) {
	return nil, nil
}

func (ud *UnitDag4Test) GetContractJury(contractId []byte) (*modules.ElectionNode, error) {
	return nil, nil
}
func (ud *UnitDag4Test) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return nil, nil, nil
}
func (ud *UnitDag4Test) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return nil, nil
}

// create txs
func createTxs(address string) []*modules.Transaction {
	txs := make([]*modules.Transaction, 0)

	sign, _ := hex.DecodeString("2c731f854ef544796b2e86c61b1a9881a0148da0c1001f0da5bd2074d2b8360367e2e0a57de91a5cfe92b79721692741f47588036cf0101f34dab1bfda0eb030")
	pubKey, _ := hex.DecodeString("0386df0aef707cc5bc8d115c2576f844d2734b05040ef2541e691763f802092c09")
	unlockScript := tokenengine.GenerateP2PKHUnlockScript(sign, pubKey)
	a := modules.NewPTNAsset()
	addr, _ := common.StringToAddress(address)
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	for j := 0; j < 16; j++ {
		tx := modules.NewTransaction([]*modules.Message{})
		output := modules.NewTxOut(uint64(j+10), lockScript, a)
		tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, modules.NewPaymentPayload([]*modules.Input{modules.NewTxIn(modules.NewOutPoint(common.NewSelfHash(),
			0, 0), unlockScript)}, []*modules.Output{output})))
		txs = append(txs, tx)
	}
	return txs
}
func mockPtnUtxos() map[modules.OutPoint]*modules.Utxo {
	result := map[modules.OutPoint]*modules.Utxo{}
	p1 := modules.NewOutPoint(common.NewSelfHash(), 0, 0)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	utxo1 := &modules.Utxo{Asset: asset1, Amount: 100, LockTime: 0}
	utxo2 := &modules.Utxo{Asset: asset1, Amount: 200, LockTime: 0}

	result[*p1] = utxo1
	p2 := modules.NewOutPoint(common.NewSelfHash(), 1, 0)
	result[*p2] = utxo2
	return result
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
	utxodb := storage.NewUtxoDb(db, tokenengine.Instance)
	mutex := new(sync.RWMutex)
	unitchain := &UnitDag4Test{db, utxodb, *mutex, nil, 10000, new(event.Feed), nil}
	config := DefaultTxPoolConfig
	config.GlobalSlots = 4096

	utxos := mockPtnUtxos()
	for outpoint, utxo := range utxos {
		utxodb.SaveUtxoEntity(&outpoint, utxo)
	}

	pool := NewTxPool(config, freecache.NewCache(1*1024*1024), unitchain, tokenengine.Instance)
	defer pool.Stop()

	var pending_cache, queue_cache, all, origin int
	address := "P13pBrshF6JU7QhMmzJjXx3mWHh13YHAUAa"
	txs := createTxs(address)
	fmt.Println("range txs start...  , spent time:", time.Since(t0))
	// Import the batch and verify that limits have been enforced
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
		p_tx := TxtoTxpoolTx(tx)
		txpool_txs = append(txpool_txs, p_tx)
		if i == len(txs)-1 {
			pool_tx = p_tx
		}
	}

	t1 := time.Now()
	fmt.Println("addlocals start.... ", t1)
	pool.AddLocals(txs)
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
			log.Debug("account matched.", "pending addr:", address, "amont:", len(list))
		}
	}
	fmt.Println("defer start.... ", time.Now().Unix()-t0.Unix())
	//  test GetSortedTxs{}
	unit_hash := common.HexToHash("0x0e7e7e3bd7c1e9ce440089712d61de38f925eb039f152ae03c6688ed714af729")
	defer func(p *TxPool) {
		if txs, total := p.GetSortedTxs(unit_hash, 1); uint64(total.Float64()) > parameter.CurrentSysParameters.UnitMaxSize {
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
			poolTxs := pool.AllTxpoolTxs()
			for _, tx := range poolTxs {
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
		log.Debugf("data:%d,%d,%d,%d,%d", origin, all, pool.AllLength(), pending_cache, queue_cache)
		fmt.Println("defer over.... spending time：", time.Now().Unix()-t0.Unix())
	}(pool)
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
	fmt.Println("getProscer:", l)
	// 去重
	m := make(map[int]int)
	for _, u := range l {
		m[u] = u
	}
	l = make([]int, 0)
	for _, u := range m {
		l = append(l, u)
	}

	if len(l) < 1 {
		fmt.Println("failed.", l)
	} else {
		fmt.Println("rm repeat:", l)
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

func TestPriorityHeap(t *testing.T) {
	txs := createTxs("P13pBrshF6JU7QhMmzJjXx3mWHh13YHAUAa")
	p_txs := make([]*modules.TxPoolTransaction, 0)
	list := new(priorityHeap)
	for _, tx := range txs {
		priority := rand.Float64()
		str := strconv.FormatFloat(priority, 'f', -1, 64)
		ptx := &modules.TxPoolTransaction{Tx: tx, Priority_lvl: str}
		p_txs = append(p_txs, ptx)
		list.Push(ptx)
	}
	count := 0
	biger := new(modules.TxPoolTransaction)
	bad := new(modules.TxPoolTransaction)
	for list.Len() > 0 {
		inter := list.Pop()
		if inter != nil {
			ptx, ok := inter.(*modules.TxPoolTransaction)
			if ok {
				if count == 0 {
					biger.Priority_lvl = ptx.Priority_lvl
				}
				if count > 1 {
					bp, _ := strconv.ParseFloat(biger.Priority_lvl, 64)
					pp, _ := strconv.ParseFloat(ptx.Priority_lvl, 64)
					if bp < pp {
						biger = ptx
						t.Fatal(fmt.Sprintf("sort.Sort.priorityHeap is failed.biger:  %s ,ptx: %s  , index: %d ", biger.Priority_lvl, ptx.Priority_lvl, ptx.Index))
					} else {
						bad = ptx
					}
				}
				count++

			}
		} else {
			log.Debug("pop error: the interTx is nil ", "count", count)
			break
		}
	}
	log.Debug("all pop end. ", "count", count)
	log.Debug("best priority  tx: ", "info", biger)
	log.Debug("bad priority  tx: ", "info", bad)
}
