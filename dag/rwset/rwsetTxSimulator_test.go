package rwset

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/stretchr/testify/assert"
)

func TestRwSetTxSimulator_GetTokenBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dag := dag.NewMockIDag(mockCtrl)
	simulator := &RwSetTxSimulator{}
	simulator.dag = dag
	mockUtxos := mockUtxos()
	mockPtnUtxos := mockPtnUtxos()
	dag.EXPECT().GetAddrUtxos(gomock.Any()).Return(mockUtxos, nil).AnyTimes()
	dag.EXPECT().GetAddr1TokenUtxos(gomock.Any(), gomock.Any()).Return(mockPtnUtxos, nil).AnyTimes()
	addr := common.HexToAddress("P18WsGzDNcwhxTyELzDK2vv7S6f3XSWXuTg")
	balance, err := simulator.GetTokenBalance("PalletOne", addr, nil)
	assert.Nil(t, err)
	assert.True(t, len(balance) == 2, "mock has 2 asset,but current is "+strconv.Itoa(len(balance)))
	for k, v := range balance {
		t.Logf("Key:{%s},Value:%d", k.String(), v)
	}
	ptnAsset := &modules.Asset{AssetId: modules.PTNCOIN}
	balance1, err := simulator.GetTokenBalance("PalletOne", addr, ptnAsset)
	assert.Nil(t, err)
	assert.True(t, len(balance1) == 1, "for PTN asset, only need return 1 row")
	assert.Equal(t, balance1[*ptnAsset], uint64(300), "sum PTN must 300")
}
func mockUtxos() map[modules.OutPoint]*modules.Utxo {
	result := map[modules.OutPoint]*modules.Utxo{}
	p1 := modules.NewOutPoint(common.Hash{}, 0, 0)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	fmt.Printf("Mock asset1:%s\n", asset1.String())
	utxo1 := &modules.Utxo{Asset: asset1, Amount: 100, LockTime: 0}
	utxo2 := &modules.Utxo{Asset: asset1, Amount: 200, LockTime: 0}
	asset2 := &modules.Asset{AssetId: modules.BTCCOIN}
	fmt.Printf("Mock asset2:%s\n", asset2.String())
	utxo3 := &modules.Utxo{Asset: asset2, Amount: 500, LockTime: 0}
	result[*p1] = utxo1
	p2 := modules.NewOutPoint(common.Hash{}, 1, 0)
	result[*p2] = utxo2
	p3 := modules.NewOutPoint(common.Hash{}, 2, 1)
	result[*p3] = utxo3
	return result
}
func mockPtnUtxos() map[modules.OutPoint]*modules.Utxo {
	result := map[modules.OutPoint]*modules.Utxo{}
	p1 := modules.NewOutPoint(common.Hash{}, 0, 0)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	fmt.Printf("Mock asset1:%s\n", asset1.String())
	utxo1 := &modules.Utxo{Asset: asset1, Amount: 100, LockTime: 0}
	utxo2 := &modules.Utxo{Asset: asset1, Amount: 200, LockTime: 0}

	result[*p1] = utxo1
	p2 := modules.NewOutPoint(common.Hash{}, 1, 0)
	result[*p2] = utxo2
	return result
}

