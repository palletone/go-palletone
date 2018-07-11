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

	pool := NewTxPool(config, unitchain)
	defer pool.Stop()

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 5)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey()
		// add address balance  and save db.

		//pool.currentState.AddBalance(crypto.PubkeyToAddress(keys[i].PublicKey), big.NewInt(1000000))
	}
	// Generate and queue a batch of transactions
	nonces := make(map[common.Address]uint64)

	txs := modules.Transactions{}
	for i, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		for j := 0; j < int(config.AccountSlots)*1; j++ {
			txs = append(txs, transaction(nonces[addr], addr, uint64(i)+100, key))
			nonces[addr]++
		}
	}
	// Import the batch and verify that limits have been enforced
	//pool.AddRemotes(txs)
	for i, tx := range txs {
		txs[i].Txsize = tx.Size()
		if txs[i].Txsize > 0 {
			continue
		} else {
			log.Println("bad tx:", tx.Hash().String(), tx.Txsize)
		}
	}
	pool.AddLocals(txs)

	log.Println("pending:", len(pool.pending))
	log.Println("queue:", len(pool.queue))
	for addr, list := range pool.queue {
		if list.Len() != int(config.AccountSlots) {
			t.Errorf("addr %x: total pending transactions mismatch: have %d, want %d", addr, list.Len(), config.AccountSlots)
		} else {
			log.Println("account matched.", "queue addr:", addr.String(), "amont:", list.Len())
			for _, tx := range list.txs.cache {
				log.Println("cache tx:", tx.Txsize, tx.Hash().String(), tx.GetPriorityLvl())
			}
			for key, tx := range list.txs.items {
				log.Println("iteme tx:", key, tx.Txsize, tx.Hash().String(), tx.GetPriorityLvl())
			}
		}
	}
	for addr, list := range pool.pending {
		if list.Len() != int(config.AccountSlots) {
			t.Errorf("addr %x: total pending transactions mismatch: have %d, want %d", addr, list.Len(), config.AccountSlots)
		} else {
			log.Println("account matched.", "pending addr:", addr.String(), "amont:", list.Len())
		}
	}
}
func transaction(nonce uint64, addr common.Address, txfee uint64, key *ecdsa.PrivateKey) *modules.Transaction {
	return pricedTransaction(nonce, addr, new(big.Int).SetUint64(txfee), key)
}
func pricedTransaction(nonce uint64, addr common.Address, txfee *big.Int, key *ecdsa.PrivateKey) *modules.Transaction {
	tx := modules.NewTransaction(nonce, addr, txfee, nil)
	h := tx.Hash()
	// tx.TxFee = new(big.Int.SetUint64(txfee))
	sig, err := crypto.Sign(h[:], key)
	if err != nil {
		return tx
	}
	tx.From.TxAuthentifier.R = string(sig)
	return tx
}
