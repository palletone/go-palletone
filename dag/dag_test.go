package dag

import (
	"testing"

	"crypto/ecdsa"
	"encoding/hex"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	dagcomm "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/stretchr/testify/assert"
)

func TestCreateUnit(t *testing.T) {

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

	//txpool := txspool.NewTxPool(txspool.DefaultTxPoolConfig, test_dag)
	if err := test_dag.SaveUnit(unit, nil, true); err != nil {
		log.Error("Save unit error", "error", err.Error())
		return
	}
	// log.Info("Save unit success")
	genesis, err0 := test_dag.GetGenesisUnit()
	log.Info("get genesiss info", "error", err0, "info", genesis)
}
func createUnit() (*modules.Unit, error) {
	asset := modules.NewPTNAsset()

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
	tplPayload := modules.NewContractTplPayload([]byte("contract_template0000"), 1024,
		[]byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159}, modules.ContractError{})
	// new msg
	msg := modules.NewMessage(modules.APP_CONTRACT_TPL, tplPayload)
	msgs := []*modules.Message{msg0}
	// new transactions
	tx := modules.NewTransaction(msgs[:])
	tx1 := modules.NewTransaction(append(msgs, msg))
	tx1 = tx1
	txs := modules.Transactions{tx}
	// new unit

	unit, err := dagcomm.NewGenesisUnit(txs, 1536451201, asset, -1, common.Hash{})
	log.Info("create unit success.", "error", err, "hash", unit.Hash().String())
	return unit, err
}

func TestTxCountAndUnitSize(t *testing.T) {
	sign, _ := hex.DecodeString("2c731f854ef544796b2e86c61b1a9881a0148da0c1001f0da5bd2074d2b8360367e2e0a57de91a5cfe92b79721692741f47588036cf0101f34dab1bfda0eb030")
	pubKey, _ := hex.DecodeString("0386df0aef707cc5bc8d115c2576f844d2734b05040ef2541e691763f802092c09")
	unlockScript := tokenengine.GenerateP2PKHUnlockScript(sign, pubKey)
	a := modules.NewPTNAsset()
	addr, _ := common.StringToAddress("P13pBrshF6JU7QhMmzJjXx3mWHh13YHAUAa")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	for i := 1; i < 100000; i *= 2 {
		txs := modules.Transactions{}
		for j := 0; j < i; j++ {
			tx := modules.NewTransaction([]*modules.Message{})
			tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, modules.NewPaymentPayload([]*modules.Input{modules.NewTxIn(modules.NewOutPoint(common.Hash{}, 0, 0), unlockScript)},
				[]*modules.Output{modules.NewTxOut(1, lockScript, a)})))
			txs = append(txs, tx)
		}
		unit := modules.NewUnit(newHeader(), txs)
		t.Logf("Tx count:%d,Unit size:%s", i, unit.Size().String())
	}
}
func newHeader() *modules.Header {
	key := new(ecdsa.PrivateKey)
	key, _ = crypto.GenerateKey()
	h := new(modules.Header)
	//h.AssetIDs = append(h.AssetIDs, modules.PTNCOIN)
	au := modules.Authentifier{}
	//address := crypto.PubkeyToAddress(&key.PublicKey)

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &modules.ChainIndex{}
	h.Number.AssetID = modules.PTNCOIN
	h.Number.Index = uint64(333333)
	h.Extra = make([]byte, 20)
	h.ParentsHash = append(h.ParentsHash, h.TxRoot)

	h.TxRoot = h.Hash()
	sig, _ := crypto.Sign(h.TxRoot[:], key)
	au.Signature = sig
	au.PubKey = crypto.CompressPubkey(&key.PublicKey)
	h.Authors = au
	return h
}
func TestDag_InsertHeaderDag(t *testing.T) {
	dag, err := setupDag()
	assert.NotNil(t, dag)
	assert.Nil(t, err)
	headers := []*modules.Header{newHeader()}

	dag.InsertHeaderDag(headers)
}
func setupDag() (*Dag, error) {
	db, err := ptndb.NewMemDatabase()
	if err != nil {
		log.Error("Init db error", "error", err.Error())
		return nil, err
	}
	unit, _ := createUnit()
	// save unit
	initDag, err := NewDag4GenesisInit(db)
	if err != nil {
		log.Error("New dag error", "error", err.Error())
		return nil, err
	}
	//txpool := txspool.NewTxPool(txspool.DefaultTxPoolConfig, test_dag)
	if err := initDag.SaveUnit(unit, nil, true); err != nil {
		log.Error("Save unit error", "error", err.Error())
		return nil, err
	}
	test_dag, err := NewDagForTest(db)
	return test_dag, err
}

//func TestDag_GetGenesisUnit(t *testing.T) {
//	db,_:=ptndb.NewLDBDatabase("./leveldb",0,128)
//	dag,_:=NewDag4GenesisInit(db)
//	txid:= common.HexToHash("0x10f0375ea48aa09099b0148d8a19fc3ac297a22b29bff0dfa99546f7af0fb57c")
//	tx,err:= dag.GetTransactionOnly(txid)
//	assert.Nil(t,err)
//	t.Log(tx.Hash().String())
//	for i,msg:=range tx.TxMessages{
//		data,_:= json.Marshal(msg.Payload)
//		t.Logf("Message[%d], APP:%v,%s",i,msg.App,string(data))
//	}
//	assert.Equal(t,txid,tx.Hash())
//}
