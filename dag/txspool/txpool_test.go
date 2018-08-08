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
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/event"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
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
	t.Parallel()

	// Create the pool to test the limit enforcement with
	//db, _ := palletdb.NewMemDatabase()
	//unitchain := &testUnitDag{db, nil, 10000, new(event.Feed)}
	unitchain := dagcommon.NewDag()

	config := testTxPoolConfig
	config.GlobalSlots = 4096
	var queue_cache, queue_item, pending_cache, pending_item, all, origin int
	//pool := NewTxPool(config, unitchain)
	unitchain = unitchain //would recover
	pool := NewTxPool(config, unitchain)

	defer pool.Stop()

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 5)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey()
		// add address balance  and save db.
	}
	// Generate and queue a batch of transactions
	nonces := make(map[common.Address]uint64)

	txs := modules.Transactions{}
	for i, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		for j := 0; j < int(config.AccountSlots)*1; j++ {
			txs = append(txs, transaction(nonces[addr], uint64(i)+100, key))
			nonces[addr]++
		}
	}
	log.Println("all addr:", len(nonces))
	// Import the batch and verify that limits have been enforced
	//pool.AddRemotes(txs)
	for i, tx := range txs {
		if txs[i].Txsize > 0 {
			continue
		} else {
			log.Println("bad tx:", tx.Hash().String(), tx.Txsize)
		}
	}
	origin = len(txs)
	pool.AddLocals(txs)

	log.Println("pending:", len(pool.pending))
	log.Println("queue:", len(pool.queue))
	for addr, list := range pool.queue {
		if list.Len() != int(config.AccountSlots) {
			//t.Errorf("addr %x: total queue transactions mismatch: have %d, want %d", addr, list.Len(), config.AccountSlots)
		}
		// Println queue list.
		queue_cache = len(list.txs.cache)
		queue_item = len(list.txs.items)
		log.Println("queue addr: ", addr.String(), "amont:", list.Len())

	}
	for addr, list := range pool.pending {
		if list.Len() != int(config.AccountSlots) {
			t.Errorf("addr %x: total pending transactions mismatch: have %d, want %d", addr, list.Len(), config.AccountSlots)
		} else {
			log.Println("account matched.", "pending addr:", addr.String(), "amont:", list.Len())
		}
		pending_cache = len(list.txs.cache)
		pending_item = len(list.txs.items)
	}
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
					if txs[i].PriorityLvl() < txs[i+1].PriorityLvl() {
						t.Error("sorted failed.", i, tx.PriorityLvl())
					}
				}

			}
			all = len(txs)
			for key := range nonces {
				log.Println("address: ", key.String())
			}
		}
		log.Println(origin, all, queue_cache, queue_item, pending_cache, pending_item, len(nonces))
	}(pool)

}
func transaction(nonce uint64, txfee uint64, key *ecdsa.PrivateKey) *modules.Transaction {
	return pricedTransaction(nonce, new(big.Int).SetUint64(txfee), key)
}
func pricedTransaction(nonce uint64, txfee *big.Int, key *ecdsa.PrivateKey) *modules.Transaction {

	sig := make([]byte, 65)

	tx := modules.NewTransaction(nonce, txfee, sig)
	h := tx.TxHash

	sig, err := crypto.Sign(h[:], key)
	if err != nil {
		return tx
	}

	tx.From.R = sig[:32]
	tx.From.S = sig[32:64]
	tx.From.V = sig[64:]
	tx.From.Address = dagcommon.RSVtoAddress(tx).String()
	return tx
}
