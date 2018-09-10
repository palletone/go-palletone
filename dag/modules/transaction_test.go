package modules

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
)

// The values in those tests are from the Transaction Tests
// at github.com/ethereum/tests.

func TestTransactionEncode(t *testing.T) {

	pay1s := PaymentPayload{
		LockTime: 12345,
	}
	output := NewTxOut(1, []byte{}, &Asset{})
	pay1s.AddTxOut(*output)

	msg := &Message{
		App:     APP_PAYMENT,
		Payload: pay1s,
	}
	msg2 := &Message{
		App:     APP_TEXT,
		Payload: TextPayload{Text: []byte("Hello PalletOne")},
	}
	emptyTx := NewTransaction(
		[]*Message{msg, msg},
		1234,
	)

	rightvrsTx := NewTransaction(
		[]*Message{msg, msg2, msg},
		12345,
	)

	emptyTx.SetHash(common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87"))
	rightvrsTx.SetHash(common.HexToHash("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"))
	txb, err := rlp.EncodeToBytes(rightvrsTx)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	should := common.FromHex("f9010aa0000000000000000000000000b94f5374fce5edbc8e2a8697c15331677e6ebf0bf8e4f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878823039")
	if !bytes.Equal(txb, should) {
		log.Error("encoded RLP mismatch", "error", txb)
	}
	rlp_hash := new(common.Hash)
	*rlp_hash = rlp.RlpHash(rightvrsTx)
	rightvrsTx.SetHash(*rlp_hash)

	tx := new(Transaction)
	rlp.DecodeBytes(txb, tx)
	//if tx.Locktime != 12345 {
	//	log.Error("decode RLP mismatch", "error", txb)
	//}
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
	if len(payment.Output) == 0 {
		t.Error("payment out decode error.")
	}
	fmt.Printf("PaymentData:%+v", payment)
	tx.SetHash(rlp.RlpHash(tx))
	if tx.TxHash != rightvrsTx.TxHash {
		log.Error("tx hash mismatch ", "right_hash", rightvrsTx.TxHash, "tx_hash", tx.TxHash)
	}

}
