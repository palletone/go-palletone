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
	"github.com/palletone/go-palletone/tokenengine"
	"reflect"
	"testing"
	"time"

	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/storage"
)

func mockUnitRepository() *UnitRepository {
	db, _ := ptndb.NewMemDatabase()
	//l := plog.NewTestLog()
	return NewUnitRepository4Db(db, tokenengine.Instance)
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

	gUnit, _ := NewGenesisUnit(modules.Transactions{tx}, time.Now().Unix(), asset, -1, common.Hash{})

	log.Debug("Genesis unit struct:")
	log.Debugf("parent units:%#x", gUnit.UnitHeader.ParentsHash)
	//log.Println("asset ids:", gUnit.UnitHeader.AssetIDs)
	log.Debugf("group_sign:%x", gUnit.UnitHeader.GroupSign)
	log.Debugf("Root:%x", gUnit.UnitHeader.TxRoot)
	log.Debugf("Number:%s", gUnit.UnitHeader.Number.String())

}

//func TestGenGenesisConfigPayload(t *testing.T) {
//	var genesisConf core.Genesis
//	genesisConf.InitialParameters.DepositRate = "0.02"
//
//	genesisConf.InitialParameters.MediatorInterval = 10
//
//	payloads, err := GenGenesisConfigPayload(&genesisConf, &modules.Asset{})
//
//	if err != nil {
//		log.Debug("TestGenGenesisConfigPayload", "err", err)
//	}
//	for _, payload := range payloads {
//
//		for _, w := range payload.WriteSet {
//			k := w.Key
//			v := w.Value
//			log.Debug("Key:", k, v)
//		}
//	}
//}

func TestSaveUnit(t *testing.T) {
	rep := mockUnitRepository()

	addr := common.Address{}
	addr.SetString("P12EA8oRMJbAtKHbaXGy8MGgzM8AMPYxkN1")
	//ks := keystore.NewKeyStore("./keystore", 1<<18, 1)

	p := common.Hash{}
	p.SetString("0000000000000000000000000000000")
	aid := modules.AssetId{}
	aid.SetBytes([]byte("xxxxxxxxxxxxxxxxxx"))
	header := new(modules.Header)
	header.ParentsHash = append(header.ParentsHash, p)
	header.Number = &modules.ChainIndex{AssetID: modules.PTNCOIN, Index: 0}
	//header.AssetIDs = []modules.AssetId{aid}
	key, _ := crypto.GenerateKey()
	//addr0 := crypto.PubkeyToAddress(&key.PublicKey)

	sig, err := crypto.Sign(header.Hash().Bytes(), key)
	if err != nil {
		log.Debug("sign header occured error: ", err)
	}
	auth := new(modules.Authentifier)
	auth.Signature = sig
	auth.PubKey = crypto.CompressPubkey(&key.PublicKey)

	header.Authors = *auth
	contractTplPayload := modules.NewContractTplPayload([]byte("contract_template0000"),
		1024,
		[]byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159}, modules.ContractError{})
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Version: &modules.StateVersion{
		Height:  &modules.ChainIndex{},
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
		"testDeploy", nil, nil, readSet, writeSet, modules.ContractError{})

	invokePayload := &modules.ContractInvokePayload{
		ContractId: []byte("contract0000"),
		ReadSet:    readSet,
		WriteSet: []modules.ContractWriteSet{
			{
				Key:   "name",
				Value: modules.ToPayloadMapValueBytes("Alice"),
			},
			{
				Key:      "age",
				IsDelete: true,
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

	if err := rep.SaveUnit(unit, true); err != nil {
		log.Debug("TestSaveUnit", "err", err)
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
	assert.Equal(t, bytes, data)
	fmt.Printf("%q", data)
}

//func TestCreateUnit(t *testing.T) {
//
//	rep := mockUnitRepository()
//	addr := common.Address{} // minner addr
//	addr.SetString("P1FYoQg1QHxAuBEgDy7c5XDWh3GLzLTmrNM")
//	//units, err := CreateUnit(&addr, time.Now())
//	units, err := rep.CreateUnit(&addr, nil, time.Now())
//	if err != nil {
//		log.Println("create unit error:", err)
//	} else {
//		log.Println("New unit:", units)
//	}
//}

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
	aid := modules.AssetId{}
	aid.SetBytes([]byte("xxxxxxxxxxxxxxxxxx"))
	uid := modules.UniqueId{}
	uid.SetBytes([]byte{0xff, 0xee})
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
			UniqueId: uid,
		},
	}
	payment := &modules.PaymentPayload{
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
	t.Log("data", tx2)
	b, _ := rlp.EncodeToBytes(&tx2)
	t.Log("rlp", b)
	var tx modules.Transaction
	//if err := rlp.DecodeBytes(b, &tx); err != nil {
	//	fmt.Println("TestPaymentTransactionRLP error:", err.Error())
	//} else {
	//	for _, msg := range tx.TxMessages {
	//		if msg.App == modules.APP_PAYMENT {
	//			var pl *modules.PaymentPayload
	//			pl, ok := msg.Payload.(*modules.PaymentPayload)
	//			if !ok {
	//				fmt.Println("Payment payload ExtractFrInterface error:", err.Error())
	//			} else {
	//				fmt.Println("Payment payload:", pl)
	//			}
	//		}
	//	}
	//}
	err := rlp.DecodeBytes(b, &tx)
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			var pl *modules.PaymentPayload
			pl, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				fmt.Println("Payment payload ExtractFrInterface error:", err.Error())
			} else {
				fmt.Println("Payment payload:", pl)
				t.Log("11111111")
				assert.Equal(t, payment, pl)
			}
		}

	}
	t.Log("data", tx)
	assert.Equal(t, tx2.Hash(), tx.Hash())

}

