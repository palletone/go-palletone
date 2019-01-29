package modules

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

var hash = common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac")

type TestA struct {
	A        uint64
	B        string
	Parent   *TestA
	Children []*TestA
}

func TestRlpA(t *testing.T) {
	a := &TestA{A: 123, B: "abc"}
	b1, _ := rlp.EncodeToBytes(a)
	a1 := &TestA{}
	t.Logf("Rlp:%x", b1)
	b := &TestA{}
	a1.Parent = b
	b2, _ := rlp.EncodeToBytes(a1)
	t.Logf("Rlp:%x", b2)
	err := rlp.DecodeBytes(b2, a1)
	if err != nil {
		t.Log(err.Error())
	}
}

type TestInput struct {
	SignatureScript  []byte
	Extra            []byte // if user creating a new asset, this field should be it's config data. Otherwise it is null.
	PreviousTxHash   common.Hash
	PreviousMsgIndex uint32 // message index in transaction
	PreviousOutIndex uint32
}
type TestPayment struct {
	Inputs   []*TestInput
	Outputs  []*Output
	LockTime uint32
}

func newTestInput(txHash common.Hash, msgIndex, outIndex uint32, unlockScript, extra []byte) *TestInput {
	return &TestInput{SignatureScript: unlockScript, Extra: extra, PreviousTxHash: txHash, PreviousMsgIndex: msgIndex, PreviousOutIndex: outIndex}
}
func newCoinbaseInput(unlockScript, extra []byte) *TestInput {
	return &TestInput{SignatureScript: unlockScript, Extra: extra}
}

type TestOutput struct {
	TxHash       common.Hash `json:"txhash"`        // reference Utxo struct key field
	MessageIndex uint32      `json:"message_index"` // message index in transaction
	OutIndex     uint32      `json:"out_index"`
}

func assertRlpHashEqual(t assert.TestingT, a, b interface{}) {
	hash1 := util.RlpHash(a)
	hash2 := util.RlpHash(b)
	assert.Equal(t, hash1, hash2)
}

//func assertRlpTxEqual(t assert.TestingT, a, b interface{}) {
//	tx1 := test
//	tx2 := Transaction{}
//	assert.Equal(t, hash1, hash2)
//}

//func TestCompare(t *testing.T) {
//	a1 := &TestA{A: 1, B: "A1"}
//	a2 := &TestA{A: 2, B: "A2"}
//	a3 := &TestA{A: 3, B: "A3", Parent: a1}
//	a11 := &TestA{A: 1, B: "A1"}
//	a22 := &TestA{A: 2, B: "A2", Parent: &TestA{}}
//
//	assert.True(t, cmp.Equal(a1, a11))
//	assert.False(t, cmp.Equal(a2, a22))
//	assert.True(t, cmp.Equal(a3.Parent, a11))
//}

func TestInput_RLP(t *testing.T) {
	input := newTestInput(common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), 123, 9999, []byte{1, 2, 3}, nil)

	bytes, err := rlp.EncodeToBytes(input)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	input2 := &TestInput{}
	err = rlp.DecodeBytes(bytes, input2)
	assert.Nil(t, err)
	//assert.Equal(t, input, input2)
	hash1 := util.RlpHash(input)
	hash2 := util.RlpHash(input2)
	assert.Equal(t, hash1, hash2)

	input3 := NewTxIn(NewOutPoint(input.PreviousTxHash, input.PreviousMsgIndex, input.PreviousOutIndex), input.SignatureScript)
	input3.Extra = input.Extra
	bytes3, _ := rlp.EncodeToBytes(input3)
	assert.Equal(t, bytes, bytes3)
	assertRlpHashEqual(t, input, input3)
}

func TestCoinbaseInput_RLP(t *testing.T) {
	input := newCoinbaseInput([]byte("unlock"), []byte("extra"))
	input.Extra = []byte{1, 2, 3}
	//ptr := *(*byte)(unsafe.Pointer(&input))
	t.Log("data", input)
	t.Log("data", input.PreviousTxHash)
	t.Log("data:", input.SignatureScript == nil)
	bytes, err := rlp.EncodeToBytes(input)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	input2 := &TestInput{}
	err = rlp.DecodeBytes(bytes, input2)
	t.Log("data", input2)
	t.Log("data:", input2.SignatureScript == nil)
	assert.Nil(t, err)
	assert.Equal(t, input, input2)

	input3 := NewTxIn(nil, input.SignatureScript)
	input3.Extra = input.Extra
	bytes3, _ := rlp.EncodeToBytes(input3)
	assert.Equal(t, bytes, bytes3)
	assertRlpHashEqual(t, input, input3)
}

func TestOutput_Rlp(t *testing.T) {
	a := &Asset{AssetId: PTNCOIN}
	t.Logf("PTN:%x", a.Bytes())
	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)

	t.Logf("Output data:%x", output)
	bytes, err := rlp.EncodeToBytes(output)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	output2 := &Output{}
	err = rlp.DecodeBytes(bytes, output2)
	t.Logf("Output2 data:%x", output2)
	assert.Nil(t, err)
	assert.Equal(t, output, output2)
}

