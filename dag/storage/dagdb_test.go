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
 *  * @date 2018
 *
 */

package storage

import (
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

var hash = common.HexToHash("0xfa4329fbb03fdd5d538a9a01a9af3b6f13e31d476ef9731adbee8bc4df688144")

func TestGetUnit(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	dagdb := NewDagDb(db)
	u, err := dagdb.GetHeaderByHash(common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"))
	assert.Nil(t, u, "empty db, must return nil Unit")
	assert.NotNil(t, err)
}

func TestPrintHashList(t *testing.T) {
	hash1 := common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
	hash2 := common.HexToHash("0xddff4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d493ee")
	txsHash := []common.Hash{hash1, hash2}
	t.Logf("%x", txsHash)

}

func TestGetHeader(t *testing.T) {
	key, _ := crypto.MyCryptoLib.KeyGen()
	pubKey,_:=crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h := new(modules.Header)
	au := modules.Authentifier{}
	address := crypto.PubkeyBytesToAddress(pubKey)
	t.Log("address:", address)

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &modules.ChainIndex{}
	h.Number.AssetID = modules.PTNCOIN
	h.Number.Index = uint64(333333)
	h.Extra = make([]byte, 20)
	h.ParentsHash = append(h.ParentsHash, h.TxRoot)
	h.TxRoot = common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	sig, _ := crypto.MyCryptoLib.Sign(key,h.TxRoot[:])
	au.Signature = sig
	au.PubKey = pubKey
	h.Authors = au
	h.Time = 123

	t.Logf("%#v", h)

	db, _ := ptndb.NewMemDatabase()
	dagdb := NewDagDb(db)

	err := dagdb.SaveHeader(h)
	assert.Nil(t, err)
	dbHeader, err := dagdb.GetHeaderByHash(h.Hash())
	assert.Nil(t, err)
	t.Logf("%#v", dbHeader)
	assertRlpHashEqual(t, h, dbHeader)
}

func TestGetTransaction(t *testing.T) {
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
	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}
	msg2 := &modules.Message{
		App:     modules.APP_DATA,
		Payload: &modules.DataPayload{MainData: []byte("Hello PalletOne"), ExtraData: []byte("Hi PalletOne")},
	}
	req := &modules.ContractInvokeRequestPayload{ContractId: []byte{123}, Args: [][]byte{{0x11}, {0x22}}, Timeout: 300}
	msg3 := &modules.Message{App: modules.APP_CONTRACT_INVOKE_REQUEST, Payload: req}
	tx := modules.NewTransaction(
		[]*modules.Message{msg, msg2, msg3},
	)
	t.Logf("%#v", tx)
	db, _ := ptndb.NewMemDatabase()
	dagdb := NewDagDb(db)

	err := dagdb.SaveTransaction(tx)
	assert.Nil(t, err)
	t.Log("tx ", tx)
	tx2, err := dagdb.GetTransactionOnly(tx.Hash())
	assert.Nil(t, err)
	t.Logf("%#v", tx2)
	assertRlpHashEqual(t, tx, tx2)
}
