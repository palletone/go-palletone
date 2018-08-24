package modules

import (
	"bytes"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
)

// The values in those tests are from the Transaction Tests
// at github.com/ethereum/tests.
var (
	msg = Message{
		App:     "payment",
		Payload: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	emptyTx = NewTransaction(
		[]Message{msg, msg},
		1234,
	)

	rightvrsTx = NewTransaction(
		[]Message{msg, msg, msg},
		12345,
	)
)

func TestTransactionEncode(t *testing.T) {
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
	if tx.Locktime != 12345 {
		log.Error("decode RLP mismatch", "error", txb)
	}

	tx.SetHash(rlp.RlpHash(tx))
	if tx.TxHash != rightvrsTx.TxHash {
		log.Error("tx hash mismatch ", "right_hash", rightvrsTx.TxHash, "tx_hash", tx.TxHash)
	}

}
