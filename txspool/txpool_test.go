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
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/validator"
	"github.com/stretchr/testify/assert"
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
//func NewTxPool4Test() *TxPool {
//	//l := log.NewTestLog()
//	testDag := NewUnitDag4Test()
//	//validat:=&validator.ValidatorAllPass{}
//	return NewTxPool(testTxPoolConfig,
//		freecache.NewCache(1*1024*1024),
//		testDag, tokenengine.Instance)
//}

func NewUnitDag4Test() *UnitDag4Test {
	db, _ := palletdb.NewMemDatabase()
	utxodb := storage.NewUtxoDb(db, tokenengine.Instance, false)

	propdb := storage.NewPropertyDb(db)
	hash := common.HexToHash("0x0e7e7e3bd7c1e9ce440089712d61de38f925eb039f152ae03c6688ed714af729")
	b := []byte("hello")
	h := modules.NewHeader([]common.Hash{hash}, common.Hash{}, b, b, b, b, []uint16{},
		modules.PTNCOIN, 0, int64(1598766666))
	propdb.SetNewestUnit(h)
	mutex := new(sync.RWMutex)

	ud := &UnitDag4Test{db, utxodb, *mutex, nil, 10000, new(event.Feed), make(map[string]map[modules.OutPoint]*modules.Utxo)}
	return ud
}

func (ud *UnitDag4Test) CurrentUnit(token modules.AssetId) *modules.Unit {
	b := []byte("test pool")
	h := modules.NewHeader([]common.Hash{}, common.Hash{}, b, b, b, b, []uint16{},
		token, 0, int64(1598766666))
	return modules.NewUnit(h, nil)
}

func (ud *UnitDag4Test) GetUnitByHash(hash common.Hash) (*modules.Unit, error) {
	return ud.CurrentUnit(modules.PTNCOIN), nil
}
func (q *UnitDag4Test) GetMediators() map[common.Address]bool {
	return nil
}
func (q *UnitDag4Test) GetDb() ptndb.Database {
	return q.Db
}

func (q *UnitDag4Test) GetJurorReward(jurorAdd common.Address) common.Address {
	return jurorAdd
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

func (q *UnitDag4Test) GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
	return common.Hash{}, &modules.ChainIndex{}, nil
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

func (ud *UnitDag4Test) GetJurorByAddrHash(addrHash common.Hash) (*modules.JurorDeposit, error) {
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
	if outpoint.TxHash == common.BytesToHash([]byte("0")) {
		t := time.Now().AddDate(0, 0, -1).Unix()
		return &modules.Utxo{Amount: Ptn2Dao(10), Timestamp: uint64(t), Asset: modules.NewPTNAsset()}, nil
	}
	return nil, fmt.Errorf("not found!")
}
func (ud *UnitDag4Test) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	return nil, nil
}
func (ud *UnitDag4Test) GetTxOutput(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	return ud.GetUtxoEntry(outpoint)
}
func (ud *UnitDag4Test) GetUtxoView(tx *modules.Transaction) (*UtxoViewpoint, error) {
	neededSet := make(map[modules.OutPoint]struct{})
	preout := modules.OutPoint{TxHash: tx.Hash()}
	for i, msgcopy := range tx.TxMessages() {
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

func (ud *UnitDag4Test) GetTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	return nil, nil
}

func (ud *UnitDag4Test) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return common.Hash{}, fmt.Errorf("the txhash not found.")
}

func (ud *UnitDag4Test) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return nil, nil
}

func (ud *UnitDag4Test) GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error) {
	return []common.Address{}, nil, nil
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
func (ud *UnitDag4Test) CheckReadSetValid(contractId []byte, readSet []modules.ContractReadSet) bool {
	return true
}
func (ud *UnitDag4Test) GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	return common.Address{}, nil
}
func (ud *UnitDag4Test) IsContractDeveloper(addr common.Address) bool {
	return true
}