func TestPaymentPayload_Rlp(t *testing.T) {
	pay := newTestPayment(true)
	t.Logf("Pay:%#v", pay)
	bytes, err := rlp.EncodeToBytes(pay)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	pay2 := &TestPayment{}
	err = rlp.DecodeBytes(bytes, pay2)
	assert.Nil(t, err)
	t.Logf("Pay:%#v", pay2)
	assertRlpHashEqual(t, pay, pay2)
}
func newTestPayment(includeCoinbase bool) *TestPayment {
	pay := &TestPayment{LockTime: 123, Inputs: []*TestInput{}, Outputs: []*Output{}}
	if includeCoinbase {
		pay.Inputs = append(pay.Inputs, newCoinbaseInput([]byte("test"), []byte("Extra")))
	}
	input := newTestInput(common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), 123, 9999, []byte{1, 2, 3}, nil)

	pay.Inputs = append(pay.Inputs, input)
	a := &Asset{AssetId: PTNCOIN}

	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	pay.Outputs = append(pay.Outputs, output)
	return pay
}

//
//func TestPaymentPayload_Rlp(t *testing.T) {
//	pay:=newTestPayment()
//	bytes,err:= rlp.EncodeToBytes(pay)
//	assert.Nil(t,err)
//	t.Logf("Rlp data:%x",bytes)
//	pay2:=&PaymentPayload{}
//	err=rlp.DecodeBytes(bytes,pay2)
//	assert.Nil(t,err)
//	t.Logf("Pay:%#v",pay2)
//	assert.Equal(t,pay.Inputs[0].PreviousOutPoint.TxHash,pay2.Inputs[0].PreviousOutPoint.TxHash)
//}
//func newTestPayment() *PaymentPayload {
//	pay := &PaymentPayload{LockTime: 123}
//
//	a := &Asset{AssetId: PTNCOIN}
//
//	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
//	pay.AddTxOut(output)
//	hash := common.HexToHash("0xe01c4bae7b396bc3c9bcb9275cef479560141c2010b6537abd78795bc935a2dd")
//	input := NewTxIn(NewOutPoint(hash, 0, 1), common.Hex2Bytes("0x40e608a3b177442c6c3476850078f48220b70c4efcdd4cb10ce62773d38231cff91c947d0f082b4854bf8675850f198f99b3981815c0e2527ecd790c26920f748821038cc8c907b29a58b00f8c2590303bfc93c69d773b9da204337678865ee0cafadb"))
//	pay.AddTxIn(input)
//	return pay
//
//	}
//
func TestDataPayload_Rlp(t *testing.T) {
	pay := &DataPayload{MainData: []byte("test"), ExtraData: []byte("test2")}
	t.Logf("Pay:%#v", pay)
	bytes, err := rlp.EncodeToBytes(pay)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	pay2 := &DataPayload{}
	err = rlp.DecodeBytes(bytes, pay2)
	assert.Nil(t, err)
	t.Logf("Pay:%#v", pay2)
	assertRlpHashEqual(t, pay, pay2)
}

//
func TestContractTplPayload_Rlp(t *testing.T) {
	pay := newTestContractTpl()
	t.Logf("Pay:%#v", pay)
	bytes, err := rlp.EncodeToBytes(pay)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	pay2 := &ContractTplPayload{}
	err = rlp.DecodeBytes(bytes, pay2)
	assert.Nil(t, err)
	t.Logf("Pay:%#v", pay2)
	assertRlpHashEqual(t, pay, pay2)
}

func newTestContractTpl() *ContractTplPayload {
	pay := &ContractTplPayload{}
	pay = NewContractTplPayload([]byte("1"), "test", "test", "test", 123, []byte("test"))

	return pay
}

type TestContractInvokeRequestPayload struct {
	ContractId   []byte   `json:"contract_id"` // contract id
	FunctionName string   `json:"function_name"`
	Args         [][]byte `json:"args"` // contract arguments list
	Timeout      uint32   `json:"timeout"`
}

func TestContractInvokeReqPayload_Rlp(t *testing.T) {
	pay := newTestContractInvokeReq()
	t.Logf("Pay:%#v", pay)
	bytes, err := rlp.EncodeToBytes(pay)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	pay2 := &TestContractInvokeRequestPayload{}
	err = rlp.DecodeBytes(bytes, pay2)
	assert.Nil(t, err)
	t.Logf("Pay:%#v", pay2)
	assertRlpHashEqual(t, pay, pay2)
}

func newTestContractInvokeReq() *TestContractInvokeRequestPayload {
	a := []byte("AAAA")
	b := []byte("BBBBBBBBBBB")
	args := [][]byte{a, b, nil}
	pay := &TestContractInvokeRequestPayload{[]byte("ContractId"), "Func1", args, 3}

	return pay
}

func TestContractInvokeResultPayload_Rlp(t *testing.T) {
	pay := newTestContractInvokeResult()
	t.Logf("Pay:%#v", pay)
	bytes, err := rlp.EncodeToBytes(pay)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	pay2 := &ContractInvokePayload{}
	err = rlp.DecodeBytes(bytes, pay2)
	assert.Nil(t, err)
	t.Logf("Pay:%#v", pay2)
	assertRlpHashEqual(t, pay, pay2)
}

func newTestContractInvokeResult() *ContractInvokePayload {
	a := []byte("AAAA")
	b := []byte("BBBBBBBBBBB")
	args := [][]byte{a, b, nil}

	version := &StateVersion{&ChainIndex{PTNCOIN, true, 100}, 2}
	read1 := ContractReadSet{"A", version, []byte("This is value")}
	readset := []ContractReadSet{read1}
	write1 := ContractWriteSet{false, "Key1", []byte("This is value2")}
	wset := []ContractWriteSet{write1}

	pay := &ContractInvokePayload{[]byte("ContractId"), "Func1", args, readset, wset, nil}

	return pay
}
