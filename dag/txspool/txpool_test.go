package txspool

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/event"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
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
		GasLimit: ud.gasLimit,
	}, nil)
}

func (ud *testUnitDag) GetUnit(hash common.Hash, number uint64) *modules.Unit {
	return ud.CurrentUnit()
}

func (ud *testUnitDag) StateAt(common.Hash) (*palletdb.MemDatabase, error) {
	return ud.Db, nil
}

func (ud *testUnitDag) SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription {
	return ud.chainHeadFeed.Subscribe(ch)
}

// Tests that if the transaction count belonging to multiple accounts go above
// some hard threshold, if they are under the minimum guaranteed slot count then
// the transactions are still kept.
func TestTransactionAddingTxs(t *testing.T) {
	t.Parallel()

	// Create the pool to test the limit enforcement with
	db, _ := palletdb.NewMemDatabase()
	unitchain := &testUnitDag{db, nil, 10000, new(event.Feed)}

	config := testTxPoolConfig
	config.GlobalSlots = 0
	var queue_cache, queue_item, all, origin int
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
			t.Errorf("addr %x: total pending transactions mismatch: have %d, want %d", addr, list.Len(), config.AccountSlots)
		}
		// Println queue list.
		queue_cache = len(list.txs.cache)
		queue_item = len(list.txs.items)
		log.Println("queue addr: ", addr.String(), "amont:", list.Len())
		// for _, tx := range list.txs.cache {
		// 	log.Println("cache tx:", tx.Hash().String(), tx.Txsize, tx.PriorityLvl())
		// }
		// for key, tx := range list.txs.items {
		// 	log.Println("iteme tx:", key, tx.Hash().String(), tx.Txsize, tx.PriorityLvl())
		// }

	}
	for addr, list := range pool.pending {
		if list.Len() != int(config.AccountSlots) {
			t.Errorf("addr %x: total pending transactions mismatch: have %d, want %d", addr, list.Len(), config.AccountSlots)
		} else {
			log.Println("account matched.", "pending addr:", addr.String(), "amont:", list.Len())
		}
	}
	//  test GetSortedTxs{}
	defer func(p *TxPool) {
		txs := pool.GetSortedTxs()

		for i, tx := range txs {
			if i < len(txs)-1 {
				if txs[i].PriorityLvl() < txs[i+1].PriorityLvl() {
					t.Error("sorted failed.", i, tx.PriorityLvl())
				}
			}
		}
		all = len(txs)
		for key, _ := range nonces {
			log.Println("address: ", key.String())
		}
		log.Println(origin, all, queue_cache, queue_item, txs[10], len(nonces))
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
	return tx
}
