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
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

var testTxPoolConfig TxPoolConfig

func init() {
	testTxPoolConfig = DefaultTxPoolConfig
	testTxPoolConfig.Journal = "test_transactions.rlp"
}

type testUnitDag struct {
	Db *palletdb.MemDatabase

	GenesisUnit   *modules.Unit
	gasLimit      uint64
	chainHeadFeed *event.Feed
}

func (ud *testUnitDag) CurrentUnit() *modules.Unit {
	return modules.NewUnit(&modules.Header{
		Extra: []byte("test pool"),
	}, nil)
}

func (ud *testUnitDag) GetUnit(hash common.Hash) *modules.Unit {
	return ud.CurrentUnit()
}

func (ud *testUnitDag) StateAt(common.Hash) (*palletdb.MemDatabase, error) {
	return ud.Db, nil
}

func (ud *testUnitDag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return ud.chainHeadFeed.Subscribe(ch)
}

// Tests that if the transaction count belonging to multiple accounts go above
// some hard threshold, if they are under the minimum guaranteed slot count then
// the transactions are still kept.
func TestTransactionAddingTxs(t *testing.T) {
	t0 := time.Now()
	fmt.Println("script start.... ", t0)
	//t.Parallel()

	// Create the pool to test the limit enforcement with
	db, _ := palletdb.NewMemDatabase()
	unitchain := &testUnitDag{db, nil, 10000, new(event.Feed)}
	//unitchain := dag.NewDag()

	config := testTxPoolConfig
	config.GlobalSlots = 4096
	var pending_cache, queue_cache, all, origin int
	//pool := NewTxPool(config, unitchain)
	pool := NewTxPool(config, unitchain)

	defer pool.Stop()

	txs := modules.Transactions{}

	msgs := make([]modules.Message, 0)
	addr := new(common.Address)
	addr.SetString("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	script := tokenengine.GenerateP2PKHLockScript(addr.Bytes())
	// step. compute total income

	totalIncome := uint64(100000000)
	// step2. create payload
	createT := big.Int{}
	input := modules.Input{
		Extra: createT.SetInt64(time.Now().Unix()).Bytes(),
	}
	output := modules.Output{
		Value: totalIncome,
		Asset: modules.Asset{
			AssertId: modules.BTCCOIN,
			ChainId:  1},
		PkScript: script,
	}
	payload0 := modules.PaymentPayload{
		Inputs:  []modules.Input{input},
		Outputs: []modules.Output{output},
	}
	copy(input.Extra[:], []byte("1234567890"))
	payload1 := modules.PaymentPayload{
		Inputs:  []modules.Input{input},
		Outputs: []modules.Output{output},
	}
	copy(input.Extra[:], []byte("0987654321"))
	payload2 := modules.PaymentPayload{
		Inputs:  []modules.Input{input},
		Outputs: []modules.Output{output},
	}
	msgs = append(msgs, *modules.NewMessage(modules.APP_PAYMENT+"1", payload0), *modules.NewMessage(modules.APP_PAYMENT+"2", payload1), *modules.NewMessage(modules.APP_PAYMENT+"3", payload2))
	for j := 0; j < int(config.AccountSlots)*1; j++ {
		txs = append(txs, transaction(msgs, uint32(j+100)))
	}
	fmt.Println("range txs start.... ", time.Now().Unix()-t0.Unix())
	// Import the batch and verify that limits have been enforced
	//	pool.AddRemotes(txs)
	for i, tx := range txs {
		if txs[i].Size() > 0 {
			continue
		} else {
			log.Println("bad tx:", tx.Hash().String(), tx.Size())
		}
	}

	origin = len(txs)
	txpool_txs := make([]*modules.TxPoolTransaction, 0)
	for _, tx := range txs {
		txpool_txs = append(txpool_txs, TxtoTxpoolTx(pool, tx))
	}

	t1 := time.Now()
	fmt.Println("addlocals start.... ", (t1.Unix() - t0.Unix()))
	pool.AddLocals(txpool_txs)

	log.Println("pending:", len(pool.pending))
	fmt.Println("addlocals over.... ", time.Now().Unix()-t0.Unix())
	for hash := range pool.pending {
		if len(pool.pending) != int(config.AccountSlots) {
			t.Errorf("addr %x: total pending transactions mismatch: have %d, want %d", hash.String(), len(pool.pending), config.AccountSlots)
		} else {
			log.Println("account matched.", "pending addr:", addr.String(), "amont:", len(pool.pending))
		}
	}
	fmt.Println("defer start.... ", time.Now().Unix()-t0.Unix())
	//  test GetSortedTxs{}
	defer func(p *TxPool) {
		if txs, total := pool.GetSortedTxs(); total.Float64() > dagconfig.DefaultConfig.UnitTxSize {
			all = len(txs)
			msg := fmt.Sprintf("total %v:total sizeof transactions is unexpected", total.Float64())
			t.Error(msg)
		} else {
			log.Println(" total size is :", total, total.Float64(), "the cout: ", len(txs))
			for i, tx := range txs {
				if i < len(txs)-1 {
					if txs[i].Priority_lvl < txs[i+1].Priority_lvl {
						t.Error("sorted failed.", i, tx.Priority_lvl)
					}
				}

			}
			all = len(txs)
			pending_cache = len(pool.pending)
			queue_cache = len(pool.queue)
			//txsss := PoolTxstoTxs(txs)
			//fmt.Println("txssssss===========", txsss[0])
		}
		log.Println(origin, all, len(pool.all), pending_cache, queue_cache)
		fmt.Println("defer over.... ", time.Now().Unix()-t0.Unix())
	}(pool)

}
func transaction(msg []modules.Message, lock uint32) *modules.Transaction {
	return pricedTransaction(msg, lock)
}
func pricedTransaction(msg []modules.Message, lock uint32) *modules.Transaction {
	tx := modules.NewTransaction(msg, lock)
	tx.SetHash(rlp.RlpHash(tx))
	return tx
}