func TestContractTplPayloadTransactionRLP(t *testing.T) {
	//rep := mockUnitRepository()
	// TODO test ContractTplPayload
	contractTplPayload := modules.ContractTplPayload{
		TemplateId: []byte("contract_template0000"),
		ByteCode:   []byte{175, 52, 23, 180, 156, 109, 17, 232, 166, 226, 84, 225, 173, 184, 229, 159},
		//Name:       "TestContractTpl",
		//Path:       "./contract",
	}
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Version: &modules.StateVersion{
		Height:  &modules.ChainIndex{},
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
	//rep := mockUnitRepository()
	// TODO test ContractTplPayload
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Version: &modules.StateVersion{
		Height:  &modules.ChainIndex{},
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
		//TemplateId: []byte("contract_template0000"),
		ContractId: []byte("contract0000"),
		Name:       "testdeploy",
		//ExecutionTime: et,
		//Jury:     []common.Address{addr},
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

func creatFeeTx(isContractTx bool, pubKey [][]byte, amount uint64, aid modules.AssetId) *modules.TxPoolTransaction {
	tx := modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	if isContractTx {
		sigs := make([]modules.SignatureSet, 0)
		for _, pk := range pubKey {
			sigSet := modules.SignatureSet{
				PubKey: pk,
			}
			sigs = append(sigs, sigSet)
		}
		conSig := &modules.Message{
			App: modules.APP_CONTRACT_STOP_REQUEST,
		}
		msgSig := &modules.Message{
			App: modules.APP_SIGNATURE,
			Payload: &modules.SignaturePayload{
				Signatures: sigs,
			},
		}
		tx.TxMessages = append(tx.TxMessages, conSig)
		tx.TxMessages = append(tx.TxMessages, msgSig)
	}

	txPTx := &modules.TxPoolTransaction{
		Tx:    &tx,
		TxFee: make([]*modules.Addition, 0),
	}
	txPTx.TxFee = append(txPTx.TxFee, &modules.Addition{
		Amount: amount,
		Asset: &modules.Asset{
			AssetId: aid,
		}})
	return txPTx
}

//
//func TestComputeTxFees(t *testing.T) {
//	m, _ := common.StringToAddress("P1K7JsRvDc5THJe6TrtfdRNxp6ZkNiboy9z")
//	txs := make([]*modules.TxPoolTransaction, 0)
//	pks := make([][]byte, 0)
//	aId := modules.AssetId{}
//	tx := &modules.TxPoolTransaction{}
//
//	//1
//	pks = [][]byte{
//		{0x01}, {0x02}, {0x03}, {0x04}, {0x05}}
//	aId = modules.AssetId{'p', 't', 'n'}
//	tx = creatFeeTx(true, pks, 10, aId)
//	txs = append(txs, tx)
//
//	//	log.Info("TestComputeTxFees", "txs:", tx)
//	/*
//		//2
//		pks = [][]byte{
//			{0x01}, {0x02}, {0x03}, {0x04}}
//		aId = modules.AssetId{'p', 't', 'n'}
//		tx = creatFeeTx(true, pks, 10, aId)
//		txs = append(txs, tx)
//
//		//3
//		pks = [][]byte{
//			{0x05}, {0x06}, {0x07}, {0x08}}
//		aId = modules.AssetId{'p', 't', 'n'}
//		tx = creatFeeTx(true, pks, 10, aId)
//		txs = append(txs, tx)
//
//		//4
//		pks = [][]byte{
//			{0x01}, {0x02}, {0x03}, {0x04}}
//		aId = modules.AssetId{'a', 'b', 'c'}
//		tx = creatFeeTx(true, pks, 10, aId)
//		txs = append(txs, tx)
//
//		//5
//		pks = [][]byte{
//			{0x01}, {0x02}, {0x03}, {0x04}}
//		aId = modules.AssetId{'a', 'b', 'c'}
//		tx = creatFeeTx(true, pks, 10, aId)
//		txs = append(txs, tx)
//	*/
//	//log.Info("TestComputeTxFees", "txs:", txs)
//	ads, err := ComputeTxFees(&m, txs)
//	log.Info("TestComputeTxFees", "txs:", ads)
//	if err == nil {
//		outAds := arrangeAdditionFeeList(ads)
//		log.Debug("TestComputeTxFees", "outAds:", outAds)
//		rewards := map[common.Address][]modules.AmountAsset{}
//
//		for _, ad := range outAds {
//
//			reward, ok := rewards[ad.Addr]
//			if !ok {
//				reward = []modules.AmountAsset{}
//			}
//			reward = addIncome(reward, ad.AmountAsset)
//			rewards[ad.Addr] = reward
//			//totalIncome += ad.AmountAsset.Amount
//		}
//		coinbase := createCoinbasePaymentMsg(rewards)
//
//		log.Debug("TestComputeTxFees", "coinbase", coinbase, "rewards", rewards)
//
//	}
//}

func TestContractStateVrf(t *testing.T) {
	contractId := []byte("TestContractVrf")
	eleW := modules.ElectionNode{
		JuryCount: 0,
		EleList: []modules.ElectionInf{
			{
				Proof: []byte("abc"), PublicKey: []byte("def"),
			},
		},
	}
	ver := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	log.Debug("TestContractStateVrf", "ElectionNode", eleW)

	db, _ := ptndb.NewMemDatabase()
	statedb := storage.NewStateDb(db)
	//eleW_bytes, _ := rlp.EncodeToBytes(eleW)
	//write
	err := statedb.SaveContractJury(contractId, eleW, ver)
	assert.Nil(t, err)

	//ws := modules.NewWriteSet("ElectionList", eleW_bytes)
	//if statedb.SaveContractState(contractId, ws, ver) != nil {
	//	log.Debug("TestContractStateVrf, SaveContractState fail")
	//	return
	//}

	//read
	eler, err := statedb.GetContractJury(contractId)
	assert.Nil(t, err)
	t.Logf("%v", eler)
	//eleByte, _, err := statedb.GetContractState(contractId, "ElectionList")
	//if err != nil {
	//	log.Debug("TestContractStateVrf, GetContractState fail", "error", err)
	//	return
	//}
	//var eler []modules.ElectionInf
	//err1 := rlp.DecodeBytes(eleByte, &eler)
	//log.Infof("%v", err1)
	//log.Infof("%v", eler)
}

func TestContractRlpEncode(t *testing.T) {
	ads := []string{"P1QFTh1Xq2JpfTbu9bfaMfWh2sR1nHrMV8z", "P1NHVBFRkooh8HD9SvtvU3bpbeVmuGKPPuF",
		"P1PpgjUC7Nkxgi5KdKCGx2tMu6F5wfPGrVX", "P1MBXJypFCsQpafDGi9ivEooR8QiYmxq4qw"}

	addrs := make([]common.Hash, 0)
	for _, s := range ads {
		a, _ := common.StringToAddress(s)
		addrs = append(addrs, util.RlpHash(a))
	}
	log.Debug("TestContractRlpEncode", "addrHash", addrs)

	addrBytes, err := rlp.EncodeToBytes(addrs)
	if err != nil {
		log.Debug("TestContractRlpEncode", "EncodeToBytes err", err)
		return
	}
	log.Debug("TestContractRlpEncode", "EncodeToBytes", addrBytes)

	var addh []common.Hash
	//addh := make([]common.Hash, 4)
	err = rlp.DecodeBytes(addrBytes, &addh)
	if err != nil {
		log.Debug("TestContractRlpEncode", "DecodeBytes err", err)
		return
	}

	log.Debug("TestContractRlpEncode", "addh", addh)
}

func TestContractTxsIllegal(t *testing.T) {
	//make tx
	readSet := []modules.ContractReadSet{}
	readSet = append(readSet, modules.ContractReadSet{Key: "name", Version: &modules.StateVersion{
		Height:  &modules.ChainIndex{Index: 123},
		TxIndex: 1,
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
	addr.SetString("PC2EA8oRMJbAtKHbaXGy8MGgzM8AMPYxkN1")
	deployPayload := &modules.ContractDeployPayload{
		//TemplateId: []byte("contract_template0000"),
		ContractId: []byte("contract0000"),
		Name:       "testDeploy",
		ReadSet:    readSet,
		WriteSet:   writeSet,
	}
	tx1 := &modules.Transaction{
		TxMessages: []*modules.Message{
			{
				App:     modules.APP_CONTRACT_DEPLOY_REQUEST,
				Payload: nil,
			},
			{
				App:     modules.APP_CONTRACT_DEPLOY,
				Payload: deployPayload,
			},
		},
	}
	txs := make([]*modules.Transaction, 0)
	txs = append(txs, tx1)

	//dag
	db, _ := ptndb.NewMemDatabase()
	statedb := storage.NewStateDb(db)

	//set state
	ver := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	err := statedb.SaveContractState([]byte("contract0000"), &writeSet[0], ver)
	if err != nil {
		log.Debug("TestContractTxIllegal", "SaveContractState err", err)
	}

	//mark
	markTxsIllegal(statedb, txs)
}
func markTxsIllegal(dag storage.IStateDb, txs []*modules.Transaction) {
	for _, tx := range txs {
		if !tx.IsContractTx() {
			continue
		}
		if tx.IsSystemContract() {
			continue
		}
		var readSet []modules.ContractReadSet
		var contractId []byte

		for _, msg := range tx.TxMessages {
			switch msg.App {
			case modules.APP_CONTRACT_DEPLOY:
				payload := msg.Payload.(*modules.ContractDeployPayload)
				readSet = payload.ReadSet
				contractId = payload.ContractId
			case modules.APP_CONTRACT_INVOKE:
				payload := msg.Payload.(*modules.ContractInvokePayload)
				readSet = payload.ReadSet
				contractId = payload.ContractId
			case modules.APP_CONTRACT_STOP:
				payload := msg.Payload.(*modules.ContractStopPayload)
				readSet = payload.ReadSet
				contractId = payload.ContractId
			}
		}
		valid := checkReadSetValid(dag, contractId, readSet)
		tx.Illegal = !valid
	}
}

// func TestCoinbase(t *testing.T) {
// 	db, _ := ptndb.NewLDBDatabase("D:\\test\\node1\\palletone\\leveldb", 700, 1024)
// 	rep := NewUnitRepository4Db(db)
// 	addr, _ := common.StringToAddress("P19VvSXKfTwE7HwUAVNbJrdUJzCYu9tJCDv")
// 	txs := readFile()
// 	t.Logf("Tx Count:%d", len(txs))
// 	tt := time.Now()
// 	ads, _ := rep.ComputeTxFeesAllocate(addr, txs)
// 	outAds := arrangeAdditionFeeList(ads)
// 	coinbase, rewards, _ := rep.CreateCoinbase(outAds, 6)
// 	t.Logf("create coinbase tx cost time %s", time.Since(tt))
// 	js, _ := json.Marshal(coinbase)
// 	t.Log(string(js))
// 	t.Log(rewards)
// }
// func readFile() []*modules.Transaction {
// 	txs := []*modules.Transaction{}
// 	rw, err := os.Open("D:\\test\\node1\\geneSignResult.txt")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer rw.Close()
// 	rb := bufio.NewReader(rw)
// 	i := 0
// 	for {
// 		line, _, err := rb.ReadLine()
// 		if err == io.EOF {
// 			break
// 		}
// 		data, _ := hex.DecodeString(string(line))
// 		tx := &modules.Transaction{}
// 		rlp.DecodeBytes(data, tx)
// 		txs = append(txs, tx)
// 		i++
// 		if i > 100000 {
// 			return txs
// 		}
// 	}
// 	return txs
// }
