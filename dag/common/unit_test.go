package common

import (
	"log"
	"testing"
	"time"

	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

func TestNewGenesisUnit(t *testing.T) {
	gUnit, _ := NewGenesisUnit(modules.Transactions{}, time.Now().Unix())

	log.Println("Genesis unit struct:")
	log.Println("--- Genesis unit header --- ")
	log.Println("parent units:", gUnit.UnitHeader.ParentsHash)
	log.Println("asset ids:", gUnit.UnitHeader.AssetIDs)
	log.Println("witness:", gUnit.UnitHeader.Witness)
	log.Println("Root:", gUnit.UnitHeader.TxRoot)
	log.Println("Number:", gUnit.UnitHeader.Number)

}

func TestGenGenesisConfigPayload(t *testing.T) {
	var genesisConf core.Genesis
	genesisConf.SystemConfig.DepositRate = 0.02

	genesisConf.InitialParameters.MediatorInterval = 10

	payload, err := GenGenesisConfigPayload(&genesisConf, &modules.Asset{})

	if err != nil {
		log.Println(err)
	}

	for k, v := range payload.ConfigSet {
		log.Println(k, v)
	}
}

func TestSaveUnit(t *testing.T) {
	if storage.Dbconn == nil {
		log.Println("dbconn is nil , renew db  start ...")
		storage.Dbconn = storage.ReNewDbConn(dagconfig.DbPath)
	}

	addr := common.Address{}
	addr.SetString("P12EA8oRMJbAtKHbaXGy8MGgzM8AMPYxkN1")
	//ks := keystore.NewKeyStore("./keystore", 1<<18, 1)

	p := common.Hash{}
	p.SetString("0000000000000000000000000000000")
	aid := modules.IDType16{}
	aid.SetBytes([]byte("xxxxxxxxxxxxxxxxxx"))
	header := new(modules.Header)
	header.ParentsHash = append(header.ParentsHash, p)
	header.AssetIDs = []modules.IDType16{aid}
	key, _ := crypto.GenerateKey()
	addr0 := crypto.PubkeyToAddress(key.PublicKey)

	sig, err := crypto.Sign(header.Hash().Bytes(), key)
	if err != nil {
		log.Println("sign header occured error: ", err)
	}
	auth := new(modules.Authentifier)
	auth.R = sig[:32]
	auth.S = sig[32:64]
	auth.V = sig[64:]
	auth.Address = addr0.String()
	header.Authors = auth
	contractTplPayload := modules.ContractTplPayload{
		TemplateId: common.HexToHash("contract_template0000"),
		Bytecode:   []byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159},
		Name:       "TestContractTpl",
		Path:       "./contract",
	}
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Value: &modules.StateVersion{
		Height:  GenesisHeight(),
		TxIndex: 0,
	}})
	writeSet := []modules.PayloadMapStruct{
		{
			Key:   "name",
			Value: "Joe",
		},
		{
			Key:   "age",
			Value: 10,
		},
	}
	deployPayload := modules.ContractDeployPayload{
		TemplateId: common.HexToHash("contract_template0000"),
		ContractId: []byte("contract0000"),
		ReadSet:    readSet,
		WriteSet:   writeSet,
	}

	invokePayload := modules.ContractInvokePayload{
		ContractId: []byte("contract0000"),
		Args:       [][]byte{[]byte("initial")},
		ReadSet:    readSet,
		WriteSet: []modules.PayloadMapStruct{
			{
				Key:   "name",
				Value: "Alice",
			},
			{
				Key: "age",
				Value: modules.DelContractState{
					IsDelete: true,
				},
			},
		},
	}
	tx1 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_CONTRACT_TPL,
				Payload: contractTplPayload,
			},
		},
	}
	tx1.CreationDate = tx1.CreateDate()
	tx1.TxHash = tx1.Hash()

	tx2 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_CONTRACT_DEPLOY,
				Payload: deployPayload,
			},
		},
	}
	tx2.CreationDate = tx2.CreateDate()
	tx2.TxHash = tx2.Hash()

	tx3 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_CONTRACT_INVOKE,
				Payload: invokePayload,
			},
		}}
	tx3.CreationDate = tx3.CreateDate()
	tx3.TxHash = tx3.Hash()

	txs := modules.Transactions{}
	//txs = append(txs, &tx1)
	txs = append(txs, &tx2)
	//txs = append(txs, &tx3)
	unit := modules.Unit{
		UnitHeader: header,
		Txs:        txs,
	}
	unit.UnitSize = unit.Size()
	unit.UnitHash = unit.Hash()

	if err := SaveUnit(unit, false); err != nil {
		log.Println(err)
	}
}

func TestGetstate(t *testing.T) {
	key := fmt.Sprintf("%s%s",
		storage.CONTRACT_STATE_PREFIX,
		"contract0000")
	data := storage.GetPrefix([]byte(key))
	for k, v := range data {
		fmt.Println("key=", k, " ,value=", v)
	}
}

type TestByte string

func TestRlpDecode(t *testing.T) {
	var t1, t2, t3 TestByte
	t1 = "111"
	t2 = "222"
	t3 = "333"

	bytes := []TestByte{t1, t2, t3}
	encodeBytes, _ := rlp.EncodeToBytes(bytes)
	var data []TestByte
	rlp.DecodeBytes(encodeBytes, &data)
	fmt.Printf("%q", data)
}

func TestCreateUnit(t *testing.T) {
	addr := common.Address{} // minner addr
	addr.SetString("P1FYoQg1QHxAuBEgDy7c5XDWh3GLzLTmrNM")
	//units, err := CreateUnit(&addr, time.Now())
	units, err := CreateUnit(&addr, nil, nil, time.Now())
	if err != nil {
		log.Println("create unit error:", err)
	} else {
		log.Println("New unit:", units)
	}
}

func TestGetContractState(t *testing.T) {
	version, value := storage.GetContractState("contract_template0000", "name")
	log.Println(version)
	log.Println(value)
	data := storage.GetTplAllState("contract_template0000")
	for k, v := range data {
		log.Println(k, v)
	}
}
