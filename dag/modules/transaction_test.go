package modules

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/constants"
)

// The values in those tests are from the Transaction Tests
// at github.com/ethereum/tests.
func TestTransactionHash(t *testing.T) {
	tx := &Transaction{}
	//tx.SetHash(common.HexToHash("e01c4bae7b396bc3c9bcb9275cef479560141c2010b6537abd78795bc935a2dd"))
	t.Log(tx.Hash().String())
}
func TestTransactionJson(t *testing.T) {
	pay1s := PaymentPayload{
		LockTime: 12345,
	}
	output := NewTxOut(99999999999999999, []byte{0xee, 0xbb}, NewPTNAsset())
	pay1s.AddTxOut(output)
	hash := common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	input := NewTxIn(NewOutPoint(&hash, 0, 1), []byte{})
	pay1s.AddTxIn(input)
	msg := &Message{
		App:     APP_PAYMENT,
		Payload: pay1s,
	}
	txmsg := NewTransaction(
		[]*Message{msg},
	)
	data, err := json.Marshal(txmsg)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("json ", string(data))
	}
	//
	txmsgNew := NewTransaction(
		[]*Message{msg},
	)
	errNew := json.Unmarshal(data, txmsgNew)
	if errNew != nil {
		fmt.Println(errNew)
	} else {
		fmt.Println("zzzz ", txmsgNew.TxMessages[0].Payload)
		data1, err1 := json.Marshal(txmsgNew.TxMessages[0].Payload)
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		fmt.Println(string(data1))
		payment := new(PaymentPayload)
		err2 := json.Unmarshal(data1, &payment)
		if err2 != nil {
			fmt.Println(err2)
		} else {
			fmt.Println(payment.Outputs[0].Value)
		}
	}
}
func TestTransactionEncode(t *testing.T) {

	pay1s := PaymentPayload{
		LockTime: 12345,
	}
	output := NewTxOut(1, []byte{0xee, 0xbb}, NewPTNAsset())
	pay1s.AddTxOut(output)
	hash := common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	input := NewTxIn(NewOutPoint(&hash, 0, 1), []byte{})
	pay1s.AddTxIn(input)
	msg := &Message{
		App:     APP_PAYMENT,
		Payload: pay1s,
	}
	msg2 := &Message{
		App:     APP_TEXT,
		Payload: TextPayload{TextHash: []byte("Hello PalletOne")},
	}
	//txmsg2 := NewTransaction(
	//	[]*Message{msg, msg},
	//)
	req := &ContractInvokeRequestPayload{ContractId: []byte{0xcc}, FunctionName: "TestFun", Args: [][]byte{[]byte{0x11}, {0x22}}}
	msg3 := &Message{App: APP_CONTRACT_INVOKE_REQUEST, Payload: req}
	txmsg3 := NewTransaction(
		[]*Message{msg, msg2, msg3},
	)

	//emptyTx.SetHash(common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87"))
	//rightvrsTx.SetHash(common.HexToHash("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"))
	txb, err := rlp.EncodeToBytes(txmsg3)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	//should := common.FromHex("f9010aa0000000000000000000000000b94f5374fce5edbc8e2a8697c15331677e6ebf0bf8e4f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878823039")
	//if !bytes.Equal(txb, should) {
	//	log.Error("encoded RLP mismatch", "error", txb)
	//}
	//rlp_hash := new(common.Hash)
	//*rlp_hash = rlp.RlpHash(txmsg3)
	//rightvrsTx.SetHash(*rlp_hash)
	// storage test

	tx := new(Transaction)
	err = rlp.DecodeBytes(txb, tx)
	if err != nil {
		t.Error(err)
	}
	//if tx.Locktime != 12345 {
	//	log.Error("decode RLP mismatch", "error", txb)
	//}
	//fmt.Println("tx:= ", tx)
	for _, msg := range tx.Messages() {
		if msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			fmt.Println("msg", pay.Inputs, pay.Outputs)
			for _, out := range pay.Outputs {
				fmt.Println("info:= ", out)
			}
		}
	}
	if len(tx.TxMessages) != 3 {
		t.Error("Rlp decode message count error")
	}
	msg0 := tx.TxMessages[0]
	if msg0.App != APP_PAYMENT {
		t.Error("Payment decode error")
	}
	payment := msg0.Payload.(*PaymentPayload)
	if payment.LockTime != 12345 {
		t.Error("payment locktime decode error.")
	}
	if len(payment.Outputs) == 0 {
		t.Error("payment out decode error.")
	}
	if len(payment.Inputs) == 0 {
		t.Error("payment input decode error.")
	}
	fmt.Printf("PaymentData:%+v", payment)
	//tx.SetHash(rlp.RlpHash(tx))
	//if tx.TxHash != rightvrsTx.TxHash {
	//	log.Error("tx hash mismatch ", "right_hash", rightvrsTx.TxHash, "tx_hash", tx.TxHash)
	//}

}
func TestIDType16Hex(t *testing.T) {
	PTNCOIN := IDType16{'p', 't', 'n', 'c', 'o', 'i', 'n'}
	fmt.Println("ptn hex:", PTNCOIN.String())
	fmt.Println("ptn hex:", PTNCOIN)
	fmt.Println("btc hex:", BTCCOIN.String())
	key := fmt.Sprintf("%s_%s_1_%d", constants.UNIT_NUMBER_PREFIX, "abc", 100)
	slice := strings.Split(key, fmt.Sprintf("%s_%s_1_", constants.UNIT_NUMBER_PREFIX, "abc"))
	fmt.Println("result:", len(slice), "0:", slice[0], "1:", slice[1])

	var tx Transaction
	str := "{\"txhash\":\"0xaa0fbe87c07b063cd6a88ab8e2c0075bec35bc80a56956cd50ce98aad3febca6\",\"messages\":[{\"App\":0,\"Payload\":{\"Inputs\":[{\"PreviousOutPoint\":null,\"SignatureScript\":null,\"Extra\":\"W+vkvg==\"}],\"Outputs\":[{\"Value\":100000000,\"PkScript\":\"dqkUj1ulfgUxOae0LG5IueWUIzBQk2WIrA==\",\"Asset\":{\"asset_id\":[119,169,59,162,215,104,17,232,157,4,140,133,144,10,158,67],\"unique_id\":[119,169,59,162,215,104,17,232,157,4,140,133,144,10,158,67],\"chain_id\":1}}],\"LockTime\":0}}]}"
	err := json.Unmarshal([]byte(str), &tx)
	fmt.Println("error: ", err)
	for _, msg := range tx.Messages() {
		fmt.Println("info: ", msg.Payload)
		data, err := json.Marshal(msg.Payload)
		if err != nil {
			return
		}
		payment := new(PaymentPayload)
		err1 := json.Unmarshal(data, &payment)
		for j, out := range payment.Outputs {
			fmt.Println("payment: ", err1, j, out)
		}

	}
}
func TestTransaction_EncodeRLP_Size(t *testing.T) {
	pay1s := PaymentPayload{
		LockTime: 12345,
	}
	a := &Asset{AssetId: PTNCOIN}

	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	pay1s.AddTxOut(output)

	msg := &Message{
		App:     APP_PAYMENT,
		Payload: pay1s,
	}
	tx := NewTransaction(
		[]*Message{},
	)
	for i := 1; i < 1000; i++ {
		tx.AddMessage(msg)
		txb, _ := rlp.EncodeToBytes(tx)
		t.Logf("input count:{%d}, encode tx size:%d\n", i, len(txb))
	}
}
