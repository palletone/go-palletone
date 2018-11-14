package dag

import (
	"github.com/palletone/go-palletone/dag/common"
	"testing"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
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
	asset := new(modules.Asset)
	asset.AssetId = modules.PTNCOIN
	asset.UniqueId = modules.PTNCOIN
	asset.ChainId = 1
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

	unit, err := common.NewGenesisUnit(txs, 123, asset)
	log.Info("create unit success.", "hash", unit.Hash().String())
	// save unit
	test_dag, err := NewDag4GenesisInit(db)
	if err != nil {
		log.Error("New dag error", "error", err.Error())
		return
	}
	if err := test_dag.SaveUnit(unit, true); err != nil {
		log.Error("Save unit error", "error", err.Error())
		return
	}
	// log.Info("Save unit success")
	genesis, err0 := test_dag.GetGenesisUnit(0)
	log.Info("get genesiss info", "error", err0, "info", genesis)
}
