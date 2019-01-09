/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package common

import (
	"log"
	"reflect"
	"testing"
	"time"

	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

func mockUnitRepository() *UnitRepository {
	db, _ := ptndb.NewMemDatabase()
	//l := plog.NewTestLog()
	return NewUnitRepository4Db(db)
}

//func mockUnitRepositoryLeveldb(path string) *UnitRepository {
//	db, _ := ptndb.NewLDBDatabase(path, 0, 0)
//	return NewUnitRepository4Db(db)
//}

func TestGenesisUnit(t *testing.T) {
	payload := new(modules.PaymentPayload)
	payload.LockTime = 999

	msg := modules.NewMessage(modules.APP_PAYMENT, payload)
	msgs := make([]*modules.Message, 0)
	tx := modules.NewTransaction(append(msgs, msg))
	asset := modules.NewPTNAsset()

	gUnit, _ := NewGenesisUnit(modules.Transactions{tx}, time.Now().Unix(), asset)

	log.Println("Genesis unit struct:")
	log.Println("parent units:", gUnit.UnitHeader.ParentsHash)
	log.Println("asset ids:", gUnit.UnitHeader.AssetIDs)
	log.Println("group_sign:", gUnit.UnitHeader.GroupSign)
	log.Println("Root:", gUnit.UnitHeader.TxRoot)
	log.Println("Number:", gUnit.UnitHeader.Number.String())

}

func TestGenGenesisConfigPayload(t *testing.T) {
	var genesisConf core.Genesis
	genesisConf.SystemConfig.DepositRate = "0.02"

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
	rep := mockUnitRepository()

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
	addr0 := crypto.PubkeyToAddress(&key.PublicKey)

	sig, err := crypto.Sign(header.Hash().Bytes(), key)
	if err != nil {
		log.Println("sign header occured error: ", err)
	}
	auth := new(modules.Authentifier)
	auth.R = sig[:32]
	auth.S = sig[32:64]
	auth.V = sig[64:]
	auth.Address = addr0
	header.Authors = *auth
	contractTplPayload := modules.NewContractTplPayload([]byte("contract_template0000"),
		"TestContractTpl", "./contract", "1.1.1", 1024,
		[]byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159})
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Version: &modules.StateVersion{
		Height:  rep.GenesisHeight(),
		TxIndex: 0,
	}})
	writeSet := []modules.ContractWriteSet{
		{
			Key:   "name",
			Value: modules.ToPayloadMapValueBytes("Joe"),
		},
		{
			Key:   "age",
			Value: modules.ToPayloadMapValueBytes(10),
		},
	}
	deployPayload := modules.NewContractDeployPayload([]byte("contract_template0000"), []byte("contract0000"),
		"testDeploy", nil, 10, nil, readSet, writeSet)

	invokePayload := &modules.ContractInvokePayload{
		ContractId: []byte("contract0000"),
		Args:       [][]byte{[]byte("initial")},
		ReadSet:    readSet,
		WriteSet: []modules.ContractWriteSet{
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
		TxMessages: []*modules.Message{
			{
				App:     modules.APP_CONTRACT_TPL,
				Payload: contractTplPayload,
			},
		},
	}
	t.Logf("Tx Hash:%s", tx1.Hash().String())
	//tx1.TxHash = tx1.Hash()

	tx2 := modules.Transaction{
		TxMessages: []*modules.Message{
			{
				App:     modules.APP_CONTRACT_DEPLOY,
				Payload: deployPayload,
			},
		},
	}
	//tx2.TxHash = tx2.Hash()
	t.Logf("Tx Hash:%s", tx2.Hash().String())
	tx3 := modules.Transaction{
		TxMessages: []*modules.Message{
			{
				App:     modules.APP_CONTRACT_INVOKE,
				Payload: invokePayload,
			},
		}}
	//tx3.TxHash = tx3.Hash()
	t.Logf("Tx Hash:%s", tx3.Hash().String())
	txs := modules.Transactions{}
	//txs = append(txs, &tx1)
	txs = append(txs, &tx2)
	//txs = append(txs, &tx3)
	unit := &modules.Unit{
		UnitHeader: header,
		Txs:        txs,
	}
	unit.UnitSize = unit.Size()
	unit.UnitHash = unit.Hash()

	if err := rep.SaveUnit(unit, nil, true, true); err != nil {
		log.Println(err)
	}
}

//func TestGetstate(t *testing.T) {
//	rep:=mockUnitRepository()
//	key := fmt.Sprintf("%s%s",
//		storage.CONTRACT_STATE_PREFIX,
//		"contract0000")
//	data := rep.GetPrefix(Dbconn, []byte(key))
//	for k, v := range data {
//		fmt.Println("key=", k, " ,value=", v)
//	}
//}

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

	rep := mockUnitRepository()
	addr := common.Address{} // minner addr
	addr.SetString("P1FYoQg1QHxAuBEgDy7c5XDWh3GLzLTmrNM")
	//units, err := CreateUnit(&addr, time.Now())
	units, err := rep.CreateUnit(&addr, nil, nil, time.Now())
	if err != nil {
		log.Println("create unit error:", err)
	} else {
		log.Println("New unit:", units)
	}
}

