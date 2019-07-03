/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package storage

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestUnitNumberIndex(t *testing.T) {
	key1 := fmt.Sprintf("%s_%s_%d", constants.UNIT_HASH_NUMBER_PREFIX, modules.BTCCOIN.String(), 10000)
	key2 := fmt.Sprintf("%s_%s_%d", constants.UNIT_HASH_NUMBER_PREFIX, modules.PTNCOIN.String(), 678934)

	if key1 != "nh_btcoin_10000" {
		log.Debug("not equal.", "key1", key1)
	}
	if key2 != "nh_ptncoin_678934" {
		log.Debug("not equal.", "key2", key2)
	}
}

func TestGetBody(t *testing.T) {
	dbconn, _ := ptndb.NewMemDatabase()
	if dbconn == nil {
		fmt.Println("Connect to db error.")
		return
	}
	key := "ub0x6fc88cbedc9c99d238c10374274443d4460de9162795faf8a3442abe33db72fa"
	data, err := dbconn.Get([]byte(key))
	if err != nil {
		fmt.Println("get body hashs error:", err, string(key))
		return
	}
	var txhashs []common.Hash
	if err := rlp.DecodeBytes(data, &txhashs); err != nil {
		fmt.Println("decode hashs error:", err)
	}

	for in, hash := range txhashs {
		fmt.Println("index:", in, "hash:", hash.String())
		key1 := append(constants.TRANSACTION_PREFIX, []byte(hash.String())...)
		data1, err1 := dbconn.Get(key1)
		if err1 != nil {
			fmt.Println("get body hashs error:", err1, string(key))
			return
		}
		tx := new(modules.Transaction)

		if err := rlp.DecodeBytes(data1, &tx); err != nil {
			fmt.Println("decode tx error:", string(key1), err)
		}
		for _, msg := range tx.TxMessages {
			fmt.Println("tx msg info ", msg)
			payment, ok := msg.Payload.(*modules.PaymentPayload)
			fmt.Println("payment info ", ok, payment)
		}
	}
}

func TestRLPTxDecode(t *testing.T) {
	pay1s := &modules.PaymentPayload{
		LockTime: 12345,
	}
	output := modules.NewTxOut(1, []byte{0xee, 0xbb}, modules.NewPTNAsset())
	pay1s.AddTxOut(output)
	hash := common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	input := modules.Input{}
	input.PreviousOutPoint = modules.NewOutPoint(hash, 0, 1)
	input.SignatureScript = []byte{}
	input.Extra = []byte("Coinbase")
	fmt.Println(input)
	fmt.Println(input.PreviousOutPoint)
	pay1s.AddTxIn(&input)
	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}
	msg2 := &modules.Message{
		App:     modules.APP_DATA,
		Payload: &modules.DataPayload{MainData: []byte("Hello PalletOne"), ExtraData: []byte("Hi PalletOne")},
	}

	req := &modules.ContractInvokeRequestPayload{ContractId: []byte{0xcc}, Args: [][]byte{[]byte{0x11}, {0x22}}}
	msg3 := &modules.Message{App: modules.APP_CONTRACT_INVOKE_REQUEST, Payload: req}
	txmsg3 := modules.NewTransaction(
		[]*modules.Message{msg, msg2, msg3},
	)
	dbconn, _ := ptndb.NewMemDatabase()
	tx_bytes, _ := rlp.EncodeToBytes(txmsg3)
	key := []byte("this_is_testing_tx_encode_decode")
	dbconn.Put(key, tx_bytes)

	val, _ := dbconn.Get(key)
	tx := new(modules.Transaction)
	rlp.DecodeBytes(val, &tx)
	for _, msg := range tx.Messages() {
		if msg.App == modules.APP_PAYMENT {
			pay, ok := msg.Payload.(*modules.PaymentPayload)
			fmt.Println("断言结果：", ok)
			for _, out := range pay.Outputs {
				fmt.Println("output:= ", out)
			}
			for _, in := range pay.Inputs {
				fmt.Println("input:= ", in)
			}

		} else if msg.App == modules.APP_DATA {
			text := msg.Payload.(*modules.DataPayload)
			fmt.Println("msg_app", msg.App, "text", string(text.MainData))
		} else {
			req := msg.Payload.(*modules.ContractInvokeRequestPayload)
			fmt.Println("msg_app", msg.App, "req", req)
		}

	}
	assert.Equal(t, txmsg3.Hash(), tx.Hash())

}