// create txs
func createTxs() []*modules.Transaction {
	txs := make([]*modules.Transaction, 0)
	hash0 := common.BytesToHash([]byte("0"))
	hash1 := common.BytesToHash([]byte("1"))
	txA := newTestPaymentTx(hash0)
	txB := newTestPaymentTx(txA.Hash())
	txC := newTestPaymentTx(txB.Hash())
	txD := newTestPaymentTx(txC.Hash())
	txX := newTestPaymentTx(hash1)
	txY := newTestPaymentTx(txX.Hash())
	txs = append(txs, txA, txB, txC, txD, txX, txY)

	return txs
}
func mockPtnUtxos() map[*modules.OutPoint]*modules.Utxo {
	result := map[*modules.OutPoint]*modules.Utxo{}
	p1 := modules.NewOutPoint(common.NewSelfHash(), 0, 0)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	utxo1 := &modules.Utxo{Asset: asset1, Amount: 100, LockTime: 0}
	utxo2 := &modules.Utxo{Asset: asset1, Amount: 200, LockTime: 0}

	result[p1] = utxo1
	p2 := modules.NewOutPoint(common.NewSelfHash(), 1, 0)
	result[p2] = utxo2

	p3 := modules.NewOutPoint(common.BytesToHash([]byte("0")), 0, 0)
	t := time.Now().AddDate(0, 0, -1).Unix()
	utxo3 := &modules.Utxo{Amount: Ptn2Dao(10), Timestamp: uint64(t), Asset: modules.NewPTNAsset()}
	result[p3] = utxo3
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
	utxodb := storage.NewUtxoDb(db, tokenengine.Instance, false)
	mutex := new(sync.RWMutex)
	unitchain := &UnitDag4Test{db, utxodb, *mutex, nil, 10000, new(event.Feed), nil}
	config := DefaultTxPoolConfig
	config.GlobalSlots = 4096
	config.NoLocals = true

	utxos := mockPtnUtxos()
	for outpoint, utxo := range utxos {
		utxodb.SaveUtxoEntity(outpoint, utxo)
	}
	pool := NewTxPool4DI(config, freecache.NewCache(1*1024*1024), unitchain,
		tokenengine.Instance, &validator.ValidatorAllPass{})
	defer pool.Stop()
	pool.startJournal(config)
	var pending_cache, queue_cache, all, origin int
	txs := createTxs()
	origin = len(txs)
	pool_tx := new(TxPoolTransaction)

	for i, tx := range txs {
		p_tx := TxtoTxpoolTx(tx)
		if i == len(txs)-1 {
			pool_tx = p_tx
		}
	}
	pool.AddLocals(txs)
	pendingTxs, _ := pool.pending()
	pending := 0
	p_txs := make([]*TxPoolTransaction, 0)
	for _, txs := range pendingTxs {
		for _, tx := range txs {
			pending++
			p_txs = append(p_txs, tx)
		}
	}
	assert.Equal(t, 0, 0)
	fmt.Println("addlocals over.... ", time.Now().Unix()-t0.Unix())
	//  test GetSortedTxs{}
	//unit_hash := common.HexToHash("0x0e7e7e3bd7c1e9ce440089712d61de38f925eb039f152ae03c6688ed714af729")
	defer func(p *TxPool) {
		sortedtxs, _ := p.GetSortedTxs()
		total := 0
		for _, tx := range sortedtxs {
			total += tx.Tx.SerializeSize()
		}
		log.Debugf(" total size is :%v ,the cout:%d ", total, len(txs))
		//for i, tx := range sortedtxs {
		//	if i < len(txs)-1 {
		//		if sortedtxs[i].Priority_lvl < sortedtxs[i+1].Priority_lvl {
		//			t.Error("sorted failed.", i, tx.Priority_lvl)
		//		}
		//	}
		//}
		//all = len(sortedtxs)

		all = len(sortedtxs)
		for i := 0; i < all-1; i++ {
			txpl := sortedtxs[i].Priority_lvl
			if txpl < sortedtxs[i+1].Priority_lvl {
				t.Error("sorted failed.", i, txpl)
			}
		}

		poolTxs := pool.AllTxpoolTxs()
		for _, tx := range poolTxs {
			if tx.Pending {
				pending_cache++ //6
			} else {
				queue_cache++ // 0
			}
		}
		//  add tx : failed , and discared the tx.
		err := p.addTx(pool_tx, !pool.config.NoLocals)
		assert.NotNil(t, err)
		err1 := p.DeleteTxByHash(pool_tx.Tx.Hash())
		if err1 != nil {
			log.Debug("DeleteTxByHash failed ", "error", err1)
		}
		err2 := p.addTx(pool_tx, !pool.config.NoLocals)
		if err2 == nil {
			log.Debug("addtx again info success")
		} else {
			log.Error("test added tx failed.", "error", err2)
		}
		log.Debugf("data:%d,%d,%d,%d,%d", origin, all, pool.AllLength(), pending_cache, queue_cache)
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

func TestPriorityHeap(t *testing.T) {
	txs := createTxs()
	p_txs := make([]*TxPoolTransaction, 0)
	list := new(priorityHeap)
	for _, tx := range txs {
		priority := rand.Float64()
		str := strconv.FormatFloat(priority, 'f', -1, 64)
		ptx := &TxPoolTransaction{Tx: tx, Priority_lvl: str}
		p_txs = append(p_txs, ptx)
		list.Push(ptx)
	}
	count := 0
	biger := new(TxPoolTransaction)
	bad := new(TxPoolTransaction)
	for list.Len() > 0 {
		inter := list.Pop()
		if inter != nil {
			ptx, ok := inter.(*TxPoolTransaction)
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
func TestGetProscerTx(t *testing.T) {
	db, _ := palletdb.NewMemDatabase()
	utxodb := storage.NewUtxoDb(db, tokenengine.Instance, false)
	mutex := new(sync.RWMutex)
	unitchain := &UnitDag4Test{db, utxodb, *mutex, nil, 10000, new(event.Feed), nil}
	config := DefaultTxPoolConfig
	config.GlobalSlots = 4096
	config.NoLocals = true

	utxos := mockPtnUtxos1()
	for outpoint, utxo := range utxos {
		utxodb.SaveUtxoEntity(outpoint, utxo)
	}
	pool := NewTxPool4DI(config, freecache.NewCache(1*1024*1024), unitchain,
		tokenengine.Instance, &validator.ValidatorAllPass{})
	defer pool.Stop()
	pool.startJournal(config)

	hash0 := common.BytesToHash([]byte("0"))
	txA := newTestPaymentTx(hash0)
	t.Logf("Tx A:%s", txA.Hash().String())
	txB := newCcInvokeRequest(txA.Hash())
	t.Logf("Tx B:%s", txB.Hash().String())
	txC := newCcInvokeFullTx(txB.Hash())
	t.Logf("Tx C:%s", txC.Hash().String())
	t.Logf("Tx C req:%s", txC.RequestHash().String())

	data, _ := json.Marshal(txC)
	t.Logf("Tx hash:%s,  C:%s, ", txC.Hash().String(), string(data))

	txD := newTestPaymentTx(txC.RequestHash()) //交易D是基于TxC在Request的时候的UTXO产生的
	t.Logf("Tx D:%s", txD.Hash().String())
	data1, _ := json.Marshal(txD)
	t.Logf("Tx hash:%s,  D:%s, ", txD.Hash().String(), string(data1))

	txs := make([]*modules.Transaction, 0)
	txs = append(txs, txD, txB, txA, txC)
	errs := pool.AddLocals(txs)
	for _, err := range errs {
		t.Logf("addLocals error:%s", err.Error())
	}
	defer func(p *TxPool) {

		count := p.Count()
		assert.Equal(t, 4, count)
		sortedTxs, _ := pool.GetSortedTxs()
		for index, tx := range sortedTxs {
			t.Logf("index:%d, hash:%s", index, tx.Tx.Hash().String())
		}
		if len(sortedTxs) == 4 {
			assert.Equal(t, txA.Hash().String(), sortedTxs[0].Tx.Hash().String())
			assert.Equal(t, txB.Hash().String(), sortedTxs[1].Tx.Hash().String())
			assert.Equal(t, txC.Hash().String(), sortedTxs[2].Tx.Hash().String())
			assert.Equal(t, txD.Hash().String(), sortedTxs[3].Tx.Hash().String())
		}
	}(pool)
}

func newTestPaymentTx(preTxHash common.Hash) *modules.Transaction {
	pay1s := &modules.PaymentPayload{
		LockTime: 0,
	}

	output := modules.NewTxOut(Ptn2Dao(10), []byte{0xee, 0xbb}, modules.NewPTNAsset())
	pay1s.AddTxOut(output)

	input := modules.Input{}
	input.PreviousOutPoint = modules.NewOutPoint(preTxHash, 0, 0)
	input.SignatureScript = []byte{}
	input.Extra = []byte("Test")

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}
	tx := modules.NewTransaction([]*modules.Message{msg})
	return tx
}
func Ptn2Dao(ptn uint64) uint64 {
	return ptn * 100000000
}

func newCcInvokeRequest(preTxHash common.Hash) *modules.Transaction {
	req := newTestPaymentTx(preTxHash)
	invoke := &modules.ContractInvokeRequestPayload{
		ContractId: []byte("PC1"),
		Args:       [][]byte{[]byte("put"), []byte("a")},
		Timeout:    0,
	}
	req.AddMessage(modules.NewMessage(modules.APP_CONTRACT_INVOKE_REQUEST, invoke))
	return req
}
func newCcInvokeFullTx(preTxHash common.Hash) *modules.Transaction {
	req := newCcInvokeRequest(preTxHash)
	result := &modules.ContractInvokePayload{
		ContractId: []byte("PC1"),
		Args:       [][]byte{[]byte("put"), []byte("a")},
		ReadSet:    nil,
		WriteSet:   nil,
		Payload:    []byte("ok"),
		ErrMsg:     modules.ContractError{},
	}
	req.AddMessage(modules.NewMessage(modules.APP_CONTRACT_INVOKE, result))
	return req
}
func mockPtnUtxos1() map[*modules.OutPoint]*modules.Utxo {
	result := map[*modules.OutPoint]*modules.Utxo{}
	p := modules.NewOutPoint(common.BytesToHash([]byte("0")), 0, 0)
	t := time.Now().AddDate(0, 0, -1).Unix()
	utxo := &modules.Utxo{Amount: Ptn2Dao(10), Timestamp: uint64(t), Asset: modules.NewPTNAsset()}
	result[p] = utxo
	return result
}
