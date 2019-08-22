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
 *  * @date 2018-2019
 *
 */

package validator

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/stretchr/testify/assert"
	"time"
)

var privKeyBytes, _ = hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")

func getAccount() (*ecdsa.PrivateKey, []byte, common.Address) {
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	pubKey := crypto.CompressPubkey(&privKey.PublicKey)
	addr := crypto.PubkeyBytesToAddress(pubKey)
	return privKey, pubKey, addr
}
func newCache() palletcache.ICache {
	return freecache.NewCache(100 * 1024)
}
func TestValidate_ValidateTx_EmptyTx_NoPayment(t *testing.T) {
	tx := &modules.Transaction{} //Empty Tx
	validat := NewValidate(nil, nil, nil, nil, newCache())
	_, _, err := validat.ValidateTx(tx, true)
	assert.NotNil(t, err)
	t.Log(err)
	tx.AddMessage(modules.NewMessage(modules.APP_DATA, &modules.DataPayload{MainData: []byte("m")}))
	_, _, err = validat.ValidateTx(tx, true)
	assert.NotNil(t, err)
	t.Log(err)
}
func TestValidate_ValidateTx_MsgCodeIncorrect(t *testing.T) {
	tx := &modules.Transaction{}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, &modules.DataPayload{MainData: []byte("m")}))

	validat := NewValidate(nil, nil, nil, nil, newCache())
	_, _, err := validat.ValidateTx(tx, true)
	assert.NotNil(t, err)
	t.Log(err)

}

var hash1 = common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac")

// func TestValidate_ValidateTx_IncorrectFee(t *testing.T) {
// 	tx := &modules.Transaction{}
// 	outPoint := &modules.OutPoint{hash1, 0, 1}
// 	pay1 := newTestPayment(outPoint, 2000)
// 	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay1))
// 	signTx(tx, outPoint)
// 	utxoq := &testutxoQuery{}
// 	validat := NewValidate(nil, utxoq, nil)
// 	err := validat.ValidateTx(tx, false)
// 	assert.NotNil(t, err)
// 	t.Log(err)

// 	pay1.Outputs[0].Value = 999
// 	signTx(tx, outPoint)
// 	err = validat.ValidateTx(tx, false)
// 	assert.Nil(t, err)
// 	t.Log(err)

// }

func signTx(tx *modules.Transaction, outPoint *modules.OutPoint) {
	privKey, _, addr := getAccount()
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	lockScripts := map[modules.OutPoint][]byte{
		*outPoint: lockScript[:],
	}

	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		s, e := crypto.MyCryptoLib.Sign(privKeyBytes, msg)
		return s, e
	}
	var hashtype uint32
	hashtype = 1
	_, e := tokenengine.Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if e != nil {
		fmt.Println(e.Error())
	}
}

type testutxoQuery struct {
}

