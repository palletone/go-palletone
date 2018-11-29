package dag

import (
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	dagcomm "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/txspool"
)

func TestCreateUnit(t *testing.T) {
	//path := "E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb"
	//
	//dagconfig.DbPath = path
	//db, err := storage.Init(path, 16, 16)
	db, err := ptndb.NewMemDatabase()
	if err != nil {
		log.Error("Init db error", "error", err.Error())
		return
	}
	unit, _ := createUnit()
	// save unit
	test_dag, err := NewDag4GenesisInit(db)
	if err != nil {
		log.Error("New dag error", "error", err.Error())
		return
	}

	txpool := txspool.NewTxPool(txspool.DefaultTxPoolConfig, test_dag, log.New("newgenesis"))
	if err := test_dag.SaveUnit(unit, txpool, true); err != nil {
		log.Error("Save unit error", "error", err.Error())
		return
	}
	// log.Info("Save unit success")
	genesis, err0 := test_dag.GetGenesisUnit(0)
	log.Info("get genesiss info", "error", err0, "info", genesis)
}
func createUnit() (*modules.Unit, error) {
	asset := new(modules.Asset)
	asset.AssetId = modules.PTNCOIN
	asset.UniqueId = modules.PTNCOIN
	// new payload tpl payload
	inputs := make([]*modules.Input, 0)
	in := new(modules.Input)
	in.Extra = []byte("jay")
	inputs = append(inputs, in)
	outputs := make([]*modules.Output, 0)
	out := new(modules.Output)
	out.Value = 1100000000
	out.Asset = asset
	outputs = append(outputs, out)
	payment := modules.NewPaymentPayload(inputs, outputs)
	msg0 := modules.NewMessage(modules.APP_PAYMENT, payment)
	tplPayload := modules.NewContractTplPayload([]byte("contract_template0000"),
		"TestContractTpl", "./contract", "1.1.1", 1024,
		[]byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159})
	// new msg
	msg := modules.NewMessage(modules.APP_CONTRACT_TPL, tplPayload)
	msgs := []*modules.Message{msg0}
	// new transactions
	tx := modules.NewTransaction(msgs[:])
	tx1 := modules.NewTransaction(append(msgs, msg))
	tx1 = tx1
	txs := modules.Transactions{tx}
	// new unit

	unit, err := dagcomm.NewGenesisUnit(txs, 1536451201, asset)
	log.Info("create unit success.", "error", err, "hash", unit.Hash().String())
	return unit, err
}
func TestDagRefreshUtxos(t *testing.T) {
	//db := storage.ReNewDbConn("/Users/jay/code/gocode/src/github.com/palletone/go-palletone/bin/work/palletone/gptn/leveldb/")
	db, _ := ptndb.NewMemDatabase()
	test_dag, _ := NewDag4GenesisInit(db)

	txpool := txspool.NewTxPool(txspool.DefaultTxPoolConfig, test_dag, log.New("newgenesis"))
	// txpool := txspool.NewTxPool4Test()
	dag_test, err := NewDagForTest(db, txpool)
	if err != nil {
		t.Fatal("New dag for test is faild,error: ", err)
	}
	unit, err := createUnit()
	if err != nil {
		log.Info("create test unit is failed.", "error", err)
		return
	}
	dag_test.SaveUnit(unit, txpool, true)
	// 添加delhash

	// unit := dag_test.GetCurrentUnit(modules.PTNCOIN)
	data := make(map[modules.OutPoint]*modules.Utxo)
	dag_test.utxos_cache[unit.Hash()] = data
	log.Debug("this unit hash info", "hash", unit.Hash().String())
	dag_test.Memdag.PushDelHashs([]common.Hash{unit.Hash()})
	log.Info("start refresh cache utxos.", "cache_len", len(dag_test.utxos_cache))

	dag_test.RefreshCacheUtxos()

	log.Info("stop refresh cache utxos.", "cache_len", len(dag_test.utxos_cache))

}
