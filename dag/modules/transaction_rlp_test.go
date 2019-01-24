package modules

import (
	"testing"
	"github.com/palletone/go-palletone/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)
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
	assert.Nil(t,err)
}

func TestOutput_Rlp(t *testing.T){
	a := &Asset{AssetId: PTNCOIN}
	t.Logf("PTN:%x",a.Bytes())
	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	bytes,err:= rlp.EncodeToBytes(output)
	assert.Nil(t,err)
	t.Logf("Rlp data:%x",bytes)
	output2:=&Output{}
	err=rlp.DecodeBytes(bytes,output2)
	assert.Nil(t,err)
}
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
func newTestPayment() *PaymentPayload{
	pay:=&PaymentPayload{LockTime:123}
	a := &Asset{AssetId: PTNCOIN}

	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	pay.AddTxOut(output)
	hash:=common.HexToHash("0xe01c4bae7b396bc3c9bcb9275cef479560141c2010b6537abd78795bc935a2dd")
	input:=NewTxIn(NewOutPoint(hash,0,1),common.Hex2Bytes("0x40e608a3b177442c6c3476850078f48220b70c4efcdd4cb10ce62773d38231cff91c947d0f082b4854bf8675850f198f99b3981815c0e2527ecd790c26920f748821038cc8c907b29a58b00f8c2590303bfc93c69d773b9da204337678865ee0cafadb"))
	pay.AddTxIn(input)
	return pay
	}