func (u *testutxoQuery) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	_, _, addr := getAccount()
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	if outpoint.TxHash == hash1 {
		return &modules.Utxo{Asset: modules.NewPTNAsset(), Amount: 1000, PkScript: lockScript}, nil
	}
	return nil, errors.New("Incorrect Hash")
}
func (u *testutxoQuery) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	return nil, nil
}
func newTestPayment(point *modules.OutPoint, outAmt uint64) *modules.PaymentPayload {
	pay1s := &modules.PaymentPayload{
		LockTime: 12345,
	}
	a := &modules.Asset{AssetId: modules.PTNCOIN}

	output := modules.NewTxOut(outAmt, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	pay1s.AddTxOut(output)
	i := modules.NewTxIn(point, []byte{})
	pay1s.AddTxIn(i)
	return pay1s
}

func TestGetRequestTx(t *testing.T) {
	msgs := make([]*modules.Message, 0)
	msg := modules.Message{}
	// payment msg
	msg.App = modules.APP_PAYMENT
	input := make([]*modules.Input, 0)
	out := make([]*modules.Output, 0)
	input = []*modules.Input{&modules.Input{PreviousOutPoint: modules.NewOutPoint(common.HexToHash("0xb17041fe6ef735b8be14f1f54b7b888b663c3074730cc8f82455d69450a533bf"), 0, 0), SignatureScript: []byte("test_sig"), Extra: []byte("jay")}}
	out = []*modules.Output{&modules.Output{Value: 10000, PkScript: []byte("test_pk"), Asset: modules.NewPTNAsset()}}
	pay := modules.NewPaymentPayload(input, out)
	msg.Payload = pay
	msgs = append(msgs, &msg)

	// contact msg
	//msg1 := new(modules.Message)
	//msg1.App = modules.APP_CONTRACT_INVOKE_REQUEST
	//invoke_req := new(modules.ContractInvokeRequestPayload)
	//
	//invoke_req.ContractId = []byte("test_contact_invoke_request")
	//invoke_req.ContractId = []byte("test_contract_id")
	//invoke_req.Args = make([][]byte, 0)
	//invoke_req.Timeout = 10 * time.Second
	//msg1.Payload = invoke_req
	//msgs = append(msgs, msg1)

	// contact stop msg
	msg2 := new(modules.Message)
	stop := new(modules.ContractStopRequestPayload)
	stop.ContractId = []byte("test_contact_stop_id")
	stop.Txid = "0xb17041fe6ef735b8be14f1f54b7b888b663c3074730cc8f82455d69450a533bf"
	stop.DeleteImage = true
	msg2.App = modules.APP_CONTRACT_STOP_REQUEST
	msg2.Payload = stop
	msgs = append(msgs, msg2)

	tx := modules.NewTransaction(msgs)
	req_tx := tx.GetRequestTx()
	if req_tx == nil {
		t.Fatal("get req_tx failed.")
		return
	}

	t_msgs := tx.Messages()
	r_msgs := req_tx.Messages()
	t_hash := tx.Hash()
	r_hash := req_tx.Hash()
	if t_hash != r_hash {
		log.Println("t_hash:", t_hash.String())
		log.Println("r_hash:", r_hash.String())
		for _, msg := range t_msgs {
			log.Println("tmsg:", msg.App, msg.Payload)
		}
		for _, msg := range r_msgs {
			log.Println("rmsg:", msg.App, msg.Payload)
		}
		t.Fatal("failed.", t_msgs, r_msgs)
	} else {
		for _, msg := range r_msgs {
			log.Println("rmsg:", msg.App, msg.Payload)
		}
	}
}
func TestValidateDoubleSpendOn1Tx(t *testing.T) {
	outPoint := modules.NewOutPoint(hash1, 0, 1)
	pay1 := newTestPayment(outPoint, 1)

	tx := &modules.Transaction{}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay1))

	signTx(tx, outPoint)
	utxoq := &testutxoQuery{}
	validate := NewValidate(nil, utxoq, nil, nil, newCache())
	_, _, err := validate.ValidateTx(tx, true)
	assert.Nil(t, err)
	pay2 := newTestPayment(outPoint, 2)
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay2))
	signTx(tx, outPoint)
	_, _, err1 := validate.ValidateTx(tx, true)
	assert.NotNil(t, err1)
	t.Log(err1)
}

//构造一个上千Input的交易，验证时间要多久？
func TestValidateLargeInputPayment(t *testing.T) {
	N := 1000
	tx := &modules.Transaction{}
	pay := &modules.PaymentPayload{Inputs: []*modules.Input{}, Outputs: []*modules.Output{}}
	lockScripts := map[modules.OutPoint][]byte{}
	privKey, _, addr := getAccount()
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	for i := 0; i < N; i++ {
		outpoint := modules.NewOutPoint(hash1, 0, uint32(i))
		lockScripts[*outpoint] = lockScript
		in := modules.NewTxIn(outpoint, nil)
		pay.Inputs = append(pay.Inputs, in)
	}
	output := modules.NewTxOut(100, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), modules.NewPTNAsset())
	pay.AddTxOut(output)
	tx.TxMessages = []*modules.Message{modules.NewMessage(modules.APP_PAYMENT, pay)}
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		s, e := crypto.MyCryptoLib.Sign(privKeyBytes, msg)
		return s, e
	}
	var hashtype uint32
	hashtype = 1
	_, e := tokenengine.Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if e != nil {
		fmt.Println(e.Error())
	}
	//data, _ := json.Marshal(tx)
	//t.Logf("Signed Tx:%s", string(data))

	utxoq := &testutxoQuery{}
	validate := NewValidate(nil, utxoq, nil, nil, newCache())
	_, _, err := validate.ValidateTx(tx, true)

	t1 := time.Now()
	//validate := NewValidate(nil, utxoq, nil, nil)
	_, _, err = validate.ValidateTx(tx, true)
	t.Logf("Validate send time:%s", time.Since(t1))
	assert.Nil(t, err)
}