//func TestGetContractState(t *testing.T) {
//	rep:=mockUnitRepository()
//	version, value := rep.GetContractState( "contract_template0000", "name")
//	log.Println(version)
//	log.Println(value)
//	data := rep.GetTplAllState( "contract_template0000")
//	for k, v := range data {
//		log.Println(k, v)
//	}
//}

//func TestGetContractTplState(t *testing.T) {
//	rep := mockUnitRepositoryLeveldb("D:\\test\\levedb")
//	version, bytecode, name, path := rep.statedb.GetContractTpl([]byte{161, 143, 28, 181, 148, 91, 16, 93, 244, 3, 53, 1, 129, 124, 224, 247, 234, 93, 92, 36, 193, 44, 0, 194, 159, 239, 237, 151, 224, 47, 240, 84})
//	log.Println("version=", version.Height, version.TxIndex)
//	log.Println("bytecode=", bytecode)
//	log.Println("name=", name)
//	log.Println("path=", path)
//}

//func TestGetTransaction(t *testing.T) {
//	rep := mockUnitRepositoryLeveldb("E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb")
//	txHash := common.Hash{}
//	if err := txHash.SetHexString("0xe146ada75bbdeeebac5902e553154104f4349d3ddab3a5ffbdc3d33c9d72792b"); err != nil {
//		log.Println("get tx hex hash error:", err.Error())
//		return
//	}
//	tx, unithash, number, index := rep.dagdb.GetTransaction(txHash)
//	fmt.Println("tx:", tx)
//	fmt.Println("unithash:", unithash)
//	fmt.Println("number:", number)
//	fmt.Println("index:", index)
//}

func TestPaymentTransactionRLP(t *testing.T) {
	p := common.Hash{}
	p.SetString("0000000000000000022222222222")
	aid := modules.IDType16{}
	aid.SetBytes([]byte("xxxxxxxxxxxxxxxxxx"))

	// TODO test PaymentPayload
	txin := modules.Input{
		PreviousOutPoint: &modules.OutPoint{
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
		Asset: &modules.Asset{
			AssetId:  aid,
			UniqueId: aid,
		},
	}
	payment := modules.PaymentPayload{
		Inputs:   []*modules.Input{&txin},
		Outputs:  []*modules.Output{&txout},
		LockTime: 12,
	}

	tx2 := modules.Transaction{
		TxMessages: []*modules.Message{
			{
				App:     modules.APP_PAYMENT,
				Payload: payment,
			},
		},
	}
	//tx2.TxHash = tx2.Hash()
	fmt.Println("Original data:", payment)
	b, _ := rlp.EncodeToBytes(tx2)
	var tx modules.Transaction
	if err := rlp.DecodeBytes(b, &tx); err != nil {
		fmt.Println("TestPaymentTransactionRLP error:", err.Error())
	} else {
		for _, msg := range tx.TxMessages {
			if msg.App == modules.APP_PAYMENT {
				var pl *modules.PaymentPayload
				pl, ok := msg.Payload.(*modules.PaymentPayload)
				if !ok {
					fmt.Println("Payment payload ExtractFrInterface error:", err.Error())
				} else {
					fmt.Println("Payment payload:", pl)
				}
			}
		}
	}

}

func TestContractTplPayloadTransactionRLP(t *testing.T) {
	rep := mockUnitRepository()
	// TODO test ContractTplPayload
	contractTplPayload := modules.ContractTplPayload{
		TemplateId: []byte("contract_template0000"),
		Bytecode:   []byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159},
		Name:       "TestContractTpl",
		Path:       "./contract",
	}
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Version: &modules.StateVersion{
		Height:  rep.GenesisHeight(),
		TxIndex: 0,
	}})
	tx1 := modules.Transaction{
		TxMessages: []*modules.Message{
			{
				App:     modules.APP_CONTRACT_TPL,
				Payload: contractTplPayload,
			},
		},
	}
	//tx1.TxHash = tx1.Hash()

	fmt.Println(">>>>>>>>  Original transaction:")
	fmt.Println(">>>>>>>>  hash:", tx1.Hash())
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
			fmt.Println("======== hash:", txDecode.Hash())
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
	rep := mockUnitRepository()
	// TODO test ContractTplPayload
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Version: &modules.StateVersion{
		Height:  rep.GenesisHeight(),
		TxIndex: 0,
	}})
	writeSet := []modules.ContractWriteSet{
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
	//et := time.Duration(12)
	deployPayload := modules.ContractDeployPayload{
		TemplateId: []byte("contract_template0000"),
		ContractId: []byte("contract0000"),
		Name:       "testdeploy",
		Args:       [][]byte{[]byte{1, 2, 3}, []byte{4, 5, 6}},
		//ExecutionTime: et,
		Jury:     []common.Address{addr},
		ReadSet:  readSet,
		WriteSet: writeSet,
	}
	tx1 := modules.Transaction{
		TxMessages: []*modules.Message{
			{
				App:     modules.APP_CONTRACT_DEPLOY,
				Payload: deployPayload,
			},
		},
	}
	//tx1.TxHash = tx1.Hash()

	fmt.Println(">>>>>>>>  Original transaction:")
	fmt.Println(">>>>>>>>  hash:", tx1.Hash())
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
			fmt.Println("======== hash:", txDecode.Hash())
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
						//fmt.Println(deployPayload.ExecutionTime)
					}
				}
			}
		}
	}
}
