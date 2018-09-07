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
	"reflect"
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
	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
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
		TemplateId: []byte("contract_template0000"),
		Bytecode:   []byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159},
		Name:       "TestContractTpl",
		Path:       "./contract",
	}
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Value: &modules.StateVersion{
		Height:  GenesisHeight(Dbconn),
		TxIndex: 0,
	}})
	writeSet := []modules.PayloadMapStruct{
		{
			Key:   "name",
			Value: modules.ToPayloadMapValueBytes("Joe"),
		},
		{
			Key:   "age",
			Value: modules.ToPayloadMapValueBytes(10),
		},
	}
	deployPayload := modules.ContractDeployPayload{
		TemplateId: []byte("contract_template0000"),
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
				Value: modules.ToPayloadMapValueBytes("Alice"),
			},
			{
				Key: "age",
				Value: modules.ToPayloadMapValueBytes(modules.DelContractState{
					IsDelete: true,
				}),
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
	tx1.TxHash = tx1.Hash()

	tx2 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_CONTRACT_DEPLOY,
				Payload: deployPayload,
			},
		},
	}
	tx2.TxHash = tx2.Hash()

	tx3 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_CONTRACT_INVOKE,
				Payload: invokePayload,
			},
		}}
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

	if err := SaveUnit(Dbconn, unit, true); err != nil {
		log.Println(err)
	}
}

func TestGetstate(t *testing.T) {
	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	key := fmt.Sprintf("%s%s",
		storage.CONTRACT_STATE_PREFIX,
		"contract0000")
	data := storage.GetPrefix(Dbconn, []byte(key))
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

	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	addr := common.Address{} // minner addr
	addr.SetString("P1FYoQg1QHxAuBEgDy7c5XDWh3GLzLTmrNM")
	//units, err := CreateUnit(&addr, time.Now())
	units, err := CreateUnit(Dbconn, &addr, nil, nil, time.Now())
	if err != nil {
		log.Println("create unit error:", err)
	} else {
		log.Println("New unit:", units)
	}
}

func TestGetContractState(t *testing.T) {
	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	version, value := storage.GetContractState(Dbconn, "contract_template0000", "name")
	log.Println(version)
	log.Println(value)
	data := storage.GetTplAllState(Dbconn, "contract_template0000")
	for k, v := range data {
		log.Println(k, v)
	}
}

func TestPaymentTransactionRLP(t *testing.T) {
	p := common.Hash{}
	p.SetString("0000000000000000022222222222")
	aid := modules.IDType16{}
	aid.SetBytes([]byte("xxxxxxxxxxxxxxxxxx"))

	// TODO test PaymentPayload
	txin := modules.Input{
		PreviousOutPoint: modules.OutPoint{
			TxHash:       p,
			MessageIndex: 1234,
			OutIndex:     12344,
		},
		SignatureScript: []byte("1234567890"),
		Extra:           []byte("990019202020"),
	}
	txout := modules.Output{
		Value:    1,
		PkScript: []byte("kssssssssssssssssssslsll"),
		Asset: modules.Asset{
			AssetId:  aid,
			UniqueId: aid,
			ChainId:  11,
		},
	}
	payment := modules.PaymentPayload{
		Input:    []*modules.Input{&txin},
		Output:   []*modules.Output{&txout},
		LockTime: 12,
	}

	tx2 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_PAYMENT,
				Payload: payment,
			},
		},
	}
	tx2.TxHash = tx2.Hash()
	fmt.Println("Original data:", payment)
	b, _ := rlp.EncodeToBytes(tx2)
	var tx modules.Transaction
	if err := rlp.DecodeBytes(b, &tx); err != nil {
		fmt.Println("TestPaymentTransactionRLP error:", err.Error())
	} else {
		for _, msg := range tx.TxMessages {
			if msg.App == modules.APP_PAYMENT {
				var pl modules.PaymentPayload
				if err := pl.ExtractFrInterface(msg.Payload); err != nil {
					fmt.Println("Payment payload ExtractFrInterface error:", err.Error())
				} else {
					fmt.Println("Payment payload:", pl)
				}
			}
		}
	}

}