func TestNewTxMgr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dag := dag.NewMockIDag(mockCtrl)
	simulator := &RwSetTxSimulator{}
	simulator.dag = dag
	mockUnit := getCurrentUnit()
	dag.EXPECT().GetCurrentUnit(gomock.Any()).Return(mockUnit).AnyTimes()
	tx := createTx()
	asid, _ := modules.NewAssetId("12345", modules.AssetType_FungibleToken, 0, tx.RequestHash().Bytes(), modules.UniqueIdType_Null)

	rwm, err := NewRwSetMgr("test")
	if err != nil {
		t.Fatal(err)
		return
	}
	chain_id := asid.String()
	log.Println("chainId:", chain_id)
	ts, err1 := rwm.NewTxSimulator(dag, chain_id, tx.Hash().String(), true)
	if err1 != nil {
		t.Fatal(err1)
		return
	}

	addr := common.HexToAddress("P19s57FnXFNg5Aa5HExw681qycqsHXys29L")
	token := &modules.FungibleToken{Name: "測試", Symbol: "test", Decimals: 0, TotalSupply: 1, SupplyAddress: addr.String()}
	define, _ := json.Marshal(token)
	if err := ts.DefineToken("jay", 0, define, "P19s57FnXFNg5Aa5HExw681qycqsHXys29L"); err != nil {
		t.Fatal(err)
		return
	}

	err1 = ts.SupplyToken("jay", asid.Bytes(), nil, 10000, "P19s57FnXFNg5Aa5HExw681qycqsHXys29L")
	assert.Nil(t, err1)
	result, err2 := ts.GetTokenSupplyData("jay")
	assert.Nil(t, err2)
	assert.True(t, len(result) == 1, "for 'jay' ,only need result 1 row.")
	assert.Equal(t, asid.Bytes(), result[0].AssetId, "result asseid is not equal.")
	// getbalence
	asset := modules.Asset{AssetId: asid}
	mockUtxos := mockTestUtxos(asid)
	dag.EXPECT().GetAddr1TokenUtxos(gomock.Any(), gomock.Any()).Return(mockUtxos, nil).AnyTimes()
	amounts, _ := ts.GetTokenBalance("jay", addr, &asset)
	for assid, amount := range amounts {
		log.Printf("asstid:%s ,amount:%d", assid.String(), amount)
	}
	// done && close
	dag.EXPECT().Close().Return().AnyTimes()
	ts.Done()
	// txsimulator 执行结束后关闭它
	assert.Nil(t, rwm.CloseTxSimulator(chain_id, tx.Hash().String()))
}
func getCurrentUnit() *modules.Unit {
	txs := modules.Transactions{createTx()}
	hash := common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	pubKey, _ := hex.DecodeString("038cc8c907b29a58b00f8c2590303bfc93c69d773b9da204337678865ee0cafadb")
	//addr:= crypto.PubkeyBytesToAddress(pubKey)
	header := &modules.Header{}
	header.ParentsHash = []common.Hash{hash}
	header.TxRoot = core.DeriveSha(txs)
	headerHash := header.HashWithoutAuthor()
	sign, _ := crypto.Sign(headerHash[:], privKey)
	header.Authors = modules.Authentifier{PubKey: pubKey, Signature: sign}
	header.Time = time.Now().Unix()
	header.Number = &modules.ChainIndex{modules.NewPTNIdType(), 1}
	return &modules.Unit{UnitHeader: header, Txs: txs, ReceivedAt: time.Now()}
}

func createTx() *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	addr, _ := common.StringToAddress("P1KFk1oV2W5N3Ek86Cxv64RZH2RBGMAQoAC")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	output := modules.NewTxOut(99, lockScript, modules.NewPTNAsset())
	pay1s.AddTxOut(output)
	hash := common.HexToHash("1")
	input := modules.Input{}
	input.PreviousOutPoint = modules.NewOutPoint(hash, 0, 0)
	input.SignatureScript = []byte{}

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}

	tx := modules.NewTransaction(
		[]*modules.Message{msg},
	)
	lockScripts := map[modules.OutPoint][]byte{
		*input.PreviousOutPoint: lockScript,
	}
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		return crypto.Sign(hash, privKey)
	}
	_, err := tokenengine.Instance.SignTxAllPaymentInput(tx, 1, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		log.Println(err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	log.Printf("UnlockScript:%x", unlockScript)
	return tx
}
func mockTestUtxos(assid modules.AssetId) map[modules.OutPoint]*modules.Utxo {
	result := map[modules.OutPoint]*modules.Utxo{}
	p1 := modules.NewOutPoint(common.Hash{}, 0, 0)
	asset1 := &modules.Asset{AssetId: assid}
	fmt.Printf("Mock asset1:%s\n", asset1.String())
	utxo1 := &modules.Utxo{Asset: asset1, Amount: 1, LockTime: 0}
	utxo2 := &modules.Utxo{Asset: asset1, Amount: 10000, LockTime: 0}

	result[*p1] = utxo1
	p2 := modules.NewOutPoint(common.Hash{}, 1, 0)
	result[*p2] = utxo2
	return result
}

func TestConvertReadMap2Slice(t *testing.T) {
	rd := make(map[string]*KVRead)
	r1 := &KVRead{key: "bbb", value: []byte("bbb"), version: &modules.StateVersion{}}
	r2 := &KVRead{key: "ccc", value: []byte("ccc"), version: &modules.StateVersion{}}
	r3 := &KVRead{key: "aaa", value: []byte("aaa"), version: &modules.StateVersion{}}
	rd["bbb"] = r1
	rd["ccc"] = r2
	rd["aaa"] = r3
	result := convertReadMap2Slice(rd)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "aaa", result[0].GetKey())
	for i, r := range result {
		t.Logf("%d,%#v", i, r)
	}
}
