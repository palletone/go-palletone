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
	"encoding/hex"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/parameter"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/stretchr/testify/assert"
)

func TestValidate_ValidateUnitTxs(t *testing.T) {
	parameter.CurrentSysParameters.GenerateUnitReward = 0
	//构造一个Unit包含3个Txs，
	//0是Coinbase，收集3Dao手续费
	//1是普通Tx，100->99 付1Dao手续费，产生1Utxo
	//2是普通Tx， 99->97 付2Dao手续费，使用Tx1中的一个Utxo，产生1Utxo
	tx0 := newCoinbaseTx()
	tx1 := newTx1(t)
	outPoint := modules.NewOutPoint(tx1.Hash(), 0, 0)
	tx2 := newTx2(t, outPoint)
	txs := modules.Transactions{tx0, tx1, tx2}

	utxoQuery := &mockUtxoQuery{}
	mockStatedbQuery := &mockStatedbQuery{}
	validate := NewValidate(nil, utxoQuery, mockStatedbQuery, nil, newCache())
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	code := validate.validateTransactions(txs, time.Now().Unix(), addr)
	assert.Equal(t, code, TxValidationCode_VALID)
}

type mockStatedbQuery struct {
}

func (q *mockStatedbQuery) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return nil, nil
}
func (q *mockStatedbQuery) GetMediators() map[common.Address]bool {
	return nil
}

func (q *mockStatedbQuery) GetMediator(add common.Address) *core.Mediator {
	return nil
}
func (q *mockStatedbQuery) GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error) {
	return []common.Address{},nil,nil
}
//获得系统配置的最低手续费要求
func (q *mockStatedbQuery) GetMinFee() (*modules.AmountAsset, error) {
	return &modules.AmountAsset{Asset: modules.NewPTNAsset(), Amount: uint64(1)}, nil
}
func (q *mockStatedbQuery) GetContractJury(contractId []byte) (*modules.ElectionNode, error) {
	return nil, nil
}
func (q *mockStatedbQuery) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return nil, nil, nil
}
func (q *mockStatedbQuery) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return map[string]*modules.ContractStateValue{}, nil
}

type mockUtxoQuery struct {
}

func (q *mockUtxoQuery) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	return nil, nil
}

func (q *mockUtxoQuery) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	hash := common.HexToHash("1")
	//result := map[*modules.OutPoint]*modules.Utxo{}
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	utxo := &modules.Utxo{Amount: 100, LockTime: 0, Asset: modules.NewPTNAsset(), PkScript: lockScript}
	if outpoint.TxHash == hash {
		return utxo, nil
	}
	//result[modules.NewOutPoint(hash, 0, 0)] = utxo
	//if u, ok := result[outpoint]; ok {
	//	return u, nil
	//}
	return nil, errors.New("No utxo found")
}

func newCoinbaseTx() *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	output := modules.NewTxOut(3, lockScript, modules.NewPTNAsset())
	pay1s.AddTxOut(output)
	input := modules.Input{}
	input.Extra = []byte("Coinbase")

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}

	tx := modules.NewTransaction(
		[]*modules.Message{msg},
	)
	return tx
}

func newTx1(t *testing.T) *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
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
	//Sign

	lockScripts := map[modules.OutPoint][]byte{
		*input.PreviousOutPoint: lockScript,
	}
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		return crypto.MyCryptoLib.Sign(privKeyBytes, msg)
	}
	_, err := tokenengine.Instance.SignTxAllPaymentInput(tx, 1, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)

	return tx
}
func newTx2(t *testing.T, outpoint *modules.OutPoint) *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	output := modules.NewTxOut(97, []byte{}, modules.NewPTNAsset())
	pay1s.AddTxOut(output)
	input := modules.Input{}
	input.PreviousOutPoint = outpoint
	input.SignatureScript = []byte{}

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}

	tx := modules.NewTransaction(
		[]*modules.Message{msg},
	)
	//Sign
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	lockScripts := map[modules.OutPoint][]byte{
		*input.PreviousOutPoint: lockScript,
	}
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		return crypto.MyCryptoLib.Sign(privKeyBytes, msg)
	}
	_, err := tokenengine.Instance.SignTxAllPaymentInput(tx, 1, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)
	return tx
}
func newHeader(txs modules.Transactions) *modules.Header {
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
	return header
}
func TestValidate_ValidateHeader(t *testing.T) {
	tx := newTx1(t)

	header := newHeader(modules.Transactions{tx})
	v := NewValidate(nil, nil, nil, nil, newCache())
	vresult := v.validateHeaderExceptGroupSig(header)
	t.Log(vresult)
	assert.Equal(t, vresult, TxValidationCode_VALID)
}

func TestSignAndVerifyATx(t *testing.T) {

	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")

	pubKeyBytes, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(privKeyBytes)
	pubKeyHash := crypto.Hash160(pubKeyBytes)
	t.Logf("Public Key:%x", pubKeyBytes)
	addr := crypto.PubkeyBytesToAddress(pubKeyBytes)
	t.Logf("Addr:%s", addr.String())
	lockScript := tokenengine.Instance.GenerateP2PKHLockScript(pubKeyHash)
	t.Logf("UTXO lock script:%x", lockScript)

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	asset0 := &modules.Asset{}
	payment.AddTxOut(modules.NewTxOut(1, lockScript, asset0))
	payment2 := &modules.PaymentPayload{}
	utxoTxId2 := common.HexToHash("1651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint2 := modules.NewOutPoint(utxoTxId2, 1, 1)
	txIn2 := modules.NewTxIn(outPoint2, []byte{})
	payment2.AddTxIn(txIn2)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	payment2.AddTxOut(modules.NewTxOut(1, lockScript, asset1))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment2))

	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_DATA, &modules.DataPayload{MainData: []byte("Hello PalletOne")}))

	lockScripts := map[modules.OutPoint][]byte{
		*outPoint:  lockScript[:],
		*outPoint2: tokenengine.Instance.GenerateP2PKHLockScript(pubKeyHash),
	}
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	addr: privKey,
	//}
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return pubKeyBytes, nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		return crypto.MyCryptoLib.Sign(privKeyBytes, hash)
	}
	var hashtype uint32
	hashtype = 1
	_, err := tokenengine.Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)

}
func TestTime(t *testing.T) {
	ti, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-08-02 00:00:00", time.Local)
	t.Log(ti.Format("2006-01-02 15:04:05"))
	t.Log(ti.Unix())
}
