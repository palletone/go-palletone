package dag

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"testing"
)

func TestCreateUnit(t *testing.T) {
	path := "E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb"

	dagconfig.DbPath = path
	db, err := storage.Init(path, 16, 16)
	if err != nil {
		log.Error("Init db error", "error", err.Error())
		return
	}
	dag, err := NewDag(db)
	if err != nil {
		log.Error("New dag error", "error", err.Error())
		return
	}
	// new payload tpl payload
	tplPayload := modules.NewContractTplPayload([]byte("contract_template0000"),
		"TestContractTpl", "./contract", "1.1.1", 1024,
		[]byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159})
	// new msg
	msg := modules.NewMessage(modules.APP_CONTRACT_TPL, tplPayload)
	msgs := []*modules.Message{msg}
	// new transactions
	tx := modules.NewTransaction(msgs, 0)
	txs := modules.Transactions{tx}
	// new unit
	unit, err := dag.CreateUnitForTest(txs)
	if err != nil {
		log.Error("CreateUnitForTest error", "error", err.Error())
		return
	}
	// save unit
	if err := common.SaveUnit(dag.Db, *unit, false); err != nil {
		log.Error("Save unit error", "error", err.Error())
		return
	}
	log.Info("Save unit success")
}
