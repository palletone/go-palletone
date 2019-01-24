package modules

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/go-cmp/cmp"
	"github.com/palletone/go-palletone/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInput_RLP(t *testing.T) {
	input:=NewTxIn(NewOutPoint(common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"),123,9999),[]byte{1,2,3})
	bytes,err:= rlp.EncodeToBytes(input)
	assert.Nil(t,err)
	t.Logf("Rlp data:%x",bytes)
	input2:=&Input{}
	err=rlp.DecodeBytes(bytes,input2)
	assert.Nil(t,err)
}

func TestCoinbaseInput_RLP(t *testing.T) {
	input:=NewTxIn(nil,nil)
	input.Extra=[]byte{0xff,0xee}
	bytes,err:= rlp.EncodeToBytes(input)
	assert.Nil(t,err)
	t.Logf("Rlp data:%x",bytes)
	input2:=&Input{}
	err=rlp.DecodeBytes(bytes,input2)
	assert.Nil(t,err)
}
func TestHash_Rlp(t *testing.T){
	type A struct{
		Hash common.Hash
	}

	 hash:=common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac")
	 a:=&A{hash}
	bytes,err:= rlp.EncodeToBytes(a)
	assert.Nil(t,err)
	t.Logf("Rlp data:%x",bytes)
	hash2:=&A{}
	err=rlp.DecodeBytes(bytes,hash2)
	t.Logf("Rlp data:%x",hash2.Hash)
	assert.Nil(t,err)

//
//func TestInput_RLP(t *testing.T) {
//	tinput:=&inputTemp{TxHash:common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"),MessageIndex:123,OutIndex:666,SignatureScript:[]byte{1,2,3},NullOutPoint:false}
//	bytes,err:= rlp.EncodeToBytes(tinput)
//	assert.Nil(t,err)
//	t.Logf("Rlp data:%x",bytes)
//
//	input:=NewTxIn(NewOutPoint(tinput.TxHash,tinput.MessageIndex,tinput.OutIndex),tinput.SignatureScript)
//	bytes2,err:=rlp.EncodeToBytes(input)
//	assert.Nil(t,err)
//	//assert.Equal(t,bytes,bytes2)
//	input2:=&Input{}
//	err=rlp.DecodeBytes(bytes2,input2)
//	assert.Nil(t,err)
//}
//func TestHash_Rlp(t *testing.T) {
//	type A struct {
//		Hash common.Hash
//	}
//
//	hash := common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac")
//	a := &A{hash}
//	bytes, err := rlp.EncodeToBytes(a)
//	assert.Nil(t, err)
//	t.Logf("Rlp data:%x", bytes)
//	hash2 := &A{}
//	err = rlp.DecodeBytes(bytes, hash2)
//	assert.Nil(t, err)
//
//}
//
//func TestOutput_Rlp(t *testing.T) {
//	a := &Asset{AssetId: PTNCOIN}
//	t.Logf("PTN:%x", a.Bytes())
//	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
//
//	t.Logf("Output data:%x",output)
//	bytes,err:= rlp.EncodeToBytes(output)
//	assert.Nil(t,err)
//	t.Logf("Rlp data:%x",bytes)
//	output2:=&Output{}
//	err=rlp.DecodeBytes(bytes,output2)
//	t.Logf("Output2 data:%x",output2)
//	assert.Nil(t,err)
//	assert.Equal(t,output,output2)
//}
//func TestPaymentPayload_Rlp(t *testing.T) {
//	pay:=newTestPayment()
//	t.Logf("Pay:%#v",pay)
//	bytes,err:= rlp.EncodeToBytes(pay)
//	assert.Nil(t,err)
//	t.Logf("Rlp data:%x",bytes)
//	pay2:=&PaymentPayload{}
//	err=rlp.DecodeBytes(bytes,pay2)
//	assert.Nil(t,err)
//	t.Logf("Pay:%#v",pay2)
//	assert.Equal(t,pay.Inputs[0].PreviousOutPoint.TxHash,pay2.Inputs[0].PreviousOutPoint.TxHash)
//}
//func newTestPayment() *PaymentPayload{
//	pay:=&PaymentPayload{LockTime:123}
//
//	bytes, err := rlp.EncodeToBytes(output)
//	assert.Nil(t, err)
//	t.Logf("Rlp data:%x", bytes)
//	output2 := &Output{}
//	err = rlp.DecodeBytes(bytes, output2)
//	assert.Nil(t, err)
//	assert.True(t, cmp.Equal(output, output2))
//}
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
//func TestDataPayload_Rlp(t *testing.T) {
//	pay:=&DataPayload{MainData:[]byte("test"),ExtraData:[]byte("test2")}
//	t.Logf("Pay:%#v",pay)
//	bytes,err:= rlp.EncodeToBytes(pay)
//	assert.Nil(t,err)
//	t.Logf("Rlp data:%x",bytes)
//	pay2:=&DataPayload{}
//	err=rlp.DecodeBytes(bytes,pay2)
//	assert.Nil(t,err)
//	t.Logf("Pay:%#v",pay2)
//	assert.Equal(t,pay.ExtraData,pay2.ExtraData)
//	assert.Equal(t,pay.MainData,pay2.MainData)
//}
//
//func TestContractTplPayload_Rlp(t *testing.T) {
//	pay:=newTestContractTpl()
//	t.Logf("Pay:%#v",pay)
//	bytes,err:= rlp.EncodeToBytes(pay)
//	assert.Nil(t,err)
//	t.Logf("Rlp data:%x",bytes)
//	pay2:=&ContractTplPayload{}
//	err=rlp.DecodeBytes(bytes,pay2)
//	assert.Nil(t,err)
//	t.Logf("Pay:%#v",pay2)
//	assert.Equal(t,pay,pay2)
//}
//
//func newTestContractTpl() *ContractTplPayload{
//	pay:=&ContractTplPayload{}
//	pay = NewContractTplPayload([]byte("1"),"test","test","test",123,[]byte("test"))
//
//	return pay
//}
//
//func TestContractDeployPayload_Rlp(t *testing.T) {
//	pay:=newTestContractDeloy()
//	t.Logf("Pay:%#v",pay)
//	bytes,err:= rlp.EncodeToBytes(pay)
//	assert.Nil(t,err)
//	t.Logf("Rlp data:%x",bytes)
//	pay2:=&ContractTplPayload{}
//	err=rlp.DecodeBytes(bytes,pay2)
//	assert.Nil(t,err)
//	t.Logf("Pay:%#v",pay2)
//	assert.Equal(t,pay,pay2)
//}

//func newTestContractDeloy() *ContractDeployPayload{
//	pay:=&ContractDeployPayload{}
//	pay = NewContractDeployPayload([]byte("1"),[]byte("123"),"test",[][]byte("test"),"test",123,[]byte("test"))
//
//	return pay
//}

}

type TestA struct {
	A        int
	B        string
	Parent   *TestA
	Children []*TestA
}

func TestCompare(t *testing.T) {
	a1 := &TestA{A: 1, B: "A1"}
	a2 := &TestA{A: 2, B: "A2"}
	a3 := &TestA{A: 3, B: "A3", Parent: a1}
	a11 := &TestA{A: 1, B: "A1"}
	a22 := &TestA{A: 2, B: "A2", Parent: &TestA{}}
	assert.True(t, cmp.Equal(a1, a11))
	assert.False(t, cmp.Equal(a2, a22))
	assert.True(t, cmp.Equal(a3.Parent, a11))
}