func TestContractTplPayloadTransactionRLP(t *testing.T) {
	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	// TODO test ContractTplPayload
	contractTplPayload := modules.ContractTplPayload{
		TemplateId: []byte("contract_template0000"),
		Bytecode:   []byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159},
		Name:       "TestContractTpl",
		Path:       "./contract",
	}
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Value: &modules.StateVersion{
		Height:  GenesisHeight(Dbconn),
		TxIndex: 0,
	}})
	tx1 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_CONTRACT_TPL,
				Payload: contractTplPayload,
			},
		},
	}
	tx1.TxHash = tx1.Hash()

	fmt.Println(">>>>>>>>  Original transaction:")
	fmt.Println(">>>>>>>>  hash:", tx1.TxHash)
	for index, msg := range tx1.TxMessages {
		fmt.Printf(">>>>>>>>  message[%d]:%v\n", index, msg)
		fmt.Println(">>>>>>>> payload type:", reflect.TypeOf(msg.Payload))
		fmt.Printf(">>>>>>>>  message[%d] payload:%v\n", index, msg.Payload)
	}
	encodeData, err := rlp.EncodeToBytes(tx1)
	if err != nil {
		fmt.Println("Encode tx1 error:", err.Error())
	} else {
		txDecode := new(modules.Transaction)
		if err := rlp.DecodeBytes(encodeData, txDecode); err != nil {
			fmt.Println("Decode tx error:", err.Error())
		} else {
			fmt.Println("======== Decode transaction:")
			fmt.Println("======== hash:", txDecode.TxHash)
			for index, msg := range txDecode.TxMessages {
				fmt.Printf("======== message[%d]:%v\n", index, msg)
				switch msg.App {
				case modules.APP_CONTRACT_TPL:
					tplPayload, ok := msg.Payload.(*modules.ContractTplPayload)
					if !ok {
						fmt.Println("Contract template payload ExtractFrInterface error:")
					} else {
						fmt.Println("Contract template payload:", tplPayload)
					}
				}
			}
		}
	}
}

func TestContractDeployPayloadTransactionRLP(t *testing.T) {
	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	if Dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	// TODO test ContractTplPayload
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Value: &modules.StateVersion{
		Height:  GenesisHeight(Dbconn),
		TxIndex: 0,
	}})
	writeSet := []modules.PayloadMapStruct{
		{
			Key:   "name",
			Value: []byte("Joe"),
		},
		{
			Key:   "age",
			Value: modules.ToPayloadMapValueBytes(uint8(10)),
		},
	}
	addr := common.Address{}
	addr.SetString("P12EA8oRMJbAtKHbaXGy8MGgzM8AMPYxkN1")
	et := time.Duration(12)
	deployPayload := modules.ContractDeployPayload{
		TemplateId:   []byte("contract_template0000"),
		ContractId:   []byte("contract0000"),
		Name:         "testdeploy",
		Args:         [][]byte{[]byte{1, 2, 3}, []byte{4, 5, 6}},
		Excutiontime: et,
		Jury:         []common.Address{addr},
		ReadSet:      readSet,
		WriteSet:     writeSet,
	}
	tx1 := modules.Transaction{
		TxMessages: []modules.Message{
			{
				App:     modules.APP_CONTRACT_DEPLOY,
				Payload: deployPayload,
			},
		},
	}
	tx1.TxHash = tx1.Hash()

	fmt.Println(">>>>>>>>  Original transaction:")
	fmt.Println(">>>>>>>>  hash:", tx1.TxHash)
	for index, msg := range tx1.TxMessages {
		fmt.Printf(">>>>>>>>  message[%d]:%v\n", index, msg)
		fmt.Println(">>>>>>>> payload type:", reflect.TypeOf(msg.Payload))
		fmt.Printf(">>>>>>>>  message[%d] payload:%v\n", index, msg.Payload)
	}
	encodeData, err := rlp.EncodeToBytes(&tx1)
	if err != nil {
		fmt.Println("Encode tx1 error:", err.Error())
	} else {
		txDecode := new(modules.Transaction)
		if err := rlp.DecodeBytes(encodeData, txDecode); err != nil {
			fmt.Println("Encode data:", encodeData)
			fmt.Println("Decode tx error:", err.Error())
		} else {
			fmt.Println("======== Decode transaction:")
			fmt.Println("======== hash:", txDecode.TxHash)
			for index, msg := range txDecode.TxMessages {
				fmt.Printf("======== message[%d]:%v\n", index, msg)
				switch msg.App {
				case modules.APP_CONTRACT_DEPLOY:
					deployPayload, ok := msg.Payload.(*modules.ContractDeployPayload)
					if !ok {
						fmt.Println("Get contract deploypayload error.")
					} else {
						fmt.Println(deployPayload.Name)
						fmt.Println(deployPayload.ContractId)
						fmt.Println(deployPayload.Excutiontime)
					}
				}
			}
		}
	}
}
